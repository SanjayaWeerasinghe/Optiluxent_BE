DELETE FROM document_sequences WHERE document_type IN ('PURCHASE_REQUEST','PURCHASE_ORDER','GOODS_RECEIPT','PURCHASE_INVOICE');
DROP TABLE IF EXISTS purchase_invoice_lines;
DROP TABLE IF EXISTS purchase_invoices;
DROP TABLE IF EXISTS goods_receipt_lines;
DROP TABLE IF EXISTS goods_receipts;
DROP TABLE IF EXISTS purchase_order_lines;
DROP TABLE IF EXISTS purchase_orders;
DROP TABLE IF EXISTS purchase_request_lines;
DROP TABLE IF EXISTS purchase_requests;
