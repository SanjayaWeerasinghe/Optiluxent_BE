-- Master Data: HR (job_positions, employees)

CREATE TABLE IF NOT EXISTS job_positions (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code         VARCHAR(20)  NOT NULL,
    name         VARCHAR(200) NOT NULL,
    department_id BIGINT      REFERENCES departments(id),
    is_active    BOOLEAN      NOT NULL DEFAULT true,
    created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_job_positions_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_job_positions_tenant_id     ON job_positions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_job_positions_department_id ON job_positions(department_id);

CREATE TRIGGER update_job_positions_updated_at
    BEFORE UPDATE ON job_positions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS employees (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code             VARCHAR(20)  NOT NULL,
    first_name       VARCHAR(100) NOT NULL,
    last_name        VARCHAR(100) NOT NULL,
    display_name     VARCHAR(200),
    gender           VARCHAR(10)  CHECK (gender IN ('MALE','FEMALE','OTHER')),
    date_of_birth    DATE,
    nic_number       VARCHAR(20),
    job_position_id  BIGINT       REFERENCES job_positions(id),
    department_id    BIGINT       REFERENCES departments(id),
    manager_id       BIGINT       REFERENCES employees(id),
    employment_type  VARCHAR(20)  NOT NULL DEFAULT 'PERMANENT' CHECK (employment_type IN ('PERMANENT','CONTRACT','PART_TIME','INTERN')),
    date_joined      DATE         NOT NULL,
    date_left        DATE,
    email            VARCHAR(255),
    phone            VARCHAR(50),
    mobile           VARCHAR(50),
    bank_id          BIGINT       REFERENCES banks(id),
    bank_account_no  VARCHAR(50),
    basic_salary     NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency_id      BIGINT       NOT NULL REFERENCES currencies(id),
    is_active        BOOLEAN      NOT NULL DEFAULT true,
    notes            TEXT,
    created_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMP,
    CONSTRAINT uidx_employees_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_employees_tenant_id       ON employees(tenant_id);
CREATE INDEX IF NOT EXISTS idx_employees_department_id   ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_employees_job_position_id ON employees(job_position_id);
CREATE INDEX IF NOT EXISTS idx_employees_manager_id      ON employees(manager_id);
CREATE INDEX IF NOT EXISTS idx_employees_is_active       ON employees(is_active);
CREATE INDEX IF NOT EXISTS idx_employees_deleted_at      ON employees(deleted_at);

CREATE TRIGGER update_employees_updated_at
    BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
