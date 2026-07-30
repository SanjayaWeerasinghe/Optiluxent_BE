-- Purchase Requests no longer surface a Type picker (or the supplier field
-- attached to the "With Supplier" Type) — the supplier decision belongs to
-- the downstream Purchase Order, not the request itself.
--
-- Null any existing PR references to those Types first, then remove the
-- Types themselves (cascade cleans up document_type_fields and
-- document_field_values).
UPDATE purchase_requests
SET document_type_id = NULL
WHERE document_type_id IN (
    SELECT id FROM document_types
    WHERE model = 'PR' AND code IN ('WITH_SUPPLIER', 'WITHOUT_SUPPLIER')
);

DELETE FROM document_types
WHERE model = 'PR' AND code IN ('WITH_SUPPLIER', 'WITHOUT_SUPPLIER');
