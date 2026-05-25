-- Master Data: Financial tables
-- currencies, exchange_rates, chart_of_accounts, cost_centers, payment_terms,
-- banks, company_bank_accounts, tax_codes, tax_groups, tax_group_codes
-- Also wires companies.base_currency_id FK (deferred from migration 000010)

CREATE TABLE IF NOT EXISTS currencies (
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(3)  NOT NULL,
    name       VARCHAR(100) NOT NULL,
    symbol     VARCHAR(10) NOT NULL,
    is_base    BOOLEAN     NOT NULL DEFAULT false,
    is_active  BOOLEAN     NOT NULL DEFAULT true,
    created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_currencies_code UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_currencies_is_base   ON currencies(is_base);
CREATE INDEX IF NOT EXISTS idx_currencies_is_active ON currencies(is_active);

CREATE TRIGGER update_currencies_updated_at
    BEFORE UPDATE ON currencies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add the FK from companies.base_currency_id now that currencies exists
ALTER TABLE companies
    ADD CONSTRAINT fk_companies_base_currency
    FOREIGN KEY (base_currency_id) REFERENCES currencies(id);

CREATE TABLE IF NOT EXISTS exchange_rates (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT    NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    from_currency_id BIGINT    NOT NULL REFERENCES currencies(id),
    to_currency_id   BIGINT    NOT NULL REFERENCES currencies(id),
    rate             NUMERIC(20, 6) NOT NULL CHECK (rate > 0),
    effective_date   DATE      NOT NULL,
    source           VARCHAR(30) NOT NULL DEFAULT 'manual',
    created_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exchange_rates_tenant_id        ON exchange_rates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_from_currency_id ON exchange_rates(from_currency_id);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_effective_date   ON exchange_rates(effective_date);

CREATE TABLE IF NOT EXISTS chart_of_accounts (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT      NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code          VARCHAR(20) NOT NULL,
    name          VARCHAR(200) NOT NULL,
    account_type  VARCHAR(20) NOT NULL CHECK (account_type IN ('ASSET','LIABILITY','EQUITY','REVENUE','EXPENSE')),
    parent_id     BIGINT      REFERENCES chart_of_accounts(id),
    currency_id   BIGINT      NOT NULL REFERENCES currencies(id),
    is_controlled BOOLEAN     NOT NULL DEFAULT false,
    is_active     BOOLEAN     NOT NULL DEFAULT true,
    level         SMALLINT    NOT NULL DEFAULT 1,
    created_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP,
    CONSTRAINT uidx_coa_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_coa_tenant_id  ON chart_of_accounts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_coa_parent_id  ON chart_of_accounts(parent_id);
CREATE INDEX IF NOT EXISTS idx_coa_deleted_at ON chart_of_accounts(deleted_at);

CREATE TRIGGER update_chart_of_accounts_updated_at
    BEFORE UPDATE ON chart_of_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS cost_centers (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT      NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code          VARCHAR(20) NOT NULL,
    name          VARCHAR(100) NOT NULL,
    department_id BIGINT      REFERENCES departments(id),
    is_active     BOOLEAN     NOT NULL DEFAULT true,
    created_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP,
    CONSTRAINT uidx_cost_centers_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_cost_centers_tenant_id ON cost_centers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cost_centers_deleted_at ON cost_centers(deleted_at);

CREATE TRIGGER update_cost_centers_updated_at
    BEFORE UPDATE ON cost_centers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS payment_terms (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code             VARCHAR(20)  NOT NULL,
    name             VARCHAR(100) NOT NULL,
    due_days         SMALLINT     NOT NULL DEFAULT 0,
    discount_days    SMALLINT     NOT NULL DEFAULT 0,
    discount_percent NUMERIC(5,2) NOT NULL DEFAULT 0,
    is_active        BOOLEAN      NOT NULL DEFAULT true,
    created_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_payment_terms_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_payment_terms_tenant_id ON payment_terms(tenant_id);

CREATE TRIGGER update_payment_terms_updated_at
    BEFORE UPDATE ON payment_terms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS banks (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    branch_name VARCHAR(200),
    swift_code  VARCHAR(11),
    address     TEXT,
    is_active   BOOLEAN   NOT NULL DEFAULT true,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_banks_updated_at
    BEFORE UPDATE ON banks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS company_bank_accounts (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    bank_id        BIGINT       NOT NULL REFERENCES banks(id),
    account_number VARCHAR(50)  NOT NULL,
    account_name   VARCHAR(200) NOT NULL,
    currency_id    BIGINT       NOT NULL REFERENCES currencies(id),
    gl_account_id  BIGINT       NOT NULL REFERENCES chart_of_accounts(id),
    is_default     BOOLEAN      NOT NULL DEFAULT false,
    is_active      BOOLEAN      NOT NULL DEFAULT true,
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_company_bank_accounts_tenant_id ON company_bank_accounts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_company_bank_accounts_bank_id   ON company_bank_accounts(bank_id);

CREATE TRIGGER update_company_bank_accounts_updated_at
    BEFORE UPDATE ON company_bank_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS tax_codes (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code         VARCHAR(20)  NOT NULL,
    name         VARCHAR(100) NOT NULL,
    tax_type     VARCHAR(10)  NOT NULL CHECK (tax_type IN ('VAT','WHT','SVAT','EXEMPT')),
    rate         NUMERIC(8,4) NOT NULL CHECK (rate >= 0),
    gl_account_id BIGINT      NOT NULL REFERENCES chart_of_accounts(id),
    is_active    BOOLEAN      NOT NULL DEFAULT true,
    created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_tax_codes_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_tax_codes_tenant_id ON tax_codes(tenant_id);

CREATE TRIGGER update_tax_codes_updated_at
    BEFORE UPDATE ON tax_codes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS tax_groups (
    id         BIGSERIAL PRIMARY KEY,
    tenant_id  BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_tax_groups_tenant_name UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tax_groups_tenant_id ON tax_groups(tenant_id);

CREATE TRIGGER update_tax_groups_updated_at
    BEFORE UPDATE ON tax_groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS tax_group_codes (
    tax_group_id BIGINT NOT NULL REFERENCES tax_groups(id) ON DELETE CASCADE,
    tax_code_id  BIGINT NOT NULL REFERENCES tax_codes(id)  ON DELETE CASCADE,
    PRIMARY KEY (tax_group_id, tax_code_id)
);
