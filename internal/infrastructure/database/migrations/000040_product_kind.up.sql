-- Product kind — separates our own goods from customer-owned refining intake
-- and bill-only service items. Defaults every existing row to PRODUCT so the
-- current behaviour is unchanged.
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS kind VARCHAR(20) NOT NULL DEFAULT 'PRODUCT'
        CHECK (kind IN ('PRODUCT', 'SERVICE', 'REFINING_INTAKE'));

CREATE INDEX IF NOT EXISTS idx_products_kind ON products(kind);

-- Seed the single shared product used for customer crude-oil intake.
-- Only for tenants that have at least one unit of measure defined — a tenant
-- with no master data isn't ready to receive oil anyway. Skipped tenants can
-- create the product manually via master-data once they've set up UOMs.
INSERT INTO products (tenant_id, code, name, kind, product_type, base_uom_id,
                      is_purchased, is_sold, is_manufactured, is_active)
SELECT t.id, 'CRUDE-REFINE', 'Crude Oil for Refining', 'REFINING_INTAKE', 'RAW_MATERIAL',
       u.id,
       FALSE, FALSE, FALSE, TRUE
FROM tenants t
JOIN LATERAL (
    SELECT id FROM units_of_measure WHERE tenant_id = t.id ORDER BY id LIMIT 1
) u ON TRUE
WHERE NOT EXISTS (SELECT 1 FROM products WHERE tenant_id = t.id AND code = 'CRUDE-REFINE');
