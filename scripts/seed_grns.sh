#!/bin/sh
# GRN Seed — Kadahapola Exports Ltd
# Creates 3 partial GRNs across PO00006, PO00008, PO00007
# Each GRN is confirmed so invoice lines are auto-populated
# Run: docker exec -e SEED_PASSWORD=Admin@1234 erp-api-dev sh /app/scripts/seed_grns.sh

BASE="http://localhost:3000/api/v1"

# ── Login ──────────────────────────────────────────────────────────────────────
echo ">>> Logging in..."
LOGIN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@optiluxent.com","password":"'"$SEED_PASSWORD"'"}')
TOKEN=$(echo "$LOGIN" | jq -r '.data.access_token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "ERROR: Login failed."
  echo "$LOGIN"
  exit 1
fi
echo "    Token acquired."

post() { curl -s -X POST "$BASE/$1" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$2"; }
get()  { curl -s         "$BASE/$1" -H "Authorization: Bearer $TOKEN"; }
idof() { echo "$1" | jq -r '.data.id // empty'; }
ok()   { echo "$1" | jq -r '.data.status // empty'; }

# Known IDs from previous seeds (PO lines + locations)
# PO00006 lines: cloves line_id=7 (product_id=4, qty=300, LKR3100)
# PO00007 lines: cardamom line_id=8 (product_id=42, qty=150, LKR4350)
#                turmeric  line_id=9 (product_id=43, qty=250, LKR650)
# PO00008 lines: cinnamon  line_id=10 (product_id=2, qty=500, LKR1450 -2%)
#                blk pepper line_id=11 (product_id=3, qty=200, LKR2750)
# Locations: A1=2 (Dry Spices), A2=3 (Teas), B1=4 (Oils)
# Suppliers: CEYLON=2, INDIAN=8, GALLE=43

# ══════════════════════════════════════════════════════════════════════════════
# GRN-001 — PARTIAL: PO00006 Cloves (200 of 300 kg)
# First truck delivery arrived; second 100kg shipment still at estate
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " GRN-001  PO00006 — Cloves partial 200/300 kg"
echo "════════════════════════════════════════"

r=$(post "procurement/goods-receipts" '{
  "po_id":        6,
  "supplier_id":  43,
  "receipt_date": "2026-06-11",
  "warehouse_id": 1,
  "notes": "First delivery from Galle Cloves Estate. 200kg of 300kg ordered. Remaining 100kg expected June 17 (pending drying completion at estate)."
}')
GRN1=$(idof "$r")
echo "  Created GRN-001 id=$GRN1 status=$(ok "$r")"

post "procurement/goods-receipts/$GRN1/items" '{
  "po_line_id":  7,
  "product_id":  4,
  "quantity":    200,
  "uom_id":      1,
  "location_id": 2,
  "unit_cost":   3100,
  "description": "Grade A Whole Cloves — 1st delivery, 200kg",
  "notes":       "Moisture test passed 11.2%. Lab cert attached. Batch CLV-2026-06-A."
}' > /dev/null
echo "  Added: Cloves 200 kg → Rack A1 (Dry Spices) @ LKR 3,100"

r=$(post "procurement/goods-receipts/$GRN1/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# GRN-002 — PARTIAL: PO00008 Cinnamon (350/500 kg) + Black Pepper (200/200 kg)
# Cinnamon short-shipped — supplier stockout on Grade A; Pepper fully delivered
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " GRN-002  PO00008 — Cinnamon 350/500 kg + Black Pepper 200/200 kg"
echo "════════════════════════════════════════"

r=$(post "procurement/goods-receipts" '{
  "po_id":        8,
  "supplier_id":  2,
  "receipt_date": "2026-06-10",
  "warehouse_id": 1,
  "notes": "Ceylon Spice Traders delivery. Black Pepper fully delivered. Cinnamon short — only 350 of 500kg Grade A available. Remaining 150kg promised next batch June 20."
}')
GRN2=$(idof "$r")
echo "  Created GRN-002 id=$GRN2"

post "procurement/goods-receipts/$GRN2/items" '{
  "po_line_id":  10,
  "product_id":  2,
  "quantity":    350,
  "uom_id":      1,
  "location_id": 2,
  "unit_cost":   1450,
  "description": "Grade A Ceylon Cinnamon Sticks — partial 350kg of 500kg",
  "notes":       "Shortfall 150kg due to supplier Grade A stockout. DN ref: CST-2026-0610."
}' > /dev/null
echo "  Added: Cinnamon 350 kg (of 500) → Rack A1 @ LKR 1,450"

post "procurement/goods-receipts/$GRN2/items" '{
  "po_line_id":  11,
  "product_id":  3,
  "quantity":    200,
  "uom_id":      1,
  "location_id": 2,
  "unit_cost":   2750,
  "description": "Whole Black Pepper — fully delivered 200kg",
  "notes":       "Phytosanitary cert attached. All 200kg received. Batch BPP-2026-06-B."
}' > /dev/null
echo "  Added: Black Pepper 200 kg (full) → Rack A1 @ LKR 2,750"

r=$(post "procurement/goods-receipts/$GRN2/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# GRN-003 — PARTIAL: PO00007 Turmeric only (200/250 kg); Cardamom NOT received
# Cardamom batch rejected at intake — curcumin content below spec (<3%)
# Turmeric partial — last 50kg held for re-drying
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " GRN-003  PO00007 — Turmeric 200/250 kg (Cardamom rejected, not received)"
echo "════════════════════════════════════════"

r=$(post "procurement/goods-receipts" '{
  "po_id":        7,
  "supplier_id":  8,
  "receipt_date": "2026-06-12",
  "warehouse_id": 1,
  "notes": "Indian Spice Exporters delivery. Cardamom 150kg REJECTED at QC — curcumin content 1.8%, spec requires min 3%. Returned to supplier. Turmeric partial: 200 of 250kg; remaining 50kg on re-drying rack, ETA June 18."
}')
GRN3=$(idof "$r")
echo "  Created GRN-003 id=$GRN3"

post "procurement/goods-receipts/$GRN3/items" '{
  "po_line_id":  9,
  "product_id":  43,
  "quantity":    200,
  "uom_id":      1,
  "location_id": 2,
  "unit_cost":   650,
  "description": "Fresh Turmeric Rhizomes — 200kg of 250kg ordered",
  "notes":       "Curcumin 3.4% — passed. Remaining 50kg pending re-drying. Batch TRM-2026-06-C."
}' > /dev/null
echo "  Added: Turmeric 200 kg (of 250) → Rack A1 @ LKR 650"
echo "  NOTE: Cardamom 150kg NOT added — rejected at QC intake, returned to supplier"

r=$(post "procurement/goods-receipts/$GRN3/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# Show invoice state after all GRNs confirmed
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " INVOICE STATE AFTER GRN CONFIRMATION"
echo "════════════════════════════════════════"

for PO_ID in 6 7 8; do
  PO_CODE=$(get "procurement/purchase-orders/$PO_ID" | jq -r '.data.code')
  INV=$(get "procurement/purchase-invoices" | jq -r ".data[] | select(.po_id == $PO_ID and .status == \"DRAFT\")")
  INV_ID=$(echo "$INV" | jq -r '.id')
  INV_CODE=$(echo "$INV" | jq -r '.code')
  INV_TOTAL=$(echo "$INV" | jq -r '.total_amount')
  echo ""
  echo "  $PO_CODE → $INV_CODE (DRAFT)  total=LKR$INV_TOTAL"
  LINES=$(get "procurement/purchase-invoices/$INV_ID/lines" 2>/dev/null \
    || curl -s "$BASE/procurement/purchase-invoices/$INV_ID" -H "Authorization: Bearer $TOKEN" | jq -r '.data.lines // empty')
  if [ -n "$LINES" ] && [ "$LINES" != "null" ]; then
    echo "$LINES" | jq -r '.[] | "    line \(.line_number): product_id=\(.product_id) qty=\(.quantity) price=LKR\(.unit_price) total=LKR\(.line_total)"' 2>/dev/null || \
    echo "$LINES" | jq -r '.data[] | "    line \(.line_number): product_id=\(.product_id) qty=\(.quantity) price=LKR\(.unit_price) total=LKR\(.line_total)"' 2>/dev/null
  fi
done

echo ""
echo "════════════════════════════════════════"
echo " SUMMARY"
echo "════════════════════════════════════════"
echo "  GRN-001 id=$GRN1 → PO00006  Cloves     200/300 kg   LKR 620,000"
echo "  GRN-002 id=$GRN2 → PO00008  Cinnamon   350/500 kg   LKR 507,550"
echo "                               BlackPepper 200/200 kg  LKR 550,000"
echo "  GRN-003 id=$GRN3 → PO00007  Turmeric   200/250 kg   LKR 130,000"
echo "                               Cardamom     0/150 kg   (REJECTED — not received)"
echo ""
echo "  ALL DONE"
