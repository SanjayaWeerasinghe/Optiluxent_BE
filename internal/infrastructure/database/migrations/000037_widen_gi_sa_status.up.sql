-- Allow GI + SA to reach CANCELLED. Draft-cancel is a common user action
-- and both tables' CHECK constraints previously only permitted
-- DRAFT / CONFIRMED, causing the new Cancel workflow to fail.

ALTER TABLE goods_issues       DROP CONSTRAINT IF EXISTS goods_issues_status_check;
ALTER TABLE goods_issues       ADD  CONSTRAINT goods_issues_status_check
    CHECK (status IN ('DRAFT', 'CONFIRMED', 'CANCELLED'));

ALTER TABLE stock_adjustments  DROP CONSTRAINT IF EXISTS stock_adjustments_status_check;
ALTER TABLE stock_adjustments  ADD  CONSTRAINT stock_adjustments_status_check
    CHECK (status IN ('DRAFT', 'CONFIRMED', 'CANCELLED'));
