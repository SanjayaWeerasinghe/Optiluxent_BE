package procurement

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"erp-system/internal/infrastructure/events"

	"github.com/google/uuid"
)

// QCAutoCreator is the minimal interface procurement needs from the inventory module
// to spin up a Material QC when a WITH_PO / WITHOUT_PO GRN is confirmed.
// inventory.Service satisfies this via its CreateAutoQC method.
type QCAutoLine struct {
	ProductID uint
	VariantID *uint
	Quantity  float64
}

type QCAutoCreator interface {
	CreateAutoQC(
		ctx context.Context,
		tenantID, userID uint,
		qcType, refType string,
		refID, warehouseID uint,
		notes string,
		lines []QCAutoLine,
	) error
}

// MOProducedBumper is the minimal interface procurement needs from the
// manufacturing module to increment ProductionOrder.produced_qty when a
// PRODUCTION_OUTPUT GRN is confirmed. manufacturing.Service satisfies this
// via its BumpProducedQty method.
type MOProducedBumper interface {
	BumpProducedQty(ctx context.Context, tenantID, moID, productID, uomID uint, qty float64) error
}

// DocumentTypeResolver looks up seeded Types by system_key and vice-versa.
// Satisfied by masterdata/documenttypes.Service.
type DocumentTypeResolver interface {
	// ResolveSystemKey returns the system_key for a Type id, "" if user-defined.
	ResolveSystemKey(ctx context.Context, tenantID, id uint) (string, error)
	// FindSystemType returns the DocumentType id for a (model, system_key)
	// pair — used when the caller wants to auto-select a system Type by key.
	FindSystemType(ctx context.Context, tenantID uint, model, systemKey string) (uint, error)
}

// ProductKindProvider exposes just the piece of the products module the
// stock-touching flows need: what Kind a given product is. Used to reject
// SERVICE-kind products from lines that would otherwise create stock rows.
// Satisfied by masterdata/products.Service.
type ProductKindProvider interface {
	Kind(ctx context.Context, tenantID, productID uint) (string, error)
}

// FinancePostGL is the tiny interface procurement uses to trigger a
// journal entry after the PI post + payment record. Satisfied by
// finance.Service directly (structural match).
type FinancePostGL interface {
	PostPIToGL(ctx context.Context, tenantID, userID, piID uint) (interface{}, error)
	// RecordInvoicePayment inserts a payments row via the finance module
	// so the payment log is populated in addition to updating the
	// invoice's paid_amount scalar. Best-effort — swallowed on error.
	RecordInvoicePaymentPI(ctx context.Context, tenantID, userID, piID uint, amount float64, method, referenceNo, notes string, bankAccountID *uint, paymentDate string) error
}

type Service struct {
	repo Repository
	bus  events.EventBus
	qc   QCAutoCreator
	mo   MOProducedBumper
	dt   DocumentTypeResolver
	pk   ProductKindProvider
	fin  FinancePostGL
}

func NewService(repo Repository) *Service                        { return &Service{repo: repo} }
func (s *Service) SetEventBus(bus events.EventBus)               { s.bus = bus }
func (s *Service) SetQCAutoCreator(qc QCAutoCreator)             { s.qc = qc }
func (s *Service) SetMOProducedBumper(mo MOProducedBumper)       { s.mo = mo }
func (s *Service) SetDocumentTypeResolver(dt DocumentTypeResolver) { s.dt = dt }
func (s *Service) SetProductKindProvider(pk ProductKindProvider)  { s.pk = pk }
func (s *Service) SetFinancePoster(fin FinancePostGL)             { s.fin = fin }

// ApplyInvoicePayment is called by the finance module when it records a
// payment against a PI. Bumps paid_amount + status. Returns supplier +
// currency so finance can stamp its payment row. Uses the same half-cent
// rounding tolerance as RecordPayment.
func (s *Service) ApplyInvoicePayment(ctx context.Context, tenantID, invoiceID uint, amount float64) (partyID, currencyID uint, err error) {
	inv, err := s.repo.GetInvoice(ctx, tenantID, invoiceID)
	if err != nil {
		return 0, 0, fmt.Errorf("purchase invoice not found")
	}
	if inv.Status == InvStatusDraft {
		return 0, 0, fmt.Errorf("invoice must be posted before recording payment")
	}
	if inv.Status == InvStatusCancelled {
		return 0, 0, fmt.Errorf("cannot record payment on a cancelled invoice")
	}
	if amount <= 0 {
		return 0, 0, fmt.Errorf("payment amount must be greater than zero")
	}
	remaining := inv.TotalAmount - inv.PaidAmount
	if amount > remaining+0.005 {
		return 0, 0, fmt.Errorf("payment amount %.2f exceeds remaining balance %.2f", amount, remaining)
	}
	inv.PaidAmount += amount
	if inv.PaidAmount >= inv.TotalAmount-0.005 {
		inv.Status = InvStatusPaid
	} else {
		inv.Status = InvStatusPartial
	}
	if err := s.repo.UpdateInvoice(ctx, inv); err != nil {
		return 0, 0, err
	}
	return inv.SupplierID, inv.CurrencyID, nil
}

// LookupInvoicePI implements finance.InvoiceLookup for PI headers.
func (s *Service) LookupInvoicePI(ctx context.Context, tenantID, invoiceID uint) (partyID, currencyID uint, totalAmount, taxAmount float64, err error) {
	inv, err := s.repo.GetInvoice(ctx, tenantID, invoiceID)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return inv.SupplierID, inv.CurrencyID, inv.TotalAmount, inv.TaxAmount, nil
}

// isNonStockableProduct returns true if the product is a SERVICE kind and
// therefore should never appear on stock-hitting lines (GRN items, DO lines,
// etc.). Silently returns false when the provider isn't wired — behaviour
// stays unchanged for callers that haven't opted in yet.
func (s *Service) isNonStockableProduct(ctx context.Context, tenantID, productID uint) bool {
	if s.pk == nil || productID == 0 {
		return false
	}
	kind, err := s.pk.Kind(ctx, tenantID, productID)
	if err != nil {
		return false
	}
	return kind == "SERVICE"
}

// resolveDocTypeKey looks up a DocumentType's system_key with a safe fallback.
// Returns "" (no behavioural switch) if the resolver is not wired, the id is
// zero, or the Type is user-defined.
func (s *Service) resolveDocTypeKey(ctx context.Context, tenantID uint, id *uint) string {
	if s.dt == nil || id == nil || *id == 0 {
		return ""
	}
	key, _ := s.dt.ResolveSystemKey(ctx, tenantID, *id)
	return key
}

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
	_ = s.bus.Publish(ctx, events.StreamProcurement, ev)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func today() string { return time.Now().Format("2006-01-02") }

func isDuplicate(err error) bool {
	return strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique")
}

func calcPOTotals(po *PurchaseOrder) {
	var sub, tax float64
	for _, l := range po.Lines {
		sub += l.LineTotal
		tax += l.TaxAmount
	}
	po.Subtotal = sub
	po.TaxAmount = tax
	po.TotalAmount = sub + tax - po.DiscountAmount
}

func calcInvoiceTotals(inv *PurchaseInvoice) {
	var sub, tax float64
	for _, l := range inv.Lines {
		sub += l.LineTotal
		tax += l.TaxAmount
	}
	inv.Subtotal = sub
	inv.TaxAmount = tax
	inv.TotalAmount = sub + tax - inv.DiscountAmount
}

// poLineTotal = qty × price × (1 - discount/100)
func poLineTotal(l *POLine) float64 {
	return l.Quantity * l.UnitPrice * (1 - l.DiscountPct/100)
}

func invoiceLineTotal(l *InvoiceLine) float64 {
	return l.Quantity * l.UnitPrice * (1 - l.DiscountPct/100)
}

func transferRatioOrOne(r float64) float64 {
	if r <= 0 {
		return 1
	}
	return r
}

// ── Purchase Requests ─────────────────────────────────────────────────────────

func (s *Service) ListPRs(ctx context.Context, tenantID uint, status string, limit, offset int) ([]PurchaseRequest, error) {
	return s.repo.ListPRs(ctx, tenantID, status, limit, offset)
}
func (s *Service) CountPRs(ctx context.Context, tenantID uint, status string) (int64, error) {
	return s.repo.CountPRs(ctx, tenantID, status)
}

func (s *Service) GetPR(ctx context.Context, tenantID, id uint) (*PurchaseRequest, error) {
	return s.repo.GetPR(ctx, tenantID, id)
}

func (s *Service) CreatePR(ctx context.Context, tenantID, userID uint, req *CreatePRRequest) (*PurchaseRequest, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "PURCHASE_REQUEST")
	if err != nil {
		return nil, fmt.Errorf("failed to generate PR code: %w", err)
	}
	d := req.RequestDate
	if d == "" {
		d = today()
	}
	pr := &PurchaseRequest{
		TenantID:       tenantID,
		Code:           code,
		DocumentTypeID: req.DocumentTypeID,
		RequestDate:    d,
		RequiredDate:   req.RequiredDate,
		RequestedBy:    req.RequestedBy,
		DepartmentID:   req.DepartmentID,
		Status:         PRStatusDraft,
		Notes:          req.Notes,
		CreatedBy:      &userID,
	}
	return pr, s.repo.CreatePR(ctx, pr)
}

func (s *Service) UpdatePR(ctx context.Context, tenantID, id uint, req *UpdatePRRequest) (*PurchaseRequest, error) {
	pr, err := s.repo.GetPR(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase request not found")
	}
	if pr.Status != PRStatusDraft {
		return nil, fmt.Errorf("only DRAFT purchase requests can be edited")
	}
	pr.DocumentTypeID = req.DocumentTypeID
	pr.RequiredDate = req.RequiredDate
	pr.RequestedBy = req.RequestedBy
	pr.DepartmentID = req.DepartmentID
	if req.Notes != "" {
		pr.Notes = req.Notes
	}
	return pr, s.repo.UpdatePR(ctx, pr)
}

func (s *Service) DeletePR(ctx context.Context, tenantID, id uint) error {
	pr, err := s.repo.GetPR(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("purchase request not found")
	}
	if pr.Status != PRStatusDraft {
		return fmt.Errorf("only DRAFT purchase requests can be deleted")
	}
	return s.repo.DeletePR(ctx, tenantID, id)
}

func (s *Service) SubmitPR(ctx context.Context, tenantID, id uint) (*PurchaseRequest, error) {
	pr, err := s.repo.GetPR(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase request not found")
	}
	if pr.Status != PRStatusDraft {
		return nil, fmt.Errorf("only DRAFT purchase requests can be submitted")
	}
	if len(pr.Lines) == 0 {
		return nil, fmt.Errorf("purchase request must have at least one item")
	}
	pr.Status = PRStatusPending
	return pr, s.repo.UpdatePR(ctx, pr)
}

func (s *Service) ApprovePR(ctx context.Context, tenantID, id, approverID uint) (*PurchaseRequest, error) {
	pr, err := s.repo.GetPR(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase request not found")
	}
	if pr.Status != PRStatusPending {
		return nil, fmt.Errorf("only PENDING_APPROVAL purchase requests can be approved")
	}
	now := time.Now()
	pr.Status = PRStatusApproved
	pr.ApprovedBy = &approverID
	pr.ApprovedAt = &now
	return pr, s.repo.UpdatePR(ctx, pr)
}

func (s *Service) RejectPR(ctx context.Context, tenantID, id, approverID uint) (*PurchaseRequest, error) {
	pr, err := s.repo.GetPR(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase request not found")
	}
	if pr.Status != PRStatusPending {
		return nil, fmt.Errorf("only PENDING_APPROVAL purchase requests can be rejected")
	}
	pr.Status = PRStatusRejected
	pr.ApprovedBy = &approverID
	return pr, s.repo.UpdatePR(ctx, pr)
}

func (s *Service) CancelPR(ctx context.Context, tenantID, id uint) (*PurchaseRequest, error) {
	pr, err := s.repo.GetPR(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase request not found")
	}
	if pr.Status == PRStatusApproved || pr.Status == PRStatusCancelled {
		return nil, fmt.Errorf("cannot cancel a %s purchase request", pr.Status)
	}
	pr.Status = PRStatusCancelled
	return pr, s.repo.UpdatePR(ctx, pr)
}

// ── PR Items ──────────────────────────────────────────────────────────────────

func (s *Service) ListPRItems(ctx context.Context, tenantID, prID uint) ([]PRLine, error) {
	if _, err := s.repo.GetPR(ctx, tenantID, prID); err != nil {
		return nil, fmt.Errorf("purchase request not found")
	}
	return s.repo.ListPRItems(ctx, prID)
}

func (s *Service) GetPRItem(ctx context.Context, tenantID, id uint) (*PRLine, error) {
	return s.repo.GetPRItem(ctx, tenantID, id)
}

func (s *Service) AddPRItem(ctx context.Context, tenantID, prID uint, req *AddPRItemRequest) (*PRLine, error) {
	pr, err := s.repo.GetPR(ctx, tenantID, prID)
	if err != nil {
		return nil, fmt.Errorf("purchase request not found")
	}
	if pr.Status != PRStatusDraft {
		return nil, fmt.Errorf("cannot add items to a %s purchase request", pr.Status)
	}
	line := &PRLine{
		PRID:           prID,
		TenantID:       tenantID,
		LineNumber:     len(pr.Lines) + 1,
		ProductID:      req.ProductID,
		VariantID:      req.VariantID,
		Description:    req.Description,
		Quantity:       req.Quantity,
		UOMID:          req.UOMID,
		EstimatedPrice: req.EstimatedPrice,
		TransferRatio:  transferRatioOrOne(req.TransferRatio),
		CurrencyID:     req.CurrencyID,
		Notes:          req.Notes,
	}
	return line, s.repo.AddPRItem(ctx, line)
}

func (s *Service) UpdatePRItem(ctx context.Context, tenantID, prID, itemID uint, req *UpdatePRItemRequest) (*PRLine, error) {
	pr, err := s.repo.GetPR(ctx, tenantID, prID)
	if err != nil {
		return nil, fmt.Errorf("purchase request not found")
	}
	if pr.Status != PRStatusDraft {
		return nil, fmt.Errorf("cannot edit items on a %s purchase request", pr.Status)
	}
	line, err := s.repo.GetPRItem(ctx, tenantID, itemID)
	if err != nil {
		return nil, fmt.Errorf("item not found")
	}
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.Description = req.Description
	line.Quantity = req.Quantity
	line.UOMID = req.UOMID
	line.EstimatedPrice = req.EstimatedPrice
	line.TransferRatio = transferRatioOrOne(req.TransferRatio)
	line.CurrencyID = req.CurrencyID
	line.Notes = req.Notes
	return line, s.repo.UpdatePRItem(ctx, line)
}

func (s *Service) DeletePRItem(ctx context.Context, tenantID, id uint) error {
	return s.repo.DeletePRItem(ctx, tenantID, id)
}

// ── Purchase Orders ───────────────────────────────────────────────────────────

func (s *Service) ListPOs(ctx context.Context, tenantID uint, status string, supplierID *uint, limit, offset int) ([]PurchaseOrder, error) {
	return s.repo.ListPOs(ctx, tenantID, status, supplierID, limit, offset)
}
func (s *Service) CountPOs(ctx context.Context, tenantID uint, status string, supplierID *uint) (int64, error) {
	return s.repo.CountPOs(ctx, tenantID, status, supplierID)
}

func (s *Service) GetPO(ctx context.Context, tenantID, id uint) (*PurchaseOrder, error) {
	return s.repo.GetPO(ctx, tenantID, id)
}

func (s *Service) CreatePO(ctx context.Context, tenantID, userID uint, req *CreatePORequest) (*PurchaseOrder, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "PURCHASE_ORDER")
	if err != nil {
		return nil, fmt.Errorf("failed to generate PO code: %w", err)
	}
	d := req.OrderDate
	if d == "" {
		d = today()
	}
	er := req.ExchangeRate
	if er == 0 {
		er = 1
	}
	po := &PurchaseOrder{
		TenantID:       tenantID,
		Code:           code,
		DocumentTypeID: req.DocumentTypeID,
		SupplierID:     req.SupplierID,
		PRID:           req.PRID,
		OrderDate:      d,
		ExpectedDate:   req.ExpectedDate,
		CurrencyID:     req.CurrencyID,
		ExchangeRate:   er,
		PaymentTermID:  req.PaymentTermID,
		WarehouseID:    req.WarehouseID,
		Status:         POStatusDraft,
		Notes:          req.Notes,
		CreatedBy:      &userID,
	}
	if err := s.repo.CreatePO(ctx, po); err != nil {
		if isDuplicate(err) {
			return nil, fmt.Errorf("purchase order code '%s' already exists", code)
		}
		return nil, err
	}
	// Auto-create a draft Purchase Invoice so the expected bill exists from day one.
	// Lines are populated later as GRNs are confirmed (actual received quantities).
	if invCode, err := s.repo.NextCode(ctx, tenantID, "PURCHASE_INVOICE"); err == nil {
		inv := &PurchaseInvoice{
			TenantID:      tenantID,
			Code:          invCode,
			SupplierID:    po.SupplierID,
			POID:          po.ID,
			InvoiceDate:   d,
			CurrencyID:    po.CurrencyID,
			ExchangeRate:  er,
			PaymentTermID: po.PaymentTermID,
			Status:        InvStatusDraft,
			CreatedBy:     &userID,
		}
		_ = s.repo.CreateInvoice(ctx, inv)
	}
	return po, nil
}

func (s *Service) UpdatePO(ctx context.Context, tenantID, id uint, req *UpdatePORequest) (*PurchaseOrder, error) {
	po, err := s.repo.GetPO(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found")
	}
	if po.Status != POStatusDraft {
		return nil, fmt.Errorf("only DRAFT purchase orders can be edited")
	}
	po.DocumentTypeID = req.DocumentTypeID
	po.ExpectedDate = req.ExpectedDate
	po.PaymentTermID = req.PaymentTermID
	if req.ExchangeRate != nil {
		po.ExchangeRate = *req.ExchangeRate
	}
	if req.DiscountAmount != nil {
		po.DiscountAmount = *req.DiscountAmount
	}
	if req.Notes != "" {
		po.Notes = req.Notes
	}
	calcPOTotals(po)
	return po, s.repo.UpdatePO(ctx, po)
}

func (s *Service) DeletePO(ctx context.Context, tenantID, id uint) error {
	po, err := s.repo.GetPO(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("purchase order not found")
	}
	if po.Status != POStatusDraft {
		return fmt.Errorf("only DRAFT purchase orders can be deleted")
	}
	return s.repo.DeletePO(ctx, tenantID, id)
}

func (s *Service) ConfirmPO(ctx context.Context, tenantID, id, userID uint) (*PurchaseOrder, error) {
	po, err := s.repo.GetPO(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found")
	}
	if po.Status != POStatusDraft {
		return nil, fmt.Errorf("only DRAFT purchase orders can be confirmed")
	}
	if len(po.Lines) == 0 {
		return nil, fmt.Errorf("purchase order must have at least one item")
	}
	now := time.Now()
	po.Status = POStatusConfirmed
	po.ConfirmedBy = &userID
	po.ConfirmedAt = &now
	if err := s.repo.UpdatePO(ctx, po); err != nil {
		return nil, err
	}
	return po, nil
}

func (s *Service) CancelPO(ctx context.Context, tenantID, id uint) (*PurchaseOrder, error) {
	po, err := s.repo.GetPO(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found")
	}
	if po.Status == POStatusReceived || po.Status == POStatusCancelled {
		return nil, fmt.Errorf("cannot cancel a %s purchase order", po.Status)
	}
	po.Status = POStatusCancelled
	return po, s.repo.UpdatePO(ctx, po)
}

// ── PO Items ──────────────────────────────────────────────────────────────────

func (s *Service) ListPOItems(ctx context.Context, tenantID, poID uint) ([]POLine, error) {
	if _, err := s.repo.GetPO(ctx, tenantID, poID); err != nil {
		return nil, fmt.Errorf("purchase order not found")
	}
	return s.repo.ListPOItems(ctx, poID)
}

func (s *Service) GetPOItem(ctx context.Context, tenantID, id uint) (*POLine, error) {
	return s.repo.GetPOItem(ctx, tenantID, id)
}

func (s *Service) AddPOItem(ctx context.Context, tenantID, poID uint, req *AddPOItemRequest) (*POLine, error) {
	po, err := s.repo.GetPO(ctx, tenantID, poID)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found")
	}
	if po.Status != POStatusDraft {
		return nil, fmt.Errorf("cannot add items to a %s purchase order", po.Status)
	}
	// Cap against source PR when this PO was created from one.
	if po.PRID != nil {
		if err := s.ValidatePOQtyAgainstPR(ctx, tenantID, *po.PRID, req.ProductID, req.Quantity, 0); err != nil {
			return nil, err
		}
	}
	line := &POLine{
		POID:          poID,
		TenantID:      tenantID,
		LineNumber:    len(po.Lines) + 1,
		ProductID:     req.ProductID,
		VariantID:     req.VariantID,
		Description:   req.Description,
		Quantity:      req.Quantity,
		UOMID:         req.UOMID,
		UnitPrice:     req.UnitPrice,
		DiscountPct:   req.DiscountPct,
		TaxCodeID:     req.TaxCodeID,
		TransferRatio: transferRatioOrOne(req.TransferRatio),
		Notes:         req.Notes,
	}
	line.LineTotal = poLineTotal(line)
	if err := s.repo.AddPOItem(ctx, line); err != nil {
		return nil, err
	}
	po.Lines = append(po.Lines, *line)
	calcPOTotals(po)
	_ = s.repo.UpdatePO(ctx, po)
	return line, nil
}

func (s *Service) UpdatePOItem(ctx context.Context, tenantID, poID, itemID uint, req *UpdatePOItemRequest) (*POLine, error) {
	po, err := s.repo.GetPO(ctx, tenantID, poID)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found")
	}
	if po.Status != POStatusDraft {
		return nil, fmt.Errorf("cannot edit items on a %s purchase order", po.Status)
	}
	line, err := s.repo.GetPOItem(ctx, tenantID, itemID)
	if err != nil {
		return nil, fmt.Errorf("item not found")
	}
	// Cap against source PR when this PO was created from one. Exclude the
	// current line so its own pre-edit qty doesn't self-count.
	if po.PRID != nil {
		if err := s.ValidatePOQtyAgainstPR(ctx, tenantID, *po.PRID, req.ProductID, req.Quantity, line.ID); err != nil {
			return nil, err
		}
	}
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.Description = req.Description
	line.Quantity = req.Quantity
	line.UOMID = req.UOMID
	line.UnitPrice = req.UnitPrice
	line.DiscountPct = req.DiscountPct
	line.TaxCodeID = req.TaxCodeID
	line.TransferRatio = transferRatioOrOne(req.TransferRatio)
	line.Notes = req.Notes
	line.LineTotal = poLineTotal(line)
	if err := s.repo.UpdatePOItem(ctx, line); err != nil {
		return nil, err
	}
	// Recalculate PO totals
	po, _ = s.repo.GetPO(ctx, tenantID, poID)
	calcPOTotals(po)
	_ = s.repo.UpdatePO(ctx, po)
	return line, nil
}

func (s *Service) DeletePOItem(ctx context.Context, tenantID, poID, itemID uint) error {
	po, err := s.repo.GetPO(ctx, tenantID, poID)
	if err != nil {
		return fmt.Errorf("purchase order not found")
	}
	if po.Status != POStatusDraft {
		return fmt.Errorf("cannot delete items from a %s purchase order", po.Status)
	}
	if err := s.repo.DeletePOItem(ctx, tenantID, itemID); err != nil {
		return err
	}
	po, _ = s.repo.GetPO(ctx, tenantID, poID)
	calcPOTotals(po)
	return s.repo.UpdatePO(ctx, po)
}

// ── Goods Receipts ────────────────────────────────────────────────────────────

func (s *Service) ListGRNsByMO(ctx context.Context, tenantID, moID uint) ([]GoodsReceipt, error) {
	return s.repo.ListGRNsByMO(ctx, tenantID, moID)
}

func (s *Service) ListGRNs(ctx context.Context, tenantID uint, status string, poID *uint, limit, offset int) ([]GoodsReceipt, error) {
	return s.repo.ListGRNs(ctx, tenantID, status, poID, limit, offset)
}
func (s *Service) CountGRNs(ctx context.Context, tenantID uint, status string, poID *uint) (int64, error) {
	return s.repo.CountGRNs(ctx, tenantID, status, poID)
}

func (s *Service) GetGRN(ctx context.Context, tenantID, id uint) (*GoodsReceipt, error) {
	return s.repo.GetGRN(ctx, tenantID, id)
}

func (s *Service) CreateGRN(ctx context.Context, tenantID, userID uint, req *CreateGRNRequest) (*GoodsReceipt, error) {
	// Guard: cannot link a GRN to a PO whose invoice is already posted/paid
	if req.POID != nil {
		if locked, err := s.repo.POInvoiceIsLocked(ctx, tenantID, *req.POID); err == nil && locked {
			return nil, fmt.Errorf("purchase order already has a completed invoice — cannot add more receipts")
		}
	}
	code, err := s.repo.NextCode(ctx, tenantID, "GOODS_RECEIPT")
	if err != nil {
		return nil, fmt.Errorf("failed to generate GRN code: %w", err)
	}
	d := req.ReceiptDate
	if d == "" {
		d = today()
	}
	// DocumentTypeID is optional on create; if the caller didn't pick one,
	// auto-select the system Type that matches the linkage on the request.
	docTypeID := req.DocumentTypeID
	if docTypeID == nil {
		var wantKey string
		switch {
		case req.MOID != nil:
			wantKey = GRNTypeProductionOutput
		case req.POID != nil:
			wantKey = GRNTypeWithPO
		default:
			wantKey = GRNTypeWithoutPO
		}
		if s.dt != nil {
			if id, err := s.dt.FindSystemType(ctx, tenantID, "GRN", wantKey); err == nil && id != 0 {
				docTypeID = &id
			}
		}
	}
	grn := &GoodsReceipt{
		TenantID:       tenantID,
		Code:           code,
		DocumentTypeID: docTypeID,
		POID:           req.POID,
		MOID:           req.MOID,
		SupplierID:     req.SupplierID,
		ReceiptDate:    d,
		WarehouseID:    req.WarehouseID,
		Status:         GRNStatusDraft,
		Notes:          req.Notes,
		CreatedBy:      &userID,
	}
	return grn, s.repo.CreateGRN(ctx, grn)
}

func (s *Service) CancelGRN(ctx context.Context, tenantID, id uint) error {
	row, err := s.repo.GetGRN(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("goods receipt not found")
	}
	if row.Status != GRNStatusDraft {
		return fmt.Errorf("only DRAFT can be cancelled")
	}
	return s.repo.SetGRNStatus(ctx, tenantID, id, GRNStatusCancelled)
}

func (s *Service) ConfirmGRN(ctx context.Context, tenantID, id, userID uint) error {
	grn, err := s.repo.GetGRN(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("goods receipt not found")
	}
	if grn.Status != GRNStatusDraft {
		return fmt.Errorf("only DRAFT goods receipts can be confirmed")
	}
	if len(grn.Lines) == 0 {
		return fmt.Errorf("goods receipt must have at least one item")
	}
	if err := s.repo.ConfirmGRN(ctx, tenantID, id, userID); err != nil {
		return err
	}
	// Auto-append GRN lines to the PO's draft invoice
	if grn.POID != nil {
		s.appendGRNLinesToDraftInvoice(ctx, tenantID, grn)
	}
	// Resolve the seeded system_key for the GRN's DocumentType. Empty means
	// the Type is user-defined and no downstream behaviour is triggered.
	sysKey := s.resolveDocTypeKey(ctx, tenantID, grn.DocumentTypeID)

	// Auto-create a QC for GRNs whose type warrants one.
	// WITH_PO / WITHOUT_PO           → MATERIAL_QC (inbound goods)
	// PRODUCTION_OUTPUT              → PRODUCT_QC  (own production, bottled goods)
	// CUSTOMER_RETURN / PRODUCTION_RETURN / user-defined → no QC.
	if s.qc != nil {
		var qcSysKey string
		switch sysKey {
		case GRNTypeWithPO, GRNTypeWithoutPO:
			qcSysKey = "MATERIAL_QC"
		case GRNTypeProductionOutput:
			qcSysKey = "PRODUCT_QC"
		}
		if qcSysKey != "" {
			lines := make([]QCAutoLine, 0, len(grn.Lines))
			for _, gl := range grn.Lines {
				lines = append(lines, QCAutoLine{
					ProductID: gl.ProductID,
					VariantID: gl.VariantID,
					Quantity:  gl.Quantity,
				})
			}
			_ = s.qc.CreateAutoQC(ctx, tenantID, userID,
				qcSysKey, "GRN", grn.ID, grn.WarehouseID,
				"Auto-created from "+grn.Code+" ("+sysKey+")",
				lines)
		}
	}
	// Bump the Manufacturing Order's produced_qty for PRODUCTION_OUTPUT GRNs.
	if s.mo != nil && sysKey == GRNTypeProductionOutput && grn.MOID != nil {
		for _, gl := range grn.Lines {
			_ = s.mo.BumpProducedQty(ctx, tenantID, *grn.MOID, gl.ProductID, gl.UOMID, gl.Quantity)
		}
	}
	return nil
}

// appendGRNLinesToDraftInvoice adds lines from a confirmed GRN to the PO's draft invoice.
// Prices are sourced from PO lines; falls back to GRN unit_cost if no PO line is linked.
// This is best-effort — failures do not roll back the GRN confirmation.
func (s *Service) appendGRNLinesToDraftInvoice(ctx context.Context, tenantID uint, grn *GoodsReceipt) {
	inv, err := s.repo.GetDraftInvoiceByPOID(ctx, tenantID, *grn.POID)
	if err != nil {
		return // no draft invoice — nothing to append to
	}
	nextLine := len(inv.Lines) + 1
	for _, grnLine := range grn.Lines {
		unitPrice := grnLine.UnitCost
		if grnLine.POLineID != nil {
			if poLine, err := s.repo.GetPOItem(ctx, tenantID, *grnLine.POLineID); err == nil {
				unitPrice = poLine.UnitPrice
			}
		}
		grnLineID := grnLine.ID
		line := &InvoiceLine{
			InvoiceID:  inv.ID,
			TenantID:   tenantID,
			GRNLineID:  &grnLineID,
			LineNumber: nextLine,
			ProductID:  grnLine.ProductID,
			VariantID:  grnLine.VariantID,
			Quantity:   grnLine.Quantity,
			UOMID:      grnLine.UOMID,
			UnitPrice:  unitPrice,
		}
		line.LineTotal = invoiceLineTotal(line)
		if err := s.repo.AddInvoiceLine(ctx, line); err == nil {
			inv.Lines = append(inv.Lines, *line)
			nextLine++
		}
	}
	calcInvoiceTotals(inv)
	_ = s.repo.UpdateInvoice(ctx, inv)
}

// ── GRN Items ─────────────────────────────────────────────────────────────────

func (s *Service) ListGRNItems(ctx context.Context, tenantID, grnID uint) ([]GRNLine, error) {
	if _, err := s.repo.GetGRN(ctx, tenantID, grnID); err != nil {
		return nil, fmt.Errorf("goods receipt not found")
	}
	return s.repo.ListGRNItems(ctx, grnID)
}

func (s *Service) GetGRNItem(ctx context.Context, tenantID, id uint) (*GRNLine, error) {
	return s.repo.GetGRNItem(ctx, tenantID, id)
}

func (s *Service) AddGRNItem(ctx context.Context, tenantID, grnID uint, req *AddGRNItemRequest) (*GRNLine, error) {
	grn, err := s.repo.GetGRN(ctx, tenantID, grnID)
	if err != nil {
		return nil, fmt.Errorf("goods receipt not found")
	}
	if grn.Status != GRNStatusDraft {
		return nil, fmt.Errorf("cannot add items to a %s goods receipt", grn.Status)
	}
	if s.isNonStockableProduct(ctx, tenantID, req.ProductID) {
		return nil, fmt.Errorf("service products cannot be added to a goods receipt")
	}
	// Cap against source PO line when this GRN line is tied to one.
	if req.POLineID != nil {
		if err := s.ValidateGRNQtyAgainstPOLine(ctx, tenantID, *req.POLineID, req.Quantity, 0); err != nil {
			return nil, err
		}
	}
	ratio := transferRatioOrOne(req.TransferRatio)
	line := &GRNLine{
		GRNID:         grnID,
		TenantID:      tenantID,
		POLineID:      req.POLineID,
		LineNumber:    len(grn.Lines) + 1,
		ProductID:     req.ProductID,
		VariantID:     req.VariantID,
		Quantity:      req.Quantity,
		UOMID:         req.UOMID,
		LocationID:    req.LocationID,
		UnitCost:      req.UnitCost,
		TotalCost:     req.Quantity * req.UnitCost,
		TransferRatio: ratio,
		Notes:         req.Notes,
	}
	return line, s.repo.AddGRNItem(ctx, line)
}

func (s *Service) UpdateGRNItem(ctx context.Context, tenantID, grnID, itemID uint, req *UpdateGRNItemRequest) (*GRNLine, error) {
	grn, err := s.repo.GetGRN(ctx, tenantID, grnID)
	if err != nil {
		return nil, fmt.Errorf("goods receipt not found")
	}
	if grn.Status != GRNStatusDraft {
		return nil, fmt.Errorf("cannot edit items on a %s goods receipt", grn.Status)
	}
	line, err := s.repo.GetGRNItem(ctx, tenantID, itemID)
	if err != nil {
		return nil, fmt.Errorf("item not found")
	}
	// Cap against source PO line. previousQty = this line's current qty so the
	// delta (new - previous) is what we compare against remaining.
	if line.POLineID != nil {
		if err := s.ValidateGRNQtyAgainstPOLine(ctx, tenantID, *line.POLineID, req.Quantity, line.Quantity); err != nil {
			return nil, err
		}
	}
	ratio := transferRatioOrOne(req.TransferRatio)
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.Quantity = req.Quantity
	line.UOMID = req.UOMID
	line.LocationID = req.LocationID
	line.UnitCost = req.UnitCost
	line.TotalCost = req.Quantity * req.UnitCost
	line.TransferRatio = ratio
	line.Notes = req.Notes
	return line, s.repo.UpdateGRNItem(ctx, line)
}

// ── Auto-Invoice from GRN ─────────────────────────────────────────────────────

// AutoInvoiceFromGRN creates a draft purchase invoice pre-populated from a
// confirmed GRN. Quantities come from GRN lines; prices come from the linked
// PO lines (falling back to GRN unit_cost when no PO line is linked).
func (s *Service) AutoInvoiceFromGRN(ctx context.Context, tenantID, grnID, userID uint) (*PurchaseInvoice, error) {
	grn, err := s.repo.GetGRN(ctx, tenantID, grnID)
	if err != nil {
		return nil, fmt.Errorf("goods receipt not found")
	}
	if grn.Status != GRNStatusConfirmed {
		return nil, fmt.Errorf("only CONFIRMED goods receipts can auto-generate an invoice")
	}
	if grn.POID == nil {
		return nil, fmt.Errorf("goods receipt is not linked to a purchase order")
	}
	if len(grn.Lines) == 0 {
		return nil, fmt.Errorf("goods receipt has no items")
	}

	po, err := s.repo.GetPO(ctx, tenantID, *grn.POID)
	if err != nil {
		return nil, fmt.Errorf("linked purchase order not found")
	}

	code, err := s.repo.NextCode(ctx, tenantID, "PURCHASE_INVOICE")
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice code: %w", err)
	}

	inv := &PurchaseInvoice{
		TenantID:      tenantID,
		Code:          code,
		SupplierID:    po.SupplierID,
		POID:          po.ID,
		InvoiceDate:   today(),
		CurrencyID:    po.CurrencyID,
		ExchangeRate:  po.ExchangeRate,
		PaymentTermID: po.PaymentTermID,
		Status:        InvStatusDraft,
		CreatedBy:     &userID,
	}
	if err := s.repo.CreateInvoice(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	for i, grnLine := range grn.Lines {
		unitPrice := grnLine.UnitCost // fallback: GRN cost → invoice price

		// Pull PO unit price for proper 3-way match
		if grnLine.POLineID != nil {
			if poLine, err := s.repo.GetPOItem(ctx, tenantID, *grnLine.POLineID); err == nil {
				unitPrice = poLine.UnitPrice
			}
		}

		grnLineID := grnLine.ID
		line := &InvoiceLine{
			InvoiceID:  inv.ID,
			TenantID:   tenantID,
			GRNLineID:  &grnLineID,
			LineNumber: i + 1,
			ProductID:  grnLine.ProductID,
			VariantID:  grnLine.VariantID,
			Quantity:   grnLine.Quantity,
			UOMID:      grnLine.UOMID,
			UnitPrice:  unitPrice,
		}
		line.LineTotal = invoiceLineTotal(line)
		if err := s.repo.AddInvoiceLine(ctx, line); err != nil {
			return inv, fmt.Errorf("created invoice %s but failed on line %d: %w", code, i+1, err)
		}
		inv.Lines = append(inv.Lines, *line)
	}

	calcInvoiceTotals(inv)
	_ = s.repo.UpdateInvoice(ctx, inv)

	return inv, nil
}

// ── Purchase Invoices ─────────────────────────────────────────────────────────

func (s *Service) ListInvoices(ctx context.Context, tenantID uint, status string, supplierID *uint, limit, offset int) ([]PurchaseInvoice, error) {
	return s.repo.ListInvoices(ctx, tenantID, status, supplierID, limit, offset)
}
func (s *Service) CountInvoices(ctx context.Context, tenantID uint, status string, supplierID *uint) (int64, error) {
	return s.repo.CountInvoices(ctx, tenantID, status, supplierID)
}

func (s *Service) GetInvoice(ctx context.Context, tenantID, id uint) (*PurchaseInvoice, error) {
	return s.repo.GetInvoice(ctx, tenantID, id)
}

func (s *Service) CreateInvoice(ctx context.Context, tenantID, userID uint, req *CreateInvoiceRequest) (*PurchaseInvoice, error) {
	// Validate PO belongs to tenant
	po, err := s.repo.GetPO(ctx, tenantID, req.POID)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found")
	}
	code, err := s.repo.NextCode(ctx, tenantID, "PURCHASE_INVOICE")
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice code: %w", err)
	}
	d := req.InvoiceDate
	if d == "" {
		d = today()
	}
	er := req.ExchangeRate
	if er == 0 {
		er = po.ExchangeRate
	}
	payTerm := req.PaymentTermID
	if payTerm == nil {
		payTerm = po.PaymentTermID
	}
	inv := &PurchaseInvoice{
		TenantID:            tenantID,
		Code:                code,
		DocumentTypeID:      req.DocumentTypeID,
		SupplierID:          po.SupplierID,
		POID:                req.POID,
		InvoiceDate:         d,
		DueDate:             req.DueDate,
		SupplierInvoiceNo:   req.SupplierInvoiceNo,
		SupplierInvoiceDate: req.SupplierInvoiceDate,
		CurrencyID:          req.CurrencyID,
		ExchangeRate:        er,
		PaymentTermID:       payTerm,
		Status:              InvStatusDraft,
		Notes:               req.Notes,
		CreatedBy:           &userID,
	}
	return inv, s.repo.CreateInvoice(ctx, inv)
}

func (s *Service) AddInvoiceLine(ctx context.Context, tenantID, invoiceID uint, req *AddInvoiceLineRequest) (*InvoiceLine, error) {
	inv, err := s.repo.GetInvoice(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("purchase invoice not found")
	}
	if inv.Status != InvStatusDraft {
		return nil, fmt.Errorf("cannot add lines to a %s purchase invoice", inv.Status)
	}
	line := &InvoiceLine{
		InvoiceID:         invoiceID,
		TenantID:          tenantID,
		GRNLineID:         req.GRNLineID,
		LineNumber:        len(inv.Lines) + 1,
		ProductID:         req.ProductID,
		VariantID:         req.VariantID,
		Description:       req.Description,
		Quantity:          req.Quantity,
		UOMID:             req.UOMID,
		UnitPrice:         req.UnitPrice,
		DiscountPct:       req.DiscountPct,
		TaxCodeID:         req.TaxCodeID,
		SupplierInvoiceNo: req.SupplierInvoiceNo,
		Notes:             req.Notes,
	}
	line.LineTotal = invoiceLineTotal(line)
	if err := s.repo.AddInvoiceLine(ctx, line); err != nil {
		return nil, err
	}
	inv.Lines = append(inv.Lines, *line)
	calcInvoiceTotals(inv)
	_ = s.repo.UpdateInvoice(ctx, inv)
	return line, nil
}

func (s *Service) DeleteInvoiceLine(ctx context.Context, tenantID, invoiceID, lineID uint) error {
	inv, err := s.repo.GetInvoice(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("purchase invoice not found")
	}
	if inv.Status != InvStatusDraft {
		return fmt.Errorf("cannot delete lines from a %s invoice", inv.Status)
	}
	if err := s.repo.DeleteInvoiceLine(ctx, tenantID, lineID); err != nil {
		return err
	}
	inv, _ = s.repo.GetInvoice(ctx, tenantID, invoiceID)
	calcInvoiceTotals(inv)
	return s.repo.UpdateInvoice(ctx, inv)
}

// InvoicePostedPayload is the event body published when a purchase invoice is posted.
// A future financial module can subscribe to procurement.invoice.posted on
// erp:procurement to create AP liability entries and GL postings.
type InvoicePostedPayload struct {
	InvoiceID      uint    `json:"invoice_id"`
	Code           string  `json:"code"`
	POID           uint    `json:"po_id"`
	SupplierID     uint    `json:"supplier_id"`
	InvoiceDate    string  `json:"invoice_date"`
	DueDate        *string `json:"due_date"`
	CurrencyID     uint    `json:"currency_id"`
	ExchangeRate   float64 `json:"exchange_rate"`
	Subtotal       float64 `json:"subtotal"`
	TaxAmount      float64 `json:"tax_amount"`
	DiscountAmount float64 `json:"discount_amount"`
	TotalAmount    float64 `json:"total_amount"`
	PostedBy       uint    `json:"posted_by"`
}

func (s *Service) PostInvoice(ctx context.Context, tenantID, id, userID uint) (*PurchaseInvoice, error) {
	inv, err := s.repo.GetInvoice(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase invoice not found")
	}
	if inv.Status != InvStatusDraft {
		return nil, fmt.Errorf("only DRAFT invoices can be posted")
	}
	if len(inv.Lines) == 0 {
		return nil, fmt.Errorf("invoice must have at least one line")
	}
	now := time.Now()
	inv.Status = InvStatusPosted
	inv.PostedBy = &userID
	inv.PostedAt = &now
	if err := s.repo.UpdateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	s.publish(ctx, events.TypeInvoicePosted, tenantID, userID, InvoicePostedPayload{
		InvoiceID:      inv.ID,
		Code:           inv.Code,
		POID:           inv.POID,
		SupplierID:     inv.SupplierID,
		InvoiceDate:    inv.InvoiceDate,
		DueDate:        inv.DueDate,
		CurrencyID:     inv.CurrencyID,
		ExchangeRate:   inv.ExchangeRate,
		Subtotal:       inv.Subtotal,
		TaxAmount:      inv.TaxAmount,
		DiscountAmount: inv.DiscountAmount,
		TotalAmount:    inv.TotalAmount,
		PostedBy:       userID,
	})
	// Best-effort GL post — failure only logs, doesn't roll back the post.
	if s.fin != nil {
		_, _ = s.fin.PostPIToGL(ctx, tenantID, userID, inv.ID)
	}
	return inv, nil
}

// PaymentRecordedPayload is the event body published when a payment is recorded.
// A future financial module can subscribe to procurement.payment.recorded on
// erp:procurement to create cash/bank credit entries and update AP balances.
type PaymentRecordedPayload struct {
	InvoiceID   uint    `json:"invoice_id"`
	Code        string  `json:"code"`
	SupplierID  uint    `json:"supplier_id"`
	Amount      float64 `json:"amount"`
	PaidAmount  float64 `json:"paid_amount"`
	TotalAmount float64 `json:"total_amount"`
	Remaining   float64 `json:"remaining"`
	Status      string  `json:"status"`
	PaymentDate string  `json:"payment_date"`
	PaymentRef  string  `json:"payment_ref"`
}

func (s *Service) RecordPayment(ctx context.Context, tenantID, id uint, req *RecordPaymentRequest) (*PurchaseInvoice, error) {
	inv, err := s.repo.GetInvoice(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase invoice not found")
	}
	if inv.Status == InvStatusDraft {
		return nil, fmt.Errorf("invoice must be posted before recording payment")
	}
	if inv.Status == InvStatusPaid {
		return nil, fmt.Errorf("invoice is already fully paid")
	}
	if inv.Status == InvStatusCancelled {
		return nil, fmt.Errorf("cannot record payment on a cancelled invoice")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("payment amount must be greater than zero")
	}
	remaining := inv.TotalAmount - inv.PaidAmount
	// Allow half a cent of overpay to absorb float rounding — the same
	// tolerance used when deciding when the invoice becomes fully paid.
	if req.Amount > remaining+0.005 {
		return nil, fmt.Errorf("payment amount %.2f exceeds remaining balance %.2f", req.Amount, remaining)
	}
	inv.PaidAmount += req.Amount
	// Fully paid when the running total covers the invoice, within half a
	// cent — accumulating multiple float payments can leave a sub-cent gap.
	if inv.PaidAmount >= inv.TotalAmount-0.005 {
		inv.Status = InvStatusPaid
	} else {
		inv.Status = InvStatusPartial
	}
	if err := s.repo.UpdateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	payDate := req.PaymentDate
	if payDate == "" {
		payDate = today()
	}
	s.publish(ctx, events.TypePaymentRecorded, tenantID, 0, PaymentRecordedPayload{
		InvoiceID:   inv.ID,
		Code:        inv.Code,
		SupplierID:  inv.SupplierID,
		Amount:      req.Amount,
		PaidAmount:  inv.PaidAmount,
		TotalAmount: inv.TotalAmount,
		Remaining:   inv.TotalAmount - inv.PaidAmount,
		Status:      inv.Status,
		PaymentDate: payDate,
		PaymentRef:  req.PaymentRef,
	})
	// The historic /pay endpoint didn't collect method or bank_account, so
	// we default to OTHER. Callers wanting rich metadata should POST to
	// /finance/payments instead. Best-effort — swallow errors.
	if s.fin != nil {
		_ = s.fin.RecordInvoicePaymentPI(ctx, tenantID, 0, inv.ID, req.Amount,
			"OTHER", req.PaymentRef, "", nil, payDate)
	}
	return inv, nil
}

func (s *Service) CancelInvoice(ctx context.Context, tenantID, id uint) (*PurchaseInvoice, error) {
	inv, err := s.repo.GetInvoice(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("purchase invoice not found")
	}
	if inv.Status == InvStatusPaid || inv.Status == InvStatusCancelled {
		return nil, fmt.Errorf("cannot cancel a %s invoice", inv.Status)
	}
	inv.Status = InvStatusCancelled
	return inv, s.repo.UpdateInvoice(ctx, inv)
}
