-- Cross-module back-references from downstream inventory + procurement docs to
-- the Manufacturing Order that spawned them. Used by the MO dashboard to
-- aggregate requested / issued / produced quantities against one order.
--
-- Nullable — pre-existing docs stay untouched. All are simple FKs to
-- production_orders(id); goods_issues already has a generic
-- reference_type/reference_id pair which we standardise as
-- ('PRODUCTION_ORDER', <mo.id>). Adding an explicit mo_id column too would
-- duplicate; leave GI alone and read it via reference_id.

ALTER TABLE material_requests
    ADD COLUMN mo_id BIGINT REFERENCES production_orders(id);
CREATE INDEX IF NOT EXISTS idx_mr_mo_id ON material_requests(mo_id);

ALTER TABLE goods_transfers
    ADD COLUMN mo_id BIGINT REFERENCES production_orders(id);
CREATE INDEX IF NOT EXISTS idx_gt_mo_id ON goods_transfers(mo_id);

ALTER TABLE goods_receipts
    ADD COLUMN mo_id BIGINT REFERENCES production_orders(id);
CREATE INDEX IF NOT EXISTS idx_grn_mo_id ON goods_receipts(mo_id);

-- Multi-role UOM support for Materials + Products.
--
-- Each master carries three (nullable) role FKs — procurement / stock /
-- production. Different flows read the appropriate one:
--   Purchasing docs → procurement_uom_id (materials only)
--   Stock ledger    → stock_uom_id
--   Production      → production_uom_id
--
-- The Product/Material Master forms populate these at edit time. Coconut oil
-- example — Product "Refined Coconut Oil 500ml Bottle":
--   procurement_uom_id = NULL  (not purchased directly)
--   stock_uom_id       = BOT   (bottles in the warehouse)
--   production_uom_id  = ML    (measured in millilitres on the shop floor)

-- Products get two extra role FKs; procurement isn't a role for products
-- since they aren't purchased as raw materials (they're the output).
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS stock_uom_id      BIGINT REFERENCES units_of_measure(id),
    ADD COLUMN IF NOT EXISTS production_uom_id BIGINT REFERENCES units_of_measure(id);

-- Materials already carry all three role UOMs split across sub-tables:
--   mm_purchasing.purchasing_uom_id
--   mm_warehouse.stocking_uom_id
--   mm_manufacturing.production_uom_id
-- Nothing to add on mm_materials.

-- Cross-UOM conversion rows anchored on Product OR Material.
-- Exactly one of product_id / material_id is set; the other stays NULL.
-- Rows are role-agnostic: they define "how many `to_uom` units make one
-- `from_uom` unit for this item". The dashboard reads the specific pair it
-- needs (e.g. BOT → L for a 500ml product = 0.5).
CREATE TABLE IF NOT EXISTS uom_conversions (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id   BIGINT        REFERENCES products(id)      ON DELETE CASCADE,
    material_id  BIGINT        REFERENCES mm_materials(id)  ON DELETE CASCADE,
    from_uom_id  BIGINT        NOT NULL REFERENCES units_of_measure(id),
    to_uom_id    BIGINT        NOT NULL REFERENCES units_of_measure(id),
    ratio        NUMERIC(18,6) NOT NULL,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CHECK ((product_id IS NULL) <> (material_id IS NULL)),
    CONSTRAINT uidx_uom_conv UNIQUE (tenant_id, product_id, material_id, from_uom_id, to_uom_id)
);
CREATE INDEX IF NOT EXISTS idx_uom_conv_product  ON uom_conversions(product_id)  WHERE product_id  IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_uom_conv_material ON uom_conversions(material_id) WHERE material_id IS NOT NULL;
