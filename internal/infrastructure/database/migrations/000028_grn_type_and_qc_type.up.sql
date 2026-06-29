-- Add grn_type to goods_receipts.
-- WITH_PO and WITHOUT_PO trigger Material QC.
-- CUSTOMER_RETURN and PRODUCTION_RETURN bypass QC.
ALTER TABLE goods_receipts
    ADD COLUMN grn_type VARCHAR(30) NOT NULL DEFAULT 'WITH_PO';

-- Backfill existing rows: po_id IS NOT NULL → WITH_PO, otherwise WITHOUT_PO.
UPDATE goods_receipts
   SET grn_type = CASE WHEN po_id IS NULL THEN 'WITHOUT_PO' ELSE 'WITH_PO' END
 WHERE grn_type = 'WITH_PO';

-- Add qc_type to quality_checks.
-- MATERIAL_QC = inbound goods (GRN). PRODUCT_QC = production output.
ALTER TABLE quality_checks
    ADD COLUMN qc_type VARCHAR(30) NOT NULL DEFAULT 'MATERIAL_QC';
