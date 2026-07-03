package procurement

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

type stockLedgerRow struct {
	productID, variantID, warehouseID, locationID uint64
	qty, unitCost, total                          float64
	date                                          string
}

type Repository interface {
	// Document sequence
	NextCode(ctx context.Context, tenantID uint, docType string) (string, error)

	// Purchase Requests
	ListPRs(ctx context.Context, tenantID uint, status string) ([]PurchaseRequest, error)
	GetPR(ctx context.Context, tenantID, id uint) (*PurchaseRequest, error)
	CreatePR(ctx context.Context, pr *PurchaseRequest) error
	UpdatePR(ctx context.Context, pr *PurchaseRequest) error
	DeletePR(ctx context.Context, tenantID, id uint) error

	// PR Items
	ListPRItems(ctx context.Context, prID uint) ([]PRLine, error)
	GetPRItem(ctx context.Context, tenantID, id uint) (*PRLine, error)
	AddPRItem(ctx context.Context, line *PRLine) error
	UpdatePRItem(ctx context.Context, line *PRLine) error
	DeletePRItem(ctx context.Context, tenantID, id uint) error

	// Purchase Orders
	ListPOs(ctx context.Context, tenantID uint, status string, supplierID *uint) ([]PurchaseOrder, error)
	GetPO(ctx context.Context, tenantID, id uint) (*PurchaseOrder, error)
	CreatePO(ctx context.Context, po *PurchaseOrder) error
	UpdatePO(ctx context.Context, po *PurchaseOrder) error
	DeletePO(ctx context.Context, tenantID, id uint) error

	// PO Items
	ListPOItems(ctx context.Context, poID uint) ([]POLine, error)
	GetPOItem(ctx context.Context, tenantID, id uint) (*POLine, error)
	AddPOItem(ctx context.Context, line *POLine) error
	UpdatePOItem(ctx context.Context, line *POLine) error
	DeletePOItem(ctx context.Context, tenantID, id uint) error

	// Goods Receipts
	ListGRNs(ctx context.Context, tenantID uint, status string, poID *uint) ([]GoodsReceipt, error)
	ListGRNsByMO(ctx context.Context, tenantID, moID uint) ([]GoodsReceipt, error)
	GetGRN(ctx context.Context, tenantID, id uint) (*GoodsReceipt, error)
	CreateGRN(ctx context.Context, grn *GoodsReceipt) error
	UpdateGRN(ctx context.Context, grn *GoodsReceipt) error
	ConfirmGRN(ctx context.Context, tenantID, grnID uint, confirmedBy uint) error

	// GRN Items
	ListGRNItems(ctx context.Context, grnID uint) ([]GRNLine, error)
	GetGRNItem(ctx context.Context, tenantID, id uint) (*GRNLine, error)
	AddGRNItem(ctx context.Context, line *GRNLine) error
	UpdateGRNItem(ctx context.Context, line *GRNLine) error

	// Purchase Invoices
	ListInvoices(ctx context.Context, tenantID uint, status string, supplierID *uint) ([]PurchaseInvoice, error)
	GetInvoice(ctx context.Context, tenantID, id uint) (*PurchaseInvoice, error)
	GetDraftInvoiceByPOID(ctx context.Context, tenantID, poID uint) (*PurchaseInvoice, error)
	POInvoiceIsLocked(ctx context.Context, tenantID, poID uint) (bool, error)
	CreateInvoice(ctx context.Context, inv *PurchaseInvoice) error
	UpdateInvoice(ctx context.Context, inv *PurchaseInvoice) error
	AddInvoiceLine(ctx context.Context, line *InvoiceLine) error
	DeleteInvoiceLine(ctx context.Context, tenantID, id uint) error
}

type dbRepository struct {
	db       *gorm.DB
	ledgerDB *sql.DB
}

func NewRepository(db *gorm.DB, ledgerDB *sql.DB) Repository {
	return &dbRepository{db: db, ledgerDB: ledgerDB}
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

// ── Purchase Requests ─────────────────────────────────────────────────────────

func (r *dbRepository) ListPRs(ctx context.Context, tenantID uint, status string) ([]PurchaseRequest, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []PurchaseRequest
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetPR(ctx context.Context, tenantID, id uint) (*PurchaseRequest, error) {
	var pr PurchaseRequest
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&pr).Error
	return &pr, err
}

func (r *dbRepository) CreatePR(ctx context.Context, pr *PurchaseRequest) error {
	return r.db.WithContext(ctx).Create(pr).Error
}

func (r *dbRepository) UpdatePR(ctx context.Context, pr *PurchaseRequest) error {
	return r.db.WithContext(ctx).Save(pr).Error
}

func (r *dbRepository) DeletePR(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&PurchaseRequest{}).Error
}

// ── PR Items ──────────────────────────────────────────────────────────────────

func (r *dbRepository) ListPRItems(ctx context.Context, prID uint) ([]PRLine, error) {
	var rows []PRLine
	return rows, r.db.WithContext(ctx).Where("pr_id = ?", prID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) GetPRItem(ctx context.Context, tenantID, id uint) (*PRLine, error) {
	var line PRLine
	return &line, r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&line).Error
}

func (r *dbRepository) AddPRItem(ctx context.Context, line *PRLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) UpdatePRItem(ctx context.Context, line *PRLine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) DeletePRItem(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&PRLine{}).Error
}

// ── Purchase Orders ───────────────────────────────────────────────────────────

func (r *dbRepository) ListPOs(ctx context.Context, tenantID uint, status string, supplierID *uint) ([]PurchaseOrder, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if supplierID != nil {
		q = q.Where("supplier_id = ?", *supplierID)
	}
	var rows []PurchaseOrder
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetPO(ctx context.Context, tenantID, id uint) (*PurchaseOrder, error) {
	var po PurchaseOrder
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&po).Error
	return &po, err
}

func (r *dbRepository) CreatePO(ctx context.Context, po *PurchaseOrder) error {
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *dbRepository) UpdatePO(ctx context.Context, po *PurchaseOrder) error {
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *dbRepository) DeletePO(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&PurchaseOrder{}).Error
}

// ── PO Items ──────────────────────────────────────────────────────────────────

func (r *dbRepository) ListPOItems(ctx context.Context, poID uint) ([]POLine, error) {
	var rows []POLine
	return rows, r.db.WithContext(ctx).Where("po_id = ?", poID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) GetPOItem(ctx context.Context, tenantID, id uint) (*POLine, error) {
	var line POLine
	return &line, r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&line).Error
}

func (r *dbRepository) AddPOItem(ctx context.Context, line *POLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) UpdatePOItem(ctx context.Context, line *POLine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

func (r *dbRepository) DeletePOItem(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&POLine{}).Error
}

// ── Goods Receipts ────────────────────────────────────────────────────────────

func (r *dbRepository) ListGRNs(ctx context.Context, tenantID uint, status string, poID *uint) ([]GoodsReceipt, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if poID != nil {
		q = q.Where("po_id = ?", *poID)
	}
	var rows []GoodsReceipt
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) ListGRNsByMO(ctx context.Context, tenantID, moID uint) ([]GoodsReceipt, error) {
	var rows []GoodsReceipt
	return rows, r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND mo_id = ?", tenantID, moID).
		Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetGRN(ctx context.Context, tenantID, id uint) (*GoodsReceipt, error) {
	var grn GoodsReceipt
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&grn).Error
	return &grn, err
}

func (r *dbRepository) CreateGRN(ctx context.Context, grn *GoodsReceipt) error {
	return r.db.WithContext(ctx).Create(grn).Error
}

func (r *dbRepository) UpdateGRN(ctx context.Context, grn *GoodsReceipt) error {
	return r.db.WithContext(ctx).Save(grn).Error
}

// ConfirmGRN confirms the GRN and updates PO received_qty atomically in PostgreSQL,
// then writes stock ledger rows to ClickHouse (best-effort, does not roll back the GRN).
func (r *dbRepository) ConfirmGRN(ctx context.Context, tenantID, grnID uint, confirmedBy uint) error {
	var ledgerRows []stockLedgerRow

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var grn GoodsReceipt
		if err := tx.Preload("Lines").
			Where("tenant_id = ? AND id = ?", tenantID, grnID).First(&grn).Error; err != nil {
			return err
		}
		if grn.Status != GRNStatusDraft {
			return fmt.Errorf("GRN is already %s", grn.Status)
		}
		now := time.Now()
		grn.Status = GRNStatusConfirmed
		grn.ConfirmedBy = &confirmedBy
		grn.ConfirmedAt = &now
		if err := tx.Save(&grn).Error; err != nil {
			return err
		}
		// Gate stock posting on grn_type: WITH_PO and WITHOUT_PO must pass QC first.
		// CUSTOMER_RETURN and PRODUCTION_RETURN post immediately (no QC).
		postStockNow := grn.GRNType != GRNTypeWithPO && grn.GRNType != GRNTypeWithoutPO
		for _, line := range grn.Lines {
			ratio := line.TransferRatio
			if ratio <= 0 {
				ratio = 1
			}
			stockQty := line.Quantity * ratio
			stockUnitCost := line.UnitCost / ratio
			stockTotal := math.Round(stockQty*stockUnitCost*100) / 100

			var variantID, locationID uint64
			if line.VariantID != nil {
				variantID = uint64(*line.VariantID)
			}
			if line.LocationID != nil {
				locationID = uint64(*line.LocationID)
			}
			if postStockNow {
				ledgerRows = append(ledgerRows, stockLedgerRow{
					productID:   uint64(line.ProductID),
					variantID:   variantID,
					warehouseID: uint64(grn.WarehouseID),
					locationID:  locationID,
					qty:         stockQty,
					unitCost:    stockUnitCost,
					total:       stockTotal,
					date:        grn.ReceiptDate,
				})

				if err := tx.Exec(`
					INSERT INTO stock_balances
						(tenant_id, product_id, variant_id, warehouse_id, location_id, quantity, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, NOW())
					ON CONFLICT ON CONSTRAINT uidx_stock_balances
					DO UPDATE SET quantity = stock_balances.quantity + EXCLUDED.quantity, updated_at = NOW()`,
					tenantID, line.ProductID, line.VariantID, grn.WarehouseID, line.LocationID, stockQty,
				).Error; err != nil {
					return err
				}
			}

			// Always update PO line received_qty — physical receipt has happened
			// even though stock is held in QC limbo until passed.
			if line.POLineID != nil {
				tx.Exec(`UPDATE purchase_order_lines SET received_qty = received_qty + ? WHERE id = ?`,
					line.Quantity, *line.POLineID)
			}
		}
		if grn.POID != nil {
			tx.Exec(`
				UPDATE purchase_orders SET status =
					CASE WHEN (
						SELECT SUM(received_qty) >= SUM(quantity)
						FROM purchase_order_lines WHERE po_id = ?
					) THEN 'RECEIVED' ELSE 'PARTIAL' END
				WHERE id = ? AND status IN ('CONFIRMED','PARTIAL')`,
				*grn.POID, *grn.POID)
		}
		return nil
	})
	if err != nil {
		return err
	}

	r.writeStockLedger(ctx, tenantID, grnID, confirmedBy, ledgerRows)
	return nil
}

func (r *dbRepository) writeStockLedger(ctx context.Context, tenantID, grnID, confirmedBy uint, rows []stockLedgerRow) {
	if r.ledgerDB == nil || len(rows) == 0 {
		return
	}
	for _, row := range rows {
		_, _ = r.ledgerDB.ExecContext(ctx, `
			INSERT INTO stock_ledger
				(tenant_id, product_id, variant_id, warehouse_id, location_id,
				 transaction_type, reference_type, reference_id,
				 quantity, unit_cost, total_cost, transaction_date, created_by)
			VALUES (?, ?, ?, ?, ?, 'PURCHASE', 'GOODS_RECEIPT', ?, ?, ?, ?, ?, ?)`,
			uint64(tenantID), row.productID, row.variantID,
			row.warehouseID, row.locationID,
			uint64(grnID), row.qty, row.unitCost, row.total, row.date, uint64(confirmedBy),
		)
	}
}

// ── GRN Items ─────────────────────────────────────────────────────────────────

func (r *dbRepository) ListGRNItems(ctx context.Context, grnID uint) ([]GRNLine, error) {
	var rows []GRNLine
	return rows, r.db.WithContext(ctx).Where("grn_id = ?", grnID).Order("line_number").Find(&rows).Error
}

func (r *dbRepository) GetGRNItem(ctx context.Context, tenantID, id uint) (*GRNLine, error) {
	var line GRNLine
	return &line, r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&line).Error
}

func (r *dbRepository) AddGRNItem(ctx context.Context, line *GRNLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) UpdateGRNItem(ctx context.Context, line *GRNLine) error {
	return r.db.WithContext(ctx).Save(line).Error
}

// ── Purchase Invoices ─────────────────────────────────────────────────────────

func (r *dbRepository) ListInvoices(ctx context.Context, tenantID uint, status string, supplierID *uint) ([]PurchaseInvoice, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if supplierID != nil {
		q = q.Where("supplier_id = ?", *supplierID)
	}
	var rows []PurchaseInvoice
	return rows, q.Order("created_at DESC").Find(&rows).Error
}

func (r *dbRepository) GetInvoice(ctx context.Context, tenantID, id uint) (*PurchaseInvoice, error) {
	var inv PurchaseInvoice
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND id = ?", tenantID, id).First(&inv).Error
	return &inv, err
}

// GetDraftInvoiceByPOID returns the most recent DRAFT invoice for a PO.
func (r *dbRepository) GetDraftInvoiceByPOID(ctx context.Context, tenantID, poID uint) (*PurchaseInvoice, error) {
	var inv PurchaseInvoice
	err := r.db.WithContext(ctx).Preload("Lines").
		Where("tenant_id = ? AND po_id = ? AND status = ?", tenantID, poID, InvStatusDraft).
		Order("created_at DESC").First(&inv).Error
	return &inv, err
}

// POInvoiceIsLocked returns true when the PO already has a posted/paid/partial invoice
// (i.e. no longer in DRAFT or CANCELLED) — meaning no more GRNs can be linked.
func (r *dbRepository) POInvoiceIsLocked(ctx context.Context, tenantID, poID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&PurchaseInvoice{}).
		Where("tenant_id = ? AND po_id = ? AND status NOT IN (?, ?)",
			tenantID, poID, InvStatusDraft, InvStatusCancelled).
		Count(&count).Error
	return count > 0, err
}

func (r *dbRepository) CreateInvoice(ctx context.Context, inv *PurchaseInvoice) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *dbRepository) UpdateInvoice(ctx context.Context, inv *PurchaseInvoice) error {
	return r.db.WithContext(ctx).Save(inv).Error
}

func (r *dbRepository) AddInvoiceLine(ctx context.Context, line *InvoiceLine) error {
	return r.db.WithContext(ctx).Create(line).Error
}

func (r *dbRepository) DeleteInvoiceLine(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&InvoiceLine{}).Error
}
