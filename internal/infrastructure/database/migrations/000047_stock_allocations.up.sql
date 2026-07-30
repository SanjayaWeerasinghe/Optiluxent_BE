-- Stock Allocations ledger. Reserves stock for pending downstream docs
-- (Sales Orders, Material Requests, Goods Issues, Goods Transfers) so the
-- next line-add against the same stock sees "available = quantity − active
-- allocations" instead of raw on-hand qty.
--
-- Scope key mirrors stock_balances' uniqueness key (tenant, product,
-- variant, warehouse, location, production) so refining segregation via
-- production_id works out of the box.
--
-- Lifecycle:
--   Reserve on draft-line create  → status = ACTIVE
--   Cancel/reject the source doc  → status = CANCELLED  (ReleaseByDoc)
--   Terminal action (DO confirm / GI confirm / GT send) → status = CONSUMED
--
-- One row per (source_type, source_id) — a single line writes exactly one
-- allocation row over its lifetime; updates flip the status column.

CREATE TABLE IF NOT EXISTS stock_allocations (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id    BIGINT        NOT NULL REFERENCES products(id),
    variant_id    BIGINT        REFERENCES product_variants(id),
    warehouse_id  BIGINT        NOT NULL REFERENCES warehouses(id),
    location_id   BIGINT        REFERENCES storage_locations(id),
    production_id BIGINT        REFERENCES production_orders(id),
    quantity      NUMERIC(18,4) NOT NULL CHECK (quantity > 0),
    source_type   VARCHAR(20)   NOT NULL
                                CHECK (source_type IN ('SO_LINE','MR_LINE','GI_LINE','GT_LINE')),
    source_id     BIGINT        NOT NULL,
    source_doc_id BIGINT        NOT NULL,
    status        VARCHAR(15)   NOT NULL DEFAULT 'ACTIVE'
                                CHECK (status IN ('ACTIVE','CONSUMED','CANCELLED')),
    notes         VARCHAR(500),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    -- One live allocation per source line. A line's row is created on first
    -- Reserve and flipped through statuses; no second row ever created for
    -- the same line.
    CONSTRAINT uidx_stock_allocations_source UNIQUE (source_type, source_id)
);

-- Fast SUM(active) queries per scope key (product + warehouse + variant +
-- location + production). Partial index skips the archived rows so the
-- hot path is always small.
CREATE INDEX IF NOT EXISTS idx_alloc_active_stock
    ON stock_allocations (tenant_id, product_id, warehouse_id, variant_id, location_id, production_id)
    WHERE status = 'ACTIVE';

-- Fast "release all allocations for this doc" queries used on cancel.
CREATE INDEX IF NOT EXISTS idx_alloc_by_doc
    ON stock_allocations (source_type, source_doc_id)
    WHERE status = 'ACTIVE';
