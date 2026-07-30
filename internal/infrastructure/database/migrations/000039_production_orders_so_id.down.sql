DROP INDEX IF EXISTS idx_production_orders_so_id;
ALTER TABLE production_orders DROP COLUMN IF EXISTS so_id;
