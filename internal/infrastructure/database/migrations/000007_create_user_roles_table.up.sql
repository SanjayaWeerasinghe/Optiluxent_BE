-- Create user_roles join table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id   BIGINT NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    role_id   BIGINT NOT NULL REFERENCES roles(id)   ON DELETE CASCADE,
    tenant_id BIGINT REFERENCES tenants(id)          ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_role_id   ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_id ON user_roles(tenant_id);
