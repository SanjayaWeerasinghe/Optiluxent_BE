package inventory

import (
	"context"
	"fmt"
	"time"
)

// Service implements all inventory business logic, wrapping the Repository.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func today() string { return time.Now().Format("2006-01-02") }

// ── Material Requests ─────────────────────────────────────────────────────────

func (s *Service) ListMRs(ctx context.Context, tenantID uint, status string) ([]MaterialRequest, error) {
	return s.repo.ListMRs(ctx, tenantID, status)
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
		TenantID:     tenantID,
		Code:         code,
		RequestedBy:  req.RequestedBy,
		DepartmentID: req.DepartmentID,
		WarehouseID:  req.WarehouseID,
		NeededDate:   d,
		Status:       MRStatusDraft,
		Notes:        req.Notes,
		CreatedBy:    userID,
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

func (s *Service) ApproveMR(ctx context.Context, tenantID, id, userID uint) error {
	mr, err := s.repo.GetMR(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusDraft {
		return fmt.Errorf("can only approve DRAFT material requests")
	}
	return s.repo.ApproveMR(ctx, tenantID, id, userID)
}

func (s *Service) RejectMR(ctx context.Context, tenantID, id, userID uint, reason string) error {
	mr, err := s.repo.GetMR(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusDraft {
		return fmt.Errorf("can only reject DRAFT material requests")
	}
	return s.repo.RejectMR(ctx, tenantID, id, userID, reason)
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
	return line, s.repo.AddMRLine(ctx, line)
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
	return line, s.repo.UpdateMRLine(ctx, line)
}

func (s *Service) DeleteMRLine(ctx context.Context, tenantID, mrID, lineID uint) error {
	mr, err := s.repo.GetMR(ctx, tenantID, mrID)
	if err != nil {
		return fmt.Errorf("material request not found")
	}
	if mr.Status != MRStatusDraft {
		return fmt.Errorf("cannot delete lines from a %s material request", mr.Status)
	}
	return s.repo.DeleteMRLine(ctx, tenantID, lineID)
}

// ── Goods Transfers ───────────────────────────────────────────────────────────

func (s *Service) ListTransfers(ctx context.Context, tenantID uint, status string) ([]GoodsTransfer, error) {
	return s.repo.ListTransfers(ctx, tenantID, status)
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
	return s.repo.SendTransfer(ctx, tenantID, id, userID)
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
	return line, s.repo.AddTransferLine(ctx, line)
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
	return line, s.repo.UpdateTransferLine(ctx, line)
}

func (s *Service) DeleteTransferLine(ctx context.Context, tenantID, transferID, lineID uint) error {
	t, err := s.repo.GetTransfer(ctx, tenantID, transferID)
	if err != nil {
		return fmt.Errorf("goods transfer not found")
	}
	if t.Status != GTStatusDraft {
		return fmt.Errorf("cannot delete lines from a %s goods transfer", t.Status)
	}
	return s.repo.DeleteTransferLine(ctx, tenantID, lineID)
}

// ── Goods Issues ──────────────────────────────────────────────────────────────

func (s *Service) ListIssues(ctx context.Context, tenantID uint, status, reason string) ([]GoodsIssue, error) {
	return s.repo.ListIssues(ctx, tenantID, status, reason)
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
		TenantID:      tenantID,
		Code:          code,
		IssueDate:     d,
		WarehouseID:   req.WarehouseID,
		IssueReason:   req.IssueReason,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		Status:        GIStatusDraft,
		Notes:         req.Notes,
		CreatedBy:     userID,
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
	gi.IssueReason = req.IssueReason
	gi.ReferenceType = req.ReferenceType
	gi.ReferenceID = req.ReferenceID
	if req.Notes != "" {
		gi.Notes = req.Notes
	}
	return gi, s.repo.UpdateIssue(ctx, gi)
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
	return s.repo.ConfirmIssue(ctx, tenantID, id, userID)
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
	return line, s.repo.AddIssueLine(ctx, line)
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
	return line, s.repo.UpdateIssueLine(ctx, line)
}

func (s *Service) DeleteIssueLine(ctx context.Context, tenantID, issueID, lineID uint) error {
	gi, err := s.repo.GetIssue(ctx, tenantID, issueID)
	if err != nil {
		return fmt.Errorf("goods issue not found")
	}
	if gi.Status != GIStatusDraft {
		return fmt.Errorf("cannot delete lines from a %s goods issue", gi.Status)
	}
	return s.repo.DeleteIssueLine(ctx, tenantID, lineID)
}

// ── Stock Adjustments ─────────────────────────────────────────────────────────

func (s *Service) ListAdjustments(ctx context.Context, tenantID uint, status string) ([]StockAdjustment, error) {
	return s.repo.ListAdjustments(ctx, tenantID, status)
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
		AdjustReason:   req.AdjustReason,
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
	sa.AdjustReason = req.AdjustReason
	if req.Notes != "" {
		sa.Notes = req.Notes
	}
	return sa, s.repo.UpdateAdjustment(ctx, sa)
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

func (s *Service) ListQualityChecks(ctx context.Context, tenantID uint, status string) ([]QualityCheck, error) {
	return s.repo.ListQualityChecks(ctx, tenantID, status)
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
	qcType := req.QCType
	if qcType == "" {
		qcType = QCTypeMaterial
	}
	qc := &QualityCheck{
		TenantID:      tenantID,
		Code:          code,
		QCType:        qcType,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		WarehouseID:   req.WarehouseID,
		CheckDate:     d,
		Status:        QCStatusPending,
		InspectorID:   req.InspectorID,
		Notes:         req.Notes,
		CreatedBy:     userID,
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
	return s.repo.SubmitQualityCheck(ctx, tenantID, id, userID)
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
	qc := &QualityCheck{
		TenantID:      tenantID,
		Code:          code,
		QCType:        qcType,
		ReferenceType: refType,
		ReferenceID:   &refID,
		WarehouseID:   warehouseID,
		CheckDate:     today(),
		Status:        QCStatusPending,
		Notes:         notes,
		CreatedBy:     userID,
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
