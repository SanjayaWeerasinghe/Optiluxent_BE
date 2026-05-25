-- Purchase Invoice: po_id becomes required; add supplier invoice tracking;
-- grn_id is retired from the header (items track GRN at line level via grn_line_id).

ALTER TABLE purchase_invoices
    ADD COLUMN IF NOT EXISTS supplier_invoice_no   VARCHAR(100),
    ADD COLUMN IF NOT EXISTS supplier_invoice_date DATE;

-- Enforce po_id NOT NULL going forward (backfill existing rows with a sentinel if needed)
UPDATE purchase_invoices SET po_id = 0 WHERE po_id IS NULL;
ALTER TABLE purchase_invoices ALTER COLUMN po_id SET NOT NULL;

-- grn_id on the header is no longer meaningful; keep column for historical data but stop using it
-- (a DROP COLUMN can be done in a future cleanup migration)

ALTER TABLE purchase_invoice_lines
    ADD COLUMN IF NOT EXISTS supplier_invoice_no VARCHAR(100);
