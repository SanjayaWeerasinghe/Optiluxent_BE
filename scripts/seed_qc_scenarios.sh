#!/bin/sh
# QC Scenarios Seed — Kadahapola Exports Ltd
# Covers 6 QC outcomes across the 3 trigger sources:
#   1. GRN WITH_PO       → MATERIAL_QC fully PASSED
#   2. GRN WITH_PO       → MATERIAL_QC fully FAILED (entire batch rejected)
#   3. GRN WITHOUT_PO    → MATERIAL_QC PARTIAL (some pass, some fail)
#   4. GRN WITHOUT_PO    → MATERIAL_QC fully PASSED (ad-hoc receipt)
#   5. Production Output → PRODUCT_QC fully PASSED
#   6. Production Output → PRODUCT_QC fully FAILED (label defect, needs rework)
#
# Run: docker exec -e SEED_PASSWORD=Admin@1234 erp-api-dev sh /app/scripts/seed_qc_scenarios.sh

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

post() { curl -s -X POST  "$BASE/$1" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$2"; }
put()  { curl -s -X PUT   "$BASE/$1" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$2"; }
get()  { curl -s          "$BASE/$1" -H "Authorization: Bearer $TOKEN"; }
idof() { echo "$1" | jq -r '.data.id // empty'; }

# Cloves stock helper for verification (Cloves = product_id 4 in WH001 = 1)
stockOf() {
  PROD=$1; WH=$2
  get "inventory/stock?warehouse_id=${WH}&product_id=${PROD}" | jq -r '[.data[]? | .quantity] | add // 0'
}

# Find the auto-QC for a GRN
qcFor() {
  REF_TYPE=$1; REF_ID=$2
  get "inventory/quality-checks" | jq --argjson r "$REF_ID" --arg t "$REF_TYPE" \
    -r '.data[] | select(.reference_type == $t and .reference_id == $r) | .id' | head -1
}

# Grade a QC line and submit the QC
gradeAndSubmit() {
  QC_ID=$1; PRODUCT_ID=$2; CHECKED=$3; PASSED=$4; FAILED=$5; RESULT=$6; REASON=$7
  LINE_ID=$(get "inventory/quality-checks/$QC_ID" | jq -r '.data.lines[0].id')
  put "inventory/quality-checks/$QC_ID/items/$LINE_ID" "{
    \"product_id\":$PRODUCT_ID,
    \"qty_checked\":$CHECKED,
    \"qty_passed\":$PASSED,
    \"qty_failed\":$FAILED,
    \"result\":\"$RESULT\",
    \"rejection_reason\":\"$REASON\"
  }" > /dev/null
  post "inventory/quality-checks/$QC_ID/start" '{}' > /dev/null
  post "inventory/quality-checks/$QC_ID/submit" '{}' > /dev/null
}

# ── Re-fetch IDs ───────────────────────────────────────────────────────────────
echo ">>> Fetching IDs..."
PROD_LIST=$(get "masterdata/products?per_page=100")
CINN_ID=$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CINN-STCK") | .id')
BLKP_ID=$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="BLKPEP")    | .id')
CLOV_ID=$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CLOVES")    | .id')
TEA_ID=$(echo "$PROD_LIST"  | jq -r '.data[] | select(.code=="CEYTEA")    | .id')
CINNP_ID=$(echo "$PROD_LIST"| jq -r '.data[] | select(.code=="CINN-PWD-250") | .id')
CARDM_ID=$(echo "$PROD_LIST"| jq -r '.data[] | select(.code=="CARDM")     | .id')

CEYLON_SUPP=2   # SUPP001 Ceylon Spice Traders
INDIAN_SUPP=8   # SUPP004 Indian Spice Exporters
MATARA_SUPP=6   # SUPP002 Matara Spice Growers
GALLE_SUPP=43   # SUPP005 Galle Cloves Estate

WH1=1; WH2=2
LOC_A1=2     # Rack A1 Dry Spices (WH1)
LOC_FG=7     # FG location in WH2
KG=1; PKT=8

echo "  Products: CINN=$CINN_ID BLKP=$BLKP_ID CLOV=$CLOV_ID TEA=$TEA_ID CINNP=$CINNP_ID CARDM=$CARDM_ID"

# ══════════════════════════════════════════════════════════════════════════════
# Scenario 1 — GRN WITH_PO → MATERIAL_QC fully PASSED (Cinnamon 100 kg)
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " S1 — GRN WITH_PO → QC PASSED (Cinnamon 100 kg)"
echo "════════════════════════════════════════"
BEFORE=$(stockOf $CINN_ID $WH1)
echo "  Cinnamon stock BEFORE: $BEFORE kg"

GRN=$(post "procurement/goods-receipts" "{
  \"grn_type\":\"WITH_PO\",\"po_id\":8,\"supplier_id\":$CEYLON_SUPP,
  \"receipt_date\":\"2026-06-28\",\"warehouse_id\":$WH1,
  \"notes\":\"S1: Ceylon Spice cinnamon delivery, clean batch expected\"
}" | jq -r '.data.id')
post "procurement/goods-receipts/$GRN/items" "{
  \"po_line_id\":10,\"product_id\":$CINN_ID,\"quantity\":100,\"uom_id\":$KG,
  \"location_id\":$LOC_A1,\"unit_cost\":1450,
  \"notes\":\"Sealed and labeled correctly\"
}" > /dev/null
post "procurement/goods-receipts/$GRN/confirm" '{}' > /dev/null
echo "  GRN id=$GRN confirmed; stock after confirm: $(stockOf $CINN_ID $WH1) kg (gated)"

QC=$(qcFor "GRN" "$GRN")
gradeAndSubmit "$QC" "$CINN_ID" 100 100 0 "PASSED" ""
echo "  QC id=$QC graded: 100 PASSED / 0 FAILED → SUBMITTED"
echo "  Stock AFTER QC: $(stockOf $CINN_ID $WH1) kg  (expected $((BEFORE + 100)))"

# ══════════════════════════════════════════════════════════════════════════════
# Scenario 2 — GRN WITH_PO → MATERIAL_QC fully FAILED (Black Pepper 80 kg)
# Supplier shipped mold-contaminated batch; entire delivery rejected
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " S2 — GRN WITH_PO → QC FAILED (Black Pepper 80 kg, mold contamination)"
echo "════════════════════════════════════════"
BEFORE=$(stockOf $BLKP_ID $WH1)
echo "  Black Pepper stock BEFORE: $BEFORE kg"

GRN=$(post "procurement/goods-receipts" "{
  \"grn_type\":\"WITH_PO\",\"po_id\":8,\"supplier_id\":$CEYLON_SUPP,
  \"receipt_date\":\"2026-06-28\",\"warehouse_id\":$WH1,
  \"notes\":\"S2: Pepper batch — physical inspection raised concerns\"
}" | jq -r '.data.id')
post "procurement/goods-receipts/$GRN/items" "{
  \"po_line_id\":11,\"product_id\":$BLKP_ID,\"quantity\":80,\"uom_id\":$KG,
  \"location_id\":$LOC_A1,\"unit_cost\":2750,
  \"notes\":\"Held at QC bay for inspection\"
}" > /dev/null
post "procurement/goods-receipts/$GRN/confirm" '{}' > /dev/null
echo "  GRN id=$GRN confirmed; stock after confirm: $(stockOf $BLKP_ID $WH1) kg (gated)"

QC=$(qcFor "GRN" "$GRN")
gradeAndSubmit "$QC" "$BLKP_ID" 80 0 80 "FAILED" "Mold contamination detected throughout batch; visible white spots on 60% of berries, moisture 15.2% (max 12%). Entire batch rejected, returning to supplier."
echo "  QC id=$QC graded: 0 PASSED / 80 FAILED → SUBMITTED"
echo "  Stock AFTER QC: $(stockOf $BLKP_ID $WH1) kg  (expected $BEFORE — no stock movement)"

# ══════════════════════════════════════════════════════════════════════════════
# Scenario 3 — GRN WITHOUT_PO → MATERIAL_QC PARTIAL (Cardamom 50 kg sample)
# Trader sample 35 pass / 15 fail
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " S3 — GRN WITHOUT_PO → QC PARTIAL (Cardamom 50 kg sample, 35 pass / 15 fail)"
echo "════════════════════════════════════════"
BEFORE=$(stockOf $CARDM_ID $WH1)
echo "  Cardamom stock BEFORE: $BEFORE kg"

GRN=$(post "procurement/goods-receipts" "{
  \"grn_type\":\"WITHOUT_PO\",\"supplier_id\":$INDIAN_SUPP,
  \"receipt_date\":\"2026-06-28\",\"warehouse_id\":$WH1,
  \"notes\":\"S3: Indian Spice trader sample — evaluation lot, no PO\"
}" | jq -r '.data.id')
post "procurement/goods-receipts/$GRN/items" "{
  \"product_id\":$CARDM_ID,\"quantity\":50,\"uom_id\":$KG,
  \"location_id\":$LOC_A1,\"unit_cost\":4350,
  \"notes\":\"Free sample for grade evaluation\"
}" > /dev/null
post "procurement/goods-receipts/$GRN/confirm" '{}' > /dev/null
echo "  GRN id=$GRN confirmed; stock after confirm: $(stockOf $CARDM_ID $WH1) kg (gated)"

QC=$(qcFor "GRN" "$GRN")
gradeAndSubmit "$QC" "$CARDM_ID" 50 35 15 "PARTIAL" "15kg below Grade 7mm spec (avg 5.8mm). 35kg meets spec and accepted; smaller pods returned."
echo "  QC id=$QC graded: 35 PASSED / 15 FAILED → SUBMITTED"
echo "  Stock AFTER QC: $(stockOf $CARDM_ID $WH1) kg  (expected $((BEFORE + 35)))"

# ══════════════════════════════════════════════════════════════════════════════
# Scenario 4 — GRN WITHOUT_PO → MATERIAL_QC fully PASSED (Tea 30 kg)
# Supplier-provided sample for evaluation, accepted
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " S4 — GRN WITHOUT_PO → QC PASSED (Tea 30 kg supplier sample)"
echo "════════════════════════════════════════"
BEFORE=$(stockOf $TEA_ID $WH1)
echo "  Ceylon Tea stock BEFORE: $BEFORE kg"

GRN=$(post "procurement/goods-receipts" "{
  \"grn_type\":\"WITHOUT_PO\",\"supplier_id\":$MATARA_SUPP,
  \"receipt_date\":\"2026-06-28\",\"warehouse_id\":$WH1,
  \"notes\":\"S4: Matara Spice new harvest sample, no PO\"
}" | jq -r '.data.id')
post "procurement/goods-receipts/$GRN/items" "{
  \"product_id\":$TEA_ID,\"quantity\":30,\"uom_id\":$KG,
  \"location_id\":$LOC_A1,\"unit_cost\":1850,
  \"notes\":\"New season tea — first plucking\"
}" > /dev/null
post "procurement/goods-receipts/$GRN/confirm" '{}' > /dev/null
echo "  GRN id=$GRN confirmed; stock after confirm: $(stockOf $TEA_ID $WH1) kg (gated)"

QC=$(qcFor "GRN" "$GRN")
gradeAndSubmit "$QC" "$TEA_ID" 30 30 0 "PASSED" ""
echo "  QC id=$QC graded: 30 PASSED / 0 FAILED → SUBMITTED"
echo "  Stock AFTER QC: $(stockOf $TEA_ID $WH1) kg  (expected $((BEFORE + 30)))"

# ══════════════════════════════════════════════════════════════════════════════
# Scenario 5 — Production Output → PRODUCT_QC fully PASSED (Cinn Powder 200 PKT)
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " S5 — Production Output → QC PASSED (Cinnamon Powder Packs 200)"
echo "════════════════════════════════════════"
BEFORE=$(stockOf $CINNP_ID $WH2)
echo "  Cinnamon Pack stock BEFORE: $BEFORE PKT"

ORD=$(post "manufacturing/orders" "{
  \"product_id\":$CINNP_ID,\"uom_id\":$PKT,\"planned_qty\":200,
  \"warehouse_id\":$WH2,\"notes\":\"S5: scheduled production run\"
}" | jq -r '.data.id')
OUT=$(post "manufacturing/orders/$ORD/outputs" "{
  \"product_id\":$CINNP_ID,\"uom_id\":$PKT,\"quantity\":200,\"unit_cost\":205,
  \"warehouse_id\":$WH2,\"location_id\":$LOC_FG,
  \"notes\":\"Batch CIN-PWD-2026-06-S5\"
}" | jq -r '.data.id')
echo "  Order id=$ORD, Output id=$OUT (200 PKT); stock after add: $(stockOf $CINNP_ID $WH2) PKT (gated)"

QC=$(qcFor "PRODUCTION_OUTPUT" "$OUT")
gradeAndSubmit "$QC" "$CINNP_ID" 200 200 0 "PASSED" ""
echo "  QC id=$QC graded: 200 PASSED / 0 FAILED → SUBMITTED"
echo "  Stock AFTER QC: $(stockOf $CINNP_ID $WH2) PKT  (expected $((BEFORE + 200)))"

# ══════════════════════════════════════════════════════════════════════════════
# Scenario 6 — Production Output → PRODUCT_QC fully FAILED (Cinn Powder 100 PKT)
# Label print misalignment; entire batch held for relabeling
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " S6 — Production Output → QC FAILED (Cinnamon Powder Packs 100, label defect)"
echo "════════════════════════════════════════"
BEFORE=$(stockOf $CINNP_ID $WH2)
echo "  Cinnamon Pack stock BEFORE: $BEFORE PKT"

ORD=$(post "manufacturing/orders" "{
  \"product_id\":$CINNP_ID,\"uom_id\":$PKT,\"planned_qty\":100,
  \"warehouse_id\":$WH2,\"notes\":\"S6: production run, label issue mid-run\"
}" | jq -r '.data.id')
OUT=$(post "manufacturing/orders/$ORD/outputs" "{
  \"product_id\":$CINNP_ID,\"uom_id\":$PKT,\"quantity\":100,\"unit_cost\":205,
  \"warehouse_id\":$WH2,\"location_id\":$LOC_FG,
  \"notes\":\"Batch CIN-PWD-2026-06-S6 — held for inspection\"
}" | jq -r '.data.id')
echo "  Order id=$ORD, Output id=$OUT (100 PKT); stock after add: $(stockOf $CINNP_ID $WH2) PKT (gated)"

QC=$(qcFor "PRODUCTION_OUTPUT" "$OUT")
gradeAndSubmit "$QC" "$CINNP_ID" 100 0 100 "FAILED" "Label printer misaligned mid-run. 100 packs have batch code printed off-position, regulatory non-compliance for export. Holding for relabel batch next week."
echo "  QC id=$QC graded: 0 PASSED / 100 FAILED → SUBMITTED"
echo "  Stock AFTER QC: $(stockOf $CINNP_ID $WH2) PKT  (expected $BEFORE — no stock movement)"

# ══════════════════════════════════════════════════════════════════════════════
# Final state summary
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " QC STATE SUMMARY"
echo "════════════════════════════════════════"
get "inventory/quality-checks" | jq -r '.data[0:10][] | "  \(.code)  type=\(.qc_type)  ref=\(.reference_type)/\(.reference_id)  status=\(.status)"'

echo ""
echo "════════════════════════════════════════"
echo " SCENARIO RECAP"
echo "════════════════════════════════════════"
echo "  S1  GRN WITH_PO    → QC PASSED   Cinnamon  100/100 → +100 stock"
echo "  S2  GRN WITH_PO    → QC FAILED   Pepper      0/80  → no stock (mold)"
echo "  S3  GRN WITHOUT_PO → QC PARTIAL  Cardamom   35/50  → +35 stock"
echo "  S4  GRN WITHOUT_PO → QC PASSED   Tea        30/30  → +30 stock"
echo "  S5  Production     → QC PASSED   Cinn Pkt 200/200  → +200 stock"
echo "  S6  Production     → QC FAILED   Cinn Pkt   0/100  → no stock (label defect)"
echo ""
echo "  ALL DONE"
