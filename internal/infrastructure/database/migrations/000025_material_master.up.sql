-- Comprehensive Material Master
CREATE TABLE IF NOT EXISTS mm_materials (
    id                      BIGSERIAL    PRIMARY KEY,
    tenant_id               BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code                    VARCHAR(50)  NOT NULL,
    name                    VARCHAR(200) NOT NULL,
    material_type           VARCHAR(20)  NOT NULL DEFAULT 'RAW_MATERIAL'
                                CHECK (material_type IN ('RAW_MATERIAL','SEMI_FINISHED','SERVICE')),
    -- Basic
    color                   VARCHAR(100),
    category_id             BIGINT       REFERENCES product_categories(id),
    article_code            VARCHAR(100),
    ref2                    VARCHAR(100),
    barcode                 VARCHAR(100),
    description             TEXT,
    -- Purchasing
    purchasing_uom_id       BIGINT       REFERENCES units_of_measure(id),
    under_delivery_pct      NUMERIC(5,2) NOT NULL DEFAULT 0,
    over_delivery_pct       NUMERIC(5,2) NOT NULL DEFAULT 0,
    -- Manufacturing
    production_uom_id       BIGINT       REFERENCES units_of_measure(id),
    reorder_qty_level       NUMERIC(18,4) NOT NULL DEFAULT 0,
    safety_level            NUMERIC(18,4) NOT NULL DEFAULT 0,
    production_days         INTEGER      NOT NULL DEFAULT 0,
    delivery_days           INTEGER      NOT NULL DEFAULT 0,
    grn_days                INTEGER      NOT NULL DEFAULT 0,
    procurement_repeat_days INTEGER      NOT NULL DEFAULT 0,
    -- Warehouse
    stocking_uom_id         BIGINT       REFERENCES units_of_measure(id),
    stock_removal           VARCHAR(10)  NOT NULL DEFAULT 'FIFO'
                                CHECK (stock_removal IN ('FIFO','FEFO','LIFO')),
    storage_main            BOOLEAN      NOT NULL DEFAULT FALSE,
    storage_damaged         BOOLEAN      NOT NULL DEFAULT FALSE,
    storage_hold            BOOLEAN      NOT NULL DEFAULT FALSE,
    batch_process           BOOLEAN      NOT NULL DEFAULT FALSE,
    production_date_check   BOOLEAN      NOT NULL DEFAULT FALSE,
    expiry_date_check       BOOLEAN      NOT NULL DEFAULT FALSE,
    qc_check                BOOLEAN      NOT NULL DEFAULT FALSE,
    grn_with_po_uom_image   BOOLEAN      NOT NULL DEFAULT FALSE,

    is_active               BOOLEAN      NOT NULL DEFAULT TRUE,
    notes                   TEXT,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_mm_materials UNIQUE (tenant_id, code)
);

-- Vendor pricing lines (many per material)
CREATE TABLE IF NOT EXISTS mm_vendors (
    id              BIGSERIAL    PRIMARY KEY,
    material_id     BIGINT       NOT NULL REFERENCES mm_materials(id) ON DELETE CASCADE,
    tenant_id       BIGINT       NOT NULL,
    vendor_id       BIGINT       REFERENCES parties(id),
    article_no      VARCHAR(100),
    delivery_days   INTEGER      NOT NULL DEFAULT 0,
    cost            NUMERIC(18,4),
    currency_id     BIGINT       REFERENCES currencies(id),
    projected_price NUMERIC(18,4),
    moq             NUMERIC(18,4),
    is_default      BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- UOM conversion ratios (many per material)
CREATE TABLE IF NOT EXISTS mm_measurements (
    id               BIGSERIAL    PRIMARY KEY,
    material_id      BIGINT       NOT NULL REFERENCES mm_materials(id) ON DELETE CASCADE,
    tenant_id        BIGINT       NOT NULL,
    base_uom_id      BIGINT       NOT NULL REFERENCES units_of_measure(id),
    target_uom_id    BIGINT       NOT NULL REFERENCES units_of_measure(id),
    conversion_ratio NUMERIC(18,6) NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_mm_materials_tenant ON mm_materials(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mm_materials_type   ON mm_materials(tenant_id, material_type);
CREATE INDEX IF NOT EXISTS idx_mm_vendors_mat      ON mm_vendors(material_id);
CREATE INDEX IF NOT EXISTS idx_mm_measurements_mat ON mm_measurements(material_id);
