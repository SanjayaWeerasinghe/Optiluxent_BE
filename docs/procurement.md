# Procurement Module — Reference Documentation

> Module path: `internal/modules/procurement`  
> Base URL: `/api/v1/procurement`  
> Last updated: 2026-05-23

---

## Table of Contents

1. [Overview](#1-overview)
2. [Document Flow](#2-document-flow)
3. [Status Lifecycles](#3-status-lifecycles)
4. [Transfer Ratio](#4-transfer-ratio)
5. [Auto-Invoice Behaviour](#5-auto-invoice-behaviour)
6. [GRN Guard](#6-grn-guard)
7. [Partial Payment](#7-partial-payment)
8. [Financial Integration (Event Bus)](#8-financial-integration-event-bus)
9. [RBAC Permissions](#9-rbac-permissions)
10. [API Reference](#10-api-reference)
    - [Purchase Requests](#101-purchase-requests)
    - [PR Items](#102-pr-items)
    - [Purchase Orders](#103-purchase-orders)
    - [PO Items](#104-po-items)
    - [Goods Receipts](#105-goods-receipts)
    - [GRN Items](#106-grn-items)
    - [Purchase Invoices](#107-purchase-invoices)
    - [Invoice Lines](#108-invoice-lines)
11. [Data Models](#11-data-models)

---

## 1. Overview

The Procurement module manages the full purchasing lifecycle from requisition through to invoice payment. It implements a 3-way matching workflow:

```
Purchase Request (PR)
        ↓
Purchase Order (PO)
        ↓
Goods Receipt (GRN)     ←── stock ledger updated on confirm
        ↓
Purchase Invoice (PINV) ←── auto-created on PO confirm, lines auto-appended on GRN confirm
        ↓
Payment Recording
```

All documents are tenant-scoped. Document codes are generated automatically using the shared sequence service (`NextCode`).

---

## 2. Document Flow

| Step | Action | Who triggers |
|------|--------|-------------|
| 1 | Create PR in DRAFT | Requester |
| 2 | Add items to PR | Requester |
| 3 | Submit PR → PENDING_APPROVAL | Requester |
| 4 | Approve or Reject PR | Approver |
| 5 | Create PO (optionally from approved PR) | Buyer |
| 6 | Add items to PO | Buyer |
| 7 | Confirm PO → CONFIRMED | Approver — **auto-creates a DRAFT invoice** |
| 8 | Create GRN (linked to PO) | Warehouse |
| 9 | Add items to GRN | Warehouse |
| 10 | Confirm GRN — **stock ledger updated; invoice lines auto-appended** | Warehouse |
| 11 | Edit/add supplier invoice details to PINV | AP clerk |
| 12 | Post invoice → POSTED — **procurement.invoice.posted event emitted** | AP clerk |
| 13 | Record payments (partial or full) — **procurement.payment.recorded event emitted** | AP clerk |

---

## 3. Status Lifecycles

### Purchase Request

```
DRAFT → PENDING_APPROVAL → APPROVED
                         → REJECTED
DRAFT / PENDING_APPROVAL → CANCELLED
```

- Only `DRAFT` PRs can be edited or have items added/removed.
- Only `PENDING_APPROVAL` PRs can be approved or rejected.

### Purchase Order

```
DRAFT → CONFIRMED → PARTIAL → RECEIVED
DRAFT / CONFIRMED / PARTIAL → CANCELLED
```

- Only `DRAFT` POs can be edited or have items added/removed/deleted.
- `PARTIAL` = at least one GRN has been confirmed.
- `RECEIVED` = all lines fully received (set by GRN confirmation logic).

### Goods Receipt

```
DRAFT → CONFIRMED
DRAFT → CANCELLED
```

- Items can only be added/edited on `DRAFT` GRNs.
- Confirmation triggers stock ledger update (irreversible).

### Purchase Invoice

```
DRAFT → POSTED → PARTIAL → PAID
DRAFT / POSTED / PARTIAL → CANCELLED
```

- Lines can only be added/removed on `DRAFT` invoices.
- Once `POSTED`, the invoice is locked from further line edits.
- `PARTIAL` = some payment recorded but total not reached.
- `PAID` = `paid_amount >= total_amount`.

---

## 4. Transfer Ratio

Transfer ratio (`transfer_ratio`) bridges a **purchase UOM** to the **stock UOM** when they differ (e.g., buying in cartons of 12 units).

| Field | Formula |
|-------|---------|
| `stock_qty` | `grn_qty × transfer_ratio` |
| `stock_unit_cost` | `grn_unit_cost ÷ transfer_ratio` |

**Default:** `1.0` (purchase UOM = stock UOM).

Transfer ratio is carried on all line types:

- `purchase_request_lines.transfer_ratio` — estimated conversion for budgeting
- `purchase_order_lines.transfer_ratio` — actual conversion for receiving
- `goods_receipt_lines.transfer_ratio` — applied when confirming to stock ledger

**Example:**  
Buying 10 cartons at LKR 300/carton; carton = 12 bottles.  
`transfer_ratio = 12` → stock receives 120 bottles at LKR 25/bottle.

---

## 5. Auto-Invoice Behaviour

### On PO Confirm

When a PO is confirmed, the system automatically creates a DRAFT `PurchaseInvoice` linked to that PO:

- Inherits `supplier_id`, `currency_id`, `exchange_rate`, `payment_term_id` from PO
- `status = DRAFT`, `total_amount = 0` (no lines yet)
- `invoice_date = today`

### On GRN Confirm

When a GRN is confirmed (and it is linked to a PO), the system finds the PO's DRAFT invoice and appends new lines — one per GRN line:

- `quantity` from GRN line
- `unit_price` sourced from the linked PO line (`po_line_id`) — falls back to `unit_cost` if no PO line
- Invoice totals are recalculated

This is **best-effort**: if the invoice append fails, the GRN confirmation still succeeds (stock is already updated).

### Manual Invoice Creation

You can also create invoices manually via `POST /purchase-invoices`. A PO reference (`po_id`) is required. Use this when auto-creation is insufficient or when back-entering historical invoices.

---

## 6. GRN Guard

A new GRN **cannot be linked** to a PO if that PO's invoice is in a locked state:

| Invoice Status | GRN creation allowed? |
|---------------|----------------------|
| DRAFT | Yes |
| CANCELLED | Yes |
| POSTED | **No** |
| PARTIAL | **No** |
| PAID | **No** |

Error response: `400 purchase order already has a completed invoice — cannot add more receipts`

---

## 7. Partial Payment

`POST /purchase-invoices/:id/pay` accepts a partial amount. Rules:

- Invoice must be `POSTED` or `PARTIAL` (not `DRAFT`, `PAID`, or `CANCELLED`)
- `amount` must be > 0 and ≤ remaining balance (`total_amount - paid_amount`)
- After each payment: `paid_amount += amount`
  - If `paid_amount >= total_amount` → status becomes `PAID`
  - Otherwise → status becomes `PARTIAL`

---

## 8. Financial Integration (Event Bus)

The procurement module publishes events to the `erp:procurement` Redis Stream. A future financial module subscribes to these to create accounting entries without any code change to procurement.

### Stream: `erp:procurement`

#### `procurement.invoice.posted`

Published when a DRAFT invoice is posted.

```json
{
  "invoice_id": 12,
  "code": "PINV-2026-0001",
  "po_id": 5,
  "supplier_id": 3,
  "invoice_date": "2026-05-23",
  "due_date": "2026-06-22",
  "currency_id": 1,
  "exchange_rate": 1.0,
  "subtotal": 3600.00,
  "tax_amount": 0.00,
  "discount_amount": 0.00,
  "total_amount": 3600.00,
  "posted_by": 1
}
```

**Intended financial action:** Create AP liability journal entry (Dr Purchases / Cr Accounts Payable).

#### `procurement.payment.recorded`

Published after each payment is recorded.

```json
{
  "invoice_id": 12,
  "code": "PINV-2026-0001",
  "supplier_id": 3,
  "amount": 1800.00,
  "paid_amount": 1800.00,
  "total_amount": 3600.00,
  "remaining": 1800.00,
  "status": "PARTIAL",
  "payment_date": "2026-05-23",
  "payment_ref": "CHQ-001"
}
```

**Intended financial action:** Dr Accounts Payable / Cr Cash or Bank.

### Subscribing (future financial module)

```go
func (m *FinancialModule) RegisterEvents(bus events.EventBus) {
    bus.Subscribe(events.StreamProcurement, "financial", "fin-worker", func(ev events.Event) error {
        switch ev.Type {
        case events.TypeInvoicePosted:
            // unmarshal InvoicePostedPayload → create AP journal entry
        case events.TypePaymentRecorded:
            // unmarshal PaymentRecordedPayload → debit AP, credit cash/bank
        }
        return nil
    })
}
```

---

## 9. RBAC Permissions

All procurement routes require a valid JWT. Actions are gated by Casbin:

| Permission | Grants access to |
|-----------|-----------------|
| `procurement:read` | GET endpoints (list, get, get items) |
| `procurement:write` | POST/PUT for create, update, submit, cancel, pay |
| `procurement:delete` | DELETE endpoints |
| `procurement:approve` | Approve/reject PR; confirm PO; confirm GRN; post invoice |

---

## 10. API Reference

All requests require:
```
Authorization: Bearer <access_token>
Content-Type: application/json   (for POST/PUT)
```

---

### 10.1 Purchase Requests

#### List PRs
```
GET /api/v1/procurement/purchase-requests?status=DRAFT
```
Query params: `status` (optional) — `DRAFT | PENDING_APPROVAL | APPROVED | REJECTED | CANCELLED`

#### Create PR
```
POST /api/v1/procurement/purchase-requests
```
```json
{
  "request_date": "2026-05-23",
  "required_date": "2026-06-01",
  "requested_by": 2,
  "department_id": 1,
  "notes": "Monthly raw material requisition"
}
```

#### Get PR
```
GET /api/v1/procurement/purchase-requests/:id
```

#### Update PR *(DRAFT only)*
```
PUT /api/v1/procurement/purchase-requests/:id
```
```json
{
  "required_date": "2026-06-10",
  "notes": "Updated delivery date"
}
```

#### Delete PR *(DRAFT only)*
```
DELETE /api/v1/procurement/purchase-requests/:id
```

#### Submit PR
```
POST /api/v1/procurement/purchase-requests/:id/submit
```
Transitions `DRAFT → PENDING_APPROVAL`. Requires at least one item.

#### Approve PR
```
POST /api/v1/procurement/purchase-requests/:id/approve
```
Transitions `PENDING_APPROVAL → APPROVED`.

#### Reject PR
```
POST /api/v1/procurement/purchase-requests/:id/reject
```
Transitions `PENDING_APPROVAL → REJECTED`.

#### Cancel PR
```
POST /api/v1/procurement/purchase-requests/:id/cancel
```
Allowed from `DRAFT` or `PENDING_APPROVAL`.

---

### 10.2 PR Items

#### List PR Items
```
GET /api/v1/procurement/purchase-requests/:id/items
```

#### Add PR Item *(DRAFT PR only)*
```
POST /api/v1/procurement/purchase-requests/:id/items
```
```json
{
  "product_id": 10,
  "variant_id": null,
  "description": "Raw cotton bale",
  "quantity": 50,
  "uom_id": 3,
  "estimated_price": 850.00,
  "transfer_ratio": 1.0,
  "currency_id": 1,
  "notes": ""
}
```

#### Get PR Item
```
GET /api/v1/procurement/purchase-requests/:id/items/:itemId
```

#### Update PR Item *(DRAFT PR only)*
```
PUT /api/v1/procurement/purchase-requests/:id/items/:itemId
```
Same body as Add.

#### Delete PR Item *(DRAFT PR only)*
```
DELETE /api/v1/procurement/purchase-requests/:id/items/:itemId
```

---

### 10.3 Purchase Orders

#### List POs
```
GET /api/v1/procurement/purchase-orders?status=DRAFT&supplier_id=3
```
Query params: `status`, `supplier_id` (both optional)

#### Create PO
```
POST /api/v1/procurement/purchase-orders
```
```json
{
  "supplier_id": 3,
  "pr_id": 1,
  "order_date": "2026-05-23",
  "expected_date": "2026-06-05",
  "currency_id": 1,
  "exchange_rate": 1.0,
  "payment_term_id": 2,
  "warehouse_id": 1,
  "notes": "Linked to PR-2026-0001"
}
```

#### Get PO
```
GET /api/v1/procurement/purchase-orders/:id
```

#### Update PO *(DRAFT only)*
```
PUT /api/v1/procurement/purchase-orders/:id
```
```json
{
  "expected_date": "2026-06-10",
  "exchange_rate": 1.05,
  "payment_term_id": 3,
  "discount_amount": 100.00,
  "notes": "Revised expected date"
}
```

#### Delete PO *(DRAFT only)*
```
DELETE /api/v1/procurement/purchase-orders/:id
```

#### Confirm PO
```
POST /api/v1/procurement/purchase-orders/:id/confirm
```
Transitions `DRAFT → CONFIRMED`. Requires at least one item. **Auto-creates a DRAFT invoice.**

#### Cancel PO
```
POST /api/v1/procurement/purchase-orders/:id/cancel
```
Not allowed if status is `RECEIVED` or already `CANCELLED`.

---

### 10.4 PO Items

#### List PO Items
```
GET /api/v1/procurement/purchase-orders/:id/items
```

#### Add PO Item *(DRAFT PO only)*
```
POST /api/v1/procurement/purchase-orders/:id/items
```
```json
{
  "product_id": 10,
  "variant_id": null,
  "description": "Raw cotton bale",
  "quantity": 50,
  "uom_id": 3,
  "unit_price": 900.00,
  "discount_pct": 0,
  "tax_code_id": 1,
  "transfer_ratio": 1.0,
  "notes": ""
}
```
`line_total = quantity × unit_price × (1 - discount_pct / 100)`

#### Get PO Item
```
GET /api/v1/procurement/purchase-orders/:id/items/:itemId
```

#### Update PO Item *(DRAFT PO only)*
```
PUT /api/v1/procurement/purchase-orders/:id/items/:itemId
```
Same body as Add.

#### Delete PO Item *(DRAFT PO only)*
```
DELETE /api/v1/procurement/purchase-orders/:id/items/:itemId
```

---

### 10.5 Goods Receipts

#### List GRNs
```
GET /api/v1/procurement/goods-receipts?status=DRAFT&po_id=5
```
Query params: `status`, `po_id` (both optional)

#### Create GRN
```
POST /api/v1/procurement/goods-receipts
```
```json
{
  "po_id": 5,
  "supplier_id": 3,
  "receipt_date": "2026-05-23",
  "warehouse_id": 1,
  "notes": "Partial delivery"
}
```
`po_id` is optional (standalone GRN without PO). If provided, the linked PO's invoice must be in `DRAFT` or `CANCELLED` status — see [GRN Guard](#6-grn-guard).

#### Get GRN
```
GET /api/v1/procurement/goods-receipts/:id
```

#### Confirm GRN
```
POST /api/v1/procurement/goods-receipts/:id/confirm
```
Transitions `DRAFT → CONFIRMED`. Requires at least one item.  
- Updates `stock_ledger` in the same DB transaction.
- Applies `transfer_ratio` when writing stock quantities/costs.
- Appends lines to the PO's DRAFT invoice (best-effort).

---

### 10.6 GRN Items

#### List GRN Items
```
GET /api/v1/procurement/goods-receipts/:id/items
```

#### Add GRN Item *(DRAFT GRN only)*
```
POST /api/v1/procurement/goods-receipts/:id/items
```
```json
{
  "po_line_id": 7,
  "product_id": 10,
  "variant_id": null,
  "quantity": 12,
  "uom_id": 3,
  "location_id": 2,
  "unit_cost": 300.00,
  "transfer_ratio": 1.0,
  "notes": ""
}
```
`po_line_id` links this receipt line to the originating PO line (used for 3-way match pricing on invoice).

#### Get GRN Item
```
GET /api/v1/procurement/goods-receipts/:id/items/:itemId
```

#### Update GRN Item *(DRAFT GRN only)*
```
PUT /api/v1/procurement/goods-receipts/:id/items/:itemId
```
Same body as Add (without `po_line_id`).

---

### 10.7 Purchase Invoices

#### List Invoices
```
GET /api/v1/procurement/purchase-invoices?status=DRAFT&supplier_id=3
```
Query params: `status`, `supplier_id` (both optional)

#### Create Invoice *(manual)*
```
POST /api/v1/procurement/purchase-invoices
```
```json
{
  "po_id": 5,
  "supplier_invoice_no": "SUP-INV-2026-100",
  "supplier_invoice_date": "2026-05-22",
  "invoice_date": "2026-05-23",
  "due_date": "2026-06-22",
  "currency_id": 1,
  "exchange_rate": 1.0,
  "payment_term_id": 2,
  "notes": ""
}
```
`po_id` is required. `supplier_id`, `exchange_rate`, and `payment_term_id` are inherited from the PO if omitted.

#### Get Invoice
```
GET /api/v1/procurement/purchase-invoices/:id
```

#### Post Invoice
```
POST /api/v1/procurement/purchase-invoices/:id/post
```
Transitions `DRAFT → POSTED`. Requires at least one line. Emits `procurement.invoice.posted` event.

#### Record Payment *(POSTED or PARTIAL only)*
```
POST /api/v1/procurement/purchase-invoices/:id/pay
```
```json
{
  "amount": 1800.00,
  "payment_date": "2026-05-23",
  "payment_ref": "CHQ-001",
  "notes": "First instalment"
}
```
Emits `procurement.payment.recorded` event. See [Partial Payment](#7-partial-payment).

#### Cancel Invoice
```
POST /api/v1/procurement/purchase-invoices/:id/cancel
```
Not allowed if status is `PAID` or already `CANCELLED`.

---

### 10.8 Invoice Lines

#### Add Invoice Line *(DRAFT invoice only)*
```
POST /api/v1/procurement/purchase-invoices/:id/lines
```
```json
{
  "grn_line_id": 4,
  "product_id": 10,
  "variant_id": null,
  "description": "Raw cotton bale",
  "quantity": 12,
  "uom_id": 3,
  "unit_price": 900.00,
  "discount_pct": 0,
  "tax_code_id": 1,
  "supplier_invoice_no": "SUP-INV-2026-100",
  "notes": ""
}
```
`grn_line_id` is optional but recommended for 3-way matching traceability.  
`supplier_invoice_no` at line level supports scenarios where one purchase invoice consolidates lines from multiple supplier invoices.

#### Delete Invoice Line *(DRAFT invoice only)*
```
DELETE /api/v1/procurement/purchase-invoices/:id/lines/:lineId
```

---

## 11. Data Models

### PurchaseRequest

| Field | Type | Notes |
|-------|------|-------|
| `id` | uint | PK |
| `tenant_id` | uint | Multi-tenant scope |
| `code` | string | Auto-generated, e.g. `PR-2026-0001` |
| `request_date` | date | Defaults to today |
| `required_date` | date? | Optional delivery need-by date |
| `requested_by` | uint? | FK → users |
| `department_id` | uint? | FK → departments |
| `status` | string | `DRAFT \| PENDING_APPROVAL \| APPROVED \| REJECTED \| CANCELLED` |
| `notes` | string | |
| `approved_by` | uint? | FK → users |
| `approved_at` | timestamp? | |
| `created_by` | uint? | |
| `lines` | PRLine[] | Eager loaded on GET |

### PRLine

| Field | Type | Notes |
|-------|------|-------|
| `product_id` | uint | Required |
| `variant_id` | uint? | |
| `quantity` | float | Required, > 0 |
| `uom_id` | uint | Purchase UOM |
| `estimated_price` | float | Budgetary price |
| `transfer_ratio` | float | Purchase UOM → stock UOM, default 1 |
| `currency_id` | uint? | Price currency if different from base |

### PurchaseOrder

| Field | Type | Notes |
|-------|------|-------|
| `code` | string | Auto-generated, e.g. `PO-2026-0001` |
| `supplier_id` | uint | Required |
| `pr_id` | uint? | Optional link to originating PR |
| `currency_id` | uint | Required |
| `exchange_rate` | float | Default 1 |
| `warehouse_id` | uint | Receiving warehouse |
| `subtotal` | float | Computed from lines |
| `tax_amount` | float | Computed from lines |
| `discount_amount` | float | Header-level discount |
| `total_amount` | float | `subtotal + tax - discount` |
| `status` | string | `DRAFT \| CONFIRMED \| PARTIAL \| RECEIVED \| CANCELLED` |

### POLine

| Field | Type | Notes |
|-------|------|-------|
| `unit_price` | float | Required |
| `discount_pct` | float | 0–100 |
| `tax_code_id` | uint? | |
| `tax_amount` | float | Computed |
| `line_total` | float | `qty × price × (1 - disc/100)` |
| `transfer_ratio` | float | Default 1 |
| `received_qty` | float | Updated on GRN confirm |
| `billed_qty` | float | Updated on invoice line add |

### GoodsReceipt

| Field | Type | Notes |
|-------|------|-------|
| `code` | string | Auto-generated, e.g. `GRN-2026-0001` |
| `po_id` | uint? | Optional PO reference |
| `supplier_id` | uint | Required |
| `warehouse_id` | uint | Required |
| `status` | string | `DRAFT \| CONFIRMED \| CANCELLED` |

### GRNLine

| Field | Type | Notes |
|-------|------|-------|
| `po_line_id` | uint? | Links to PO line for 3-way match |
| `quantity` | float | Purchase UOM quantity |
| `unit_cost` | float | Cost in purchase UOM |
| `total_cost` | float | `quantity × unit_cost` |
| `transfer_ratio` | float | Applied to stock ledger on confirm |
| `location_id` | uint? | Bin/rack within warehouse |

### PurchaseInvoice

| Field | Type | Notes |
|-------|------|-------|
| `code` | string | Auto-generated, e.g. `PINV-2026-0001` |
| `po_id` | uint | Required — one invoice per PO |
| `supplier_id` | uint | Inherited from PO |
| `supplier_invoice_no` | string | Supplier's own invoice reference |
| `supplier_invoice_date` | date? | Date on supplier's document |
| `currency_id` | uint | |
| `exchange_rate` | float | |
| `subtotal / tax_amount / discount_amount / total_amount` | float | Computed from lines |
| `paid_amount` | float | Accumulates with each payment |
| `status` | string | `DRAFT \| POSTED \| PARTIAL \| PAID \| CANCELLED` |
| `posted_by / posted_at` | uint? / timestamp? | Set on post |

### InvoiceLine

| Field | Type | Notes |
|-------|------|-------|
| `grn_line_id` | uint? | Traceability link to GRN line |
| `unit_price` | float | Sourced from PO line on auto-append |
| `discount_pct` | float | 0–100 |
| `tax_code_id` | uint? | |
| `line_total` | float | `qty × price × (1 - disc/100)` |
| `supplier_invoice_no` | string | Line-level supplier invoice reference |
