DELETE FROM document_type_fields
WHERE document_type_id IN (
    SELECT id FROM document_types WHERE system_key IN ('REFINING_INTAKE')
);

DELETE FROM document_types
WHERE system_key IN ('REFINERY_SERVICE', 'REFINING_INTAKE', 'REFINING_PLAN');
