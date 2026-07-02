#!/bin/sh
# Extended Master Data Seed — Kadahapola Exports Ltd
# Covers: MRP Material Master, Opening Stock Entries, Additional Contacts & HR
# Run inside the erp-api-dev container: sh /app/scripts/seed_extended.sh

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

# ── Re-fetch IDs we depend on ──────────────────────────────────────────────────
PROD_LIST=$(get "masterdata/products?per_page=100")
CINN_STCK=$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CINN-STCK") | .id')
BLKPEP=$(echo "$PROD_LIST"   | jq -r '.data[] | select(.code=="BLKPEP") | .id')
CLOVES=$(echo "$PROD_LIST"   | jq -r '.data[] | select(.code=="CLOVES") | .id')
CEYTEA=$(echo "$PROD_LIST"   | jq -r '.data[] | select(.code=="CEYTEA") | .id')
COCOIL=$(echo "$PROD_LIST"   | jq -r '.data[] | select(.code=="COCO-OIL") | .id')
JAR=$(echo "$PROD_LIST"      | jq -r '.data[] | select(.code=="JAR-250") | .id')
CINN_PKT=$(echo "$PROD_LIST" | jq -r '.data[] | select(.code=="CINN-PWD-250") | .id')
echo "  Products — CINN_STCK=$CINN_STCK BLKPEP=$BLKPEP CLOVES=$CLOVES CEYTEA=$CEYTEA COCOIL=$COCOIL JAR=$JAR CINN_PKT=$CINN_PKT"

UOM_LIST=$(get "masterdata/products/uoms?per_page=50")
KG_ID=$(echo "$UOM_LIST"  | jq -r '.data[] | select(.code=="KG") | .id')
L_ID=$(echo "$UOM_LIST"   | jq -r '.data[] | select(.code=="L") | .id')
PCS_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="PCS") | .id')
PKT_ID=$(echo "$UOM_LIST" | jq -r '.data[] | select(.code=="PKT") | .id')
echo "  UOMs — KG=$KG_ID L=$L_ID PCS=$PCS_ID PKT=$PKT_ID"

WH_LIST=$(get "masterdata/inventory/warehouses?per_page=20")
WH1_ID=$(echo "$WH_LIST" | jq -r '.data[] | select(.code=="WH001") | .id')
WH2_ID=$(echo "$WH_LIST" | jq -r '.data[] | select(.code=="WH002") | .id')
echo "  Warehouses — WH1=$WH1_ID WH2=$WH2_ID"

LOC_LIST=$(get "masterdata/inventory/warehouses/$WH1_ID/locations?per_page=20")
LOC_A1=$(echo "$LOC_LIST" | jq -r '.data[] | select(.code=="WH001-A1") | .id')
LOC_A2=$(echo "$LOC_LIST" | jq -r '.data[] | select(.code=="WH001-A2") | .id')
LOC_B1=$(echo "$LOC_LIST" | jq -r '.data[] | select(.code=="WH001-B1") | .id')
echo "  Locations — A1=$LOC_A1 A2=$LOC_A2 B1=$LOC_B1"

LOC2_LIST=$(get "masterdata/inventory/warehouses/$WH2_ID/locations?per_page=20")
LOC_FG1=$(echo "$LOC2_LIST" | jq -r '.data[] | select(.code=="WH002-FG1") | .id')
echo "  WH2 Loc — FG1=$LOC_FG1"

PARTY_LIST=$(get "masterdata/contacts/parties?per_page=50")
SUPP1=$(echo "$PARTY_LIST" | jq -r '.data[] | select(.code=="SUPP001") | .id')
DEPT_LIST=$(get "masterdata/organization/departments?per_page=20")
PROD_DEPT=$(echo "$DEPT_LIST" | jq -r '.data[] | select(.code=="PROD") | .id')
FIN_DEPT=$(echo "$DEPT_LIST"  | jq -r '.data[] | select(.code=="FIN") | .id')
SALES_DEPT=$(echo "$DEPT_LIST"| jq -r '.data[] | select(.code=="SALES") | .id')
PUR_DEPT=$(echo "$DEPT_LIST"  | jq -r '.data[] | select(.code=="PUR") | .id')

CURR_LIST=$(get "masterdata/financial/currencies?per_page=20")
LKR_ID=$(echo "$CURR_LIST" | jq -r '.data[] | select(.code=="LKR") | .id')

echo ""
echo "════════════════════════════════════════"
echo " 1. MRP MATERIAL MASTER"
echo "════════════════════════════════════════"

# Cinnamon Sticks — Class A raw material, buy externally
if [ -n "$CINN_STCK" ]; then
  r=$(post "masterdata/mrp" '{
    "product_id":'"$CINN_STCK"',
    "material_type":"RAW_MATERIAL",
    "mrp_type":"REORDER_POINT",
    "procurement_type":"EXTERNAL",
    "purchase_lead_time":5,
    "production_lead_time":0,
    "safety_stock":100,
    "reorder_point":200,
    "min_order_qty":50,
    "lot_size_type":"LOT_FOR_LOT",
    "abc_class":"A",
    "shelf_life_days":730,
    "notes":"Primary export spice — maintain high stock level"
  }')
  echo "  Cinnamon Sticks MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Black Pepper — Class A raw material
if [ -n "$BLKPEP" ]; then
  r=$(post "masterdata/mrp" '{
    "product_id":'"$BLKPEP"',
    "material_type":"RAW_MATERIAL",
    "mrp_type":"REORDER_POINT",
    "procurement_type":"EXTERNAL",
    "purchase_lead_time":7,
    "production_lead_time":0,
    "safety_stock":75,
    "reorder_point":150,
    "min_order_qty":25,
    "lot_size_type":"LOT_FOR_LOT",
    "abc_class":"A",
    "shelf_life_days":540,
    "notes":"High demand export item"
  }')
  echo "  Black Pepper MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Cloves — Class B raw material (expensive, buy less frequently)
if [ -n "$CLOVES" ]; then
  r=$(post "masterdata/mrp" '{
    "product_id":'"$CLOVES"',
    "material_type":"RAW_MATERIAL",
    "mrp_type":"REORDER_POINT",
    "procurement_type":"EXTERNAL",
    "purchase_lead_time":10,
    "production_lead_time":0,
    "safety_stock":20,
    "reorder_point":50,
    "min_order_qty":10,
    "lot_size_type":"LOT_FOR_LOT",
    "abc_class":"B",
    "shelf_life_days":365,
    "notes":"High value item — order carefully to avoid overstock"
  }')
  echo "  Cloves MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Ceylon Tea — Class A raw material
if [ -n "$CEYTEA" ]; then
  r=$(post "masterdata/mrp" '{
    "product_id":'"$CEYTEA"',
    "material_type":"RAW_MATERIAL",
    "mrp_type":"REORDER_POINT",
    "procurement_type":"EXTERNAL",
    "purchase_lead_time":3,
    "production_lead_time":0,
    "safety_stock":100,
    "reorder_point":250,
    "min_order_qty":50,
    "lot_size_type":"LOT_FOR_LOT",
    "abc_class":"A",
    "shelf_life_days":365,
    "notes":"Seasonal availability — build stock during harvest"
  }')
  echo "  Ceylon Tea MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Coconut Oil — Class B
if [ -n "$COCOIL" ]; then
  r=$(post "masterdata/mrp" '{
    "product_id":'"$COCOIL"',
    "material_type":"RAW_MATERIAL",
    "mrp_type":"REORDER_POINT",
    "procurement_type":"EXTERNAL",
    "purchase_lead_time":7,
    "production_lead_time":0,
    "safety_stock":50,
    "reorder_point":100,
    "min_order_qty":20,
    "lot_size_type":"LOT_FOR_LOT",
    "abc_class":"B",
    "shelf_life_days":720,
    "notes":"Store in cool dry conditions away from direct sunlight"
  }')
  echo "  Coconut Oil MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Glass Jar 250g — Class C packaging (buy in fixed lots of 500)
if [ -n "$JAR" ]; then
  FIXED_500=500
  r=$(post "masterdata/mrp" '{
    "product_id":'"$JAR"',
    "material_type":"RAW_MATERIAL",
    "mrp_type":"REORDER_POINT",
    "procurement_type":"EXTERNAL",
    "purchase_lead_time":14,
    "production_lead_time":0,
    "safety_stock":500,
    "reorder_point":1000,
    "min_order_qty":500,
    "max_order_qty":5000,
    "lot_size_type":"FIXED_LOT",
    "fixed_lot_size":500,
    "abc_class":"C",
    "notes":"Order in boxes of 500 — 2-week lead time from supplier"
  }')
  echo "  Glass Jar MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Cinnamon Powder Pack 250g — SEMI_FINISHED, IN_HOUSE production
if [ -n "$CINN_PKT" ]; then
  r=$(post "masterdata/mrp" '{
    "product_id":'"$CINN_PKT"',
    "material_type":"SEMI_FINISHED",
    "mrp_type":"MRP",
    "procurement_type":"IN_HOUSE",
    "purchase_lead_time":0,
    "production_lead_time":2,
    "safety_stock":200,
    "reorder_point":500,
    "min_order_qty":100,
    "lot_size_type":"LOT_FOR_LOT",
    "abc_class":"A",
    "shelf_life_days":540,
    "notes":"MRP-driven production — uses BOM-CINN-250"
  }')
  echo "  Cinnamon Powder Pack MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi

echo ""
echo "════════════════════════════════════════"
echo " 2. OPENING STOCK ENTRIES"
echo "════════════════════════════════════════"

TODAY="2026-01-01T00:00:00Z"

# Cinnamon Sticks — 500 kg in WH001 Rack A1
if [ -n "$CINN_STCK" ] && [ -n "$WH1_ID" ]; then
  r=$(post "masterdata/inventory/stock/entries" '{
    "product_id":'"$CINN_STCK"',
    "warehouse_id":'"$WH1_ID"',
    "location_id":'"$LOC_A1"',
    "transaction_type":"OPENING",
    "quantity":500,
    "unit_cost":850,
    "transaction_date":"'"$TODAY"'",
    "notes":"Opening stock — cinnamon sticks grade A"
  }')
  echo "  Cinnamon Sticks Stock: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Black Pepper — 200 kg in WH001 Rack A1
if [ -n "$BLKPEP" ] && [ -n "$WH1_ID" ]; then
  r=$(post "masterdata/inventory/stock/entries" '{
    "product_id":'"$BLKPEP"',
    "warehouse_id":'"$WH1_ID"',
    "location_id":'"$LOC_A1"',
    "transaction_type":"OPENING",
    "quantity":200,
    "unit_cost":920,
    "transaction_date":"'"$TODAY"'",
    "notes":"Opening stock — black pepper whole"
  }')
  echo "  Black Pepper Stock: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Cloves — 80 kg in WH001 Rack A1
if [ -n "$CLOVES" ] && [ -n "$WH1_ID" ]; then
  r=$(post "masterdata/inventory/stock/entries" '{
    "product_id":'"$CLOVES"',
    "warehouse_id":'"$WH1_ID"',
    "location_id":'"$LOC_A1"',
    "transaction_type":"OPENING",
    "quantity":80,
    "unit_cost":2800,
    "transaction_date":"'"$TODAY"'",
    "notes":"Opening stock — cloves whole"
  }')
  echo "  Cloves Stock: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Ceylon Tea — 300 kg in WH001 Rack A2
if [ -n "$CEYTEA" ] && [ -n "$WH1_ID" ]; then
  r=$(post "masterdata/inventory/stock/entries" '{
    "product_id":'"$CEYTEA"',
    "warehouse_id":'"$WH1_ID"',
    "location_id":'"$LOC_A2"',
    "transaction_type":"OPENING",
    "quantity":300,
    "unit_cost":650,
    "transaction_date":"'"$TODAY"'",
    "notes":"Opening stock — Ceylon black tea BOP"
  }')
  echo "  Ceylon Tea Stock: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Coconut Oil — 120 L in WH001 Rack B1
if [ -n "$COCOIL" ] && [ -n "$WH1_ID" ]; then
  r=$(post "masterdata/inventory/stock/entries" '{
    "product_id":'"$COCOIL"',
    "warehouse_id":'"$WH1_ID"',
    "location_id":'"$LOC_B1"',
    "transaction_type":"OPENING",
    "quantity":120,
    "unit_cost":480,
    "transaction_date":"'"$TODAY"'",
    "notes":"Opening stock — virgin coconut oil"
  }')
  echo "  Coconut Oil Stock: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Glass Jar 250g — 3000 pcs in WH002 FG Rack 1
if [ -n "$JAR" ] && [ -n "$WH2_ID" ]; then
  r=$(post "masterdata/inventory/stock/entries" '{
    "product_id":'"$JAR"',
    "warehouse_id":'"$WH2_ID"',
    "location_id":'"$LOC_FG1"',
    "transaction_type":"OPENING",
    "quantity":3000,
    "unit_cost":45,
    "transaction_date":"'"$TODAY"'",
    "notes":"Opening stock — 250g glass jars"
  }')
  echo "  Glass Jar Stock: $(echo $r | jq -r '.data.id // .error.message')"
fi

# Cinnamon Powder Packs — 800 pkt in WH002 FG Rack 1
if [ -n "$CINN_PKT" ] && [ -n "$WH2_ID" ]; then
  r=$(post "masterdata/inventory/stock/entries" '{
    "product_id":'"$CINN_PKT"',
    "warehouse_id":'"$WH2_ID"',
    "location_id":'"$LOC_FG1"',
    "transaction_type":"OPENING",
    "quantity":800,
    "unit_cost":280,
    "transaction_date":"'"$TODAY"'",
    "notes":"Opening stock — cinnamon powder 250g packs"
  }')
  echo "  Cinnamon Powder Packs Stock: $(echo $r | jq -r '.data.id // .error.message')"
fi

echo ""
echo "════════════════════════════════════════"
echo " 3. ADDITIONAL JOB POSITIONS"
echo "════════════════════════════════════════"
r=$(post "masterdata/hr/job-positions" '{"code":"LOGOFF","name":"Logistics Officer","department_id":'"$PROD_DEPT"'}')
echo "  Logistics: $(echo $r | jq -r '.data.id // .error.message')"; LOGOFF=$(id "$r")
r=$(post "masterdata/hr/job-positions" '{"code":"FINMGR","name":"Finance Manager","department_id":'"$FIN_DEPT"'}')
echo "  Fin Manager: $(echo $r | jq -r '.data.id // .error.message')"; FINMGR=$(id "$r")
r=$(post "masterdata/hr/job-positions" '{"code":"ITTECH","name":"IT Technician","department_id":'"$PROD_DEPT"'}')
echo "  IT Tech: $(echo $r | jq -r '.data.id // .error.message')"; ITTECH=$(id "$r")

JP_LIST=$(get "masterdata/hr/job-positions?per_page=50")
LOGOFF=${LOGOFF:-$(echo "$JP_LIST" | jq -r '.data[] | select(.code=="LOGOFF") | .id')}
FINMGR=${FINMGR:-$(echo "$JP_LIST" | jq -r '.data[] | select(.code=="FINMGR") | .id')}

echo ""
echo "════════════════════════════════════════"
echo " 4. ADDITIONAL EMPLOYEES"
echo "════════════════════════════════════════"
BOC_ID=$(get "masterdata/financial/banks?per_page=20" | jq -r '.data[] | select(.name | startswith("Bank of Ceylon")) | .id' | head -1)
HNB_ID=$(get "masterdata/financial/banks?per_page=20" | jq -r '.data[] | select(.name == "Hatton National Bank") | .id' | head -1)

r=$(post "masterdata/hr/employees" '{
  "code":"EMP006",
  "first_name":"Chamara",
  "last_name":"Wickramasinghe",
  "display_name":"Chamara Wickramasinghe",
  "gender":"MALE",
  "date_of_birth":"1983-07-18",
  "nic_number":"831995420V",
  "job_position_id":'"$LOGOFF"',
  "department_id":'"$PROD_DEPT"',
  "employment_type":"PERMANENT",
  "date_joined":"2018-09-01",
  "email":"chamara.w@kadahapola.lk",
  "phone":"+94 41 2200006",
  "mobile":"+94 77 6000006",
  "bank_id":'"$BOC_ID"',
  "bank_account_no":"BOC-EMP-006",
  "basic_salary":55000,
  "currency_id":'"$LKR_ID"'
}')
echo "  Chamara W: $(echo $r | jq -r '.data.id // .error.message')"

r=$(post "masterdata/hr/employees" '{
  "code":"EMP007",
  "first_name":"Sanduni",
  "last_name":"Rathnayake",
  "display_name":"Sanduni Rathnayake",
  "gender":"FEMALE",
  "date_of_birth":"1995-02-14",
  "nic_number":"953453218V",
  "job_position_id":'"$FINMGR"',
  "department_id":'"$FIN_DEPT"',
  "employment_type":"PERMANENT",
  "date_joined":"2023-01-10",
  "email":"sanduni.r@kadahapola.lk",
  "phone":"+94 41 2200007",
  "mobile":"+94 70 7000007",
  "bank_id":'"$HNB_ID"',
  "bank_account_no":"HNB-EMP-007",
  "basic_salary":95000,
  "currency_id":'"$LKR_ID"'
}')
echo "  Sanduni R: $(echo $r | jq -r '.data.id // .error.message')"

r=$(post "masterdata/hr/employees" '{
  "code":"EMP008",
  "first_name":"Tharaka",
  "last_name":"Bandara",
  "display_name":"Tharaka Bandara",
  "gender":"MALE",
  "date_of_birth":"1998-11-30",
  "nic_number":"983351890V",
  "job_position_id":'"$LOGOFF"',
  "department_id":'"$PROD_DEPT"',
  "employment_type":"CONTRACT",
  "date_joined":"2024-03-01",
  "email":"tharaka.b@kadahapola.lk",
  "phone":"+94 41 2200008",
  "mobile":"+94 76 8000008",
  "bank_id":'"$BOC_ID"',
  "bank_account_no":"BOC-EMP-008",
  "basic_salary":42000,
  "currency_id":'"$LKR_ID"'
}')
echo "  Tharaka B: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " 5. ADDITIONAL BUSINESS PARTIES"
echo "════════════════════════════════════════"
PT_LIST=$(get "masterdata/financial/payment-terms?per_page=20")
NET30_ID=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="NET30") | .id')
NET15_ID=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="NET15") | .id')
NET60_ID=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="NET60") | .id')
ADV50_ID=$(echo "$PT_LIST" | jq -r '.data[] | select(.code=="ADV50") | .id')

CURR_LIST2=$(get "masterdata/financial/currencies?per_page=20")
GBP_ID=$(echo "$CURR_LIST2" | jq -r '.data[] | select(.code=="GBP") | .id')
USD_ID=$(echo "$CURR_LIST2" | jq -r '.data[] | select(.code=="USD") | .id')

r=$(post "masterdata/contacts/parties" '{
  "code":"CUST005",
  "name":"London Herb Co Ltd",
  "legal_name":"London Herb Company Limited",
  "party_type":"CUSTOMER",
  "type":"COMPANY",
  "currency_id":'"$GBP_ID"',
  "payment_term_id":'"$NET30_ID"',
  "credit_limit":40000,
  "notes":"UK distributor for organic spices and teas"
}')
echo "  UK Customer: $(echo $r | jq -r '.data.id // .error.message')"; UK_CUST=$(id "$r")

r=$(post "masterdata/contacts/parties" '{
  "code":"CUST006",
  "name":"Natural Foods Australia",
  "legal_name":"Natural Foods Australia Pty Ltd",
  "party_type":"CUSTOMER",
  "type":"COMPANY",
  "currency_id":'"$USD_ID"',
  "payment_term_id":'"$NET60_ID"',
  "credit_limit":35000,
  "notes":"Australian bulk buyer for health food market"
}')
echo "  AU Customer: $(echo $r | jq -r '.data.id // .error.message')"

r=$(post "masterdata/contacts/parties" '{
  "code":"SUPP005",
  "name":"Galle Cloves Estate",
  "legal_name":"Galle Cloves Estate (Pvt) Ltd",
  "party_type":"SUPPLIER",
  "type":"COMPANY",
  "currency_id":'"$LKR_ID"',
  "payment_term_id":'"$NET30_ID"',
  "credit_limit":0,
  "notes":"Primary clove supplier from Galle district"
}')
echo "  Cloves Supplier: $(echo $r | jq -r '.data.id // .error.message')"; CLOV_SUPP=$(id "$r")

r=$(post "masterdata/contacts/parties" '{
  "code":"SUPP006",
  "name":"Ratnapura Tea Growers",
  "legal_name":"Ratnapura Tea Growers Association",
  "party_type":"SUPPLIER",
  "type":"COMPANY",
  "currency_id":'"$LKR_ID"',
  "payment_term_id":'"$NET15_ID"',
  "credit_limit":0,
  "notes":"Cooperative tea supplier — Ratnapura high-grown region"
}')
echo "  Tea Supplier: $(echo $r | jq -r '.data.id // .error.message')"

# Re-fetch to get IDs if duplicate
PARTY_LIST=$(get "masterdata/contacts/parties?per_page=100")
UK_CUST=${UK_CUST:-$(echo "$PARTY_LIST" | jq -r '.data[] | select(.code=="CUST005") | .id')}
CLOV_SUPP=${CLOV_SUPP:-$(echo "$PARTY_LIST" | jq -r '.data[] | select(.code=="SUPP005") | .id')}

# Contact persons for new parties
if [ -n "$UK_CUST" ]; then
  r=$(post "masterdata/contacts/parties/$UK_CUST/contacts" '{"name":"James Thornton","designation":"Head of Procurement","phone":"+44 20 7123 4567","mobile":"+44 7700 123456","email":"j.thornton@londonherbco.uk","is_primary":true}')
  echo "  UK Contact: $(echo $r | jq -r '.data.id // .error.message')"
fi
if [ -n "$CLOV_SUPP" ]; then
  r=$(post "masterdata/contacts/parties/$CLOV_SUPP/contacts" '{"name":"Aruna Gunawardena","designation":"Estate Manager","phone":"+94 91 2245678","mobile":"+94 77 9000009","email":"aruna.g@gallecloves.lk","is_primary":true}')
  echo "  Cloves Supplier Contact: $(echo $r | jq -r '.data.id // .error.message')"
fi

echo ""
echo "════════════════════════════════════════"
echo " 6. ADDITIONAL PRODUCT CATEGORIES & PRODUCTS"
echo "════════════════════════════════════════"
r=$(post "masterdata/products/categories" '{"code":"HERB","name":"Herbal Products"}')
echo "  Herbs Cat: $(echo $r | jq -r '.data.id // .error.message')"; HERB_CAT=$(id "$r")
r=$(post "masterdata/products/categories" '{"code":"BULK","name":"Bulk Commodities"}')
echo "  Bulk Cat: $(echo $r | jq -r '.data.id // .error.message')"; BULK_CAT=$(id "$r")

PCAT_LIST=$(get "masterdata/products/categories?per_page=20")
HERB_CAT=${HERB_CAT:-$(echo "$PCAT_LIST" | jq -r '.data[] | select(.code=="HERB") | .id')}
BULK_CAT=${BULK_CAT:-$(echo "$PCAT_LIST" | jq -r '.data[] | select(.code=="BULK") | .id')}
SPICE_CAT=$(echo "$PCAT_LIST" | jq -r '.data[] | select(.code=="SPICE") | .id')

r=$(post "masterdata/products" '{
  "code":"CARDM",
  "name":"Cardamom Pods Green",
  "product_type":"FINISHED",
  "category_id":'"$SPICE_CAT"',
  "base_uom_id":'"$KG_ID"',
  "purchase_uom_id":'"$KG_ID"',
  "sales_uom_id":'"$KG_ID"',
  "cost_price":4500,
  "standard_price":6000,
  "is_purchased":true,
  "is_sold":true,
  "is_manufactured":false
}')
echo "  Cardamom: $(echo $r | jq -r '.data.id // .error.message')"; CARDM_ID=$(id "$r")

r=$(post "masterdata/products" '{
  "code":"TURMRIC",
  "name":"Turmeric Root Whole",
  "product_type":"FINISHED",
  "category_id":'"$SPICE_CAT"',
  "base_uom_id":'"$KG_ID"',
  "purchase_uom_id":'"$KG_ID"',
  "sales_uom_id":'"$KG_ID"',
  "cost_price":380,
  "standard_price":580,
  "is_purchased":true,
  "is_sold":true,
  "is_manufactured":false
}')
echo "  Turmeric: $(echo $r | jq -r '.data.id // .error.message')"; TURM_ID=$(id "$r")

r=$(post "masterdata/products" '{
  "code":"LEMON-GRASS",
  "name":"Dried Lemongrass",
  "product_type":"FINISHED",
  "category_id":'"$HERB_CAT"',
  "base_uom_id":'"$KG_ID"',
  "purchase_uom_id":'"$KG_ID"',
  "sales_uom_id":'"$KG_ID"',
  "cost_price":520,
  "standard_price":780,
  "is_purchased":true,
  "is_sold":true,
  "is_manufactured":false
}')
echo "  Lemongrass: $(echo $r | jq -r '.data.id // .error.message')"; LGRASS_ID=$(id "$r")

r=$(post "masterdata/products" '{
  "code":"MORINGA",
  "name":"Moringa Leaf Powder 100g Pack",
  "product_type":"FINISHED",
  "category_id":'"$HERB_CAT"',
  "base_uom_id":'"$PKT_ID"',
  "purchase_uom_id":'"$PKT_ID"',
  "sales_uom_id":'"$PKT_ID"',
  "cost_price":180,
  "standard_price":320,
  "is_purchased":false,
  "is_sold":true,
  "is_manufactured":true
}')
echo "  Moringa Pack: $(echo $r | jq -r '.data.id // .error.message')"; MORINGA_ID=$(id "$r")

# Re-fetch IDs for newly created products
PROD_LIST2=$(get "masterdata/products?per_page=100")
CARDM_ID=${CARDM_ID:-$(echo "$PROD_LIST2" | jq -r '.data[] | select(.code=="CARDM") | .id')}
TURM_ID=${TURM_ID:-$(echo "$PROD_LIST2" | jq -r '.data[] | select(.code=="TURMRIC") | .id')}
MORINGA_ID=${MORINGA_ID:-$(echo "$PROD_LIST2" | jq -r '.data[] | select(.code=="MORINGA") | .id')}

# MRP for new products
if [ -n "$CARDM_ID" ]; then
  r=$(post "masterdata/mrp" '{
    "product_id":'"$CARDM_ID"',
    "material_type":"RAW_MATERIAL",
    "mrp_type":"REORDER_POINT",
    "procurement_type":"EXTERNAL",
    "purchase_lead_time":10,
    "production_lead_time":0,
    "safety_stock":15,
    "reorder_point":30,
    "min_order_qty":10,
    "lot_size_type":"LOT_FOR_LOT",
    "abc_class":"B",
    "shelf_life_days":365,
    "notes":"Very high value — order conservatively"
  }')
  echo "  Cardamom MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi
if [ -n "$TURM_ID" ]; then
  r=$(post "masterdata/mrp" '{
    "product_id":'"$TURM_ID"',
    "material_type":"RAW_MATERIAL",
    "mrp_type":"REORDER_POINT",
    "procurement_type":"EXTERNAL",
    "purchase_lead_time":5,
    "production_lead_time":0,
    "safety_stock":100,
    "reorder_point":200,
    "min_order_qty":50,
    "lot_size_type":"LOT_FOR_LOT",
    "abc_class":"C",
    "shelf_life_days":545
  }')
  echo "  Turmeric MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi
if [ -n "$MORINGA_ID" ]; then
  r=$(post "masterdata/mrp" '{
    "product_id":'"$MORINGA_ID"',
    "material_type":"SEMI_FINISHED",
    "mrp_type":"MRP",
    "procurement_type":"IN_HOUSE",
    "purchase_lead_time":0,
    "production_lead_time":3,
    "safety_stock":100,
    "reorder_point":300,
    "min_order_qty":100,
    "lot_size_type":"FIXED_LOT",
    "fixed_lot_size":200,
    "abc_class":"B",
    "shelf_life_days":365,
    "notes":"Manufactured in-house — batch size 200 packs"
  }')
  echo "  Moringa Pack MRP: $(echo $r | jq -r '.data.id // .error.message')"
fi

echo ""
echo "════════════════════════════════════════"
echo " 7. ADDITIONAL MATERIAL CATEGORIES"
echo "════════════════════════════════════════"
MC_LIST=$(get "masterdata/material-categories?per_page=50")
RAWMAT_ID=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="RAW-MAT") | .id' | head -1)
RAWSPICE_CAT=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="RAW-SPICE") | .id')
PKGMAT_CAT=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="PKG-MAT") | .id')
SEMI_CAT=$(echo "$MC_LIST" | jq -r '.data[] | select(.code=="SEMI") | .id')

r=$(post "masterdata/material-categories" '{"code":"RAW-HERB","name":"Raw Herbs and Botanicals","parent_id":'"$RAWMAT_ID"'}')
echo "  Raw Herbs: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/material-categories" '{"code":"RAW-OIL","name":"Raw Oils and Extracts","parent_id":'"$RAWMAT_ID"'}')
echo "  Raw Oils: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/material-categories" '{"code":"PKG-FLEX","name":"Flexible Packaging","parent_id":'"$PKGMAT_CAT"'}')
echo "  Flex Packaging: $(echo $r | jq -r '.data.id // .error.message')"
r=$(post "masterdata/material-categories" '{"code":"SEMI-POWD","name":"Powdered Semi-Finished","parent_id":'"$SEMI_CAT"'}')
echo "  Semi Powder: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " 8. ADDITIONAL MATERIALS"
echo "════════════════════════════════════════"
MC_LIST2=$(get "masterdata/material-categories?per_page=50")
RAWHERB_CAT=$(echo "$MC_LIST2" | jq -r '.data[] | select(.code=="RAW-HERB") | .id')
RAWOIL_CAT=$(echo "$MC_LIST2"  | jq -r '.data[] | select(.code=="RAW-OIL") | .id')
FLEXKG_CAT=$(echo "$MC_LIST2"  | jq -r '.data[] | select(.code=="PKG-FLEX") | .id')
SEMIPWD_CAT=$(echo "$MC_LIST2" | jq -r '.data[] | select(.code=="SEMI-POWD") | .id')

r=$(post "masterdata/materials" '{
  "code":"RM-CARDM-001",
  "name":"Green Cardamom Pods",
  "material_type":"RAW_MATERIAL",
  "category_id":'"${RAWSPICE_CAT:-$RAWMAT_ID}"',
  "color":"Green",
  "description":"Premium green cardamom pods from hill country",
  "notes":"Store at low humidity — highly aromatic"
}')
echo "  Cardamom Material: $(echo $r | jq -r '.data.id // .error.message')"; MAT_CARDM=$(id "$r")

r=$(post "masterdata/materials" '{
  "code":"RM-TURM-001",
  "name":"Turmeric Root (Dried)",
  "material_type":"RAW_MATERIAL",
  "category_id":'"${RAWSPICE_CAT:-$RAWMAT_ID}"',
  "color":"Orange/Yellow",
  "description":"Sun-dried turmeric root from local farms"
}')
echo "  Turmeric Material: $(echo $r | jq -r '.data.id // .error.message')"; MAT_TURM=$(id "$r")

r=$(post "masterdata/materials" '{
  "code":"RM-LGRASS-001",
  "name":"Fresh Lemongrass Stalks",
  "material_type":"RAW_MATERIAL",
  "category_id":'"${RAWHERB_CAT:-$RAWMAT_ID}"',
  "color":"Green",
  "description":"Fresh lemongrass stalks for drying"
}')
echo "  Lemongrass Material: $(echo $r | jq -r '.data.id // .error.message')"

r=$(post "masterdata/materials" '{
  "code":"RM-MORINGA-001",
  "name":"Moringa Leaves (Fresh)",
  "material_type":"RAW_MATERIAL",
  "category_id":'"${RAWHERB_CAT:-$RAWMAT_ID}"',
  "color":"Green",
  "description":"Fresh moringa leaves for powder processing"
}')
echo "  Moringa Leaf Material: $(echo $r | jq -r '.data.id // .error.message')"

r=$(post "masterdata/materials" '{
  "code":"PKG-POUCH-100",
  "name":"Stand-up Pouch 100g",
  "material_type":"RAW_MATERIAL",
  "category_id":'"${FLEXKG_CAT:-$PKGMAT_CAT}"',
  "color":"Kraft/White",
  "article_code":"SUP-100G",
  "barcode":"4099876543210",
  "description":"100g kraft stand-up zipper pouch for herbs and powders"
}')
echo "  Pouch Material: $(echo $r | jq -r '.data.id // .error.message')"; MAT_POUCH=$(id "$r")

r=$(post "masterdata/materials" '{
  "code":"SF-MORINGA-PWD",
  "name":"Moringa Leaf Powder (Bulk)",
  "material_type":"SEMI_FINISHED",
  "category_id":'"${SEMIPWD_CAT:-$SEMI_CAT}"',
  "color":"Green",
  "description":"Dried and ground moringa leaf powder — bulk pre-packaging"
}')
echo "  Moringa Powder SF: $(echo $r | jq -r '.data.id // .error.message')"; MAT_MORNPWD=$(id "$r")

r=$(post "masterdata/materials" '{
  "code":"SRV-CERT",
  "name":"Organic Certification Audit",
  "material_type":"SERVICE",
  "description":"Annual third-party organic certification audit service"
}')
echo "  Certification Service: $(echo $r | jq -r '.data.id // .error.message')"

echo ""
echo "════════════════════════════════════════"
echo " ALL DONE"
echo "════════════════════════════════════════"
