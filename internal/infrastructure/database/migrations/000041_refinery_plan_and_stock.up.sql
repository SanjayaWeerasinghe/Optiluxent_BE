-- Refinery Service groundwork:
--   1. stock_ledger + stock_balances gain a production_id stamp so
--      REFINING_INTAKE stock is segregated per production. NULL means the
--      row is unscoped general stock (existing behaviour).
--   2. production_plans gain so_id + document_type_id.
--   3. production_plan_inputs holds the chemicals the Plan calls for; they
--      copy into MO.ProductionResource when the Production is created.

-- 1a. stock_ledger stamp
ALTER TABLE stock_ledger
    ADD COLUMN IF NOT EXISTS production_id BIGINT REFERENCES production_orders(id);
CREATE INDEX IF NOT EXISTS idx_stock_ledger_production ON stock_ledger(production_id)
    WHERE production_id IS NOT NULL;

-- 1b. stock_balances stamp — needs to be part of the uniqueness so we can
-- carry a separate balance per production for REFINING_INTAKE stock.
ALTER TABLE stock_balances
    ADD COLUMN IF NOT EXISTS production_id BIGINT REFERENCES production_orders(id);
ALTER TABLE stock_balances
    DROP CONSTRAINT IF EXISTS uidx_stock_balances;
ALTER TABLE stock_balances
    ADD CONSTRAINT uidx_stock_balances
        UNIQUE NULLS NOT DISTINCT
        (tenant_id, product_id, variant_id, warehouse_id, location_id, production_id);
CREATE INDEX IF NOT EXISTS idx_stock_balances_production ON stock_balances(production_id)
    WHERE production_id IS NOT NULL;

-- 2. Production Plan gains SO link + Type classification.
ALTER TABLE production_plans
    ADD COLUMN IF NOT EXISTS so_id            BIGINT REFERENCES sales_orders(id),
    ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
CREATE INDEX IF NOT EXISTS idx_production_plans_so_id ON production_plans(so_id);

-- 3. Anticipated inputs for the Plan. Copied into ProductionResource
-- rows when a Production is manually created from the released Plan.
CREATE TABLE IF NOT EXISTS production_plan_inputs (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id      BIGINT       NOT NULL REFERENCES production_plans(id) ON DELETE CASCADE,
    line_number  INT          NOT NULL DEFAULT 1,
    product_id   BIGINT       NOT NULL REFERENCES products(id),
    quantity     NUMERIC(18,4) NOT NULL DEFAULT 0,
    uom_id       BIGINT       NOT NULL REFERENCES units_of_measure(id),
    notes        VARCHAR(500),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pli_plan ON production_plan_inputs(plan_id);
CREATE INDEX IF NOT EXISTS idx_pli_tenant ON production_plan_inputs(tenant_id);
