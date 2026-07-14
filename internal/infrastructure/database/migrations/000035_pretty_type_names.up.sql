-- Prettify the display names for seeded Document Types. Codes stay as the
-- machine key (WITH_PO, CUSTOMER_RETURN, ...), names become the
-- human-friendly label used in dropdowns + list columns.

UPDATE document_types SET name = 'With Purchase Order'         WHERE model = 'GRN' AND code = 'WITH_PO';
UPDATE document_types SET name = 'Without Purchase Order'      WHERE model = 'GRN' AND code = 'WITHOUT_PO';
UPDATE document_types SET name = 'Customer Return'             WHERE model = 'GRN' AND code = 'CUSTOMER_RETURN';
UPDATE document_types SET name = 'Production Return'           WHERE model = 'GRN' AND code = 'PRODUCTION_RETURN';
UPDATE document_types SET name = 'Production Output'           WHERE model = 'GRN' AND code = 'PRODUCTION_OUTPUT';

UPDATE document_types SET name = 'For Production'              WHERE model = 'GI'  AND code = 'PRODUCTION';
UPDATE document_types SET name = 'For Sale'                    WHERE model = 'GI'  AND code = 'SALE';
UPDATE document_types SET name = 'Expense'                     WHERE model = 'GI'  AND code = 'EXPENSE';
UPDATE document_types SET name = 'Damage'                      WHERE model = 'GI'  AND code = 'DAMAGE';
UPDATE document_types SET name = 'Adjustment'                  WHERE model = 'GI'  AND code = 'ADJUSTMENT';
UPDATE document_types SET name = 'QC Rejection'                WHERE model = 'GI'  AND code = 'QC_REJECTION';
UPDATE document_types SET name = 'Other'                       WHERE model = 'GI'  AND code = 'OTHER';

UPDATE document_types SET name = 'Stocktake'                   WHERE model = 'SA'  AND code = 'STOCKTAKE';
UPDATE document_types SET name = 'Damage'                      WHERE model = 'SA'  AND code = 'DAMAGE';
UPDATE document_types SET name = 'Expiry'                      WHERE model = 'SA'  AND code = 'EXPIRY';
UPDATE document_types SET name = 'Write-Off'                   WHERE model = 'SA'  AND code = 'WRITE_OFF';
UPDATE document_types SET name = 'Correction'                  WHERE model = 'SA'  AND code = 'CORRECTION';
UPDATE document_types SET name = 'Other'                       WHERE model = 'SA'  AND code = 'OTHER';

UPDATE document_types SET name = 'Material Quality Check'      WHERE model = 'QC'  AND code = 'MATERIAL_QC';
UPDATE document_types SET name = 'Product Quality Check'       WHERE model = 'QC'  AND code = 'PRODUCT_QC';
