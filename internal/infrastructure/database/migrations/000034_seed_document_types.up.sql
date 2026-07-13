-- Seed user-defined Document Types + their picker fields per tenant.
--
-- These are "user" Types (no system_key) — pure classification. They ship
-- pre-configured with the picker slots the business asked for:
--   PR  → With Supplier / Without Supplier
--   PO  → With PR / Without PR
--   GRN → (system Types already exist — attach picker fields to them)
--   MR  → For Production / For Departments / For Sales
--   GI  → With MR / With SO
--
-- Each block is ON CONFLICT DO NOTHING so re-running is a no-op.

-- ── PR Types ─────────────────────────────────────────────────────────────────
INSERT INTO document_types (tenant_id, model, code, name, system_key)
SELECT t.id, 'PR', v.code, v.name, NULL
FROM tenants t CROSS JOIN (VALUES
  ('WITH_SUPPLIER',    'With Supplier'),
  ('WITHOUT_SUPPLIER', 'Without Supplier')
) v(code, name)
ON CONFLICT (tenant_id, model, code) DO NOTHING;

-- PR "With Supplier" → Supplier picker field
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'supplier', 'Supplier', 'Supplier', TRUE, 1
FROM document_types dt
WHERE dt.model = 'PR' AND dt.code = 'WITH_SUPPLIER'
ON CONFLICT (document_type_id, code) DO NOTHING;

-- ── PO Types ─────────────────────────────────────────────────────────────────
INSERT INTO document_types (tenant_id, model, code, name, system_key)
SELECT t.id, 'PO', v.code, v.name, NULL
FROM tenants t CROSS JOIN (VALUES
  ('WITH_PR',    'With Purchase Request'),
  ('WITHOUT_PR', 'Without Purchase Request')
) v(code, name)
ON CONFLICT (tenant_id, model, code) DO NOTHING;

-- PO "With PR" → PR picker
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'source_pr', 'Purchase Request', 'PR', TRUE, 1
FROM document_types dt
WHERE dt.model = 'PO' AND dt.code = 'WITH_PR'
ON CONFLICT (document_type_id, code) DO NOTHING;

-- ── GRN Types — the system Types were seeded by 000033. Attach picker fields
--     to the ones that benefit from a linked-doc slot. ────────────────────────

-- WITH_PO   → Purchase Order picker
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'source_po', 'Purchase Order', 'PO', FALSE, 1
FROM document_types dt
WHERE dt.model = 'GRN' AND dt.code = 'WITH_PO'
ON CONFLICT (document_type_id, code) DO NOTHING;

-- CUSTOMER_RETURN → Customer picker + Sales Invoice picker
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'customer', 'Customer', 'Customer', TRUE, 1
FROM document_types dt
WHERE dt.model = 'GRN' AND dt.code = 'CUSTOMER_RETURN'
ON CONFLICT (document_type_id, code) DO NOTHING;

INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'source_invoice', 'Sales Invoice', 'SI', FALSE, 2
FROM document_types dt
WHERE dt.model = 'GRN' AND dt.code = 'CUSTOMER_RETURN'
ON CONFLICT (document_type_id, code) DO NOTHING;

-- PRODUCTION_OUTPUT / PRODUCTION_RETURN → Manufacturing Order picker
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'source_mo', 'Manufacturing Order', 'MO', FALSE, 1
FROM document_types dt
WHERE dt.model = 'GRN' AND dt.code IN ('PRODUCTION_OUTPUT', 'PRODUCTION_RETURN')
ON CONFLICT (document_type_id, code) DO NOTHING;

-- ── MR Types ─────────────────────────────────────────────────────────────────
INSERT INTO document_types (tenant_id, model, code, name, system_key)
SELECT t.id, 'MR', v.code, v.name, NULL
FROM tenants t CROSS JOIN (VALUES
  ('FOR_PRODUCTION',  'For Production'),
  ('FOR_DEPARTMENTS', 'For Departments'),
  ('FOR_SALES',       'For Sales')
) v(code, name)
ON CONFLICT (tenant_id, model, code) DO NOTHING;

-- MR "For Production" → Manufacturing Order picker
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'target_mo', 'Manufacturing Order', 'MO', TRUE, 1
FROM document_types dt
WHERE dt.model = 'MR' AND dt.code = 'FOR_PRODUCTION'
ON CONFLICT (document_type_id, code) DO NOTHING;

-- MR "For Departments" → Department picker
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'target_department', 'Department', 'Department', TRUE, 1
FROM document_types dt
WHERE dt.model = 'MR' AND dt.code = 'FOR_DEPARTMENTS'
ON CONFLICT (document_type_id, code) DO NOTHING;

-- MR "For Sales" → Sales Order picker
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'target_so', 'Sales Order', 'SO', TRUE, 1
FROM document_types dt
WHERE dt.model = 'MR' AND dt.code = 'FOR_SALES'
ON CONFLICT (document_type_id, code) DO NOTHING;

-- ── GI Types ─────────────────────────────────────────────────────────────────
INSERT INTO document_types (tenant_id, model, code, name, system_key)
SELECT t.id, 'GI', v.code, v.name, NULL
FROM tenants t CROSS JOIN (VALUES
  ('WITH_MR', 'With Material Request'),
  ('WITH_SO', 'With Sales Order')
) v(code, name)
ON CONFLICT (tenant_id, model, code) DO NOTHING;

-- GI "With MR" → MR picker
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'source_mr', 'Material Request', 'MR', TRUE, 1
FROM document_types dt
WHERE dt.model = 'GI' AND dt.code = 'WITH_MR'
ON CONFLICT (document_type_id, code) DO NOTHING;

-- GI "With SO" → SO picker
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'source_so', 'Sales Order', 'SO', TRUE, 1
FROM document_types dt
WHERE dt.model = 'GI' AND dt.code = 'WITH_SO'
ON CONFLICT (document_type_id, code) DO NOTHING;
