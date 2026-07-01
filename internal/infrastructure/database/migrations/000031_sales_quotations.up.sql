-- Sales Quotations — precedes the Sales Order in the sales lifecycle.
-- Flow: DRAFT → SENT → ACCEPTED (auto-creates SO with sq_id back-ref)
--                    → REJECTED
--                    → EXPIRED (once valid_until passes without acceptance)
--                    → CANCELLED (from DRAFT or SENT)

CREATE TABLE IF NOT EXISTS sales_quotations (
    id                    BIGSERIAL PRIMARY KEY,
    tenant_id             BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code                  VARCHAR(50)   NOT NULL,
    customer_id           BIGINT        NOT NULL REFERENCES parties(id),
    quotation_date        DATE          NOT NULL DEFAULT CURRENT_DATE,
    valid_until           DATE,
    currency_id           BIGINT        NOT NULL REFERENCES currencies(id),
    exchange_rate         NUMERIC(18,6) NOT NULL DEFAULT 1,
    payment_term_id       BIGINT        REFERENCES payment_terms(id),
    warehouse_id          BIGINT        NOT NULL REFERENCES warehouses(id),
    status                VARCHAR(20)   NOT NULL DEFAULT 'DRAFT'
                              CHECK (status IN ('DRAFT','SENT','ACCEPTED','REJECTED','EXPIRED','CANCELLED')),
    subtotal              NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_amount            NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_amount       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount          NUMERIC(18,2) NOT NULL DEFAULT 0,
    customer_reference    VARCHAR(100),
    terms_and_conditions  TEXT,
    notes                 TEXT,
    -- Set when the quotation is ACCEPTED and an SO is auto-created; forms the
    -- back-reference for reports and drill-through.
    converted_so_id       BIGINT        REFERENCES sales_orders(id),
    accepted_by           BIGINT,
    accepted_at           TIMESTAMP,
    rejected_by           BIGINT,
    created_by            BIGINT,
    created_at            TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMP,
    CONSTRAINT uidx_sales_quotations_tenant_code UNIQUE (tenant_id, code)
);

CREATE INDEX IF NOT EXISTS idx_sq_tenant_id   ON sales_quotations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sq_customer_id ON sales_quotations(customer_id);
CREATE INDEX IF NOT EXISTS idx_sq_status      ON sales_quotations(status);
CREATE INDEX IF NOT EXISTS idx_sq_deleted_at  ON sales_quotations(deleted_at);

CREATE TRIGGER update_sales_quotations_updated_at
    BEFORE UPDATE ON sales_quotations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS sales_quotation_lines (
    id           BIGSERIAL PRIMARY KEY,
    sq_id        BIGINT        NOT NULL REFERENCES sales_quotations(id) ON DELETE CASCADE,
    tenant_id    BIGINT        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
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

CREATE INDEX IF NOT EXISTS idx_sq_lines_sq_id ON sales_quotation_lines(sq_id);

CREATE TRIGGER update_sales_quotation_lines_updated_at
    BEFORE UPDATE ON sales_quotation_lines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Back-reference from SO → SQ so accepted quotations trace forward.
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS sq_id BIGINT REFERENCES sales_quotations(id);
CREATE INDEX IF NOT EXISTS idx_so_sq_id ON sales_orders(sq_id);

-- Doc sequence for quotations — prefix SQ, five-digit padding, per-tenant.
INSERT INTO document_sequences (tenant_id, document_type, prefix, next_number, padding, suffix)
SELECT t.id, 'SALES_QUOTATION', 'SQ', 1, 5, ''
FROM tenants t
WHERE NOT EXISTS (
    SELECT 1 FROM document_sequences ds
    WHERE ds.tenant_id = t.id AND ds.document_type = 'SALES_QUOTATION'
);
