-- Create feature_flags table
CREATE TABLE IF NOT EXISTS feature_flags (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    enabled     BOOLEAN      NOT NULL DEFAULT false,
    rules       JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- Enforce uniqueness: one flag name per tenant, one global flag name (tenant_id IS NULL)
CREATE UNIQUE INDEX IF NOT EXISTS uidx_feature_flags_tenant_name
    ON feature_flags(tenant_id, name) WHERE tenant_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uidx_feature_flags_global_name
    ON feature_flags(name) WHERE tenant_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_feature_flags_tenant_id ON feature_flags(tenant_id);
CREATE INDEX IF NOT EXISTS idx_feature_flags_name      ON feature_flags(name);
CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled   ON feature_flags(enabled);

CREATE TRIGGER update_feature_flags_updated_at
    BEFORE UPDATE ON feature_flags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
