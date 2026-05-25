# Master Data Module — Implementation Plan

**Phase**: Phase 1 (Pre-Business Modules)
**Depends On**: Batch 5 (Production Readiness) ✅ must be complete
**Developer**: Sanjaya Weerasinghe
**Status**: ⏳ Not Started
**Business Context**: Manufacturing + Sales & Distribution, Multi-currency, Multi-warehouse

---

## Overview

Master Data is the shared reference layer that every business module reads from. It is implemented as a single `masterdata` module (registered with the Module Registry from Batch 4), containing 7 submodules. No business logic lives here — only CRUD, validation, and reference data management.

All other modules (Sales, Inventory, Finance, Production, Procurement, HR) depend on Master Data. It must be fully seeded before any transactional module can operate.

**Multi-currency rule**: Every entity that carries a monetary value stores:
- `currency_id` — foreign key to `currencies`
- `exchange_rate` — decimal snapshot at time of creation/update
- All amounts stored in both transaction currency and base currency (LKR)

---

## Module Structure

```
internal/modules/masterdata/
├── module.go                              ← Module interface impl; registers all submodules
├── migrations/                            ← All masterdata SQL migration files
│   ├── md_001_organization.up.sql
│   ├── md_001_organization.down.sql
│   ├── md_002_financial.up.sql
│   ├── md_002_financial.down.sql
│   ├── md_003_contacts.up.sql
│   ├── md_003_contacts.down.sql
│   ├── md_004_products.up.sql
│   ├── md_004_products.down.sql
│   ├── md_005_inventory.up.sql
│   ├── md_005_inventory.down.sql
│   ├── md_006_manufacturing.up.sql
│   ├── md_006_manufacturing.down.sql
│   ├── md_007_hr.up.sql
│   └── md_007_hr.down.sql
├── organization/
│   ├── entity.go
│   ├── repository.go
│   ├── service.go
│   ├── handler.go
│   ├── dto.go
│   └── routes.go
├── financial/
│   ├── entity.go
│   ├── repository.go
│   ├── service.go
│   ├── handler.go
│   ├── dto.go
│   └── routes.go
├── contacts/
│   ├── entity.go
│   ├── repository.go
│   ├── service.go
│   ├── handler.go
│   ├── dto.go
│   └── routes.go
├── products/
│   ├── entity.go
│   ├── repository.go
│   ├── service.go
│   ├── handler.go
│   ├── dto.go
│   └── routes.go
├── inventory/
│   ├── entity.go
│   ├── repository.go
│   ├── service.go
│   ├── handler.go
│   ├── dto.go
│   └── routes.go
├── manufacturing/
│   ├── entity.go
│   ├── repository.go
│   ├── service.go
│   ├── handler.go
│   ├── dto.go
│   └── routes.go
└── hr/
    ├── entity.go
    ├── repository.go
    ├── service.go
    ├── handler.go
    ├── dto.go
    └── routes.go
```

---

## module.go — Top-Level Module Registration

```go
// internal/modules/masterdata/module.go

type MasterDataModule struct {
    db     *gorm.DB
    cache  cache.Cache
    logger *zap.Logger
    bus    events.EventBus
}

func (m *MasterDataModule) Name() string         { return "masterdata" }
func (m *MasterDataModule) Dependencies() []string { return []string{} } // no deps; foundational

func (m *MasterDataModule) Initialize(deps modules.Dependencies) error     { ... }
func (m *MasterDataModule) RegisterRoutes(router fiber.Router)              { ... } // mounts all 7 submodule routes
func (m *MasterDataModule) RegisterEvents(bus events.EventBus)              { ... } // publishes masterdata.* events
func (m *MasterDataModule) Migrate(db *gorm.DB) error                       { ... } // runs md_001 → md_007
func (m *MasterDataModule) Shutdown(ctx context.Context) error               { ... }
```

---

## Submodule 1: Organization

**Route Prefix**: `/api/v1/masterdata/organization`

### Entities (`organization/entity.go`)

```go
type Company struct {
    ID              uint      `gorm:"primarykey"`
    Name            string    `gorm:"not null"`
    LegalName       string
    TaxRegNumber    string
    Logo            string    // URL/path
    Email           string
    Phone           string
    Website         string
    AddressLine1    string
    AddressLine2    string
    City            string
    StateID         uint
    CountryID       uint
    PostalCode      string
    BaseCurrencyID  uint      `gorm:"not null"`
    FiscalYearStart int       // month number (1-12)
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type Department struct {
    ID               uint       `gorm:"primarykey"`
    TenantID         uint       `gorm:"not null;index"`
    Code             string     `gorm:"not null"`
    Name             string     `gorm:"not null"`
    ParentID         *uint      // self-referential for hierarchy
    ManagerEmployeeID *uint
    IsActive         bool       `gorm:"default:true"`
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        gorm.DeletedAt `gorm:"index"`
}

type FiscalYear struct {
    ID        uint      `gorm:"primarykey"`
    TenantID  uint      `gorm:"not null;index"`
    Name      string    `gorm:"not null"`         // e.g. "FY 2025/2026"
    StartDate time.Time `gorm:"not null"`
    EndDate   time.Time `gorm:"not null"`
    IsClosed  bool      `gorm:"default:false"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type AccountingPeriod struct {
    ID           uint      `gorm:"primarykey"`
    TenantID     uint      `gorm:"not null;index"`
    FiscalYearID uint      `gorm:"not null"`
    PeriodNumber int       `gorm:"not null"`      // 1-12
    Name         string    `gorm:"not null"`      // e.g. "April 2025"
    StartDate    time.Time `gorm:"not null"`
    EndDate      time.Time `gorm:"not null"`
    IsClosed     bool      `gorm:"default:false"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type DocumentSequence struct {
    ID          uint   `gorm:"primarykey"`
    TenantID    uint   `gorm:"not null;index"`
    DocumentType string `gorm:"not null"`      // SALES_ORDER, PURCHASE_ORDER, INVOICE, etc.
    Prefix      string `gorm:"not null"`       // SO, PO, INV
    NextNumber  int    `gorm:"default:1"`
    Padding     int    `gorm:"default:5"`      // SO-00001
    Suffix      string                         // optional year suffix: SO-2025-00001
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Country struct {
    ID        uint   `gorm:"primarykey"`
    Code      string `gorm:"uniqueIndex;not null"` // ISO 3166-1 alpha-2
    Name      string `gorm:"not null"`
    PhoneCode string
    IsActive  bool   `gorm:"default:true"`
}

type State struct {
    ID        uint   `gorm:"primarykey"`
    CountryID uint   `gorm:"not null;index"`
    Code      string `gorm:"not null"`
    Name      string `gorm:"not null"`
}
```

### API Endpoints

```
GET    /api/v1/masterdata/organization/company              — Get company profile
PUT    /api/v1/masterdata/organization/company              — Update company profile

GET    /api/v1/masterdata/organization/departments          — List departments (tree)
POST   /api/v1/masterdata/organization/departments          — Create department
GET    /api/v1/masterdata/organization/departments/:id      — Get department
PUT    /api/v1/masterdata/organization/departments/:id      — Update department
DELETE /api/v1/masterdata/organization/departments/:id      — Deactivate department

GET    /api/v1/masterdata/organization/fiscal-years         — List fiscal years
POST   /api/v1/masterdata/organization/fiscal-years         — Create fiscal year
PUT    /api/v1/masterdata/organization/fiscal-years/:id/close — Close fiscal year

GET    /api/v1/masterdata/organization/accounting-periods   — List periods (filter by fiscal year)
POST   /api/v1/masterdata/organization/accounting-periods/:id/close — Close period

GET    /api/v1/masterdata/organization/document-sequences   — List sequences
POST   /api/v1/masterdata/organization/document-sequences   — Create sequence
PUT    /api/v1/masterdata/organization/document-sequences/:id — Update sequence

GET    /api/v1/masterdata/organization/countries            — List countries
GET    /api/v1/masterdata/organization/countries/:id/states — List states for country
```

---

## Submodule 2: Financial

**Route Prefix**: `/api/v1/masterdata/financial`

### Entities (`financial/entity.go`)

```go
type Currency struct {
    ID        uint   `gorm:"primarykey"`
    Code      string `gorm:"uniqueIndex;not null"` // ISO 4217: LKR, USD, EUR
    Name      string `gorm:"not null"`
    Symbol    string `gorm:"not null"`             // Rs, $, €
    IsBase    bool   `gorm:"default:false"`        // only one base currency
    IsActive  bool   `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type ExchangeRate struct {
    ID             uint      `gorm:"primarykey"`
    TenantID       uint      `gorm:"not null;index"`
    FromCurrencyID uint      `gorm:"not null"`
    ToCurrencyID   uint      `gorm:"not null"`      // always base currency
    Rate           float64   `gorm:"not null"`      // 1 FROM = Rate TO
    EffectiveDate  time.Time `gorm:"not null;index"`
    Source         string    // manual, bank_feed, api
    CreatedAt      time.Time
    // Note: rates are never edited; insert a new record to correct
}

// ChartOfAccount uses a nested set or adjacency list for hierarchy
type ChartOfAccount struct {
    ID           uint    `gorm:"primarykey"`
    TenantID     uint    `gorm:"not null;index"`
    Code         string  `gorm:"not null"`
    Name         string  `gorm:"not null"`
    AccountType  string  `gorm:"not null"` // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
    ParentID     *uint
    CurrencyID   uint    `gorm:"not null"`
    IsControlled bool    `gorm:"default:false"` // controlled by sub-ledger (AR, AP, Bank)
    IsActive     bool    `gorm:"default:true"`
    Level        int     // computed: depth in hierarchy
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type CostCenter struct {
    ID           uint   `gorm:"primarykey"`
    TenantID     uint   `gorm:"not null;index"`
    Code         string `gorm:"not null"`
    Name         string `gorm:"not null"`
    DepartmentID *uint
    IsActive     bool   `gorm:"default:true"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type PaymentTerm struct {
    ID              uint    `gorm:"primarykey"`
    TenantID        uint    `gorm:"not null;index"`
    Code            string  `gorm:"not null"`
    Name            string  `gorm:"not null"`   // Net 30, Immediate, 50% Advance
    DueDays         int     `gorm:"default:0"`
    DiscountDays    int     `gorm:"default:0"`
    DiscountPercent float64 `gorm:"default:0"`
    IsActive        bool    `gorm:"default:true"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type Bank struct {
    ID        uint   `gorm:"primarykey"`
    Name      string `gorm:"not null"`
    BranchName string
    SwiftCode string
    Address   string
    IsActive  bool   `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type CompanyBankAccount struct {
    ID            uint   `gorm:"primarykey"`
    TenantID      uint   `gorm:"not null;index"`
    BankID        uint   `gorm:"not null"`
    AccountNumber string `gorm:"not null"`
    AccountName   string `gorm:"not null"`
    CurrencyID    uint   `gorm:"not null"`
    GLAccountID   uint   `gorm:"not null"` // links to ChartOfAccount
    IsDefault     bool   `gorm:"default:false"`
    IsActive      bool   `gorm:"default:true"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type TaxCode struct {
    ID          uint    `gorm:"primarykey"`
    TenantID    uint    `gorm:"not null;index"`
    Code        string  `gorm:"not null"`
    Name        string  `gorm:"not null"`
    TaxType     string  `gorm:"not null"` // VAT, WHT, SVAT, EXEMPT
    Rate        float64 `gorm:"not null"` // percentage: 18.00
    GLAccountID uint    `gorm:"not null"` // tax payable/receivable account
    IsActive    bool    `gorm:"default:true"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type TaxGroup struct {
    ID        uint   `gorm:"primarykey"`
    TenantID  uint   `gorm:"not null;index"`
    Name      string `gorm:"not null"`
    IsActive  bool   `gorm:"default:true"`
    TaxCodes  []TaxCode `gorm:"many2many:tax_group_codes;"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### API Endpoints

```
GET    /api/v1/masterdata/financial/currencies              — List currencies
POST   /api/v1/masterdata/financial/currencies              — Create currency
PUT    /api/v1/masterdata/financial/currencies/:id          — Update currency
PUT    /api/v1/masterdata/financial/currencies/:id/set-base — Set as base currency

GET    /api/v1/masterdata/financial/exchange-rates          — List rates (filter by currency, date range)
POST   /api/v1/masterdata/financial/exchange-rates          — Add exchange rate
GET    /api/v1/masterdata/financial/exchange-rates/latest   — Get latest rate per currency pair

GET    /api/v1/masterdata/financial/chart-of-accounts       — List CoA (tree or flat with ?flat=true)
POST   /api/v1/masterdata/financial/chart-of-accounts       — Create account
GET    /api/v1/masterdata/financial/chart-of-accounts/:id   — Get account
PUT    /api/v1/masterdata/financial/chart-of-accounts/:id   — Update account
DELETE /api/v1/masterdata/financial/chart-of-accounts/:id   — Deactivate account

GET    /api/v1/masterdata/financial/cost-centers            — List cost centers
POST   /api/v1/masterdata/financial/cost-centers            — Create cost center
PUT    /api/v1/masterdata/financial/cost-centers/:id        — Update cost center

GET    /api/v1/masterdata/financial/payment-terms           — List payment terms
POST   /api/v1/masterdata/financial/payment-terms           — Create payment term
PUT    /api/v1/masterdata/financial/payment-terms/:id       — Update payment term

GET    /api/v1/masterdata/financial/banks                   — List banks
POST   /api/v1/masterdata/financial/banks                   — Create bank

GET    /api/v1/masterdata/financial/bank-accounts           — List company bank accounts
POST   /api/v1/masterdata/financial/bank-accounts           — Create bank account
PUT    /api/v1/masterdata/financial/bank-accounts/:id       — Update bank account

GET    /api/v1/masterdata/financial/tax-codes               — List tax codes
POST   /api/v1/masterdata/financial/tax-codes               — Create tax code
PUT    /api/v1/masterdata/financial/tax-codes/:id           — Update tax code

GET    /api/v1/masterdata/financial/tax-groups              — List tax groups
POST   /api/v1/masterdata/financial/tax-groups              — Create tax group
PUT    /api/v1/masterdata/financial/tax-groups/:id          — Update tax group
```

---

## Submodule 3: Contacts

**Route Prefix**: `/api/v1/masterdata/contacts`

### Entities (`contacts/entity.go`)

```go
// Shared base for Customer and Supplier
type Party struct {
    ID            uint   `gorm:"primarykey"`
    TenantID      uint   `gorm:"not null;index"`
    Code          string `gorm:"not null"`           // auto-generated: CUST-0001, SUPP-0001
    Name          string `gorm:"not null"`
    LegalName     string
    PartyType     string `gorm:"not null"`           // CUSTOMER, SUPPLIER, BOTH
    Type          string `gorm:"not null"`           // INDIVIDUAL, COMPANY
    TaxRegNumber  string
    CurrencyID    uint   `gorm:"not null"`           // default transaction currency
    PaymentTermID *uint
    CreditLimit   float64 `gorm:"default:0"`        // customers only
    IsActive      bool    `gorm:"default:true"`
    Notes         string
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     gorm.DeletedAt `gorm:"index"`
}

type ContactPerson struct {
    ID          uint   `gorm:"primarykey"`
    TenantID    uint   `gorm:"not null;index"`
    PartyID     uint   `gorm:"not null;index"`
    Name        string `gorm:"not null"`
    Designation string
    Phone       string
    Mobile      string
    Email       string
    IsPrimary   bool   `gorm:"default:false"`
    IsActive    bool   `gorm:"default:true"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Address struct {
    ID           uint   `gorm:"primarykey"`
    TenantID     uint   `gorm:"not null;index"`
    PartyID      uint   `gorm:"not null;index"`
    AddressType  string `gorm:"not null"` // BILLING, SHIPPING, BOTH
    AddressLine1 string `gorm:"not null"`
    AddressLine2 string
    City         string
    StateID      *uint
    CountryID    uint   `gorm:"not null"`
    PostalCode   string
    IsPrimary    bool   `gorm:"default:false"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Note**: Customer and Supplier are both backed by the `parties` table with `party_type` discriminator. This avoids duplicating contact/address logic. Queries for customers use `WHERE party_type IN ('CUSTOMER', 'BOTH')`, suppliers use `WHERE party_type IN ('SUPPLIER', 'BOTH')`.

### API Endpoints

```
GET    /api/v1/masterdata/contacts/customers                — List customers (paginated, searchable)
POST   /api/v1/masterdata/contacts/customers                — Create customer
GET    /api/v1/masterdata/contacts/customers/:id            — Get customer with contacts + addresses
PUT    /api/v1/masterdata/contacts/customers/:id            — Update customer
DELETE /api/v1/masterdata/contacts/customers/:id            — Soft-delete customer

GET    /api/v1/masterdata/contacts/suppliers                — List suppliers
POST   /api/v1/masterdata/contacts/suppliers                — Create supplier
GET    /api/v1/masterdata/contacts/suppliers/:id            — Get supplier with contacts + addresses
PUT    /api/v1/masterdata/contacts/suppliers/:id            — Update supplier
DELETE /api/v1/masterdata/contacts/suppliers/:id            — Soft-delete supplier

POST   /api/v1/masterdata/contacts/:party_id/contacts       — Add contact person
PUT    /api/v1/masterdata/contacts/:party_id/contacts/:id   — Update contact person
DELETE /api/v1/masterdata/contacts/:party_id/contacts/:id   — Remove contact person

POST   /api/v1/masterdata/contacts/:party_id/addresses      — Add address
PUT    /api/v1/masterdata/contacts/:party_id/addresses/:id  — Update address
DELETE /api/v1/masterdata/contacts/:party_id/addresses/:id  — Remove address
```

---

## Submodule 4: Products

**Route Prefix**: `/api/v1/masterdata/products`

### Entities (`products/entity.go`)

```go
type ProductCategory struct {
    ID        uint   `gorm:"primarykey"`
    TenantID  uint   `gorm:"not null;index"`
    Code      string `gorm:"not null"`
    Name      string `gorm:"not null"`
    ParentID  *uint  // hierarchical categories
    IsActive  bool   `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Product struct {
    ID              uint    `gorm:"primarykey"`
    TenantID        uint    `gorm:"not null;index"`
    Code            string  `gorm:"not null"`           // auto-generated or manual
    Name            string  `gorm:"not null"`
    Description     string
    CategoryID      uint    `gorm:"not null"`
    ProductType     string  `gorm:"not null"`  // RAW_MATERIAL, SEMI_FINISHED, FINISHED_GOOD, CONSUMABLE, SERVICE
    BaseUoMID       uint    `gorm:"not null"`  // base unit of measure
    PurchaseUoMID   *uint   // if different from base
    SalesUoMID      *uint   // if different from base
    TaxGroupID      *uint
    StockAccountID  *uint   // GL account for stock valuation
    CogsAccountID   *uint   // GL account for cost of goods sold
    RevenueAccountID *uint  // GL account for sales revenue
    CostPrice       float64 `gorm:"default:0"`
    SalePrice       float64 `gorm:"default:0"`
    MinStockLevel   float64 `gorm:"default:0"` // reorder point
    MaxStockLevel   float64 `gorm:"default:0"`
    LeadTimeDays    int     `gorm:"default:0"`
    IsTrackedSerial bool    `gorm:"default:false"` // track by serial number
    IsTrackedLot    bool    `gorm:"default:false"` // track by lot/batch
    IsActive        bool    `gorm:"default:true"`
    Barcode         string
    Weight          float64
    WeightUoMID     *uint
    Notes           string
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type UnitOfMeasure struct {
    ID       uint   `gorm:"primarykey"`
    Code     string `gorm:"uniqueIndex;not null"` // PCS, KG, M, L, DOZ
    Name     string `gorm:"not null"`
    UoMType  string `gorm:"not null"` // QUANTITY, WEIGHT, LENGTH, VOLUME, TIME
    IsActive bool   `gorm:"default:true"`
}

type UoMConversion struct {
    ID         uint    `gorm:"primarykey"`
    TenantID   uint    `gorm:"not null;index"`
    FromUoMID  uint    `gorm:"not null"`
    ToUoMID    uint    `gorm:"not null"`
    Factor     float64 `gorm:"not null"` // 1 FROM = Factor TO  (e.g. 1 DOZ = 12 PCS → factor=12)
    ProductID  *uint   // NULL = global conversion; set for product-specific
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type PriceList struct {
    ID         uint      `gorm:"primarykey"`
    TenantID   uint      `gorm:"not null;index"`
    Name       string    `gorm:"not null"`
    CurrencyID uint      `gorm:"not null"`
    ValidFrom  *time.Time
    ValidTo    *time.Time
    IsDefault  bool      `gorm:"default:false"`
    IsActive   bool      `gorm:"default:true"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type PriceListItem struct {
    ID          uint    `gorm:"primarykey"`
    TenantID    uint    `gorm:"not null;index"`
    PriceListID uint    `gorm:"not null;index"`
    ProductID   uint    `gorm:"not null;index"`
    MinQty      float64 `gorm:"default:0"`
    Price       float64 `gorm:"not null"`
    CurrencyID  uint    `gorm:"not null"` // must match PriceList.CurrencyID
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### API Endpoints

```
GET    /api/v1/masterdata/products/categories               — List categories (tree)
POST   /api/v1/masterdata/products/categories               — Create category
PUT    /api/v1/masterdata/products/categories/:id           — Update category
DELETE /api/v1/masterdata/products/categories/:id           — Deactivate

GET    /api/v1/masterdata/products/items                    — List products (filter by type, category, active)
POST   /api/v1/masterdata/products/items                    — Create product
GET    /api/v1/masterdata/products/items/:id                — Get product (full detail)
PUT    /api/v1/masterdata/products/items/:id                — Update product
DELETE /api/v1/masterdata/products/items/:id                — Deactivate product

GET    /api/v1/masterdata/products/uom                      — List units of measure
POST   /api/v1/masterdata/products/uom                      — Create UoM
PUT    /api/v1/masterdata/products/uom/:id                  — Update UoM

GET    /api/v1/masterdata/products/uom-conversions          — List conversions (filter by product)
POST   /api/v1/masterdata/products/uom-conversions          — Create conversion
PUT    /api/v1/masterdata/products/uom-conversions/:id      — Update conversion
DELETE /api/v1/masterdata/products/uom-conversions/:id      — Delete conversion

GET    /api/v1/masterdata/products/price-lists              — List price lists
POST   /api/v1/masterdata/products/price-lists              — Create price list
GET    /api/v1/masterdata/products/price-lists/:id/items    — List items in price list
POST   /api/v1/masterdata/products/price-lists/:id/items    — Add item to price list
PUT    /api/v1/masterdata/products/price-lists/:id/items/:item_id — Update price list item
DELETE /api/v1/masterdata/products/price-lists/:id/items/:item_id — Remove item
```

---

## Submodule 5: Inventory

**Route Prefix**: `/api/v1/masterdata/inventory`

### Entities (`inventory/entity.go`)

```go
type Warehouse struct {
    ID           uint   `gorm:"primarykey"`
    TenantID     uint   `gorm:"not null;index"`
    Code         string `gorm:"not null"`
    Name         string `gorm:"not null"`
    AddressLine1 string
    AddressLine2 string
    City         string
    StateID      *uint
    CountryID    uint
    IsActive     bool  `gorm:"default:true"`
    IsDefault    bool  `gorm:"default:false"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type StorageLocation struct {
    ID           uint   `gorm:"primarykey"`
    TenantID     uint   `gorm:"not null;index"`
    WarehouseID  uint   `gorm:"not null;index"`
    Code         string `gorm:"not null"`
    Name         string `gorm:"not null"`
    LocationType string `gorm:"not null"` // INPUT, OUTPUT, STORAGE, QUALITY, SCRAP
    IsActive     bool   `gorm:"default:true"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type StockAdjustmentReason struct {
    ID             uint   `gorm:"primarykey"`
    TenantID       uint   `gorm:"not null;index"`
    Code           string `gorm:"not null"`
    Name           string `gorm:"not null"`
    AdjustmentType string `gorm:"not null"` // ADD, SUBTRACT
    IsActive       bool   `gorm:"default:true"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### API Endpoints

```
GET    /api/v1/masterdata/inventory/warehouses               — List warehouses
POST   /api/v1/masterdata/inventory/warehouses               — Create warehouse
GET    /api/v1/masterdata/inventory/warehouses/:id           — Get warehouse with locations
PUT    /api/v1/masterdata/inventory/warehouses/:id           — Update warehouse
DELETE /api/v1/masterdata/inventory/warehouses/:id           — Deactivate warehouse

GET    /api/v1/masterdata/inventory/warehouses/:id/locations — List locations in warehouse
POST   /api/v1/masterdata/inventory/warehouses/:id/locations — Create location
PUT    /api/v1/masterdata/inventory/locations/:id            — Update location
DELETE /api/v1/masterdata/inventory/locations/:id            — Deactivate location

GET    /api/v1/masterdata/inventory/adjustment-reasons       — List adjustment reasons
POST   /api/v1/masterdata/inventory/adjustment-reasons       — Create reason
PUT    /api/v1/masterdata/inventory/adjustment-reasons/:id   — Update reason
```

---

## Submodule 6: Manufacturing

**Route Prefix**: `/api/v1/masterdata/manufacturing`

### Entities (`manufacturing/entity.go`)

```go
type WorkCenter struct {
    ID              uint    `gorm:"primarykey"`
    TenantID        uint    `gorm:"not null;index"`
    Code            string  `gorm:"not null"`
    Name            string  `gorm:"not null"`
    DepartmentID    *uint
    WarehouseID     *uint   // physical location
    CapacityPerHour float64 `gorm:"default:1"` // units producible per hour
    CostPerHour     float64 `gorm:"default:0"`
    CurrencyID      uint    `gorm:"not null"`
    IsActive        bool    `gorm:"default:true"`
    Notes           string
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type Operation struct {
    ID                uint   `gorm:"primarykey"`
    TenantID          uint   `gorm:"not null;index"`
    Code              string `gorm:"not null"`
    Name              string `gorm:"not null"` // Cutting, Assembling, Finishing, Painting
    DefaultWorkCenterID *uint
    DefaultDurationMins int   `gorm:"default:0"`
    IsActive          bool   `gorm:"default:true"`
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

type BOM struct {
    ID          uint    `gorm:"primarykey"`
    TenantID    uint    `gorm:"not null;index"`
    ProductID   uint    `gorm:"not null;index"`
    BOMType     string  `gorm:"not null"` // MANUFACTURE, KIT, PHANTOM
    Quantity    float64 `gorm:"not null"` // output quantity this BOM produces
    UoMID       uint    `gorm:"not null"`
    Version     string  `gorm:"default:'1.0'"`
    IsActive    bool    `gorm:"default:true"`
    Notes       string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
    Lines       []BOMLine `gorm:"foreignKey:BOMID"`
}

type BOMLine struct {
    ID          uint    `gorm:"primarykey"`
    TenantID    uint    `gorm:"not null;index"`
    BOMID       uint    `gorm:"not null;index"`
    Sequence    int     `gorm:"not null"`
    ProductID   uint    `gorm:"not null"`   // component / raw material
    Quantity    float64 `gorm:"not null"`
    UoMID       uint    `gorm:"not null"`
    OperationID *uint   // which operation consumes this component
    ScrapPercent float64 `gorm:"default:0"` // expected waste %
    IsOptional  bool    `gorm:"default:false"`
    Notes       string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Routing struct {
    ID        uint   `gorm:"primarykey"`
    TenantID  uint   `gorm:"not null;index"`
    ProductID uint   `gorm:"not null;index"`
    Version   string `gorm:"default:'1.0'"`
    IsActive  bool   `gorm:"default:true"`
    Notes     string
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    Lines     []RoutingLine `gorm:"foreignKey:RoutingID"`
}

type RoutingLine struct {
    ID             uint    `gorm:"primarykey"`
    TenantID       uint    `gorm:"not null;index"`
    RoutingID      uint    `gorm:"not null;index"`
    Sequence       int     `gorm:"not null"`
    OperationID    uint    `gorm:"not null"`
    WorkCenterID   uint    `gorm:"not null"`
    SetupMins      float64 `gorm:"default:0"`
    DurationMins   float64 `gorm:"not null"`
    IsOptional     bool    `gorm:"default:false"`
    Notes          string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### API Endpoints

```
GET    /api/v1/masterdata/manufacturing/work-centers             — List work centers
POST   /api/v1/masterdata/manufacturing/work-centers             — Create work center
GET    /api/v1/masterdata/manufacturing/work-centers/:id         — Get work center
PUT    /api/v1/masterdata/manufacturing/work-centers/:id         — Update work center
DELETE /api/v1/masterdata/manufacturing/work-centers/:id         — Deactivate

GET    /api/v1/masterdata/manufacturing/operations               — List operations
POST   /api/v1/masterdata/manufacturing/operations               — Create operation
PUT    /api/v1/masterdata/manufacturing/operations/:id           — Update operation

GET    /api/v1/masterdata/manufacturing/bom                      — List BOMs (filter by product)
POST   /api/v1/masterdata/manufacturing/bom                      — Create BOM with lines
GET    /api/v1/masterdata/manufacturing/bom/:id                  — Get BOM with all lines
PUT    /api/v1/masterdata/manufacturing/bom/:id                  — Update BOM header
POST   /api/v1/masterdata/manufacturing/bom/:id/lines            — Add BOM line
PUT    /api/v1/masterdata/manufacturing/bom/:id/lines/:line_id   — Update BOM line
DELETE /api/v1/masterdata/manufacturing/bom/:id/lines/:line_id   — Remove BOM line
DELETE /api/v1/masterdata/manufacturing/bom/:id                  — Deactivate BOM

GET    /api/v1/masterdata/manufacturing/routings                 — List routings (filter by product)
POST   /api/v1/masterdata/manufacturing/routings                 — Create routing with lines
GET    /api/v1/masterdata/manufacturing/routings/:id             — Get routing with all lines
PUT    /api/v1/masterdata/manufacturing/routings/:id             — Update routing header
POST   /api/v1/masterdata/manufacturing/routings/:id/lines       — Add routing line
PUT    /api/v1/masterdata/manufacturing/routings/:id/lines/:line_id — Update routing line
DELETE /api/v1/masterdata/manufacturing/routings/:id/lines/:line_id — Remove routing line
DELETE /api/v1/masterdata/manufacturing/routings/:id             — Deactivate routing
```

---

## Submodule 7: HR

**Route Prefix**: `/api/v1/masterdata/hr`

### Entities (`hr/entity.go`)

```go
type JobPosition struct {
    ID           uint   `gorm:"primarykey"`
    TenantID     uint   `gorm:"not null;index"`
    Code         string `gorm:"not null"`
    Title        string `gorm:"not null"`
    DepartmentID *uint
    IsActive     bool   `gorm:"default:true"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type Employee struct {
    ID             uint      `gorm:"primarykey"`
    TenantID       uint      `gorm:"not null;index"`
    Code           string    `gorm:"not null"` // EMP-0001
    FirstName      string    `gorm:"not null"`
    LastName       string    `gorm:"not null"`
    NIC            string    `gorm:"not null"`  // National ID Card
    DateOfBirth    time.Time
    Gender         string
    Phone          string
    Email          string
    Address        string
    DepartmentID   *uint
    JobPositionID  *uint
    HireDate       time.Time `gorm:"not null"`
    TerminationDate *time.Time
    WorkScheduleID *uint
    UserID         *uint     // link to system user (optional)
    IsActive       bool      `gorm:"default:true"`
    Notes          string
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type WorkSchedule struct {
    ID           uint   `gorm:"primarykey"`
    TenantID     uint   `gorm:"not null;index"`
    Name         string `gorm:"not null"`  // Morning Shift, Night Shift, Standard
    ShiftStart   string `gorm:"not null"`  // HH:MM format
    ShiftEnd     string `gorm:"not null"`  // HH:MM format
    WorkingDays  string `gorm:"not null"`  // JSON: ["MON","TUE","WED","THU","FRI"]
    IsActive     bool   `gorm:"default:true"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### API Endpoints

```
GET    /api/v1/masterdata/hr/job-positions                  — List job positions
POST   /api/v1/masterdata/hr/job-positions                  — Create job position
PUT    /api/v1/masterdata/hr/job-positions/:id              — Update job position

GET    /api/v1/masterdata/hr/employees                      — List employees (filter by dept, active)
POST   /api/v1/masterdata/hr/employees                      — Create employee
GET    /api/v1/masterdata/hr/employees/:id                  — Get employee detail
PUT    /api/v1/masterdata/hr/employees/:id                  — Update employee
DELETE /api/v1/masterdata/hr/employees/:id                  — Terminate employee (soft-delete)

GET    /api/v1/masterdata/hr/work-schedules                 — List work schedules
POST   /api/v1/masterdata/hr/work-schedules                 — Create work schedule
PUT    /api/v1/masterdata/hr/work-schedules/:id             — Update work schedule
```

---

## Database Migrations

**Naming convention**: `md_{number}_{submodule}.up/down.sql`
**Location**: `internal/modules/masterdata/migrations/`

### Migration Order (must run in this sequence due to FK dependencies)

```
md_001 — organization
  Creates: countries, states, companies, departments, fiscal_years,
           accounting_periods, document_sequences

md_002 — financial
  Creates: currencies, exchange_rates, chart_of_accounts, cost_centers,
           payment_terms, banks, company_bank_accounts, tax_codes,
           tax_groups, tax_group_codes (junction)
  FKs to: md_001 (companies.base_currency_id → currencies)

md_003 — contacts
  Creates: parties, contact_persons, addresses
  FKs to: md_001 (countries, states), md_002 (currencies, payment_terms)

md_004 — products
  Creates: product_categories, products, units_of_measure, uom_conversions,
           price_lists, price_list_items
  FKs to: md_002 (tax_groups, currencies, chart_of_accounts)

md_005 — inventory
  Creates: warehouses, storage_locations, stock_adjustment_reasons
  FKs to: md_001 (countries, states)

md_006 — manufacturing
  Creates: work_centers, operations, boms, bom_lines, routings, routing_lines
  FKs to: md_004 (products, units_of_measure), md_002 (currencies),
           md_001 (departments), md_005 (warehouses)

md_007 — hr
  Creates: job_positions, employees, work_schedules
  FKs to: md_001 (departments)
```

---

## Implementation Order

### Phase 1: Organization + Financial (Steps 1–4)
1. Write and run `md_001_organization` migration; implement Organization entities, repos, services, handlers, routes
2. Write and run `md_002_financial` migration; implement Financial entities, repos, services, handlers, routes
3. Seed initial data: base currency (LKR), default CoA chart, Sri Lanka as country, default payment terms
4. Test all organization + financial endpoints with Postman

### Phase 2: Contacts + Products (Steps 5–8)
5. Write and run `md_003_contacts` migration; implement Contacts submodule (Party base + Customer/Supplier views)
6. Write and run `md_004_products` migration; implement Products submodule
7. Implement UoM conversions with validation (circular conversion check)
8. Test contacts and products endpoints

### Phase 3: Inventory + Manufacturing (Steps 9–12)
9. Write and run `md_005_inventory` migration; implement Inventory submodule
10. Write and run `md_006_manufacturing` migration; implement Manufacturing submodule
11. Implement BOM line + routing line CRUD (with sequence reordering)
12. Test inventory and manufacturing endpoints; validate BOM references only RAW_MATERIAL or SEMI_FINISHED products

### Phase 4: HR + Module Registration (Steps 13–15)
13. Write and run `md_007_hr` migration; implement HR submodule
14. Implement `module.go`: wire all 7 submodule route registrations, migration runner, event registrations
15. Register `masterdata` module in the module registry; run full integration test

---

## Cross-Cutting Rules

**Every write endpoint must**:
- Validate all FK references exist and belong to the same tenant before saving
- Return the created/updated resource in the response (no empty 201s)
- Emit a domain event: `masterdata.{entity}.created`, `masterdata.{entity}.updated`, `masterdata.{entity}.deleted`

**Soft delete rule**:
- Master data is never hard-deleted
- DELETE endpoints set `is_active = false` (or set `deleted_at` for GORM soft delete)
- Inactive records are excluded from list endpoints unless `?include_inactive=true` is passed

**Multi-currency rule** on monetary fields:
- All `cost_price`, `sale_price`, `cost_per_hour`, `credit_limit` fields are stored in the record's own currency
- The currency is always explicit on the entity (never implicit)
- Business modules are responsible for converting to base currency when journaling

**Code auto-generation**:
- Product, Customer, Supplier, Employee codes use the `document_sequences` table
- Sequence codes: CUST, SUPP, PROD, EMP

---

## Dependencies on Previous Batches

- **Batch 3**: Auth middleware (all masterdata endpoints are protected), tenant context
- **Batch 4**: RBAC enforcer (`masterdata:read`, `masterdata:write` permissions), tenant scope middleware, audit logger (logs all changes), event bus (publishes masterdata events), module registry (masterdata registers itself)
- **Batch 5**: Validation (custom validators on codes, currency codes, date ranges), Swagger annotations on all handlers, seeder (initial currencies, CoA, countries)

---

## Permissions Required

Add to `internal/domain/rbac/permissions.go`:
```
masterdata:read     — view any master data
masterdata:write    — create and update master data
masterdata:delete   — deactivate master data
```

Default role assignments:
- `super_admin`, `admin` — all three
- `manager` — `masterdata:read`, `masterdata:write`
- `user` — `masterdata:read`
- `guest` — `masterdata:read`

---

## Total Entities: 35 tables across 7 submodules

| Submodule | Tables |
|-----------|--------|
| Organization | companies, departments, fiscal_years, accounting_periods, document_sequences, countries, states |
| Financial | currencies, exchange_rates, chart_of_accounts, cost_centers, payment_terms, banks, company_bank_accounts, tax_codes, tax_groups, tax_group_codes |
| Contacts | parties, contact_persons, addresses |
| Products | product_categories, products, units_of_measure, uom_conversions, price_lists, price_list_items |
| Inventory | warehouses, storage_locations, stock_adjustment_reasons |
| Manufacturing | work_centers, operations, boms, bom_lines, routings, routing_lines |
| HR | job_positions, employees, work_schedules |

---

## Success Criteria

- [ ] All 7 submodules registered and routes mounted under `/api/v1/masterdata/`
- [ ] All 35 tables created via migrations; migrations roll back cleanly
- [ ] Multi-currency: every monetary field carries a `currency_id`; exchange rates stored per entry with `effective_date`
- [ ] Tenant isolation: all entities scoped by `tenant_id`; no cross-tenant data leakage
- [ ] Soft delete on all entities: `DELETE` deactivates, not removes; `?include_inactive=true` query param works
- [ ] FK validation: creating a product with an invalid `category_id` returns 400 with clear error
- [ ] BOM validation: BOM lines may only reference RAW_MATERIAL, SEMI_FINISHED, or CONSUMABLE product types
- [ ] Document sequences generate correctly formatted codes (CUST-00001, etc.) with no gaps or collisions under concurrency
- [ ] All endpoints have Swagger annotations
- [ ] All changes appear in audit log
- [ ] Domain events published for every create/update/deactivate
- [ ] Postman collection covers full CRUD for all 7 submodules
- [ ] Integration tests pass for all submodules

---

## Estimated Time: 12–16 hours

---

**Author**: Sanjaya Weerasinghe
**Date**: 2026-05-22
**Next**: Start with Phase 1 — Organization submodule (migration + entities + handlers)
