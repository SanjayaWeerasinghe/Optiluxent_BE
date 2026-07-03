package manufacturing

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// ConvertProductQty converts a (productID, qty, fromUOM) to `toUOM` using the
// per-product conversion table. Used by the MO dashboard to normalise
// heterogeneous quantities into a single unit (litres) for comparison.
//
// Returns (converted, true) on success; (qty, false) if no conversion row
// exists — the caller can then decide to skip the item or treat qty as-is.
func ConvertProductQty(ctx context.Context, db *gorm.DB, productID, fromUOM, toUOM uint, qty float64) (float64, bool) {
	if fromUOM == toUOM {
		return qty, true
	}
	var ratio float64
	err := db.WithContext(ctx).Raw(
		`SELECT ratio FROM uom_conversions
		 WHERE product_id = ? AND from_uom_id = ? AND to_uom_id = ? LIMIT 1`,
		productID, fromUOM, toUOM,
	).Row().Scan(&ratio)
	if err == nil {
		return qty * ratio, true
	}
	// Try the inverse direction — one row often serves both ways.
	err = db.WithContext(ctx).Raw(
		`SELECT ratio FROM uom_conversions
		 WHERE product_id = ? AND from_uom_id = ? AND to_uom_id = ? LIMIT 1`,
		productID, toUOM, fromUOM,
	).Row().Scan(&ratio)
	if err == nil && ratio != 0 {
		return qty / ratio, true
	}
	if err == sql.ErrNoRows {
		return qty, false
	}
	return qty, false
}

// ConvertMaterialQty is the same idea anchored on a material_id.
func ConvertMaterialQty(ctx context.Context, db *gorm.DB, materialID, fromUOM, toUOM uint, qty float64) (float64, bool) {
	if fromUOM == toUOM {
		return qty, true
	}
	var ratio float64
	err := db.WithContext(ctx).Raw(
		`SELECT ratio FROM uom_conversions
		 WHERE material_id = ? AND from_uom_id = ? AND to_uom_id = ? LIMIT 1`,
		materialID, fromUOM, toUOM,
	).Row().Scan(&ratio)
	if err == nil {
		return qty * ratio, true
	}
	err = db.WithContext(ctx).Raw(
		`SELECT ratio FROM uom_conversions
		 WHERE material_id = ? AND from_uom_id = ? AND to_uom_id = ? LIMIT 1`,
		materialID, toUOM, fromUOM,
	).Row().Scan(&ratio)
	if err == nil && ratio != 0 {
		return qty / ratio, true
	}
	if err == sql.ErrNoRows {
		return qty, false
	}
	return qty, false
}
