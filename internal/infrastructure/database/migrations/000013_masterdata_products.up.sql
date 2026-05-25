-- Master Data: Products (product_categories, units_of_measure, products, product_variants, product_prices)

CREATE TABLE IF NOT EXISTS product_categories (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code        VARCHAR(20)  NOT NULL,
    name        VARCHAR(200) NOT NULL,
    parent_id   BIGINT       REFERENCES product_categories(id),
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP,
    CONSTRAINT uidx_product_categories_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_product_categories_tenant_id ON product_categories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_product_categories_parent_id ON product_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_product_categories_deleted_at ON product_categories(deleted_at);

CREATE TRIGGER update_product_categories_updated_at
    BEFORE UPDATE ON product_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS units_of_measure (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT      NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code          VARCHAR(20) NOT NULL,
    name          VARCHAR(100) NOT NULL,
    uom_type      VARCHAR(20) NOT NULL DEFAULT 'UNIT' CHECK (uom_type IN ('UNIT','WEIGHT','VOLUME','LENGTH','AREA','TIME')),
    base_uom_id   BIGINT      REFERENCES units_of_measure(id),
    conversion_factor NUMERIC(18,6) NOT NULL DEFAULT 1,
    is_active     BOOLEAN     NOT NULL DEFAULT true,
    created_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_units_of_measure_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_units_of_measure_tenant_id ON units_of_measure(tenant_id);

CREATE TRIGGER update_units_of_measure_updated_at
    BEFORE UPDATE ON units_of_measure
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS products (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(50)   NOT NULL,
    name            VARCHAR(200)  NOT NULL,
    description     TEXT,
    product_type    VARCHAR(20)   NOT NULL DEFAULT 'FINISHED' CHECK (product_type IN ('RAW_MATERIAL','WIP','FINISHED','SERVICE','CONSUMABLE')),
    category_id     BIGINT        REFERENCES product_categories(id),
    base_uom_id     BIGINT        NOT NULL REFERENCES units_of_measure(id),
    purchase_uom_id BIGINT        REFERENCES units_of_measure(id),
    sales_uom_id    BIGINT        REFERENCES units_of_measure(id),
    tax_code_id     BIGINT        REFERENCES tax_codes(id),
    cost_price      NUMERIC(18,4) NOT NULL DEFAULT 0,
    standard_price  NUMERIC(18,4) NOT NULL DEFAULT 0,
    min_stock_qty   NUMERIC(18,4) NOT NULL DEFAULT 0,
    reorder_qty     NUMERIC(18,4) NOT NULL DEFAULT 0,
    lead_time_days  INT           NOT NULL DEFAULT 0,
    is_purchased    BOOLEAN       NOT NULL DEFAULT true,
    is_sold         BOOLEAN       NOT NULL DEFAULT true,
    is_manufactured BOOLEAN       NOT NULL DEFAULT false,
    is_active       BOOLEAN       NOT NULL DEFAULT true,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,
    CONSTRAINT uidx_products_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_products_tenant_id    ON products(tenant_id);
CREATE INDEX IF NOT EXISTS idx_products_category_id  ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_product_type ON products(product_type);
CREATE INDEX IF NOT EXISTS idx_products_is_active    ON products(is_active);
CREATE INDEX IF NOT EXISTS idx_products_deleted_at   ON products(deleted_at);

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS product_variants (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id   BIGINT        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    code         VARCHAR(50)   NOT NULL,
    name         VARCHAR(200)  NOT NULL,
    attributes   JSONB,
    cost_price   NUMERIC(18,4) NOT NULL DEFAULT 0,
    sales_price  NUMERIC(18,4) NOT NULL DEFAULT 0,
    is_active    BOOLEAN       NOT NULL DEFAULT true,
    created_at   TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_product_variants_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX IF NOT EXISTS idx_product_variants_tenant_id  ON product_variants(tenant_id);

CREATE TRIGGER update_product_variants_updated_at
    BEFORE UPDATE ON product_variants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS product_prices (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id    BIGINT        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    party_id      BIGINT        REFERENCES parties(id) ON DELETE CASCADE,
    price_type    VARCHAR(20)   NOT NULL DEFAULT 'SALES' CHECK (price_type IN ('SALES','PURCHASE','SPECIAL')),
    currency_id   BIGINT        NOT NULL REFERENCES currencies(id),
    price         NUMERIC(18,4) NOT NULL,
    min_qty       NUMERIC(18,4) NOT NULL DEFAULT 0,
    effective_from DATE         NOT NULL,
    effective_to  DATE,
    is_active     BOOLEAN       NOT NULL DEFAULT true,
    created_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_prices_product_id ON product_prices(product_id);
CREATE INDEX IF NOT EXISTS idx_product_prices_tenant_id  ON product_prices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_product_prices_party_id   ON product_prices(party_id);

CREATE TRIGGER update_product_prices_updated_at
    BEFORE UPDATE ON product_prices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
