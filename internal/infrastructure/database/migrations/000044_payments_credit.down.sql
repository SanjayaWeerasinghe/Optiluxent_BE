DELETE FROM document_sequences WHERE document_type IN ('PAYMENT_IN','PAYMENT_OUT');
ALTER TABLE parties DROP COLUMN IF EXISTS credit_type;
DROP TABLE IF EXISTS finance_settings;
DROP TABLE IF EXISTS payments;
