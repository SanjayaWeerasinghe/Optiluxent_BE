package manufacturing

import (
	"context"
	"fmt"
)

// ── Dashboard data readers (injected from other modules) ──────────────────────

// MRReader / GRNReader are thin cross-module ports the dashboard uses to gather
// linked documents against a Manufacturing Order. inventory.Service and
// procurement.Service satisfy these directly via their existing List*ByMO
// methods.
type MRLineRow struct {
	ProductID    uint    `json:"product_id"`
	UOMID        uint    `json:"uom_id"`
	RequestedQty float64 `json:"requested_qty"`
	IssuedQty    float64 `json:"issued_qty"`
}

type LinkedMR struct {
	ID     uint        `json:"id"`
	Code   string      `json:"code"`
	Status string      `json:"status"`
	Lines  []MRLineRow `json:"lines"`
}

type StockLineRow struct {
	ProductID uint    `json:"product_id"`
	UOMID     uint    `json:"uom_id"`
	Quantity  float64 `json:"quantity"`
}

type LinkedGI struct {
	ID     uint           `json:"id"`
	Code   string         `json:"code"`
	Status string         `json:"status"`
	Lines  []StockLineRow `json:"lines"`
}

type LinkedGT struct {
	ID     uint           `json:"id"`
	Code   string         `json:"code"`
	Status string         `json:"status"`
	Lines  []StockLineRow `json:"lines"`
}

type LinkedGRN struct {
	ID      uint           `json:"id"`
	Code    string         `json:"code"`
	Status  string         `json:"status"`
	GRNType string         `json:"grn_type"`
	Lines   []StockLineRow `json:"lines"`
}

type LinkedQC struct {
	ID         uint    `json:"id"`
	Code       string  `json:"code"`
	QCType     string  `json:"qc_type"`
	Status     string  `json:"status"`
	QtyChecked float64 `json:"qty_checked"`
	QtyPassed  float64 `json:"qty_passed"`
	QtyFailed  float64 `json:"qty_failed"`
	RefType    string  `json:"reference_type"`
	RefID      *uint   `json:"reference_id"`
}

type DashboardReaders struct {
	ListMRsByMO       func(ctx context.Context, tenantID, moID uint) ([]LinkedMR, error)
	ListGIsByMO       func(ctx context.Context, tenantID, moID uint) ([]LinkedGI, error)
	ListGTsByMO       func(ctx context.Context, tenantID, moID uint) ([]LinkedGT, error)
	ListGRNsByMO      func(ctx context.Context, tenantID, moID uint) ([]LinkedGRN, error)
	ListQCsForGRNs    func(ctx context.Context, tenantID uint, grnIDs []uint) ([]LinkedQC, error)
}

func (s *Service) SetDashboardReaders(r DashboardReaders) { s.dashReaders = r }

// ── Response packet ───────────────────────────────────────────────────────────

type DashboardTotals struct {
	// Denominators / planned
	PlannedLitres float64 `json:"planned_litres"`

	// Aggregates
	RequestedLitres         float64 `json:"requested_litres"`
	IssuedLitres            float64 `json:"issued_litres"`
	ProducedLitres          float64 `json:"produced_litres"`           // all bottled (pass + fail)
	ProducedOKLitres        float64 `json:"produced_ok_litres"`        // QC-passed
	ProducedQCFailedLitres  float64 `json:"produced_qc_failed_litres"` // thrown-out
	WastageUpstreamLitres   float64 `json:"wastage_upstream_litres"`   // issued − produced (refining loss)

	// Percentages
	WastageUpstreamPct float64 `json:"wastage_upstream_pct"`
	ProducedOKPct      float64 `json:"produced_ok_pct"`
}

type DashboardResponse struct {
	Order            *ProductionOrder `json:"order"`
	MaterialRequests []LinkedMR       `json:"material_requests"`
	Issues           []LinkedGI       `json:"issues"`
	Transfers        []LinkedGT       `json:"transfers"`
	GRNs             []LinkedGRN      `json:"grns"`
	QCs              []LinkedQC       `json:"qcs"`
	Totals           DashboardTotals  `json:"totals"`
}

// GetDashboard assembles the full dashboard packet for an MO.
func (s *Service) GetDashboard(ctx context.Context, tenantID, moID uint) (*DashboardResponse, error) {
	order, err := s.repo.GetOrder(ctx, tenantID, moID)
	if err != nil {
		return nil, fmt.Errorf("production order not found: %w", err)
	}
	resp := &DashboardResponse{Order: order}

	r := s.dashReaders
	if r.ListMRsByMO != nil {
		resp.MaterialRequests, _ = r.ListMRsByMO(ctx, tenantID, moID)
	}
	if r.ListGIsByMO != nil {
		resp.Issues, _ = r.ListGIsByMO(ctx, tenantID, moID)
	}
	if r.ListGTsByMO != nil {
		resp.Transfers, _ = r.ListGTsByMO(ctx, tenantID, moID)
	}
	if r.ListGRNsByMO != nil {
		resp.GRNs, _ = r.ListGRNsByMO(ctx, tenantID, moID)
	}
	if r.ListQCsForGRNs != nil {
		grnIDs := make([]uint, 0, len(resp.GRNs))
		for _, g := range resp.GRNs {
			grnIDs = append(grnIDs, g.ID)
		}
		if len(grnIDs) > 0 {
			resp.QCs, _ = r.ListQCsForGRNs(ctx, tenantID, grnIDs)
		}
	}

	resp.Totals = s.computeTotals(ctx, order, resp)
	return resp, nil
}

// computeTotals normalises heterogeneous quantities to litres for comparison.
// Any line that can't be converted falls back to raw qty (a warning would be
// better but is unimportant right now).
func (s *Service) computeTotals(ctx context.Context, order *ProductionOrder, r *DashboardResponse) DashboardTotals {
	toLitres := func(productID, uomID uint, qty float64) float64 {
		if s.db == nil {
			return qty
		}
		// Litres UOM: we use the order's UOM as our "canonical" unit for the
		// dashboard. If a line is in a different UOM, convert product → order-uom.
		if uomID == order.UOMID {
			return qty
		}
		if v, ok := ConvertProductQty(ctx, s.db, productID, uomID, order.UOMID, qty); ok {
			return v
		}
		return qty
	}

	var t DashboardTotals
	t.PlannedLitres = order.PlannedQty

	for _, mr := range r.MaterialRequests {
		for _, l := range mr.Lines {
			t.RequestedLitres += toLitres(l.ProductID, l.UOMID, l.RequestedQty)
			t.IssuedLitres    += toLitres(l.ProductID, l.UOMID, l.IssuedQty)
		}
	}
	// GIs may exist without MRs — include them too, but avoid double-count when
	// we already summed via the MR issued_qty roll-up.
	if len(r.MaterialRequests) == 0 {
		for _, gi := range r.Issues {
			for _, l := range gi.Lines {
				t.IssuedLitres += toLitres(l.ProductID, l.UOMID, l.Quantity)
			}
		}
	}

	// Produced total = ALL confirmed PRODUCTION_OUTPUT GRN line qty (in litres).
	// QC-failed comes from the linked Product QCs.
	grnLitresByProduct := map[uint]float64{}
	for _, g := range r.GRNs {
		if g.Status != "CONFIRMED" {
			continue
		}
		for _, l := range g.Lines {
			v := toLitres(l.ProductID, l.UOMID, l.Quantity)
			t.ProducedLitres += v
			grnLitresByProduct[l.ProductID] += v
		}
	}

	_ = grnLitresByProduct // kept for potential per-product roll-ups later

	// Derive per-GRN produced litres and apply the QC's own pass/fail ratio,
	// matching QC by ReferenceType/ID = "GRN"/grn.ID.
	qcByGRN := map[uint]*LinkedQC{}
	for i := range r.QCs {
		q := &r.QCs[i]
		if q.RefType == "GRN" && q.RefID != nil {
			qcByGRN[*q.RefID] = q
		}
	}
	for _, g := range r.GRNs {
		if g.Status != "CONFIRMED" {
			continue
		}
		var grnLitres float64
		for _, l := range g.Lines {
			grnLitres += toLitres(l.ProductID, l.UOMID, l.Quantity)
		}
		qc := qcByGRN[g.ID]
		if qc == nil || (qc.QtyPassed+qc.QtyFailed) <= 0 {
			// No QC decided yet — count as OK provisionally
			t.ProducedOKLitres += grnLitres
			continue
		}
		total := qc.QtyPassed + qc.QtyFailed
		t.ProducedOKLitres       += grnLitres * (qc.QtyPassed / total)
		t.ProducedQCFailedLitres += grnLitres * (qc.QtyFailed / total)
	}

	// Refining loss = what we issued but did not appear as any bottled output.
	t.WastageUpstreamLitres = t.IssuedLitres - t.ProducedLitres
	if t.WastageUpstreamLitres < 0 {
		t.WastageUpstreamLitres = 0
	}

	if t.IssuedLitres > 0 {
		t.WastageUpstreamPct = round2(t.WastageUpstreamLitres / t.IssuedLitres * 100)
		t.ProducedOKPct      = round2(t.ProducedOKLitres / t.IssuedLitres * 100)
	}
	return t
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
