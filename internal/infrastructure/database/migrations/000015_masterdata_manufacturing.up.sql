-- Master Data: Manufacturing (bill_of_materials, bom_lines, work_centers, routings, routing_operations)

CREATE TABLE IF NOT EXISTS bill_of_materials (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(50)  NOT NULL,
    product_id      BIGINT       NOT NULL REFERENCES products(id),
    variant_id      BIGINT       REFERENCES product_variants(id),
    quantity        NUMERIC(18,4) NOT NULL DEFAULT 1,
    uom_id          BIGINT       NOT NULL REFERENCES units_of_measure(id),
    bom_type        VARCHAR(20)  NOT NULL DEFAULT 'MANUFACTURE' CHECK (bom_type IN ('MANUFACTURE','KIT','SUBCONTRACT')),
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    notes           TEXT,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,
    CONSTRAINT uidx_bom_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_bom_tenant_id   ON bill_of_materials(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bom_product_id  ON bill_of_materials(product_id);
CREATE INDEX IF NOT EXISTS idx_bom_deleted_at  ON bill_of_materials(deleted_at);

CREATE TRIGGER update_bill_of_materials_updated_at
    BEFORE UPDATE ON bill_of_materials
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS bom_lines (
    id              BIGSERIAL PRIMARY KEY,
    bom_id          BIGINT        NOT NULL REFERENCES bill_of_materials(id) ON DELETE CASCADE,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    component_id    BIGINT        NOT NULL REFERENCES products(id),
    variant_id      BIGINT        REFERENCES product_variants(id),
    quantity        NUMERIC(18,4) NOT NULL,
    uom_id          BIGINT        NOT NULL REFERENCES units_of_measure(id),
    scrap_percent   NUMERIC(5,2)  NOT NULL DEFAULT 0,
    sequence        INT           NOT NULL DEFAULT 10,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bom_lines_bom_id       ON bom_lines(bom_id);
CREATE INDEX IF NOT EXISTS idx_bom_lines_component_id ON bom_lines(component_id);

CREATE TRIGGER update_bom_lines_updated_at
    BEFORE UPDATE ON bom_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS work_centers (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(20)   NOT NULL,
    name            VARCHAR(200)  NOT NULL,
    capacity        NUMERIC(10,2) NOT NULL DEFAULT 1,
    cost_per_hour   NUMERIC(18,4) NOT NULL DEFAULT 0,
    currency_id     BIGINT        NOT NULL REFERENCES currencies(id),
    is_active       BOOLEAN       NOT NULL DEFAULT true,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,
    CONSTRAINT uidx_work_centers_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_work_centers_tenant_id  ON work_centers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_work_centers_deleted_at ON work_centers(deleted_at);

CREATE TRIGGER update_work_centers_updated_at
    BEFORE UPDATE ON work_centers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS routings (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code        VARCHAR(50)  NOT NULL,
    product_id  BIGINT       NOT NULL REFERENCES products(id),
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    notes       TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP,
    CONSTRAINT uidx_routings_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_routings_tenant_id   ON routings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_routings_product_id  ON routings(product_id);
CREATE INDEX IF NOT EXISTS idx_routings_deleted_at  ON routings(deleted_at);

CREATE TRIGGER update_routings_updated_at
    BEFORE UPDATE ON routings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS routing_operations (
    id              BIGSERIAL PRIMARY KEY,
    routing_id      BIGINT        NOT NULL REFERENCES routings(id) ON DELETE CASCADE,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    sequence        INT           NOT NULL DEFAULT 10,
    name            VARCHAR(200)  NOT NULL,
    work_center_id  BIGINT        NOT NULL REFERENCES work_centers(id),
    setup_time      NUMERIC(10,2) NOT NULL DEFAULT 0,
    cycle_time      NUMERIC(10,2) NOT NULL DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_routing_operations_routing_id     ON routing_operations(routing_id);
CREATE INDEX IF NOT EXISTS idx_routing_operations_work_center_id ON routing_operations(work_center_id);

CREATE TRIGGER update_routing_operations_updated_at
    BEFORE UPDATE ON routing_operations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
