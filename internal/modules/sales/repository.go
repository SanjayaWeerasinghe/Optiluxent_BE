package sales

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	// Sales Quotations
	NextSQCode(ctx context.Context, tenantID uint) (string, error)
	CreateSQ(ctx context.Context, sq *SalesQuotation) error
	ListSQs(ctx context.Context, tenantID uint, customerID *uint, status string, limit, offset int) ([]SalesQuotation, error)
	CountSQs(ctx context.Context, tenantID uint, customerID *uint, status string) (int64, error)
	GetSQ(ctx context.Context, tenantID, id uint) (*SalesQuotation, error)
	UpdateSQ(ctx context.Context, sq *SalesQuotation) error
	DeleteSQ(ctx context.Context, tenantID, id uint) error
	AddSQLine(ctx context.Context, line *SQLine) error
	GetSQLine(ctx context.Context, tenantID, sqID, lineID uint) (*SQLine, error)
	ListSQLines(ctx context.Context, tenantID, sqID uint) ([]SQLine, error)
	UpdateSQLine(ctx context.Context, line *SQLine) error
	DeleteSQLine(ctx context.Context, tenantID, sqID, lineID uint) error

	// Sales Orders
	NextSOCode(ctx context.Context, tenantID uint) (string, error)
	CreateSO(ctx context.Context, so *SalesOrder) error
	ListSOs(ctx context.Context, tenantID uint, customerID *uint, status string, limit, offset int) ([]SalesOrder, error)
	CountSOs(ctx context.Context, tenantID uint, customerID *uint, status string) (int64, error)
	GetSO(ctx context.Context, tenantID, id uint) (*SalesOrder, error)
	UpdateSO(ctx context.Context, so *SalesOrder) error
	DeleteSO(ctx context.Context, tenantID, id uint) error
	AddSOLine(ctx context.Context, line *SOLine) error
	GetSOLine(ctx context.Context, tenantID, soID, lineID uint) (*SOLine, error)
	ListSOLines(ctx context.Context, tenantID, soID uint) ([]SOLine, error)
	UpdateSOLine(ctx context.Context, line *SOLine) error
	DeleteSOLine(ctx context.Context, tenantID, soID, lineID uint) error

	// Delivery Orders
	NextDOCode(ctx context.Context, tenantID uint) (string, error)
	CreateDO(ctx context.Context, do *DeliveryOrder) error
	ListDOs(ctx context.Context, tenantID uint, soID *uint, status string, limit, offset int) ([]DeliveryOrder, error)
	CountDOs(ctx context.Context, tenantID uint, soID *uint, status string) (int64, error)
	GetDO(ctx context.Context, tenantID, id uint) (*DeliveryOrder, error)
	UpdateDO(ctx context.Context, do *DeliveryOrder) error
	AddDOLine(ctx context.Context, line *DOLine) error
	GetDOLine(ctx context.Context, tenantID, doID, lineID uint) (*DOLine, error)
	ListDOLines(ctx context.Context, tenantID, doID uint) ([]DOLine, error)
	UpdateDOLine(ctx context.Context, line *DOLine) error
	DeleteDOLine(ctx context.Context, tenantID, doID, lineID uint) error
	ConfirmDO(ctx context.Context, tenantID, doID, userID uint, now time.Time) error
	SetDOStatus(ctx context.Context, tenantID, id uint, status string) error

	// Sales Invoices
	NextSICode(ctx context.Context, tenantID uint) (string, error)
	CreateSI(ctx context.Context, si *SalesInvoice) error
	ListSIs(ctx context.Context, tenantID uint, customerID *uint, status string, limit, offset int) ([]SalesInvoice, error)
	CountSIs(ctx context.Context, tenantID uint, customerID *uint, status string) (int64, error)
	GetSI(ctx context.Context, tenantID, id uint) (*SalesInvoice, error)
	GetDraftInvoiceBySOID(ctx context.Context, tenantID, soID uint) (*SalesInvoice, error)
	UpdateSI(ctx context.Context, si *SalesInvoice) error
	AddSILine(ctx context.Context, line *SILine) error
	GetSILine(ctx context.Context, tenantID, invoiceID, lineID uint) (*SILine, error)
	ListSILines(ctx context.Context, tenantID, invoiceID uint) ([]SILine, error)
	DeleteSILine(ctx context.Context, tenantID, invoiceID, lineID uint) error
}

type dbRepository struct {
	db       *gorm.DB
	ledgerDB *sql.DB
}

func NewRepository(db *gorm.DB, ledgerDB *sql.DB) Repository {
	return &dbRepository{db: db, ledgerDB: ledgerDB}
}

// nextCode fetches the next formatted document code and increments the counter atomically.
func (r *dbRepository) nextCode(ctx context.Context, tenantID uint, docType string) (string, error) {
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

func (r *dbRepository) NextSOCode(ctx context.Context, tenantID uint) (string, error) {
	return r.nextCode(ctx, tenantID, "SALES_ORDER")
}

func (r *dbRepository) NextDOCode(ctx context.Context, tenantID uint) (string, error) {
	return r.nextCode(ctx, tenantID, "DELIVERY_ORDER")
}

func (r *dbRepository) NextSICode(ctx context.Context, tenantID uint) (string, error) {
	return r.nextCode(ctx, tenantID, "SALES_INVOICE")
}

func (r *dbRepository) NextSQCode(ctx context.Context, tenantID uint) (string, error) {
	return r.nextCode(ctx, tenantID, "SALES_QUOTATION")
}

// ── Sales Quotations ──────────────────────────────────────────────────────────

func (r *dbRepository) CreateSQ(ctx context.Context, sq *SalesQuotation) error {
	return r.db.WithContext(ctx).Create(sq).Error
}

func (r *dbRepository) ListSQs(ctx context.Context, tenantID uint, customerID *uint, status string, limit, offset int) ([]SalesQuotation, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q = q.Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []SalesQuotation
	return rows, q.Find(&rows).Error
}

func (r *dbRepository) CountSQs(ctx context.Context, tenantID uint, customerID *uint, status string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&SalesQuotation{}).Where("tenant_id = ?", tenantID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var n int64
	return n, q.Count(&n).Error
}

func (r *dbRepository) GetSQ(ctx context.Context, tenantID, id uint) (*SalesQuotation, error) {
	var sq SalesQuotation
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&sq).Error
	return &sq, err
}

func (r *dbRepository) UpdateSQ(ctx context.Context, sq *SalesQuotation) error {
	return r.db.WithContext(ctx).Save(sq).Error
}

func (r *dbRepository) DeleteSQ(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&SalesQuotation{}).Error
}

// ── SQ Lines ──────────────────────────────────────────────────────────────────

func (r *dbRepository) AddSQLine(ctx context.Context, line *SQLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) GetSQLine(ctx context.Context, tenantID, sqID, lineID uint) (*SQLine, error) {
	var line SQLine
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND sq_id = ? AND id = ?", tenantID, sqID, lineID).
		First(&line).Error
	return &line, err
}

func (r *dbRepository) ListSQLines(ctx context.Context, tenantID, sqID uint) ([]SQLine, error) {
	var rows []SQLine
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND sq_id = ?", tenantID, sqID).
		Order("line_number").Find(&rows).Error
}

func (r *dbRepository) UpdateSQLine(ctx context.Context, line *SQLine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) DeleteSQLine(ctx context.Context, tenantID, sqID, lineID uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND sq_id = ? AND id = ?", tenantID, sqID, lineID).
		Delete(&SQLine{}).Error
}

// ── Sales Orders ──────────────────────────────────────────────────────────────

func (r *dbRepository) CreateSO(ctx context.Context, so *SalesOrder) error {
	return r.db.WithContext(ctx).Create(so).Error
}

func (r *dbRepository) ListSOs(ctx context.Context, tenantID uint, customerID *uint, status string, limit, offset int) ([]SalesOrder, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q = q.Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []SalesOrder
	return rows, q.Find(&rows).Error
}

func (r *dbRepository) CountSOs(ctx context.Context, tenantID uint, customerID *uint, status string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&SalesOrder{}).Where("tenant_id = ?", tenantID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var n int64
	return n, q.Count(&n).Error
}

func (r *dbRepository) GetSO(ctx context.Context, tenantID, id uint) (*SalesOrder, error) {
	var so SalesOrder
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&so).Error
	return &so, err
}

func (r *dbRepository) UpdateSO(ctx context.Context, so *SalesOrder) error {
	return r.db.WithContext(ctx).Save(so).Error
}

func (r *dbRepository) DeleteSO(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&SalesOrder{}).Error
}

// ── SO Lines ──────────────────────────────────────────────────────────────────

func (r *dbRepository) AddSOLine(ctx context.Context, line *SOLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) GetSOLine(ctx context.Context, tenantID, soID, lineID uint) (*SOLine, error) {
	var line SOLine
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND so_id = ? AND id = ?", tenantID, soID, lineID).
		First(&line).Error
	return &line, err
}

func (r *dbRepository) ListSOLines(ctx context.Context, tenantID, soID uint) ([]SOLine, error) {
	var rows []SOLine
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND so_id = ?", tenantID, soID).
		Order("line_number").Find(&rows).Error
}

func (r *dbRepository) UpdateSOLine(ctx context.Context, line *SOLine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) DeleteSOLine(ctx context.Context, tenantID, soID, lineID uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND so_id = ? AND id = ?", tenantID, soID, lineID).
		Delete(&SOLine{}).Error
}

// ── Delivery Orders ───────────────────────────────────────────────────────────

func (r *dbRepository) CreateDO(ctx context.Context, do *DeliveryOrder) error {
	return r.db.WithContext(ctx).Create(do).Error
}

func (r *dbRepository) ListDOs(ctx context.Context, tenantID uint, soID *uint, status string, limit, offset int) ([]DeliveryOrder, error) {
	// Preload lines — the SI extras panel rolls up delivered qty per product
	// across every DO linked to an SO. Same pattern as procurement.ListGRNs.
	q := r.db.WithContext(ctx).Preload("Lines").Where("tenant_id = ?", tenantID)
	if soID != nil {
		q = q.Where("so_id = ?", *soID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q = q.Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []DeliveryOrder
	return rows, q.Find(&rows).Error
}

func (r *dbRepository) CountDOs(ctx context.Context, tenantID uint, soID *uint, status string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&DeliveryOrder{}).Where("tenant_id = ?", tenantID)
	if soID != nil {
		q = q.Where("so_id = ?", *soID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var n int64
	return n, q.Count(&n).Error
}

func (r *dbRepository) GetDO(ctx context.Context, tenantID, id uint) (*DeliveryOrder, error) {
	var do DeliveryOrder
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&do).Error
	return &do, err
}

func (r *dbRepository) UpdateDO(ctx context.Context, do *DeliveryOrder) error {
	return r.db.WithContext(ctx).Save(do).Error
}

// SetDOStatus is a plain status transition used by Cancel — no side-effects.
func (r *dbRepository) SetDOStatus(ctx context.Context, tenantID, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&DeliveryOrder{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("status", status).Error
}

// ── DO Lines ──────────────────────────────────────────────────────────────────

func (r *dbRepository) AddDOLine(ctx context.Context, line *DOLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) GetDOLine(ctx context.Context, tenantID, doID, lineID uint) (*DOLine, error) {
	var line DOLine
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND do_id = ? AND id = ?", tenantID, doID, lineID).
		First(&line).Error
	return &line, err
}

func (r *dbRepository) ListDOLines(ctx context.Context, tenantID, doID uint) ([]DOLine, error) {
	var rows []DOLine
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND do_id = ?", tenantID, doID).
		Order("line_number").Find(&rows).Error
}

func (r *dbRepository) UpdateDOLine(ctx context.Context, line *DOLine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) DeleteDOLine(ctx context.Context, tenantID, doID, lineID uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND do_id = ? AND id = ?", tenantID, doID, lineID).
		Delete(&DOLine{}).Error
}

// ConfirmDO confirms the delivery order, decrements stock_balances, updates SO delivered_qty,
// updates SO status, and writes stock ledger rows to ClickHouse (best-effort).
func (r *dbRepository) ConfirmDO(ctx context.Context, tenantID, doID, userID uint, now time.Time) error {
	type ledgerRow struct {
		productID, variantID, warehouseID, locationID uint64
		qty, unitCost, total                          float64
		date                                          string
	}
	var ledgerRows []ledgerRow

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var do DeliveryOrder
		if err := tx.Preload("Lines").
			Where("tenant_id = ? AND id = ?", tenantID, doID).First(&do).Error; err != nil {
			return err
		}
		if do.Status != DOStatusDraft {
			return fmt.Errorf("delivery order is already %s", do.Status)
		}

		do.Status = DOStatusConfirmed
		do.ConfirmedBy = &userID
		do.ConfirmedAt = &now
		if err := tx.Save(&do).Error; err != nil {
			return err
		}

		for _, line := range do.Lines {
			var variantID, locationID uint64
			if line.VariantID != nil {
				variantID = uint64(*line.VariantID)
			}
			if line.LocationID != nil {
				locationID = uint64(*line.LocationID)
			}

			ledgerRows = append(ledgerRows, ledgerRow{
				productID:   uint64(line.ProductID),
				variantID:   variantID,
				warehouseID: uint64(do.WarehouseID),
				locationID:  locationID,
				qty:         line.Quantity,
				unitCost:    line.UnitCost,
				total:       line.TotalCost,
				date:        do.DeliveryDate,
			})

			// Decrement stock_balances (negative quantity = outgoing)
			if err := tx.Exec(`
				INSERT INTO stock_balances
					(tenant_id, product_id, variant_id, warehouse_id, location_id, quantity, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, NOW())
				ON CONFLICT ON CONSTRAINT uidx_stock_balances
				DO UPDATE SET quantity = stock_balances.quantity - EXCLUDED.quantity, updated_at = NOW()`,
				tenantID, line.ProductID, line.VariantID, do.WarehouseID, line.LocationID, line.Quantity,
			).Error; err != nil {
				return err
			}

			// Update delivered_qty on the matching SO line
			if line.SOLineID != nil {
				tx.Exec(`UPDATE sales_order_lines SET delivered_qty = delivered_qty + ? WHERE id = ?`,
					line.Quantity, *line.SOLineID)
			}
		}

		// Update SO status if linked
		if do.SOID != nil {
			tx.Exec(`
				UPDATE sales_orders SET status =
					CASE WHEN (
						SELECT SUM(delivered_qty) >= SUM(quantity)
						FROM sales_order_lines WHERE so_id = ?
					) THEN 'DELIVERED' ELSE 'PARTIAL' END
				WHERE id = ? AND status IN ('CONFIRMED','PARTIAL')`,
				*do.SOID, *do.SOID)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Write to stock ledger (best-effort, does not roll back the DO confirmation)
	if r.ledgerDB != nil && len(ledgerRows) > 0 {
		for _, row := range ledgerRows {
			_, _ = r.ledgerDB.ExecContext(ctx, `
				INSERT INTO stock_ledger
					(tenant_id, product_id, variant_id, warehouse_id, location_id,
					 transaction_type, reference_type, reference_id,
					 quantity, unit_cost, total_cost, transaction_date, created_by)
				VALUES (?, ?, ?, ?, ?, 'SALE', 'DELIVERY_ORDER', ?, ?, ?, ?, ?, ?)`,
				uint64(tenantID), row.productID, row.variantID,
				row.warehouseID, row.locationID,
				uint64(doID), -row.qty, row.unitCost, row.total, row.date, uint64(userID),
			)
		}
	}

	return nil
}

// ── Sales Invoices ────────────────────────────────────────────────────────────

func (r *dbRepository) CreateSI(ctx context.Context, si *SalesInvoice) error {
	return r.db.WithContext(ctx).Create(si).Error
}

func (r *dbRepository) ListSIs(ctx context.Context, tenantID uint, customerID *uint, status string, limit, offset int) ([]SalesInvoice, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q = q.Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []SalesInvoice
	return rows, q.Find(&rows).Error
}

func (r *dbRepository) CountSIs(ctx context.Context, tenantID uint, customerID *uint, status string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&SalesInvoice{}).Where("tenant_id = ?", tenantID)
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var n int64
	return n, q.Count(&n).Error
}

func (r *dbRepository) GetSI(ctx context.Context, tenantID, id uint) (*SalesInvoice, error) {
	var si SalesInvoice
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&si).Error
	return &si, err
}

// GetDraftInvoiceBySOID returns the most recent DRAFT invoice for an SO.
func (r *dbRepository) GetDraftInvoiceBySOID(ctx context.Context, tenantID, soID uint) (*SalesInvoice, error) {
	var si SalesInvoice
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND so_id = ? AND status = ?", tenantID, soID, SIStatusDraft).
		Order("created_at DESC").First(&si).Error
	return &si, err
}

func (r *dbRepository) UpdateSI(ctx context.Context, si *SalesInvoice) error {
	return r.db.WithContext(ctx).Save(si).Error
}

// ── SI Lines ──────────────────────────────────────────────────────────────────

func (r *dbRepository) AddSILine(ctx context.Context, line *SILine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) GetSILine(ctx context.Context, tenantID, invoiceID, lineID uint) (*SILine, error) {
	var line SILine
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND invoice_id = ? AND id = ?", tenantID, invoiceID, lineID).
		First(&line).Error
	return &line, err
}

func (r *dbRepository) ListSILines(ctx context.Context, tenantID, invoiceID uint) ([]SILine, error) {
	var rows []SILine
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND invoice_id = ?", tenantID, invoiceID).
		Order("line_number").Find(&rows).Error
}

func (r *dbRepository) DeleteSILine(ctx context.Context, tenantID, invoiceID, lineID uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND invoice_id = ? AND id = ?", tenantID, invoiceID, lineID).
		Delete(&SILine{}).Error
}
