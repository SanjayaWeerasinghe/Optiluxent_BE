-- Rollback inventory module migration

DELETE FROM document_sequences WHERE document_type IN
    ('MATERIAL_REQUEST','GOODS_TRANSFER','GOODS_ISSUE','STOCK_ADJUSTMENT','QUALITY_CHECK');

DROP TABLE IF EXISTS quality_check_lines;
DROP TABLE IF EXISTS quality_checks;
DROP TABLE IF EXISTS stock_adjustment_lines;
DROP TABLE IF EXISTS stock_adjustments;
DROP TABLE IF EXISTS goods_issue_lines;
DROP TABLE IF EXISTS goods_issues;
DROP TABLE IF EXISTS goods_transfer_lines;
DROP TABLE IF EXISTS goods_transfers;
DROP TABLE IF EXISTS material_request_lines;
DROP TABLE IF EXISTS material_requests;
DROP TABLE IF EXISTS stock_balances;
