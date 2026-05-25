-- Master Data: Inventory (warehouses, storage_locations, stock_ledger)

CREATE TABLE IF NOT EXISTS warehouses (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code         VARCHAR(20)  NOT NULL,
    name         VARCHAR(200) NOT NULL,
    address      TEXT,
    is_active    BOOLEAN      NOT NULL DEFAULT true,
    created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMP,
    CONSTRAINT uidx_warehouses_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_warehouses_tenant_id  ON warehouses(tenant_id);
CREATE INDEX IF NOT EXISTS idx_warehouses_deleted_at ON warehouses(deleted_at);

CREATE TRIGGER update_warehouses_updated_at
    BEFORE UPDATE ON warehouses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS storage_locations (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    warehouse_id BIGINT       NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    code         VARCHAR(20)  NOT NULL,
    name         VARCHAR(200) NOT NULL,
    location_type VARCHAR(20) NOT NULL DEFAULT 'STORAGE' CHECK (location_type IN ('STORAGE','RECEIVING','SHIPPING','QUALITY','DAMAGED')),
    is_active    BOOLEAN      NOT NULL DEFAULT true,
    created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_storage_locations_warehouse_code UNIQUE (warehouse_id, code)
);

CREATE INDEX IF NOT EXISTS idx_storage_locations_warehouse_id ON storage_locations(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_storage_locations_tenant_id    ON storage_locations(tenant_id);

CREATE TRIGGER update_storage_locations_updated_at
    BEFORE UPDATE ON storage_locations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS stock_ledger (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id       BIGINT        NOT NULL REFERENCES products(id),
    variant_id       BIGINT        REFERENCES product_variants(id),
    warehouse_id     BIGINT        NOT NULL REFERENCES warehouses(id),
    location_id      BIGINT        REFERENCES storage_locations(id),
    transaction_type VARCHAR(30)   NOT NULL CHECK (transaction_type IN ('OPENING','PURCHASE','SALES','PRODUCTION','ADJUSTMENT','TRANSFER_IN','TRANSFER_OUT','RETURN_IN','RETURN_OUT')),
    reference_type   VARCHAR(50),
    reference_id     BIGINT,
    quantity         NUMERIC(18,4) NOT NULL,
    unit_cost        NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost       NUMERIC(18,4) NOT NULL DEFAULT 0,
    transaction_date DATE          NOT NULL,
    notes            TEXT,
    created_at       TIMESTAMP     NOT NULL DEFAULT NOW(),
    created_by       BIGINT
);

CREATE INDEX IF NOT EXISTS idx_stock_ledger_tenant_id    ON stock_ledger(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_product_id   ON stock_ledger(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_warehouse_id ON stock_ledger(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_date         ON stock_ledger(transaction_date);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_ref          ON stock_ledger(reference_type, reference_id);
