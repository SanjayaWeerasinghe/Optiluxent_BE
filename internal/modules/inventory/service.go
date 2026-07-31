package inventory

import (
	"context"
	"fmt"
	"time"
)

// Service implements all inventory business logic, wrapping the Repository.
type Service struct {
	repo  Repository
	dt    DocumentTypeResolver
	dmg   DamagedBinResolver
	alloc *AllocationService
}

// DocumentTypeResolver is the tiny interface inventory needs from the
// masterdata documenttypes service — the same shape as procurement's.
type DocumentTypeResolver interface {
	ResolveSystemKey(ctx context.Context, tenantID, id uint) (string, error)
	FindSystemType(ctx context.Context, tenantID uint, model, systemKey string) (uint, error)
}

// DamagedBinResolver resolves (or lazily creates) the DAMAGED-type
// storage_location for a given warehouse. Called from
// SubmitQualityCheck when a QC has qty_failed that needs somewhere to
// land — the flow never dead-ends because a warehouse forgot to seed
// a damaged bin. Satisfied by masterdata/inventory.Service.
type DamagedBinResolver interface {
	EnsureDamagedLocation(ctx context.Context, tenantID, warehouseID uint) (uint, error)
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// SetAllocationService — wired at Initialize so the MR/GI/GT line hooks and
// GI-confirm / GT-send flows can reserve/consume/release stock allocations.
func (s *Service) SetAllocationService(a *AllocationService) { s.alloc = a }

// AllocationService — the module-neutral surface exposed to cross-module
// callers (sales, procurement) via cmd/api/main.go's adapter wiring. Returns
// nil if allocations aren't wired yet — callers should nil-check.
func (s *Service) Allocation() *AllocationService { return s.alloc }

// SetDocumentTypeResolver is wired at startup from cmd/api/main.go so the
// service can auto-select seeded Types (e.g. MATERIAL_QC) when a caller
// doesn't explicitly pick one, and resolve system_key on legacy behaviour.
func (s *Service) SetDocumentTypeResolver(dt DocumentTypeResolver) { s.dt = dt }

// SetDamagedBinResolver wires the storage-locations service so
// SubmitQualityCheck can look up / create a DAMAGED bin for the
// failed-qty stock post.
func (s *Service) SetDamagedBinResolver(dmg DamagedBinResolver) { s.dmg = dmg }

// DamagedBinFor is a small pass-through used by the repository layer
// (which doesn't hold references to sibling modules) to reach the
// resolver via the service.
func (s *Service) DamagedBinFor(ctx context.Context, tenantID, warehouseID uint) (uint, error) {
	if s.dmg == nil {
		return 0, fmt.Errorf("damaged-bin resolver not wired")
	}
	return s.dmg.EnsureDamagedLocation(ctx, tenantID, warehouseID)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func today() string { return time.Now().Format("2006-01-02") }

// productKind — small helper used by the allocation hooks to decide whether
// a scope key should include production_id (refining segregation). Reads
// through the allocation repo's DB() handle since the main inventory repo
// interface doesn't expose one.
func (s *Service) productKind(ctx context.Context, productID uint) string {
	if s.alloc == nil {
		return ""
	}
	var kind string
	_ = s.alloc.repo.DB().WithContext(ctx).
		Raw(`SELECT kind FROM products WHERE id = ?`, productID).Row().Scan(&kind)
	return kind
}

// scopeForMR — build the allocation ScopeKey for an MR line. If the MR is
// linked to a Manufacturing Order and the product is REFINING_INTAKE,
// segregate against that MO's production_id.
func (s *Service) scopeForMR(ctx context.Context, mr *MaterialRequest, line *MRLine) ScopeKey {
	sk := ScopeKey{
		ProductID:   line.ProductID,
		VariantID:   line.VariantID,
		WarehouseID: mr.WarehouseID,
	}
	if mr.MOID != nil && s.productKind(ctx, line.ProductID) == "REFINING_INTAKE" {
		mo := *mr.MOID
		sk.ProductionID = &mo
	}
	return sk
}

// reserveForMRLine — the common Reserve call reused by AddMRLine and
// UpdateMRLine. Bubbles a clean error message up to the caller/API layer.
func (s *Service) reserveForMRLine(ctx context.Context, tenantID uint, mr *MaterialRequest, line *MRLine) error {
	_, err := s.alloc.Reserve(ctx, ReserveRequest{
		TenantID:    tenantID,
		ScopeKey:    s.scopeForMR(ctx, mr, line),
		Quantity:    line.RequestedQty,
		SourceType:  AllocSourceMRLine,
		SourceID:    line.ID,
		SourceDocID: mr.ID,
		Notes:       fmt.Sprintf("MR %s line %d", mr.Code, line.LineNumber),
		OnUpdate:    true, // idempotent: drops any pre-existing alloc for this line first
	})
	return err
}

// scopeForGI — the scope key for a GI line at its issue warehouse. Standalone
// GIs use general stock; MR-linked GIs consume the MR's allocation instead
// (so this helper is only called from the standalone branch).
func (s *Service) scopeForGI(ctx context.Context, gi *GoodsIssue, line *GILine, productionID *uint) ScopeKey {
	sk := ScopeKey{
		ProductID:   line.ProductID,
		VariantID:   line.VariantID,
		WarehouseID: gi.WarehouseID,
	}
	if productionID != nil && s.productKind(ctx, line.ProductID) == "REFINING_INTAKE" {
		sk.ProductionID = productionID
	}
	return sk
}

// scopeForGT — the scope key for a GT line at its source warehouse.
func (s *Service) scopeForGT(ctx context.Context, gt *GoodsTransfer, line *GTLine) ScopeKey {
	sk := ScopeKey{
		ProductID:   line.ProductID,
		VariantID:   line.VariantID,
		WarehouseID: gt.FromWarehouseID,
	}
	_ = ctx
	return sk
}

// ── Material Requests ─────────────────────────────────────────────────────────

func (s *Service) ListMRsByMO(ctx context.Context, tenantID, moID uint) ([]MaterialRequest, error) {
	return s.repo.ListMRsByMO(ctx, tenantID, moID)
}

func (s *Service) ListTransfersByMO(ctx context.Context, tenantID, moID uint) ([]GoodsTransfer, error) {
	return s.repo.ListTransfersByMO(ctx, tenantID, moID)
}

func (s *Service) ListIssuesByMO(ctx context.Context, tenantID, moID uint) ([]GoodsIssue, error) {
	return s.repo.ListIssuesByMO(ctx, tenantID, moID)
}

// ListQCsByGRNIDs returns the QC records whose reference_type is 'GRN' and
// reference_id is in the given list. Used by the MO dashboard to compute
// pass/fail metrics on production output GRNs.
func (s *Service) ListQCsByGRNIDs(ctx context.Context, tenantID uint, grnIDs []uint) ([]QualityCheck, error) {
	return s.repo.ListQCsByRefs(ctx, tenantID, "GRN", grnIDs)
}

func (s *Service) ListMRs(ctx context.Context, tenantID uint, status string, limit, offset int) ([]MaterialRequest, error) {
	return s.repo.ListMRs(ctx, tenantID, status, limit, offset)
}
func (s *Service) CountMRs(ctx context.Context, tenantID uint, status string) (int64, error) {
	return s.repo.CountMRs(ctx, tenantID, status)
}

func (s *Service) GetMR(ctx context.Context, tenantID, id uint) (*MaterialRequest, error) {
	return s.repo.GetMR(ctx, tenantID, id)
}

func (s *Service) CreateMR(ctx context.Context, tenantID, userID uint, req *CreateMRRequest) (*MaterialRequest, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "MATERIAL_REQUEST")
	if err != nil {
		return nil, fmt.Errorf("failed to generate material request code: %w", err)
	}
	d := req.NeededDate
	if d == "" {
		d = today()
	}
	mr := &MaterialRequest{
		TenantID:       tenantID,
		Code:           code,
		DocumentTypeID: req.DocumentTypeID,
		MOID:           req.MOID,
		RequestedBy:    req.RequestedBy,
		DepartmentID:   req.DepartmentID,
		WarehouseID:    req.WarehouseID,
		NeededDate:     d,
		Status:         MRStatusDraft,
		Notes:          req.Notes,
		CreatedBy:      userID,
	}
	return mr, s.repo.CreateMR(ctx, mr)
}

func (s *Service) UpdateMR(ctx context.Context, tenantID, id uint, req *UpdateMRRequest) (*MaterialRequest, error) {
	mr, err := s.repo.GetMR(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusDraft {
		return nil, fmt.Errorf("only DRAFT material requests can be edited")
	}
	mr.DocumentTypeID = req.DocumentTypeID
	mr.DepartmentID = req.DepartmentID
	mr.WarehouseID = req.WarehouseID
	if req.NeededDate != "" {
		mr.NeededDate = req.NeededDate
	}
	if req.Notes != "" {
		mr.Notes = req.Notes
	}
	return mr, s.repo.UpdateMR(ctx, mr)
}

// SubmitMR moves a DRAFT MR into PENDING_APPROVAL. Only after this can
// approvers Approve / Reject it — mirroring the workflow the FE now shows.
func (s *Service) SubmitMR(ctx context.Context, tenantID, id, userID uint) error {
	mr, err := s.repo.GetMR(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusDraft {
		return fmt.Errorf("can only submit DRAFT material requests")
	}
	if len(mr.Lines) == 0 {
		return fmt.Errorf("cannot submit an empty material request")
	}
	return s.repo.SetMRStatus(ctx, tenantID, id, MRStatusPendingApproval)
}

func (s *Service) ApproveMR(ctx context.Context, tenantID, id, userID uint) error {
	mr, err := s.repo.GetMR(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusPendingApproval {
		return fmt.Errorf("can only approve material requests awaiting approval")
	}
	return s.repo.ApproveMR(ctx, tenantID, id, userID)
}

func (s *Service) RejectMR(ctx context.Context, tenantID, id, userID uint, reason string) error {
	mr, err := s.repo.GetMR(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusPendingApproval {
		return fmt.Errorf("can only reject material requests awaiting approval")
	}
	if s.alloc != nil {
		_ = s.alloc.ReleaseByDoc(ctx, tenantID, AllocSourceMRLine, id)
	}
	return s.repo.RejectMR(ctx, tenantID, id, userID, reason)
}

// CancelMR — voluntary abandonment. Allowed from DRAFT or PENDING_APPROVAL.
func (s *Service) CancelMR(ctx context.Context, tenantID, id uint) error {
	mr, err := s.repo.GetMR(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusDraft && mr.Status != MRStatusPendingApproval {
		return fmt.Errorf("only DRAFT / PENDING_APPROVAL material requests can be cancelled")
	}
	if s.alloc != nil {
		_ = s.alloc.ReleaseByDoc(ctx, tenantID, AllocSourceMRLine, id)
	}
	return s.repo.SetMRStatus(ctx, tenantID, id, MRStatusCancelled)
}

func (s *Service) ListMRLines(ctx context.Context, tenantID, mrID uint) ([]MRLine, error) {
	if _, err := s.repo.GetMR(ctx, tenantID, mrID); err != nil {
		return nil, fmt.Errorf("material request not found")
	}
	return s.repo.ListMRLines(ctx, mrID)
}

func (s *Service) AddMRLine(ctx context.Context, tenantID, mrID uint, req *AddMRLineRequest) (*MRLine, error) {
	mr, err := s.repo.GetMR(ctx, tenantID, mrID)
	if err != nil {
		return nil, fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusDraft {
		return nil, fmt.Errorf("cannot add lines to a %s material request", mr.Status)
	}
	line := &MRLine{
		TenantID:     tenantID,
		MRID:         mrID,
		LineNumber:   len(mr.Lines) + 1,
		ProductID:    req.ProductID,
		VariantID:    req.VariantID,
		UOMId:        req.UOMId,
		RequestedQty: req.RequestedQty,
		Notes:        req.Notes,
	}
	if err := s.repo.AddMRLine(ctx, line); err != nil {
		return nil, err
	}
	// Reserve stock immediately on draft-line create. Refining MRs (linked
	// to a MO consuming REFINING_INTAKE products) reserve from the
	// segregated pool via production_id; general MRs reserve from the
	// warehouse's default pool.
	if s.alloc != nil {
		if err := s.reserveForMRLine(ctx, tenantID, mr, line); err != nil {
			// Roll back the line we just wrote so the caller sees an atomic
			// "line + reservation" pair; the FE would otherwise show a line
			// that has no allocation and gets rejected at approve time.
			_ = s.repo.DeleteMRLine(ctx, tenantID, line.ID)
			return nil, err
		}
	}
	return line, nil
}

func (s *Service) UpdateMRLine(ctx context.Context, tenantID, mrID, lineID uint, req *UpdateMRLineRequest) (*MRLine, error) {
	mr, err := s.repo.GetMR(ctx, tenantID, mrID)
	if err != nil {
		return nil, fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusDraft {
		return nil, fmt.Errorf("cannot edit lines on a %s material request", mr.Status)
	}
	var line *MRLine
	for i := range mr.Lines {
		if mr.Lines[i].ID == lineID {
			line = &mr.Lines[i]
			break
		}
	}
	if line == nil {
		return nil, fmt.Errorf("line not found")
	}
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.UOMId = req.UOMId
	line.RequestedQty = req.RequestedQty
	line.Notes = req.Notes
	if err := s.repo.UpdateMRLine(ctx, line); err != nil {
		return nil, err
	}
	// Re-reserve on qty/product change. Reserve(OnUpdate) drops the prior
	// alloc row first so the ACTIVE sum doesn't count us twice.
	if s.alloc != nil {
		if err := s.reserveForMRLine(ctx, tenantID, mr, line); err != nil {
			return nil, err
		}
	}
	return line, nil
}

func (s *Service) DeleteMRLine(ctx context.Context, tenantID, mrID, lineID uint) error {
	mr, err := s.repo.GetMR(ctx, tenantID, mrID)
	if err != nil {
		return fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusDraft {
		return fmt.Errorf("cannot delete lines from a %s material request", mr.Status)
	}
	// Release the reservation for this line before the row disappears.
	if s.alloc != nil {
		_ = s.alloc.Release(ctx, tenantID, AllocSourceMRLine, lineID)
	}
	return s.repo.DeleteMRLine(ctx, tenantID, lineID)
}

// ── Goods Transfers ───────────────────────────────────────────────────────────

func (s *Service) ListTransfers(ctx context.Context, tenantID uint, status string, limit, offset int) ([]GoodsTransfer, error) {
	return s.repo.ListTransfers(ctx, tenantID, status, limit, offset)
}
func (s *Service) CountTransfers(ctx context.Context, tenantID uint, status string) (int64, error) {
	return s.repo.CountTransfers(ctx, tenantID, status)
}

func (s *Service) GetTransfer(ctx context.Context, tenantID, id uint) (*GoodsTransfer, error) {
	return s.repo.GetTransfer(ctx, tenantID, id)
}

func (s *Service) CreateTransfer(ctx context.Context, tenantID, userID uint, req *CreateTransferRequest) (*GoodsTransfer, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "GOODS_TRANSFER")
	if err != nil {
		return nil, fmt.Errorf("failed to generate goods transfer code: %w", err)
	}
	if req.FromWarehouseID == req.ToWarehouseID {
		return nil, fmt.Errorf("source and destination warehouses must be different")
	}
	d := req.TransferDate
	if d == "" {
		d = today()
	}
	t := &GoodsTransfer{
		TenantID:        tenantID,
		Code:            code,
		DocumentTypeID:  req.DocumentTypeID,
		MOID:            req.MOID,
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		TransferDate:    d,
		Status:          GTStatusDraft,
		Notes:           req.Notes,
		CreatedBy:       userID,
	}
	return t, s.repo.CreateTransfer(ctx, t)
}

func (s *Service) UpdateTransfer(ctx context.Context, tenantID, id uint, req *UpdateTransferRequest) (*GoodsTransfer, error) {
	t, err := s.repo.GetTransfer(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("goods transfer not found")
	}
	if t.Status != GTStatusDraft {
		return nil, fmt.Errorf("only DRAFT goods transfers can be edited")
	}
	if req.FromWarehouseID == req.ToWarehouseID {
		return nil, fmt.Errorf("source and destination warehouses must be different")
	}
	t.DocumentTypeID = req.DocumentTypeID
	t.MOID = req.MOID
	t.FromWarehouseID = req.FromWarehouseID
	t.ToWarehouseID = req.ToWarehouseID
	if req.TransferDate != "" {
		t.TransferDate = req.TransferDate
	}
	if req.Notes != "" {
		t.Notes = req.Notes
	}
	return t, s.repo.UpdateTransfer(ctx, t)
}

func (s *Service) SendTransfer(ctx context.Context, tenantID, id, userID uint) error {
	t, err := s.repo.GetTransfer(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("goods transfer not found")
	}
	if t.Status != GTStatusDraft {
		return fmt.Errorf("only DRAFT goods transfers can be sent")
	}
	if len(t.Lines) == 0 {
		return fmt.Errorf("goods transfer must have at least one line")
	}
	if err := s.repo.SendTransfer(ctx, tenantID, id, userID); err != nil {
		return err
	}
	if s.alloc != nil {
		for _, l := range t.Lines {
			_ = s.alloc.Consume(ctx, tenantID, AllocSourceGTLine, l.ID)
		}
	}
	return nil
}

func (s *Service) ReceiveTransfer(ctx context.Context, tenantID, id, userID uint) error {
	t, err := s.repo.GetTransfer(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("goods transfer not found")
	}
	if t.Status != GTStatusInTransit {
		return fmt.Errorf("only IN_TRANSIT goods transfers can be received")
	}
	return s.repo.ReceiveTransfer(ctx, tenantID, id, userID)
}

func (s *Service) ListTransferLines(ctx context.Context, tenantID, transferID uint) ([]GTLine, error) {
	if _, err := s.repo.GetTransfer(ctx, tenantID, transferID); err != nil {
		return nil, fmt.Errorf("goods transfer not found")
	}
	return s.repo.ListTransferLines(ctx, transferID)
}

func (s *Service) AddTransferLine(ctx context.Context, tenantID, transferID uint, req *AddTransferLineRequest) (*GTLine, error) {
	t, err := s.repo.GetTransfer(ctx, tenantID, transferID)
	if err != nil {
		return nil, fmt.Errorf("goods transfer not found")
	}
	if t.Status != GTStatusDraft {
		return nil, fmt.Errorf("cannot add lines to a %s goods transfer", t.Status)
	}
	line := &GTLine{
		TenantID:       tenantID,
		TransferID:     transferID,
		LineNumber:     len(t.Lines) + 1,
		ProductID:      req.ProductID,
		VariantID:      req.VariantID,
		FromLocationID: req.FromLocationID,
		ToLocationID:   req.ToLocationID,
		UOMId:          req.UOMId,
		Quantity:       req.Quantity,
		Notes:          req.Notes,
	}
	if err := s.repo.AddTransferLine(ctx, line); err != nil {
		return nil, err
	}
	if s.alloc != nil {
		if err := s.reserveForGTLine(ctx, tenantID, t, line); err != nil {
			_ = s.repo.DeleteTransferLine(ctx, tenantID, line.ID)
			return nil, err
		}
	}
	return line, nil
}

func (s *Service) UpdateTransferLine(ctx context.Context, tenantID, transferID, lineID uint, req *UpdateTransferLineRequest) (*GTLine, error) {
	t, err := s.repo.GetTransfer(ctx, tenantID, transferID)
	if err != nil {
		return nil, fmt.Errorf("goods transfer not found")
	}
	if t.Status != GTStatusDraft {
		return nil, fmt.Errorf("cannot edit lines on a %s goods transfer", t.Status)
	}
	var line *GTLine
	for i := range t.Lines {
		if t.Lines[i].ID == lineID {
			line = &t.Lines[i]
			break
		}
	}
	if line == nil {
		return nil, fmt.Errorf("line not found")
	}
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.FromLocationID = req.FromLocationID
	line.ToLocationID = req.ToLocationID
	line.UOMId = req.UOMId
	line.Quantity = req.Quantity
	line.Notes = req.Notes
	if err := s.repo.UpdateTransferLine(ctx, line); err != nil {
		return nil, err
	}
	if s.alloc != nil {
		if err := s.reserveForGTLine(ctx, tenantID, t, line); err != nil {
			return nil, err
		}
	}
	return line, nil
}

func (s *Service) DeleteTransferLine(ctx context.Context, tenantID, transferID, lineID uint) error {
	t, err := s.repo.GetTransfer(ctx, tenantID, transferID)
	if err != nil {
		return fmt.Errorf("goods transfer not found")
	}
	if t.Status != GTStatusDraft {
		return fmt.Errorf("cannot delete lines from a %s goods transfer", t.Status)
	}
	if s.alloc != nil {
		_ = s.alloc.Release(ctx, tenantID, AllocSourceGTLine, lineID)
	}
	return s.repo.DeleteTransferLine(ctx, tenantID, lineID)
}

// reserveForGTLine — the Reserve call reused by AddTransferLine / UpdateTransferLine.
func (s *Service) reserveForGTLine(ctx context.Context, tenantID uint, gt *GoodsTransfer, line *GTLine) error {
	_, err := s.alloc.Reserve(ctx, ReserveRequest{
		TenantID:    tenantID,
		ScopeKey:    s.scopeForGT(ctx, gt, line),
		Quantity:    line.Quantity,
		SourceType:  AllocSourceGTLine,
		SourceID:    line.ID,
		SourceDocID: gt.ID,
		Notes:       fmt.Sprintf("GT %s line %d", gt.Code, line.LineNumber),
		OnUpdate:    true,
	})
	return err
}

// ── Goods Issues ──────────────────────────────────────────────────────────────

func (s *Service) ListIssues(ctx context.Context, tenantID uint, status, reason string, limit, offset int) ([]GoodsIssue, error) {
	return s.repo.ListIssues(ctx, tenantID, status, reason, limit, offset)
}
func (s *Service) CountIssues(ctx context.Context, tenantID uint, status, reason string) (int64, error) {
	return s.repo.CountIssues(ctx, tenantID, status, reason)
}

func (s *Service) GetIssue(ctx context.Context, tenantID, id uint) (*GoodsIssue, error) {
	return s.repo.GetIssue(ctx, tenantID, id)
}

func (s *Service) CreateIssue(ctx context.Context, tenantID, userID uint, req *CreateIssueRequest) (*GoodsIssue, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "GOODS_ISSUE")
	if err != nil {
		return nil, fmt.Errorf("failed to generate goods issue code: %w", err)
	}
	d := req.IssueDate
	if d == "" {
		d = today()
	}
	gi := &GoodsIssue{
		TenantID:       tenantID,
		Code:           code,
		IssueDate:      d,
		WarehouseID:    req.WarehouseID,
		DocumentTypeID: req.DocumentTypeID,
		MRID:           req.MRID,
		ReferenceType:  req.ReferenceType,
		ReferenceID:    req.ReferenceID,
		Status:         GIStatusDraft,
		Notes:          req.Notes,
		CreatedBy:      userID,
	}
	return gi, s.repo.CreateIssue(ctx, gi)
}

func (s *Service) UpdateIssue(ctx context.Context, tenantID, id uint, req *UpdateIssueRequest) (*GoodsIssue, error) {
	gi, err := s.repo.GetIssue(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("goods issue not found")
	}
	if gi.Status != GIStatusDraft {
		return nil, fmt.Errorf("only DRAFT goods issues can be edited")
	}
	if req.IssueDate != "" {
		gi.IssueDate = req.IssueDate
	}
	gi.WarehouseID = req.WarehouseID
	gi.DocumentTypeID = req.DocumentTypeID
	gi.MRID = req.MRID
	gi.ReferenceType = req.ReferenceType
	gi.ReferenceID = req.ReferenceID
	if req.Notes != "" {
		gi.Notes = req.Notes
	}
	return gi, s.repo.UpdateIssue(ctx, gi)
}

func (s *Service) CancelIssue(ctx context.Context, tenantID, id uint) error {
	row, err := s.repo.GetIssue(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("goods issue not found")
	}
	if row.Status != GIStatusDraft {
		return fmt.Errorf("only DRAFT can be cancelled")
	}
	if s.alloc != nil {
		_ = s.alloc.ReleaseByDoc(ctx, tenantID, AllocSourceGILine, id)
	}
	return s.repo.SetIssueStatus(ctx, tenantID, id, GIStatusCancelled)
}

func (s *Service) ConfirmIssue(ctx context.Context, tenantID, id, userID uint) error {
	gi, err := s.repo.GetIssue(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("goods issue not found")
	}
	if gi.Status != GIStatusDraft {
		return fmt.Errorf("only DRAFT goods issues can be confirmed")
	}
	if len(gi.Lines) == 0 {
		return fmt.Errorf("goods issue must have at least one line")
	}
	// Cap check: for MR-linked and production-order-linked GIs, ensure this
	// confirm won't push any MR line's cumulative issued_qty past its
	// requested_qty. Fires before the actual confirm so the transaction
	// aborts cleanly on breach.
	if err := s.validateGIAgainstMRCap(ctx, tenantID, gi); err != nil {
		return err
	}
	if err := s.repo.ConfirmIssue(ctx, tenantID, id, userID); err != nil {
		return err
	}
	// Consume the allocations for this GI. If MR-linked, each MR line's
	// allocation flips CONSUMED; if standalone, each GI line's own
	// allocation flips CONSUMED.
	if s.alloc != nil {
		if gi.MRID != nil {
			// Match GI product back to MR line to consume the right alloc.
			mrLines, _ := s.repo.ListMRLines(ctx, *gi.MRID)
			for _, ml := range mrLines {
				for _, l := range gi.Lines {
					if l.ProductID == ml.ProductID {
						_ = s.alloc.Consume(ctx, tenantID, AllocSourceMRLine, ml.ID)
					}
				}
			}
		} else {
			for _, l := range gi.Lines {
				_ = s.alloc.Consume(ctx, tenantID, AllocSourceGILine, l.ID)
			}
		}
	}
	// If this GI is a production issue (reference_type=PRODUCTION_ORDER),
	// walk all MRs linked to the same MO and bump each MR line's
	// issued_qty by the matching product's issued quantity.
	// Best-effort — a failure here doesn't unwind the GI confirmation.
	if gi.ReferenceType == "PRODUCTION_ORDER" && gi.ReferenceID != nil {
		s.bumpMRIssuedQtyForMO(ctx, tenantID, *gi.ReferenceID, gi.Lines)
	}
	return nil
}

// validateGIAgainstMRCap — pre-flight check that a GI confirm won't push any
// linked MR line's cumulative issued_qty past its requested_qty. Handles
// both single-MR GIs (gi.mr_id set) and production-order GIs (may touch
// multiple MRs by product across the same MO).
func (s *Service) validateGIAgainstMRCap(ctx context.Context, tenantID uint, gi *GoodsIssue) error {
	// Collect per-product delta from this GI.
	perProduct := make(map[uint]float64, len(gi.Lines))
	for _, l := range gi.Lines {
		perProduct[l.ProductID] += l.Quantity
	}

	// The set of MRs this GI's confirm will touch.
	var mrsToCheck []MaterialRequest
	if gi.MRID != nil {
		mr, err := s.repo.GetMR(ctx, tenantID, *gi.MRID)
		if err == nil {
			mrsToCheck = []MaterialRequest{*mr}
		}
	}
	if gi.ReferenceType == "PRODUCTION_ORDER" && gi.ReferenceID != nil {
		moMRs, _ := s.repo.ListMRsByMO(ctx, tenantID, *gi.ReferenceID)
		mrsToCheck = append(mrsToCheck, moMRs...)
	}
	if len(mrsToCheck) == 0 {
		return nil // standalone GI — no MR to cap against
	}

	// The production-order path can span MULTIPLE MRs; a single product's
	// remaining capacity is the sum of remaining across those MRs. Aggregate
	// remaining per product before comparing.
	perProductRemaining := make(map[uint]float64, len(perProduct))
	for _, mr := range mrsToCheck {
		for _, ln := range mr.Lines {
			perProductRemaining[ln.ProductID] += ln.RequestedQty - ln.IssuedQty
		}
	}
	for productID, delta := range perProduct {
		remaining, ok := perProductRemaining[productID]
		if !ok {
			continue // GI line for a product the MR didn't request — allow (matches legacy behavior)
		}
		if delta > remaining+0.005 {
			return fmt.Errorf(
				"GI line for product %d exceeds MR requested remaining: %.4f > %.4f",
				productID, delta, remaining,
			)
		}
	}
	return nil
}

// bumpMRIssuedQtyForMO adds each GI line's quantity onto the matching MR line
// (by product_id) across every MR that references the same MO.
func (s *Service) bumpMRIssuedQtyForMO(ctx context.Context, tenantID, moID uint, giLines []GILine) {
	mrs, err := s.repo.ListMRsByMO(ctx, tenantID, moID)
	if err != nil || len(mrs) == 0 {
		return
	}
	// Sum GI qty per product so a GI issuing the same product twice counts once
	perProduct := make(map[uint]float64, len(giLines))
	for _, l := range giLines {
		perProduct[l.ProductID] += l.Quantity
	}
	for i := range mrs {
		mr := &mrs[i]
		for j := range mr.Lines {
			line := &mr.Lines[j]
			if delta, ok := perProduct[line.ProductID]; ok && delta > 0 {
				line.IssuedQty += delta
				_ = s.repo.UpdateMRLine(ctx, line)
			}
		}
	}
}

func (s *Service) ListIssueLines(ctx context.Context, tenantID, issueID uint) ([]GILine, error) {
	if _, err := s.repo.GetIssue(ctx, tenantID, issueID); err != nil {
		return nil, fmt.Errorf("goods issue not found")
	}
	return s.repo.ListIssueLines(ctx, issueID)
}

func (s *Service) AddIssueLine(ctx context.Context, tenantID, issueID uint, req *AddIssueLineRequest) (*GILine, error) {
	gi, err := s.repo.GetIssue(ctx, tenantID, issueID)
	if err != nil {
		return nil, fmt.Errorf("goods issue not found")
	}
	if gi.Status != GIStatusDraft {
		return nil, fmt.Errorf("cannot add lines to a %s goods issue", gi.Status)
	}
	line := &GILine{
		TenantID:   tenantID,
		IssueID:    issueID,
		LineNumber: len(gi.Lines) + 1,
		ProductID:  req.ProductID,
		VariantID:  req.VariantID,
		LocationID: req.LocationID,
		UOMId:      req.UOMId,
		Quantity:   req.Quantity,
		UnitCost:   req.UnitCost,
		TotalCost:  req.Quantity * req.UnitCost,
		Notes:      req.Notes,
	}
	if err := s.repo.AddIssueLine(ctx, line); err != nil {
		return nil, err
	}
	// Standalone GIs need their own allocation. GIs linked to an MR already
	// have stock reserved by the MR, so we skip — the confirm flow will
	// Consume the MR_LINE alloc instead.
	if s.alloc != nil && gi.MRID == nil {
		if err := s.reserveForGILine(ctx, tenantID, gi, line); err != nil {
			_ = s.repo.DeleteIssueLine(ctx, tenantID, line.ID)
			return nil, err
		}
	}
	return line, nil
}

func (s *Service) UpdateIssueLine(ctx context.Context, tenantID, issueID, lineID uint, req *UpdateIssueLineRequest) (*GILine, error) {
	gi, err := s.repo.GetIssue(ctx, tenantID, issueID)
	if err != nil {
		return nil, fmt.Errorf("goods issue not found")
	}
	if gi.Status != GIStatusDraft {
		return nil, fmt.Errorf("cannot edit lines on a %s goods issue", gi.Status)
	}
	var line *GILine
	for i := range gi.Lines {
		if gi.Lines[i].ID == lineID {
			line = &gi.Lines[i]
			break
		}
	}
	if line == nil {
		return nil, fmt.Errorf("line not found")
	}
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.LocationID = req.LocationID
	line.UOMId = req.UOMId
	line.Quantity = req.Quantity
	line.UnitCost = req.UnitCost
	line.TotalCost = req.Quantity * req.UnitCost
	line.Notes = req.Notes
	if err := s.repo.UpdateIssueLine(ctx, line); err != nil {
		return nil, err
	}
	if s.alloc != nil && gi.MRID == nil {
		if err := s.reserveForGILine(ctx, tenantID, gi, line); err != nil {
			return nil, err
		}
	}
	return line, nil
}

func (s *Service) DeleteIssueLine(ctx context.Context, tenantID, issueID, lineID uint) error {
	gi, err := s.repo.GetIssue(ctx, tenantID, issueID)
	if err != nil {
		return fmt.Errorf("goods issue not found")
	}
	if gi.Status != GIStatusDraft {
		return fmt.Errorf("cannot delete lines from a %s goods issue", gi.Status)
	}
	if s.alloc != nil {
		_ = s.alloc.Release(ctx, tenantID, AllocSourceGILine, lineID)
	}
	return s.repo.DeleteIssueLine(ctx, tenantID, lineID)
}

// reserveForGILine — Reserve for standalone GI lines. MR-linked GIs use the
// MR's allocation instead so we don't double-count.
func (s *Service) reserveForGILine(ctx context.Context, tenantID uint, gi *GoodsIssue, line *GILine) error {
	var productionID *uint
	if gi.ReferenceType == "PRODUCTION_ORDER" && gi.ReferenceID != nil {
		productionID = gi.ReferenceID
	}
	_, err := s.alloc.Reserve(ctx, ReserveRequest{
		TenantID:    tenantID,
		ScopeKey:    s.scopeForGI(ctx, gi, line, productionID),
		Quantity:    line.Quantity,
		SourceType:  AllocSourceGILine,
		SourceID:    line.ID,
		SourceDocID: gi.ID,
		Notes:       fmt.Sprintf("GI %s line %d", gi.Code, line.LineNumber),
		OnUpdate:    true,
	})
	return err
}

// ── Stock Adjustments ─────────────────────────────────────────────────────────

func (s *Service) ListAdjustments(ctx context.Context, tenantID uint, status string, limit, offset int) ([]StockAdjustment, error) {
	return s.repo.ListAdjustments(ctx, tenantID, status, limit, offset)
}
func (s *Service) CountAdjustments(ctx context.Context, tenantID uint, status string) (int64, error) {
	return s.repo.CountAdjustments(ctx, tenantID, status)
}

func (s *Service) GetAdjustment(ctx context.Context, tenantID, id uint) (*StockAdjustment, error) {
	return s.repo.GetAdjustment(ctx, tenantID, id)
}

func (s *Service) CreateAdjustment(ctx context.Context, tenantID, userID uint, req *CreateAdjustmentRequest) (*StockAdjustment, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "STOCK_ADJUSTMENT")
	if err != nil {
		return nil, fmt.Errorf("failed to generate stock adjustment code: %w", err)
	}
	d := req.AdjustmentDate
	if d == "" {
		d = today()
	}
	sa := &StockAdjustment{
		TenantID:       tenantID,
		Code:           code,
		AdjustmentDate: d,
		WarehouseID:    req.WarehouseID,
		DocumentTypeID: req.DocumentTypeID,
		Status:         SAStatusDraft,
		Notes:          req.Notes,
		CreatedBy:      userID,
	}
	return sa, s.repo.CreateAdjustment(ctx, sa)
}

func (s *Service) UpdateAdjustment(ctx context.Context, tenantID, id uint, req *UpdateAdjustmentRequest) (*StockAdjustment, error) {
	sa, err := s.repo.GetAdjustment(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("stock adjustment not found")
	}
	if sa.Status != SAStatusDraft {
		return nil, fmt.Errorf("only DRAFT stock adjustments can be edited")
	}
	if req.AdjustmentDate != "" {
		sa.AdjustmentDate = req.AdjustmentDate
	}
	sa.WarehouseID = req.WarehouseID
	sa.DocumentTypeID = req.DocumentTypeID
	if req.Notes != "" {
		sa.Notes = req.Notes
	}
	return sa, s.repo.UpdateAdjustment(ctx, sa)
}

func (s *Service) CancelAdjustment(ctx context.Context, tenantID, id uint) error {
	row, err := s.repo.GetAdjustment(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("stock adjustment not found")
	}
	if row.Status != SAStatusDraft {
		return fmt.Errorf("only DRAFT can be cancelled")
	}
	return s.repo.SetAdjustmentStatus(ctx, tenantID, id, SAStatusCancelled)
}

func (s *Service) ConfirmAdjustment(ctx context.Context, tenantID, id, userID uint) error {
	sa, err := s.repo.GetAdjustment(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("stock adjustment not found")
	}
	if sa.Status != SAStatusDraft {
		return fmt.Errorf("only DRAFT stock adjustments can be confirmed")
	}
	if len(sa.Lines) == 0 {
		return fmt.Errorf("stock adjustment must have at least one line")
	}
	return s.repo.ConfirmAdjustment(ctx, tenantID, id, userID)
}

func (s *Service) ListAdjustmentLines(ctx context.Context, tenantID, adjID uint) ([]SALine, error) {
	if _, err := s.repo.GetAdjustment(ctx, tenantID, adjID); err != nil {
		return nil, fmt.Errorf("stock adjustment not found")
	}
	return s.repo.ListAdjustmentLines(ctx, adjID)
}

func (s *Service) AddAdjustmentLine(ctx context.Context, tenantID, adjID uint, req *AddAdjustmentLineRequest) (*SALine, error) {
	sa, err := s.repo.GetAdjustment(ctx, tenantID, adjID)
	if err != nil {
		return nil, fmt.Errorf("stock adjustment not found")
	}
	if sa.Status != SAStatusDraft {
		return nil, fmt.Errorf("cannot add lines to a %s stock adjustment", sa.Status)
	}
	line := &SALine{
		TenantID:     tenantID,
		AdjustmentID: adjID,
		LineNumber:   len(sa.Lines) + 1,
		ProductID:    req.ProductID,
		VariantID:    req.VariantID,
		LocationID:   req.LocationID,
		UOMId:        req.UOMId,
		Quantity:     req.Quantity,
		UnitCost:     req.UnitCost,
		Notes:        req.Notes,
	}
	return line, s.repo.AddAdjustmentLine(ctx, line)
}

func (s *Service) UpdateAdjustmentLine(ctx context.Context, tenantID, adjID, lineID uint, req *UpdateAdjustmentLineRequest) (*SALine, error) {
	sa, err := s.repo.GetAdjustment(ctx, tenantID, adjID)
	if err != nil {
		return nil, fmt.Errorf("stock adjustment not found")
	}
	if sa.Status != SAStatusDraft {
		return nil, fmt.Errorf("cannot edit lines on a %s stock adjustment", sa.Status)
	}
	var line *SALine
	for i := range sa.Lines {
		if sa.Lines[i].ID == lineID {
			line = &sa.Lines[i]
			break
		}
	}
	if line == nil {
		return nil, fmt.Errorf("line not found")
	}
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.LocationID = req.LocationID
	line.UOMId = req.UOMId
	line.Quantity = req.Quantity
	line.UnitCost = req.UnitCost
	line.Notes = req.Notes
	return line, s.repo.UpdateAdjustmentLine(ctx, line)
}

func (s *Service) DeleteAdjustmentLine(ctx context.Context, tenantID, adjID, lineID uint) error {
	sa, err := s.repo.GetAdjustment(ctx, tenantID, adjID)
	if err != nil {
		return fmt.Errorf("stock adjustment not found")
	}
	if sa.Status != SAStatusDraft {
		return fmt.Errorf("cannot delete lines from a %s stock adjustment", sa.Status)
	}
	return s.repo.DeleteAdjustmentLine(ctx, tenantID, lineID)
}

// ── Quality Checks ────────────────────────────────────────────────────────────

func (s *Service) ListQualityChecks(ctx context.Context, tenantID uint, status string, limit, offset int) ([]QualityCheck, error) {
	return s.repo.ListQualityChecks(ctx, tenantID, status, limit, offset)
}
func (s *Service) CountQualityChecks(ctx context.Context, tenantID uint, status string) (int64, error) {
	return s.repo.CountQualityChecks(ctx, tenantID, status)
}

func (s *Service) GetQualityCheck(ctx context.Context, tenantID, id uint) (*QualityCheck, error) {
	return s.repo.GetQualityCheck(ctx, tenantID, id)
}

func (s *Service) CreateQualityCheck(ctx context.Context, tenantID, userID uint, req *CreateQualityCheckRequest) (*QualityCheck, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "QUALITY_CHECK")
	if err != nil {
		return nil, fmt.Errorf("failed to generate quality check code: %w", err)
	}
	d := req.CheckDate
	if d == "" {
		d = today()
	}
	// Auto-select the MATERIAL_QC system Type if the caller didn't pick one.
	docTypeID := req.DocumentTypeID
	if docTypeID == nil && s.dt != nil {
		if id, err := s.dt.FindSystemType(ctx, tenantID, "QC", QCTypeMaterial); err == nil && id != 0 {
			docTypeID = &id
		}
	}
	qc := &QualityCheck{
		TenantID:       tenantID,
		Code:           code,
		DocumentTypeID: docTypeID,
		ReferenceType:  req.ReferenceType,
		ReferenceID:    req.ReferenceID,
		WarehouseID:    req.WarehouseID,
		CheckDate:      d,
		Status:         QCStatusPending,
		InspectorID:    req.InspectorID,
		Notes:          req.Notes,
		CreatedBy:      userID,
	}
	return qc, s.repo.CreateQualityCheck(ctx, qc)
}

func (s *Service) StartQualityCheck(ctx context.Context, tenantID, id, userID uint) (*QualityCheck, error) {
	qc, err := s.repo.GetQualityCheck(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("quality check not found")
	}
	if qc.Status != QCStatusPending {
		return nil, fmt.Errorf("only PENDING quality checks can be started")
	}
	if err := s.repo.UpdateQualityCheckStatus(ctx, tenantID, id, QCStatusInProgress); err != nil {
		return nil, err
	}
	qc.Status = QCStatusInProgress
	return qc, nil
}

func (s *Service) SubmitQualityCheck(ctx context.Context, tenantID, id, userID uint) error {
	qc, err := s.repo.GetQualityCheck(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("quality check not found")
	}
	if qc.Status == QCStatusPassed || qc.Status == QCStatusFailed || qc.Status == QCStatusPartial {
		return fmt.Errorf("quality check is already finalized with status %s", qc.Status)
	}
	// Resolve the DAMAGED bin only if at least one line has failed qty —
	// keeps the "all passed" fast-path free of masterdata lookups.
	var damagedBinID uint
	for _, l := range qc.Lines {
		if l.QtyFailed > 0 {
			bin, err := s.DamagedBinFor(ctx, tenantID, qc.WarehouseID)
			if err != nil {
				return fmt.Errorf("could not resolve DAMAGED bin for warehouse %d: %w", qc.WarehouseID, err)
			}
			damagedBinID = bin
			break
		}
	}
	return s.repo.SubmitQualityCheck(ctx, tenantID, id, userID, damagedBinID)
}

func (s *Service) ListQCLines(ctx context.Context, tenantID, checkID uint) ([]QCLine, error) {
	if _, err := s.repo.GetQualityCheck(ctx, tenantID, checkID); err != nil {
		return nil, fmt.Errorf("quality check not found")
	}
	return s.repo.ListQCLines(ctx, checkID)
}

func (s *Service) AddQCLine(ctx context.Context, tenantID, checkID uint, req *AddQCLineRequest) (*QCLine, error) {
	qc, err := s.repo.GetQualityCheck(ctx, tenantID, checkID)
	if err != nil {
		return nil, fmt.Errorf("quality check not found")
	}
	if qc.Status == QCStatusPassed || qc.Status == QCStatusFailed || qc.Status == QCStatusPartial {
		return nil, fmt.Errorf("cannot add lines to a finalized quality check")
	}
	result := req.Result
	if result == "" {
		result = QCStatusPending
	}
	line := &QCLine{
		TenantID:        tenantID,
		CheckID:         checkID,
		LineNumber:      len(qc.Lines) + 1,
		ProductID:       req.ProductID,
		VariantID:       req.VariantID,
		QtyChecked:      req.QtyChecked,
		QtyPassed:       req.QtyPassed,
		QtyFailed:       req.QtyFailed,
		Result:          result,
		RejectionReason: req.RejectionReason,
		Notes:           req.Notes,
	}
	return line, s.repo.AddQCLine(ctx, line)
}

func (s *Service) UpdateQCLine(ctx context.Context, tenantID, checkID, lineID uint, req *UpdateQCLineRequest) (*QCLine, error) {
	qc, err := s.repo.GetQualityCheck(ctx, tenantID, checkID)
	if err != nil {
		return nil, fmt.Errorf("quality check not found")
	}
	if qc.Status == QCStatusPassed || qc.Status == QCStatusFailed || qc.Status == QCStatusPartial {
		return nil, fmt.Errorf("cannot edit lines on a finalized quality check")
	}
	var line *QCLine
	for i := range qc.Lines {
		if qc.Lines[i].ID == lineID {
			line = &qc.Lines[i]
			break
		}
	}
	if line == nil {
		return nil, fmt.Errorf("line not found")
	}
	result := req.Result
	if result == "" {
		result = QCStatusPending
	}
	line.ProductID = req.ProductID
	line.VariantID = req.VariantID
	line.QtyChecked = req.QtyChecked
	line.QtyPassed = req.QtyPassed
	line.QtyFailed = req.QtyFailed
	line.Result = result
	line.RejectionReason = req.RejectionReason
	line.Notes = req.Notes
	return line, s.repo.UpdateQCLine(ctx, line)
}

// ── Auto-QC creation (called by other modules) ────────────────────────────────

// QCAutoLine is a minimal product+qty pair used by other modules to create QC lines
// when they trigger an auto-QC (e.g. procurement on GRN confirm, manufacturing on output).
type QCAutoLine struct {
	ProductID uint
	VariantID *uint
	Quantity  float64
}

// CreateAutoQC creates a PENDING QualityCheck with the supplied lines.
// Used by other modules to trigger QC from cross-module events (GRN confirm, production output).
// Failures are returned but callers typically treat them as non-fatal.
func (s *Service) CreateAutoQC(
	ctx context.Context,
	tenantID, userID uint,
	qcType, refType string,
	refID, warehouseID uint,
	notes string,
	lines []QCAutoLine,
) (*QualityCheck, error) {
	if qcType == "" {
		qcType = QCTypeMaterial
	}
	code, err := s.repo.NextCode(ctx, tenantID, "QUALITY_CHECK")
	if err != nil {
		return nil, fmt.Errorf("failed to generate QC code: %w", err)
	}
	// Look up the seeded QC Type row for this qcType key so the FK is set.
	var docTypeID *uint
	if s.dt != nil {
		if id, err := s.dt.FindSystemType(ctx, tenantID, "QC", qcType); err == nil && id != 0 {
			docTypeID = &id
		}
	}
	qc := &QualityCheck{
		TenantID:       tenantID,
		Code:           code,
		DocumentTypeID: docTypeID,
		ReferenceType:  refType,
		ReferenceID:    &refID,
		WarehouseID:    warehouseID,
		CheckDate:      today(),
		Status:         QCStatusPending,
		Notes:          notes,
		CreatedBy:      userID,
	}
	if err := s.repo.CreateQualityCheck(ctx, qc); err != nil {
		return nil, err
	}
	for i, l := range lines {
		line := &QCLine{
			TenantID:   tenantID,
			CheckID:    qc.ID,
			LineNumber: i + 1,
			ProductID:  l.ProductID,
			VariantID:  l.VariantID,
			QtyChecked: l.Quantity,
			Result:     QCStatusPending,
		}
		if err := s.repo.AddQCLine(ctx, line); err != nil {
			return qc, fmt.Errorf("QC %d created but line %d failed: %w", qc.ID, i+1, err)
		}
	}
	return qc, nil
}

// ── Stock Balances ────────────────────────────────────────────────────────────

func (s *Service) GetStockBalance(ctx context.Context, tenantID uint, warehouseID *uint, productID *uint) ([]StockBalanceRow, error) {
	return s.repo.GetStockBalance(ctx, tenantID, warehouseID, productID)
}
