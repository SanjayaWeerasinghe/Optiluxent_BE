-- Extend procurement line tables with transfer_ratio and (for PR) per-line currency

ALTER TABLE purchase_request_lines
    ADD COLUMN IF NOT EXISTS transfer_ratio NUMERIC(18,6) NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS currency_id    BIGINT REFERENCES currencies(id);

ALTER TABLE purchase_order_lines
    ADD COLUMN IF NOT EXISTS transfer_ratio NUMERIC(18,6) NOT NULL DEFAULT 1;

ALTER TABLE goods_receipt_lines
    ADD COLUMN IF NOT EXISTS transfer_ratio NUMERIC(18,6) NOT NULL DEFAULT 1;
