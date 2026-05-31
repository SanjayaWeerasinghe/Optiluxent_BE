-- Drop old tables in reverse dependency order
DROP TABLE IF EXISTS mm_measurements CASCADE;
DROP TABLE IF EXISTS mm_vendors CASCADE;
DROP TABLE IF EXISTS mm_warehouse CASCADE;
DROP TABLE IF EXISTS mm_manufacturing CASCADE;
DROP TABLE IF EXISTS mm_purchasing CASCADE;
DROP TABLE IF EXISTS mm_materials CASCADE;
DROP TABLE IF EXISTS mm_categories CASCADE;

-- 1. Categories (standalone, hierarchical)
CREATE TABLE mm_categories (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parent_id   BIGINT REFERENCES mm_categories(id) ON DELETE SET NULL,
    code        VARCHAR(50)  NOT NULL,
    name        VARCHAR(200) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_mm_categories UNIQUE (tenant_id, code)
);
CREATE INDEX idx_mm_categories_tenant  ON mm_categories(tenant_id);
CREATE INDEX idx_mm_categories_parent  ON mm_categories(parent_id);

-- 2. Core material identity
CREATE TABLE mm_materials (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code          VARCHAR(50)  NOT NULL,
    name          VARCHAR(200) NOT NULL,
    material_type VARCHAR(20)  NOT NULL DEFAULT 'RAW_MATERIAL'
                      CHECK (material_type IN ('RAW_MATERIAL','SEMI_FINISHED','SERVICE')),
    color         VARCHAR(100),
    category_id   BIGINT REFERENCES mm_categories(id) ON DELETE SET NULL,
    article_code  VARCHAR(100),
    ref2          VARCHAR(100),
    barcode       VARCHAR(100),
    description   TEXT,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    notes         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_mm_materials UNIQUE (tenant_id, code)
);
CREATE INDEX idx_mm_materials_tenant ON mm_materials(tenant_id);
CREATE INDEX idx_mm_materials_type   ON mm_materials(tenant_id, material_type);

-- 3. Purchasing (1:1)
CREATE TABLE mm_purchasing (
    id                BIGSERIAL PRIMARY KEY,
    material_id       BIGINT NOT NULL REFERENCES mm_materials(id) ON DELETE CASCADE,
    tenant_id         BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    purchasing_uom_id BIGINT REFERENCES units_of_measure(id) ON DELETE SET NULL,
    under_delivery_pct NUMERIC(10,4) NOT NULL DEFAULT 0,
    over_delivery_pct  NUMERIC(10,4) NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_mm_purchasing UNIQUE (material_id)
);

-- 4. Manufacturing (1:1)
CREATE TABLE mm_manufacturing (
    id                      BIGSERIAL PRIMARY KEY,
    material_id             BIGINT NOT NULL REFERENCES mm_materials(id) ON DELETE CASCADE,
    tenant_id               BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    production_uom_id       BIGINT REFERENCES units_of_measure(id) ON DELETE SET NULL,
    reorder_qty_level       NUMERIC(18,4) NOT NULL DEFAULT 0,
    safety_level            NUMERIC(18,4) NOT NULL DEFAULT 0,
    production_days         INT NOT NULL DEFAULT 0,
    delivery_days           INT NOT NULL DEFAULT 0,
    grn_days                INT NOT NULL DEFAULT 0,
    procurement_repeat_days INT NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_mm_manufacturing UNIQUE (material_id)
);

-- 5. Warehouse (1:1)
CREATE TABLE mm_warehouse (
    id                    BIGSERIAL PRIMARY KEY,
    material_id           BIGINT NOT NULL REFERENCES mm_materials(id) ON DELETE CASCADE,
    tenant_id             BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    stocking_uom_id       BIGINT REFERENCES units_of_measure(id) ON DELETE SET NULL,
    stock_removal         VARCHAR(10) NOT NULL DEFAULT 'FIFO'
                              CHECK (stock_removal IN ('FIFO','FEFO','LIFO')),
    storage_main          BOOLEAN NOT NULL DEFAULT FALSE,
    storage_damaged       BOOLEAN NOT NULL DEFAULT FALSE,
    storage_hold          BOOLEAN NOT NULL DEFAULT FALSE,
    batch_process         BOOLEAN NOT NULL DEFAULT FALSE,
    production_date_check BOOLEAN NOT NULL DEFAULT FALSE,
    expiry_date_check     BOOLEAN NOT NULL DEFAULT FALSE,
    qc_check              BOOLEAN NOT NULL DEFAULT FALSE,
    grn_with_po_uom_image BOOLEAN NOT NULL DEFAULT FALSE,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_mm_warehouse UNIQUE (material_id)
);

-- 6. Vendors (M:1)
CREATE TABLE mm_vendors (
    id              BIGSERIAL PRIMARY KEY,
    material_id     BIGINT NOT NULL REFERENCES mm_materials(id) ON DELETE CASCADE,
    tenant_id       BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    vendor_id       BIGINT REFERENCES parties(id) ON DELETE SET NULL,
    article_no      VARCHAR(100),
    delivery_days   INT NOT NULL DEFAULT 0,
    cost            NUMERIC(18,4),
    currency_id     BIGINT REFERENCES currencies(id) ON DELETE SET NULL,
    projected_price NUMERIC(18,4),
    moq             NUMERIC(18,4),
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_mm_vendors_material ON mm_vendors(material_id);

-- 7. Measurements (M:1)
CREATE TABLE mm_measurements (
    id               BIGSERIAL PRIMARY KEY,
    material_id      BIGINT NOT NULL REFERENCES mm_materials(id) ON DELETE CASCADE,
    tenant_id        BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    base_uom_id      BIGINT NOT NULL REFERENCES units_of_measure(id) ON DELETE RESTRICT,
    target_uom_id    BIGINT NOT NULL REFERENCES units_of_measure(id) ON DELETE RESTRICT,
    conversion_ratio NUMERIC(18,6) NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_mm_measurements_material ON mm_measurements(material_id);
