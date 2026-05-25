-- Sales module: sales_orders, delivery_orders, sales_invoices

-- ── Sales Orders ──────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS sales_orders (
    id                    BIGSERIAL PRIMARY KEY,
    tenant_id             BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code                  VARCHAR(50)   NOT NULL,
    customer_id           BIGINT        NOT NULL REFERENCES parties(id),
    order_date            DATE          NOT NULL DEFAULT CURRENT_DATE,
    expected_delivery_date DATE,
    currency_id           BIGINT        NOT NULL REFERENCES currencies(id),
    exchange_rate         NUMERIC(18,6) NOT NULL DEFAULT 1,
    payment_term_id       BIGINT        REFERENCES payment_terms(id),
    warehouse_id          BIGINT        NOT NULL REFERENCES warehouses(id),
    status                VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                              CHECK (status IN ('DRAFT','CONFIRMED','PARTIAL','DELIVERED','CANCELLED')),
    subtotal              NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_amount            NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_amount       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount          NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes                 TEXT,
    confirmed_by          BIGINT,
    confirmed_at          TIMESTAMP,
    created_by            BIGINT,
    created_at            TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMP,
    CONSTRAINT uidx_sales_orders_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_so_tenant_id    ON sales_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_so_customer_id  ON sales_orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_so_status       ON sales_orders(status);
CREATE INDEX IF NOT EXISTS idx_so_deleted_at   ON sales_orders(deleted_at);

CREATE TRIGGER update_sales_orders_updated_at
    BEFORE UPDATE ON sales_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS sales_order_lines (
    id            BIGSERIAL PRIMARY KEY,
    so_id         BIGINT        NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    tenant_id     BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    line_number   INT           NOT NULL DEFAULT 1,
    product_id    BIGINT        NOT NULL REFERENCES products(id),
    variant_id    BIGINT        REFERENCES product_variants(id),
    description   VARCHAR(500),
    quantity      NUMERIC(18,4) NOT NULL,
    uom_id        BIGINT        NOT NULL REFERENCES units_of_measure(id),
    unit_price    NUMERIC(18,4) NOT NULL DEFAULT 0,
    discount_pct  NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tax_code_id   BIGINT        REFERENCES tax_codes(id),
    tax_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    line_total    NUMERIC(18,2) NOT NULL DEFAULT 0,
    delivered_qty NUMERIC(18,4) NOT NULL DEFAULT 0,
    invoiced_qty  NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes         TEXT,
    created_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_so_lines_so_id ON sales_order_lines(so_id);

CREATE TRIGGER update_sales_order_lines_updated_at
    BEFORE UPDATE ON sales_order_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Delivery Orders ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS delivery_orders (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code          VARCHAR(50)  NOT NULL,
    so_id         BIGINT       REFERENCES sales_orders(id),
    customer_id   BIGINT       NOT NULL REFERENCES parties(id),
    delivery_date DATE         NOT NULL DEFAULT CURRENT_DATE,
    warehouse_id  BIGINT       NOT NULL REFERENCES warehouses(id),
    status        VARCHAR(20)  NOT NULL DEFAULT 'DRAFT'
                      CHECK (status IN ('DRAFT','CONFIRMED','CANCELLED')),
    notes         TEXT,
    confirmed_by  BIGINT,
    confirmed_at  TIMESTAMP,
    created_by    BIGINT,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP,
    CONSTRAINT uidx_delivery_orders_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_do_tenant_id    ON delivery_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_do_so_id        ON delivery_orders(so_id);
CREATE INDEX IF NOT EXISTS idx_do_customer_id  ON delivery_orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_do_status       ON delivery_orders(status);
CREATE INDEX IF NOT EXISTS idx_do_deleted_at   ON delivery_orders(deleted_at);

CREATE TRIGGER update_delivery_orders_updated_at
    BEFORE UPDATE ON delivery_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS delivery_order_lines (
    id          BIGSERIAL PRIMARY KEY,
    do_id       BIGINT        NOT NULL REFERENCES delivery_orders(id) ON DELETE CASCADE,
    tenant_id   BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    so_line_id  BIGINT        REFERENCES sales_order_lines(id),
    line_number INT           NOT NULL DEFAULT 1,
    product_id  BIGINT        NOT NULL REFERENCES products(id),
    variant_id  BIGINT        REFERENCES product_variants(id),
    location_id BIGINT        REFERENCES storage_locations(id),
    uom_id      BIGINT        NOT NULL REFERENCES units_of_measure(id),
    quantity    NUMERIC(18,4) NOT NULL,
    unit_cost   NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost  NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes       TEXT,
    created_at  TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_do_lines_do_id ON delivery_order_lines(do_id);

CREATE TRIGGER update_delivery_order_lines_updated_at
    BEFORE UPDATE ON delivery_order_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Sales Invoices ────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS sales_invoices (
    id                 BIGSERIAL PRIMARY KEY,
    tenant_id          BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code               VARCHAR(50)   NOT NULL,
    customer_id        BIGINT        NOT NULL REFERENCES parties(id),
    so_id              BIGINT        NOT NULL REFERENCES sales_orders(id),
    invoice_date       DATE          NOT NULL DEFAULT CURRENT_DATE,
    due_date           DATE,
    customer_po_number VARCHAR(100),
    currency_id        BIGINT        NOT NULL REFERENCES currencies(id),
    exchange_rate      NUMERIC(18,6) NOT NULL DEFAULT 1,
    payment_term_id    BIGINT        REFERENCES payment_terms(id),
    status             VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                           CHECK (status IN ('DRAFT','POSTED','PARTIAL','PAID','CANCELLED')),
    subtotal           NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_amount         NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount       NUMERIC(18,2) NOT NULL DEFAULT 0,
    paid_amount        NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes              TEXT,
    posted_by          BIGINT,
    posted_at          TIMESTAMP,
    created_by         BIGINT,
    created_at         TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMP,
    CONSTRAINT uidx_sales_invoices_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_sinv_tenant_id    ON sales_invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sinv_customer_id  ON sales_invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_sinv_so_id        ON sales_invoices(so_id);
CREATE INDEX IF NOT EXISTS idx_sinv_status       ON sales_invoices(status);
CREATE INDEX IF NOT EXISTS idx_sinv_deleted_at   ON sales_invoices(deleted_at);

CREATE TRIGGER update_sales_invoices_updated_at
    BEFORE UPDATE ON sales_invoices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS sales_invoice_lines (
    id           BIGSERIAL PRIMARY KEY,
    invoice_id   BIGINT        NOT NULL REFERENCES sales_invoices(id) ON DELETE CASCADE,
    tenant_id    BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    do_line_id   BIGINT        REFERENCES delivery_order_lines(id),
    line_number  INT           NOT NULL DEFAULT 1,
    product_id   BIGINT        NOT NULL REFERENCES products(id),
    variant_id   BIGINT        REFERENCES product_variants(id),
    description  VARCHAR(500),
    quantity     NUMERIC(18,4) NOT NULL,
    uom_id       BIGINT        NOT NULL REFERENCES units_of_measure(id),
    unit_price   NUMERIC(18,4) NOT NULL DEFAULT 0,
    discount_pct NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tax_code_id  BIGINT        REFERENCES tax_codes(id),
    tax_amount   NUMERIC(18,2) NOT NULL DEFAULT 0,
    line_total   NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes        TEXT,
    created_at   TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sinv_lines_invoice_id ON sales_invoice_lines(invoice_id);

CREATE TRIGGER update_sales_invoice_lines_updated_at
    BEFORE UPDATE ON sales_invoice_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ── Default Document Sequences ────────────────────────────────────────────────

INSERT INTO document_sequences (tenant_id, document_type, prefix, next_number, padding, suffix)
SELECT t.id, v.doc_type, v.prefix, 1, 5, ''
FROM tenants t
CROSS JOIN (VALUES
    ('SALES_ORDER',    'SO'),
    ('DELIVERY_ORDER', 'DO'),
    ('SALES_INVOICE',  'SI')
) AS v(doc_type, prefix)
WHERE NOT EXISTS (
    SELECT 1 FROM document_sequences ds
    WHERE ds.tenant_id = t.id AND ds.document_type = v.doc_type
);
