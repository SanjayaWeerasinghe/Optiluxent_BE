-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(255),
    is_system   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP
);

-- Unique role name per tenant (NULL tenant = system-wide role)
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_tenant_name
    ON roles(COALESCE(tenant_id, 0), name)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_roles_tenant_id  ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_is_system  ON roles(is_system);
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles(deleted_at);

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
