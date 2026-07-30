-- HR module extensions. Employees master already has basic_salary + join
-- date; this migration adds the transactional tables + a couple of
-- long-missing employee columns.

-- ── Employee columns ────────────────────────────────────────────────────────
ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS contract_end_date DATE,
    ADD COLUMN IF NOT EXISTS cv_url            VARCHAR(500) DEFAULT '';

-- ── Family members ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS employee_family (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    employee_id    BIGINT       NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    relation       VARCHAR(30)  NOT NULL,                      -- SPOUSE / CHILD / PARENT / SIBLING / OTHER
    full_name      VARCHAR(200) NOT NULL,
    date_of_birth  DATE,
    is_dependent   BOOLEAN      NOT NULL DEFAULT FALSE,
    notes          VARCHAR(500),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_family_employee ON employee_family(employee_id);

-- ── Emergency contacts ─────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS employee_emergency_contact (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    employee_id    BIGINT       NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    full_name      VARCHAR(200) NOT NULL,
    relation       VARCHAR(30)  NOT NULL,
    phone          VARCHAR(50)  NOT NULL,
    alt_phone      VARCHAR(50),
    address        VARCHAR(500),
    is_primary     BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_emergency_employee ON employee_emergency_contact(employee_id);

-- ── Attendance ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS attendance (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    employee_id   BIGINT       NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    attend_date   DATE         NOT NULL,
    check_in      TIME,
    check_out     TIME,
    status        VARCHAR(20)  NOT NULL DEFAULT 'PRESENT'
                  CHECK (status IN ('PRESENT','ABSENT','HALF_DAY','LEAVE','HOLIDAY')),
    hours_worked  NUMERIC(6,2) NOT NULL DEFAULT 0,
    notes         VARCHAR(500),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_attendance_emp_date UNIQUE (tenant_id, employee_id, attend_date)
);
CREATE INDEX IF NOT EXISTS idx_attendance_date ON attendance(attend_date);

-- ── Salary revisions (append-only) ─────────────────────────────────────────
-- Every change to basic_salary should also insert a row here so the FE
-- can show a history strip; the current effective salary is whichever
-- row has the latest effective_date <= today.
CREATE TABLE IF NOT EXISTS salary_history (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    employee_id     BIGINT       NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    effective_date  DATE         NOT NULL,
    basic_salary    NUMERIC(18,2) NOT NULL,
    reason          VARCHAR(200),
    created_by      BIGINT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_salary_history_emp ON salary_history(employee_id, effective_date DESC);
