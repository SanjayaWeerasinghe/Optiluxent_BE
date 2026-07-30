-- Double-entry GL storage. One journal_entries header per source event
-- (SI post, PI post, payment). Balanced gl_lines child rows carry the
-- debit / credit / account / party split. Reversal-of pointer lets a
-- future "cancel invoice" flow write a mirror entry without deleting
-- history.

CREATE TABLE IF NOT EXISTS journal_entries (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code           VARCHAR(50)  NOT NULL,
    post_date      DATE         NOT NULL,
    narration      VARCHAR(500) NOT NULL,
    source_type    VARCHAR(30)  NOT NULL,
    source_id      BIGINT       NOT NULL,
    is_posted      BOOLEAN      NOT NULL DEFAULT TRUE,
    is_reversal_of BIGINT       REFERENCES journal_entries(id),
    created_by     BIGINT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uidx_journal_entries_tenant_code UNIQUE (tenant_id, code)
);
CREATE INDEX IF NOT EXISTS idx_je_source ON journal_entries (source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_je_date   ON journal_entries (post_date);

CREATE TABLE IF NOT EXISTS gl_lines (
    id             BIGSERIAL PRIMARY KEY,
    tenant_id      BIGINT       NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    journal_id     BIGINT       NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    line_number    INT          NOT NULL,
    account_id     BIGINT       NOT NULL REFERENCES chart_of_accounts(id),
    party_id       BIGINT       REFERENCES parties(id),
    cost_center_id BIGINT       REFERENCES cost_centers(id),
    debit          NUMERIC(18,4) NOT NULL DEFAULT 0,
    credit         NUMERIC(18,4) NOT NULL DEFAULT 0,
    description    VARCHAR(500),
    -- Exactly one side per line: either debit > 0 XOR credit > 0.
    CONSTRAINT gl_lines_one_side CHECK (
        debit  >= 0 AND credit >= 0
        AND (debit  = 0 OR credit = 0)
        AND (debit  > 0 OR credit > 0)
    )
);
CREATE INDEX IF NOT EXISTS idx_gl_lines_journal ON gl_lines (journal_id);
CREATE INDEX IF NOT EXISTS idx_gl_lines_account ON gl_lines (account_id);
CREATE INDEX IF NOT EXISTS idx_gl_lines_party   ON gl_lines (party_id) WHERE party_id IS NOT NULL;

-- Sequence for JE codes.
INSERT INTO document_sequences (tenant_id, document_type, prefix, next_number, padding, suffix)
SELECT t.id, 'JOURNAL_ENTRY', 'JE-', 1, 5, ''
FROM tenants t
ON CONFLICT ON CONSTRAINT uidx_document_sequences_tenant_type DO NOTHING;
