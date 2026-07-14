-- Direct link from a Goods Issue to its source Material Request.
-- Nullable — pre-existing issues stay untouched. Used by the FE to
-- auto-populate GI lines from an approved MR on Initiate.
ALTER TABLE goods_issues
    ADD COLUMN IF NOT EXISTS mr_id BIGINT REFERENCES material_requests(id);
CREATE INDEX IF NOT EXISTS idx_gi_mr_id ON goods_issues(mr_id);
