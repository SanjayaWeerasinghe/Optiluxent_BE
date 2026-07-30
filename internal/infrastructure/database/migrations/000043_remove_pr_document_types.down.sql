-- Restore the PR "With Supplier" and "Without Supplier" Types (from 000034)
-- so a downgrade returns the schema to its previous state.
INSERT INTO document_types (tenant_id, model, code, name, system_key)
SELECT t.id, 'PR', v.code, v.name, NULL
FROM tenants t
CROSS JOIN (VALUES
    ('WITH_SUPPLIER',    'With Supplier'),
    ('WITHOUT_SUPPLIER', 'Without Supplier')
) AS v(code, name);

INSERT INTO document_type_fields (tenant_id, document_type_id, code, label, kind, is_required, display_order)
SELECT dt.tenant_id, dt.id, 'supplier', 'Supplier', 'Supplier', TRUE, 1
FROM document_types dt
WHERE dt.model = 'PR' AND dt.code = 'WITH_SUPPLIER';
