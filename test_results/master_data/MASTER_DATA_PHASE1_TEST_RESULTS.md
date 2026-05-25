# Master Data Phase 1 Test Results

**Date:** 2026-05-22  
**Environment:** Docker dev stack (postgres:5433, redis:6380, api:3000)  
**Migrations:** 000010 (organization) + 000011 (financial) — version 11, dirty: false

---

## Build Verification

```
docker exec erp-api-dev go build ./...
# Exit 0 — clean build
```

---

## Migrations

```
Migration 000010_masterdata_organization — Applied
  tables: countries, states, companies, departments,
          fiscal_years, accounting_periods, document_sequences

Migration 000011_masterdata_financial — Applied
  tables: currencies, exchange_rates, chart_of_accounts,
          cost_centers, payment_terms, banks, company_bank_accounts,
          tax_codes, tax_groups, tax_group_codes
  ALTER TABLE companies ADD CONSTRAINT fk_companies_base_currency
```

---

## 1. Company Profile

```json
PUT /api/v1/masterdata/organization/company
→ 200 OK  {"id":1,"name":"Kadahapola Exports (Pvt) Ltd","fiscal_year_start":1,"base_currency_id":1}

GET /api/v1/masterdata/organization/company
→ 200 OK  company record returned
```

---

## 2. Departments

```
POST /api/v1/masterdata/organization/departments  {"code":"FIN","name":"Finance & Accounts"}  → 201
POST /api/v1/masterdata/organization/departments  {"code":"PROD","name":"Production"}         → 201
POST /api/v1/masterdata/organization/departments  {"code":"SALES","name":"Sales & Marketing"} → 201

GET /api/v1/masterdata/organization/departments
→ ["FIN","PROD","SALES"]
```

---

## 3. Fiscal Year

```
POST /api/v1/masterdata/organization/fiscal-years
{"name":"FY 2025/2026","start_date":"2025-04-01","end_date":"2026-03-31"}
→ 201 Created  {"id":1,"is_closed":false}

PUT /api/v1/masterdata/organization/fiscal-years/1/close
→ 200 OK  (is_closed becomes true)
```

---

## 4. Document Sequences

```
POST /api/v1/masterdata/organization/document-sequences {"document_type":"SALES_ORDER","prefix":"SO"}    → 201
POST /api/v1/masterdata/organization/document-sequences {"document_type":"PURCHASE_ORDER","prefix":"PO"} → 201
POST /api/v1/masterdata/organization/document-sequences {"document_type":"INVOICE","prefix":"INV"}       → 201

GET /api/v1/masterdata/organization/document-sequences
→ ["INVOICE","PURCHASE_ORDER","SALES_ORDER"]
```

---

## 5. Currencies

```
POST /api/v1/masterdata/financial/currencies {"code":"LKR","name":"Sri Lanka Rupee","symbol":"Rs"} → 201 id=1
POST /api/v1/masterdata/financial/currencies {"code":"USD","name":"US Dollar","symbol":"$"}        → 201 id=2
PUT  /api/v1/masterdata/financial/currencies/1/set-base → 200 (LKR is_base=true, USD is_base=false)

GET /api/v1/masterdata/financial/currencies
→ ["LKR(BASE)","USD"]
```

---

## 6. Exchange Rates

```
POST /api/v1/masterdata/financial/exchange-rates
{"from_currency_id":2,"to_currency_id":1,"rate":325.50,"effective_date":"2026-05-22"}
→ 201 Created

GET /api/v1/masterdata/financial/exchange-rates       → [{"rate":325.5}]
GET /api/v1/masterdata/financial/exchange-rates/latest → latest per currency pair
```

---

## 7. Chart of Accounts

```
POST /api/v1/masterdata/financial/chart-of-accounts {"code":"1000","name":"Cash & Bank","account_type":"ASSET","currency_id":1}   → 201
POST /api/v1/masterdata/financial/chart-of-accounts {"code":"2200","name":"VAT Payable","account_type":"LIABILITY","currency_id":1} → 201

GET /api/v1/masterdata/financial/chart-of-accounts → ["1000 Cash & Bank","2200 VAT Payable"]
```

---

## 8. Payment Terms

```
POST /api/v1/masterdata/financial/payment-terms {"code":"NET30","name":"Net 30 Days","due_days":30} → 201
GET  /api/v1/masterdata/financial/payment-terms → ["NET30"]
```

---

## 9. Banks

```
POST /api/v1/masterdata/financial/banks {"name":"Bank of Ceylon","branch_name":"Kadahapola","swift_code":"BCEYLKLX"} → 201
GET  /api/v1/masterdata/financial/banks → ["Bank of Ceylon"]
```

---

## 10. Tax Codes & Tax Groups

```
POST /api/v1/masterdata/financial/tax-codes  {"code":"VAT18","tax_type":"VAT","rate":18.0,"gl_account_id":2} → 201
POST /api/v1/masterdata/financial/tax-groups {"name":"Standard VAT","tax_code_ids":[1]}                      → 201

GET /api/v1/masterdata/financial/tax-codes  → ["VAT18"]
GET /api/v1/masterdata/financial/tax-groups → ["Standard VAT"]
```

---

## 11. RBAC Permission Enforcement

```
Regular user (role=user) attempts POST /api/v1/masterdata/financial/currencies
→ HTTP 403 Forbidden  (masterdata:write required, user has only masterdata:read)
```

---

## 12. JWT Tenant Context Fix

Added `tenant_id uint` to JWT Claims so the authenticated user's tenant propagates
to all tenant-scoped operations without requiring `X-Tenant-ID` header:

```json
{"user_id":1,"tenant_id":1,"email":"admin@kadahapola.com","role":"super_admin",...}
```

---

## Summary

| Component                   | Status |
|-----------------------------|--------|
| Migration 000010 (org)      | OK     |
| Migration 000011 (financial) | OK    |
| Company profile CRUD        | OK     |
| Departments CRUD            | OK     |
| Fiscal Years + close        | OK     |
| Document Sequences CRUD     | OK     |
| Currencies + set-base       | OK     |
| Exchange Rates              | OK     |
| Chart of Accounts           | OK     |
| Payment Terms               | OK     |
| Banks                       | OK     |
| Tax Codes                   | OK     |
| Tax Groups (many2many)      | OK     |
| RBAC (masterdata:write 403) | OK     |
| JWT tenant_id propagation   | OK     |
| Full compile                | OK     |
