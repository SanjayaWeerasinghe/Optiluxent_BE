-- Inventory module: stock_balances, material_requests, goods_transfers,
--                   goods_issues, stock_adjustments, quality_checks

-- ── Stock Balances ────────────────────────────────────────────────────────────
-- Maintained in real-time via UPSERT on every stock-movement transaction.
-- NULLS NOT DISTINCT so (product, NULL variant, warehouse, NULL location)
-- resolves to a single row rather than always inserting a new one.

CREATE TABLE IF NOT EXISTS stock_balances (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    product_id    BIGINT        NOT NULL REFERENCES products(id),
    variant_id    BIGINT        REFERENCES product_variants(id),
    warehouse_id  BIGINT        NOT NULL REFERENCES warehouses(id),
    location_id   BIGINT        REFERENCES storage_locations(id),
    quantity      NUMERIC(18,4) NOT NULL DEFAULT 0,
    reserved_qty  NUMERIC(18,4) NOT NULL DEFAULT 0,
    updated_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_stock_balances
        UNIQUE NULLS NOT DISTINCT (tenant_id, product_id, variant_id, warehouse_id, location_id)
);

CREATE INDEX IF NOT EXISTS idx_stock_balances_tenant      ON stock_balances(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stock_balances_product     ON stock_balances(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_balances_warehouse   ON stock_balances(warehouse_id);

-- ── Material Requests ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS material_requests (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(50)  NOT NULL,
    requested_by    BIGINT       NOT NULL,
    department_id   BIGINT       REFERENCES departments(id),
    warehouse_id    BIGINT       NOT NULL REFERENCES warehouses(id),
    needed_date     DATE         NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'DRAFT'
                        CHECK (status IN ('DRAFT','APPROVED','REJECTED','FULFILLED','PARTIAL')),
    notes           TEXT,
    approved_by     BIGINT,
    approved_at     TIMESTAMP,
    rejected_by     BIGINT,
    rejected_at     TIMESTAMP,
    reject_reason   TEXT,
    created_by      BIGINT       NOT NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_material_requests_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_mr_tenant_id ON material_requests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mr_status    ON material_requests(status);

CREATE TRIGGER update_material_requests_updated_at
    BEFORE UPDATE ON material_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS material_request_lines (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    mr_id          BIGINT        NOT NULL REFERENCES material_requests(id) ON DELETE CASCADE,
    line_number    INT           NOT NULL DEFAULT 1,
    product_id     BIGINT        NOT NULL REFERENCES products(id),
    variant_id     BIGINT        REFERENCES product_variants(id),
    uom_id         BIGINT        NOT NULL REFERENCES units_of_measure(id),
    requested_qty  NUMERIC(18,4) NOT NULL,
    issued_qty     NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes          TEXT
);

CREATE INDEX IF NOT EXISTS idx_mr_lines_mr_id ON material_request_lines(mr_id);

-- ── Goods Transfers ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS goods_transfers (
    id                BIGSERIAL PRIMARY KEY,
    tenant_id         BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code              VARCHAR(50)  NOT NULL,
    from_warehouse_id BIGINT       NOT NULL REFERENCES warehouses(id),
    to_warehouse_id   BIGINT       NOT NULL REFERENCES warehouses(id),
    transfer_date     DATE         NOT NULL,
    status            VARCHAR(20)  NOT NULL DEFAULT 'DRAFT'
                          CHECK (status IN ('DRAFT','IN_TRANSIT','RECEIVED','CANCELLED')),
    notes             TEXT,
    sent_by           BIGINT,
    sent_at           TIMESTAMP,
    received_by       BIGINT,
    received_at       TIMESTAMP,
    created_by        BIGINT       NOT NULL,
    created_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_goods_transfers_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_gt_tenant_id ON goods_transfers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gt_status    ON goods_transfers(status);

CREATE TRIGGER update_goods_transfers_updated_at
    BEFORE UPDATE ON goods_transfers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS goods_transfer_lines (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    transfer_id      BIGINT        NOT NULL REFERENCES goods_transfers(id) ON DELETE CASCADE,
    line_number      INT           NOT NULL DEFAULT 1,
    product_id       BIGINT        NOT NULL REFERENCES products(id),
    variant_id       BIGINT        REFERENCES product_variants(id),
    from_location_id BIGINT        REFERENCES storage_locations(id),
    to_location_id   BIGINT        REFERENCES storage_locations(id),
    uom_id           BIGINT        NOT NULL REFERENCES units_of_measure(id),
    quantity         NUMERIC(18,4) NOT NULL,
    received_qty     NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes            TEXT
);

CREATE INDEX IF NOT EXISTS idx_gt_lines_transfer_id ON goods_transfer_lines(transfer_id);

-- ── Goods Issues ──────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS goods_issues (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code           VARCHAR(50)  NOT NULL,
    issue_date     DATE         NOT NULL,
    warehouse_id   BIGINT       NOT NULL REFERENCES warehouses(id),
    issue_reason   VARCHAR(50)  NOT NULL
                       CHECK (issue_reason IN ('PRODUCTION','SALE','EXPENSE','DAMAGE','ADJUSTMENT','QC_REJECTION','OTHER')),
    reference_type VARCHAR(50),
    reference_id   BIGINT,
    status         VARCHAR(20)  NOT NULL DEFAULT 'DRAFT'
                       CHECK (status IN ('DRAFT','CONFIRMED')),
    notes          TEXT,
    created_by     BIGINT       NOT NULL,
    confirmed_by   BIGINT,
    confirmed_at   TIMESTAMP,
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_goods_issues_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_gi_tenant_id ON goods_issues(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gi_status    ON goods_issues(status);
CREATE INDEX IF NOT EXISTS idx_gi_reason    ON goods_issues(issue_reason);

CREATE TRIGGER update_goods_issues_updated_at
    BEFORE UPDATE ON goods_issues
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS goods_issue_lines (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    issue_id    BIGINT        NOT NULL REFERENCES goods_issues(id) ON DELETE CASCADE,
    line_number INT           NOT NULL DEFAULT 1,
    product_id  BIGINT        NOT NULL REFERENCES products(id),
    variant_id  BIGINT        REFERENCES product_variants(id),
    location_id BIGINT        REFERENCES storage_locations(id),
    uom_id      BIGINT        NOT NULL REFERENCES units_of_measure(id),
    quantity    NUMERIC(18,4) NOT NULL,
    unit_cost   NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost  NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes       TEXT
);

CREATE INDEX IF NOT EXISTS idx_gi_lines_issue_id ON goods_issue_lines(issue_id);

-- ── Stock Adjustments ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS stock_adjustments (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code             VARCHAR(50)  NOT NULL,
    adjustment_date  DATE         NOT NULL,
    warehouse_id     BIGINT       NOT NULL REFERENCES warehouses(id),
    adjust_reason    VARCHAR(50)  NOT NULL
                         CHECK (adjust_reason IN ('STOCKTAKE','DAMAGE','EXPIRY','WRITE_OFF','CORRECTION','OTHER')),
    status           VARCHAR(20)  NOT NULL DEFAULT 'DRAFT'
                         CHECK (status IN ('DRAFT','CONFIRMED')),
    notes            TEXT,
    created_by       BIGINT       NOT NULL,
    confirmed_by     BIGINT,
    confirmed_at     TIMESTAMP,
    created_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_stock_adjustments_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_sa_tenant_id ON stock_adjustments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sa_status    ON stock_adjustments(status);

CREATE TRIGGER update_stock_adjustments_updated_at
    BEFORE UPDATE ON stock_adjustments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS stock_adjustment_lines (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    adjustment_id BIGINT        NOT NULL REFERENCES stock_adjustments(id) ON DELETE CASCADE,
    line_number   INT           NOT NULL DEFAULT 1,
    product_id    BIGINT        NOT NULL REFERENCES products(id),
    variant_id    BIGINT        REFERENCES product_variants(id),
    location_id   BIGINT        REFERENCES storage_locations(id),
    uom_id        BIGINT        NOT NULL REFERENCES units_of_measure(id),
    quantity      NUMERIC(18,4) NOT NULL,
    unit_cost     NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes         TEXT
);

CREATE INDEX IF NOT EXISTS idx_sa_lines_adjustment_id ON stock_adjustment_lines(adjustment_id);

-- ── Quality Checks ────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS quality_checks (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code           VARCHAR(50)  NOT NULL,
    reference_type VARCHAR(50),
    reference_id   BIGINT,
    warehouse_id   BIGINT       NOT NULL REFERENCES warehouses(id),
    check_date     DATE         NOT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'PENDING'
                       CHECK (status IN ('PENDING','IN_PROGRESS','PASSED','FAILED','PARTIAL')),
    inspector_id   BIGINT,
    notes          TEXT,
    created_by     BIGINT       NOT NULL,
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_quality_checks_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_qc_tenant_id ON quality_checks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_qc_status    ON quality_checks(status);

CREATE TRIGGER update_quality_checks_updated_at
    BEFORE UPDATE ON quality_checks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS quality_check_lines (
    id               BIGSERIAL PRIMARY KEY,
    tenant_id        BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    check_id         BIGINT        NOT NULL REFERENCES quality_checks(id) ON DELETE CASCADE,
    line_number      INT           NOT NULL DEFAULT 1,
    product_id       BIGINT        NOT NULL REFERENCES products(id),
    variant_id       BIGINT        REFERENCES product_variants(id),
    qty_checked      NUMERIC(18,4) NOT NULL,
    qty_passed       NUMERIC(18,4) NOT NULL DEFAULT 0,
    qty_failed       NUMERIC(18,4) NOT NULL DEFAULT 0,
    result           VARCHAR(20)   NOT NULL DEFAULT 'PENDING'
                         CHECK (result IN ('PENDING','PASSED','FAILED','PARTIAL')),
    rejection_reason TEXT,
    notes            TEXT
);

CREATE INDEX IF NOT EXISTS idx_qc_lines_check_id ON quality_check_lines(check_id);

-- ── Document Sequences ────────────────────────────────────────────────────────

INSERT INTO document_sequences (tenant_id, document_type, prefix, next_number, padding, suffix)
SELECT t.id, v.doc_type, v.prefix, 1, 5, ''
FROM tenants t
CROSS JOIN (VALUES
    ('MATERIAL_REQUEST', 'MR'),
    ('GOODS_TRANSFER',   'GT'),
    ('GOODS_ISSUE',      'GI'),
    ('STOCK_ADJUSTMENT', 'SA'),
    ('QUALITY_CHECK',    'QC')
) AS v(doc_type, prefix)
WHERE NOT EXISTS (
    SELECT 1 FROM document_sequences ds
    WHERE ds.tenant_id = t.id AND ds.document_type = v.doc_type
);
