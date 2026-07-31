package finance

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	// Document code
	NextCode(ctx context.Context, tenantID uint, docType string) (string, error)

	// Payments
	CreatePayment(ctx context.Context, p *Payment) error
	UpdatePayment(ctx context.Context, p *Payment) error
	GetPayment(ctx context.Context, tenantID, id uint) (*Payment, error)
	// ListPayments — `limit=0` = unbounded. Offset added for pagination.
	ListPayments(ctx context.Context, tenantID uint, invoiceKind string, invoiceID uint, partyID uint, direction string, limit, offset int) ([]Payment, error)
	CountPayments(ctx context.Context, tenantID uint, invoiceKind string, invoiceID uint, partyID uint, direction string) (int64, error)
	SumPaymentsByInvoice(ctx context.Context, tenantID uint, invoiceKind string, invoiceID uint) (float64, error)

	// Journal entries. `limit=0` = unbounded.
	CreateJournalEntry(ctx context.Context, je *JournalEntry) error
	GetJournalEntry(ctx context.Context, tenantID, id uint) (*JournalEntry, error)
	ListJournalEntries(ctx context.Context, tenantID uint, sourceType string, sourceID uint, limit, offset int) ([]JournalEntry, error)
	CountJournalEntries(ctx context.Context, tenantID uint, sourceType string, sourceID uint) (int64, error)

	// Settings
	GetSettings(ctx context.Context, tenantID uint) (*FinanceSettings, error)
	UpsertSettings(ctx context.Context, s *FinanceSettings) error

	// Aging
	AgingRows(ctx context.Context, tenantID uint, kind string, asOf string) ([]AgingRow, error)

	// Outstanding by party (SUM of open invoices).
	PartyOutstanding(ctx context.Context, tenantID, partyID uint, kind string) (float64, error)

	// Party credit type — thin passthrough so ConfirmSO can gate without
	// pulling the whole parties module.
	PartyCreditProfile(ctx context.Context, tenantID, partyID uint) (creditType string, creditLimit float64, err error)
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

// ── Document code (uses shared document_sequences) ──────────────────────────

func (r *dbRepository) NextCode(ctx context.Context, tenantID uint, docType string) (string, error) {
	var code string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var seq struct {
			Prefix     string
			NextNumber int
			Padding    int
			Suffix     string
		}
		row := tx.Raw(`
			SELECT prefix, next_number, padding, suffix
			FROM document_sequences
			WHERE tenant_id = ? AND document_type = ?
			FOR UPDATE`, tenantID, docType).Row()
		if err := row.Scan(&seq.Prefix, &seq.NextNumber, &seq.Padding, &seq.Suffix); err != nil {
			return fmt.Errorf("document sequence not found for %s: %w", docType, err)
		}
		code = fmt.Sprintf("%s%0*d%s", seq.Prefix, seq.Padding, seq.NextNumber, seq.Suffix)
		return tx.Exec(`UPDATE document_sequences SET next_number = next_number + 1 WHERE tenant_id = ? AND document_type = ?`,
			tenantID, docType).Error
	})
	return code, err
}

// ── Payments ────────────────────────────────────────────────────────────────

func (r *dbRepository) CreatePayment(ctx context.Context, p *Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *dbRepository) UpdatePayment(ctx context.Context, p *Payment) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *dbRepository) GetPayment(ctx context.Context, tenantID, id uint) (*Payment, error) {
	var p Payment
	return &p, r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&p).Error
}

// paymentFilterWhere — factored so ListPayments + CountPayments share filters.
func paymentFilterWhere(q *gorm.DB, tenantID uint, invoiceKind string, invoiceID, partyID uint, direction string) *gorm.DB {
	q = q.Where("tenant_id = ?", tenantID)
	if invoiceKind != "" {
		q = q.Where("invoice_kind = ?", invoiceKind)
	}
	if invoiceID > 0 {
		q = q.Where("invoice_id = ?", invoiceID)
	}
	if partyID > 0 {
		q = q.Where("party_id = ?", partyID)
	}
	if direction != "" {
		q = q.Where("direction = ?", direction)
	}
	return q
}

func (r *dbRepository) ListPayments(ctx context.Context, tenantID uint, invoiceKind string, invoiceID, partyID uint, direction string, limit, offset int) ([]Payment, error) {
	q := paymentFilterWhere(r.db.WithContext(ctx), tenantID, invoiceKind, invoiceID, partyID, direction).
		Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []Payment
	return rows, q.Find(&rows).Error
}

func (r *dbRepository) CountPayments(ctx context.Context, tenantID uint, invoiceKind string, invoiceID, partyID uint, direction string) (int64, error) {
	var n int64
	return n, paymentFilterWhere(r.db.WithContext(ctx).Model(&Payment{}), tenantID, invoiceKind, invoiceID, partyID, direction).Count(&n).Error
}

func (r *dbRepository) SumPaymentsByInvoice(ctx context.Context, tenantID uint, invoiceKind string, invoiceID uint) (float64, error) {
	var sum float64
	err := r.db.WithContext(ctx).
		Raw(`SELECT COALESCE(SUM(amount), 0) FROM payments WHERE tenant_id = ? AND invoice_kind = ? AND invoice_id = ?`,
			tenantID, invoiceKind, invoiceID).Row().Scan(&sum)
	return sum, err
}

// ── Journal entries ─────────────────────────────────────────────────────────

func (r *dbRepository) CreateJournalEntry(ctx context.Context, je *JournalEntry) error {
	// GORM's Create with a slice of Lines populates the FK automatically
	// via the foreignKey:JournalID association on the parent.
	return r.db.WithContext(ctx).Create(je).Error
}

func (r *dbRepository) GetJournalEntry(ctx context.Context, tenantID, id uint) (*JournalEntry, error) {
	var je JournalEntry
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&je).Error
	return &je, err
}

// jeFilterWhere — shared filter body for List + Count.
func jeFilterWhere(q *gorm.DB, tenantID uint, sourceType string, sourceID uint) *gorm.DB {
	q = q.Where("tenant_id = ?", tenantID)
	if sourceType != "" {
		q = q.Where("source_type = ?", sourceType)
	}
	if sourceID > 0 {
		q = q.Where("source_id = ?", sourceID)
	}
	return q
}

func (r *dbRepository) ListJournalEntries(ctx context.Context, tenantID uint, sourceType string, sourceID uint, limit, offset int) ([]JournalEntry, error) {
	q := jeFilterWhere(r.db.WithContext(ctx), tenantID, sourceType, sourceID).Order("post_date DESC, id DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []JournalEntry
	return rows, q.Find(&rows).Error
}

func (r *dbRepository) CountJournalEntries(ctx context.Context, tenantID uint, sourceType string, sourceID uint) (int64, error) {
	var n int64
	return n, jeFilterWhere(r.db.WithContext(ctx).Model(&JournalEntry{}), tenantID, sourceType, sourceID).Count(&n).Error
}

// ── Settings ────────────────────────────────────────────────────────────────

func (r *dbRepository) GetSettings(ctx context.Context, tenantID uint) (*FinanceSettings, error) {
	var s FinanceSettings
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		// Return an empty settings row so callers can uniformly nil-check
		// individual account IDs.
		return &FinanceSettings{TenantID: tenantID}, nil
	}
	return &s, err
}

func (r *dbRepository) UpsertSettings(ctx context.Context, s *FinanceSettings) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO finance_settings (tenant_id, ar_account_id, ap_account_id, cash_account_id,
		                               sales_revenue_id, purchase_expense_id, tax_account_id, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			ar_account_id       = EXCLUDED.ar_account_id,
			ap_account_id       = EXCLUDED.ap_account_id,
			cash_account_id     = EXCLUDED.cash_account_id,
			sales_revenue_id    = EXCLUDED.sales_revenue_id,
			purchase_expense_id = EXCLUDED.purchase_expense_id,
			tax_account_id      = EXCLUDED.tax_account_id,
			updated_at          = NOW()`,
		s.TenantID, s.ARAccountID, s.APAccountID, s.CashAccountID,
		s.SalesRevenueID, s.PurchaseExpenseID, s.TaxAccountID).Error
}

// ── Aging ───────────────────────────────────────────────────────────────────

func (r *dbRepository) AgingRows(ctx context.Context, tenantID uint, kind, asOf string) ([]AgingRow, error) {
	// kind = 'AR' → open sales invoices for customers.
	// kind = 'AP' → open purchase invoices for suppliers.
	// asOf defaults to CURRENT_DATE in the query if empty.
	table, partyCol, invoiceKindLit := "sales_invoices", "customer_id", "SI"
	if kind == "AP" {
		table, partyCol, invoiceKindLit = "purchase_invoices", "supplier_id", "PI"
	}
	// Always pass asOf as a bound text param cast to date server-side so
	// the placeholder count is deterministic. Default to CURRENT_DATE
	// when caller left it blank.
	if asOf == "" {
		asOf = "CURRENT_DATE"
	}
	asOfExpr := "?"
	args := []interface{}{asOf, tenantID}
	if asOf == "CURRENT_DATE" {
		asOfExpr = "CURRENT_DATE"
		args = []interface{}{tenantID}
	}
	// Cast due_date + invoice_date to text so a NULL comes back as an
	// empty string — Go's sql.Scan of a Date column rejects NULL into a
	// non-pointer string. asOfExpr is bound as a text param and cast to
	// date inside the SQL, so callers don't need to worry about format.
	query := fmt.Sprintf(`
		SELECT id, code, '%s' AS invoice_kind, %s AS party_id,
		       TO_CHAR(invoice_date, 'YYYY-MM-DD') AS invoice_date,
		       COALESCE(TO_CHAR(due_date, 'YYYY-MM-DD'), '') AS due_date,
		       total_amount, paid_amount,
		       (total_amount - paid_amount) AS outstanding_amount,
		       CASE WHEN due_date IS NULL THEN 0
		            ELSE GREATEST(0, (%s::date - due_date)) END AS days_overdue,
		       CASE
		           WHEN due_date IS NULL OR (%s::date - due_date) <= 30 THEN '0-30'
		           WHEN (%s::date - due_date) <= 60 THEN '31-60'
		           WHEN (%s::date - due_date) <= 90 THEN '61-90'
		           ELSE '90+'
		       END AS bucket
		FROM %s
		WHERE tenant_id = ?
		  AND status IN ('POSTED','PARTIAL')
		ORDER BY due_date NULLS LAST, id`,
		invoiceKindLit, partyCol, asOfExpr, asOfExpr, asOfExpr, asOfExpr, table)

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AgingRow
	for rows.Next() {
		var a AgingRow
		if err := rows.Scan(&a.InvoiceID, &a.InvoiceCode, &a.InvoiceKind, &a.PartyID,
			&a.InvoiceDate, &a.DueDate,
			&a.TotalAmount, &a.PaidAmount, &a.OutstandingAmount,
			&a.DaysOverdue, &a.Bucket); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ── Party outstanding ───────────────────────────────────────────────────────

func (r *dbRepository) PartyOutstanding(ctx context.Context, tenantID, partyID uint, kind string) (float64, error) {
	table, partyCol := "sales_invoices", "customer_id"
	if kind == "AP" {
		table, partyCol = "purchase_invoices", "supplier_id"
	}
	var sum float64
	err := r.db.WithContext(ctx).
		Raw(fmt.Sprintf(`SELECT COALESCE(SUM(total_amount - paid_amount), 0)
		                 FROM %s
		                 WHERE tenant_id = ? AND %s = ? AND status IN ('POSTED','PARTIAL')`, table, partyCol),
			tenantID, partyID).Row().Scan(&sum)
	return sum, err
}

// ── Party credit profile ────────────────────────────────────────────────────

func (r *dbRepository) PartyCreditProfile(ctx context.Context, tenantID, partyID uint) (string, float64, error) {
	var creditType string
	var creditLimit float64
	err := r.db.WithContext(ctx).
		Raw(`SELECT credit_type, credit_limit FROM parties WHERE tenant_id = ? AND id = ?`,
			tenantID, partyID).Row().Scan(&creditType, &creditLimit)
	return creditType, creditLimit, err
}
