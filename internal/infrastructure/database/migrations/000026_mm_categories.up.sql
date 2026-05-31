CREATE TABLE IF NOT EXISTS mm_categories (
    id         BIGSERIAL    PRIMARY KEY,
    tenant_id  BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code       VARCHAR(50)  NOT NULL,
    name       VARCHAR(200) NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_mm_categories UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_mm_categories_tenant ON mm_categories(tenant_id);

-- Re-point mm_materials.category_id to mm_categories
ALTER TABLE mm_materials DROP CONSTRAINT IF EXISTS mm_materials_category_id_fkey;
ALTER TABLE mm_materials
    ADD CONSTRAINT mm_materials_category_id_fkey
    FOREIGN KEY (category_id) REFERENCES mm_categories(id);
