#!/bin/sh
# Master Data Seed Script — Kadahapola Exports Ltd
# Run inside the erp-api-dev container: sh /app/scripts/seed_masterdata.sh

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
put()  { curl -s -X PUT  "$BASE/$1" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$2"; }
get()  { curl -s "$BASE/$1" -H "Authorization: Bearer $TOKEN"; }
id()   { echo "$1" | jq -r '.data.id // .data[0].id // empty' 2>/dev/null; }

echo ""
echo "════════════════════════════════════════"
echo " 1. CURRENCIES"
echo "════════════════════════════════════════"
r=$(post "masterdata/financial/currencies" '{"code":"EUR","name":"Euro","symbol":"€"}')
echo "  EUR: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/currencies" '{"code":"GBP","name":"British Pound","symbol":"£"}')
echo "  GBP: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/currencies" '{"code":"SGD","name":"Singapore Dollar","symbol":"S$"}')
echo "  SGD: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/currencies" '{"code":"AED","name":"UAE Dirham","symbol":"AED"}')
echo "  AED: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/currencies" '{"code":"INR","name":"Indian Rupee","symbol":"Rs"}')
echo "  INR: $(echo $r | jq -r '.data.id // .error.message')"

# Always re-fetch all currency IDs by code (handles both first run and repeat runs)
CURR_LIST=$(get "masterdata/financial/currencies")
LKR_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="LKR") | .id')
USD_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="USD") | .id')
EUR_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="EUR") | .id')
GBP_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="GBP") | .id')
SGD_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="SGD") | .id')
AED_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="AED") | .id')
INR_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="INR") | .id')
echo "  LKR=$LKR_ID  USD=$USD_ID  EUR=$EUR_ID  GBP=$GBP_ID  SGD=$SGD_ID  AED=$AED_ID  INR=$INR_ID"

echo ""
echo "════════════════════════════════════════"
echo " 2. PAYMENT TERMS"
echo "════════════════════════════════════════"
r=$(post "masterdata/financial/payment-terms" '{"code":"IMM","name":"Immediate","due_days":0,"discount_days":0,"discount_percent":0}')
echo "  IMM: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/payment-terms" '{"code":"NET15","name":"Net 15 Days","due_days":15,"discount_days":0,"discount_percent":0}')
echo "  NET15: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/payment-terms" '{"code":"NET60","name":"Net 60 Days","due_days":60,"discount_days":0,"discount_percent":0}')
echo "  NET60: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/payment-terms" '{"code":"ADV50","name":"50% Advance","due_days":30,"discount_days":0,"discount_percent":0}')
echo "  ADV50: $(echo $r | jq -r '.data.id // .error.message')"

PT_LIST=$(get "masterdata/financial/payment-terms")
IMM_PT=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="IMM") | .id')
NET15_PT=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="NET15") | .id')
NET30_ID=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="NET30") | .id')
NET60_PT=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="NET60") | .id')
ADV50_PT=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="ADV50") | .id')
echo "  IMM=$IMM_PT NET15=$NET15_PT NET30=$NET30_ID NET60=$NET60_PT ADV50=$ADV50_PT"

echo ""
echo "════════════════════════════════════════"
echo " 3. CHART OF ACCOUNTS"
echo "════════════════════════════════════════"
r=$(post "masterdata/financial/chart-of-accounts" '{"code":"1100","name":"Trade Receivables","account_type":"ASSET","currency_id":'"$LKR_ID"'}')
echo "  AR: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/chart-of-accounts" '{"code":"1200","name":"Inventory Asset","account_type":"ASSET","currency_id":'"$LKR_ID"'}')
echo "  Inventory: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/chart-of-accounts" '{"code":"2100","name":"Trade Payables","account_type":"LIABILITY","currency_id":'"$LKR_ID"'}')
echo "  AP: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/chart-of-accounts" '{"code":"4000","name":"Sales Revenue","account_type":"REVENUE","currency_id":'"$LKR_ID"'}')
echo "  Revenue: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/chart-of-accounts" '{"code":"5000","name":"Cost of Goods Sold","account_type":"EXPENSE","currency_id":'"$LKR_ID"'}')
echo "  COGS: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/chart-of-accounts" '{"code":"6100","name":"Salary Expense","account_type":"EXPENSE","currency_id":'"$LKR_ID"'}')
echo "  Salary: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/chart-of-accounts" '{"code":"2300","name":"VAT Output","account_type":"LIABILITY","currency_id":'"$LKR_ID"'}')
echo "  VAT Out: $(echo $r | jq -r '.data.id // .error.message')"

COA_LIST=$(get "masterdata/financial/chart-of-accounts")
CASH_ID=$(echo "$COA_LIST" | jq -r '.data[] | select(.code=="1000") | .id')
VATOUT_ID=$(echo "$COA_LIST" | jq -r '.data[] | select(.code=="2300") | .id')
AR_ID=$(echo "$COA_LIST" | jq -r '.data[] | select(.code=="1100") | .id')
AP_ID=$(echo "$COA_LIST" | jq -r '.data[] | select(.code=="2100") | .id')
echo "  Cash=$CASH_ID  VATOUT=$VATOUT_ID  AR=$AR_ID  AP=$AP_ID"

echo ""
echo "════════════════════════════════════════"
echo " 4. TAX CODES"
echo "════════════════════════════════════════"
r=$(post "masterdata/financial/tax-codes" '{"code":"VAT00","name":"VAT 0% (Export)","tax_type":"VAT","rate":0,"gl_account_id":'"$VATOUT_ID"'}')
echo "  VAT0: $(echo $r | jq -r '.data.id // .error.message')"; VAT0_ID=$(id "$r")
r=$(post "masterdata/financial/tax-codes" '{"code":"VAT08","name":"VAT 8%","tax_type":"VAT","rate":8,"gl_account_id":'"$VATOUT_ID"'}')
echo "  VAT8: $(echo $r | jq -r '.data.id // .error.message')"; VAT8_ID=$(id "$r")

TC_LIST=$(get "masterdata/financial/tax-codes")
VAT18_ID=$(echo "$TC_LIST" | jq -r '.data[] | select(.code=="VAT18") | .id')
VAT0_ID=${VAT0_ID:-$(echo "$TC_LIST" | jq -r '.data[] | select(.code=="VAT00") | .id')}
VAT8_ID=${VAT8_ID:-$(echo "$TC_LIST" | jq -r '.data[] | select(.code=="VAT08") | .id')}
echo "  VAT18=$VAT18_ID  VAT0=$VAT0_ID  VAT8=$VAT8_ID"

echo ""
echo "════════════════════════════════════════"
echo " 5. TAX GROUPS"
echo "════════════════════════════════════════"
r=$(post "masterdata/financial/tax-groups" '{"name":"Standard VAT","tax_code_ids":['"$VAT18_ID"']}')
echo "  Standard: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/tax-groups" '{"name":"Export Zero Rate","tax_code_ids":['"$VAT0_ID"']}')
echo "  Export: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " 6. BANKS & BANK ACCOUNTS"
echo "════════════════════════════════════════"
r=$(post "masterdata/financial/banks" '{"name":"Hatton National Bank","branch_name":"Colombo Main","swift_code":"HBLILKLX","address":"HNB Towers, Colombo 03"}')
echo "  HNB: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/banks" '{"name":"Sampath Bank","branch_name":"Kandy Branch","swift_code":"BSAMLKLX","address":"Sampath Centre, Kandy"}')
echo "  Sampath: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/banks" '{"name":"Commercial Bank","branch_name":"Matara Branch","swift_code":"CCEYLKLX","address":"Commercial House, Matara"}')
echo "  ComBank: $(echo $r | jq -r '.data.id // .error.message')"

BANK_LIST=$(get "masterdata/financial/banks")
BOC_ID=$(echo "$BANK_LIST" | jq -r '.data[] | select(.name | startswith("Bank of Ceylon")) | .id' | head -1)
HNB_ID=$(echo "$BANK_LIST" | jq -r '.data[] | select(.name == "Hatton National Bank") | .id' | head -1)
SAMPATH_ID=$(echo "$BANK_LIST" | jq -r '.data[] | select(.name == "Sampath Bank") | .id' | head -1)
COM_BANK_ID=$(echo "$BANK_LIST" | jq -r '.data[] | select(.name == "Commercial Bank") | .id' | head -1)
echo "  BOC=$BOC_ID HNB=$HNB_ID Sampath=$SAMPATH_ID"

r=$(post "masterdata/financial/bank-accounts" '{"bank_id":'"$BOC_ID"',"account_number":"7001234567","account_name":"Kadahapola Exports - LKR Current","currency_id":'"$LKR_ID"',"gl_account_id":'"$CASH_ID"',"is_default":true}')
echo "  BOC LKR Acct: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/bank-accounts" '{"bank_id":'"$HNB_ID"',"account_number":"HNB-001-USD-2025","account_name":"Kadahapola Exports - USD Account","currency_id":'"$USD_ID"',"gl_account_id":'"$CASH_ID"',"is_default":false}')
echo "  HNB USD Acct: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " 7. EXCHANGE RATES"
echo "════════════════════════════════════════"
r=$(post "masterdata/financial/exchange-rates" '{"from_currency_id":'"$USD_ID"',"to_currency_id":'"$LKR_ID"',"rate":320.50,"effective_date":"2026-01-01T00:00:00Z","source":"Central Bank"}')
echo "  USD/LKR: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/exchange-rates" '{"from_currency_id":'"$EUR_ID"',"to_currency_id":'"$LKR_ID"',"rate":345.75,"effective_date":"2026-01-01T00:00:00Z","source":"Central Bank"}')
echo "  EUR/LKR: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/exchange-rates" '{"from_currency_id":'"$GBP_ID"',"to_currency_id":'"$LKR_ID"',"rate":405.20,"effective_date":"2026-01-01T00:00:00Z","source":"Central Bank"}')
echo "  GBP/LKR: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/exchange-rates" '{"from_currency_id":'"$SGD_ID"',"to_currency_id":'"$LKR_ID"',"rate":238.40,"effective_date":"2026-01-01T00:00:00Z","source":"Central Bank"}')
echo "  SGD/LKR: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/exchange-rates" '{"from_currency_id":'"$AED_ID"',"to_currency_id":'"$LKR_ID"',"rate":87.30,"effective_date":"2026-01-01T00:00:00Z","source":"Central Bank"}')
echo "  AED/LKR: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " 8. DEPARTMENTS & COST CENTERS"
echo "════════════════════════════════════════"
r=$(post "masterdata/organization/departments" '{"code":"QA","name":"Quality Assurance"}')
echo "  QA Dept: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/organization/departments" '{"code":"PUR","name":"Procurement"}')
echo "  Procurement: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/organization/departments" '{"code":"ADMIN","name":"Administration"}')
echo "  Admin: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/organization/departments" '{"code":"IT","name":"Information Technology"}')
echo "  IT: $(echo $r | jq -r '.data.id // .error.message')"

DEPT_LIST=$(get "masterdata/organization/departments")
QA_DEPT=$(echo "$DEPT_LIST" | jq -r '.data[] | select(.code=="QA") | .id')
PUR_DEPT=$(echo "$DEPT_LIST" | jq -r '.data[] | select(.code=="PUR") | .id')
FIN_DEPT=$(echo "$DEPT_LIST" | jq -r '.data[] | select(.code=="FIN") | .id')
PROD_DEPT=$(echo "$DEPT_LIST" | jq -r '.data[] | select(.code=="PROD") | .id')
SALES_DEPT=$(echo "$DEPT_LIST" | jq -r '.data[] | select(.code=="SALES") | .id')
echo "  QA=$QA_DEPT PUR=$PUR_DEPT FIN=$FIN_DEPT PROD=$PROD_DEPT SALES=$SALES_DEPT"

r=$(post "masterdata/financial/cost-centers" '{"code":"CC-FIN","name":"Finance Cost Center","department_id":'"$FIN_DEPT"'}')
echo "  Finance CC: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/cost-centers" '{"code":"CC-PROD","name":"Production Cost Center","department_id":'"$PROD_DEPT"'}')
echo "  Production CC: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/financial/cost-centers" '{"code":"CC-SALES","name":"Sales Cost Center","department_id":'"$SALES_DEPT"'}')
echo "  Sales CC: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " 9. FISCAL YEAR & DOC SEQUENCES"
echo "════════════════════════════════════════"
r=$(post "masterdata/organization/fiscal-years" '{"name":"FY 2025-2026","start_date":"2025-04-01T00:00:00Z","end_date":"2026-03-31T00:00:00Z"}')
echo "  FY25-26: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/organization/fiscal-years" '{"name":"FY 2026-2027","start_date":"2026-04-01T00:00:00Z","end_date":"2027-03-31T00:00:00Z"}')
echo "  FY26-27: $(echo $r | jq -r '.data.id // .error.message')"

r=$(post "masterdata/organization/document-sequences" '{"document_type":"PURCHASE_ORDER","prefix":"PO","next_number":1001,"padding":5,"suffix":""}')
echo "  PO Seq: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/organization/document-sequences" '{"document_type":"SALES_ORDER","prefix":"SO","next_number":1001,"padding":5,"suffix":""}')
echo "  SO Seq: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/organization/document-sequences" '{"document_type":"SALES_INVOICE","prefix":"INV","next_number":1001,"padding":5,"suffix":""}')
echo "  INV Seq: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/organization/document-sequences" '{"document_type":"GOODS_RECEIPT","prefix":"GRN","next_number":1001,"padding":5,"suffix":""}')
echo "  GRN Seq: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/organization/document-sequences" '{"document_type":"GOODS_ISSUE","prefix":"GI","next_number":1001,"padding":5,"suffix":""}')
echo "  GI Seq: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " 10. UNITS OF MEASURE"
echo "════════════════════════════════════════"
UOM_LIST=$(get "masterdata/products/uoms")
KG_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="KG") | .id')
echo "  KG=$KG_ID"
r=$(post "masterdata/products/uoms" '{"code":"G","name":"Gram","uom_type":"WEIGHT","base_uom_id":'"$KG_ID"',"conversion_factor":0.001}')
echo "  G: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/products/uoms" '{"code":"MT","name":"Metric Ton","uom_type":"WEIGHT","base_uom_id":'"$KG_ID"',"conversion_factor":1000}')
echo "  MT: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/products/uoms" '{"code":"L","name":"Litre","uom_type":"VOLUME","conversion_factor":1}')
echo "  L: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/products/uoms" '{"code":"PCS","name":"Pieces","uom_type":"UNIT","conversion_factor":1}')
echo "  PCS: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/products/uoms" '{"code":"BOX","name":"Box","uom_type":"UNIT","conversion_factor":1}')
echo "  BOX: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/products/uoms" '{"code":"M","name":"Metre","uom_type":"LENGTH","conversion_factor":1}')
echo "  M: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/products/uoms" '{"code":"PKT","name":"Packet","uom_type":"UNIT","conversion_factor":1}')
echo "  PKT: $(echo $r | jq -r '.data.id // .error.message')"

# Re-fetch all UOM IDs
UOM_LIST=$(get "masterdata/products/uoms")
KG_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="KG") | .id')
G_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="G") | .id')
MT_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="MT") | .id')
L_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="L") | .id')
PCS_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="PCS") | .id')
BOX_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="BOX") | .id')
M_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="M") | .id')
PKT_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="PKT") | .id')
echo "  KG=$KG_ID G=$G_ID MT=$MT_ID L=$L_ID PCS=$PCS_ID BOX=$BOX_ID PKT=$PKT_ID"

# Add ML after L is confirmed
r=$(post "masterdata/products/uoms" '{"code":"ML","name":"Millilitre","uom_type":"VOLUME","base_uom_id":'"$L_ID"',"conversion_factor":0.001}')
echo "  ML: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " 11. PRODUCT CATEGORIES"
echo "════════════════════════════════════════"
SPICE_CAT=$(get "masterdata/products/categories" | jq -r '.data[] | select(.code=="SPICE") | .id')
echo "  Spices=$SPICE_CAT"
r=$(post "masterdata/products/categories" '{"code":"TEA","name":"Teas & Infusions"}')
echo "  Teas: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/products/categories" '{"code":"OILS","name":"Essential & Carrier Oils"}')
echo "  Oils: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/products/categories" '{"code":"PKG","name":"Packaging Materials"}')
echo "  Packaging: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/products/categories" '{"code":"FG","name":"Finished Goods"}')
echo "  Finished: $(echo $r | jq -r '.data.id // .error.message')"

PCAT_LIST=$(get "masterdata/products/categories")
SPICE_CAT=$(echo "$PCAT_LIST" | jq -r '.data[] | select(.code=="SPICE") | .id')
TEA_CAT=$(echo "$PCAT_LIST" | jq -r '.data[] | select(.code=="TEA") | .id')
OILS_CAT=$(echo "$PCAT_LIST" | jq -r '.data[] | select(.code=="OILS") | .id')
PKG_CAT=$(echo "$PCAT_LIST" | jq -r '.data[] | select(.code=="PKG") | .id')
FG_CAT=$(echo "$PCAT_LIST" | jq -r '.data[] | select(.code=="FG") | .id')
echo "  SPICE=$SPICE_CAT TEA=$TEA_CAT OILS=$OILS_CAT PKG=$PKG_CAT FG=$FG_CAT"

echo ""
echo "════════════════════════════════════════"
echo " 12. PRODUCTS"
echo "════════════════════════════════════════"
r=$(post "masterdata/products" '{"code":"CINN-STCK","name":"Ceylon Cinnamon Sticks","product_type":"FINISHED","category_id":'"$SPICE_CAT"',"base_uom_id":'"$KG_ID"',"purchase_uom_id":'"$KG_ID"',"sales_uom_id":'"$KG_ID"',"cost_price":850,"standard_price":1200,"is_purchased":true,"is_sold":true,"is_manufactured":false}')
echo "  Cinnamon Sticks: $(echo $r | jq -r '.data.id // .error.message')"; CINN_STCK=$(id "$r")
r=$(post "masterdata/products" '{"code":"BLKPEP","name":"Black Pepper Whole","product_type":"FINISHED","category_id":'"$SPICE_CAT"',"base_uom_id":'"$KG_ID"',"purchase_uom_id":'"$KG_ID"',"sales_uom_id":'"$KG_ID"',"cost_price":920,"standard_price":1400,"is_purchased":true,"is_sold":true,"is_manufactured":false}')
echo "  Black Pepper: $(echo $r | jq -r '.data.id // .error.message')"; BLKPEP_ID=$(id "$r")
r=$(post "masterdata/products" '{"code":"CLOVES","name":"Cloves Whole","product_type":"FINISHED","category_id":'"$SPICE_CAT"',"base_uom_id":'"$KG_ID"',"purchase_uom_id":'"$KG_ID"',"sales_uom_id":'"$KG_ID"',"cost_price":2800,"standard_price":3800,"is_purchased":true,"is_sold":true,"is_manufactured":false}')
echo "  Cloves: $(echo $r | jq -r '.data.id // .error.message')"; CLOVES_ID=$(id "$r")
r=$(post "masterdata/products" '{"code":"CEYTEA","name":"Ceylon Black Tea BOP","product_type":"FINISHED","category_id":'"$TEA_CAT"',"base_uom_id":'"$KG_ID"',"purchase_uom_id":'"$KG_ID"',"sales_uom_id":'"$KG_ID"',"cost_price":650,"standard_price":950,"is_purchased":true,"is_sold":true,"is_manufactured":false}')
echo "  Ceylon Tea: $(echo $r | jq -r '.data.id // .error.message')"; TEA_ID=$(id "$r")
r=$(post "masterdata/products" '{"code":"COCO-OIL","name":"Virgin Coconut Oil","product_type":"FINISHED","category_id":'"$OILS_CAT"',"base_uom_id":'"$L_ID"',"purchase_uom_id":'"$L_ID"',"sales_uom_id":'"$L_ID"',"cost_price":480,"standard_price":780,"is_purchased":true,"is_sold":true,"is_manufactured":false}')
echo "  Coconut Oil: $(echo $r | jq -r '.data.id // .error.message')"; COCOIL_ID=$(id "$r")
r=$(post "masterdata/products" '{"code":"JAR-250","name":"Glass Jar 250g","product_type":"CONSUMABLE","category_id":'"$PKG_CAT"',"base_uom_id":'"$PCS_ID"',"purchase_uom_id":'"$BOX_ID"',"sales_uom_id":'"$PCS_ID"',"cost_price":45,"standard_price":65,"is_purchased":true,"is_sold":false,"is_manufactured":false}')
echo "  Glass Jar: $(echo $r | jq -r '.data.id // .error.message')"; JAR_ID=$(id "$r")
r=$(post "masterdata/products" '{"code":"CINN-PWD-250","name":"Ceylon Cinnamon Powder 250g Pack","product_type":"FINISHED","category_id":'"$FG_CAT"',"base_uom_id":'"$PKT_ID"',"purchase_uom_id":'"$PKT_ID"',"sales_uom_id":'"$PKT_ID"',"cost_price":280,"standard_price":450,"is_purchased":false,"is_sold":true,"is_manufactured":true}')
echo "  Cinnamon Powder Pack: $(echo $r | jq -r '.data.id // .error.message')"; CINN_PKT=$(id "$r")

# Re-fetch product IDs if creation was skipped (duplicate)
PROD_LIST=$(get "masterdata/products")
CINN_STCK=${CINN_STCK:-$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CINN-STCK") | .id')}
JAR_ID=${JAR_ID:-$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="JAR-250") | .id')}
CINN_PKT=${CINN_PKT:-$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CINN-PWD-250") | .id')}
echo "  CINN_STCK=$CINN_STCK JAR=$JAR_ID CINN_PKT=$CINN_PKT"

echo ""
echo "════════════════════════════════════════"
echo " 13. BUSINESS PARTIES (CONTACTS)"
echo "════════════════════════════════════════"
r=$(post "masterdata/contacts/parties" '{"code":"CUST002","name":"German Spice House GmbH","legal_name":"German Spice House GmbH","party_type":"CUSTOMER","type":"COMPANY","currency_id":'"$EUR_ID"',"payment_term_id":'"$NET30_ID"',"credit_limit":50000}')
echo "  DE Customer: $(echo $r | jq -r '.data.id // .error.message')"; DE_CUST=$(id "$r")
r=$(post "masterdata/contacts/parties" '{"code":"CUST003","name":"Singapore Flavours Pte Ltd","legal_name":"Singapore Flavours Pte Ltd","party_type":"CUSTOMER","type":"COMPANY","currency_id":'"$SGD_ID"',"payment_term_id":'"$IMM_PT"',"credit_limit":30000}')
echo "  SG Customer: $(echo $r | jq -r '.data.id // .error.message')"; SG_CUST=$(id "$r")
r=$(post "masterdata/contacts/parties" '{"code":"CUST004","name":"Dubai Spice Trading LLC","legal_name":"Dubai Spice Trading LLC","party_type":"CUSTOMER","type":"COMPANY","currency_id":'"$AED_ID"',"payment_term_id":'"$NET15_PT"',"credit_limit":25000}')
echo "  AE Customer: $(echo $r | jq -r '.data.id // .error.message')"; AE_CUST=$(id "$r")
r=$(post "masterdata/contacts/parties" '{"code":"SUPP002","name":"Matara Spice Growers Coop","legal_name":"Matara Spice Growers Cooperative Society","party_type":"SUPPLIER","type":"COMPANY","currency_id":'"$LKR_ID"',"payment_term_id":'"$NET30_ID"',"credit_limit":0}')
echo "  LK Supplier 2: $(echo $r | jq -r '.data.id // .error.message')"; LK_SUPP2=$(id "$r")
r=$(post "masterdata/contacts/parties" '{"code":"SUPP003","name":"Lanka Packaging Solutions","legal_name":"Lanka Packaging Solutions (Pvt) Ltd","party_type":"SUPPLIER","type":"COMPANY","currency_id":'"$LKR_ID"',"payment_term_id":'"$NET15_PT"',"credit_limit":0}')
echo "  Packaging Supplier: $(echo $r | jq -r '.data.id // .error.message')"; PKG_SUPP=$(id "$r")
r=$(post "masterdata/contacts/parties" '{"code":"SUPP004","name":"Indian Spice Exporters","legal_name":"Indian Spice Exporters Pvt Ltd","party_type":"SUPPLIER","type":"COMPANY","currency_id":'"$INR_ID"',"payment_term_id":'"$ADV50_PT"',"credit_limit":0}')
echo "  IN Supplier: $(echo $r | jq -r '.data.id // .error.message')"; IN_SUPP=$(id "$r")

PARTY_LIST=$(get "masterdata/contacts/parties")
SUPP1_ID=$(echo "$PARTY_LIST" | jq -r '.data[] | select(.code=="SUPP001") | .id')
DE_CUST=${DE_CUST:-$(echo "$PARTY_LIST" | jq -r '.data[] | select(.code=="CUST002") | .id')}
SG_CUST=${SG_CUST:-$(echo "$PARTY_LIST" | jq -r '.data[] | select(.code=="CUST003") | .id')}
LK_SUPP2=${LK_SUPP2:-$(echo "$PARTY_LIST" | jq -r '.data[] | select(.code=="SUPP002") | .id')}
PKG_SUPP=${PKG_SUPP:-$(echo "$PARTY_LIST" | jq -r '.data[] | select(.code=="SUPP003") | .id')}
echo "  SUPP1=$SUPP1_ID DE_CUST=$DE_CUST SG=$SG_CUST LK_SUPP2=$LK_SUPP2 PKG=$PKG_SUPP"

# Contact persons
if [ -n "$DE_CUST" ]; then
  r=$(post "masterdata/contacts/parties/$DE_CUST/contacts" '{"name":"Hans Mueller","designation":"Purchasing Manager","phone":"+49 30 1234567","mobile":"+49 170 9876543","email":"hans.mueller@germanspice.de","is_primary":true}')
  echo "  DE Contact: $(echo $r | jq -r '.data.id // .error.message')"
fi
if [ -n "$SG_CUST" ]; then
  r=$(post "masterdata/contacts/parties/$SG_CUST/contacts" '{"name":"Lim Wei Jie","designation":"Director of Sourcing","phone":"+65 6234 5678","mobile":"+65 9123 4567","email":"weijie@sgflavours.com","is_primary":true}')
  echo "  SG Contact: $(echo $r | jq -r '.data.id // .error.message')"
fi
if [ -n "$LK_SUPP2" ]; then
  r=$(post "masterdata/contacts/parties/$LK_SUPP2/contacts" '{"name":"Sunil Perera","designation":"Secretary","phone":"+94 41 2234567","mobile":"+94 77 1234567","email":"sunil@mataracoop.lk","is_primary":true}')
  echo "  LK Supp Contact: $(echo $r | jq -r '.data.id // .error.message')"
fi

echo ""
echo "════════════════════════════════════════"
echo " 14. WAREHOUSES & LOCATIONS"
echo "════════════════════════════════════════"
WH1_ID=$(get "masterdata/inventory/warehouses" | jq -r '.data[] | select(.code=="WH001") | .id')
echo "  Existing WH001=$WH1_ID"
r=$(post "masterdata/inventory/warehouses/$WH1_ID/locations" '{"code":"WH001-A1","name":"Rack A1 - Dry Spices","location_type":"STORAGE"}')
echo "  Loc A1: $(echo $r | jq -r '.data.id // .error.message')"; LOC_A1=$(id "$r")
r=$(post "masterdata/inventory/warehouses/$WH1_ID/locations" '{"code":"WH001-A2","name":"Rack A2 - Teas","location_type":"STORAGE"}')
echo "  Loc A2: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/inventory/warehouses/$WH1_ID/locations" '{"code":"WH001-B1","name":"Bin B1 - Oils and Liquids","location_type":"STORAGE"}')
echo "  Loc B1: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/inventory/warehouses/$WH1_ID/locations" '{"code":"WH001-QC","name":"QC Holding Area","location_type":"QUALITY"}')
echo "  Loc QC: $(echo $r | jq -r '.data.id // .error.message')"

r=$(post "masterdata/inventory/warehouses" '{"code":"WH002","name":"Finished Goods Warehouse","address":"Export Zone, Kadahapola"}')
echo "  WH002 FG: $(echo $r | jq -r '.data.id // .error.message')"; WH2_ID=$(id "$r")
WH2_ID=${WH2_ID:-$(get "masterdata/inventory/warehouses" | jq -r '.data[] | select(.code=="WH002") | .id')}
if [ -n "$WH2_ID" ]; then
  r=$(post "masterdata/inventory/warehouses/$WH2_ID/locations" '{"code":"WH002-FG1","name":"FG Rack 1 - Packed Products","location_type":"STORAGE"}')
  echo "  FG Loc 1: $(echo $r | jq -r '.data.id // .error.message')"
  r=$(post "masterdata/inventory/warehouses/$WH2_ID/locations" '{"code":"WH002-DISP","name":"Dispatch Bay","location_type":"SHIPPING"}')
  echo "  Dispatch: $(echo $r | jq -r '.data.id // .error.message')"
fi

echo ""
echo "════════════════════════════════════════"
echo " 15. JOB POSITIONS & EMPLOYEES"
echo "════════════════════════════════════════"
r=$(post "masterdata/hr/job-positions" '{"code":"QAM","name":"QA Manager","department_id":'"$QA_DEPT"'}')
echo "  QA Manager: $(echo $r | jq -r '.data.id // .error.message')"; QAM_POS=$(id "$r")
r=$(post "masterdata/hr/job-positions" '{"code":"PRDSUP","name":"Production Supervisor","department_id":'"$PROD_DEPT"'}')
echo "  Prod Sup: $(echo $r | jq -r '.data.id // .error.message')"; PRDSUP_POS=$(id "$r")
r=$(post "masterdata/hr/job-positions" '{"code":"SALESEXEC","name":"Sales Executive","department_id":'"$SALES_DEPT"'}')
echo "  Sales Exec: $(echo $r | jq -r '.data.id // .error.message')"; SALESEXEC_POS=$(id "$r")
r=$(post "masterdata/hr/job-positions" '{"code":"PROCOFF","name":"Procurement Officer","department_id":'"$PUR_DEPT"'}')
echo "  Procurement: $(echo $r | jq -r '.data.id // .error.message')"; PROCOFF_POS=$(id "$r")
r=$(post "masterdata/hr/job-positions" '{"code":"ACCACC","name":"Accounts Officer","department_id":'"$FIN_DEPT"'}')
echo "  Accounts: $(echo $r | jq -r '.data.id // .error.message')"; ACCOFF_POS=$(id "$r")

JP_LIST=$(get "masterdata/hr/job-positions")
QAM_POS=${QAM_POS:-$(echo "$JP_LIST" | jq -r '.data[] | select(.code=="QAM") | .id')}
PRDSUP_POS=${PRDSUP_POS:-$(echo "$JP_LIST" | jq -r '.data[] | select(.code=="PRDSUP") | .id')}
SALESEXEC_POS=${SALESEXEC_POS:-$(echo "$JP_LIST" | jq -r '.data[] | select(.code=="SALESEXEC") | .id')}
PROCOFF_POS=${PROCOFF_POS:-$(echo "$JP_LIST" | jq -r '.data[] | select(.code=="PROCOFF") | .id')}

r=$(post "masterdata/hr/employees" '{"code":"EMP002","first_name":"Nimal","last_name":"Silva","display_name":"Nimal Silva","gender":"MALE","date_of_birth":"1988-05-15","nic_number":"883363456V","job_position_id":'"$QAM_POS"',"department_id":'"$QA_DEPT"',"employment_type":"PERMANENT","date_joined":"2020-01-01","email":"nimal.silva@kadahapola.lk","phone":"+94 41 2200002","mobile":"+94 77 2000002","bank_id":'"$BOC_ID"',"bank_account_no":"BOC-EMP-002","basic_salary":85000,"currency_id":'"$LKR_ID"'}')
echo "  Nimal Silva: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/hr/employees" '{"code":"EMP003","first_name":"Kamala","last_name":"Jayawardena","display_name":"Kamala Jayawardena","gender":"FEMALE","date_of_birth":"1992-08-22","nic_number":"924563789V","job_position_id":'"$SALESEXEC_POS"',"department_id":'"$SALES_DEPT"',"employment_type":"PERMANENT","date_joined":"2021-03-01","email":"kamala.j@kadahapola.lk","phone":"+94 41 2200003","mobile":"+94 76 3000003","bank_id":'"$HNB_ID"',"bank_account_no":"HNB-EMP-003","basic_salary":65000,"currency_id":'"$LKR_ID"'}')
echo "  Kamala J: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/hr/employees" '{"code":"EMP004","first_name":"Roshan","last_name":"Fernando","display_name":"Roshan Fernando","gender":"MALE","date_of_birth":"1985-11-10","nic_number":"854783210V","job_position_id":'"$PRDSUP_POS"',"department_id":'"$PROD_DEPT"',"employment_type":"PERMANENT","date_joined":"2019-06-01","email":"roshan.f@kadahapola.lk","phone":"+94 41 2200004","mobile":"+94 71 4000004","bank_id":'"$SAMPATH_ID"',"bank_account_no":"SMP-EMP-004","basic_salary":70000,"currency_id":'"$LKR_ID"'}')
echo "  Roshan F: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/hr/employees" '{"code":"EMP005","first_name":"Priya","last_name":"Perera","display_name":"Priya Perera","gender":"FEMALE","date_of_birth":"1990-03-05","nic_number":"900643120V","job_position_id":'"$PROCOFF_POS"',"department_id":'"$PUR_DEPT"',"employment_type":"PERMANENT","date_joined":"2022-07-15","email":"priya.p@kadahapola.lk","phone":"+94 41 2200005","mobile":"+94 77 5000005","bank_id":'"$COM_BANK_ID"',"bank_account_no":"COM-EMP-005","basic_salary":60000,"currency_id":'"$LKR_ID"'}')
echo "  Priya P: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " 16. WORK CENTERS"
echo "════════════════════════════════════════"
WC1_ID=$(get "masterdata/manufacturing/work-centers" | jq -r '.data[0].id // empty')
echo "  Existing WC=$WC1_ID"
r=$(post "masterdata/manufacturing/work-centers" '{"code":"WC-CLEAN","name":"Cleaning and Sorting Station","capacity":500,"cost_per_hour":1200,"currency_id":'"$LKR_ID"',"notes":"Initial cleaning and quality sorting of raw spices"}')
echo "  Cleaning: $(echo $r | jq -r '.data.id // .error.message')"; WC_CLEAN=$(id "$r")
r=$(post "masterdata/manufacturing/work-centers" '{"code":"WC-GRIND","name":"Grinding Room","capacity":300,"cost_per_hour":1800,"currency_id":'"$LKR_ID"',"notes":"Industrial grinding equipment for spice powder production"}')
echo "  Grinding: $(echo $r | jq -r '.data.id // .error.message')"; WC_GRIND=$(id "$r")
r=$(post "masterdata/manufacturing/work-centers" '{"code":"WC-PACK","name":"Packaging Line","capacity":2000,"cost_per_hour":2500,"currency_id":'"$LKR_ID"',"notes":"Automated packaging line for consumer packs"}')
echo "  Packaging: $(echo $r | jq -r '.data.id // .error.message')"; WC_PACK=$(id "$r")
r=$(post "masterdata/manufacturing/work-centers" '{"code":"WC-QA","name":"Quality Lab","capacity":100,"cost_per_hour":800,"currency_id":'"$LKR_ID"',"notes":"Quality assurance testing and certification"}')
echo "  QA Lab: $(echo $r | jq -r '.data.id // .error.message')"; WC_QA=$(id "$r")

WC_LIST=$(get "masterdata/manufacturing/work-centers")
WC_CLEAN=${WC_CLEAN:-$(echo "$WC_LIST" | jq -r '.data[] | select(.code=="WC-CLEAN") | .id')}
WC_GRIND=${WC_GRIND:-$(echo "$WC_LIST" | jq -r '.data[] | select(.code=="WC-GRIND") | .id')}
WC_PACK=${WC_PACK:-$(echo "$WC_LIST" | jq -r '.data[] | select(.code=="WC-PACK") | .id')}
WC_QA=${WC_QA:-$(echo "$WC_LIST" | jq -r '.data[] | select(.code=="WC-QA") | .id')}
echo "  WC_CLEAN=$WC_CLEAN WC_GRIND=$WC_GRIND WC_PACK=$WC_PACK WC_QA=$WC_QA"

echo ""
echo "════════════════════════════════════════"
echo " 17. BOMs & ROUTINGS"
echo "════════════════════════════════════════"
if [ -n "$CINN_PKT" ]; then
  r=$(post "masterdata/manufacturing/boms" '{"code":"BOM-CINN-250","product_id":'"$CINN_PKT"',"quantity":1,"uom_id":'"$PKT_ID"',"bom_type":"MANUFACTURE","notes":"250g cinnamon powder pack BOM"}')
  echo "  BOM Cinn Pkt: $(echo $r | jq -r '.data.id // .error.message')"; BOM1=$(id "$r")
  BOM1=${BOM1:-$(get "masterdata/manufacturing/boms" | jq -r '.data[] | select(.code=="BOM-CINN-250") | .id')}
  if [ -n "$BOM1" ] && [ -n "$CINN_STCK" ]; then
    r=$(post "masterdata/manufacturing/boms/$BOM1/lines" '{"component_id":'"$CINN_STCK"',"quantity":0.28,"uom_id":'"$KG_ID"',"scrap_percent":2,"sequence":1,"notes":"Raw cinnamon sticks"}')
    echo "  BOM Line1: $(echo $r | jq -r '.data.id // .error.message')"
    r=$(post "masterdata/manufacturing/boms/$BOM1/lines" '{"component_id":'"$JAR_ID"',"quantity":1,"uom_id":'"$PCS_ID"',"scrap_percent":0,"sequence":2,"notes":"Glass jar"}')
    echo "  BOM Line2: $(echo $r | jq -r '.data.id // .error.message')"
  fi
  r=$(post "masterdata/manufacturing/routings" '{"code":"RT-CINN-250","product_id":'"$CINN_PKT"',"notes":"Cinnamon powder pack production routing"}')
  echo "  Routing: $(echo $r | jq -r '.data.id // .error.message')"; RT1=$(id "$r")
  RT1=${RT1:-$(get "masterdata/manufacturing/routings" | jq -r '.data[] | select(.code=="RT-CINN-250") | .id')}
  if [ -n "$RT1" ] && [ -n "$WC_CLEAN" ]; then
    r=$(post "masterdata/manufacturing/routings/$RT1/operations" '{"sequence":10,"name":"Cleaning and Sorting","work_center_id":'"$WC_CLEAN"',"setup_time":15,"cycle_time":60,"notes":"Sort and clean cinnamon sticks"}')
    echo "  Op1: $(echo $r | jq -r '.data.id // .error.message')"
    r=$(post "masterdata/manufacturing/routings/$RT1/operations" '{"sequence":20,"name":"Grinding","work_center_id":'"$WC_GRIND"',"setup_time":20,"cycle_time":45,"notes":"Grind to fine powder"}')
    echo "  Op2: $(echo $r | jq -r '.data.id // .error.message')"
    r=$(post "masterdata/manufacturing/routings/$RT1/operations" '{"sequence":30,"name":"Packaging","work_center_id":'"$WC_PACK"',"setup_time":10,"cycle_time":30,"notes":"Fill and seal jars"}')
    echo "  Op3: $(echo $r | jq -r '.data.id // .error.message')"
    r=$(post "masterdata/manufacturing/routings/$RT1/operations" '{"sequence":40,"name":"QA Inspection","work_center_id":'"$WC_QA"',"setup_time":5,"cycle_time":15,"notes":"Quality check before dispatch"}')
    echo "  Op4: $(echo $r | jq -r '.data.id // .error.message')"
  fi
fi

echo ""
echo "════════════════════════════════════════"
echo " 18. MATERIAL CATEGORIES"
echo "════════════════════════════════════════"
RAWMAT_ID=$(get "masterdata/material-categories" | jq -r '.data[] | select(.code=="RAW-MAT") | .id | select(. != null)' | head -1)
echo "  Existing RAW-MAT=$RAWMAT_ID"
r=$(post "masterdata/material-categories" '{"code":"PKG-MAT","name":"Packaging Materials"}')
echo "  Packaging: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/material-categories" '{"code":"SEMI","name":"Semi-Processed"}')
echo "  Semi: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/material-categories" '{"code":"CHEM","name":"Chemicals and Additives"}')
echo "  Chemicals: $(echo $r | jq -r '.data.id // .error.message')"

MC_LIST=$(get "masterdata/material-categories")
RAWMAT_ID=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="RAW-MAT") | .id' | head -1)
PKGMAT_CAT=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="PKG-MAT") | .id')
SEMI_CAT=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="SEMI") | .id')

if [ -n "$RAWMAT_ID" ]; then
  r=$(post "masterdata/material-categories" '{"code":"RAW-SPICE","name":"Raw Spices","parent_id":'"$RAWMAT_ID"'}')
  echo "  Raw Spices (sub): $(echo $r | jq -r '.data.id // .error.message')"
  r=$(post "masterdata/material-categories" '{"code":"RAW-TEA","name":"Raw Tea Leaves","parent_id":'"$RAWMAT_ID"'}')
  echo "  Raw Tea (sub): $(echo $r | jq -r '.data.id // .error.message')"
fi
if [ -n "$PKGMAT_CAT" ]; then
  r=$(post "masterdata/material-categories" '{"code":"PKG-GLASS","name":"Glass Containers","parent_id":'"$PKGMAT_CAT"'}')
  echo "  Glass (sub): $(echo $r | jq -r '.data.id // .error.message')"
  r=$(post "masterdata/material-categories" '{"code":"PKG-PAPER","name":"Paper and Labels","parent_id":'"$PKGMAT_CAT"'}')
  echo "  Paper (sub): $(echo $r | jq -r '.data.id // .error.message')"
fi

MC_LIST=$(get "masterdata/material-categories")
RAWSPICE_CAT=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="RAW-SPICE") | .id')
RAWTEA_CAT=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="RAW-TEA") | .id')
GLASS_CAT=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="PKG-GLASS") | .id')
PAPER_CAT=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="PKG-PAPER") | .id')
echo "  RAW-SPICE=$RAWSPICE_CAT RAW-TEA=$RAWTEA_CAT GLASS=$GLASS_CAT PAPER=$PAPER_CAT"

echo ""
echo "════════════════════════════════════════"
echo " 19. MATERIALS"
echo "════════════════════════════════════════"
CAT_ID=${RAWSPICE_CAT:-$RAWMAT_ID}

r=$(post "masterdata/materials" '{"code":"RM-CINN-001","name":"Cinnamon Sticks (Grade A)","material_type":"RAW_MATERIAL","category_id":'"$CAT_ID"',"color":"Brown","article_code":"CINN-GRADE-A","description":"High grade Ceylon cinnamon sticks for export","notes":"Source: Matara and Galle districts"}')
echo "  Cinn Sticks RM: $(echo $r | jq -r '.data.id // .error.message')"; MAT_CINN=$(id "$r")
r=$(post "masterdata/materials" '{"code":"RM-BLKPEP-001","name":"Black Pepper Berries","material_type":"RAW_MATERIAL","category_id":'"$CAT_ID"',"color":"Green/Black","description":"Fresh black pepper berries pre-harvest","notes":"Sourced from hill country suppliers"}')
echo "  Black Pepper RM: $(echo $r | jq -r '.data.id // .error.message')"; MAT_PEP=$(id "$r")
r=$(post "masterdata/materials" '{"code":"RM-CLOVE-001","name":"Cloves (Whole)","material_type":"RAW_MATERIAL","category_id":'"$CAT_ID"',"color":"Dark Brown","description":"Premium whole cloves from local farms"}')
echo "  Cloves RM: $(echo $r | jq -r '.data.id // .error.message')"; MAT_CLOV=$(id "$r")

RAWTEA_ID=${RAWTEA_CAT:-$RAWMAT_ID}
r=$(post "masterdata/materials" '{"code":"RM-TEA-001","name":"Ceylon Tea Leaves BOP","material_type":"RAW_MATERIAL","category_id":'"$RAWTEA_ID"',"color":"Dark Green","description":"Broken Orange Pekoe tea leaves"}')
echo "  Tea Leaves RM: $(echo $r | jq -r '.data.id // .error.message')"; MAT_TEA=$(id "$r")

PKGCAT_ID=${GLASS_CAT:-$PKGMAT_CAT}
r=$(post "masterdata/materials" '{"code":"PKG-JAR-250","name":"Glass Jar 250g Clear","material_type":"RAW_MATERIAL","category_id":'"$PKGCAT_ID"',"color":"Clear","article_code":"GJ-250","barcode":"4012345678901","description":"250g clear glass jar with metal lid"}')
echo "  Glass Jar PKG: $(echo $r | jq -r '.data.id // .error.message')"; MAT_JAR=$(id "$r")

PAPER_ID=${PAPER_CAT:-$PKGMAT_CAT}
r=$(post "masterdata/materials" '{"code":"PKG-LBL-001","name":"Product Label - Cinnamon","material_type":"RAW_MATERIAL","category_id":'"$PAPER_ID"',"color":"White/Gold","description":"Premium product labels for 250g cinnamon packs"}')
echo "  Label PKG: $(echo $r | jq -r '.data.id // .error.message')"; MAT_LBL=$(id "$r")

SEMI_CAT_ID=${SEMI_CAT:-$RAWMAT_ID}
r=$(post "masterdata/materials" '{"code":"SF-CINN-PWD","name":"Ground Cinnamon Powder","material_type":"SEMI_FINISHED","category_id":'"$SEMI_CAT_ID"',"color":"Beige","description":"Ground cinnamon powder - bulk before packaging"}')
echo "  Ground Cinnamon SF: $(echo $r | jq -r '.data.id // .error.message')"; MAT_CINNPWD=$(id "$r")
r=$(post "masterdata/materials" '{"code":"SRV-QTEST","name":"Microbiology Quality Test","material_type":"SERVICE","description":"Third-party lab microbiology testing service per batch"}')
echo "  QA Service: $(echo $r | jq -r '.data.id // .error.message')"; MAT_SRV=$(id "$r")

# Re-fetch material IDs if created in previous run
MAT_LIST=$(get "masterdata/materials")
MAT_CINN=${MAT_CINN:-$(echo "$MAT_LIST" | jq -r '.data[] | select(.code=="RM-CINN-001") | .id')}
MAT_JAR=${MAT_JAR:-$(echo "$MAT_LIST" | jq -r '.data[] | select(.code=="PKG-JAR-250") | .id')}
MAT_CINNPWD=${MAT_CINNPWD:-$(echo "$MAT_LIST" | jq -r '.data[] | select(.code=="SF-CINN-PWD") | .id')}
echo "  MAT_CINN=$MAT_CINN MAT_JAR=$MAT_JAR MAT_CINNPWD=$MAT_CINNPWD"

# Sub-tables for cinnamon sticks
if [ -n "$MAT_CINN" ]; then
  r=$(put "masterdata/materials/$MAT_CINN/purchasing" '{"purchasing_uom_id":'"$KG_ID"',"under_delivery_pct":5,"over_delivery_pct":10}')
  echo "  Cinn Purchasing: $(echo $r | jq -r '.message // .error.message')"
  r=$(put "masterdata/materials/$MAT_CINN/manufacturing" '{"production_uom_id":'"$KG_ID"',"reorder_qty_level":200,"safety_level":50,"production_days":0,"delivery_days":5,"grn_days":2,"procurement_repeat_days":30}')
  echo "  Cinn Manufacturing: $(echo $r | jq -r '.message // .error.message')"
  r=$(put "masterdata/materials/$MAT_CINN/warehouse" '{"stocking_uom_id":'"$KG_ID"',"stock_removal":"FIFO","storage_main":true,"storage_damaged":true,"storage_hold":true,"batch_process":true,"production_date_check":false,"expiry_date_check":true,"qc_check":true,"grn_with_po_uom_image":false}')
  echo "  Cinn Warehouse: $(echo $r | jq -r '.message // .error.message')"
  if [ -n "$SUPP1_ID" ]; then
    post "masterdata/materials/$MAT_CINN/vendors/replace" '[{"vendor_id":'"$SUPP1_ID"',"article_no":"CINN-A-CEYLON","delivery_days":5,"cost":850,"currency_id":'"$LKR_ID"',"projected_price":820,"moq":50,"is_default":true},{"vendor_id":'"${LK_SUPP2:-$SUPP1_ID}"',"article_no":"CINN-MAT","delivery_days":7,"cost":820,"currency_id":'"$LKR_ID"',"projected_price":800,"moq":100,"is_default":false}]' > /dev/null
    echo "  Cinn Vendors: saved"
  fi
  post "masterdata/materials/$MAT_CINN/measurements/replace" '[{"base_uom_id":'"$KG_ID"',"target_uom_id":'"$G_ID"',"conversion_ratio":1000}]' > /dev/null
  echo "  Cinn Measurements: saved"
fi

# Sub-tables for glass jar
if [ -n "$MAT_JAR" ] && [ -n "$BOX_ID" ]; then
  r=$(put "masterdata/materials/$MAT_JAR/purchasing" '{"purchasing_uom_id":'"$BOX_ID"',"under_delivery_pct":2,"over_delivery_pct":5}')
  echo "  Jar Purchasing: $(echo $r | jq -r '.message // .error.message')"
  r=$(put "masterdata/materials/$MAT_JAR/warehouse" '{"stocking_uom_id":'"$PCS_ID"',"stock_removal":"FIFO","storage_main":true,"storage_damaged":false,"storage_hold":false,"batch_process":false,"production_date_check":false,"expiry_date_check":false,"qc_check":false,"grn_with_po_uom_image":true}')
  echo "  Jar Warehouse: $(echo $r | jq -r '.message // .error.message')"
  if [ -n "$PKG_SUPP" ]; then
    post "masterdata/materials/$MAT_JAR/vendors/replace" '[{"vendor_id":'"$PKG_SUPP"',"article_no":"GJ-250-CLEAR","delivery_days":14,"cost":45,"currency_id":'"$LKR_ID"',"projected_price":42,"moq":1000,"is_default":true}]' > /dev/null
    echo "  Jar Vendors: saved"
  fi
fi

# Sub-tables for semi-finished cinnamon powder
if [ -n "$MAT_CINNPWD" ]; then
  r=$(put "masterdata/materials/$MAT_CINNPWD/manufacturing" '{"production_uom_id":'"$KG_ID"',"reorder_qty_level":100,"safety_level":20,"production_days":1,"delivery_days":0,"grn_days":0,"procurement_repeat_days":0}')
  echo "  SF Cinn Mfg: $(echo $r | jq -r '.message // .error.message')"
  r=$(put "masterdata/materials/$MAT_CINNPWD/warehouse" '{"stocking_uom_id":'"$KG_ID"',"stock_removal":"FIFO","storage_main":true,"storage_damaged":false,"storage_hold":true,"batch_process":true,"production_date_check":true,"expiry_date_check":true,"qc_check":true,"grn_with_po_uom_image":false}')
  echo "  SF Cinn Warehouse: $(echo $r | jq -r '.message // .error.message')"
fi

echo ""
echo "════════════════════════════════════════"
echo " 20. COMPANY SETTINGS"
echo "════════════════════════════════════════"
r=$(put "masterdata/organization/company" '{"name":"Kadahapola Exports","legal_name":"Kadahapola Exports (Pvt) Ltd","tax_reg_number":"VAT-115234567-7000","email":"info@kadahapola.lk","phone":"+94 41 2234500","website":"www.kadahapola.lk","address_line1":"No. 12, Export Zone Road","address_line2":"Kadahapola Industrial Estate","city":"Matara","postal_code":"81000","base_currency_id":'"$LKR_ID"',"fiscal_year_start":4}')
echo "  Company: $(echo $r | jq -r '.message // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " ALL DONE"
echo "════════════════════════════════════════"
