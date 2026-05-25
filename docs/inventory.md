# Inventory Module — Reference Documentation

> Module path: `internal/modules/inventory`  
> Base URL: `/api/v1/inventory`  
> Last updated: 2026-05-23

---

## Table of Contents

1. [Overview](#1-overview)
2. [Stock Architecture](#2-stock-architecture)
3. [Document Flow](#3-document-flow)
4. [Status Lifecycles](#4-status-lifecycles)
5. [Stock Balance Rules](#5-stock-balance-rules)
6. [Goods Issue Reasons](#6-goods-issue-reasons)
7. [Stock Adjustment Reasons](#7-stock-adjustment-reasons)
8. [Quality Check Result Logic](#8-quality-check-result-logic)
9. [RBAC Permissions](#9-rbac-permissions)
10. [API Reference](#10-api-reference)
    - [Material Requests](#101-material-requests)
    - [MR Lines](#102-mr-lines)
    - [Goods Transfers](#103-goods-transfers)
    - [Transfer Lines](#104-transfer-lines)
    - [Goods Issues](#105-goods-issues)
    - [Issue Lines](#106-issue-lines)
    - [Stock Adjustments](#107-stock-adjustments)
    - [Adjustment Lines](#108-adjustment-lines)
    - [Quality Checks](#109-quality-checks)
    - [QC Lines](#1010-qc-lines)
    - [Stock Balance](#1011-stock-balance)
11. [Data Models](#11-data-models)

---

## 1. Overview

The Inventory module manages all stock movements within the warehouse — from internal material requests through physical transfers, consumption tracking, corrections, and quality control. It works alongside the Procurement module (which handles inbound goods) to maintain accurate real-time stock levels.

```
Goods Receipt (Procurement)
         ↓ stock_balances += received qty
         
Material Request → Goods Issue
                        ↓ stock_balances -= issued qty

Goods Transfer (warehouse A → warehouse B)
         ↓ stock_balances: A -= qty, B += qty

Stock Adjustment (physical count correction)
         ↓ stock_balances ± qty

Quality Check (inspection, linked to any reference)
         ↓ informational — failed qty handled via Goods Issue (QC_REJECTION)
```

All documents are tenant-scoped and auto-coded via the shared `document_sequences` table.

---

## 2. Stock Architecture

Stock is maintained in two separate layers with distinct responsibilities:

### Layer 1 — `stock_balances` (PostgreSQL `erp_db`)

The **current quantity** per product/variant/warehouse/location. Updated atomically inside the same PostgreSQL transaction as every stock-movement operation.

```sql
-- unique per (tenant, product, variant, warehouse, location) — NULLs treated as equal
CONSTRAINT uidx_stock_balances UNIQUE NULLS NOT DISTINCT (
    tenant_id, product_id, variant_id, warehouse_id, location_id
)
```

| Column | Description |
|---|---|
| `quantity` | Current on-hand quantity |
| `reserved_qty` | Quantity reserved by unfulfilled orders (future use) |

**Every module that moves stock upserts this table:**

| Trigger | Delta |
|---|---|
| GRN confirmed (Procurement) | `+= received_qty` (after transfer_ratio) |
| Goods Issue confirmed | `-= issued_qty` (blocked if insufficient) |
| Transfer received | From: `-= qty`, To: `+= qty` |
| Adjustment confirmed | `+= qty` (positive or negative) |

### Layer 2 — `stock_ledger` (ClickHouse `erp_ledger`)

An **append-only movement log** written after the PostgreSQL transaction commits (best-effort). Never updated or deleted. Powers the movement history tab on each module page without touching the ERP database.

| Transaction type | Who writes it |
|---|---|
| `PURCHASE` | Procurement — GRN confirm |
| `GOODS_ISSUE` | Inventory — Issue confirm |
| `TRANSFER_OUT` | Inventory — Transfer receive (source) |
| `TRANSFER_IN` | Inventory — Transfer receive (destination) |
| `ADJUSTMENT` | Inventory — Adjustment confirm |

> If ClickHouse is unavailable, the stock movement still completes. The history entry is silently skipped and logged as an error. `stock_balances` is always the authoritative source of truth.

---

## 3. Document Flow

### Material Request → Goods Issue

```
Create MR (DRAFT)
    ↓ add lines
Approve MR (APPROVED)
    ↓
Create Goods Issue referencing MR (issue_reason = PRODUCTION or as needed)
    ↓ add lines
Confirm Issue (CONFIRMED) → stock_balances decremented
```

When a Goods Issue is confirmed with `reference_type = "MATERIAL_REQUEST"`, the linked MR's `issued_qty` per line is updated, and the MR status transitions to FULFILLED or PARTIAL.

### Goods Transfer

```
Create Transfer (DRAFT)  — from_warehouse → to_warehouse
    ↓ add lines
Send Transfer (IN_TRANSIT)  — physical dispatch
    ↓
Receive Transfer (RECEIVED) — stock_balances adjusted atomically
```

### Stock Adjustment

```
Create Adjustment (DRAFT)
    ↓ add lines (quantity can be negative)
Confirm Adjustment (CONFIRMED) → stock_balances adjusted
```

### Quality Check

```
Create Quality Check (PENDING)
    ↓ add lines with product + qty_checked
Start QC (IN_PROGRESS)
    ↓ fill in qty_passed / qty_failed per line
Submit QC (PASSED / FAILED / PARTIAL — auto-computed)
```

Failed quantities require a separate Goods Issue with `issue_reason = QC_REJECTION` to deduct from stock.

---

## 4. Status Lifecycles

### Material Request

```
DRAFT ──approve──→ APPROVED ──(via GI confirm)──→ FULFILLED
                              ──(partial GI)──→ PARTIAL
      ──reject──→ REJECTED
```

> Only DRAFT requests can be approved or rejected.

### Goods Transfer

```
DRAFT ──send──→ IN_TRANSIT ──receive──→ RECEIVED
      ──cancel──→ CANCELLED
IN_TRANSIT ──cancel──→ CANCELLED
```

> Send and Receive require `inventory:approve` permission.

### Goods Issue

```
DRAFT ──confirm──→ CONFIRMED
```

> Confirm checks `stock_balances.quantity >= line.quantity` for every line. Returns 400 if any line has insufficient stock.

### Stock Adjustment

```
DRAFT ──confirm──→ CONFIRMED
```

> Quantity can be positive (stock found) or negative (shrinkage, write-off).

### Quality Check

```
PENDING ──start──→ IN_PROGRESS ──submit──→ PASSED
                                        ──→ FAILED
                                        ──→ PARTIAL
```

> Submit auto-computes the overall status from line results:
> - All lines PASSED → PASSED
> - All lines FAILED → FAILED  
> - Mixed → PARTIAL

---

## 5. Stock Balance Rules

| Rule | Detail |
|---|---|
| Negative stock blocked | Goods Issue confirm returns 400 if `stock_balances.quantity < issued qty` for any line |
| Atomicity | `stock_balances` is updated inside the PostgreSQL transaction — no partial updates possible |
| NULL handling | `variant_id = NULL` and `location_id = NULL` are treated as distinct values in the unique constraint (product without variant / warehouse without location breakdown) |
| Available stock | `quantity - reserved_qty` (reserved_qty populated when Sales Orders are confirmed) |

---

## 6. Goods Issue Reasons

| Reason | Use case |
|---|---|
| `PRODUCTION` | Materials consumed in a production/work order |
| `SALE` | Direct sale without a formal sales order |
| `EXPENSE` | Internal consumption (office supplies, etc.) |
| `DAMAGE` | Goods damaged in warehouse |
| `ADJUSTMENT` | General manual correction |
| `QC_REJECTION` | Goods rejected by quality check |
| `OTHER` | Catch-all |

---

## 7. Stock Adjustment Reasons

| Reason | Use case |
|---|---|
| `STOCKTAKE` | Physical count differs from system quantity |
| `DAMAGE` | Stock written off due to damage |
| `EXPIRY` | Stock expired and removed |
| `WRITE_OFF` | Irrecoverable loss |
| `CORRECTION` | Data entry correction |
| `OTHER` | Catch-all |

Adjustment line quantities:
- **Positive** → stock increased (found extra units, received without GRN)
- **Negative** → stock decreased (units missing, written off)

---

## 8. Quality Check Result Logic

QC line result is set manually by the inspector:

| Line result | Meaning |
|---|---|
| `PENDING` | Not yet inspected |
| `PASSED` | All units pass |
| `FAILED` | All units fail |
| `PARTIAL` | Some units pass, some fail (use `qty_passed` + `qty_failed`) |

On `submit`, the overall QC status is derived:

```
all lines PASSED                → QC status = PASSED
all lines FAILED                → QC status = FAILED
any mix (including PARTIAL lines) → QC status = PARTIAL
```

Failed units must be handled with a Goods Issue (`QC_REJECTION`) to deduct them from `stock_balances`.

---

## 9. RBAC Permissions

| Permission | Grants |
|---|---|
| `inventory:read` | GET all list and detail endpoints |
| `inventory:write` | POST/PUT create and update (documents + lines) |
| `inventory:delete` | DELETE lines on draft documents |
| `inventory:approve` | Approve/Reject MR, Send/Receive Transfer, Confirm Issue, Confirm Adjustment, Start/Submit QC |

---

## 10. API Reference

All endpoints require `Authorization: Bearer <token>`. State-changing endpoints are audit-logged to ClickHouse.

---

### 10.1 Material Requests

#### `GET /inventory/material-requests`

List all material requests for the tenant.

| Query param | Type | Description |
|---|---|---|
| `status` | string | Filter: DRAFT, APPROVED, REJECTED, FULFILLED, PARTIAL |

**Response 200**
```json
{
  "status": "success",
  "message": "material requests retrieved",
  "data": [
    {
      "id": 1,
      "code": "MR00001",
      "requested_by": 3,
      "warehouse_id": 1,
      "needed_date": "2026-06-01",
      "status": "DRAFT",
      "created_by": 3,
      "created_at": "2026-05-23T10:00:00Z"
    }
  ]
}
```

---

#### `POST /inventory/material-requests`

Create a new material request in DRAFT.

**Request body**
```json
{
  "requested_by": 3,
  "department_id": 2,
  "warehouse_id": 1,
  "needed_date": "2026-06-01",
  "notes": "Monthly production materials"
}
```

**Response 201** — returns the created MR with auto-generated code (MR00001).

---

#### `GET /inventory/material-requests/:id`

Returns the MR including its lines.

---

#### `PUT /inventory/material-requests/:id`

Update a DRAFT material request.

**Request body**
```json
{
  "warehouse_id": 1,
  "needed_date": "2026-06-05",
  "notes": "Updated"
}
```

---

#### `POST /inventory/material-requests/:id/approve`

Transitions DRAFT → APPROVED. Sets `approved_by` and `approved_at`.

> Requires `inventory:approve`

**Response 200**
```json
{ "status": "success", "message": "material request approved" }
```

---

#### `POST /inventory/material-requests/:id/reject`

Transitions DRAFT → REJECTED.

> Requires `inventory:approve`

**Request body**
```json
{ "reason": "Budget not approved this month" }
```

---

### 10.2 MR Lines

#### `GET /inventory/material-requests/:id/lines`

#### `POST /inventory/material-requests/:id/lines`

**Request body**
```json
{
  "product_id": 5,
  "variant_id": null,
  "uom_id": 1,
  "requested_qty": 50,
  "notes": "Standard grade only"
}
```

#### `PUT /inventory/material-requests/:id/lines/:lineId`

Same body as Add.

#### `DELETE /inventory/material-requests/:id/lines/:lineId`

---

### 10.3 Goods Transfers

#### `GET /inventory/transfers`

| Query param | Description |
|---|---|
| `status` | DRAFT, IN_TRANSIT, RECEIVED, CANCELLED |

#### `POST /inventory/transfers`

**Request body**
```json
{
  "from_warehouse_id": 1,
  "to_warehouse_id": 2,
  "transfer_date": "2026-06-01",
  "notes": "Monthly replenishment to branch"
}
```

**Response 201** — returns transfer with auto-generated code (GT00001).

#### `GET /inventory/transfers/:id`

Returns transfer with lines.

#### `PUT /inventory/transfers/:id`

Update DRAFT transfer header (same body as create).

#### `POST /inventory/transfers/:id/send`

DRAFT → IN_TRANSIT. Sets `sent_by` and `sent_at`.

> Stock is **not** adjusted at send time — only on receive.

#### `POST /inventory/transfers/:id/receive`

IN_TRANSIT → RECEIVED. **Atomically adjusts `stock_balances`**:
- From warehouse/location: `quantity -= line.quantity`
- To warehouse/location: `quantity += line.quantity`

Then writes `TRANSFER_OUT` and `TRANSFER_IN` pairs to ClickHouse (best-effort).

---

### 10.4 Transfer Lines

#### `GET /inventory/transfers/:id/lines`

#### `POST /inventory/transfers/:id/lines`

**Request body**
```json
{
  "product_id": 5,
  "variant_id": null,
  "from_location_id": 3,
  "to_location_id": 7,
  "uom_id": 1,
  "quantity": 100
}
```

#### `PUT /inventory/transfers/:id/lines/:lineId`

#### `DELETE /inventory/transfers/:id/lines/:lineId`

---

### 10.5 Goods Issues

#### `GET /inventory/issues`

| Query param | Description |
|---|---|
| `status` | DRAFT, CONFIRMED |
| `reason` | PRODUCTION, SALE, EXPENSE, DAMAGE, ADJUSTMENT, QC_REJECTION, OTHER |

#### `POST /inventory/issues`

**Request body**
```json
{
  "issue_date": "2026-06-01",
  "warehouse_id": 1,
  "issue_reason": "PRODUCTION",
  "reference_type": "MATERIAL_REQUEST",
  "reference_id": 7,
  "notes": "For Work Order WO-2026-001"
}
```

#### `GET /inventory/issues/:id`

Returns issue with lines.

#### `PUT /inventory/issues/:id`

Update DRAFT issue header.

#### `POST /inventory/issues/:id/confirm`

DRAFT → CONFIRMED.

**Stock check** — before confirming, the system verifies for each line:
```
stock_balances.quantity >= line.quantity
```
If any line fails this check, the entire confirm is rejected with:
```json
{ "status": "error", "message": "insufficient stock for product 5" }
```

On success, `stock_balances` is decremented atomically and `GOODS_ISSUE` entries are written to ClickHouse.

---

### 10.6 Issue Lines

#### `GET /inventory/issues/:id/lines`

#### `POST /inventory/issues/:id/lines`

**Request body**
```json
{
  "product_id": 5,
  "variant_id": null,
  "location_id": 3,
  "uom_id": 1,
  "quantity": 20,
  "unit_cost": 250.00
}
```

`total_cost` is computed automatically as `quantity × unit_cost`.

#### `PUT /inventory/issues/:id/lines/:lineId`

#### `DELETE /inventory/issues/:id/lines/:lineId`

---

### 10.7 Stock Adjustments

#### `GET /inventory/adjustments`

| Query param | Description |
|---|---|
| `status` | DRAFT, CONFIRMED |

#### `POST /inventory/adjustments`

**Request body**
```json
{
  "adjustment_date": "2026-06-01",
  "warehouse_id": 1,
  "adjust_reason": "STOCKTAKE",
  "notes": "Quarterly physical count"
}
```

#### `GET /inventory/adjustments/:id`

Returns adjustment with lines.

#### `PUT /inventory/adjustments/:id`

Update DRAFT adjustment header.

#### `POST /inventory/adjustments/:id/confirm`

DRAFT → CONFIRMED.

Applies signed quantity deltas to `stock_balances` atomically and writes `ADJUSTMENT` entries to ClickHouse.

> No stock minimum check — negative balances are allowed for adjustments (system correction scenarios).

---

### 10.8 Adjustment Lines

#### `GET /inventory/adjustments/:id/lines`

#### `POST /inventory/adjustments/:id/lines`

**Request body**
```json
{
  "product_id": 5,
  "variant_id": null,
  "location_id": 3,
  "uom_id": 1,
  "quantity": -5,
  "unit_cost": 250.00,
  "notes": "5 units missing in physical count"
}
```

> `quantity` may be negative (deduction) or positive (addition).

#### `PUT /inventory/adjustments/:id/lines/:lineId`

#### `DELETE /inventory/adjustments/:id/lines/:lineId`

---

### 10.9 Quality Checks

#### `GET /inventory/quality-checks`

| Query param | Description |
|---|---|
| `status` | PENDING, IN_PROGRESS, PASSED, FAILED, PARTIAL |

#### `POST /inventory/quality-checks`

**Request body**
```json
{
  "reference_type": "GOODS_RECEIPT",
  "reference_id": 12,
  "warehouse_id": 1,
  "check_date": "2026-06-01",
  "inspector_id": 4,
  "notes": "Standard incoming inspection"
}
```

`reference_type` and `reference_id` are optional — a QC can be standalone.

#### `GET /inventory/quality-checks/:id`

Returns QC with lines.

#### `POST /inventory/quality-checks/:id/start`

PENDING → IN_PROGRESS. Sets `inspector_id` timestamp.

#### `POST /inventory/quality-checks/:id/submit`

IN_PROGRESS → PASSED / FAILED / PARTIAL (auto-computed from line results).

**Result logic:**

| Line results | Overall status |
|---|---|
| All PASSED | PASSED |
| All FAILED | FAILED |
| Any mix | PARTIAL |

> QC does not automatically adjust stock. Use a Goods Issue (`QC_REJECTION`) to deduct failed units from `stock_balances`.

---

### 10.10 QC Lines

#### `GET /inventory/quality-checks/:id/lines`

#### `POST /inventory/quality-checks/:id/lines`

**Request body**
```json
{
  "product_id": 5,
  "variant_id": null,
  "qty_checked": 100,
  "qty_passed": 95,
  "qty_failed": 5,
  "result": "PARTIAL",
  "rejection_reason": "Surface scratches on 5 units",
  "notes": ""
}
```

#### `PUT /inventory/quality-checks/:id/lines/:lineId`

Same body as Add.

---

### 10.11 Stock Balance

#### `GET /inventory/stock`

Returns current on-hand quantities from the `stock_balances` table. Excludes rows with `quantity = 0`.

| Query param | Type | Description |
|---|---|---|
| `warehouse_id` | uint | Filter to a specific warehouse |
| `product_id` | uint | Filter to a specific product |

**Response 200**
```json
{
  "status": "success",
  "message": "stock balance retrieved",
  "data": [
    {
      "product_id": 5,
      "variant_id": null,
      "warehouse_id": 1,
      "location_id": 3,
      "quantity": 245.0000,
      "reserved_qty": 0.0000
    }
  ]
}
```

> `quantity - reserved_qty` = available to issue.

---

## 11. Data Models

### MaterialRequest

| Field | Type | Notes |
|---|---|---|
| `id` | uint | Auto-generated |
| `tenant_id` | uint | Tenant scope |
| `code` | string | Auto-generated (MR00001) |
| `requested_by` | uint | User ID of requester |
| `department_id` | *uint | Optional department reference |
| `warehouse_id` | uint | Source warehouse for materials |
| `needed_date` | date | When materials are needed |
| `status` | string | DRAFT / APPROVED / REJECTED / FULFILLED / PARTIAL |
| `notes` | string | |
| `approved_by` | *uint | Set on approve |
| `approved_at` | *datetime | |
| `rejected_by` | *uint | Set on reject |
| `reject_reason` | string | Required on reject |
| `created_by` | uint | |

### MRLine

| Field | Type | Notes |
|---|---|---|
| `mr_id` | uint | Parent MR |
| `product_id` | uint | |
| `variant_id` | *uint | Optional product variant |
| `uom_id` | uint | Unit of measure |
| `requested_qty` | decimal(18,4) | Quantity needed |
| `issued_qty` | decimal(18,4) | Quantity fulfilled so far |

### GoodsTransfer

| Field | Type | Notes |
|---|---|---|
| `code` | string | Auto-generated (GT00001) |
| `from_warehouse_id` | uint | Source warehouse |
| `to_warehouse_id` | uint | Destination warehouse |
| `transfer_date` | date | |
| `status` | string | DRAFT / IN_TRANSIT / RECEIVED / CANCELLED |
| `sent_by` | *uint | Set on send |
| `sent_at` | *datetime | |
| `received_by` | *uint | Set on receive |
| `received_at` | *datetime | |

### GTLine

| Field | Type | Notes |
|---|---|---|
| `from_location_id` | *uint | Optional source bin/location |
| `to_location_id` | *uint | Optional destination bin/location |
| `quantity` | decimal(18,4) | Quantity dispatched |
| `received_qty` | decimal(18,4) | Quantity received (for partial receipt tracking) |

### GoodsIssue

| Field | Type | Notes |
|---|---|---|
| `code` | string | Auto-generated (GI00001) |
| `issue_date` | date | |
| `warehouse_id` | uint | |
| `issue_reason` | string | PRODUCTION / SALE / EXPENSE / DAMAGE / ADJUSTMENT / QC_REJECTION / OTHER |
| `reference_type` | string | Optional: MATERIAL_REQUEST, WORK_ORDER, etc. |
| `reference_id` | *uint | Optional: ID of the referenced document |
| `status` | string | DRAFT / CONFIRMED |

### GILine

| Field | Type | Notes |
|---|---|---|
| `quantity` | decimal(18,4) | Must not exceed available stock |
| `unit_cost` | decimal(18,4) | |
| `total_cost` | decimal(18,4) | Computed: quantity × unit_cost |

### StockAdjustment

| Field | Type | Notes |
|---|---|---|
| `code` | string | Auto-generated (SA00001) |
| `adjustment_date` | date | |
| `warehouse_id` | uint | |
| `adjust_reason` | string | STOCKTAKE / DAMAGE / EXPIRY / WRITE_OFF / CORRECTION / OTHER |
| `status` | string | DRAFT / CONFIRMED |

### SALine

| Field | Type | Notes |
|---|---|---|
| `quantity` | decimal(18,4) | Positive = add, Negative = deduct |
| `unit_cost` | decimal(18,4) | |

### QualityCheck

| Field | Type | Notes |
|---|---|---|
| `code` | string | Auto-generated (QC00001) |
| `reference_type` | string | Optional: GOODS_RECEIPT, GOODS_TRANSFER, etc. |
| `reference_id` | *uint | |
| `warehouse_id` | uint | |
| `check_date` | date | |
| `status` | string | PENDING / IN_PROGRESS / PASSED / FAILED / PARTIAL |
| `inspector_id` | *uint | |

### QCLine

| Field | Type | Notes |
|---|---|---|
| `qty_checked` | decimal(18,4) | Total units inspected |
| `qty_passed` | decimal(18,4) | Units that passed |
| `qty_failed` | decimal(18,4) | Units that failed |
| `result` | string | PENDING / PASSED / FAILED / PARTIAL |
| `rejection_reason` | string | Required when result is FAILED or PARTIAL |

### StockBalance (response only)

| Field | Type | Notes |
|---|---|---|
| `product_id` | uint | |
| `variant_id` | *uint | |
| `warehouse_id` | uint | |
| `location_id` | *uint | |
| `quantity` | decimal(18,4) | Current on-hand |
| `reserved_qty` | decimal(18,4) | Reserved by open orders |
