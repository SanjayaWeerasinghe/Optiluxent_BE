-- Reverse 000033: restore the old enum columns, back-fill from
-- document_types.system_key, then drop the new tables + FKs.

-- 1. Re-add old enum columns.
ALTER TABLE goods_receipts    ADD COLUMN IF NOT EXISTS grn_type      VARCHAR(30) NOT NULL DEFAULT 'WITH_PO';
ALTER TABLE goods_issues      ADD COLUMN IF NOT EXISTS issue_reason  VARCHAR(50) NOT NULL DEFAULT 'OTHER';
ALTER TABLE stock_adjustments ADD COLUMN IF NOT EXISTS adjust_reason VARCHAR(50) NOT NULL DEFAULT 'OTHER';
ALTER TABLE quality_checks    ADD COLUMN IF NOT EXISTS qc_type       VARCHAR(30) NOT NULL DEFAULT 'MATERIAL_QC';

-- 2. Back-fill from the seeded Types' system_key.
UPDATE goods_receipts g
SET grn_type = dt.system_key
FROM document_types dt WHERE dt.id = g.document_type_id AND dt.system_key IS NOT NULL;

UPDATE goods_issues gi
SET issue_reason = dt.system_key
FROM document_types dt WHERE dt.id = gi.document_type_id AND dt.system_key IS NOT NULL;

UPDATE stock_adjustments sa
SET adjust_reason = dt.system_key
FROM document_types dt WHERE dt.id = sa.document_type_id AND dt.system_key IS NOT NULL;

UPDATE quality_checks qc
SET qc_type = dt.system_key
FROM document_types dt WHERE dt.id = qc.document_type_id AND dt.system_key IS NOT NULL;

-- 3. Drop the document_type_id columns from every doc table.
ALTER TABLE purchase_requests  DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE purchase_orders    DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE purchase_invoices  DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE goods_receipts     DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE material_requests  DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE goods_transfers    DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE goods_issues       DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE stock_adjustments  DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE quality_checks     DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE sales_quotations   DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE sales_orders       DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE delivery_orders    DROP COLUMN IF EXISTS document_type_id;
ALTER TABLE sales_invoices     DROP COLUMN IF EXISTS document_type_id;

-- 4. Drop the new tables (order matters due to FKs).
DROP TABLE IF EXISTS document_field_values;
DROP TABLE IF EXISTS document_type_fields;
DROP TABLE IF EXISTS document_types;
