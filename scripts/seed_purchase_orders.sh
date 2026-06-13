#!/bin/sh
# Purchase Order Seed — Kadahapola Exports Ltd
# Creates 4 POs linked to PRs, covering: DRAFT, CONFIRMED, CANCELLED statuses
# Run: docker exec -e SEED_PASSWORD=Admin@1234 erp-api-dev sh /app/scripts/seed_purchase_orders.sh

BASE="http://localhost:3000/api/v1"

# ── Login ──────────────────────────────────────────────────────────────────────
echo ">>> Logging in..."
LOGIN=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@optiluxent.com","password":"'"$SEED_PASSWORD"'"}')
TOKEN=$(echo "$LOGIN" | jq -r '.data.access_token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "ERROR: Login failed. Set SEED_PASSWORD env var."
  echo "$LOGIN"
  exit 1
fi
echo "    Token acquired."

post() { curl -s -X POST "$BASE/$1" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$2"; }
get()  { curl -s         "$BASE/$1" -H "Authorization: Bearer $TOKEN"; }
idof() { echo "$1" | jq -r '.data.id // empty'; }
ok()   { echo "$1" | jq -r '.data.status // .data.id // empty'; }

# ── Re-fetch Master Data IDs ───────────────────────────────────────────────────
echo ">>> Fetching master data IDs..."

PROD_LIST=$(get "masterdata/products?per_page=100")
CINN_ID=$(echo "$PROD_LIST"  | jq -r '.data[] | select(.code=="CINN-STCK")   | .id')
BLKP_ID=$(echo "$PROD_LIST"  | jq -r '.data[] | select(.code=="BLKPEP")      | .id')
CLOV_ID=$(echo "$PROD_LIST"  | jq -r '.data[] | select(.code=="CLOVES")      | .id')
JAR_ID=$(echo "$PROD_LIST"   | jq -r '.data[] | select(.code=="JAR-250")     | .id')
CINNP_ID=$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CINN-PWD-250") | .id')
CARDM_ID=$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CARDM")       | .id')
TURM_ID=$(echo "$PROD_LIST"  | jq -r '.data[] | select(.code=="TURMRIC")     | .id')

UOM_LIST=$(get "masterdata/products/uoms?per_page=20")
KG_ID=$(echo "$UOM_LIST"  | jq -r '.data[] | select(.code=="KG")  | .id')
L_ID=$(echo "$UOM_LIST"   | jq -r '.data[] | select(.code=="L")   | .id')
PCS_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="PCS") | .id')
PKT_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="PKT") | .id')

CURR_LIST=$(get "masterdata/financial/currencies?per_page=20")
LKR_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="LKR") | .id')

PT_LIST=$(get "masterdata/financial/payment-terms?per_page=20")
NET30_ID=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="NET30") | .id')
NET15_ID=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="NET15") | .id')
ADV50_ID=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="ADV50") | .id')

TAX_LIST=$(get "masterdata/financial/tax-codes?per_page=20")
VAT00_ID=$(echo "$TAX_LIST" | jq -r '.data[] | select(.code=="VAT00") | .id')
VAT08_ID=$(echo "$TAX_LIST" | jq -r '.data[] | select(.code=="VAT08") | .id')

SUPP_LIST=$(get "masterdata/contacts/parties?party_type=SUPPLIER&per_page=20")
CEYLON_ID=$(echo "$SUPP_LIST" | jq -r '.data[] | select(.code=="SUPP001") | .id')
LANKA_ID=$(echo "$SUPP_LIST"  | jq -r '.data[] | select(.code=="SUPP003") | .id')
GALLE_ID=$(echo "$SUPP_LIST"  | jq -r '.data[] | select(.code=="SUPP005") | .id')
INDIAN_ID=$(echo "$SUPP_LIST" | jq -r '.data[] | select(.code=="SUPP004") | .id')

WH_LIST=$(get "masterdata/inventory/warehouses?per_page=20")
WH1_ID=$(echo "$WH_LIST" | jq -r '.data[] | select(.code=="WH001") | .id')

# Re-fetch PR IDs from the active purchase requests
PR_LIST=$(get "procurement/purchase-requests")
PR1_ID=$(echo "$PR_LIST" | jq -r '.data[] | select(.notes | test("Monthly.*Cinnamon")) | .id' | head -1)
PR2_ID=$(echo "$PR_LIST" | jq -r '.data[] | select(.notes | test("packaging stock")) | .id' | head -1)
PR3_ID=$(echo "$PR_LIST" | jq -r '.data[] | select(.notes | test("London Herb Co")) | .id' | head -1)
PR6_ID=$(echo "$PR_LIST" | jq -r '.data[] | select(.notes | test("Cardamom and Turmeric")) | .id' | head -1)

echo "  Products  CINN=$CINN_ID BLKP=$BLKP_ID CLOV=$CLOV_ID JAR=$JAR_ID CINNP=$CINNP_ID CARDM=$CARDM_ID TURM=$TURM_ID"
echo "  UOMs      KG=$KG_ID PCS=$PCS_ID PKT=$PKT_ID"
echo "  LKR=$LKR_ID  VAT00=$VAT00_ID VAT08=$VAT08_ID"
echo "  Payment   NET30=$NET30_ID NET15=$NET15_ID ADV50=$ADV50_ID"
echo "  Suppliers CEYLON=$CEYLON_ID LANKA=$LANKA_ID GALLE=$GALLE_ID INDIAN=$INDIAN_ID"
echo "  WH1=$WH1_ID"
echo "  PRs       PR1=$PR1_ID PR2=$PR2_ID PR3=$PR3_ID PR6=$PR6_ID"

# ══════════════════════════════════════════════════════════════════════════════
# PO-001 — DRAFT: Monthly Spice Restock (linked to PR-001)
# PR is still DRAFT — PO raised in parallel while PR is being processed
# Supplier: Ceylon Spice Traders | Payment: NET30 | Tax: VAT00
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PO-001  DRAFT — Monthly Spice Restock (linked to PR-001)"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-orders" '{
  "supplier_id":     '"$CEYLON_ID"',
  "pr_id":           '"$PR1_ID"',
  "order_date":      "2026-06-09",
  "expected_date":   "2026-06-23",
  "currency_id":     '"$LKR_ID"',
  "exchange_rate":   1,
  "payment_term_id": '"$NET30_ID"',
  "warehouse_id":    '"$WH1_ID"',
  "notes": "Linked to PR-001. Prepared in advance — pending PR approval. Ceylon Spice Traders confirmed availability."
}')
PO1=$(idof "$r")
echo "  Created PO-001 id=$PO1 status=$(ok "$r")"

post "procurement/purchase-orders/$PO1/items" '{
  "product_id":   '"$CINN_ID"',
  "quantity":     500,
  "uom_id":       '"$KG_ID"',
  "unit_price":   1450,
  "discount_pct": 2,
  "tax_code_id":  '"$VAT00_ID"',
  "description":  "Grade A Ceylon Cinnamon Sticks — export quality",
  "notes":        "2% early-order discount negotiated"
}' > /dev/null

post "procurement/purchase-orders/$PO1/items" '{
  "product_id":   '"$BLKP_ID"',
  "quantity":     200,
  "uom_id":       '"$KG_ID"',
  "unit_price":   2750,
  "discount_pct": 0,
  "tax_code_id":  '"$VAT00_ID"',
  "description":  "Whole Black Pepper — export grade, moisture <12%",
  "notes":        "Must include phytosanitary certificate"
}' > /dev/null

echo "  Added 2 lines: Cinnamon 500kg @ LKR 1,450 (-2%) + Black Pepper 200kg @ LKR 2,750"
echo "  Totals: LKR ~1,262,000 subtotal"
echo "  Final status: DRAFT"

# ══════════════════════════════════════════════════════════════════════════════
# PO-002 — CANCELLED: Packaging Replenishment (linked to PR-002)
# PO created after PR was submitted; supplier raised prices 15% — cancelled
# Supplier: Lanka Packaging Solutions | Payment: NET15
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PO-002  CANCELLED — Packaging Replenishment (linked to PR-002)"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-orders" '{
  "supplier_id":     '"$LANKA_ID"',
  "pr_id":           '"$PR2_ID"',
  "order_date":      "2026-06-08",
  "expected_date":   "2026-06-17",
  "currency_id":     '"$LKR_ID"',
  "exchange_rate":   1,
  "payment_term_id": '"$NET15_ID"',
  "warehouse_id":    '"$WH1_ID"',
  "notes": "Linked to PR-002. Initial PO raised urgently. CANCELLED — Lanka Packaging revised quote 15% above agreed rate. Sourcing alternate supplier."
}')
PO2=$(idof "$r")
echo "  Created PO-002 id=$PO2"

post "procurement/purchase-orders/$PO2/items" '{
  "product_id":   '"$JAR_ID"',
  "quantity":     3000,
  "uom_id":       '"$PCS_ID"',
  "unit_price":   51.75,
  "discount_pct": 0,
  "tax_code_id":  '"$VAT08_ID"',
  "description":  "Glass jars 250g — food grade with airtight lid",
  "notes":        "Revised price: LKR 51.75 (originally quoted 45.00)"
}' > /dev/null

post "procurement/purchase-orders/$PO2/items" '{
  "product_id":   '"$CINNP_ID"',
  "quantity":     1500,
  "uom_id":       '"$PKT_ID"',
  "unit_price":   28.75,
  "discount_pct": 0,
  "tax_code_id":  '"$VAT08_ID"',
  "description":  "Pre-printed retail pouches 250g — design v3.2",
  "notes":        "Price increase rejected. New supplier needed."
}' > /dev/null

echo "  Added 2 lines: Jars 3000 PCS @ LKR 51.75 + Pouches 1500 PKT @ LKR 28.75"

r=$(post "procurement/purchase-orders/$PO2/cancel" '{}')
echo "  Cancelled → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# PO-003 — CONFIRMED: Cloves for UK Export Order (linked to PR-003 APPROVED)
# The ideal flow: APPROVED PR → PO issued → PO confirmed
# Confirming automatically creates a draft Purchase Invoice
# Supplier: Galle Cloves Estate | Payment: ADV50 | Tax: VAT00
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PO-003  CONFIRMED — Cloves for UK Export (linked to PR-003 APPROVED)"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-orders" '{
  "supplier_id":     '"$GALLE_ID"',
  "pr_id":           '"$PR3_ID"',
  "order_date":      "2026-06-06",
  "expected_date":   "2026-06-13",
  "currency_id":     '"$LKR_ID"',
  "exchange_rate":   1,
  "payment_term_id": '"$ADV50_ID"',
  "warehouse_id":    '"$WH1_ID"',
  "notes": "Linked to PR-003 (APPROVED). Time-critical — vessel departs June 22 for UK. 50% advance payment issued."
}')
PO3=$(idof "$r")
echo "  Created PO-003 id=$PO3"

post "procurement/purchase-orders/$PO3/items" '{
  "product_id":   '"$CLOV_ID"',
  "quantity":     300,
  "uom_id":       '"$KG_ID"',
  "unit_price":   3100,
  "discount_pct": 0,
  "tax_code_id":  '"$VAT00_ID"',
  "description":  "Grade A Whole Cloves — UK export standard, moisture <12%",
  "notes":        "Lab cert required. Supplier confirmed stock available at estate."
}' > /dev/null

echo "  Added 1 line: Cloves 300kg @ LKR 3,100  (Total: LKR 930,000)"

r=$(post "procurement/purchase-orders/$PO3/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"
echo "  NOTE: Draft Purchase Invoice auto-created by system upon PO confirmation"

# ══════════════════════════════════════════════════════════════════════════════
# PO-004 — DRAFT: New Product Raw Materials (linked to PR-006)
# PR is DRAFT pending budget — PO prepared speculatively for lead time reasons
# Supplier: Indian Spice Exporters | Payment: NET30 | Tax: VAT00
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PO-004  DRAFT — New Product Raw Materials (linked to PR-006)"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-orders" '{
  "supplier_id":     '"$INDIAN_ID"',
  "pr_id":           '"$PR6_ID"',
  "order_date":      "2026-06-09",
  "expected_date":   "2026-06-30",
  "currency_id":     '"$LKR_ID"',
  "exchange_rate":   1,
  "payment_term_id": '"$NET30_ID"',
  "warehouse_id":    '"$WH1_ID"',
  "notes": "Linked to PR-006 (DRAFT — awaiting budget). PO prepared early due to 3-week lead time from Indian Spice Exporters. Do not confirm until PR is approved."
}')
PO4=$(idof "$r")
echo "  Created PO-004 id=$PO4"

post "procurement/purchase-orders/$PO4/items" '{
  "product_id":   '"$CARDM_ID"',
  "quantity":     150,
  "uom_id":       '"$KG_ID"',
  "unit_price":   4350,
  "discount_pct": 0,
  "tax_code_id":  '"$VAT00_ID"',
  "description":  "Green Cardamom Pods Grade 7mm+ — new product line",
  "notes":        "Supplier sample lot dispatched — pending quality approval"
}' > /dev/null

post "procurement/purchase-orders/$PO4/items" '{
  "product_id":   '"$TURM_ID"',
  "quantity":     250,
  "uom_id":       '"$KG_ID"',
  "unit_price":   650,
  "discount_pct": 0,
  "tax_code_id":  '"$VAT00_ID"',
  "description":  "Fresh Turmeric Rhizomes — curcumin min 3%, lab cert required",
  "notes":        "Negotiate on price if ordering >300kg"
}' > /dev/null

echo "  Added 2 lines: Cardamom 150kg @ LKR 4,350 + Turmeric 250kg @ LKR 650"
echo "  Totals: LKR ~815,000 subtotal"
echo "  Final status: DRAFT"

# ══════════════════════════════════════════════════════════════════════════════
# Verify — list all POs and check invoice was auto-created for PO-003
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " VERIFICATION"
echo "════════════════════════════════════════"

PO_LIST=$(get "procurement/purchase-orders")
echo "  All Purchase Orders:"
echo "$PO_LIST" | jq -r '.data[] | "  id=\(.id)  \(.code)  status=\(.status)  pr_id=\(.pr_id // "-")  total=LKR\(.total_amount)"'

echo ""
INV_LIST=$(get "procurement/purchase-invoices")
echo "  Purchase Invoices (auto-created on PO confirm):"
echo "$INV_LIST" | jq -r '.data[] | "  id=\(.id)  \(.code)  status=\(.status)  po_id=\(.po_id)  total=LKR\(.total_amount)"'

echo ""
echo "════════════════════════════════════════"
echo " SUMMARY"
echo "════════════════════════════════════════"
echo "  PO-001 id=$PO1  DRAFT     ← PR-001(DRAFT)    Cinnamon 500kg + Black Pepper 200kg  Ceylon Spice Traders"
echo "  PO-002 id=$PO2  CANCELLED ← PR-002(PENDING)  Glass Jars 3000 + Pouches 1500       Lanka Packaging (price dispute)"
echo "  PO-003 id=$PO3  CONFIRMED ← PR-003(APPROVED) Cloves 300kg                          Galle Cloves Estate + auto-invoice"
echo "  PO-004 id=$PO4  DRAFT     ← PR-006(DRAFT)    Cardamom 150kg + Turmeric 250kg      Indian Spice Exporters"
echo ""
echo "  ALL DONE"
