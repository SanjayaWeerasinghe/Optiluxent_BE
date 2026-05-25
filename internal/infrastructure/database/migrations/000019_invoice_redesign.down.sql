ALTER TABLE purchase_invoice_lines DROP COLUMN IF EXISTS supplier_invoice_no;
ALTER TABLE purchase_invoices      ALTER COLUMN po_id DROP NOT NULL;
ALTER TABLE purchase_invoices      DROP COLUMN IF EXISTS supplier_invoice_date;
ALTER TABLE purchase_invoices      DROP COLUMN IF EXISTS supplier_invoice_no;
