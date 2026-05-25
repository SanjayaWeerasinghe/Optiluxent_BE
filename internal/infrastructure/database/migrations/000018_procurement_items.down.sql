ALTER TABLE goods_receipt_lines      DROP COLUMN IF EXISTS transfer_ratio;
ALTER TABLE purchase_order_lines     DROP COLUMN IF EXISTS transfer_ratio;
ALTER TABLE purchase_request_lines   DROP COLUMN IF EXISTS currency_id;
ALTER TABLE purchase_request_lines   DROP COLUMN IF EXISTS transfer_ratio;
