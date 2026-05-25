-- Manufacturing module: cost_estimates, production_plans, production_orders, post_costs

-- ── Cost Estimates (Pre-Costing) ──────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cost_estimates (
    id                       BIGSERIAL PRIMARY KEY,
    tenant_id                BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code                     VARCHAR(50)   NOT NULL,
    product_id               BIGINT        NOT NULL REFERENCES products(id),
    uom_id                   BIGINT        NOT NULL REFERENCES units_of_measure(id),
    planned_qty              NUMERIC(18,4) NOT NULL DEFAULT 0,
    bom_id                   BIGINT        REFERENCES bill_of_materials(id),
    routing_id               BIGINT        REFERENCES routings(id),
    planned_start_date       DATE,
    planned_end_date         DATE,
    estimated_material_cost  NUMERIC(18,4) NOT NULL DEFAULT 0,
    estimated_resource_cost  NUMERIC(18,4) NOT NULL DEFAULT 0,
    estimated_overhead_cost  NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_estimated_cost     NUMERIC(18,4) NOT NULL DEFAULT 0,
    status                   VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                                 CHECK (status IN ('DRAFT','APPROVED','CANCELLED')),
    notes                    TEXT,
    approved_by              BIGINT,
    approved_at              TIMESTAMP,
    created_by               BIGINT,
    created_at               TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMP,
    CONSTRAINT uidx_cost_estimates_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_cost_estimates_tenant_id  ON cost_estimates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cost_estimates_status     ON cost_estimates(status);
CREATE INDEX IF NOT EXISTS idx_cost_estimates_deleted_at ON cost_estimates(deleted_at);

CREATE TRIGGER update_cost_estimates_updated_at
    BEFORE UPDATE ON cost_estimates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS cost_estimate_lines (
    id             BIGSERIAL PRIMARY KEY,
    estimate_id    BIGINT        NOT NULL REFERENCES cost_estimates(id) ON DELETE CASCADE,
    tenant_id      BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    line_number    INT           NOT NULL DEFAULT 1,
    line_type      VARCHAR(20)   NOT NULL
                       CHECK (line_type IN ('MATERIAL','RESOURCE','LABOR','OVERHEAD')),
    product_id     BIGINT        REFERENCES products(id),
    work_center_id BIGINT        REFERENCES work_centers(id),
    description    VARCHAR(500),
    quantity       NUMERIC(18,4) NOT NULL DEFAULT 0,
    uom_id         BIGINT        REFERENCES units_of_measure(id),
    unit_cost      NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost     NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes          TEXT,
    created_at     TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cost_estimate_lines_estimate_id ON cost_estimate_lines(estimate_id);
CREATE INDEX IF NOT EXISTS idx_cost_estimate_lines_tenant_id   ON cost_estimate_lines(tenant_id);

CREATE TRIGGER update_cost_estimate_lines_updated_at
    BEFORE UPDATE ON cost_estimate_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Production Plans ──────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS production_plans (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code                VARCHAR(50)   NOT NULL,
    product_id          BIGINT        NOT NULL REFERENCES products(id),
    uom_id              BIGINT        NOT NULL REFERENCES units_of_measure(id),
    planned_qty         NUMERIC(18,4) NOT NULL DEFAULT 0,
    bom_id              BIGINT        REFERENCES bill_of_materials(id),
    routing_id          BIGINT        REFERENCES routings(id),
    estimate_id         BIGINT        REFERENCES cost_estimates(id),
    warehouse_id        BIGINT        NOT NULL REFERENCES warehouses(id),
    planned_start_date  DATE,
    planned_end_date    DATE,
    status              VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                            CHECK (status IN ('DRAFT','RELEASED','IN_PROGRESS','COMPLETED','CANCELLED')),
    notes               TEXT,
    released_by         BIGINT,
    released_at         TIMESTAMP,
    created_by          BIGINT,
    created_at          TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMP,
    CONSTRAINT uidx_production_plans_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_production_plans_tenant_id   ON production_plans(tenant_id);
CREATE INDEX IF NOT EXISTS idx_production_plans_status      ON production_plans(status);
CREATE INDEX IF NOT EXISTS idx_production_plans_deleted_at  ON production_plans(deleted_at);

CREATE TRIGGER update_production_plans_updated_at
    BEFORE UPDATE ON production_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Production Orders ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS production_orders (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code         VARCHAR(50)   NOT NULL,
    plan_id      BIGINT        REFERENCES production_plans(id),
    product_id   BIGINT        NOT NULL REFERENCES products(id),
    uom_id       BIGINT        NOT NULL REFERENCES units_of_measure(id),
    planned_qty  NUMERIC(18,4) NOT NULL DEFAULT 0,
    produced_qty NUMERIC(18,4) NOT NULL DEFAULT 0,
    warehouse_id BIGINT        NOT NULL REFERENCES warehouses(id),
    start_date   DATE,
    end_date     DATE,
    status       VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                     CHECK (status IN ('DRAFT','IN_PROGRESS','COMPLETED','CANCELLED')),
    notes        TEXT,
    completed_by BIGINT,
    completed_at TIMESTAMP,
    created_by   BIGINT,
    created_at   TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMP,
    CONSTRAINT uidx_production_orders_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_production_orders_tenant_id  ON production_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_production_orders_status     ON production_orders(status);
CREATE INDEX IF NOT EXISTS idx_production_orders_plan_id    ON production_orders(plan_id);
CREATE INDEX IF NOT EXISTS idx_production_orders_deleted_at ON production_orders(deleted_at);

CREATE TRIGGER update_production_orders_updated_at
    BEFORE UPDATE ON production_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS production_outputs (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT        NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    tenant_id    BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    line_number  INT           NOT NULL DEFAULT 1,
    product_id   BIGINT        NOT NULL REFERENCES products(id),
    uom_id       BIGINT        NOT NULL REFERENCES units_of_measure(id),
    quantity     NUMERIC(18,4) NOT NULL DEFAULT 0,
    unit_cost    NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost   NUMERIC(18,4) NOT NULL DEFAULT 0,
    warehouse_id BIGINT        REFERENCES warehouses(id),
    location_id  BIGINT        REFERENCES storage_locations(id),
    notes        TEXT,
    created_at   TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_production_outputs_order_id  ON production_outputs(order_id);
CREATE INDEX IF NOT EXISTS idx_production_outputs_tenant_id ON production_outputs(tenant_id);

CREATE TRIGGER update_production_outputs_updated_at
    BEFORE UPDATE ON production_outputs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS production_resources (
    id             BIGSERIAL PRIMARY KEY,
    order_id       BIGINT        NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    tenant_id      BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    line_number    INT           NOT NULL DEFAULT 1,
    resource_type  VARCHAR(20)   NOT NULL
                       CHECK (resource_type IN ('MATERIAL','LABOR','OVERHEAD')),
    product_id     BIGINT        REFERENCES products(id),
    work_center_id BIGINT        REFERENCES work_centers(id),
    description    VARCHAR(500),
    quantity       NUMERIC(18,4) NOT NULL DEFAULT 0,
    uom_id         BIGINT        REFERENCES units_of_measure(id),
    unit_cost      NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost     NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes          TEXT,
    created_at     TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_production_resources_order_id  ON production_resources(order_id);
CREATE INDEX IF NOT EXISTS idx_production_resources_tenant_id ON production_resources(tenant_id);

CREATE TRIGGER update_production_resources_updated_at
    BEFORE UPDATE ON production_resources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Post Costs ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS post_costs (
    id                       BIGSERIAL PRIMARY KEY,
    tenant_id                BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code                     VARCHAR(50)   NOT NULL,
    order_id                 BIGINT        NOT NULL REFERENCES production_orders(id),
    estimate_id              BIGINT        REFERENCES cost_estimates(id),
    actual_material_cost     NUMERIC(18,4) NOT NULL DEFAULT 0,
    actual_resource_cost     NUMERIC(18,4) NOT NULL DEFAULT 0,
    actual_overhead_cost     NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_actual_cost        NUMERIC(18,4) NOT NULL DEFAULT 0,
    estimated_material_cost  NUMERIC(18,4) NOT NULL DEFAULT 0,
    estimated_resource_cost  NUMERIC(18,4) NOT NULL DEFAULT 0,
    estimated_overhead_cost  NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_estimated_cost     NUMERIC(18,4) NOT NULL DEFAULT 0,
    variance_amount          NUMERIC(18,4) NOT NULL DEFAULT 0,
    variance_pct             NUMERIC(10,4) NOT NULL DEFAULT 0,
    status                   VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                                 CHECK (status IN ('DRAFT','FINALIZED')),
    notes                    TEXT,
    finalized_by             BIGINT,
    finalized_at             TIMESTAMP,
    created_by               BIGINT,
    created_at               TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMP,
    CONSTRAINT uidx_post_costs_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_post_costs_tenant_id   ON post_costs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_post_costs_status      ON post_costs(status);
CREATE INDEX IF NOT EXISTS idx_post_costs_order_id    ON post_costs(order_id);
CREATE INDEX IF NOT EXISTS idx_post_costs_estimate_id ON post_costs(estimate_id);
CREATE INDEX IF NOT EXISTS idx_post_costs_deleted_at  ON post_costs(deleted_at);

CREATE TRIGGER update_post_costs_updated_at
    BEFORE UPDATE ON post_costs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Default Document Sequences ────────────────────────────────────────────────

INSERT INTO document_sequences (tenant_id, document_type, prefix, next_number, padding, suffix)
SELECT t.id, v.doc_type, v.prefix, 1, 5, ''
FROM tenants t
CROSS JOIN (VALUES
    ('COST_ESTIMATE',    'CE-'),
    ('PRODUCTION_PLAN',  'PP-'),
    ('PRODUCTION_ORDER', 'MO-'),
    ('POST_COST',        'PC-')
) AS v(doc_type, prefix)
WHERE NOT EXISTS (
    SELECT 1 FROM document_sequences ds
    WHERE ds.tenant_id = t.id AND ds.document_type = v.doc_type
);
