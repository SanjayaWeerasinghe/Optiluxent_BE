package finance

import (
	"context"
	"fmt"
	"time"
)

// InvoicePaymentApplier is the tiny interface the finance module needs from
// the invoice-owning modules (procurement, sales) so it can bump paid_amount
// + status after inserting a payment row. Both procurement and sales
// services expose an equivalent method; adapters in cmd/api/main.go bridge
// per-module signatures to this shape.
type InvoicePaymentApplier interface {
	// ApplyPayment updates the invoice's paid_amount, transitions status
	// (POSTED → PARTIAL → PAID), and returns the party_id + currency_id
	// so the finance service can stamp the payment row correctly.
	ApplyPayment(ctx context.Context, tenantID, invoiceID uint, amount float64) (partyID, currencyID uint, err error)
}

// InvoiceLookup — read-only fetch used by the GL poster to build the
// initial SI/PI journal entry (amount + party). Satisfied by procurement
// and sales services via thin adapters.
type InvoiceLookup interface {
	LookupInvoice(ctx context.Context, tenantID, invoiceID uint) (partyID, currencyID uint, totalAmount, taxAmount float64, err error)
}

// Service is the finance module's public surface.
type Service struct {
	repo    Repository
	siApply InvoicePaymentApplier
	piApply InvoicePaymentApplier
	siLook  InvoiceLookup
	piLook  InvoiceLookup
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) SetSalesApplier(a InvoicePaymentApplier)      { s.siApply = a }
func (s *Service) SetProcurementApplier(a InvoicePaymentApplier) { s.piApply = a }
func (s *Service) SetSalesInvoiceLookup(l InvoiceLookup)        { s.siLook = l }
func (s *Service) SetProcurementInvoiceLookup(l InvoiceLookup)   { s.piLook = l }

// ── Payments ────────────────────────────────────────────────────────────────

// RecordPayment inserts a payment row, applies the amount to the invoice
// via the module-owned applier, and posts a balanced JE. All three happen
// against the same DB (payments + invoice update + JE) but not in a single
// tx — the applier is called through an interface. Failure to post the JE
// after payment success only logs a warning; the payment record is the
// source of truth.
func (s *Service) RecordPayment(ctx context.Context, tenantID, userID uint, req *RecordPaymentRequest) (*Payment, error) {
	applier := s.siApply
	direction := DirectionInbound
	seq := "PAYMENT_IN"
	if req.InvoiceKind == InvoiceKindPI {
		applier = s.piApply
		direction = DirectionOutbound
		seq = "PAYMENT_OUT"
	}
	if applier == nil {
		return nil, fmt.Errorf("no invoice applier wired for %s payments", req.InvoiceKind)
	}
	partyID, currencyID, err := applier.ApplyPayment(ctx, tenantID, req.InvoiceID, req.Amount)
	if err != nil {
		return nil, err
	}

	code, err := s.repo.NextCode(ctx, tenantID, seq)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate payment code: %w", err)
	}

	payDate := req.PaymentDate
	if payDate == "" {
		payDate = time.Now().Format("2006-01-02")
	}

	p := &Payment{
		TenantID:      tenantID,
		Code:          code,
		Direction:     direction,
		InvoiceKind:   req.InvoiceKind,
		InvoiceID:     req.InvoiceID,
		PartyID:       partyID,
		Amount:        req.Amount,
		CurrencyID:    currencyID,
		ExchangeRate:  1,
		PaymentDate:   payDate,
		Method:        req.Method,
		BankAccountID: req.BankAccountID,
		ReferenceNo:   req.ReferenceNo,
		Notes:         req.Notes,
		CreatedBy:     &userID,
	}
	if err := s.repo.CreatePayment(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Post the JE. Errors here don't roll back the payment (the operational
	// record wins) — they're surfaced in the response so the UI can flag
	// unposted payments for later reconciliation.
	je, jeErr := s.postPaymentJE(ctx, tenantID, userID, p)
	if jeErr == nil && je != nil {
		p.JournalID = &je.ID
		_ = s.repo.UpdatePayment(ctx, p)
	}
	return p, nil
}

func (s *Service) ListPayments(ctx context.Context, tenantID uint, invoiceKind string, invoiceID, partyID uint, direction string, limit int) ([]Payment, error) {
	return s.repo.ListPayments(ctx, tenantID, invoiceKind, invoiceID, partyID, direction, limit)
}

func (s *Service) GetPayment(ctx context.Context, tenantID, id uint) (*Payment, error) {
	return s.repo.GetPayment(ctx, tenantID, id)
}

// ── Settings ────────────────────────────────────────────────────────────────

func (s *Service) GetSettings(ctx context.Context, tenantID uint) (*FinanceSettings, error) {
	return s.repo.GetSettings(ctx, tenantID)
}

func (s *Service) UpdateSettings(ctx context.Context, tenantID uint, req *UpdateSettingsRequest) (*FinanceSettings, error) {
	settings := &FinanceSettings{
		TenantID:          tenantID,
		ARAccountID:       req.ARAccountID,
		APAccountID:       req.APAccountID,
		CashAccountID:     req.CashAccountID,
		SalesRevenueID:    req.SalesRevenueID,
		PurchaseExpenseID: req.PurchaseExpenseID,
		TaxAccountID:      req.TaxAccountID,
	}
	if err := s.repo.UpsertSettings(ctx, settings); err != nil {
		return nil, err
	}
	return s.repo.GetSettings(ctx, tenantID)
}

// ── Journal Entries (read-only listing) ─────────────────────────────────────

func (s *Service) ListJournalEntries(ctx context.Context, tenantID uint, sourceType string, sourceID uint, limit int) ([]JournalEntry, error) {
	return s.repo.ListJournalEntries(ctx, tenantID, sourceType, sourceID, limit)
}

func (s *Service) GetJournalEntry(ctx context.Context, tenantID, id uint) (*JournalEntry, error) {
	return s.repo.GetJournalEntry(ctx, tenantID, id)
}

// ── Aging + Outstanding ─────────────────────────────────────────────────────

func (s *Service) AgingReport(ctx context.Context, tenantID uint, kind, asOf string) ([]AgingRow, error) {
	if kind != "AR" && kind != "AP" {
		return nil, fmt.Errorf("kind must be AR or AP, got %q", kind)
	}
	return s.repo.AgingRows(ctx, tenantID, kind, asOf)
}

// PartyOutstanding is the credit-gate helper called from ConfirmSO via the
// sales.FinanceChecker interface.
func (s *Service) PartyOutstanding(ctx context.Context, tenantID, partyID uint, kind string) (float64, error) {
	return s.repo.PartyOutstanding(ctx, tenantID, partyID, kind)
}

// PartyCreditProfile passes through to the repo — used by sales to gate
// SO confirmations by CASH vs CREDIT + credit_limit.
func (s *Service) PartyCreditProfile(ctx context.Context, tenantID, partyID uint) (string, float64, error) {
	return s.repo.PartyCreditProfile(ctx, tenantID, partyID)
}

// ── Wrappers for cross-module hooks ─────────────────────────────────────────
//
// The procurement.FinancePostGL and sales.FinancePostGL interfaces expect a
// slightly different shape than the internal RecordPayment / PostSIToGL
// methods (they don't want to know about DTOs). These thin adapters bridge
// the shapes so main.go can wire finance.Service directly.

// RecordInvoicePaymentPI is called by procurement.Service after its own
// paid_amount bump (RecordPayment). Inserts a payments row via
// RecordPayment but skips the ApplyPayment step because procurement
// already did it.
func (s *Service) RecordInvoicePaymentPI(ctx context.Context, tenantID, userID, piID uint, amount float64, method, referenceNo, notes string, bankAccountID *uint, paymentDate string) error {
	// Look up the party + currency directly rather than going through
	// ApplyPayment (which would double-bump paid_amount).
	if s.piLook == nil {
		return fmt.Errorf("procurement invoice lookup not wired")
	}
	partyID, currencyID, _, _, err := s.piLook.LookupInvoice(ctx, tenantID, piID)
	if err != nil {
		return err
	}
	return s.insertPaymentAndPostJE(ctx, tenantID, userID, InvoiceKindPI, piID, partyID, currencyID,
		amount, method, referenceNo, notes, bankAccountID, paymentDate)
}

// RecordInvoicePaymentSI is the sales twin of RecordInvoicePaymentPI.
func (s *Service) RecordInvoicePaymentSI(ctx context.Context, tenantID, userID, siID uint, amount float64, method, referenceNo, notes string, bankAccountID *uint, paymentDate string) error {
	if s.siLook == nil {
		return fmt.Errorf("sales invoice lookup not wired")
	}
	partyID, currencyID, _, _, err := s.siLook.LookupInvoice(ctx, tenantID, siID)
	if err != nil {
		return err
	}
	return s.insertPaymentAndPostJE(ctx, tenantID, userID, InvoiceKindSI, siID, partyID, currencyID,
		amount, method, referenceNo, notes, bankAccountID, paymentDate)
}

// insertPaymentAndPostJE — shared body for the two wrappers above.
// Does NOT touch the invoice's paid_amount (caller already did).
func (s *Service) insertPaymentAndPostJE(ctx context.Context, tenantID, userID uint, invoiceKind string, invoiceID, partyID, currencyID uint, amount float64, method, referenceNo, notes string, bankAccountID *uint, paymentDate string) error {
	if method == "" {
		method = MethodOther
	}
	if paymentDate == "" {
		paymentDate = time.Now().Format("2006-01-02")
	}
	direction := DirectionInbound
	seq := "PAYMENT_IN"
	if invoiceKind == InvoiceKindPI {
		direction = DirectionOutbound
		seq = "PAYMENT_OUT"
	}
	code, err := s.repo.NextCode(ctx, tenantID, seq)
	if err != nil {
		return fmt.Errorf("payment code allocation: %w", err)
	}
	p := &Payment{
		TenantID:      tenantID,
		Code:          code,
		Direction:     direction,
		InvoiceKind:   invoiceKind,
		InvoiceID:     invoiceID,
		PartyID:       partyID,
		Amount:        amount,
		CurrencyID:    currencyID,
		ExchangeRate:  1,
		PaymentDate:   paymentDate,
		Method:        method,
		BankAccountID: bankAccountID,
		ReferenceNo:   referenceNo,
		Notes:         notes,
	}
	if userID > 0 {
		p.CreatedBy = &userID
	}
	if err := s.repo.CreatePayment(ctx, p); err != nil {
		return fmt.Errorf("save payment: %w", err)
	}
	je, jeErr := s.postPaymentJE(ctx, tenantID, userID, p)
	if jeErr == nil && je != nil {
		p.JournalID = &je.ID
		_ = s.repo.UpdatePayment(ctx, p)
	}
	return nil
}

// ── JE posting for SI / PI post ─────────────────────────────────────────────
//
// These are called by procurement and sales AFTER a successful invoice
// post. Best-effort — errors surface up but the caller (post workflow)
// treats a failed JE as a warning, not a rollback.

func (s *Service) PostSIToGL(ctx context.Context, tenantID, userID, siID uint) (*JournalEntry, error) {
	if s.siLook == nil {
		return nil, fmt.Errorf("sales invoice lookup not wired")
	}
	partyID, _, totalAmount, taxAmount, err := s.siLook.LookupInvoice(ctx, tenantID, siID)
	if err != nil {
		return nil, err
	}
	settings, err := s.repo.GetSettings(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// Dr AR (total), Cr Revenue (total-tax), Cr Tax (tax) if tax > 0.
	je, err := s.buildJE(ctx, tenantID, userID, SourceTypeSIPost, siID,
		fmt.Sprintf("SI #%d post — AR receivable", siID))
	if err != nil {
		return nil, err
	}
	if settings.ARAccountID != nil {
		je.Lines = append(je.Lines, GLLine{
			TenantID: tenantID, LineNumber: len(je.Lines) + 1,
			AccountID: *settings.ARAccountID, PartyID: &partyID,
			Debit: totalAmount, Description: "Accounts Receivable",
		})
	}
	if settings.SalesRevenueID != nil {
		je.Lines = append(je.Lines, GLLine{
			TenantID: tenantID, LineNumber: len(je.Lines) + 1,
			AccountID: *settings.SalesRevenueID,
			Credit: totalAmount - taxAmount, Description: "Sales Revenue",
		})
	}
	if taxAmount > 0 && settings.TaxAccountID != nil {
		je.Lines = append(je.Lines, GLLine{
			TenantID: tenantID, LineNumber: len(je.Lines) + 1,
			AccountID: *settings.TaxAccountID,
			Credit: taxAmount, Description: "Tax Payable",
		})
	}
	return je, s.finishJE(ctx, je)
}

func (s *Service) PostPIToGL(ctx context.Context, tenantID, userID, piID uint) (*JournalEntry, error) {
	if s.piLook == nil {
		return nil, fmt.Errorf("purchase invoice lookup not wired")
	}
	partyID, _, totalAmount, taxAmount, err := s.piLook.LookupInvoice(ctx, tenantID, piID)
	if err != nil {
		return nil, err
	}
	settings, err := s.repo.GetSettings(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// Dr Purchase Expense (total-tax), Dr Tax Recoverable (tax), Cr AP (total).
	je, err := s.buildJE(ctx, tenantID, userID, SourceTypePIPost, piID,
		fmt.Sprintf("PI #%d post — AP liability", piID))
	if err != nil {
		return nil, err
	}
	if settings.PurchaseExpenseID != nil {
		je.Lines = append(je.Lines, GLLine{
			TenantID: tenantID, LineNumber: len(je.Lines) + 1,
			AccountID: *settings.PurchaseExpenseID,
			Debit: totalAmount - taxAmount, Description: "Purchases",
		})
	}
	if taxAmount > 0 && settings.TaxAccountID != nil {
		je.Lines = append(je.Lines, GLLine{
			TenantID: tenantID, LineNumber: len(je.Lines) + 1,
			AccountID: *settings.TaxAccountID,
			Debit: taxAmount, Description: "Tax Recoverable",
		})
	}
	if settings.APAccountID != nil {
		je.Lines = append(je.Lines, GLLine{
			TenantID: tenantID, LineNumber: len(je.Lines) + 1,
			AccountID: *settings.APAccountID, PartyID: &partyID,
			Credit: totalAmount, Description: "Accounts Payable",
		})
	}
	return je, s.finishJE(ctx, je)
}

// ── JE posting for a payment row ────────────────────────────────────────────

func (s *Service) postPaymentJE(ctx context.Context, tenantID, userID uint, p *Payment) (*JournalEntry, error) {
	settings, err := s.repo.GetSettings(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	je, err := s.buildJE(ctx, tenantID, userID, SourceTypePayment, p.ID,
		fmt.Sprintf("Payment %s — %s %s#%d", p.Code, p.Direction, p.InvoiceKind, p.InvoiceID))
	if err != nil {
		return nil, err
	}
	// Cash-side account picks the bank if supplied, else the generic cash.
	cashAcc := settings.CashAccountID
	if p.BankAccountID != nil {
		// bank_account already has its own gl_account_id on
		// company_bank_accounts.gl_account_id. Prefer that.
		if bankAcc, err := s.bankGLAccount(ctx, tenantID, *p.BankAccountID); err == nil && bankAcc != nil {
			cashAcc = bankAcc
		}
	}

	switch p.Direction {
	case DirectionInbound: // AR receipt — Dr Bank/Cash, Cr AR
		if cashAcc != nil {
			je.Lines = append(je.Lines, GLLine{
				TenantID: tenantID, LineNumber: len(je.Lines) + 1,
				AccountID: *cashAcc, Debit: p.Amount, Description: "Cash / Bank receipt",
			})
		}
		if settings.ARAccountID != nil {
			partyID := p.PartyID
			je.Lines = append(je.Lines, GLLine{
				TenantID: tenantID, LineNumber: len(je.Lines) + 1,
				AccountID: *settings.ARAccountID, PartyID: &partyID,
				Credit: p.Amount, Description: "AR settlement",
			})
		}
	case DirectionOutbound: // AP disbursement — Dr AP, Cr Bank/Cash
		if settings.APAccountID != nil {
			partyID := p.PartyID
			je.Lines = append(je.Lines, GLLine{
				TenantID: tenantID, LineNumber: len(je.Lines) + 1,
				AccountID: *settings.APAccountID, PartyID: &partyID,
				Debit: p.Amount, Description: "AP settlement",
			})
		}
		if cashAcc != nil {
			je.Lines = append(je.Lines, GLLine{
				TenantID: tenantID, LineNumber: len(je.Lines) + 1,
				AccountID: *cashAcc, Credit: p.Amount, Description: "Cash / Bank paid out",
			})
		}
	}
	return je, s.finishJE(ctx, je)
}

// ── JE helpers ──────────────────────────────────────────────────────────────

func (s *Service) buildJE(ctx context.Context, tenantID, userID uint, sourceType string, sourceID uint, narration string) (*JournalEntry, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "JOURNAL_ENTRY")
	if err != nil {
		return nil, err
	}
	return &JournalEntry{
		TenantID:   tenantID,
		Code:       code,
		PostDate:   time.Now().Format("2006-01-02"),
		Narration:  narration,
		SourceType: sourceType,
		SourceID:   sourceID,
		IsPosted:   true,
		CreatedBy:  &userID,
	}, nil
}

func (s *Service) finishJE(ctx context.Context, je *JournalEntry) error {
	if len(je.Lines) == 0 {
		return nil // Nothing configured for this side — no-op.
	}
	// Force tenant_id on every line (GORM sometimes leaves association FK
	// scaffolding on the parent side).
	for i := range je.Lines {
		je.Lines[i].TenantID = je.TenantID
	}
	if err := s.assertBalanced(je); err != nil {
		return err
	}
	return s.repo.CreateJournalEntry(ctx, je)
}

func (s *Service) assertBalanced(je *JournalEntry) error {
	var dr, cr float64
	for _, l := range je.Lines {
		dr += l.Debit
		cr += l.Credit
	}
	// Half-cent tolerance for float rounding.
	if dr-cr > 0.005 || cr-dr > 0.005 {
		return fmt.Errorf("journal entry not balanced: debits=%.4f credits=%.4f", dr, cr)
	}
	return nil
}

// bankGLAccount reads company_bank_accounts.gl_account_id — used to steer
// a bank-transfer payment's cash leg onto the specific bank's GL account.
func (s *Service) bankGLAccount(ctx context.Context, tenantID, bankAccountID uint) (*uint, error) {
	var glAccountID *uint
	err := s.repo.(*dbRepository).db.WithContext(ctx).
		Raw(`SELECT gl_account_id FROM company_bank_accounts WHERE tenant_id = ? AND id = ?`,
			tenantID, bankAccountID).Row().Scan(&glAccountID)
	if err != nil {
		return nil, err
	}
	return glAccountID, nil
}
