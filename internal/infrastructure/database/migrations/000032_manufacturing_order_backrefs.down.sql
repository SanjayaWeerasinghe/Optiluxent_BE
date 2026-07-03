DROP TABLE IF EXISTS uom_conversions;

ALTER TABLE products
    DROP COLUMN IF EXISTS production_uom_id,
    DROP COLUMN IF EXISTS stock_uom_id;

DROP INDEX IF EXISTS idx_grn_mo_id;
DROP INDEX IF EXISTS idx_gt_mo_id;
DROP INDEX IF EXISTS idx_mr_mo_id;

ALTER TABLE goods_receipts    DROP COLUMN IF EXISTS mo_id;
ALTER TABLE goods_transfers   DROP COLUMN IF EXISTS mo_id;
ALTER TABLE material_requests DROP COLUMN IF EXISTS mo_id;
