-- Seed the three new Document Types for the Refinery Service flow.
-- - SO / REFINERY_SERVICE  → classifies the sale as a refining service
-- - GRN / REFINING_INTAKE  → intake of customer's crude oil, tagged with a
--                            source Production picker
-- - PLAN / REFINING_PLAN   → a Production Plan specifically for refining
--
-- Uses the same shape as 000034_seed_document_types: one row per tenant.

-- Refinery Service SO Type
INSERT INTO document_types (tenant_id, model, code, name, system_key, is_active)
SELECT t.id, 'SO', 'REFINERY_SERVICE', 'Refinery Service', 'REFINERY_SERVICE', TRUE
FROM tenants t
WHERE NOT EXISTS (
    SELECT 1 FROM document_types dt
    WHERE dt.tenant_id = t.id AND dt.model = 'SO' AND dt.code = 'REFINERY_SERVICE'
);

-- Refining Intake GRN Type
INSERT INTO document_types (tenant_id, model, code, name, system_key, is_active)
SELECT t.id, 'GRN', 'REFINING_INTAKE', 'Refining Intake', 'REFINING_INTAKE', TRUE
FROM tenants t
WHERE NOT EXISTS (
    SELECT 1 FROM document_types dt
    WHERE dt.tenant_id = t.id AND dt.model = 'GRN' AND dt.code = 'REFINING_INTAKE'
);

-- Picker field on the Refining Intake GRN Type: "For Production" (kind='MO')
INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'source_production', 'For Production', 'MO', TRUE, 1
FROM document_types dt
WHERE dt.system_key = 'REFINING_INTAKE'
  AND NOT EXISTS (
      SELECT 1 FROM document_type_fields dtf
      WHERE dtf.document_type_id = dt.id AND dtf.code = 'source_production'
  );

-- Refining Service Plan Type
INSERT INTO document_types (tenant_id, model, code, name, system_key, is_active)
SELECT t.id, 'PLAN', 'REFINING_PLAN', 'Refining Service Plan', 'REFINING_PLAN', TRUE
FROM tenants t
WHERE NOT EXISTS (
    SELECT 1 FROM document_types dt
    WHERE dt.tenant_id = t.id AND dt.model = 'PLAN' AND dt.code = 'REFINING_PLAN'
);
