DROP TABLE IF EXISTS production_plan_inputs;

DROP INDEX IF EXISTS idx_production_plans_so_id;
ALTER TABLE production_plans
    DROP COLUMN IF EXISTS so_id,
    DROP COLUMN IF EXISTS document_type_id;

DROP INDEX IF EXISTS idx_stock_balances_production;
ALTER TABLE stock_balances
    DROP CONSTRAINT IF EXISTS uidx_stock_balances;
ALTER TABLE stock_balances
    ADD CONSTRAINT uidx_stock_balances
        UNIQUE NULLS NOT DISTINCT
        (tenant_id, product_id, variant_id, warehouse_id, location_id);
ALTER TABLE stock_balances
    DROP COLUMN IF EXISTS production_id;

DROP INDEX IF EXISTS idx_stock_ledger_production;
ALTER TABLE stock_ledger
    DROP COLUMN IF EXISTS production_id;
