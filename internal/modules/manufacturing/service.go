package manufacturing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"erp-system/internal/infrastructure/events"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QCAutoCreator is the minimal interface manufacturing needs from the inventory
// module to spin up a Product QC whenever a production output is recorded.
// inventory.Service satisfies this via its CreateAutoQC method (wrapped at injection).
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

// Service contains all manufacturing business logic.
type Service struct {
	repo        Repository
	db          *gorm.DB // used by the dashboard aggregator + uom converter
	bus         events.EventBus
	qc          QCAutoCreator
	dashReaders DashboardReaders
}

func NewService(repo Repository, db *gorm.DB) *Service { return &Service{repo: repo, db: db} }
func (s *Service) SetEventBus(bus events.EventBus)    { s.bus = bus }
func (s *Service) SetQCAutoCreator(qc QCAutoCreator)  { s.qc = qc }

// BumpProducedQty adds the given quantity to a Manufacturing Order's produced_qty,
// converting the caller's UOM to the MO's UOM via product_uom_conversions when needed.
// Called by procurement.ConfirmGRN when a PRODUCTION_OUTPUT GRN is confirmed.
func (s *Service) BumpProducedQty(ctx context.Context, tenantID, moID, productID, uomID uint, qty float64) error {
	order, err := s.repo.GetOrder(ctx, tenantID, moID)
	if err != nil {
		return fmt.Errorf("production order not found: %w", err)
	}
	// If the caller UOM matches the MO UOM, no conversion needed. Otherwise
	// look up the ratio; fall back to raw qty if none configured so we still
	// see something in the total.
	delta := qty
	if uomID != order.UOMID {
		if converted, ok := ConvertProductQty(ctx, s.db, productID, uomID, order.UOMID, qty); ok {
			delta = converted
		}
	}
	return s.repo.IncrementProducedQty(ctx, tenantID, moID, delta)
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
	_ = s.bus.Publish(ctx, StreamManufacturing, ev)
}

// recalcEstimateTotals recomputes the header cost buckets from the line items.
func recalcEstimateTotals(est *CostEstimate) {
	var mat, res, ovh float64
	for _, l := range est.Lines {
		switch l.LineType {
		case EstimateLineTypeMaterial:
			mat += l.TotalCost
		case EstimateLineTypeResource, EstimateLineTypeLabor:
			res += l.TotalCost
		case EstimateLineTypeOverhead:
			ovh += l.TotalCost
		}
	}
	est.EstimatedMaterialCost = mat
	est.EstimatedResourceCost = res
	est.EstimatedOverheadCost = ovh
	est.TotalEstimatedCost = mat + res + ovh
}

// computeActuals derives actual costs from an order's resources.
func computeActuals(resources []ProductionResource) (mat, res, ovh float64) {
	for _, r := range resources {
		switch r.ResourceType {
		case ResourceTypeMaterial:
			mat += r.TotalCost
		case ResourceTypeLabor:
			res += r.TotalCost
		case ResourceTypeOverhead:
			ovh += r.TotalCost
		}
	}
	return
}

// ── Cost Estimates ────────────────────────────────────────────────────────────

func (s *Service) ListEstimates(ctx context.Context, tenantID uint, status string) ([]CostEstimate, error) {
	return s.repo.ListEstimates(ctx, tenantID, status)
}

func (s *Service) GetEstimate(ctx context.Context, tenantID, id uint) (*CostEstimate, error) {
	return s.repo.GetEstimate(ctx, tenantID, id)
}

func (s *Service) CreateEstimate(ctx context.Context, tenantID, userID uint, req *CreateEstimateRequest) (*CostEstimate, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "COST_ESTIMATE")
	if err != nil {
		return nil, fmt.Errorf("failed to generate cost estimate code: %w", err)
	}
	est := &CostEstimate{
		TenantID:         tenantID,
		Code:             code,
		ProductID:        req.ProductID,
		UOMID:            req.UOMID,
		PlannedQty:       req.PlannedQty,
		BOMID:            req.BOMID,
		RoutingID:        req.RoutingID,
		PlannedStartDate: req.PlannedStartDate,
		PlannedEndDate:   req.PlannedEndDate,
		Notes:            req.Notes,
		Status:           EstimateStatusDraft,
		CreatedBy:        &userID,
	}
	if err := s.repo.CreateEstimate(ctx, est); err != nil {
		return nil, err
	}
	return est, nil
}

func (s *Service) UpdateEstimate(ctx context.Context, tenantID, id uint, req *UpdateEstimateRequest) (*CostEstimate, error) {
	est, err := s.repo.GetEstimate(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("cost estimate not found")
	}
	if est.Status != EstimateStatusDraft {
		return nil, fmt.Errorf("only DRAFT cost estimates can be updated")
	}
	est.ProductID = req.ProductID
	est.UOMID = req.UOMID
	est.PlannedQty = req.PlannedQty
	est.BOMID = req.BOMID
	est.RoutingID = req.RoutingID
	est.PlannedStartDate = req.PlannedStartDate
	est.PlannedEndDate = req.PlannedEndDate
	est.Notes = req.Notes
	if err := s.repo.UpdateEstimate(ctx, est); err != nil {
		return nil, err
	}
	return est, nil
}

func (s *Service) DeleteEstimate(ctx context.Context, tenantID, id uint) error {
	est, err := s.repo.GetEstimate(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("cost estimate not found")
	}
	if est.Status != EstimateStatusDraft {
		return fmt.Errorf("only DRAFT cost estimates can be deleted")
	}
	return s.repo.DeleteEstimate(ctx, tenantID, id)
}

func (s *Service) ApproveEstimate(ctx context.Context, tenantID, id, userID uint) (*CostEstimate, error) {
	est, err := s.repo.GetEstimate(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("cost estimate not found")
	}
	if est.Status != EstimateStatusDraft {
		return nil, fmt.Errorf("only DRAFT cost estimates can be approved")
	}
	if len(est.Lines) == 0 {
		return nil, fmt.Errorf("cost estimate must have at least one line to be approved")
	}
	recalcEstimateTotals(est)
	now := time.Now()
	est.Status = EstimateStatusApproved
	est.ApprovedBy = &userID
	est.ApprovedAt = &now
	if err := s.repo.UpdateEstimate(ctx, est); err != nil {
		return nil, err
	}
	return est, nil
}

func (s *Service) CancelEstimate(ctx context.Context, tenantID, id uint) (*CostEstimate, error) {
	est, err := s.repo.GetEstimate(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("cost estimate not found")
	}
	if est.Status == EstimateStatusCancelled {
		return nil, fmt.Errorf("cost estimate is already cancelled")
	}
	est.Status = EstimateStatusCancelled
	if err := s.repo.UpdateEstimate(ctx, est); err != nil {
		return nil, err
	}
	return est, nil
}

// ── Cost Estimate Lines ───────────────────────────────────────────────────────

func (s *Service) ListEstimateLines(ctx context.Context, tenantID, estimateID uint) ([]CostEstimateLine, error) {
	// Verify the estimate belongs to tenant
	if _, err := s.repo.GetEstimate(ctx, tenantID, estimateID); err != nil {
		return nil, fmt.Errorf("cost estimate not found")
	}
	return s.repo.ListEstimateLines(ctx, estimateID)
}

func (s *Service) AddEstimateLine(ctx context.Context, tenantID, estimateID uint, req *AddEstimateLineRequest) (*CostEstimateLine, error) {
	est, err := s.repo.GetEstimate(ctx, tenantID, estimateID)
	if err != nil {
		return nil, fmt.Errorf("cost estimate not found")
	}
	if est.Status != EstimateStatusDraft {
		return nil, fmt.Errorf("lines can only be added to DRAFT cost estimates")
	}

	// Determine next line number
	lineNum := len(est.Lines) + 1

	line := &CostEstimateLine{
		EstimateID:   estimateID,
		TenantID:     tenantID,
		LineNumber:   lineNum,
		LineType:     req.LineType,
		ProductID:    req.ProductID,
		WorkCenterID: req.WorkCenterID,
		Description:  req.Description,
		Quantity:     req.Quantity,
		UOMID:        req.UOMID,
		UnitCost:     req.UnitCost,
		TotalCost:    req.Quantity * req.UnitCost,
		Notes:        req.Notes,
	}
	if err := s.repo.AddEstimateLine(ctx, line); err != nil {
		return nil, err
	}

	// Recalculate header totals
	est.Lines = append(est.Lines, *line)
	recalcEstimateTotals(est)
	if err := s.repo.UpdateEstimate(ctx, est); err != nil {
		return nil, err
	}

	return line, nil
}

func (s *Service) DeleteEstimateLine(ctx context.Context, tenantID, estimateID, lineID uint) error {
	est, err := s.repo.GetEstimate(ctx, tenantID, estimateID)
	if err != nil {
		return fmt.Errorf("cost estimate not found")
	}
	if est.Status != EstimateStatusDraft {
		return fmt.Errorf("lines can only be removed from DRAFT cost estimates")
	}
	if err := s.repo.DeleteEstimateLine(ctx, tenantID, lineID); err != nil {
		return err
	}

	// Reload and recalculate header totals
	est, err = s.repo.GetEstimate(ctx, tenantID, estimateID)
	if err != nil {
		return err
	}
	recalcEstimateTotals(est)
	return s.repo.UpdateEstimate(ctx, est)
}

// ── Production Plans ──────────────────────────────────────────────────────────

func (s *Service) ListPlans(ctx context.Context, tenantID uint, status string) ([]ProductionPlan, error) {
	return s.repo.ListPlans(ctx, tenantID, status)
}

func (s *Service) GetPlan(ctx context.Context, tenantID, id uint) (*ProductionPlan, error) {
	return s.repo.GetPlan(ctx, tenantID, id)
}

func (s *Service) CreatePlan(ctx context.Context, tenantID, userID uint, req *CreatePlanRequest) (*ProductionPlan, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "PRODUCTION_PLAN")
	if err != nil {
		return nil, fmt.Errorf("failed to generate production plan code: %w", err)
	}
	plan := &ProductionPlan{
		TenantID:         tenantID,
		Code:             code,
		SOID:             req.SOID,
		DocumentTypeID:   req.DocumentTypeID,
		ProductID:        req.ProductID,
		UOMID:            req.UOMID,
		PlannedQty:       req.PlannedQty,
		BOMID:            req.BOMID,
		RoutingID:        req.RoutingID,
		EstimateID:       req.EstimateID,
		WarehouseID:      req.WarehouseID,
		PlannedStartDate: req.PlannedStartDate,
		PlannedEndDate:   req.PlannedEndDate,
		Notes:            req.Notes,
		Status:           PlanStatusDraft,
		CreatedBy:        &userID,
	}
	if err := s.repo.CreatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) UpdatePlan(ctx context.Context, tenantID, id uint, req *UpdatePlanRequest) (*ProductionPlan, error) {
	plan, err := s.repo.GetPlan(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("production plan not found")
	}
	if plan.Status != PlanStatusDraft {
		return nil, fmt.Errorf("only DRAFT production plans can be updated")
	}
	plan.SOID = req.SOID
	plan.DocumentTypeID = req.DocumentTypeID
	plan.ProductID = req.ProductID
	plan.UOMID = req.UOMID
	plan.PlannedQty = req.PlannedQty
	plan.BOMID = req.BOMID
	plan.RoutingID = req.RoutingID
	plan.EstimateID = req.EstimateID
	plan.WarehouseID = req.WarehouseID
	plan.PlannedStartDate = req.PlannedStartDate
	plan.PlannedEndDate = req.PlannedEndDate
	plan.Notes = req.Notes
	if err := s.repo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) DeletePlan(ctx context.Context, tenantID, id uint) error {
	plan, err := s.repo.GetPlan(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("production plan not found")
	}
	if plan.Status != PlanStatusDraft {
		return fmt.Errorf("only DRAFT production plans can be deleted")
	}
	return s.repo.DeletePlan(ctx, tenantID, id)
}

// CreateOrderFromPlan spawns a draft ProductionOrder that inherits the
// released Plan's product, qty, warehouse, and inputs. Called by the
// "Create Production" workflow button on the Plan modal.
func (s *Service) CreateOrderFromPlan(ctx context.Context, tenantID, planID, userID uint) (*ProductionOrder, error) {
	plan, err := s.repo.GetPlan(ctx, tenantID, planID)
	if err != nil {
		return nil, fmt.Errorf("production plan not found")
	}
	if plan.Status != PlanStatusReleased {
		return nil, fmt.Errorf("only RELEASED plans can spawn a Production")
	}
	planIDCopy := plan.ID
	req := &CreateOrderRequest{
		PlanID:      &planIDCopy,
		ProductID:   plan.ProductID,
		UOMID:       plan.UOMID,
		PlannedQty:  plan.PlannedQty,
		WarehouseID: plan.WarehouseID,
		StartDate:   plan.PlannedStartDate,
		EndDate:     plan.PlannedEndDate,
		Notes:       plan.Notes,
	}
	return s.CreateOrder(ctx, tenantID, userID, req)
}

// ── Production Plan Inputs ────────────────────────────────────────────────────

func (s *Service) ListPlanInputs(ctx context.Context, tenantID, planID uint) ([]ProductionPlanInput, error) {
	if _, err := s.repo.GetPlan(ctx, tenantID, planID); err != nil {
		return nil, fmt.Errorf("production plan not found")
	}
	return s.repo.ListPlanInputs(ctx, tenantID, planID)
}

func (s *Service) AddPlanInput(ctx context.Context, tenantID, planID uint, req *AddPlanInputRequest) (*ProductionPlanInput, error) {
	plan, err := s.repo.GetPlan(ctx, tenantID, planID)
	if err != nil {
		return nil, fmt.Errorf("production plan not found")
	}
	if plan.Status != PlanStatusDraft {
		return nil, fmt.Errorf("only DRAFT production plans accept new inputs")
	}
	input := &ProductionPlanInput{
		PlanID:     planID,
		TenantID:   tenantID,
		LineNumber: len(plan.Inputs) + 1,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		UOMID:      req.UOMID,
		Notes:      req.Notes,
	}
	if err := s.repo.AddPlanInput(ctx, input); err != nil {
		return nil, err
	}
	return input, nil
}

func (s *Service) UpdatePlanInput(ctx context.Context, tenantID, planID, inputID uint, req *UpdatePlanInputRequest) (*ProductionPlanInput, error) {
	plan, err := s.repo.GetPlan(ctx, tenantID, planID)
	if err != nil {
		return nil, fmt.Errorf("production plan not found")
	}
	if plan.Status != PlanStatusDraft {
		return nil, fmt.Errorf("only DRAFT production plans accept input edits")
	}
	input, err := s.repo.GetPlanInput(ctx, tenantID, planID, inputID)
	if err != nil {
		return nil, fmt.Errorf("plan input not found")
	}
	input.ProductID = req.ProductID
	input.Quantity = req.Quantity
	input.UOMID = req.UOMID
	input.Notes = req.Notes
	if err := s.repo.UpdatePlanInput(ctx, input); err != nil {
		return nil, err
	}
	return input, nil
}

func (s *Service) DeletePlanInput(ctx context.Context, tenantID, planID, inputID uint) error {
	plan, err := s.repo.GetPlan(ctx, tenantID, planID)
	if err != nil {
		return fmt.Errorf("production plan not found")
	}
	if plan.Status != PlanStatusDraft {
		return fmt.Errorf("only DRAFT production plans accept input deletes")
	}
	if _, err := s.repo.GetPlanInput(ctx, tenantID, planID, inputID); err != nil {
		return fmt.Errorf("plan input not found")
	}
	return s.repo.DeletePlanInput(ctx, tenantID, inputID)
}

func (s *Service) ReleasePlan(ctx context.Context, tenantID, id, userID uint) (*ProductionPlan, error) {
	plan, err := s.repo.GetPlan(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("production plan not found")
	}
	if plan.Status != PlanStatusDraft {
		return nil, fmt.Errorf("only DRAFT production plans can be released")
	}
	now := time.Now()
	plan.Status = PlanStatusReleased
	plan.ReleasedBy = &userID
	plan.ReleasedAt = &now
	if err := s.repo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) CancelPlan(ctx context.Context, tenantID, id uint) (*ProductionPlan, error) {
	plan, err := s.repo.GetPlan(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("production plan not found")
	}
	if plan.Status == PlanStatusCompleted {
		return nil, fmt.Errorf("completed production plans cannot be cancelled")
	}
	if plan.Status == PlanStatusCancelled {
		return nil, fmt.Errorf("production plan is already cancelled")
	}
	plan.Status = PlanStatusCancelled
	if err := s.repo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// ── Production Orders ─────────────────────────────────────────────────────────

func (s *Service) ListOrders(ctx context.Context, tenantID uint, status string) ([]ProductionOrder, error) {
	return s.repo.ListOrders(ctx, tenantID, status)
}

func (s *Service) GetOrder(ctx context.Context, tenantID, id uint) (*ProductionOrder, error) {
	return s.repo.GetOrder(ctx, tenantID, id)
}

func (s *Service) CreateOrder(ctx context.Context, tenantID, userID uint, req *CreateOrderRequest) (*ProductionOrder, error) {
	code, err := s.repo.NextCode(ctx, tenantID, "PRODUCTION_ORDER")
	if err != nil {
		return nil, fmt.Errorf("failed to generate production order code: %w", err)
	}
	order := &ProductionOrder{
		TenantID:    tenantID,
		Code:        code,
		PlanID:      req.PlanID,
		SOID:        req.SOID,
		ProductID:   req.ProductID,
		UOMID:       req.UOMID,
		PlannedQty:  req.PlannedQty,
		WarehouseID: req.WarehouseID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Notes:       req.Notes,
		Status:      OrderStatusDraft,
		CreatedBy:   &userID,
	}
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}
	// If this MO was spawned from a Plan, copy the Plan's anticipated inputs
	// into MO ProductionResource rows so the shop-floor form starts with the
	// materials list pre-filled. Best-effort — silently skip if the Plan has
	// no inputs or the copy fails.
	if req.PlanID != nil {
		inputs, err := s.repo.ListPlanInputs(ctx, tenantID, *req.PlanID)
		if err == nil {
			for i, in := range inputs {
				res := &ProductionResource{
					OrderID:      order.ID,
					TenantID:     tenantID,
					LineNumber:   i + 1,
					ResourceType: ResourceTypeMaterial,
					ProductID:    &in.ProductID,
					Quantity:     in.Quantity,
					UOMID:        &in.UOMID,
					Notes:        in.Notes,
				}
				_ = s.repo.AddResource(ctx, res)
			}
		}
	}
	return order, nil
}

func (s *Service) UpdateOrder(ctx context.Context, tenantID, id uint, req *UpdateOrderRequest) (*ProductionOrder, error) {
	order, err := s.repo.GetOrder(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("production order not found")
	}
	if order.Status != OrderStatusDraft {
		return nil, fmt.Errorf("only DRAFT production orders can be updated")
	}
	order.PlanID = req.PlanID
	order.SOID = req.SOID
	order.ProductID = req.ProductID
	order.UOMID = req.UOMID
	order.PlannedQty = req.PlannedQty
	order.WarehouseID = req.WarehouseID
	order.StartDate = req.StartDate
	order.EndDate = req.EndDate
	order.Notes = req.Notes
	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *Service) DeleteOrder(ctx context.Context, tenantID, id uint) error {
	order, err := s.repo.GetOrder(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("production order not found")
	}
	if order.Status != OrderStatusDraft {
		return fmt.Errorf("only DRAFT production orders can be deleted")
	}
	return s.repo.DeleteOrder(ctx, tenantID, id)
}

func (s *Service) StartOrder(ctx context.Context, tenantID, id, userID uint) (*ProductionOrder, error) {
	order, err := s.repo.GetOrder(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("production order not found")
	}
	if order.Status != OrderStatusDraft {
		return nil, fmt.Errorf("only DRAFT production orders can be started")
	}
	order.Status = OrderStatusInProgress
	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *Service) CompleteOrder(ctx context.Context, tenantID, id, userID uint) (*ProductionOrder, error) {
	if err := s.repo.CompleteOrder(ctx, tenantID, id, userID); err != nil {
		return nil, err
	}
	order, err := s.repo.GetOrder(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	s.publish(ctx, TypeOrderCompleted, tenantID, userID, map[string]interface{}{
		"order_id":     id,
		"code":         order.Code,
		"produced_qty": order.ProducedQty,
	})
	return order, nil
}

func (s *Service) CancelOrder(ctx context.Context, tenantID, id uint) (*ProductionOrder, error) {
	order, err := s.repo.GetOrder(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("production order not found")
	}
	if order.Status == OrderStatusCompleted {
		return nil, fmt.Errorf("completed production orders cannot be cancelled")
	}
	if order.Status == OrderStatusCancelled {
		return nil, fmt.Errorf("production order is already cancelled")
	}
	order.Status = OrderStatusCancelled
	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

// ── Production Outputs ────────────────────────────────────────────────────────

func (s *Service) ListOutputs(ctx context.Context, tenantID, orderID uint) ([]ProductionOutput, error) {
	if _, err := s.repo.GetOrder(ctx, tenantID, orderID); err != nil {
		return nil, fmt.Errorf("production order not found")
	}
	return s.repo.ListOutputs(ctx, orderID)
}

func (s *Service) AddOutput(ctx context.Context, tenantID, orderID uint, req *AddOutputRequest) (*ProductionOutput, error) {
	order, err := s.repo.GetOrder(ctx, tenantID, orderID)
	if err != nil {
		return nil, fmt.Errorf("production order not found")
	}
	if order.Status != OrderStatusDraft && order.Status != OrderStatusInProgress {
		return nil, fmt.Errorf("outputs can only be added to DRAFT or IN_PROGRESS orders")
	}

	lineNum := len(order.Outputs) + 1
	output := &ProductionOutput{
		OrderID:     orderID,
		TenantID:    tenantID,
		LineNumber:  lineNum,
		ProductID:   req.ProductID,
		UOMID:       req.UOMID,
		Quantity:    req.Quantity,
		UnitCost:    req.UnitCost,
		TotalCost:   req.Quantity * req.UnitCost,
		WarehouseID: req.WarehouseID,
		LocationID:  req.LocationID,
		Notes:       req.Notes,
	}
	if err := s.repo.AddOutput(ctx, output); err != nil {
		return nil, err
	}
	// Auto-create a Product QC for the newly recorded output.
	if s.qc != nil {
		whID := order.WarehouseID
		if output.WarehouseID != nil {
			whID = *output.WarehouseID
		}
		_ = s.qc.CreateAutoQC(ctx, tenantID, 0,
			"PRODUCT_QC", "PRODUCTION_OUTPUT", output.ID, whID,
			"Auto-created from production order "+order.Code+" output line "+fmt.Sprintf("%d", output.LineNumber),
			[]QCAutoLine{{ProductID: output.ProductID, Quantity: output.Quantity}})
	}
	return output, nil
}

func (s *Service) DeleteOutput(ctx context.Context, tenantID, orderID, outputID uint) error {
	order, err := s.repo.GetOrder(ctx, tenantID, orderID)
	if err != nil {
		return fmt.Errorf("production order not found")
	}
	if order.Status != OrderStatusDraft && order.Status != OrderStatusInProgress {
		return fmt.Errorf("outputs can only be removed from DRAFT or IN_PROGRESS orders")
	}
	return s.repo.DeleteOutput(ctx, tenantID, outputID)
}

// ── Production Resources ──────────────────────────────────────────────────────

func (s *Service) ListResources(ctx context.Context, tenantID, orderID uint) ([]ProductionResource, error) {
	if _, err := s.repo.GetOrder(ctx, tenantID, orderID); err != nil {
		return nil, fmt.Errorf("production order not found")
	}
	return s.repo.ListResources(ctx, orderID)
}

func (s *Service) AddResource(ctx context.Context, tenantID, orderID uint, req *AddResourceRequest) (*ProductionResource, error) {
	order, err := s.repo.GetOrder(ctx, tenantID, orderID)
	if err != nil {
		return nil, fmt.Errorf("production order not found")
	}
	if order.Status != OrderStatusDraft && order.Status != OrderStatusInProgress {
		return nil, fmt.Errorf("resources can only be added to DRAFT or IN_PROGRESS orders")
	}

	lineNum := len(order.Resources) + 1
	resource := &ProductionResource{
		OrderID:      orderID,
		TenantID:     tenantID,
		LineNumber:   lineNum,
		ResourceType: req.ResourceType,
		ProductID:    req.ProductID,
		WorkCenterID: req.WorkCenterID,
		Description:  req.Description,
		Quantity:     req.Quantity,
		UOMID:        req.UOMID,
		UnitCost:     req.UnitCost,
		TotalCost:    req.Quantity * req.UnitCost,
		Notes:        req.Notes,
	}
	if err := s.repo.AddResource(ctx, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

func (s *Service) DeleteResource(ctx context.Context, tenantID, orderID, resourceID uint) error {
	order, err := s.repo.GetOrder(ctx, tenantID, orderID)
	if err != nil {
		return fmt.Errorf("production order not found")
	}
	if order.Status != OrderStatusDraft && order.Status != OrderStatusInProgress {
		return fmt.Errorf("resources can only be removed from DRAFT or IN_PROGRESS orders")
	}
	return s.repo.DeleteResource(ctx, tenantID, resourceID)
}

// ── Post Costs ────────────────────────────────────────────────────────────────

func (s *Service) ListPostCosts(ctx context.Context, tenantID uint, status string) ([]PostCost, error) {
	return s.repo.ListPostCosts(ctx, tenantID, status)
}

func (s *Service) GetPostCost(ctx context.Context, tenantID, id uint) (*PostCost, error) {
	return s.repo.GetPostCost(ctx, tenantID, id)
}

func (s *Service) CreatePostCost(ctx context.Context, tenantID, userID uint, req *CreatePostCostRequest) (*PostCost, error) {
	// Verify the production order belongs to tenant
	order, err := s.repo.GetOrder(ctx, tenantID, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("production order not found")
	}

	code, err := s.repo.NextCode(ctx, tenantID, "POST_COST")
	if err != nil {
		return nil, fmt.Errorf("failed to generate post-cost code: %w", err)
	}

	pc := &PostCost{
		TenantID:  tenantID,
		Code:      code,
		OrderID:   req.OrderID,
		EstimateID: req.EstimateID,
		Notes:     req.Notes,
		Status:    PostCostStatusDraft,
		CreatedBy: &userID,
	}

	// Compute actuals from order resources
	actMat, actRes, actOvh := computeActuals(order.Resources)
	pc.ActualMaterialCost = actMat
	pc.ActualResourceCost = actRes
	pc.ActualOverheadCost = actOvh
	pc.TotalActualCost = actMat + actRes + actOvh

	// If estimate_id provided, pull estimated costs
	if req.EstimateID != nil {
		est, err := s.repo.GetEstimate(ctx, tenantID, *req.EstimateID)
		if err == nil {
			pc.EstimatedMaterialCost = est.EstimatedMaterialCost
			pc.EstimatedResourceCost = est.EstimatedResourceCost
			pc.EstimatedOverheadCost = est.EstimatedOverheadCost
			pc.TotalEstimatedCost = est.TotalEstimatedCost
		}
	}

	// Calculate variance
	pc.VarianceAmount = pc.TotalActualCost - pc.TotalEstimatedCost
	if pc.TotalEstimatedCost > 0 {
		pc.VariancePct = pc.VarianceAmount / pc.TotalEstimatedCost * 100
	}

	if err := s.repo.CreatePostCost(ctx, pc); err != nil {
		return nil, err
	}
	return pc, nil
}

func (s *Service) FinalizePostCost(ctx context.Context, tenantID, id, userID uint) (*PostCost, error) {
	pc, err := s.repo.GetPostCost(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("post-cost record not found")
	}
	if pc.Status != PostCostStatusDraft {
		return nil, fmt.Errorf("only DRAFT post-cost records can be finalized")
	}

	// Re-compute actuals from current order state
	order, err := s.repo.GetOrder(ctx, tenantID, pc.OrderID)
	if err != nil {
		return nil, fmt.Errorf("production order not found")
	}
	actMat, actRes, actOvh := computeActuals(order.Resources)
	pc.ActualMaterialCost = actMat
	pc.ActualResourceCost = actRes
	pc.ActualOverheadCost = actOvh
	pc.TotalActualCost = actMat + actRes + actOvh

	// Recalculate variance
	pc.VarianceAmount = pc.TotalActualCost - pc.TotalEstimatedCost
	if pc.TotalEstimatedCost > 0 {
		pc.VariancePct = pc.VarianceAmount / pc.TotalEstimatedCost * 100
	} else {
		pc.VariancePct = 0
	}

	now := time.Now()
	pc.Status = PostCostStatusFinalized
	pc.FinalizedBy = &userID
	pc.FinalizedAt = &now

	if err := s.repo.UpdatePostCost(ctx, pc); err != nil {
		return nil, err
	}
	return pc, nil
}
