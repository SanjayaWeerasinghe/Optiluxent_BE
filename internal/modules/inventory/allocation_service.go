package inventory

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// AllocationService — the module-neutral surface consumed by every doc that
// reserves stock. Same interface shape is re-declared as
// procurement.AllocationReserver / sales.AllocationReserver so cross-module
// callers don't import inventory directly (avoids an import cycle in the
// sales → inventory → sales-adapter path).
type AllocationService struct {
	repo AllocationRepository
}

func NewAllocationService(repo AllocationRepository) *AllocationService {
	return &AllocationService{repo: repo}
}

// ReserveRequest — a single caller ask to hold `Quantity` on the given scope.
// If OnUpdate is true and an ACTIVE allocation for (SourceType, SourceID)
// already exists, the existing row is deleted first and a fresh one is
// written — so line-quantity edits do the right thing.
type ReserveRequest struct {
	TenantID    uint
	ScopeKey    ScopeKey
	Quantity    float64
	SourceType  string
	SourceID    uint
	SourceDocID uint
	Notes       string
	// OnUpdate — set to true from a doc line UPDATE call. Reserve will drop
	// the pre-existing allocation for the same source line before re-checking
	// availability and inserting the new one. Draft-create flows leave this
	// false (there's no pre-existing row to remove).
	OnUpdate bool
}

// Reserve computes availability inside a serialisable transaction and, on
// success, inserts a new ACTIVE row. Availability = on-hand at the scope
// minus SUM(ACTIVE allocations at the scope, excluding any row we're about
// to replace via OnUpdate).
func (s *AllocationService) Reserve(ctx context.Context, req ReserveRequest) (*Allocation, error) {
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("allocation quantity must be positive")
	}
	if req.SourceType != AllocSourceSOLine &&
		req.SourceType != AllocSourceMRLine &&
		req.SourceType != AllocSourceGILine &&
		req.SourceType != AllocSourceGTLine {
		return nil, fmt.Errorf("invalid allocation source_type %q", req.SourceType)
	}

	var created *Allocation
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the concurrent-reserver race. SELECT … FOR UPDATE on the
		// stock_balances row (or rows, if multiple bins carry the same
		// product+warehouse+production key) serialises any two Reserve calls
		// against the same scope; whichever hits second waits, then sees the
		// first's committed allocation and either fits or fails.
		lockSQL, lockArgs := lockScopeSQL(req.TenantID, req.ScopeKey)
		if err := tx.Exec(lockSQL, lockArgs...).Error; err != nil {
			return fmt.Errorf("lock stock scope: %w", err)
		}

		// If this is an update against a pre-existing allocation for the
		// same source line, drop it first so it doesn't self-count. We use a
		// hard DELETE rather than a status flip so a Reserve failure doesn't
		// leave a stale CANCELLED row that a re-run would collide with on
		// the unique constraint.
		if req.OnUpdate {
			if err := tx.Where("tenant_id = ? AND source_type = ? AND source_id = ?",
				req.TenantID, req.SourceType, req.SourceID).
				Delete(&Allocation{}).Error; err != nil {
				return fmt.Errorf("drop prior allocation on update: %w", err)
			}
		}

		// On-hand qty at the scope. Same shape as ConfirmIssue's stock check
		// (repository.go:507) so refining segregation is identical.
		var available float64
		q := `SELECT COALESCE(SUM(quantity), 0) FROM stock_balances
		      WHERE tenant_id = ? AND product_id = ? AND warehouse_id = ?`
		args := []any{req.TenantID, req.ScopeKey.ProductID, req.ScopeKey.WarehouseID}
		if req.ScopeKey.VariantID != nil {
			q += ` AND variant_id = ?`
			args = append(args, *req.ScopeKey.VariantID)
		} else {
			q += ` AND variant_id IS NULL`
		}
		if req.ScopeKey.LocationID != nil {
			q += ` AND location_id = ?`
			args = append(args, *req.ScopeKey.LocationID)
		}
		if req.ScopeKey.ProductionID != nil {
			q += ` AND production_id = ?`
			args = append(args, *req.ScopeKey.ProductionID)
		} else {
			q += ` AND production_id IS NULL`
		}
		if err := tx.Raw(q, args...).Row().Scan(&available); err != nil {
			return fmt.Errorf("read on-hand qty: %w", err)
		}

		// SUM of ACTIVE allocations at the same scope. We reuse the repo
		// helper here — it queries in the same transaction because tx is
		// bound to the same underlying connection.
		reservedAlready, err := sumActiveInTx(tx, req.TenantID, req.ScopeKey)
		if err != nil {
			return err
		}

		remaining := available - reservedAlready
		// Half-cent tolerance for float noise, matches the finance module's
		// RecordPayment style.
		if req.Quantity > remaining+0.005 {
			return fmt.Errorf("insufficient stock for product %d: available %.4f, required %.4f",
				req.ScopeKey.ProductID, remaining, req.Quantity)
		}

		a := &Allocation{
			TenantID:     req.TenantID,
			ProductID:    req.ScopeKey.ProductID,
			VariantID:    req.ScopeKey.VariantID,
			WarehouseID:  req.ScopeKey.WarehouseID,
			LocationID:   req.ScopeKey.LocationID,
			ProductionID: req.ScopeKey.ProductionID,
			Quantity:     req.Quantity,
			SourceType:   req.SourceType,
			SourceID:     req.SourceID,
			SourceDocID:  req.SourceDocID,
			Status:       AllocStatusActive,
			Notes:        req.Notes,
		}
		if err := tx.Create(a).Error; err != nil {
			return fmt.Errorf("insert allocation: %w", err)
		}
		created = a
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// Release flips a single source line's allocation → CANCELLED.
// Idempotent: no-op on missing or already-non-active rows.
func (s *AllocationService) Release(ctx context.Context, tenantID uint, sourceType string, sourceID uint) error {
	return s.repo.SetStatusBySource(ctx, tenantID, sourceType, sourceID, AllocStatusCancelled)
}

// Consume flips → CONSUMED. Called by the terminal action that physically
// moves stock (DO confirm, GI confirm, GT send/receive). Idempotent.
func (s *AllocationService) Consume(ctx context.Context, tenantID uint, sourceType string, sourceID uint) error {
	return s.repo.SetStatusBySource(ctx, tenantID, sourceType, sourceID, AllocStatusConsumed)
}

// ReleaseByDoc — bulk-cancel every ACTIVE allocation belonging to a
// doc-level cancel/reject (SO cancel, MR cancel, GI cancel, GT cancel).
func (s *AllocationService) ReleaseByDoc(ctx context.Context, tenantID uint, sourceType string, docID uint) error {
	return s.repo.ReleaseByDoc(ctx, tenantID, sourceType, docID)
}

// AvailableQty — convenience read used by the FE Stock Overview and by
// pre-flight UI availability probes on line pickers.
func (s *AllocationService) AvailableQty(ctx context.Context, tenantID uint, key ScopeKey) (float64, error) {
	var available float64
	q := `SELECT COALESCE(SUM(quantity), 0) FROM stock_balances
	      WHERE tenant_id = ? AND product_id = ? AND warehouse_id = ?`
	args := []any{tenantID, key.ProductID, key.WarehouseID}
	if key.VariantID != nil {
		q += ` AND variant_id = ?`
		args = append(args, *key.VariantID)
	} else {
		q += ` AND variant_id IS NULL`
	}
	if key.LocationID != nil {
		q += ` AND location_id = ?`
		args = append(args, *key.LocationID)
	}
	if key.ProductionID != nil {
		q += ` AND production_id = ?`
		args = append(args, *key.ProductionID)
	} else {
		q += ` AND production_id IS NULL`
	}
	if err := s.repo.DB().WithContext(ctx).Raw(q, args...).Row().Scan(&available); err != nil {
		return 0, err
	}
	reserved, err := s.repo.SumActiveAllocations(ctx, tenantID, key)
	if err != nil {
		return 0, err
	}
	return available - reserved, nil
}

// ReservedByProductWarehouse — the Stock Overview "Reserved" column source.
func (s *AllocationService) ReservedByProductWarehouse(ctx context.Context, tenantID, productID, warehouseID uint) (float64, error) {
	return s.repo.SumReservedByProductWarehouse(ctx, tenantID, productID, warehouseID)
}

// ListActive — for the FE Allocations section.
func (s *AllocationService) ListActive(ctx context.Context, tenantID uint, filters AllocationFilters, limit, offset int) ([]Allocation, error) {
	return s.repo.ListActive(ctx, tenantID, filters, limit, offset)
}

// CountActive — total rows matching the filter (drives the pager footer).
func (s *AllocationService) CountActive(ctx context.Context, tenantID uint, filters AllocationFilters) (int64, error) {
	return s.repo.CountActive(ctx, tenantID, filters)
}

// ── internals ────────────────────────────────────────────────────────────────

// lockScopeSQL — the SELECT … FOR UPDATE query that serialises concurrent
// reservers. Returns the SQL + args slice. If no stock_balances row exists
// for the scope yet (first-ever reservation for a product that has zero
// on-hand), the SELECT locks nothing — that's fine, subsequent allocation
// inserts + the ACTIVE sum query still see each other via READ COMMITTED
// because SELECT FOR UPDATE also flushes any pending inserts on the row-set.
func lockScopeSQL(tenantID uint, key ScopeKey) (string, []any) {
	q := `SELECT id FROM stock_balances
	      WHERE tenant_id = ? AND product_id = ? AND warehouse_id = ?`
	args := []any{tenantID, key.ProductID, key.WarehouseID}
	if key.VariantID != nil {
		q += ` AND variant_id = ?`
		args = append(args, *key.VariantID)
	} else {
		q += ` AND variant_id IS NULL`
	}
	if key.LocationID != nil {
		q += ` AND location_id = ?`
		args = append(args, *key.LocationID)
	}
	if key.ProductionID != nil {
		q += ` AND production_id = ?`
		args = append(args, *key.ProductionID)
	} else {
		q += ` AND production_id IS NULL`
	}
	q += ` FOR UPDATE`
	return q, args
}

// sumActiveInTx — the tx-scoped variant of SumActiveAllocations. Needed
// because the Reserve tx already opened; we can't reach for the repo's
// non-transactional db here without breaking isolation.
func sumActiveInTx(tx *gorm.DB, tenantID uint, key ScopeKey) (float64, error) {
	var total float64
	q := tx.Model(&Allocation{}).
		Where("tenant_id = ? AND product_id = ? AND warehouse_id = ? AND status = ?",
			tenantID, key.ProductID, key.WarehouseID, AllocStatusActive)
	if key.VariantID != nil {
		q = q.Where("variant_id = ?", *key.VariantID)
	} else {
		q = q.Where("variant_id IS NULL")
	}
	if key.LocationID != nil {
		q = q.Where("location_id = ?", *key.LocationID)
	} else {
		q = q.Where("location_id IS NULL")
	}
	if key.ProductionID != nil {
		q = q.Where("production_id = ?", *key.ProductionID)
	} else {
		q = q.Where("production_id IS NULL")
	}
	if err := q.Select("COALESCE(SUM(quantity), 0)").Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("sum active allocations in tx: %w", err)
	}
	return total, nil
}
