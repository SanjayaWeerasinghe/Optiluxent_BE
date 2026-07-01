ALTER TABLE material_request_lines  DROP COLUMN created_at, DROP COLUMN updated_at;
ALTER TABLE goods_transfer_lines    DROP COLUMN created_at, DROP COLUMN updated_at;
ALTER TABLE stock_adjustment_lines  DROP COLUMN created_at, DROP COLUMN updated_at;
ALTER TABLE goods_issue_lines       DROP COLUMN created_at, DROP COLUMN updated_at;
