-- Master Data: Contacts (parties, contact_persons, addresses)

CREATE TABLE IF NOT EXISTS parties (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(20)   NOT NULL,
    name            VARCHAR(200)  NOT NULL,
    legal_name      VARCHAR(200),
    party_type      VARCHAR(10)   NOT NULL CHECK (party_type IN ('CUSTOMER','SUPPLIER','BOTH')),
    type            VARCHAR(10)   NOT NULL DEFAULT 'COMPANY' CHECK (type IN ('INDIVIDUAL','COMPANY')),
    tax_reg_number  VARCHAR(50),
    currency_id     BIGINT        NOT NULL REFERENCES currencies(id),
    payment_term_id BIGINT        REFERENCES payment_terms(id),
    credit_limit    NUMERIC(18,2) NOT NULL DEFAULT 0,
    is_active       BOOLEAN       NOT NULL DEFAULT true,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,
    CONSTRAINT uidx_parties_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_parties_tenant_id   ON parties(tenant_id);
CREATE INDEX IF NOT EXISTS idx_parties_party_type  ON parties(party_type);
CREATE INDEX IF NOT EXISTS idx_parties_is_active   ON parties(is_active);
CREATE INDEX IF NOT EXISTS idx_parties_deleted_at  ON parties(deleted_at);

CREATE TRIGGER update_parties_updated_at
    BEFORE UPDATE ON parties
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS contact_persons (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    party_id    BIGINT       NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    name        VARCHAR(200) NOT NULL,
    designation VARCHAR(100),
    phone       VARCHAR(50),
    mobile      VARCHAR(50),
    email       VARCHAR(255),
    is_primary  BOOLEAN      NOT NULL DEFAULT false,
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contact_persons_party_id  ON contact_persons(party_id);
CREATE INDEX IF NOT EXISTS idx_contact_persons_tenant_id ON contact_persons(tenant_id);

CREATE TRIGGER update_contact_persons_updated_at
    BEFORE UPDATE ON contact_persons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS party_addresses (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    party_id      BIGINT       NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    address_type  VARCHAR(10)  NOT NULL DEFAULT 'BOTH' CHECK (address_type IN ('BILLING','SHIPPING','BOTH')),
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city          VARCHAR(100),
    state_id      BIGINT       REFERENCES states(id),
    country_id    BIGINT       NOT NULL REFERENCES countries(id),
    postal_code   VARCHAR(20),
    is_primary    BOOLEAN      NOT NULL DEFAULT false,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_party_addresses_party_id  ON party_addresses(party_id);
CREATE INDEX IF NOT EXISTS idx_party_addresses_tenant_id ON party_addresses(tenant_id);

CREATE TRIGGER update_party_addresses_updated_at
    BEFORE UPDATE ON party_addresses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
