ALTER TABLE material_requests DROP CONSTRAINT IF EXISTS material_requests_status_check;
ALTER TABLE material_requests ADD  CONSTRAINT material_requests_status_check
    CHECK (status IN ('DRAFT', 'APPROVED', 'REJECTED', 'FULFILLED', 'PARTIAL'));
