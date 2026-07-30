package procurement

import (
	"context"
	"fmt"
)

// ValidatePOQtyAgainstPR — enforces that the aggregate PO-line qty per
// product across every PO linked to `prID` does not exceed the PR's total
// requested qty per product. If `excludePOLineID` is non-zero, that line's
// quantity is subtracted from the aggregate (used on UpdatePOItem so a line
// doesn't self-count against its own new qty).
//
// The half-cent tolerance mirrors the finance module's payment-comparison
// style — it silences float rounding noise when 15.0000 needs to compare
// equal to 14.9999.
func (s *Service) ValidatePOQtyAgainstPR(
	ctx context.Context,
	tenantID, prID, productID uint,
	newQty float64,
	excludePOLineID uint,
) error {
	// 1. Sum the PR's total requested qty for this product (across every PR line).
	prLines, err := s.repo.ListPRItems(ctx, prID)
	if err != nil {
		return fmt.Errorf("read PR items: %w", err)
	}
	var prTotal float64
	for _, l := range prLines {
		if l.ProductID == productID {
			prTotal += l.Quantity
		}
	}
	if prTotal == 0 {
		// PR doesn't request this product at all. Reject — a PO line for a
		// product outside the PR is inconsistent with the "based on PR"
		// contract.
		return fmt.Errorf("product %d is not on the source PR", productID)
	}

	// 2. Sum every existing PO line for this product across every PO
	// referencing this PR. Exclude the caller's own line on Update.
	pos, err := s.repo.ListPOs(ctx, tenantID, "", nil)
	if err != nil {
		return fmt.Errorf("read POs: %w", err)
	}
	var poTotal float64
	for _, p := range pos {
		if p.PRID == nil || *p.PRID != prID {
			continue
		}
		items, err := s.repo.ListPOItems(ctx, p.ID)
		if err != nil {
			return fmt.Errorf("read PO items: %w", err)
		}
		for _, li := range items {
			if li.ProductID != productID {
				continue
			}
			if excludePOLineID > 0 && li.ID == excludePOLineID {
				continue
			}
			poTotal += li.Quantity
		}
	}

	if poTotal+newQty > prTotal+0.005 {
		return fmt.Errorf(
			"PO quantity for product %d exceeds PR remaining: %.4f + %.4f > %.4f",
			productID, poTotal, newQty, prTotal,
		)
	}
	return nil
}

// ValidateGRNQtyAgainstPOLine — cap a GRN line's qty against the
// remaining-to-receive balance on its source PO line.  `excludeGRNLineID`
// isn't used here (received_qty is the accumulator that already includes
// past receipts), but the signature keeps the shape parallel to
// ValidatePOQtyAgainstPR for consistency.
func (s *Service) ValidateGRNQtyAgainstPOLine(
	ctx context.Context,
	tenantID, poLineID uint,
	newQty float64,
	// Optional: when updating an existing GRN line, its own contribution
	// hasn't yet flowed into POLine.received_qty (that bump only happens on
	// GRN confirm). So callers on Update should pass the previous qty of
	// the line being updated, which we subtract from newQty for the cap
	// computation — same-line edits net out to just the delta.
	previousQtyOnThisLine float64,
) error {
	poLine, err := s.repo.GetPOItem(ctx, tenantID, poLineID)
	if err != nil {
		return fmt.Errorf("PO line %d not found", poLineID)
	}
	remaining := poLine.Quantity - poLine.ReceivedQty
	delta := newQty - previousQtyOnThisLine
	if delta > remaining+0.005 {
		return fmt.Errorf(
			"GRN line for product %d exceeds PO line remaining: %.4f > %.4f",
			poLine.ProductID, delta, remaining,
		)
	}
	return nil
}
