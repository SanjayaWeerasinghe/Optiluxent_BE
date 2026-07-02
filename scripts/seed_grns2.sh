#!/bin/sh
# GRN Seed (Round 2) — Kadahapola Exports Ltd
# Adds follow-up GRNs to existing POs — demonstrates multiple GRNs per PO
# Each confirmed GRN appends new lines to the draft invoice
#
# Outstanding after Round 1:
#   PO00006  Cloves        100 kg remaining
#   PO00007  Turmeric       50 kg remaining | Cardamom 150 kg rejected (supplier resending)
#   PO00008  Cinnamon      150 kg remaining
#
# Run: docker exec -e SEED_PASSWORD=Admin@1234 erp-api-dev sh /app/scripts/seed_grns2.sh

BASE="http://localhost:3000/api/v1"

echo ">>> Logging in..."
LOGIN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@optiluxent.com","password":"'"$SEED_PASSWORD"'"}')
TOKEN=$(echo "$LOGIN" | jq -r '.data.access_token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "ERROR: Login failed."; echo "$LOGIN"; exit 1
fi
echo "    Token acquired."

post() { curl -s -X POST "$BASE/$1" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$2"; }
get()  { curl -s         "$BASE/$1" -H "Authorization: Bearer $TOKEN"; }
idof() { echo "$1" | jq -r '.data.id // empty'; }
ok()   { echo "$1" | jq -r '.data.status // empty'; }

# ══════════════════════════════════════════════════════════════════════════════
# GRN-004 — PO00006: 2nd delivery — final 100 kg Cloves (completes the PO)
# Invoice PINV00006 will gain a 2nd line → total becomes LKR 930,000
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " GRN-004  PO00006 — Cloves final 100 kg (2nd and last delivery)"
echo "════════════════════════════════════════"

r=$(post "procurement/goods-receipts" '{
  "po_id":        6,
  "supplier_id":  43,
  "receipt_date": "2026-06-17",
  "warehouse_id": 1,
  "notes": "Second and final delivery from Galle Cloves Estate. Remaining 100kg from estate drying batch. PO now fully received."
}')
GRN4=$(idof "$r")
echo "  Created GRN-004 id=$GRN4 status=$(ok "$r")"

post "procurement/goods-receipts/$GRN4/items" '{
  "po_line_id":  7,
  "product_id":  4,
  "quantity":    100,
  "uom_id":      1,
  "location_id": 2,
  "unit_cost":   3100,
  "description": "Grade A Whole Cloves — 2nd delivery, final 100kg",
  "notes":       "Moisture 10.8%. Batch CLV-2026-06-B. PO fully received."
}' > /dev/null
echo "  Added: Cloves 100 kg → Rack A1 @ LKR 3,100"

r=$(post "procurement/goods-receipts/$GRN4/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# GRN-005 — PO00007: 2nd delivery — remaining Turmeric 50 kg
# Invoice PINV00009 gains a 2nd line for Turmeric, completing that line
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " GRN-005  PO00007 — Turmeric remaining 50 kg (completes Turmeric line)"
echo "════════════════════════════════════════"

r=$(post "procurement/goods-receipts" '{
  "po_id":        7,
  "supplier_id":  8,
  "receipt_date": "2026-06-18",
  "warehouse_id": 1,
  "notes": "Indian Spice Exporters — Turmeric balance 50kg from re-drying. Turmeric PO line now fully received."
}')
GRN5=$(idof "$r")
echo "  Created GRN-005 id=$GRN5"

post "procurement/goods-receipts/$GRN5/items" '{
  "po_line_id":  9,
  "product_id":  43,
  "quantity":    50,
  "uom_id":      1,
  "location_id": 2,
  "unit_cost":   650,
  "description": "Turmeric Rhizomes — final 50kg after re-drying",
  "notes":       "Curcumin 3.6%. Batch TRM-2026-06-D. Turmeric line now complete."
}' > /dev/null
echo "  Added: Turmeric 50 kg → Rack A1 @ LKR 650"

r=$(post "procurement/goods-receipts/$GRN5/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# GRN-006 — PO00007: 3rd delivery — Cardamom replacement batch 100 kg
# Supplier sent a new batch after the previous 150 kg was rejected
# Partial: only 100 of 150 kg this time; 50 kg still pending supplier
# Invoice PINV00009 gains a 3rd line for Cardamom
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " GRN-006  PO00007 — Cardamom replacement 100 kg (supplier resent after rejection)"
echo "════════════════════════════════════════"

r=$(post "procurement/goods-receipts" '{
  "po_id":        7,
  "supplier_id":  8,
  "receipt_date": "2026-06-19",
  "warehouse_id": 1,
  "notes": "Indian Spice Exporters replacement Cardamom batch. Previous 150kg was rejected (curcumin 1.8%). New batch: 100kg cleared QC at 3.2%. Remaining 50kg from new harvest expected June 28."
}')
GRN6=$(idof "$r")
echo "  Created GRN-006 id=$GRN6"

post "procurement/goods-receipts/$GRN6/items" '{
  "po_line_id":  8,
  "product_id":  42,
  "quantity":    100,
  "uom_id":      1,
  "location_id": 2,
  "unit_cost":   4350,
  "description": "Green Cardamom Pods 7mm+ — replacement batch 100kg",
  "notes":       "Curcumin 3.2% — passed QC. Batch CDM-2026-06-R1. Still 50kg outstanding."
}' > /dev/null
echo "  Added: Cardamom 100 kg → Rack A1 @ LKR 4,350"

r=$(post "procurement/goods-receipts/$GRN6/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# GRN-007 — PO00008: 2nd delivery — Cinnamon 100 kg (partial, 50 kg still due)
# Invoice PINV00008 gains a 3rd line
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " GRN-007  PO00008 — Cinnamon 100 kg (2nd delivery, 50 kg still outstanding)"
echo "════════════════════════════════════════"

r=$(post "procurement/goods-receipts" '{
  "po_id":        8,
  "supplier_id":  2,
  "receipt_date": "2026-06-20",
  "warehouse_id": 1,
  "notes": "Ceylon Spice Traders 2nd Cinnamon delivery. 100kg of remaining 150kg. Final 50kg on next truck, ETA June 25."
}')
GRN7=$(idof "$r")
echo "  Created GRN-007 id=$GRN7"

post "procurement/goods-receipts/$GRN7/items" '{
  "po_line_id":  10,
  "product_id":  2,
  "quantity":    100,
  "uom_id":      1,
  "location_id": 2,
  "unit_cost":   1450,
  "description": "Grade A Ceylon Cinnamon Sticks — 2nd delivery 100kg",
  "notes":       "Batch CINN-2026-06-B. 50kg still outstanding for final delivery."
}' > /dev/null
echo "  Added: Cinnamon 100 kg → Rack A1 @ LKR 1,450"

r=$(post "procurement/goods-receipts/$GRN7/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# GRN-008 — PO00008: 3rd and final delivery — Cinnamon last 50 kg (completes PO)
# Invoice PINV00008 gains a 4th line — PO fully received
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " GRN-008  PO00008 — Cinnamon final 50 kg (3rd delivery, PO complete)"
echo "════════════════════════════════════════"

r=$(post "procurement/goods-receipts" '{
  "po_id":        8,
  "supplier_id":  2,
  "receipt_date": "2026-06-25",
  "warehouse_id": 1,
  "notes": "Ceylon Spice Traders final Cinnamon delivery. All 500kg now received. PO fully complete."
}')
GRN8=$(idof "$r")
echo "  Created GRN-008 id=$GRN8"

post "procurement/goods-receipts/$GRN8/items" '{
  "po_line_id":  10,
  "product_id":  2,
  "quantity":    50,
  "uom_id":      1,
  "location_id": 2,
  "unit_cost":   1450,
  "description": "Grade A Ceylon Cinnamon Sticks — final 50kg, PO complete",
  "notes":       "Batch CINN-2026-06-C. All 500kg of PO now received."
}' > /dev/null
echo "  Added: Cinnamon 50 kg → Rack A1 @ LKR 1,450"

r=$(post "procurement/goods-receipts/$GRN8/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# Final state: all invoices with lines
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " INVOICE STATE — ALL LINES"
echo "════════════════════════════════════════"

for PO_ID_INV_ID in "6:6" "7:9" "8:8"; do
  PO_ID=$(echo "$PO_ID_INV_ID" | cut -d: -f1)
  INV_ID=$(echo "$PO_ID_INV_ID" | cut -d: -f2)
  INV=$(get "procurement/purchase-invoices/$INV_ID")
  INV_CODE=$(echo "$INV" | jq -r '.data.code')
  INV_TOTAL=$(echo "$INV" | jq -r '.data.total_amount')
  LINE_COUNT=$(echo "$INV" | jq -r '.data.lines | length')
  echo ""
  echo "  $INV_CODE  po_id=$PO_ID  DRAFT  total=LKR$INV_TOTAL  ($LINE_COUNT lines)"
  echo "$INV" | jq -r '.data.lines[] | "    L\(.line_number): product_id=\(.product_id)  qty=\(.quantity)  @LKR\(.unit_price)  = LKR\(.line_total)  [grn_line=\(.grn_line_id)]"'
done

echo ""
echo "════════════════════════════════════════"
echo " PO OUTSTANDING QTY SUMMARY"
echo "════════════════════════════════════════"
for PO_ID in 6 7 8; do
  PO=$(get "procurement/purchase-orders/$PO_ID")
  CODE=$(echo "$PO" | jq -r '.data.code')
  STATUS=$(echo "$PO" | jq -r '.data.status')
  echo "  $CODE [$STATUS]"
  get "procurement/purchase-orders/$PO_ID/items" \
    | jq -r '.data[] | "    product_id=\(.product_id)  ordered=\(.quantity)  received=\(.received_qty)  outstanding=\(.quantity - .received_qty)"'
done

echo ""
echo "════════════════════════════════════════"
echo " GRN HISTORY PER PO"
echo "════════════════════════════════════════"
ALL_GRNS=$(get "procurement/goods-receipts")
for PO_ID in 6 7 8; do
  PO_CODE=$(get "procurement/purchase-orders/$PO_ID" | jq -r '.data.code')
  echo "  $PO_CODE:"
  echo "$ALL_GRNS" | jq -r ".data[] | select(.po_id == $PO_ID) | \"    GRN id=\(.id) \(.code) \(.receipt_date) status=\(.status)\""
done

echo ""
echo "  ALL DONE"
