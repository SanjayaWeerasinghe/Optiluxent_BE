-- Widen the MR status CHECK constraint so the two new transitions land
-- cleanly. Added: PENDING_APPROVAL (Submit for Approval), CANCELLED.

ALTER TABLE material_requests DROP CONSTRAINT IF EXISTS material_requests_status_check;
ALTER TABLE material_requests ADD  CONSTRAINT material_requests_status_check
    CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'REJECTED', 'FULFILLED', 'PARTIAL', 'CANCELLED'));
