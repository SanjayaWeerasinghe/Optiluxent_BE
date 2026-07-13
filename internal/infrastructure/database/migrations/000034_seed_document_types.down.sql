-- Reverse the 000034 seed. Deletes the user-defined Types we inserted
-- (identified by (model, code) pairs); their fields cascade via ON DELETE.
-- Also removes the picker fields we attached to existing system GRN Types.

-- User-defined Types → drop them (their fields + values cascade)
DELETE FROM document_types
WHERE (model = 'PR' AND code IN ('WITH_SUPPLIER', 'WITHOUT_SUPPLIER'))
   OR (model = 'PO' AND code IN ('WITH_PR', 'WITHOUT_PR'))
   OR (model = 'MR' AND code IN ('FOR_PRODUCTION', 'FOR_DEPARTMENTS', 'FOR_SALES'))
   OR (model = 'GI' AND code IN ('WITH_MR', 'WITH_SO'));

-- System GRN Types stay; their fields we added are removed by field code.
DELETE FROM document_type_fields
WHERE code IN ('source_po', 'customer', 'source_invoice', 'source_mo')
  AND document_type_id IN (
    SELECT id FROM document_types
    WHERE model = 'GRN' AND code IN ('WITH_PO', 'CUSTOMER_RETURN', 'PRODUCTION_OUTPUT', 'PRODUCTION_RETURN')
  );
