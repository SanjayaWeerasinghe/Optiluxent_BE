-- Procurement module: purchase_requests, purchase_orders, goods_receipts, purchase_invoices

-- ── Purchase Requests ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS purchase_requests (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(50)  NOT NULL,
    request_date    DATE         NOT NULL DEFAULT CURRENT_DATE,
    required_date   DATE,
    requested_by    BIGINT       REFERENCES employees(id),
    department_id   BIGINT       REFERENCES departments(id),
    status          VARCHAR(20)  NOT NULL DEFAULT 'DRAFT'
                        CHECK (status IN ('DRAFT','PENDING_APPROVAL','APPROVED','REJECTED','CANCELLED')),
    notes           TEXT,
    approved_by     BIGINT,
    approved_at     TIMESTAMP,
    created_by      BIGINT,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,
    CONSTRAINT uidx_purchase_requests_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_pr_tenant_id   ON purchase_requests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pr_status      ON purchase_requests(status);
CREATE INDEX IF NOT EXISTS idx_pr_deleted_at  ON purchase_requests(deleted_at);

CREATE TRIGGER update_purchase_requests_updated_at
    BEFORE UPDATE ON purchase_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS purchase_request_lines (
    id              BIGSERIAL PRIMARY KEY,
    pr_id           BIGINT        NOT NULL REFERENCES purchase_requests(id) ON DELETE CASCADE,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    line_number     INT           NOT NULL DEFAULT 1,
    product_id      BIGINT        NOT NULL REFERENCES products(id),
    variant_id      BIGINT        REFERENCES product_variants(id),
    description     VARCHAR(500),
    quantity        NUMERIC(18,4) NOT NULL,
    uom_id          BIGINT        NOT NULL REFERENCES units_of_measure(id),
    estimated_price NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pr_lines_pr_id ON purchase_request_lines(pr_id);

CREATE TRIGGER update_purchase_request_lines_updated_at
    BEFORE UPDATE ON purchase_request_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Purchase Orders ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS purchase_orders (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(50)   NOT NULL,
    supplier_id     BIGINT        NOT NULL REFERENCES parties(id),
    pr_id           BIGINT        REFERENCES purchase_requests(id),
    order_date      DATE          NOT NULL DEFAULT CURRENT_DATE,
    expected_date   DATE,
    currency_id     BIGINT        NOT NULL REFERENCES currencies(id),
    exchange_rate   NUMERIC(18,6) NOT NULL DEFAULT 1,
    payment_term_id BIGINT        REFERENCES payment_terms(id),
    warehouse_id    BIGINT        NOT NULL REFERENCES warehouses(id),
    status          VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                        CHECK (status IN ('DRAFT','CONFIRMED','PARTIAL','RECEIVED','CANCELLED')),
    subtotal        NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes           TEXT,
    confirmed_by    BIGINT,
    confirmed_at    TIMESTAMP,
    created_by      BIGINT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,
    CONSTRAINT uidx_purchase_orders_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_po_tenant_id   ON purchase_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_po_supplier_id ON purchase_orders(supplier_id);
CREATE INDEX IF NOT EXISTS idx_po_status      ON purchase_orders(status);
CREATE INDEX IF NOT EXISTS idx_po_deleted_at  ON purchase_orders(deleted_at);

CREATE TRIGGER update_purchase_orders_updated_at
    BEFORE UPDATE ON purchase_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS purchase_order_lines (
    id              BIGSERIAL PRIMARY KEY,
    po_id           BIGINT        NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    line_number     INT           NOT NULL DEFAULT 1,
    product_id      BIGINT        NOT NULL REFERENCES products(id),
    variant_id      BIGINT        REFERENCES product_variants(id),
    description     VARCHAR(500),
    quantity        NUMERIC(18,4) NOT NULL,
    uom_id          BIGINT        NOT NULL REFERENCES units_of_measure(id),
    unit_price      NUMERIC(18,4) NOT NULL DEFAULT 0,
    discount_pct    NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tax_code_id     BIGINT        REFERENCES tax_codes(id),
    tax_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    line_total      NUMERIC(18,2) NOT NULL DEFAULT 0,
    received_qty    NUMERIC(18,4) NOT NULL DEFAULT 0,
    billed_qty      NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_po_lines_po_id ON purchase_order_lines(po_id);

CREATE TRIGGER update_purchase_order_lines_updated_at
    BEFORE UPDATE ON purchase_order_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Goods Receipts ────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS goods_receipts (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(50)   NOT NULL,
    po_id           BIGINT        REFERENCES purchase_orders(id),
    supplier_id     BIGINT        NOT NULL REFERENCES parties(id),
    receipt_date    DATE          NOT NULL DEFAULT CURRENT_DATE,
    warehouse_id    BIGINT        NOT NULL REFERENCES warehouses(id),
    status          VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                        CHECK (status IN ('DRAFT','CONFIRMED','CANCELLED')),
    notes           TEXT,
    confirmed_by    BIGINT,
    confirmed_at    TIMESTAMP,
    created_by      BIGINT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,
    CONSTRAINT uidx_goods_receipts_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_grn_tenant_id   ON goods_receipts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_grn_po_id       ON goods_receipts(po_id);
CREATE INDEX IF NOT EXISTS idx_grn_status      ON goods_receipts(status);
CREATE INDEX IF NOT EXISTS idx_grn_deleted_at  ON goods_receipts(deleted_at);

CREATE TRIGGER update_goods_receipts_updated_at
    BEFORE UPDATE ON goods_receipts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS goods_receipt_lines (
    id              BIGSERIAL PRIMARY KEY,
    grn_id          BIGINT        NOT NULL REFERENCES goods_receipts(id) ON DELETE CASCADE,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    po_line_id      BIGINT        REFERENCES purchase_order_lines(id),
    line_number     INT           NOT NULL DEFAULT 1,
    product_id      BIGINT        NOT NULL REFERENCES products(id),
    variant_id      BIGINT        REFERENCES product_variants(id),
    quantity        NUMERIC(18,4) NOT NULL,
    uom_id          BIGINT        NOT NULL REFERENCES units_of_measure(id),
    location_id     BIGINT        REFERENCES storage_locations(id),
    unit_cost       NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost      NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_grn_lines_grn_id ON goods_receipt_lines(grn_id);

CREATE TRIGGER update_goods_receipt_lines_updated_at
    BEFORE UPDATE ON goods_receipt_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Purchase Invoices ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS purchase_invoices (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(50)   NOT NULL,
    supplier_id     BIGINT        NOT NULL REFERENCES parties(id),
    grn_id          BIGINT        REFERENCES goods_receipts(id),
    po_id           BIGINT        REFERENCES purchase_orders(id),
    invoice_date    DATE          NOT NULL DEFAULT CURRENT_DATE,
    due_date        DATE,
    currency_id     BIGINT        NOT NULL REFERENCES currencies(id),
    exchange_rate   NUMERIC(18,6) NOT NULL DEFAULT 1,
    payment_term_id BIGINT        REFERENCES payment_terms(id),
    status          VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                        CHECK (status IN ('DRAFT','POSTED','PARTIAL','PAID','CANCELLED')),
    subtotal        NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    paid_amount     NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes           TEXT,
    posted_by       BIGINT,
    posted_at       TIMESTAMP,
    created_by      BIGINT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,
    CONSTRAINT uidx_purchase_invoices_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_pinv_tenant_id   ON purchase_invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pinv_supplier_id ON purchase_invoices(supplier_id);
CREATE INDEX IF NOT EXISTS idx_pinv_status      ON purchase_invoices(status);
CREATE INDEX IF NOT EXISTS idx_pinv_deleted_at  ON purchase_invoices(deleted_at);

CREATE TRIGGER update_purchase_invoices_updated_at
    BEFORE UPDATE ON purchase_invoices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS purchase_invoice_lines (
    id              BIGSERIAL PRIMARY KEY,
    invoice_id      BIGINT        NOT NULL REFERENCES purchase_invoices(id) ON DELETE CASCADE,
    tenant_id       BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    grn_line_id     BIGINT        REFERENCES goods_receipt_lines(id),
    line_number     INT           NOT NULL DEFAULT 1,
    product_id      BIGINT        NOT NULL REFERENCES products(id),
    variant_id      BIGINT        REFERENCES product_variants(id),
    description     VARCHAR(500),
    quantity        NUMERIC(18,4) NOT NULL,
    uom_id          BIGINT        NOT NULL REFERENCES units_of_measure(id),
    unit_price      NUMERIC(18,4) NOT NULL DEFAULT 0,
    discount_pct    NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tax_code_id     BIGINT        REFERENCES tax_codes(id),
    tax_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    line_total      NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pinv_lines_invoice_id ON purchase_invoice_lines(invoice_id);

CREATE TRIGGER update_purchase_invoice_lines_updated_at
    BEFORE UPDATE ON purchase_invoice_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Default Document Sequences ────────────────────────────────────────────────

INSERT INTO document_sequences (tenant_id, document_type, prefix, next_number, padding, suffix)
SELECT t.id, v.doc_type, v.prefix, 1, 5, ''
FROM tenants t
CROSS JOIN (VALUES
    ('PURCHASE_REQUEST', 'PR'),
    ('GOODS_RECEIPT',    'GRN'),
    ('PURCHASE_INVOICE', 'PINV')
) AS v(doc_type, prefix)
WHERE NOT EXISTS (
    SELECT 1 FROM document_sequences ds
    WHERE ds.tenant_id = t.id AND ds.document_type = v.doc_type
);

-- PURCHASE_ORDER may already exist from seed data; insert only if missing
INSERT INTO document_sequences (tenant_id, document_type, prefix, next_number, padding, suffix)
SELECT t.id, 'PURCHASE_ORDER', 'PO', 1, 5, ''
FROM tenants t
WHERE NOT EXISTS (
    SELECT 1 FROM document_sequences ds
    WHERE ds.tenant_id = t.id AND ds.document_type = 'PURCHASE_ORDER'
);
