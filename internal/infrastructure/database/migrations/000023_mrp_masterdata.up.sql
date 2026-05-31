-- MRP (Material Resource Planning) master data
-- Material Master: planning parameters for each product

CREATE TABLE IF NOT EXISTS mrp_material_master (
    id                   BIGSERIAL    PRIMARY KEY,
    tenant_id            BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id           BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,

    -- Planning type
    mrp_type             VARCHAR(20)  NOT NULL DEFAULT 'MRP'
                             CHECK (mrp_type IN ('MRP','REORDER_POINT','MANUAL','NO_PLANNING')),
    procurement_type     VARCHAR(20)  NOT NULL DEFAULT 'EXTERNAL'
                             CHECK (procurement_type IN ('EXTERNAL','IN_HOUSE','BOTH')),

    -- Lead times (days)
    purchase_lead_time   INTEGER      NOT NULL DEFAULT 0,
    production_lead_time INTEGER      NOT NULL DEFAULT 0,

    -- Stock levels
    safety_stock         NUMERIC(18,4) NOT NULL DEFAULT 0,
    reorder_point        NUMERIC(18,4) NOT NULL DEFAULT 0,
    min_order_qty        NUMERIC(18,4) NOT NULL DEFAULT 0,
    max_order_qty        NUMERIC(18,4),

    -- Lot sizing
    lot_size_type        VARCHAR(20)  NOT NULL DEFAULT 'LOT_FOR_LOT'
                             CHECK (lot_size_type IN ('LOT_FOR_LOT','FIXED_LOT','ECONOMIC_ORDER')),
    fixed_lot_size       NUMERIC(18,4),

    -- Classification
    abc_class            CHAR(1)      CHECK (abc_class IN ('A','B','C')),
    shelf_life_days      INTEGER,

    is_active            BOOLEAN      NOT NULL DEFAULT TRUE,
    notes                TEXT,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_mrp_material_master UNIQUE (tenant_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_mrp_material_master_tenant  ON mrp_material_master(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mrp_material_master_product ON mrp_material_master(product_id);
