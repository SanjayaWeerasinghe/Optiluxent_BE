ALTER TABLE sales_orders DROP COLUMN IF EXISTS sq_id;
DROP TABLE IF EXISTS sales_quotation_lines;
DROP TABLE IF EXISTS sales_quotations;
DELETE FROM document_sequences WHERE document_type = 'SALES_QUOTATION';
