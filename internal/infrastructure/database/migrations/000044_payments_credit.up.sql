-- Finance module foundation: a real payments log (currently paid_amount
-- is a scalar on the invoice with no method / bank / reference / date
-- history), a per-tenant settings row so the GL poster knows which
-- accounts to use, and a credit_type flag on parties so we can enforce
-- "cash-on-delivery" vs "credit-with-limit" at SO confirm time.

-- ── payments ────────────────────────────────────────────────────────────────
-- Unified for AR (customer pays us) + AP (we pay a supplier). Every row
-- points at a single invoice; an invoice can accumulate many payment
-- rows and SUM(payments.amount) must equal invoices.paid_amount.
CREATE TABLE IF NOT EXISTS payments (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(50)  NOT NULL,
    direction       VARCHAR(10)  NOT NULL CHECK (direction IN ('INBOUND','OUTBOUND')),
    invoice_kind    VARCHAR(10)  NOT NULL CHECK (invoice_kind IN ('SI','PI')),
    invoice_id      BIGINT       NOT NULL,
    party_id        BIGINT       NOT NULL REFERENCES parties(id),
    amount          NUMERIC(18,4) NOT NULL,
    currency_id     BIGINT       NOT NULL REFERENCES currencies(id),
    exchange_rate   NUMERIC(18,6) NOT NULL DEFAULT 1,
    payment_date    DATE         NOT NULL,
    method          VARCHAR(20)  NOT NULL CHECK (method IN ('CASH','BANK_TRANSFER','CHEQUE','CARD','OTHER')),
    bank_account_id BIGINT       REFERENCES company_bank_accounts(id),
    reference_no    VARCHAR(100),
    notes           TEXT,
    journal_id      BIGINT,      -- filled by GL poster after the JE is inserted
    created_by      BIGINT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_payments_tenant_code UNIQUE (tenant_id, code)
);
CREATE INDEX IF NOT EXISTS idx_payments_invoice ON payments (invoice_kind, invoice_id);
CREATE INDEX IF NOT EXISTS idx_payments_party   ON payments (party_id);
CREATE INDEX IF NOT EXISTS idx_payments_date    ON payments (payment_date);

-- ── finance_settings ────────────────────────────────────────────────────────
-- Tenant-scoped defaults consumed by the GL poster. Nullable so a tenant
-- can bootstrap without picking accounts up-front; unset accounts skip
-- posting for that side of the entry (system still functions for
-- operational payment tracking).
CREATE TABLE IF NOT EXISTS finance_settings (
    tenant_id           BIGINT PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    ar_account_id       BIGINT REFERENCES chart_of_accounts(id),
    ap_account_id       BIGINT REFERENCES chart_of_accounts(id),
    cash_account_id     BIGINT REFERENCES chart_of_accounts(id),
    sales_revenue_id    BIGINT REFERENCES chart_of_accounts(id),
    purchase_expense_id BIGINT REFERENCES chart_of_accounts(id),
    tax_account_id      BIGINT REFERENCES chart_of_accounts(id),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── parties.credit_type ─────────────────────────────────────────────────────
-- CASH = must pay every prior invoice before we confirm the next SO.
-- CREDIT = allowed up to credit_limit; SO confirm blocks when exceeded.
ALTER TABLE parties
    ADD COLUMN IF NOT EXISTS credit_type VARCHAR(10) NOT NULL DEFAULT 'CREDIT'
        CHECK (credit_type IN ('CASH','CREDIT'));

-- ── document sequences for payment codes ────────────────────────────────────
INSERT INTO document_sequences (tenant_id, document_type, prefix, next_number, padding, suffix)
SELECT t.id, v.doc_type, v.prefix, 1, 5, ''
FROM tenants t
CROSS JOIN (VALUES
    ('PAYMENT_IN',  'RCPT'),
    ('PAYMENT_OUT', 'PAY')
) AS v(doc_type, prefix)
ON CONFLICT ON CONSTRAINT uidx_document_sequences_tenant_type DO NOTHING;
