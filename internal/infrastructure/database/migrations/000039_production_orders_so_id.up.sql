-- Sales-order-linked production orders. Nullable — plan-based orders keep
-- so_id NULL and use plan_id instead. The FE Type picker decides which of
-- the two source fields the user is filling.
ALTER TABLE production_orders
    ADD COLUMN IF NOT EXISTS so_id BIGINT REFERENCES sales_orders(id);
CREATE INDEX IF NOT EXISTS idx_production_orders_so_id ON production_orders(so_id);
