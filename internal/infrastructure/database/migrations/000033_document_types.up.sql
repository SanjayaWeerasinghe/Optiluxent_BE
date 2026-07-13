-- Document Types + custom fields — unified typing system.
--
-- Replaces the four ad-hoc string enums we ship today (grn_type,
-- issue_reason, adjust_reason, qc_type) with a generic per-model Type
-- system. Users can add their own Types from Master Data, each with a
-- list of picker "fields" (each field pointing at a picker of a specific
-- entity kind: PR, PO, Customer, Supplier, Product, etc.).
--
-- The rows we seed from the old enums carry `system_key`; the BE
-- branches on that value to keep the existing behavioural workflows
-- (auto-QC on GRN confirm, MO produced_qty bump on PRODUCTION_OUTPUT,
-- MR issued_qty bump on production goods-issue).

-- ── 1. Type catalogue ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS document_types (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    model        VARCHAR(20)  NOT NULL,
    code         VARCHAR(50)  NOT NULL,
    name         VARCHAR(200) NOT NULL,
    system_key   VARCHAR(50),
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_doctype UNIQUE (tenant_id, model, code)
);
CREATE INDEX IF NOT EXISTS idx_doctype_model_active ON document_types(tenant_id, model, is_active);

-- ── 2. Field slots hanging off a Type ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS document_type_fields (
    id                BIGSERIAL PRIMARY KEY,
    tenant_id         BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    document_type_id  BIGINT       NOT NULL REFERENCES document_types(id) ON DELETE CASCADE,
    code              VARCHAR(50)  NOT NULL,
    label             VARCHAR(200) NOT NULL,
    kind              VARCHAR(20)  NOT NULL,
    is_required       BOOLEAN      NOT NULL DEFAULT FALSE,
    display_order     INT          NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_dtf UNIQUE (document_type_id, code)
);

-- ── 3. Per-document field values ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS document_field_values (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT      NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    doc_kind     VARCHAR(20) NOT NULL,
    doc_id       BIGINT      NOT NULL,
    field_id     BIGINT      NOT NULL REFERENCES document_type_fields(id) ON DELETE CASCADE,
    ref_id       BIGINT      NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_dfv UNIQUE (doc_kind, doc_id, field_id)
);
CREATE INDEX IF NOT EXISTS idx_dfv_doc ON document_field_values(doc_kind, doc_id);

-- ── 4. Add document_type_id to every transactional document ──────────────────
ALTER TABLE purchase_requests  ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE purchase_orders    ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE purchase_invoices  ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE goods_receipts     ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE material_requests  ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE goods_transfers    ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE goods_issues       ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE stock_adjustments  ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE quality_checks     ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE sales_quotations   ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE sales_orders       ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE delivery_orders    ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);
ALTER TABLE sales_invoices     ADD COLUMN IF NOT EXISTS document_type_id BIGINT REFERENCES document_types(id);

-- ── 5. Seed system Types from the existing enum values ──────────────────────
--
-- One row per (tenant, enum_value) so multi-tenant tenants each keep an
-- isolated catalogue. code == name == system_key on seeded rows; users
-- can rename them later without breaking the behavioural switch (which
-- keys on system_key).

-- GRN — grn_type has a default of WITH_PO, so every tenant needs the full set.
INSERT INTO document_types (tenant_id, model, code, name, system_key)
SELECT DISTINCT t.id, 'GRN', v.value, v.value, v.value
FROM tenants t
CROSS JOIN (VALUES ('WITH_PO'), ('WITHOUT_PO'), ('CUSTOMER_RETURN'), ('PRODUCTION_RETURN'), ('PRODUCTION_OUTPUT')) v(value)
ON CONFLICT (tenant_id, model, code) DO NOTHING;

INSERT INTO document_types (tenant_id, model, code, name, system_key)
SELECT DISTINCT t.id, 'GI', v.value, v.value, v.value
FROM tenants t
CROSS JOIN (VALUES ('PRODUCTION'), ('SALE'), ('EXPENSE'), ('DAMAGE'), ('ADJUSTMENT'), ('QC_REJECTION'), ('OTHER')) v(value)
ON CONFLICT (tenant_id, model, code) DO NOTHING;

INSERT INTO document_types (tenant_id, model, code, name, system_key)
SELECT DISTINCT t.id, 'SA', v.value, v.value, v.value
FROM tenants t
CROSS JOIN (VALUES ('STOCKTAKE'), ('DAMAGE'), ('EXPIRY'), ('WRITE_OFF'), ('CORRECTION'), ('OTHER')) v(value)
ON CONFLICT (tenant_id, model, code) DO NOTHING;

INSERT INTO document_types (tenant_id, model, code, name, system_key)
SELECT DISTINCT t.id, 'QC', v.value, v.value, v.value
FROM tenants t
CROSS JOIN (VALUES ('MATERIAL_QC'), ('PRODUCT_QC')) v(value)
ON CONFLICT (tenant_id, model, code) DO NOTHING;

-- ── 6. Back-fill document_type_id on existing rows ──────────────────────────
UPDATE goods_receipts g
SET document_type_id = dt.id
FROM document_types dt
WHERE dt.tenant_id = g.tenant_id
  AND dt.model     = 'GRN'
  AND dt.system_key = g.grn_type
  AND g.document_type_id IS NULL;

UPDATE goods_issues gi
SET document_type_id = dt.id
FROM document_types dt
WHERE dt.tenant_id = gi.tenant_id
  AND dt.model     = 'GI'
  AND dt.system_key = gi.issue_reason
  AND gi.document_type_id IS NULL;

UPDATE stock_adjustments sa
SET document_type_id = dt.id
FROM document_types dt
WHERE dt.tenant_id = sa.tenant_id
  AND dt.model     = 'SA'
  AND dt.system_key = sa.adjust_reason
  AND sa.document_type_id IS NULL;

UPDATE quality_checks qc
SET document_type_id = dt.id
FROM document_types dt
WHERE dt.tenant_id = qc.tenant_id
  AND dt.model     = 'QC'
  AND dt.system_key = qc.qc_type
  AND qc.document_type_id IS NULL;

-- ── 7. Drop the old enum columns ────────────────────────────────────────────
ALTER TABLE goods_receipts    DROP COLUMN IF EXISTS grn_type;
ALTER TABLE goods_issues      DROP COLUMN IF EXISTS issue_reason;
ALTER TABLE stock_adjustments DROP COLUMN IF EXISTS adjust_reason;
ALTER TABLE quality_checks    DROP COLUMN IF EXISTS qc_type;
