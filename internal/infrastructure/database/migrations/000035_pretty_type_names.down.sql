-- Restore raw enum-style names — matches the initial 000033 seed.
UPDATE document_types SET name = code WHERE system_key IS NOT NULL;
