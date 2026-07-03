package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ledgerRow holds data for a single stock_ledger entry written to ClickHouse.
type ledgerRow struct {
	tenantID    uint
	productID   uint
	variantID   *uint
	warehouseID uint
	locationID  *uint
	txType      string
	refType     string
	refID       uint
	qty         float64
	unitCost    float64
	totalCost   float64
	date        string
	createdBy   uint
}

// Repository defines all persistence operations for the inventory module.
type Repository interface {
	// Document sequence
	NextCode(ctx context.Context, tenantID uint, docType string) (string, error)

	// Material Requests
	ListMRs(ctx context.Context, tenantID uint, status string) ([]MaterialRequest, error)
	ListMRsByMO(ctx context.Context, tenantID, moID uint) ([]MaterialRequest, error)
	GetMR(ctx context.Context, tenantID, id uint) (*MaterialRequest, error)
	CreateMR(ctx context.Context, mr *MaterialRequest) error
	UpdateMR(ctx context.Context, mr *MaterialRequest) error
	ApproveMR(ctx context.Context, tenantID, id, userID uint) error
	RejectMR(ctx context.Context, tenantID, id, userID uint, reason string) error
	ListMRLines(ctx context.Context, mrID uint) ([]MRLine, error)
	AddMRLine(ctx context.Context, line *MRLine) error
	UpdateMRLine(ctx context.Context, line *MRLine) error
	DeleteMRLine(ctx context.Context, tenantID, id uint) error

	// Goods Transfers
	ListTransfers(ctx context.Context, tenantID uint, status string) ([]GoodsTransfer, error)
	ListTransfersByMO(ctx context.Context, tenantID, moID uint) ([]GoodsTransfer, error)
	GetTransfer(ctx context.Context, tenantID, id uint) (*GoodsTransfer, error)
	CreateTransfer(ctx context.Context, t *GoodsTransfer) error
	UpdateTransfer(ctx context.Context, t *GoodsTransfer) error
	ListTransferLines(ctx context.Context, transferID uint) ([]GTLine, error)
	AddTransferLine(ctx context.Context, line *GTLine) error
	UpdateTransferLine(ctx context.Context, line *GTLine) error
	DeleteTransferLine(ctx context.Context, tenantID, id uint) error
	SendTransfer(ctx context.Context, tenantID, id, userID uint) error
	ReceiveTransfer(ctx context.Context, tenantID, id, userID uint) error

	// Goods Issues
	ListIssues(ctx context.Context, tenantID uint, status, reason string) ([]GoodsIssue, error)
	ListIssuesByMO(ctx context.Context, tenantID, moID uint) ([]GoodsIssue, error)
	GetIssue(ctx context.Context, tenantID, id uint) (*GoodsIssue, error)
	CreateIssue(ctx context.Context, gi *GoodsIssue) error
	UpdateIssue(ctx context.Context, gi *GoodsIssue) error
	ListIssueLines(ctx context.Context, issueID uint) ([]GILine, error)
	AddIssueLine(ctx context.Context, line *GILine) error
	UpdateIssueLine(ctx context.Context, line *GILine) error
	DeleteIssueLine(ctx context.Context, tenantID, id uint) error
	ConfirmIssue(ctx context.Context, tenantID, id, userID uint) error

	// Stock Adjustments
	ListAdjustments(ctx context.Context, tenantID uint, status string) ([]StockAdjustment, error)
	GetAdjustment(ctx context.Context, tenantID, id uint) (*StockAdjustment, error)
	CreateAdjustment(ctx context.Context, sa *StockAdjustment) error
	UpdateAdjustment(ctx context.Context, sa *StockAdjustment) error
	ListAdjustmentLines(ctx context.Context, adjID uint) ([]SALine, error)
	AddAdjustmentLine(ctx context.Context, line *SALine) error
	UpdateAdjustmentLine(ctx context.Context, line *SALine) error
	DeleteAdjustmentLine(ctx context.Context, tenantID, id uint) error
	ConfirmAdjustment(ctx context.Context, tenantID, id, userID uint) error

	// Quality Checks
	ListQualityChecks(ctx context.Context, tenantID uint, status string) ([]QualityCheck, error)
	ListQCsByRefs(ctx context.Context, tenantID uint, refType string, refIDs []uint) ([]QualityCheck, error)
	GetQualityCheck(ctx context.Context, tenantID, id uint) (*QualityCheck, error)
	CreateQualityCheck(ctx context.Context, qc *QualityCheck) error
	UpdateQualityCheckStatus(ctx context.Context, tenantID, id uint, status string) error
	ListQCLines(ctx context.Context, checkID uint) ([]QCLine, error)
	AddQCLine(ctx context.Context, line *QCLine) error
	UpdateQCLine(ctx context.Context, line *QCLine) error
	SubmitQualityCheck(ctx context.Context, tenantID, id, userID uint) error

	// Stock Balances
	GetStockBalance(ctx context.Context, tenantID uint, warehouseID *uint, productID *uint) ([]StockBalanceRow, error)
}

type dbRepository struct {
	db       *gorm.DB
	ledgerDB *sql.DB
}

func NewRepository(db *gorm.DB, ledgerDB *sql.DB) Repository {
	return &dbRepository{db: db, ledgerDB: ledgerDB}
}

// ── Document Sequence ─────────────────────────────────────────────────────────

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

// ── Stock helpers ─────────────────────────────────────────────────────────────

func (r *dbRepository) upsertStockBalance(tx *gorm.DB, tenantID, productID uint, variantID *uint, warehouseID uint, locationID *uint, delta float64) error {
	return tx.Exec(`
		INSERT INTO stock_balances (tenant_id, product_id, variant_id, warehouse_id, location_id, quantity, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
		ON CONFLICT ON CONSTRAINT uidx_stock_balances
		DO UPDATE SET quantity = stock_balances.quantity + EXCLUDED.quantity, updated_at = NOW()
	`, tenantID, productID, variantID, warehouseID, locationID, delta).Error
}

func (r *dbRepository) writeToLedger(ctx context.Context, rows []ledgerRow) {
	if r.ledgerDB == nil {
		return
	}
	for _, row := range rows {
		_, _ = r.ledgerDB.ExecContext(ctx, `
			INSERT INTO stock_ledger
				(tenant_id, product_id, variant_id, warehouse_id, location_id,
				 transaction_type, reference_type, reference_id,
				 quantity, unit_cost, total_cost, transaction_date, created_by)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			row.tenantID, row.productID, row.variantID, row.warehouseID, row.locationID,
			row.txType, row.refType, row.refID,
			row.qty, row.unitCost, row.totalCost, row.date, row.createdBy,
		)
	}
}

// ── Material Requests ─────────────────────────────────────────────────────────

func (r *dbRepository) ListMRs(ctx context.Context, tenantID uint, status string) ([]MaterialRequest, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []MaterialRequest
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) ListMRsByMO(ctx context.Context, tenantID, moID uint) ([]MaterialRequest, error) {
	var rows []MaterialRequest
	return rows, r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND mo_id = ?", tenantID, moID).
		Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetMR(ctx context.Context, tenantID, id uint) (*MaterialRequest, error) {
	var mr MaterialRequest
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&mr).Error
	return &mr, err
}

func (r *dbRepository) CreateMR(ctx context.Context, mr *MaterialRequest) error {
	return r.db.WithContext(ctx).Create(mr).Error
}

func (r *dbRepository) UpdateMR(ctx context.Context, mr *MaterialRequest) error {
	return r.db.WithContext(ctx).Save(mr).Error
}

func (r *dbRepository) ApproveMR(ctx context.Context, tenantID, id, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&MaterialRequest{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"status":      MRStatusApproved,
			"approved_by": userID,
			"approved_at": now,
		}).Error
}

func (r *dbRepository) RejectMR(ctx context.Context, tenantID, id, userID uint, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&MaterialRequest{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"status":        MRStatusRejected,
			"rejected_by":   userID,
			"rejected_at":   now,
			"reject_reason": reason,
		}).Error
}

func (r *dbRepository) ListMRLines(ctx context.Context, mrID uint) ([]MRLine, error) {
	var rows []MRLine
	return rows, r.db.WithContext(ctx).Where("mr_id = ?", mrID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) AddMRLine(ctx context.Context, line *MRLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) UpdateMRLine(ctx context.Context, line *MRLine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) DeleteMRLine(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&MRLine{}).Error
}

// ── Goods Transfers ───────────────────────────────────────────────────────────

func (r *dbRepository) ListTransfers(ctx context.Context, tenantID uint, status string) ([]GoodsTransfer, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []GoodsTransfer
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) ListTransfersByMO(ctx context.Context, tenantID, moID uint) ([]GoodsTransfer, error) {
	var rows []GoodsTransfer
	return rows, r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND mo_id = ?", tenantID, moID).
		Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetTransfer(ctx context.Context, tenantID, id uint) (*GoodsTransfer, error) {
	var t GoodsTransfer
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&t).Error
	return &t, err
}

func (r *dbRepository) CreateTransfer(ctx context.Context, t *GoodsTransfer) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *dbRepository) UpdateTransfer(ctx context.Context, t *GoodsTransfer) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *dbRepository) ListTransferLines(ctx context.Context, transferID uint) ([]GTLine, error) {
	var rows []GTLine
	return rows, r.db.WithContext(ctx).Where("transfer_id = ?", transferID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) AddTransferLine(ctx context.Context, line *GTLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) UpdateTransferLine(ctx context.Context, line *GTLine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) DeleteTransferLine(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&GTLine{}).Error
}

func (r *dbRepository) SendTransfer(ctx context.Context, tenantID, id, userID uint) error {
	var t GoodsTransfer
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&t).Error; err != nil {
		return err
	}
	if t.Status != GTStatusDraft {
		return fmt.Errorf("goods transfer is already %s", t.Status)
	}
	now := time.Now()
	return r.db.WithContext(ctx).Model(&GoodsTransfer{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"status":  GTStatusInTransit,
			"sent_by": userID,
			"sent_at": now,
		}).Error
}

func (r *dbRepository) ReceiveTransfer(ctx context.Context, tenantID, id, userID uint) error {
	var ledgerRows []ledgerRow

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var t GoodsTransfer
		if err := tx.Preload("Lines").
			Where("tenant_id = ? AND id = ?", tenantID, id).First(&t).Error; err != nil {
			return err
		}
		if t.Status != GTStatusInTransit {
			return fmt.Errorf("goods transfer must be IN_TRANSIT to receive, current status: %s", t.Status)
		}
		now := time.Now()
		t.Status = GTStatusReceived
		t.ReceivedBy = &userID
		t.ReceivedAt = &now
		if err := tx.Save(&t).Error; err != nil {
			return err
		}
		for _, line := range t.Lines {
			// Subtract from source warehouse/location
			if err := r.upsertStockBalance(tx, tenantID, line.ProductID, line.VariantID, t.FromWarehouseID, line.FromLocationID, -line.Quantity); err != nil {
				return fmt.Errorf("failed to deduct stock from source on line %d: %w", line.LineNumber, err)
			}
			// Add to destination warehouse/location
			if err := r.upsertStockBalance(tx, tenantID, line.ProductID, line.VariantID, t.ToWarehouseID, line.ToLocationID, line.Quantity); err != nil {
				return fmt.Errorf("failed to add stock to destination on line %d: %w", line.LineNumber, err)
			}
			ledgerRows = append(ledgerRows, ledgerRow{
				tenantID:    tenantID,
				productID:   line.ProductID,
				variantID:   line.VariantID,
				warehouseID: t.FromWarehouseID,
				locationID:  line.FromLocationID,
				txType:      "TRANSFER_OUT",
				refType:     "GOODS_TRANSFER",
				refID:       t.ID,
				qty:         -line.Quantity,
				unitCost:    0,
				totalCost:   0,
				date:        t.TransferDate,
				createdBy:   userID,
			}, ledgerRow{
				tenantID:    tenantID,
				productID:   line.ProductID,
				variantID:   line.VariantID,
				warehouseID: t.ToWarehouseID,
				locationID:  line.ToLocationID,
				txType:      "TRANSFER_IN",
				refType:     "GOODS_TRANSFER",
				refID:       t.ID,
				qty:         line.Quantity,
				unitCost:    0,
				totalCost:   0,
				date:        t.TransferDate,
				createdBy:   userID,
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	r.writeToLedger(ctx, ledgerRows)
	return nil
}

// ── Goods Issues ──────────────────────────────────────────────────────────────

func (r *dbRepository) ListIssues(ctx context.Context, tenantID uint, status, reason string) ([]GoodsIssue, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if reason != "" {
		q = q.Where("issue_reason = ?", reason)
	}
	var rows []GoodsIssue
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) ListIssuesByMO(ctx context.Context, tenantID, moID uint) ([]GoodsIssue, error) {
	var rows []GoodsIssue
	return rows, r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND reference_type = ? AND reference_id = ?", tenantID, "PRODUCTION_ORDER", moID).
		Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetIssue(ctx context.Context, tenantID, id uint) (*GoodsIssue, error) {
	var gi GoodsIssue
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&gi).Error
	return &gi, err
}

func (r *dbRepository) CreateIssue(ctx context.Context, gi *GoodsIssue) error {
	return r.db.WithContext(ctx).Create(gi).Error
}

func (r *dbRepository) UpdateIssue(ctx context.Context, gi *GoodsIssue) error {
	return r.db.WithContext(ctx).Save(gi).Error
}

func (r *dbRepository) ListIssueLines(ctx context.Context, issueID uint) ([]GILine, error) {
	var rows []GILine
	return rows, r.db.WithContext(ctx).Where("issue_id = ?", issueID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) AddIssueLine(ctx context.Context, line *GILine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) UpdateIssueLine(ctx context.Context, line *GILine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) DeleteIssueLine(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&GILine{}).Error
}

func (r *dbRepository) ConfirmIssue(ctx context.Context, tenantID, id, userID uint) error {
	var ledgerRows []ledgerRow

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var gi GoodsIssue
		if err := tx.Preload("Lines").
			Where("tenant_id = ? AND id = ?", tenantID, id).First(&gi).Error; err != nil {
			return err
		}
		if gi.Status != GIStatusDraft {
			return fmt.Errorf("goods issue is already %s", gi.Status)
		}

		// Check sufficient stock for each line
		for _, line := range gi.Lines {
			var available float64
			q := tx.Raw(`
				SELECT COALESCE(SUM(quantity), 0) FROM stock_balances
				WHERE tenant_id = ? AND product_id = ? AND warehouse_id = ?`,
				tenantID, line.ProductID, gi.WarehouseID)
			if line.VariantID != nil {
				q = tx.Raw(`
					SELECT COALESCE(SUM(quantity), 0) FROM stock_balances
					WHERE tenant_id = ? AND product_id = ? AND warehouse_id = ? AND variant_id = ?`,
					tenantID, line.ProductID, gi.WarehouseID, *line.VariantID)
			}
			if err := q.Row().Scan(&available); err != nil {
				return fmt.Errorf("failed to check stock for product %d: %w", line.ProductID, err)
			}
			if available < line.Quantity {
				return fmt.Errorf("insufficient stock for product %d: available %.4f, required %.4f", line.ProductID, available, line.Quantity)
			}
		}

		now := time.Now()
		gi.Status = GIStatusConfirmed
		gi.ConfirmedBy = &userID
		gi.ConfirmedAt = &now

		for i := range gi.Lines {
			gi.Lines[i].TotalCost = gi.Lines[i].Quantity * gi.Lines[i].UnitCost
			if err := tx.Save(&gi.Lines[i]).Error; err != nil {
				return err
			}
			if err := r.upsertStockBalance(tx, tenantID, gi.Lines[i].ProductID, gi.Lines[i].VariantID, gi.WarehouseID, gi.Lines[i].LocationID, -gi.Lines[i].Quantity); err != nil {
				return fmt.Errorf("failed to deduct stock on line %d: %w", gi.Lines[i].LineNumber, err)
			}
			ledgerRows = append(ledgerRows, ledgerRow{
				tenantID:    tenantID,
				productID:   gi.Lines[i].ProductID,
				variantID:   gi.Lines[i].VariantID,
				warehouseID: gi.WarehouseID,
				locationID:  gi.Lines[i].LocationID,
				txType:      "GOODS_ISSUE",
				refType:     "GOODS_ISSUE",
				refID:       gi.ID,
				qty:         -gi.Lines[i].Quantity,
				unitCost:    gi.Lines[i].UnitCost,
				totalCost:   gi.Lines[i].TotalCost,
				date:        gi.IssueDate,
				createdBy:   userID,
			})
		}

		return tx.Save(&gi).Error
	})
	if err != nil {
		return err
	}

	r.writeToLedger(ctx, ledgerRows)
	return nil
}

// ── Stock Adjustments ─────────────────────────────────────────────────────────

func (r *dbRepository) ListAdjustments(ctx context.Context, tenantID uint, status string) ([]StockAdjustment, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []StockAdjustment
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetAdjustment(ctx context.Context, tenantID, id uint) (*StockAdjustment, error) {
	var sa StockAdjustment
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&sa).Error
	return &sa, err
}

func (r *dbRepository) CreateAdjustment(ctx context.Context, sa *StockAdjustment) error {
	return r.db.WithContext(ctx).Create(sa).Error
}

func (r *dbRepository) UpdateAdjustment(ctx context.Context, sa *StockAdjustment) error {
	return r.db.WithContext(ctx).Save(sa).Error
}

func (r *dbRepository) ListAdjustmentLines(ctx context.Context, adjID uint) ([]SALine, error) {
	var rows []SALine
	return rows, r.db.WithContext(ctx).Where("adjustment_id = ?", adjID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) AddAdjustmentLine(ctx context.Context, line *SALine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) UpdateAdjustmentLine(ctx context.Context, line *SALine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) DeleteAdjustmentLine(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&SALine{}).Error
}

func (r *dbRepository) ConfirmAdjustment(ctx context.Context, tenantID, id, userID uint) error {
	var ledgerRows []ledgerRow

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sa StockAdjustment
		if err := tx.Preload("Lines").
			Where("tenant_id = ? AND id = ?", tenantID, id).First(&sa).Error; err != nil {
			return err
		}
		if sa.Status != SAStatusDraft {
			return fmt.Errorf("stock adjustment is already %s", sa.Status)
		}

		now := time.Now()
		sa.Status = SAStatusConfirmed
		sa.ConfirmedBy = &userID
		sa.ConfirmedAt = &now

		for _, line := range sa.Lines {
			if err := r.upsertStockBalance(tx, tenantID, line.ProductID, line.VariantID, sa.WarehouseID, line.LocationID, line.Quantity); err != nil {
				return fmt.Errorf("failed to adjust stock on line %d: %w", line.LineNumber, err)
			}
			totalCost := line.Quantity * line.UnitCost
			ledgerRows = append(ledgerRows, ledgerRow{
				tenantID:    tenantID,
				productID:   line.ProductID,
				variantID:   line.VariantID,
				warehouseID: sa.WarehouseID,
				locationID:  line.LocationID,
				txType:      "ADJUSTMENT",
				refType:     "STOCK_ADJUSTMENT",
				refID:       sa.ID,
				qty:         line.Quantity,
				unitCost:    line.UnitCost,
				totalCost:   totalCost,
				date:        sa.AdjustmentDate,
				createdBy:   userID,
			})
		}

		return tx.Save(&sa).Error
	})
	if err != nil {
		return err
	}

	r.writeToLedger(ctx, ledgerRows)
	return nil
}

// ── Quality Checks ────────────────────────────────────────────────────────────

func (r *dbRepository) ListQualityChecks(ctx context.Context, tenantID uint, status string) ([]QualityCheck, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []QualityCheck
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) ListQCsByRefs(ctx context.Context, tenantID uint, refType string, refIDs []uint) ([]QualityCheck, error) {
	if len(refIDs) == 0 {
		return nil, nil
	}
	var rows []QualityCheck
	return rows, r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND reference_type = ? AND reference_id IN ?", tenantID, refType, refIDs).
		Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetQualityCheck(ctx context.Context, tenantID, id uint) (*QualityCheck, error) {
	var qc QualityCheck
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&qc).Error
	return &qc, err
}

func (r *dbRepository) CreateQualityCheck(ctx context.Context, qc *QualityCheck) error {
	return r.db.WithContext(ctx).Create(qc).Error
}

func (r *dbRepository) UpdateQualityCheckStatus(ctx context.Context, tenantID, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&QualityCheck{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("status", status).Error
}

func (r *dbRepository) ListQCLines(ctx context.Context, checkID uint) ([]QCLine, error) {
	var rows []QCLine
	return rows, r.db.WithContext(ctx).Where("check_id = ?", checkID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) AddQCLine(ctx context.Context, line *QCLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) UpdateQCLine(ctx context.Context, line *QCLine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) SubmitQualityCheck(ctx context.Context, tenantID, id, userID uint) error {
	qc, err := r.GetQualityCheck(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if len(qc.Lines) == 0 {
		return fmt.Errorf("quality check has no lines")
	}

	allPassed := true
	allFailed := true
	for _, line := range qc.Lines {
		if line.Result != QCStatusPassed {
			allPassed = false
		}
		if line.Result != QCStatusFailed {
			allFailed = false
		}
	}

	var overallStatus string
	switch {
	case allPassed:
		overallStatus = QCStatusPassed
	case allFailed:
		overallStatus = QCStatusFailed
	default:
		overallStatus = QCStatusPartial
	}

	// Lookup location_id from upstream reference for accurate stock placement.
	// MATERIAL_QC → goods_receipt_lines.location_id (matched by product_id within the GRN).
	// PRODUCT_QC  → production_outputs.location_id (matched by output ID).
	var ledgerRows []ledgerRow
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Finalize the QC status first
		if err := tx.Model(&QualityCheck{}).
			Where("tenant_id = ? AND id = ?", tenantID, id).
			Update("status", overallStatus).Error; err != nil {
			return err
		}
		// Post qty_passed for each line to stock_balances + collect ledger rows.
		// Skip lines that didn't pass any qty.
		txType := "QC_RELEASE"
		refType := qc.ReferenceType
		var refID uint
		if qc.ReferenceID != nil {
			refID = *qc.ReferenceID
		}
		// Build a product_id → location_id map from the upstream reference for placement.
		locByProduct := r.locationsForQCSource(tx, qc)
		for _, line := range qc.Lines {
			if line.QtyPassed <= 0 {
				continue
			}
			var locID *uint
			if l, ok := locByProduct[line.ProductID]; ok {
				locID = l
			}
			if err := r.upsertStockBalance(tx, tenantID, line.ProductID, line.VariantID, qc.WarehouseID, locID, line.QtyPassed); err != nil {
				return fmt.Errorf("failed to post stock for QC line %d: %w", line.ID, err)
			}
			ledgerRows = append(ledgerRows, ledgerRow{
				tenantID: tenantID, productID: line.ProductID, variantID: line.VariantID,
				warehouseID: qc.WarehouseID, locationID: locID,
				txType: txType, refType: refType, refID: refID,
				qty: line.QtyPassed, date: qc.CheckDate, createdBy: userID,
			})
		}
		return nil
	})
	if err != nil {
		return err
	}
	r.writeToLedger(ctx, ledgerRows)
	return nil
}

// locationsForQCSource resolves preferred storage location for each product on the
// upstream document (GRN or Production Output) so QC pass posts stock to the right bin.
// Returns an empty map if the reference type is unrecognised.
func (r *dbRepository) locationsForQCSource(tx *gorm.DB, qc *QualityCheck) map[uint]*uint {
	out := make(map[uint]*uint)
	if qc.ReferenceID == nil {
		return out
	}
	switch qc.ReferenceType {
	case "GRN":
		var rows []struct {
			ProductID  uint
			LocationID *uint
		}
		tx.Raw(`SELECT product_id, location_id FROM goods_receipt_lines WHERE grn_id = ?`, *qc.ReferenceID).Scan(&rows)
		for _, r := range rows {
			out[r.ProductID] = r.LocationID
		}
	case "PRODUCTION_OUTPUT":
		var row struct {
			ProductID  uint
			LocationID *uint
		}
		tx.Raw(`SELECT product_id, location_id FROM production_outputs WHERE id = ?`, *qc.ReferenceID).Scan(&row)
		if row.ProductID != 0 {
			out[row.ProductID] = row.LocationID
		}
	}
	return out
}

// ── Stock Balances ────────────────────────────────────────────────────────────

func (r *dbRepository) GetStockBalance(ctx context.Context, tenantID uint, warehouseID *uint, productID *uint) ([]StockBalanceRow, error) {
	q := r.db.WithContext(ctx).Table("stock_balances").
		Select("product_id, variant_id, warehouse_id, location_id, quantity, COALESCE(reserved_qty, 0) AS reserved_qty").
		Where("tenant_id = ?", tenantID)
	if warehouseID != nil {
		q = q.Where("warehouse_id = ?", *warehouseID)
	}
	if productID != nil {
		q = q.Where("product_id = ?", *productID)
	}
	var rows []StockBalanceRow
	return rows, q.Find(&rows).Error
}
