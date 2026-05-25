-- Master Data: Organization tables
-- countries, states, companies, departments, fiscal_years, accounting_periods, document_sequences
-- Note: companies.base_currency_id FK to currencies added in migration 000011

CREATE TABLE IF NOT EXISTS countries (
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(2)   NOT NULL,
    name       VARCHAR(100) NOT NULL,
    phone_code VARCHAR(10),
    is_active  BOOLEAN      NOT NULL DEFAULT true,
    CONSTRAINT uidx_countries_code UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_countries_is_active ON countries(is_active);

CREATE TABLE IF NOT EXISTS states (
    id         BIGSERIAL PRIMARY KEY,
    country_id BIGINT       NOT NULL REFERENCES countries(id) ON DELETE CASCADE,
    code       VARCHAR(10)  NOT NULL,
    name       VARCHAR(100) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_states_country_id ON states(country_id);

CREATE TABLE IF NOT EXISTS companies (
    id                 BIGSERIAL PRIMARY KEY,
    name               VARCHAR(200) NOT NULL,
    legal_name         VARCHAR(200),
    tax_reg_number     VARCHAR(50),
    logo               TEXT,
    email              VARCHAR(255),
    phone              VARCHAR(50),
    website            VARCHAR(255),
    address_line1      VARCHAR(255),
    address_line2      VARCHAR(255),
    city               VARCHAR(100),
    state_id           BIGINT REFERENCES states(id),
    country_id         BIGINT REFERENCES countries(id),
    postal_code        VARCHAR(20),
    base_currency_id   BIGINT,
    fiscal_year_start  SMALLINT NOT NULL DEFAULT 1 CHECK (fiscal_year_start BETWEEN 1 AND 12),
    created_at         TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_companies_updated_at
    BEFORE UPDATE ON companies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS departments (
    id                   BIGSERIAL PRIMARY KEY,
    tenant_id            BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code                 VARCHAR(20)  NOT NULL,
    name                 VARCHAR(100) NOT NULL,
    parent_id            BIGINT       REFERENCES departments(id),
    manager_employee_id  BIGINT,
    is_active            BOOLEAN      NOT NULL DEFAULT true,
    created_at           TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_departments_tenant_code
    ON departments(tenant_id, code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_departments_tenant_id  ON departments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_departments_parent_id  ON departments(parent_id);
CREATE INDEX IF NOT EXISTS idx_departments_deleted_at ON departments(deleted_at);

CREATE TRIGGER update_departments_updated_at
    BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS fiscal_years (
    id         BIGSERIAL PRIMARY KEY,
    tenant_id  BIGINT    NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name       VARCHAR(50) NOT NULL,
    start_date DATE      NOT NULL,
    end_date   DATE      NOT NULL,
    is_closed  BOOLEAN   NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_fiscal_years_dates CHECK (end_date > start_date)
);

CREATE INDEX IF NOT EXISTS idx_fiscal_years_tenant_id ON fiscal_years(tenant_id);

CREATE TRIGGER update_fiscal_years_updated_at
    BEFORE UPDATE ON fiscal_years
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS accounting_periods (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT    NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    fiscal_year_id BIGINT    NOT NULL REFERENCES fiscal_years(id) ON DELETE CASCADE,
    period_number  SMALLINT  NOT NULL CHECK (period_number BETWEEN 1 AND 12),
    name           VARCHAR(50) NOT NULL,
    start_date     DATE      NOT NULL,
    end_date       DATE      NOT NULL,
    is_closed      BOOLEAN   NOT NULL DEFAULT false,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_accounting_periods_fy_num UNIQUE (fiscal_year_id, period_number),
    CONSTRAINT chk_accounting_periods_dates CHECK (end_date > start_date)
);

CREATE INDEX IF NOT EXISTS idx_accounting_periods_tenant_id      ON accounting_periods(tenant_id);
CREATE INDEX IF NOT EXISTS idx_accounting_periods_fiscal_year_id ON accounting_periods(fiscal_year_id);

CREATE TRIGGER update_accounting_periods_updated_at
    BEFORE UPDATE ON accounting_periods
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS document_sequences (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT      NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL,
    prefix        VARCHAR(10) NOT NULL,
    next_number   INT         NOT NULL DEFAULT 1,
    padding       SMALLINT    NOT NULL DEFAULT 5,
    suffix        VARCHAR(20),
    created_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_document_sequences_tenant_type UNIQUE (tenant_id, document_type)
);

CREATE INDEX IF NOT EXISTS idx_document_sequences_tenant_id ON document_sequences(tenant_id);

CREATE TRIGGER update_document_sequences_updated_at
    BEFORE UPDATE ON document_sequences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
