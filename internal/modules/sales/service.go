package sales

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"erp-system/internal/infrastructure/events"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
	bus  events.EventBus
}

func NewService(repo Repository) *Service           { return &Service{repo: repo} }
func (s *Service) SetEventBus(bus events.EventBus)  { s.bus = bus }

func (s *Service) publish(ctx context.Context, eventType string, tenantID, userID uint, payload any) {
	if s.bus == nil {
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	ev := events.Event{
		ID:         uuid.NewString(),
		Type:       eventType,
		TenantID:   tenantID,
		UserID:     userID,
		OccurredAt: time.Now(),
		Payload:    json.RawMessage(raw),
	}
	_ = s.bus.Publish(ctx, events.StreamSales, ev)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func today() string { return time.Now().Format("2006-01-02") }

func isDuplicate(err error) bool {
	return strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique")
}

func soLineTotal(line *SOLine) float64 {
	return line.Quantity * line.UnitPrice * (1 - line.DiscountPct/100)
}

func siLineTotal(line *SILine) float64 {
	return line.Quantity * line.UnitPrice * (1 - line.DiscountPct/100)
}

func calcSOTotals(so *SalesOrder) {
	var sub, tax float64
	for _, l := range so.Lines {
		sub += l.LineTotal
		tax += l.TaxAmount
	}
	so.Subtotal = sub
	so.TaxAmount = tax
	so.TotalAmount = sub + tax - so.DiscountAmount
}

func calcSITotals(si *SalesInvoice) {
	var sub, tax float64
	for _, l := range si.Lines {
		sub += l.LineTotal
		tax += l.TaxAmount
	}
	si.Subtotal = sub
	si.TaxAmount = tax
	si.TotalAmount = sub + tax - si.DiscountAmount
}

func sqLineTotal(line *SQLine) float64 {
	return line.Quantity * line.UnitPrice * (1 - line.DiscountPct/100)
}

func calcSQTotals(sq *SalesQuotation) {
	var sub, tax float64
	for _, l := range sq.Lines {
		sub += l.LineTotal
		tax += l.TaxAmount
	}
	sq.Subtotal = sub
	sq.TaxAmount = tax
	sq.TotalAmount = sub + tax - sq.DiscountAmount
}

// ── Sales Quotations ──────────────────────────────────────────────────────────

func (s *Service) ListSQs(ctx context.Context, tenantID uint, customerID *uint, status string) ([]SalesQuotation, error) {
	return s.repo.ListSQs(ctx, tenantID, customerID, status)
}

func (s *Service) GetSQ(ctx context.Context, tenantID, id uint) (*SalesQuotation, error) {
	return s.repo.GetSQ(ctx, tenantID, id)
}

func (s *Service) CreateSQ(ctx context.Context, tenantID, userID uint, req CreateSQRequest) (*SalesQuotation, error) {
	code := req.Code
	if code == "" {
		var err error
		code, err = s.repo.NextSQCode(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate SQ code: %w", err)
		}
	}
	d := req.QuotationDate
	if d == "" {
		d = today()
	}
	er := req.ExchangeRate
	if er == 0 {
		er = 1
	}
	sq := &SalesQuotation{
		TenantID:           tenantID,
		Code:               code,
		CustomerID:         req.CustomerID,
		QuotationDate:      d,
		ValidUntil:         req.ValidUntil,
		CurrencyID:         req.CurrencyID,
		ExchangeRate:       er,
		PaymentTermID:      req.PaymentTermID,
		WarehouseID:        req.WarehouseID,
		Status:             SQStatusDraft,
		CustomerReference:  req.CustomerReference,
		TermsAndConditions: req.TermsAndConditions,
		Notes:              req.Notes,
		CreatedBy:          &userID,
	}
	if err := s.repo.CreateSQ(ctx, sq); err != nil {
		if isDuplicate(err) {
			return nil, fmt.Errorf("quotation code '%s' already exists", code)
		}
		return nil, err
	}
	return sq, nil
}

func (s *Service) UpdateSQ(ctx context.Context, tenantID, id uint, req UpdateSQRequest) (*SalesQuotation, error) {
	sq, err := s.repo.GetSQ(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("quotation not found")
	}
	if sq.Status != SQStatusDraft {
		return nil, fmt.Errorf("only DRAFT quotations can be edited")
	}
	sq.ValidUntil = req.ValidUntil
	sq.PaymentTermID = req.PaymentTermID
	if req.ExchangeRate != nil {
		sq.ExchangeRate = *req.ExchangeRate
	}
	if req.DiscountAmount != nil {
		sq.DiscountAmount = *req.DiscountAmount
	}
	if req.CustomerReference != "" {
		sq.CustomerReference = req.CustomerReference
	}
	if req.TermsAndConditions != "" {
		sq.TermsAndConditions = req.TermsAndConditions
	}
	if req.Notes != "" {
		sq.Notes = req.Notes
	}
	calcSQTotals(sq)
	return sq, s.repo.UpdateSQ(ctx, sq)
}

func (s *Service) DeleteSQ(ctx context.Context, tenantID, id uint) error {
	sq, err := s.repo.GetSQ(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("quotation not found")
	}
	if sq.Status != SQStatusDraft {
		return fmt.Errorf("only DRAFT quotations can be deleted")
	}
	return s.repo.DeleteSQ(ctx, tenantID, id)
}

func (s *Service) SubmitSQ(ctx context.Context, tenantID, id uint) (*SalesQuotation, error) {
	sq, err := s.repo.GetSQ(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("quotation not found")
	}
	if sq.Status != SQStatusDraft {
		return nil, fmt.Errorf("only DRAFT quotations can be sent")
	}
	if len(sq.Lines) == 0 {
		return nil, fmt.Errorf("quotation must have at least one item")
	}
	sq.Status = SQStatusSent
	return sq, s.repo.UpdateSQ(ctx, sq)
}

// AcceptSQ transitions a SENT quotation to ACCEPTED and creates a Sales Order
// with the same lines. The new SO carries sq_id as a back-reference so users
// can drill through.
func (s *Service) AcceptSQ(ctx context.Context, tenantID, userID, id uint) (*SalesQuotation, *SalesOrder, error) {
	sq, err := s.repo.GetSQ(ctx, tenantID, id)
	if err != nil {
		return nil, nil, fmt.Errorf("quotation not found")
	}
	if sq.Status != SQStatusSent {
		return nil, nil, fmt.Errorf("only SENT quotations can be accepted")
	}
	// Generate an SO code and build the new SO
	soCode, err := s.repo.NextSOCode(ctx, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate SO code: %w", err)
	}
	sqID := sq.ID
	so := &SalesOrder{
		TenantID:      tenantID,
		Code:          soCode,
		SQID:          &sqID,
		CustomerID:    sq.CustomerID,
		OrderDate:     today(),
		CurrencyID:    sq.CurrencyID,
		ExchangeRate:  sq.ExchangeRate,
		PaymentTermID: sq.PaymentTermID,
		WarehouseID:   sq.WarehouseID,
		Status:        SOStatusDraft,
		Notes:         sq.Notes,
		CreatedBy:     &userID,
	}
	if err := s.repo.CreateSO(ctx, so); err != nil {
		return nil, nil, err
	}
	// Copy each SQ line to the new SO
	for i, sl := range sq.Lines {
		soLine := &SOLine{
			SOID:        so.ID,
			TenantID:    tenantID,
			LineNumber:  i + 1,
			ProductID:   sl.ProductID,
			VariantID:   sl.VariantID,
			Description: sl.Description,
			Quantity:    sl.Quantity,
			UOMID:       sl.UOMID,
			UnitPrice:   sl.UnitPrice,
			DiscountPct: sl.DiscountPct,
			TaxCodeID:   sl.TaxCodeID,
			TaxAmount:   sl.TaxAmount,
			LineTotal:   sl.LineTotal,
			Notes:       sl.Notes,
		}
		if err := s.repo.AddSOLine(ctx, soLine); err != nil {
			return nil, nil, err
		}
		so.Lines = append(so.Lines, *soLine)
	}
	calcSOTotals(so)
	if err := s.repo.UpdateSO(ctx, so); err != nil {
		return nil, nil, err
	}
	// Mirror the procurement/sales pattern: creating an SO already spins up a
	// draft SI header (see CreateSO). Since we called repo.CreateSO directly
	// here (bypassing the service), do the same auto-invoice manually so the
	// flow stays consistent.
	if invCode, err := s.repo.NextSICode(ctx, tenantID); err == nil {
		_ = s.repo.CreateSI(ctx, &SalesInvoice{
			TenantID:      tenantID,
			Code:          invCode,
			CustomerID:    so.CustomerID,
			SOID:          so.ID,
			InvoiceDate:   today(),
			CurrencyID:    so.CurrencyID,
			ExchangeRate:  so.ExchangeRate,
			PaymentTermID: so.PaymentTermID,
			Status:        SIStatusDraft,
			CreatedBy:     &userID,
		})
	}
	// Mark the SQ as ACCEPTED and back-link
	now := time.Now()
	sq.Status = SQStatusAccepted
	sq.AcceptedBy = &userID
	sq.AcceptedAt = &now
	sq.ConvertedSOID = &so.ID
	if err := s.repo.UpdateSQ(ctx, sq); err != nil {
		return nil, nil, err
	}
	return sq, so, nil
}

func (s *Service) RejectSQ(ctx context.Context, tenantID, userID, id uint) (*SalesQuotation, error) {
	sq, err := s.repo.GetSQ(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("quotation not found")
	}
	if sq.Status != SQStatusSent {
		return nil, fmt.Errorf("only SENT quotations can be rejected")
	}
	sq.Status = SQStatusRejected
	sq.RejectedBy = &userID
	return sq, s.repo.UpdateSQ(ctx, sq)
}

func (s *Service) CancelSQ(ctx context.Context, tenantID, id uint) (*SalesQuotation, error) {
	sq, err := s.repo.GetSQ(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("quotation not found")
	}
	if sq.Status == SQStatusAccepted || sq.Status == SQStatusCancelled {
		return nil, fmt.Errorf("cannot cancel a %s quotation", sq.Status)
	}
	sq.Status = SQStatusCancelled
	return sq, s.repo.UpdateSQ(ctx, sq)
}

// ── SQ Lines ──────────────────────────────────────────────────────────────────

func (s *Service) ListSQLines(ctx context.Context, tenantID, sqID uint) ([]SQLine, error) {
	if _, err := s.repo.GetSQ(ctx, tenantID, sqID); err != nil {
		return nil, fmt.Errorf("quotation not found")
	}
	return s.repo.ListSQLines(ctx, tenantID, sqID)
}

func (s *Service) GetSQLine(ctx context.Context, tenantID, sqID, lineID uint) (*SQLine, error) {
	return s.repo.GetSQLine(ctx, tenantID, sqID, lineID)
}

func (s *Service) AddSQLine(ctx context.Context, tenantID, sqID uint, req AddSQLineRequest) (*SQLine, error) {
	sq, err := s.repo.GetSQ(ctx, tenantID, sqID)
	if err != nil {
		return nil, fmt.Errorf("quotation not found")
	}
	if sq.Status != SQStatusDraft {
		return nil, fmt.Errorf("cannot add items to a %s quotation", sq.Status)
	}
	line := &SQLine{
		SQID:        sqID,
		TenantID:    tenantID,
		LineNumber:  len(sq.Lines) + 1,
		ProductID:   req.ProductID,
		VariantID:   req.VariantID,
		Description: req.Description,
		Quantity:    req.Quantity,
		UOMID:       req.UOMID,
		UnitPrice:   req.UnitPrice,
		DiscountPct: req.DiscountPct,
		TaxCodeID:   req.TaxCodeID,
		Notes:       req.Notes,
	}
	line.LineTotal = sqLineTotal(line)
	if err := s.repo.AddSQLine(ctx, line); err != nil {
		return nil, err
	}
	sq.Lines = append(sq.Lines, *line)
	calcSQTotals(sq)
	_ = s.repo.UpdateSQ(ctx, sq)
	return line, nil
}

func (s *Service) UpdateSQLine(ctx context.Context, tenantID, sqID, lineID uint, req UpdateSQLineRequest) (*SQLine, error) {
	sq, err := s.repo.GetSQ(ctx, tenantID, sqID)
	if err != nil {
		return nil, fmt.Errorf("quotation not found")
	}
	if sq.Status != SQStatusDraft {
		return nil, fmt.Errorf("cannot edit items on a %s quotation", sq.Status)
	}
	line, err := s.repo.GetSQLine(ctx, tenantID, sqID, lineID)
	if err != nil {
		return nil, fmt.Errorf("item not found")
	}
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.Description = req.Description
	line.Quantity = req.Quantity
	line.UOMID = req.UOMID
	line.UnitPrice = req.UnitPrice
	line.DiscountPct = req.DiscountPct
	line.TaxCodeID = req.TaxCodeID
	line.Notes = req.Notes
	line.LineTotal = sqLineTotal(line)
	if err := s.repo.UpdateSQLine(ctx, line); err != nil {
		return nil, err
	}
	sq, _ = s.repo.GetSQ(ctx, tenantID, sqID)
	calcSQTotals(sq)
	_ = s.repo.UpdateSQ(ctx, sq)
	return line, nil
}

func (s *Service) DeleteSQLine(ctx context.Context, tenantID, sqID, lineID uint) error {
	sq, err := s.repo.GetSQ(ctx, tenantID, sqID)
	if err != nil {
		return fmt.Errorf("quotation not found")
	}
	if sq.Status != SQStatusDraft {
		return fmt.Errorf("cannot delete items from a %s quotation", sq.Status)
	}
	if err := s.repo.DeleteSQLine(ctx, tenantID, sqID, lineID); err != nil {
		return err
	}
	sq, _ = s.repo.GetSQ(ctx, tenantID, sqID)
	calcSQTotals(sq)
	return s.repo.UpdateSQ(ctx, sq)
}

// ── Sales Orders ──────────────────────────────────────────────────────────────

func (s *Service) ListSOs(ctx context.Context, tenantID uint, customerID *uint, status string) ([]SalesOrder, error) {
	return s.repo.ListSOs(ctx, tenantID, customerID, status)
}

func (s *Service) GetSO(ctx context.Context, tenantID, id uint) (*SalesOrder, error) {
	return s.repo.GetSO(ctx, tenantID, id)
}

func (s *Service) CreateSO(ctx context.Context, tenantID, userID uint, req CreateSORequest) (*SalesOrder, error) {
	code := req.Code
	if code == "" {
		var err error
		code, err = s.repo.NextSOCode(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate SO code: %w", err)
		}
	}
	d := req.OrderDate
	if d == "" {
		d = today()
	}
	er := req.ExchangeRate
	if er == 0 {
		er = 1
	}
	so := &SalesOrder{
		TenantID:             tenantID,
		Code:                 code,
		CustomerID:           req.CustomerID,
		OrderDate:            d,
		ExpectedDeliveryDate: req.ExpectedDeliveryDate,
		CurrencyID:           req.CurrencyID,
		ExchangeRate:         er,
		PaymentTermID:        req.PaymentTermID,
		WarehouseID:          req.WarehouseID,
		Status:               SOStatusDraft,
		Notes:                req.Notes,
		CreatedBy:            &userID,
	}
	if err := s.repo.CreateSO(ctx, so); err != nil {
		if isDuplicate(err) {
			return nil, fmt.Errorf("sales order code '%s' already exists", code)
		}
		return nil, err
	}
	// Auto-create a draft Sales Invoice so the expected bill exists from day one.
	// Lines are populated later as DOs are confirmed (actual delivered quantities).
	if invCode, err := s.repo.NextSICode(ctx, tenantID); err == nil {
		si := &SalesInvoice{
			TenantID:      tenantID,
			Code:          invCode,
			CustomerID:    so.CustomerID,
			SOID:          so.ID,
			InvoiceDate:   d,
			CurrencyID:    so.CurrencyID,
			ExchangeRate:  er,
			PaymentTermID: so.PaymentTermID,
			Status:        SIStatusDraft,
			CreatedBy:     &userID,
		}
		_ = s.repo.CreateSI(ctx, si)
	}
	return so, nil
}

func (s *Service) UpdateSO(ctx context.Context, tenantID, id uint, req UpdateSORequest) (*SalesOrder, error) {
	so, err := s.repo.GetSO(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("sales order not found")
	}
	if so.Status != SOStatusDraft {
		return nil, fmt.Errorf("only DRAFT sales orders can be edited")
	}
	so.ExpectedDeliveryDate = req.ExpectedDeliveryDate
	so.PaymentTermID = req.PaymentTermID
	if req.ExchangeRate != nil {
		so.ExchangeRate = *req.ExchangeRate
	}
	if req.DiscountAmount != nil {
		so.DiscountAmount = *req.DiscountAmount
	}
	if req.Notes != "" {
		so.Notes = req.Notes
	}
	calcSOTotals(so)
	return so, s.repo.UpdateSO(ctx, so)
}

func (s *Service) DeleteSO(ctx context.Context, tenantID, id uint) error {
	so, err := s.repo.GetSO(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("sales order not found")
	}
	if so.Status != SOStatusDraft {
		return fmt.Errorf("only DRAFT sales orders can be deleted")
	}
	return s.repo.DeleteSO(ctx, tenantID, id)
}

func (s *Service) ConfirmSO(ctx context.Context, tenantID, userID, soID uint) error {
	so, err := s.repo.GetSO(ctx, tenantID, soID)
	if err != nil {
		return fmt.Errorf("sales order not found")
	}
	if so.Status != SOStatusDraft {
		return fmt.Errorf("only DRAFT sales orders can be confirmed")
	}
	if len(so.Lines) == 0 {
		return fmt.Errorf("sales order must have at least one item")
	}
	now := time.Now()
	so.Status = SOStatusConfirmed
	so.ConfirmedBy = &userID
	so.ConfirmedAt = &now
	return s.repo.UpdateSO(ctx, so)
}

func (s *Service) CancelSO(ctx context.Context, tenantID, soID uint) error {
	so, err := s.repo.GetSO(ctx, tenantID, soID)
	if err != nil {
		return fmt.Errorf("sales order not found")
	}
	if so.Status == SOStatusDelivered || so.Status == SOStatusCancelled {
		return fmt.Errorf("cannot cancel a %s sales order", so.Status)
	}
	so.Status = SOStatusCancelled
	return s.repo.UpdateSO(ctx, so)
}

// ── SO Lines ──────────────────────────────────────────────────────────────────

func (s *Service) ListSOLines(ctx context.Context, tenantID, soID uint) ([]SOLine, error) {
	if _, err := s.repo.GetSO(ctx, tenantID, soID); err != nil {
		return nil, fmt.Errorf("sales order not found")
	}
	return s.repo.ListSOLines(ctx, tenantID, soID)
}

func (s *Service) GetSOLine(ctx context.Context, tenantID, soID, lineID uint) (*SOLine, error) {
	return s.repo.GetSOLine(ctx, tenantID, soID, lineID)
}

func (s *Service) AddSOLine(ctx context.Context, tenantID, soID uint, req AddSOLineRequest) (*SOLine, error) {
	so, err := s.repo.GetSO(ctx, tenantID, soID)
	if err != nil {
		return nil, fmt.Errorf("sales order not found")
	}
	if so.Status != SOStatusDraft {
		return nil, fmt.Errorf("cannot add items to a %s sales order", so.Status)
	}
	line := &SOLine{
		SOID:        soID,
		TenantID:    tenantID,
		LineNumber:  len(so.Lines) + 1,
		ProductID:   req.ProductID,
		VariantID:   req.VariantID,
		Description: req.Description,
		Quantity:    req.Quantity,
		UOMID:       req.UOMID,
		UnitPrice:   req.UnitPrice,
		DiscountPct: req.DiscountPct,
		TaxCodeID:   req.TaxCodeID,
		Notes:       req.Notes,
	}
	line.LineTotal = soLineTotal(line)
	if err := s.repo.AddSOLine(ctx, line); err != nil {
		return nil, err
	}
	so.Lines = append(so.Lines, *line)
	calcSOTotals(so)
	_ = s.repo.UpdateSO(ctx, so)
	return line, nil
}

func (s *Service) UpdateSOLine(ctx context.Context, tenantID, soID, lineID uint, req UpdateSOLineRequest) (*SOLine, error) {
	so, err := s.repo.GetSO(ctx, tenantID, soID)
	if err != nil {
		return nil, fmt.Errorf("sales order not found")
	}
	if so.Status != SOStatusDraft {
		return nil, fmt.Errorf("cannot edit items on a %s sales order", so.Status)
	}
	line, err := s.repo.GetSOLine(ctx, tenantID, soID, lineID)
	if err != nil {
		return nil, fmt.Errorf("item not found")
	}
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.Description = req.Description
	line.Quantity = req.Quantity
	line.UOMID = req.UOMID
	line.UnitPrice = req.UnitPrice
	line.DiscountPct = req.DiscountPct
	line.TaxCodeID = req.TaxCodeID
	line.Notes = req.Notes
	line.LineTotal = soLineTotal(line)
	if err := s.repo.UpdateSOLine(ctx, line); err != nil {
		return nil, err
	}
	// Recalculate SO totals
	so, _ = s.repo.GetSO(ctx, tenantID, soID)
	calcSOTotals(so)
	_ = s.repo.UpdateSO(ctx, so)
	return line, nil
}

func (s *Service) DeleteSOLine(ctx context.Context, tenantID, soID, lineID uint) error {
	so, err := s.repo.GetSO(ctx, tenantID, soID)
	if err != nil {
		return fmt.Errorf("sales order not found")
	}
	if so.Status != SOStatusDraft {
		return fmt.Errorf("cannot delete items from a %s sales order", so.Status)
	}
	if err := s.repo.DeleteSOLine(ctx, tenantID, soID, lineID); err != nil {
		return err
	}
	so, _ = s.repo.GetSO(ctx, tenantID, soID)
	calcSOTotals(so)
	return s.repo.UpdateSO(ctx, so)
}

// ── Delivery Orders ───────────────────────────────────────────────────────────

func (s *Service) ListDOs(ctx context.Context, tenantID uint, soID *uint, status string) ([]DeliveryOrder, error) {
	return s.repo.ListDOs(ctx, tenantID, soID, status)
}

func (s *Service) GetDO(ctx context.Context, tenantID, id uint) (*DeliveryOrder, error) {
	return s.repo.GetDO(ctx, tenantID, id)
}

func (s *Service) CreateDO(ctx context.Context, tenantID, userID uint, req CreateDORequest) (*DeliveryOrder, error) {
	code := req.Code
	if code == "" {
		var err error
		code, err = s.repo.NextDOCode(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate DO code: %w", err)
		}
	}
	d := req.DeliveryDate
	if d == "" {
		d = today()
	}
	do := &DeliveryOrder{
		TenantID:     tenantID,
		Code:         code,
		SOID:         req.SOID,
		CustomerID:   req.CustomerID,
		DeliveryDate: d,
		WarehouseID:  req.WarehouseID,
		Status:       DOStatusDraft,
		Notes:        req.Notes,
		CreatedBy:    &userID,
	}
	if err := s.repo.CreateDO(ctx, do); err != nil {
		if isDuplicate(err) {
			return nil, fmt.Errorf("delivery order code '%s' already exists", code)
		}
		return nil, err
	}
	return do, nil
}

func (s *Service) UpdateDO(ctx context.Context, tenantID, id uint, req UpdateDORequest) (*DeliveryOrder, error) {
	do, err := s.repo.GetDO(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("delivery order not found")
	}
	if do.Status != DOStatusDraft {
		return nil, fmt.Errorf("only DRAFT delivery orders can be edited")
	}
	if req.DeliveryDate != nil {
		do.DeliveryDate = *req.DeliveryDate
	}
	if req.WarehouseID != nil {
		do.WarehouseID = *req.WarehouseID
	}
	if req.Notes != "" {
		do.Notes = req.Notes
	}
	return do, s.repo.UpdateDO(ctx, do)
}

func (s *Service) ConfirmDO(ctx context.Context, tenantID, userID, doID uint) error {
	do, err := s.repo.GetDO(ctx, tenantID, doID)
	if err != nil {
		return fmt.Errorf("delivery order not found")
	}
	if do.Status != DOStatusDraft {
		return fmt.Errorf("only DRAFT delivery orders can be confirmed")
	}
	if len(do.Lines) == 0 {
		return fmt.Errorf("delivery order must have at least one item")
	}
	if err := s.repo.ConfirmDO(ctx, tenantID, doID, userID, time.Now()); err != nil {
		return err
	}
	// Auto-append DO lines to the SO's draft invoice
	if do.SOID != nil {
		s.appendDOLinesToDraftInvoice(ctx, tenantID, do)
	}
	return nil
}

// appendDOLinesToDraftInvoice adds lines from a confirmed DO to the SO's draft invoice.
// Prices are sourced from SO lines; falls back to 0 if no SO line is linked.
// Best-effort — failures do not roll back the DO confirmation.
func (s *Service) appendDOLinesToDraftInvoice(ctx context.Context, tenantID uint, do *DeliveryOrder) {
	if do.SOID == nil {
		return
	}
	si, err := s.repo.GetDraftInvoiceBySOID(ctx, tenantID, *do.SOID)
	if err != nil {
		return // no draft invoice — nothing to append to
	}
	// Pre-load all SO lines once so we can resolve prices for DO lines that
	// didn't explicitly set so_line_id (the FE doesn't expose that field).
	soLines, _ := s.repo.ListSOLines(ctx, tenantID, *do.SOID)
	soLineByProduct := make(map[uint]*SOLine, len(soLines))
	for i := range soLines {
		l := soLines[i]
		soLineByProduct[l.ProductID] = &l
	}

	nextLine := len(si.Lines) + 1
	for _, doLine := range do.Lines {
		var unitPrice float64
		var taxCodeID *uint
		var discountPct float64
		var resolvedSOLine *SOLine
		if doLine.SOLineID != nil {
			if soLine, err := s.repo.GetSOLine(ctx, tenantID, *do.SOID, *doLine.SOLineID); err == nil {
				resolvedSOLine = soLine
			}
		}
		if resolvedSOLine == nil {
			// Fallback: match SO line by product_id within the same SO.
			resolvedSOLine = soLineByProduct[doLine.ProductID]
		}
		if resolvedSOLine != nil {
			unitPrice = resolvedSOLine.UnitPrice
			taxCodeID = resolvedSOLine.TaxCodeID
			discountPct = resolvedSOLine.DiscountPct
		}
		doLineID := doLine.ID
		line := &SILine{
			InvoiceID:   si.ID,
			TenantID:    tenantID,
			DOLineID:    &doLineID,
			LineNumber:  nextLine,
			ProductID:   doLine.ProductID,
			VariantID:   doLine.VariantID,
			Quantity:    doLine.Quantity,
			UOMID:       doLine.UOMID,
			UnitPrice:   unitPrice,
			DiscountPct: discountPct,
			TaxCodeID:   taxCodeID,
		}
		line.LineTotal = siLineTotal(line)
		if err := s.repo.AddSILine(ctx, line); err == nil {
			si.Lines = append(si.Lines, *line)
			nextLine++
		}
	}
	calcSITotals(si)
	_ = s.repo.UpdateSI(ctx, si)
}

// ── DO Lines ──────────────────────────────────────────────────────────────────

func (s *Service) ListDOLines(ctx context.Context, tenantID, doID uint) ([]DOLine, error) {
	if _, err := s.repo.GetDO(ctx, tenantID, doID); err != nil {
		return nil, fmt.Errorf("delivery order not found")
	}
	return s.repo.ListDOLines(ctx, tenantID, doID)
}

func (s *Service) GetDOLine(ctx context.Context, tenantID, doID, lineID uint) (*DOLine, error) {
	return s.repo.GetDOLine(ctx, tenantID, doID, lineID)
}

func (s *Service) AddDOLine(ctx context.Context, tenantID, doID uint, req AddDOLineRequest) (*DOLine, error) {
	do, err := s.repo.GetDO(ctx, tenantID, doID)
	if err != nil {
		return nil, fmt.Errorf("delivery order not found")
	}
	if do.Status != DOStatusDraft {
		return nil, fmt.Errorf("cannot add items to a %s delivery order", do.Status)
	}
	line := &DOLine{
		DOID:       doID,
		TenantID:   tenantID,
		SOLineID:   req.SOLineID,
		LineNumber: len(do.Lines) + 1,
		ProductID:  req.ProductID,
		VariantID:  req.VariantID,
		LocationID: req.LocationID,
		UOMID:      req.UOMID,
		Quantity:   req.Quantity,
		UnitCost:   req.UnitCost,
		TotalCost:  req.Quantity * req.UnitCost,
		Notes:      req.Notes,
	}
	return line, s.repo.AddDOLine(ctx, line)
}

func (s *Service) UpdateDOLine(ctx context.Context, tenantID, doID, lineID uint, req UpdateDOLineRequest) (*DOLine, error) {
	do, err := s.repo.GetDO(ctx, tenantID, doID)
	if err != nil {
		return nil, fmt.Errorf("delivery order not found")
	}
	if do.Status != DOStatusDraft {
		return nil, fmt.Errorf("cannot edit items on a %s delivery order", do.Status)
	}
	line, err := s.repo.GetDOLine(ctx, tenantID, doID, lineID)
	if err != nil {
		return nil, fmt.Errorf("item not found")
	}
	line.SOLineID = req.SOLineID
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.LocationID = req.LocationID
	line.UOMID = req.UOMID
	line.Quantity = req.Quantity
	line.UnitCost = req.UnitCost
	line.TotalCost = req.Quantity * req.UnitCost
	line.Notes = req.Notes
	return line, s.repo.UpdateDOLine(ctx, line)
}

func (s *Service) DeleteDOLine(ctx context.Context, tenantID, doID, lineID uint) error {
	do, err := s.repo.GetDO(ctx, tenantID, doID)
	if err != nil {
		return fmt.Errorf("delivery order not found")
	}
	if do.Status != DOStatusDraft {
		return fmt.Errorf("cannot delete items from a %s delivery order", do.Status)
	}
	return s.repo.DeleteDOLine(ctx, tenantID, doID, lineID)
}

// ── Sales Invoices ────────────────────────────────────────────────────────────

func (s *Service) ListSIs(ctx context.Context, tenantID uint, customerID *uint, status string) ([]SalesInvoice, error) {
	return s.repo.ListSIs(ctx, tenantID, customerID, status)
}

func (s *Service) GetSI(ctx context.Context, tenantID, id uint) (*SalesInvoice, error) {
	return s.repo.GetSI(ctx, tenantID, id)
}

func (s *Service) CreateSI(ctx context.Context, tenantID, userID uint, req CreateSIRequest) (*SalesInvoice, error) {
	code := req.Code
	if code == "" {
		var err error
		code, err = s.repo.NextSICode(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate SI code: %w", err)
		}
	}
	d := req.InvoiceDate
	if d == "" {
		d = today()
	}
	er := req.ExchangeRate
	if er == 0 {
		er = 1
	}
	si := &SalesInvoice{
		TenantID:         tenantID,
		Code:             code,
		CustomerID:       req.CustomerID,
		SOID:             req.SOID,
		InvoiceDate:      d,
		DueDate:          req.DueDate,
		CustomerPONumber: req.CustomerPONumber,
		CurrencyID:       req.CurrencyID,
		ExchangeRate:     er,
		PaymentTermID:    req.PaymentTermID,
		Status:           SIStatusDraft,
		Notes:            req.Notes,
		CreatedBy:        &userID,
	}
	if err := s.repo.CreateSI(ctx, si); err != nil {
		if isDuplicate(err) {
			return nil, fmt.Errorf("sales invoice code '%s' already exists", code)
		}
		return nil, err
	}
	return si, nil
}

func (s *Service) AddSILine(ctx context.Context, tenantID, invoiceID uint, req AddSILineRequest) (*SILine, error) {
	si, err := s.repo.GetSI(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("sales invoice not found")
	}
	if si.Status != SIStatusDraft {
		return nil, fmt.Errorf("cannot add lines to a %s sales invoice", si.Status)
	}
	line := &SILine{
		InvoiceID:   invoiceID,
		TenantID:    tenantID,
		DOLineID:    req.DOLineID,
		LineNumber:  len(si.Lines) + 1,
		ProductID:   req.ProductID,
		VariantID:   req.VariantID,
		Description: req.Description,
		Quantity:    req.Quantity,
		UOMID:       req.UOMID,
		UnitPrice:   req.UnitPrice,
		DiscountPct: req.DiscountPct,
		TaxCodeID:   req.TaxCodeID,
		Notes:       req.Notes,
	}
	line.LineTotal = siLineTotal(line)
	if err := s.repo.AddSILine(ctx, line); err != nil {
		return nil, err
	}
	si.Lines = append(si.Lines, *line)
	calcSITotals(si)
	_ = s.repo.UpdateSI(ctx, si)
	return line, nil
}

func (s *Service) ListSILines(ctx context.Context, tenantID, invoiceID uint) ([]SILine, error) {
	if _, err := s.repo.GetSI(ctx, tenantID, invoiceID); err != nil {
		return nil, fmt.Errorf("sales invoice not found")
	}
	return s.repo.ListSILines(ctx, tenantID, invoiceID)
}

func (s *Service) DeleteSILine(ctx context.Context, tenantID, invoiceID, lineID uint) error {
	si, err := s.repo.GetSI(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("sales invoice not found")
	}
	if si.Status != SIStatusDraft {
		return fmt.Errorf("cannot delete lines from a %s sales invoice", si.Status)
	}
	if err := s.repo.DeleteSILine(ctx, tenantID, invoiceID, lineID); err != nil {
		return err
	}
	si, _ = s.repo.GetSI(ctx, tenantID, invoiceID)
	calcSITotals(si)
	return s.repo.UpdateSI(ctx, si)
}

// SIPostedPayload is the event body published when a sales invoice is posted.
type SIPostedPayload struct {
	InvoiceID    uint    `json:"invoice_id"`
	Code         string  `json:"code"`
	SOID         uint    `json:"so_id"`
	CustomerID   uint    `json:"customer_id"`
	InvoiceDate  string  `json:"invoice_date"`
	DueDate      *string `json:"due_date"`
	CurrencyID   uint    `json:"currency_id"`
	ExchangeRate float64 `json:"exchange_rate"`
	Subtotal     float64 `json:"subtotal"`
	TaxAmount    float64 `json:"tax_amount"`
	TotalAmount  float64 `json:"total_amount"`
	PostedBy     uint    `json:"posted_by"`
}

func (s *Service) PostSI(ctx context.Context, tenantID, userID, invoiceID uint) error {
	si, err := s.repo.GetSI(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("sales invoice not found")
	}
	if si.Status != SIStatusDraft {
		return fmt.Errorf("only DRAFT invoices can be posted")
	}
	if len(si.Lines) == 0 {
		return fmt.Errorf("invoice must have at least one line")
	}
	now := time.Now()
	si.Status = SIStatusPosted
	si.PostedBy = &userID
	si.PostedAt = &now
	if err := s.repo.UpdateSI(ctx, si); err != nil {
		return err
	}
	s.publish(ctx, events.TypeSIPosted, tenantID, userID, SIPostedPayload{
		InvoiceID:    si.ID,
		Code:         si.Code,
		SOID:         si.SOID,
		CustomerID:   si.CustomerID,
		InvoiceDate:  si.InvoiceDate,
		DueDate:      si.DueDate,
		CurrencyID:   si.CurrencyID,
		ExchangeRate: si.ExchangeRate,
		Subtotal:     si.Subtotal,
		TaxAmount:    si.TaxAmount,
		TotalAmount:  si.TotalAmount,
		PostedBy:     userID,
	})
	return nil
}

// SIPaymentPayload is the event body published when a payment is recorded on a sales invoice.
type SIPaymentPayload struct {
	InvoiceID   uint    `json:"invoice_id"`
	Code        string  `json:"code"`
	CustomerID  uint    `json:"customer_id"`
	Amount      float64 `json:"amount"`
	PaidAmount  float64 `json:"paid_amount"`
	TotalAmount float64 `json:"total_amount"`
	Remaining   float64 `json:"remaining"`
	Status      string  `json:"status"`
}

func (s *Service) RecordPayment(ctx context.Context, tenantID, invoiceID uint, req RecordPaymentRequest) error {
	si, err := s.repo.GetSI(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("sales invoice not found")
	}
	if si.Status == SIStatusDraft {
		return fmt.Errorf("invoice must be posted before recording payment")
	}
	if si.Status == SIStatusPaid {
		return fmt.Errorf("invoice is already fully paid")
	}
	if si.Status == SIStatusCancelled {
		return fmt.Errorf("cannot record payment on a cancelled invoice")
	}
	remaining := si.TotalAmount - si.PaidAmount
	if req.Amount > remaining {
		return fmt.Errorf("payment amount %.2f exceeds remaining balance %.2f", req.Amount, remaining)
	}
	si.PaidAmount += req.Amount
	if si.PaidAmount >= si.TotalAmount {
		si.Status = SIStatusPaid
	} else {
		si.Status = SIStatusPartial
	}
	if err := s.repo.UpdateSI(ctx, si); err != nil {
		return err
	}
	s.publish(ctx, events.TypeSIPaymentRecorded, tenantID, 0, SIPaymentPayload{
		InvoiceID:   si.ID,
		Code:        si.Code,
		CustomerID:  si.CustomerID,
		Amount:      req.Amount,
		PaidAmount:  si.PaidAmount,
		TotalAmount: si.TotalAmount,
		Remaining:   si.TotalAmount - si.PaidAmount,
		Status:      si.Status,
	})
	return nil
}

func (s *Service) CancelSI(ctx context.Context, tenantID, invoiceID uint) error {
	si, err := s.repo.GetSI(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("sales invoice not found")
	}
	if si.Status == SIStatusPaid || si.Status == SIStatusCancelled {
		return fmt.Errorf("cannot cancel a %s sales invoice", si.Status)
	}
	si.Status = SIStatusCancelled
	return s.repo.UpdateSI(ctx, si)
}
