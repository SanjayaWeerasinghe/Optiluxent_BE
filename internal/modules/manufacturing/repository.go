package manufacturing

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Repository defines all data access operations for the manufacturing module.
type Repository interface {
	// Document sequence
	NextCode(ctx context.Context, tenantID uint, docType string) (string, error)

	// Cost Estimates
	ListEstimates(ctx context.Context, tenantID uint, status string) ([]CostEstimate, error)
	GetEstimate(ctx context.Context, tenantID, id uint) (*CostEstimate, error)
	CreateEstimate(ctx context.Context, est *CostEstimate) error
	UpdateEstimate(ctx context.Context, est *CostEstimate) error
	DeleteEstimate(ctx context.Context, tenantID, id uint) error

	// Cost Estimate Lines
	ListEstimateLines(ctx context.Context, estimateID uint) ([]CostEstimateLine, error)
	AddEstimateLine(ctx context.Context, line *CostEstimateLine) error
	DeleteEstimateLine(ctx context.Context, tenantID, lineID uint) error

	// Production Plans. `limit=0` = unbounded.
	ListPlans(ctx context.Context, tenantID uint, status string, limit, offset int) ([]ProductionPlan, error)
	CountPlans(ctx context.Context, tenantID uint, status string) (int64, error)
	GetPlan(ctx context.Context, tenantID, id uint) (*ProductionPlan, error)
	CreatePlan(ctx context.Context, plan *ProductionPlan) error
	UpdatePlan(ctx context.Context, plan *ProductionPlan) error
	DeletePlan(ctx context.Context, tenantID, id uint) error

	// Production Plan Inputs
	ListPlanInputs(ctx context.Context, tenantID, planID uint) ([]ProductionPlanInput, error)
	GetPlanInput(ctx context.Context, tenantID, planID, inputID uint) (*ProductionPlanInput, error)
	AddPlanInput(ctx context.Context, input *ProductionPlanInput) error
	UpdatePlanInput(ctx context.Context, input *ProductionPlanInput) error
	DeletePlanInput(ctx context.Context, tenantID, inputID uint) error

	// Production Orders. `limit=0` = unbounded.
	ListOrders(ctx context.Context, tenantID uint, status string, limit, offset int) ([]ProductionOrder, error)
	CountOrders(ctx context.Context, tenantID uint, status string) (int64, error)
	GetOrder(ctx context.Context, tenantID, id uint) (*ProductionOrder, error)
	IncrementProducedQty(ctx context.Context, tenantID, id uint, delta float64) error
	CreateOrder(ctx context.Context, order *ProductionOrder) error
	UpdateOrder(ctx context.Context, order *ProductionOrder) error
	DeleteOrder(ctx context.Context, tenantID, id uint) error
	CompleteOrder(ctx context.Context, tenantID, orderID, userID uint) error

	// Production Outputs
	ListOutputs(ctx context.Context, orderID uint) ([]ProductionOutput, error)
	AddOutput(ctx context.Context, output *ProductionOutput) error
	DeleteOutput(ctx context.Context, tenantID, id uint) error

	// Production Resources
	ListResources(ctx context.Context, orderID uint) ([]ProductionResource, error)
	AddResource(ctx context.Context, resource *ProductionResource) error
	DeleteResource(ctx context.Context, tenantID, id uint) error

	// Post Costs
	ListPostCosts(ctx context.Context, tenantID uint, status string) ([]PostCost, error)
	GetPostCost(ctx context.Context, tenantID, id uint) (*PostCost, error)
	CreatePostCost(ctx context.Context, pc *PostCost) error
	UpdatePostCost(ctx context.Context, pc *PostCost) error
	DeletePostCost(ctx context.Context, tenantID, id uint) error
}

type dbRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &dbRepository{db: db}
}

// NextCode fetches the next formatted document code and increments the counter atomically.
func (r *dbRepository) NextCode(ctx context.Context, tenantID uint, docType string) (string, error) {
	var prefix, suffix string
	var nextNum, padding int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Raw(`
			SELECT prefix, COALESCE(suffix,''), next_number, COALESCE(padding, 5)
			FROM document_sequences
			WHERE tenant_id = ? AND document_type = ?
			FOR UPDATE`, tenantID, docType).
			Row().Scan(&prefix, &suffix, &nextNum, &padding)
		if err != nil {
			return fmt.Errorf("document sequence not found for %s: %w", docType, err)
		}
		return tx.Exec(`
			UPDATE document_sequences SET next_number = next_number + 1
			WHERE tenant_id = ? AND document_type = ?`, tenantID, docType).Error
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%0*d%s", prefix, padding, nextNum, suffix), nil
}

// upsertStockBalance adjusts the stock balance for a product/warehouse by delta.
// Uses the same ON CONFLICT constraint as the inventory module.
func upsertStockBalance(tx *gorm.DB, tenantID, productID uint, variantID *uint, warehouseID uint, locationID *uint, delta float64) error {
	return tx.Exec(`INSERT INTO stock_balances (tenant_id, product_id, variant_id, warehouse_id, location_id, quantity, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
		ON CONFLICT ON CONSTRAINT uidx_stock_balances
		DO UPDATE SET quantity = stock_balances.quantity + EXCLUDED.quantity, updated_at = NOW()
	`, tenantID, productID, variantID, warehouseID, locationID, delta).Error
}

// ── Cost Estimates ────────────────────────────────────────────────────────────

func (r *dbRepository) ListEstimates(ctx context.Context, tenantID uint, status string) ([]CostEstimate, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []CostEstimate
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetEstimate(ctx context.Context, tenantID, id uint) (*CostEstimate, error) {
	var est CostEstimate
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&est).Error
	return &est, err
}

func (r *dbRepository) CreateEstimate(ctx context.Context, est *CostEstimate) error {
	return r.db.WithContext(ctx).Create(est).Error
}

func (r *dbRepository) UpdateEstimate(ctx context.Context, est *CostEstimate) error {
	return r.db.WithContext(ctx).Save(est).Error
}

func (r *dbRepository) DeleteEstimate(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&CostEstimate{}).Error
}

// ── Cost Estimate Lines ───────────────────────────────────────────────────────

func (r *dbRepository) ListEstimateLines(ctx context.Context, estimateID uint) ([]CostEstimateLine, error) {
	var rows []CostEstimateLine
	return rows, r.db.WithContext(ctx).Where("estimate_id = ?", estimateID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) AddEstimateLine(ctx context.Context, line *CostEstimateLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) DeleteEstimateLine(ctx context.Context, tenantID, lineID uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, lineID).Delete(&CostEstimateLine{}).Error
}

// ── Production Plans ──────────────────────────────────────────────────────────

func (r *dbRepository) ListPlans(ctx context.Context, tenantID uint, status string, limit, offset int) ([]ProductionPlan, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q = q.Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []ProductionPlan
	return rows, q.Find(&rows).Error
}

func (r *dbRepository) CountPlans(ctx context.Context, tenantID uint, status string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&ProductionPlan{}).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var n int64
	return n, q.Count(&n).Error
}

func (r *dbRepository) GetPlan(ctx context.Context, tenantID, id uint) (*ProductionPlan, error) {
	var plan ProductionPlan
	err := r.db.WithContext(ctx).Preload("Inputs").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&plan).Error
	return &plan, err
}

// ── Production Plan Inputs ────────────────────────────────────────────────────

func (r *dbRepository) ListPlanInputs(ctx context.Context, tenantID, planID uint) ([]ProductionPlanInput, error) {
	var rows []ProductionPlanInput
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND plan_id = ?", tenantID, planID).
		Order("line_number ASC").Find(&rows).Error
}

func (r *dbRepository) GetPlanInput(ctx context.Context, tenantID, planID, inputID uint) (*ProductionPlanInput, error) {
	var row ProductionPlanInput
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND plan_id = ? AND id = ?", tenantID, planID, inputID).First(&row).Error
	return &row, err
}

func (r *dbRepository) AddPlanInput(ctx context.Context, input *ProductionPlanInput) error {
	return r.db.WithContext(ctx).Create(input).Error
}

func (r *dbRepository) UpdatePlanInput(ctx context.Context, input *ProductionPlanInput) error {
	return r.db.WithContext(ctx).Save(input).Error
}

func (r *dbRepository) DeletePlanInput(ctx context.Context, tenantID, inputID uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, inputID).
		Delete(&ProductionPlanInput{}).Error
}

func (r *dbRepository) CreatePlan(ctx context.Context, plan *ProductionPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *dbRepository) UpdatePlan(ctx context.Context, plan *ProductionPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *dbRepository) DeletePlan(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&ProductionPlan{}).Error
}

// ── Production Orders ─────────────────────────────────────────────────────────

func (r *dbRepository) ListOrders(ctx context.Context, tenantID uint, status string, limit, offset int) ([]ProductionOrder, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q = q.Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []ProductionOrder
	return rows, q.Find(&rows).Error
}

func (r *dbRepository) CountOrders(ctx context.Context, tenantID uint, status string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&ProductionOrder{}).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var n int64
	return n, q.Count(&n).Error
}

func (r *dbRepository) GetOrder(ctx context.Context, tenantID, id uint) (*ProductionOrder, error) {
	var order ProductionOrder
	err := r.db.WithContext(ctx).Preload("Outputs").Preload("Resources").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&order).Error
	return &order, err
}

func (r *dbRepository) CreateOrder(ctx context.Context, order *ProductionOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *dbRepository) UpdateOrder(ctx context.Context, order *ProductionOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// IncrementProducedQty adds delta (may be negative) to the order's
// produced_qty atomically so concurrent GRN confirms can't clobber each other.
func (r *dbRepository) IncrementProducedQty(ctx context.Context, tenantID, id uint, delta float64) error {
	return r.db.WithContext(ctx).Exec(
		`UPDATE production_orders SET produced_qty = produced_qty + ?, updated_at = NOW()
		 WHERE tenant_id = ? AND id = ?`,
		delta, tenantID, id,
	).Error
}

func (r *dbRepository) DeleteOrder(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&ProductionOrder{}).Error
}

// CompleteOrder transitions an order to COMPLETED inside a single transaction:
// debits MATERIAL resources from stock and credits all outputs to stock.
func (r *dbRepository) CompleteOrder(ctx context.Context, tenantID, orderID, userID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order ProductionOrder
		if err := tx.Preload("Outputs").Preload("Resources").
			Where("tenant_id = ? AND id = ?", tenantID, orderID).First(&order).Error; err != nil {
			return err
		}
		if order.Status != OrderStatusInProgress {
			return fmt.Errorf("production order is not IN_PROGRESS")
		}

		// Debit MATERIAL resources from stock (consumption is unconditional)
		for _, res := range order.Resources {
			if res.ResourceType == ResourceTypeMaterial && res.ProductID != nil {
				if err := upsertStockBalance(tx, tenantID, *res.ProductID, nil, order.WarehouseID, nil, -res.Quantity); err != nil {
					return fmt.Errorf("failed to debit material %d: %w", *res.ProductID, err)
				}
			}
		}

		// Outputs are NOT credited to stock here — they are gated on Product QC
		// passing (inventory.SubmitQualityCheck posts qty_passed to stock_balances).
		// We still track totalProduced for the order's produced_qty field.
		var totalProduced float64
		for _, out := range order.Outputs {
			totalProduced += out.Quantity
		}

		// Update order
		return tx.Model(&ProductionOrder{}).
			Where("tenant_id = ? AND id = ?", tenantID, orderID).
			Updates(map[string]interface{}{
				"status":       OrderStatusCompleted,
				"produced_qty": totalProduced,
				"completed_by": userID,
				"completed_at": gorm.Expr("NOW()"),
			}).Error
	})
}

// ── Production Outputs ────────────────────────────────────────────────────────

func (r *dbRepository) ListOutputs(ctx context.Context, orderID uint) ([]ProductionOutput, error) {
	var rows []ProductionOutput
	return rows, r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) AddOutput(ctx context.Context, output *ProductionOutput) error {
	return r.db.WithContext(ctx).Create(output).Error
}

func (r *dbRepository) DeleteOutput(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&ProductionOutput{}).Error
}

// ── Production Resources ──────────────────────────────────────────────────────

func (r *dbRepository) ListResources(ctx context.Context, orderID uint) ([]ProductionResource, error) {
	var rows []ProductionResource
	return rows, r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) AddResource(ctx context.Context, resource *ProductionResource) error {
	return r.db.WithContext(ctx).Create(resource).Error
}

func (r *dbRepository) DeleteResource(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&ProductionResource{}).Error
}

// ── Post Costs ────────────────────────────────────────────────────────────────

func (r *dbRepository) ListPostCosts(ctx context.Context, tenantID uint, status string) ([]PostCost, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []PostCost
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetPostCost(ctx context.Context, tenantID, id uint) (*PostCost, error) {
	var pc PostCost
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&pc).Error
	return &pc, err
}

func (r *dbRepository) CreatePostCost(ctx context.Context, pc *PostCost) error {
	return r.db.WithContext(ctx).Create(pc).Error
}

func (r *dbRepository) UpdatePostCost(ctx context.Context, pc *PostCost) error {
	return r.db.WithContext(ctx).Save(pc).Error
}

func (r *dbRepository) DeletePostCost(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&PostCost{}).Error
}
