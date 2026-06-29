#!/bin/sh
# Sales Seed — Kadahapola Exports Ltd
# Full SO → DO → SI lifecycle across 6 scenarios covering all statuses:
#   SO-001 DRAFT                — UK export quote pending customer signoff
#   SO-002 CONFIRMED no DO      — AU bulk order, awaiting production
#   SO-003 PARTIAL delivery     — Local distributor with 1 partial DO
#   SO-004 DELIVERED + POSTED   — Full cycle, invoice posted, unpaid
#   SO-005 DELIVERED + PARTIAL  — Full cycle with partial payment recorded
#   SO-006 CANCELLED            — Was confirmed, then cancelled
#
# Run: docker exec -e SEED_PASSWORD=Admin@1234 erp-api-dev sh /app/scripts/seed_sales.sh

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
ok()   { echo "$1" | jq -r '.data.status // .data.id // .message // empty'; }

# ── Re-fetch IDs ───────────────────────────────────────────────────────────────
echo ">>> Fetching master data IDs..."

PROD=$(get "masterdata/products?per_page=100")
CINNP_ID=$(echo "$PROD" | jq -r '.data[] | select(.code=="CINN-PWD-250") | .id')
TEA_ID=$(echo "$PROD"   | jq -r '.data[] | select(.code=="CEYTEA")       | .id')
CINN_ID=$(echo "$PROD"  | jq -r '.data[] | select(.code=="CINN-STCK")    | .id')
CLOV_ID=$(echo "$PROD"  | jq -r '.data[] | select(.code=="CLOVES")       | .id')
TURM_ID=$(echo "$PROD"  | jq -r '.data[] | select(.code=="TURMRIC")      | .id')

UOM=$(get "masterdata/products/uoms?per_page=20")
KG=$(echo "$UOM"  | jq -r '.data[] | select(.code=="KG")  | .id')
PKT=$(echo "$UOM" | jq -r '.data[] | select(.code=="PKT") | .id')

CURR=$(get "masterdata/financial/currencies?per_page=20")
LKR=$(echo "$CURR" | jq -r '.data[] | select(.code=="LKR") | .id')
USD=$(echo "$CURR" | jq -r '.data[] | select(.code=="USD") | .id')
GBP=$(echo "$CURR" | jq -r '.data[] | select(.code=="GBP") | .id')

PT=$(get "masterdata/financial/payment-terms?per_page=20")
NET30=$(echo "$PT" | jq -r '.data[] | select(.code=="NET30") | .id')
NET60=$(echo "$PT" | jq -r '.data[] | select(.code=="NET60") | .id')
IMM=$(echo "$PT"   | jq -r '.data[] | select(.code=="IMM")   | .id')
ADV50=$(echo "$PT" | jq -r '.data[] | select(.code=="ADV50") | .id')

TAX=$(get "masterdata/financial/tax-codes?per_page=20")
VAT00=$(echo "$TAX" | jq -r '.data[] | select(.code=="VAT00") | .id')
VAT08=$(echo "$TAX" | jq -r '.data[] | select(.code=="VAT08") | .id')

CUST=$(get "masterdata/contacts/parties?party_type=CUSTOMER&per_page=20")
CUST1=$(echo "$CUST" | jq -r '.data[] | select(.code=="CUST001") | .id')
CUST2=$(echo "$CUST" | jq -r '.data[] | select(.code=="CUST002") | .id')
CUST3=$(echo "$CUST" | jq -r '.data[] | select(.code=="CUST003") | .id')
CUST4=$(echo "$CUST" | jq -r '.data[] | select(.code=="CUST004") | .id')
CUST5=$(echo "$CUST" | jq -r '.data[] | select(.code=="CUST005") | .id')
CUST6=$(echo "$CUST" | jq -r '.data[] | select(.code=="CUST006") | .id')

WH=$(get "masterdata/inventory/warehouses?per_page=20")
WH2=$(echo "$WH" | jq -r '.data[] | select(.code=="WH002") | .id')

# FG location in WH002
WH2LOC=$(get "masterdata/inventory/warehouses/$WH2/locations" 2>/dev/null \
  | jq -r '.data[0].id // empty')
if [ -z "$WH2LOC" ] || [ "$WH2LOC" = "null" ]; then WH2LOC=6; fi

echo "  Products  CINNP=$CINNP_ID TEA=$TEA_ID CINN=$CINN_ID CLOV=$CLOV_ID TURM=$TURM_ID"
echo "  UOMs      KG=$KG PKT=$PKT"
echo "  Currency  LKR=$LKR USD=$USD GBP=$GBP"
echo "  Payment   NET30=$NET30 NET60=$NET60 IMM=$IMM ADV50=$ADV50"
echo "  Tax       VAT00=$VAT00 VAT08=$VAT08"
echo "  Customers C1=$CUST1 C2=$CUST2 C3=$CUST3 C4=$CUST4 C5=$CUST5 C6=$CUST6"
echo "  Warehouse WH2=$WH2 loc=$WH2LOC"

# ══════════════════════════════════════════════════════════════════════════════
# SO-001 — DRAFT: UK export quote (CUST005, GBP)
# Sales rep prepared the order; awaiting customer signoff
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " SO-001  DRAFT — UK export quote (London Herb Co)"
echo "════════════════════════════════════════"

r=$(post "sales/sales-orders" '{
  "customer_id":            '"$CUST5"',
  "order_date":             "2026-06-09",
  "expected_delivery_date": "2026-07-05",
  "currency_id":            '"$GBP"',
  "exchange_rate":          385.5,
  "payment_term_id":        '"$NET30"',
  "warehouse_id":           '"$WH2"',
  "notes": "Quote for London Herb Co Q3 order. Awaiting customer PO signoff before confirmation."
}')
SO1=$(idof "$r"); echo "  Created SO-001 id=$SO1 status=$(ok "$r")"

post "sales/sales-orders/$SO1/items" '{"product_id":'"$CINNP_ID"',"quantity":2000,"uom_id":'"$PKT"',"unit_price":1.85,"tax_code_id":'"$VAT00"',"description":"Cinnamon Powder Pack 250g — UK retail","notes":"FOB Colombo. EUR pricing pending currency hedge."}' > /dev/null
post "sales/sales-orders/$SO1/items" '{"product_id":'"$TEA_ID"',"quantity":100,"uom_id":'"$KG"',"unit_price":12.50,"tax_code_id":'"$VAT00"',"description":"Ceylon OP Grade Tea — bulk for repackaging","notes":""}' > /dev/null
echo "  Added: Cinnamon Powder 2000 PKT @ GBP 1.85 + Ceylon Tea 100 kg @ GBP 12.50"
echo "  Final status: DRAFT"

# ══════════════════════════════════════════════════════════════════════════════
# SO-002 — CONFIRMED no DO: AU bulk order (CUST006, USD)
# Customer confirmed; awaiting production batch before delivery
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " SO-002  CONFIRMED — AU bulk order (Natural Foods Australia)"
echo "════════════════════════════════════════"

r=$(post "sales/sales-orders" '{
  "customer_id":            '"$CUST6"',
  "order_date":             "2026-06-04",
  "expected_delivery_date": "2026-07-20",
  "currency_id":            '"$USD"',
  "exchange_rate":          305.0,
  "payment_term_id":        '"$NET60"',
  "warehouse_id":           '"$WH2"',
  "notes": "Bulk health food order for AU market. Awaiting cinnamon powder production batch."
}')
SO2=$(idof "$r"); echo "  Created SO-002 id=$SO2"

post "sales/sales-orders/$SO2/items" '{"product_id":'"$CINNP_ID"',"quantity":5000,"uom_id":'"$PKT"',"unit_price":2.10,"tax_code_id":'"$VAT00"',"description":"Cinnamon Powder Pack 250g — AU retail","notes":"AQIS-compliant labeling required"}' > /dev/null
post "sales/sales-orders/$SO2/items" '{"product_id":'"$TURM_ID"',"quantity":300,"uom_id":'"$KG"',"unit_price":8.50,"tax_code_id":'"$VAT00"',"description":"Turmeric Powder — bulk","notes":"Min curcumin 3%"}' > /dev/null
echo "  Added: Cinnamon Powder 5000 PKT + Turmeric 300 kg"

r=$(post "sales/sales-orders/$SO2/confirm" '{}')
echo "  Confirmed → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# SO-003 — PARTIAL: Local distributor (CUST001, LKR)
# Order confirmed → first DO delivered partial; 2nd batch pending
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " SO-003  PARTIAL delivery — Local distributor (CUST001)"
echo "════════════════════════════════════════"

r=$(post "sales/sales-orders" '{
  "customer_id":            '"$CUST1"',
  "order_date":             "2026-06-05",
  "expected_delivery_date": "2026-06-20",
  "currency_id":            '"$LKR"',
  "exchange_rate":          1,
  "payment_term_id":        '"$NET30"',
  "warehouse_id":           '"$WH2"',
  "notes": "Cargills supermarket bulk order. Split delivery: 600 first week, 400 following week."
}')
SO3=$(idof "$r"); echo "  Created SO-003 id=$SO3"

SO3_LINE_R=$(post "sales/sales-orders/$SO3/items" '{"product_id":'"$CINNP_ID"',"quantity":1000,"uom_id":'"$PKT"',"unit_price":280,"tax_code_id":'"$VAT08"',"description":"Cinnamon Powder Pack 250g — local retail","notes":""}')
SO3_LINE=$(echo "$SO3_LINE_R" | jq -r '.data.id')
echo "  Added line id=$SO3_LINE: Cinnamon Powder 1000 PKT @ LKR 280"

post "sales/sales-orders/$SO3/confirm" '{}' > /dev/null
echo "  SO confirmed"

# DO #1 — partial 600 of 1000 pkts
DO_R=$(post "sales/deliveries" '{
  "so_id":         '"$SO3"',
  "customer_id":   '"$CUST1"',
  "delivery_date": "2026-06-12",
  "warehouse_id":  '"$WH2"',
  "notes": "First batch — 600 of 1000 pkts. Remaining 400 to ship next week."
}')
DO1=$(idof "$DO_R")
post "sales/deliveries/$DO1/items" '{"so_line_id":'"$SO3_LINE"',"product_id":'"$CINNP_ID"',"quantity":600,"uom_id":'"$PKT"',"location_id":'"$WH2LOC"',"unit_cost":210,"notes":"Loaded on truck CL-3447. DN# CGS-2026-06-A."}' > /dev/null
r=$(post "sales/deliveries/$DO1/confirm" '{}')
echo "  DO-1 (600 PKT) confirmed → status: $(ok "$r")"
echo "  SO-003 outstanding: 400 PKT (next delivery pending)"

# ══════════════════════════════════════════════════════════════════════════════
# SO-004 — DELIVERED + POSTED: Full cycle (CUST002, LKR)
# SO → confirm → DO full → confirm DO → SI → add lines → post SI (unpaid)
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " SO-004  DELIVERED + POSTED invoice — Full cycle (CUST002)"
echo "════════════════════════════════════════"

r=$(post "sales/sales-orders" '{
  "customer_id":            '"$CUST2"',
  "order_date":             "2026-06-01",
  "expected_delivery_date": "2026-06-10",
  "currency_id":            '"$LKR"',
  "exchange_rate":          1,
  "payment_term_id":        '"$NET30"',
  "warehouse_id":           '"$WH2"',
  "notes": "Keells Super monthly order. Full cycle complete."
}')
SO4=$(idof "$r"); echo "  Created SO-004 id=$SO4"

SO4_LINE_R=$(post "sales/sales-orders/$SO4/items" '{"product_id":'"$CINNP_ID"',"quantity":200,"uom_id":'"$PKT"',"unit_price":275,"tax_code_id":'"$VAT08"',"description":"Cinnamon Powder Pack 250g","notes":""}')
SO4_LINE=$(echo "$SO4_LINE_R" | jq -r '.data.id')

post "sales/sales-orders/$SO4/confirm" '{}' > /dev/null
echo "  SO confirmed, line id=$SO4_LINE"

DO_R=$(post "sales/deliveries" '{"so_id":'"$SO4"',"customer_id":'"$CUST2"',"delivery_date":"2026-06-09","warehouse_id":'"$WH2"',"notes":"Full delivery of all 200 pkts. DN# KLS-2026-06-A."}')
DO2=$(idof "$DO_R")
DO2_LINE_R=$(post "sales/deliveries/$DO2/items" '{"so_line_id":'"$SO4_LINE"',"product_id":'"$CINNP_ID"',"quantity":200,"uom_id":'"$PKT"',"location_id":'"$WH2LOC"',"unit_cost":205,"notes":""}')
DO2_LINE=$(echo "$DO2_LINE_R" | jq -r '.data.id')
post "sales/deliveries/$DO2/confirm" '{}' > /dev/null
echo "  DO confirmed (200 PKT full), do_line_id=$DO2_LINE"

SI_R=$(post "sales/invoices" '{
  "customer_id":         '"$CUST2"',
  "so_id":               '"$SO4"',
  "invoice_date":        "2026-06-10",
  "due_date":            "2026-07-10",
  "customer_po_number":  "KLS-PO-2026-0410",
  "currency_id":         '"$LKR"',
  "exchange_rate":       1,
  "payment_term_id":     '"$NET30"',
  "notes": "Invoice for SO-004 full delivery"
}')
SI4=$(idof "$SI_R"); echo "  Created SI id=$SI4 status=$(ok "$SI_R")"

post "sales/invoices/$SI4/lines" '{"do_line_id":'"$DO2_LINE"',"product_id":'"$CINNP_ID"',"quantity":200,"uom_id":'"$PKT"',"unit_price":275,"tax_code_id":'"$VAT08"',"description":"Cinnamon Powder Pack 250g"}' > /dev/null
echo "  Added invoice line: 200 PKT @ LKR 275 = LKR 55,000 + VAT8%"

r=$(post "sales/invoices/$SI4/post" '{}')
echo "  Invoice posted → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# SO-005 — DELIVERED + PARTIAL PAYMENT: Big customer (CUST003, LKR)
# Full cycle + 50% partial payment recorded → SI status PARTIAL
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " SO-005  DELIVERED + PARTIAL PAID — Full cycle with 50% payment"
echo "════════════════════════════════════════"

r=$(post "sales/sales-orders" '{
  "customer_id":            '"$CUST3"',
  "order_date":             "2026-05-25",
  "expected_delivery_date": "2026-06-05",
  "currency_id":            '"$LKR"',
  "exchange_rate":          1,
  "payment_term_id":        '"$NET30"',
  "warehouse_id":           '"$WH2"',
  "notes": "Arpico bulk monthly order. Full cycle + 50% advance received."
}')
SO5=$(idof "$r"); echo "  Created SO-005 id=$SO5"

SO5_L1_R=$(post "sales/sales-orders/$SO5/items" '{"product_id":'"$CINNP_ID"',"quantity":500,"uom_id":'"$PKT"',"unit_price":275,"tax_code_id":'"$VAT08"',"description":"Cinnamon Powder Pack 250g"}')
SO5_L1=$(echo "$SO5_L1_R" | jq -r '.data.id')
SO5_L2_R=$(post "sales/sales-orders/$SO5/items" '{"product_id":'"$TEA_ID"',"quantity":80,"uom_id":'"$KG"',"unit_price":1800,"tax_code_id":'"$VAT08"',"description":"Ceylon Tea — bulk for repackaging"}')
SO5_L2=$(echo "$SO5_L2_R" | jq -r '.data.id')

post "sales/sales-orders/$SO5/confirm" '{}' > /dev/null
echo "  SO confirmed, lines: $SO5_L1, $SO5_L2"

DO_R=$(post "sales/deliveries" '{"so_id":'"$SO5"',"customer_id":'"$CUST3"',"delivery_date":"2026-06-03","warehouse_id":'"$WH2"',"notes":"Full delivery. DN# ARP-2026-06-A."}')
DO3=$(idof "$DO_R")
DO3_L1=$(post "sales/deliveries/$DO3/items" '{"so_line_id":'"$SO5_L1"',"product_id":'"$CINNP_ID"',"quantity":500,"uom_id":'"$PKT"',"location_id":'"$WH2LOC"',"unit_cost":205}' | jq -r '.data.id')
DO3_L2=$(post "sales/deliveries/$DO3/items" '{"so_line_id":'"$SO5_L2"',"product_id":'"$TEA_ID"',"quantity":80,"uom_id":'"$KG"',"location_id":'"$WH2LOC"',"unit_cost":1300}' | jq -r '.data.id')
post "sales/deliveries/$DO3/confirm" '{}' > /dev/null
echo "  DO confirmed (500 PKT + 80 kg)"

SI_R=$(post "sales/invoices" '{
  "customer_id":'"$CUST3"',"so_id":'"$SO5"',"invoice_date":"2026-06-04","due_date":"2026-07-04",
  "customer_po_number":"ARP-PO-2026-0405","currency_id":'"$LKR"',"exchange_rate":1,"payment_term_id":'"$NET30"',
  "notes":"Invoice with 50% advance applied"
}')
SI5=$(idof "$SI_R")
post "sales/invoices/$SI5/lines" '{"do_line_id":'"$DO3_L1"',"product_id":'"$CINNP_ID"',"quantity":500,"uom_id":'"$PKT"',"unit_price":275,"tax_code_id":'"$VAT08"'}' > /dev/null
post "sales/invoices/$SI5/lines" '{"do_line_id":'"$DO3_L2"',"product_id":'"$TEA_ID"',"quantity":80,"uom_id":'"$KG"',"unit_price":1800,"tax_code_id":'"$VAT08"'}' > /dev/null
echo "  Created SI id=$SI5 with 2 lines"

post "sales/invoices/$SI5/post" '{}' > /dev/null
echo "  Invoice posted"

# Total: 500*275 + 80*1800 = 137,500 + 144,000 = 281,500 + 8% VAT = 22,520 = 304,020
# 50% partial payment = LKR 152,010
r=$(post "sales/invoices/$SI5/pay" '{"amount": 152010}')
echo "  Recorded 50% partial payment LKR 152,010 → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# SO-006 — CANCELLED: was confirmed then cancelled (CUST004)
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " SO-006  CANCELLED — Confirmed then cancelled (CUST004)"
echo "════════════════════════════════════════"

r=$(post "sales/sales-orders" '{
  "customer_id":            '"$CUST4"',
  "order_date":             "2026-06-02",
  "expected_delivery_date": "2026-06-25",
  "currency_id":            '"$LKR"',
  "exchange_rate":          1,
  "payment_term_id":        '"$IMM"',
  "warehouse_id":           '"$WH2"',
  "notes": "Customer initially placed order, then cancelled due to inventory glut on their side."
}')
SO6=$(idof "$r"); echo "  Created SO-006 id=$SO6"

post "sales/sales-orders/$SO6/items" '{"product_id":'"$TEA_ID"',"quantity":50,"uom_id":'"$KG"',"unit_price":1900,"tax_code_id":'"$VAT08"',"description":"Ceylon Tea — speculative order","notes":""}' > /dev/null
echo "  Added: Ceylon Tea 50 kg @ LKR 1,900"

post "sales/sales-orders/$SO6/confirm" '{}' > /dev/null
echo "  SO confirmed"

r=$(post "sales/sales-orders/$SO6/cancel" '{}')
echo "  Cancelled → status: $(ok "$r")"

# ══════════════════════════════════════════════════════════════════════════════
# Final state
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo "════════════════════════════════════════"
echo " FINAL STATE"
echo "════════════════════════════════════════"
echo ""
echo "Sales Orders:"
get "sales/sales-orders" | jq -r '.data[] | "  id=\(.id)  \(.code)  customer=\(.customer_id)  status=\(.status)  total=\(.currency_id):\(.total_amount)"'
echo ""
echo "Delivery Orders:"
get "sales/deliveries" | jq -r '.data[] | "  id=\(.id)  \(.code)  so_id=\(.so_id // "-")  status=\(.status)  date=\(.delivery_date[0:10])"'
echo ""
echo "Sales Invoices:"
get "sales/invoices" | jq -r '.data[] | "  id=\(.id)  \(.code)  so_id=\(.so_id)  status=\(.status)  total=\(.total_amount)  paid=\(.paid_amount)"'

echo ""
echo "════════════════════════════════════════"
echo " SUMMARY"
echo "════════════════════════════════════════"
echo "  SO-001  DRAFT                — UK quote (London Herb Co, GBP)         pending signoff"
echo "  SO-002  CONFIRMED no DO      — AU bulk (Natural Foods Australia, USD) awaiting production"
echo "  SO-003  PARTIAL              — Local (CUST001, LKR)                   600/1000 delivered"
echo "  SO-004  DELIVERED + POSTED   — Local (CUST002, LKR)                   full cycle, unpaid"
echo "  SO-005  DELIVERED + PARTIAL  — Local (CUST003, LKR)                   full cycle, 50% paid"
echo "  SO-006  CANCELLED            — Local (CUST004, LKR)                   was confirmed, then cancelled"
echo ""
echo "  ALL DONE"
