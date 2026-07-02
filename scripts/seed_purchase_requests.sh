#!/bin/sh
# Purchase Request Seed — Kadahapola Exports Ltd
# Creates 6 PRs covering: DRAFT, PENDING_APPROVAL, APPROVED, REJECTED, CANCELLED
# Run: docker exec -e SEED_PASSWORD=Admin@1234 erp-api-dev sh /app/scripts/seed_purchase_requests.sh

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

post()  { curl -s -X POST  "$BASE/$1" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$2"; }
get()   { curl -s          "$BASE/$1" -H "Authorization: Bearer $TOKEN"; }
idof()  { echo "$1" | jq -r '.data.id // empty'; }
ok()    { echo "$1" | jq -r '.data.status // .data.id // empty'; }

# ── Re-fetch IDs ───────────────────────────────────────────────────────────────
echo ">>> Fetching master data IDs..."

PROD_LIST=$(get "masterdata/products?per_page=100")
CINN_ID=$(echo "$PROD_LIST"  | jq -r '.data[] | select(.code=="CINN-STCK")  | .id')
BLKP_ID=$(echo "$PROD_LIST"  | jq -r '.data[] | select(.code=="BLKPEP")     | .id')
CLOV_ID=$(echo "$PROD_LIST"  | jq -r '.data[] | select(.code=="CLOVES")     | .id')
TEA_ID=$(echo "$PROD_LIST"   | jq -r '.data[] | select(.code=="CEYTEA")     | .id')
OIL_ID=$(echo "$PROD_LIST"   | jq -r '.data[] | select(.code=="COCO-OIL")   | .id')
JAR_ID=$(echo "$PROD_LIST"   | jq -r '.data[] | select(.code=="JAR-250")    | .id')
CINNP_ID=$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CINN-PWD-250")| .id')
CARDM_ID=$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CARDM")      | .id')
TURM_ID=$(echo "$PROD_LIST"  | jq -r '.data[] | select(.code=="TURMRIC")    | .id')

UOM_LIST=$(get "masterdata/products/uoms?per_page=20")
KG_ID=$(echo "$UOM_LIST"  | jq -r '.data[] | select(.code=="KG")  | .id')
L_ID=$(echo "$UOM_LIST"   | jq -r '.data[] | select(.code=="L")   | .id')
PCS_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="PCS") | .id')
PKT_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="PKT") | .id')

CURR_LIST=$(get "masterdata/financial/currencies?per_page=20")
LKR_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="LKR") | .id')

DEPT_LIST=$(get "masterdata/organization/departments?per_page=20")
PUR_DEPT=$(echo "$DEPT_LIST"  | jq -r '.data[] | select(.code=="PUR")   | .id')
PROD_DEPT=$(echo "$DEPT_LIST" | jq -r '.data[] | select(.code=="PROD")  | .id')
SALES_DEPT=$(echo "$DEPT_LIST"| jq -r '.data[] | select(.code=="SALES") | .id')
IT_DEPT=$(echo "$DEPT_LIST"   | jq -r '.data[] | select(.code=="IT")    | .id')

echo "  Products  CINN=$CINN_ID BLKP=$BLKP_ID CLOV=$CLOV_ID TEA=$TEA_ID OIL=$OIL_ID JAR=$JAR_ID CINNP=$CINNP_ID CARDM=$CARDM_ID TURM=$TURM_ID"
echo "  UOMs      KG=$KG_ID L=$L_ID PCS=$PCS_ID PKT=$PKT_ID"
echo "  Currency  LKR=$LKR_ID"
echo "  Depts     PUR=$PUR_DEPT PROD=$PROD_DEPT SALES=$SALES_DEPT IT=$IT_DEPT"

# ══════════════════════════════════════════════════════════════════════════════
# PR-001 — DRAFT: Monthly Spice Restock
# Status: Left as DRAFT — created, items added, not yet submitted
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PR-001  DRAFT — Monthly Spice Restock"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-requests" '{
  "request_date":  "2026-06-09",
  "required_date": "2026-06-25",
  "department_id": '"$PUR_DEPT"',
  "notes": "Routine monthly restocking for export processing — Cinnamon & Black Pepper"
}')
PR1=$(idof "$r")
echo "  Created PR-001 id=$PR1 status=$(ok "$r")"

post "procurement/purchase-requests/$PR1/items" '{
  "product_id": '"$CINN_ID"', "quantity": 500, "uom_id": '"$KG_ID"',
  "estimated_price": 1500, "currency_id": '"$LKR_ID"',
  "description": "Grade A Cinnamon Sticks for export batch",
  "notes": "Preferred supplier: Ceylon Spice Traders"
}' > /dev/null

post "procurement/purchase-requests/$PR1/items" '{
  "product_id": '"$BLKP_ID"', "quantity": 200, "uom_id": '"$KG_ID"',
  "estimated_price": 2800, "currency_id": '"$LKR_ID"',
  "description": "Whole Black Pepper — export grade",
  "notes": "Check quality cert before delivery"
}' > /dev/null

echo "  Added 2 items (Cinnamon 500kg + Black Pepper 200kg)"
echo "  Final status: DRAFT"

# ══════════════════════════════════════════════════════════════════════════════
# PR-002 — PENDING_APPROVAL: Packaging Replenishment
# Status: Created, items added, then submitted → PENDING_APPROVAL
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PR-002  PENDING_APPROVAL — Packaging Replenishment"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-requests" '{
  "request_date":  "2026-06-08",
  "required_date": "2026-06-18",
  "department_id": '"$PROD_DEPT"',
  "notes": "Urgent — Q3 packaging stock below safety level. Production line will halt by week 3 without this."
}')
PR2=$(idof "$r")
echo "  Created PR-002 id=$PR2"

post "procurement/purchase-requests/$PR2/items" '{
  "product_id": '"$JAR_ID"', "quantity": 3000, "uom_id": '"$PCS_ID"',
  "estimated_price": 45, "currency_id": '"$LKR_ID"',
  "description": "Glass jar 250g — for cinnamon powder packing",
  "notes": "Must match spec: food-grade, airtight lid"
}' > /dev/null

post "procurement/purchase-requests/$PR2/items" '{
  "product_id": '"$CINNP_ID"', "quantity": 1500, "uom_id": '"$PKT_ID"',
  "estimated_price": 25, "currency_id": '"$LKR_ID"',
  "description": "Pre-printed retail pouches for cinnamon powder 250g",
  "notes": "Design v3.2 — confirm artwork before print run"
}' > /dev/null

echo "  Added 2 items (Glass Jars 3000 PCS + Pouches 1500 PKT)"

r=$(post "procurement/purchase-requests/$PR2/submit" '{}')
echo "  Submitted → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# PR-003 — APPROVED: Cloves for UK Export Order
# Status: Created → submitted → approved
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PR-003  APPROVED — Cloves for UK Export Order"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-requests" '{
  "request_date":  "2026-06-05",
  "required_date": "2026-06-15",
  "department_id": '"$PUR_DEPT"',
  "notes": "London Herb Co UK order LHC-2026-047 requires 300kg Grade A Cloves. Time-sensitive — vessel departs June 22."
}')
PR3=$(idof "$r")
echo "  Created PR-003 id=$PR3"

post "procurement/purchase-requests/$PR3/items" '{
  "product_id": '"$CLOV_ID"', "quantity": 300, "uom_id": '"$KG_ID"',
  "estimated_price": 3200, "currency_id": '"$LKR_ID"',
  "description": "Grade A Whole Cloves — UK export standard",
  "notes": "Supplier: Galle Cloves Estate. Must pass moisture test <12%."
}' > /dev/null

echo "  Added 1 item (Cloves 300kg)"

post "procurement/purchase-requests/$PR3/submit" '{}' > /dev/null
r=$(post "procurement/purchase-requests/$PR3/approve" '{}')
echo "  Submitted + Approved → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# PR-004 — REJECTED: Ceylon Tea — Stock Already Sufficient
# Status: Created → submitted → rejected
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PR-004  REJECTED — Ceylon Tea Overstock Request"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-requests" '{
  "request_date":  "2026-06-07",
  "required_date": "2026-07-01",
  "department_id": '"$SALES_DEPT"',
  "notes": "Sales team requesting advance stock build-up for anticipated Q3 demand. 500kg Ceylon Tea."
}')
PR4=$(idof "$r")
echo "  Created PR-004 id=$PR4"

post "procurement/purchase-requests/$PR4/items" '{
  "product_id": '"$TEA_ID"', "quantity": 500, "uom_id": '"$KG_ID"',
  "estimated_price": 1200, "currency_id": '"$LKR_ID"',
  "description": "Ceylon OP Grade Tea — speculative stock build",
  "notes": "Based on projected Q3 order pipeline"
}' > /dev/null

echo "  Added 1 item (Ceylon Tea 500kg)"

post "procurement/purchase-requests/$PR4/submit" '{}' > /dev/null
r=$(post "procurement/purchase-requests/$PR4/reject" '{}')
echo "  Submitted + Rejected → status: $(ok "$r")"
echo "  Rejection reason: Current WH001 stock is 1,200kg — no purchase needed until stock falls below 300kg"

# ══════════════════════════════════════════════════════════════════════════════
# PR-005 — CANCELLED: Coconut Oil Batch — Supplier Delayed
# Status: Created → submitted → cancelled
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PR-005  CANCELLED — Coconut Oil Batch (Supplier Delayed)"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-requests" '{
  "request_date":  "2026-06-03",
  "required_date": "2026-06-12",
  "department_id": '"$PUR_DEPT"',
  "notes": "Planned coconut oil procurement for June production batch. Supplier confirmed delay — cancelling and deferring to July."
}')
PR5=$(idof "$r")
echo "  Created PR-005 id=$PR5"

post "procurement/purchase-requests/$PR5/items" '{
  "product_id": '"$OIL_ID"', "quantity": 200, "uom_id": '"$L_ID"',
  "estimated_price": 800, "currency_id": '"$LKR_ID"',
  "description": "Virgin Coconut Oil — food grade, for production",
  "notes": "Matara Spice Growers Coop — preferred supplier"
}' > /dev/null

post "procurement/purchase-requests/$PR5/items" '{
  "product_id": '"$CINN_ID"', "quantity": 100, "uom_id": '"$KG_ID"',
  "estimated_price": 1500, "currency_id": '"$LKR_ID"',
  "description": "Cinnamon Sticks to accompany oil batch",
  "notes": "Same delivery consignment"
}' > /dev/null

echo "  Added 2 items (Coconut Oil 200L + Cinnamon 100kg)"

post "procurement/purchase-requests/$PR5/submit" '{}' > /dev/null
r=$(post "procurement/purchase-requests/$PR5/cancel" '{}')
echo "  Submitted + Cancelled → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# PR-006 — DRAFT: New Product Raw Materials (multi-item, multi-dept)
# Status: Left as DRAFT — new product lines, awaiting budget approval
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " PR-006  DRAFT — New Product Raw Materials"
echo "════════════════════════════════════════"

r=$(post "procurement/purchase-requests" '{
  "request_date":  "2026-06-09",
  "required_date": "2026-06-30",
  "department_id": '"$PUR_DEPT"',
  "notes": "Initial raw material procurement for new product lines: Cardamom and Turmeric. Pending board budget confirmation — do not submit until finance signs off."
}')
PR6=$(idof "$r")
echo "  Created PR-006 id=$PR6"

post "procurement/purchase-requests/$PR6/items" '{
  "product_id": '"$CARDM_ID"', "quantity": 150, "uom_id": '"$KG_ID"',
  "estimated_price": 4500, "currency_id": '"$LKR_ID"',
  "description": "Green Cardamom Pods — Grade 7mm+ for new export line",
  "notes": "Source: local Matale growers cooperative — get samples first"
}' > /dev/null

post "procurement/purchase-requests/$PR6/items" '{
  "product_id": '"$TURM_ID"', "quantity": 250, "uom_id": '"$KG_ID"',
  "estimated_price": 680, "currency_id": '"$LKR_ID"',
  "description": "Fresh Turmeric Rhizomes for drying and grinding",
  "notes": "Ensure curcumin content min 3% — request lab cert"
}' > /dev/null

echo "  Added 2 items (Cardamom 150kg + Turmeric 250kg)"
echo "  Final status: DRAFT (awaiting budget sign-off)"

# ══════════════════════════════════════════════════════════════════════════════
# Summary
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " SUMMARY"
echo "════════════════════════════════════════"
echo "  PR-001 id=$PR1  DRAFT              — Monthly Spice Restock (Cinnamon + Black Pepper)"
echo "  PR-002 id=$PR2  PENDING_APPROVAL   — Packaging Replenishment (Glass Jars + Pouches)"
echo "  PR-003 id=$PR3  APPROVED           — Cloves for UK Export Order"
echo "  PR-004 id=$PR4  REJECTED           — Ceylon Tea Overstock Request"
echo "  PR-005 id=$PR5  CANCELLED          — Coconut Oil Batch (Supplier Delayed)"
echo "  PR-006 id=$PR6  DRAFT              — New Product Raw Materials (Cardamom + Turmeric)"
echo ""
echo "  ALL DONE"
