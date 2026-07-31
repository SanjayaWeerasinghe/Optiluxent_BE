# Kadahapola ERP — Full System Setup

Zero-to-running guide for a fresh machine. Covers both the BE and FE
repos, all databases, seeded master data, and an admin login.

Assumes both repos are cloned side by side under a single parent
folder (any name will do):

```
D:\Kadahapola\               (or ~/Kadahapola on macOS/Linux)
├── Kadahapola BE\           (this repo — infra, migrations, seeder)
└── Kadahapola FE\           (React + Vite app)
```

The folder names above are what the E2E specs assume in relative paths.
If you clone elsewhere, keep the two sibling folders together.

---

## 0. Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| Docker Desktop | 4.x+ | Runs Postgres 16, Redis 7, ClickHouse, plus BE (Air hot-reload) and FE (Vite HMR) containers |
| Git | any recent | Clone the two repos |
| Node.js | 20+ | Only needed if you'll run Playwright specs from the host (recommended). Otherwise Docker handles it. |
| A GitHub account with access to `SanjayaWeerasinghe/Optiluxent_BE` and `..._FE` | — | For cloning + pushing |

Windows-specific: Docker Desktop should use the WSL2 backend. Enable it
if the installer didn't already.

---

## 1. Clone both repos

```bash
mkdir Kadahapola && cd Kadahapola
git clone https://github.com/SanjayaWeerasinghe/Optiluxent_BE.git "Kadahapola BE"
git clone https://github.com/SanjayaWeerasinghe/Optiluxent_FE.git "Kadahapola FE"

cd "Kadahapola BE" && git checkout Kadahapola && cd ..
cd "Kadahapola FE" && git checkout Kadahapola && cd ..
```

Folder names with spaces are intentional — the sibling scripts +
volume-mount paths reference them literally.

---

## 2. Configure environment

Both repos ship `.env.example` files.

**BE** — `Kadahapola BE/.env.example` → copy to `.env` and adjust if you
want non-default secrets. The defaults work for local dev; the only
values worth changing early are:

```
SEED_ADMIN_EMAIL=admin@kadahapola.com   # first login
SEED_ADMIN_PASSWORD=Admin@12345         # change this if this box is shared
JWT_SECRET=<any long random string>     # required for tokens
```

**FE** — `Kadahapola FE/.env.example` → copy to `.env`:

```
VITE_API_URL=http://localhost:3000
```

`docker-compose.dev.yml` on the BE side already injects the DB/redis/CH
hostnames into the API container's environment, so you don't need to
duplicate those in `.env` unless you're running the API on the host
instead of in Docker.

---

## 3. Bring up the BE stack

From `Kadahapola BE`:

```bash
docker compose -f docker-compose.dev.yml up -d
```

This starts five containers:

| Container | Port | What it does |
|---|---|---|
| `erp-postgres-dev` | 5433 → 5432 | Primary OLTP database (Postgres 16) |
| `erp-redis-dev` | 6379 | Token blacklist + rate limits |
| `erp-clickhouse-dev` | 8123, 9000 | High-volume audit + stock-ledger append log |
| `erp-api-dev` | 3000 | Go API — Air watches `.go` files and rebuilds on save |
| `erp-web-dev` | 5173 | (starts from the FE compose — see §5 or bring it up here too if BE compose pulls it in) |

The API container's entrypoint runs `./erp-migrate -command up` before
starting the server, so all migrations (currently through 000047) land
on the empty database automatically the first time. Watch for it in
the logs:

```bash
docker logs -f erp-api-dev
```

Wait for `HTTP server configured` before proceeding. Verify:

```bash
curl -s http://localhost:3000/health
# {"success":true,"message":"Service is healthy","data":{"status":"ok",...}}
```

If Postgres reports "database does not exist" — the volume from an
earlier attempt is fine; check the container name in the error and
either recreate the volume or verify `POSTGRES_DB=erp_db` is set.

---

## 4. Seed the super admin

Migrations create empty tables. We only seed **the tenant, the RBAC
scaffolding (permissions + roles), and the super-admin user** — the
company sets up their own master data (currencies, warehouses,
products, parties, chart of accounts, document types) through the FE
once they log in.

```bash
docker exec -e SEED_ADMIN_PASSWORD="YourStrongPassword" erp-api-dev sh -c 'cd /app && go run ./cmd/seed prod'
```

`prod` mode runs:
- `SeedDefaultTenant` — creates the tenant row the whole app is scoped to.
- `SeedPermissions` — inserts every permission the modules check against.
- `SeedRoles` — creates the built-in roles (super admin, admin, staff, …).
- `SeedAdminUser` — creates one user with the super-admin role, using
  the email + password from `SEED_ADMIN_EMAIL` (default
  `admin@kadahapola.com`) and `SEED_ADMIN_PASSWORD`.

If `SEED_ADMIN_PASSWORD` is unset, the seeder falls back to
`Admin@12345` and prints a warning. Rotate as soon as the company's
IT lead has real credentials.

> `dev` and `all` modes also exist and layer on sample master data +
> parties + products — useful for local development. **Do not use them
> when handing the system over to a real tenant.**

### What the company sets up next (via the FE, no CLI needed)

The super admin logs in and populates, in this order:

1. **Master Data → Organization** — company profile, warehouses.
2. **Master Data → Financial** — currencies, exchange rates, payment
   terms, tax codes, banks, company bank accounts, chart of accounts.
3. **Master Data → Contacts** — suppliers and customers (with
   `credit_type` CASH or CREDIT and `credit_limit`).
4. **Master Data → Products / Materials / Categories** — the SKUs the
   business actually sells and buys.
5. **Master Data → Document Types** — customise doc-type behaviour
   (e.g. mark a GRN type as PRODUCTION_OUTPUT to trigger auto-QC).
6. **Master Data → HR** — job positions, then employees.
7. **Finance → Settings** — pick the default AR / AP / Cash / Sales /
   Purchase Expense accounts. Once these are set, every SI/PI post
   will start writing balanced journal entries automatically.

Everything is empty until they populate it — nothing pretends to be
"pre-configured" with fake data.

---

## 5. Bring up the FE

From `Kadahapola FE`:

```bash
docker compose -f docker-compose.dev.yml up -d
```

Or, if you prefer host-side Node (faster HMR):

```bash
npm install
npm run dev
```

Vite dev server binds to `http://localhost:5173`. Open it, log in with
the admin credentials from step 4a, and you should land on the
Dashboard.

**Smoke checklist** — every module should render:

- **Dashboard** — cards + charts (empty until you seed transactional data)
- **Procurement** — PR / PO / GRN / PI sections
- **Inventory** — MR / GT / GI / SA / QC / Stock Overview / **Allocations** (new)
- **Sales** — SQ / SO / DO / SI
- **Manufacturing** — Plans / Production Orders / Post-Costing
- **Finance** — Dashboard / Receivables / Payables / Payments Log / JEs / Settings
- **Human Resources** — Employees / Attendance
- **Master Data** — Organization / Financial / Contacts / Products / Manufacturing / HR / Inventory / …

---

## 6. Verify end-to-end

> The E2E specs assume the sample master data that the `dev` seeder
> creates (currencies `LKR/USD/GBP`, UOMs `KG/PKT`, warehouses
> `WH001/WH002`, sample products, suppliers, customers). Under a
> clean `prod` seed the tenant has none of that — the specs will fail
> until the company (or you, on a scratch machine) adds it. Two paths:
>
> - **Handover to a real tenant** — skip §6 entirely. Verify the app
>   by logging in as the super admin and walking the FE.
> - **Local dev machine** — swap `prod` for `dev` in §4 (or run
>   `docker exec ... go run ./cmd/seed dev` again to layer the sample
>   data on top of the existing tenant), then run the specs below.

**Quick smoke (30 sec)** — runs one flow via API, no browser needed:

```bash
cd "Kadahapola FE"
npx playwright test e2e/validations/qty-caps.spec.ts --reporter=list
```

Expected: **7 passed**. This proves the qty-cap validators, stock
allocations, and refining segregation all fire.

**Full chain (5–10 min)** — walks the whole business flow through the
UI, in order. Reset local state first so the chain starts clean:

```bash
cd "Kadahapola FE"
rm -rf e2e/.state

# Procurement → Finance AP
npx playwright test e2e/finance/00-seed-settings.spec.ts
npx playwright test e2e/procurement/purchase-requests
npx playwright test e2e/procurement/purchase-orders/{01,02,03}-*.spec.ts
npx playwright test e2e/procurement/goods-receipts/{01,02,03}-*.spec.ts
npx playwright test e2e/procurement/purchase-invoices/{01,02}-*.spec.ts   # SKIP 03
npx playwright test e2e/finance/payables

# Sales → Finance AR
npx playwright test e2e/sales/sales-quotations
npx playwright test e2e/sales/sales-orders/from-sq
npx playwright test e2e/sales/delivery-orders
npx playwright test e2e/sales/sales-invoices
npx playwright test e2e/finance/receivables

# Inventory + Manufacturing + HR
npx playwright test e2e/inventory/material-requests
npx playwright test e2e/inventory/goods-transfers
npx playwright test e2e/inventory/goods-issues
npx playwright test e2e/inventory/stock-adjustments
npx playwright test e2e/inventory/quality-checks
npx playwright test e2e/manufacturing/production-plans
npx playwright test e2e/manufacturing/production-orders-normal
npx playwright test e2e/manufacturing/production-orders-refining
npx playwright test e2e/hr/employees
```

Each folder is a sequential chain — files pass state to one another via
`e2e/.state/*.json`. If a step fails, subsequent files in that chain
skip cleanly rather than false-failing.

---

## 7. Daily workflow

**Start/stop**:

```bash
docker compose -f docker-compose.dev.yml up -d      # start
docker compose -f docker-compose.dev.yml down       # stop, keep volumes
docker compose -f docker-compose.dev.yml down -v    # stop + wipe DB (nuclear)
```

**Hot reload** — Air (BE) and Vite (FE) both auto-reload on file save.
If Air appears frozen after a compile error, restart the API container:

```bash
docker restart erp-api-dev
```

**Adding a migration**:

```bash
# create the two files
touch internal/infrastructure/database/migrations/000048_your_change.{up,down}.sql

# apply
docker exec erp-api-dev sh -c 'cd /app && go run ./cmd/migrate up'
```

**Reset the DB from scratch** — wipe + re-migrate + re-seed:

```bash
docker compose -f docker-compose.dev.yml down -v
docker compose -f docker-compose.dev.yml up -d
# wait ~15s for the entrypoint to run migrations
docker exec -e SEED_ADMIN_PASSWORD="YourPwd" erp-api-dev sh -c 'cd /app && go run ./cmd/seed dev'
```

**Running one E2E spec**:

```bash
cd "Kadahapola FE"
npx playwright test e2e/validations/qty-caps.spec.ts --reporter=list
# add --headed to watch it drive the browser
```

---

## 8. Troubleshooting

| Symptom | Fix |
|---|---|
| `git push` returns `403 Permission denied to <other-user>` | Windows Credential Manager cached the wrong login. `printf "protocol=https\nhost=github.com\n\n" \| git credential-manager erase`, then re-push — a browser popup will ask for the correct GitHub login. |
| `Cannot GET /health` on port 3000 | API container hasn't finished booting. `docker logs -f erp-api-dev` — wait for `HTTP server configured`. |
| Migration says `no such table: <x>` on step 4 | Migrations haven't run. Manual trigger: `docker exec erp-api-dev sh -c 'cd /app && go run ./cmd/migrate up'`. |
| API 500 on `/api/v1/finance/*` | Finance settings not seeded. Re-run `e2e/finance/00-seed-settings.spec.ts`. |
| `insufficient stock` on E2E chains for products that "should" have stock | Prior test runs consumed the tenant's baseline. Either reset the DB (step 7 "Reset") or seed extra stock via a Stock Adjustment through the UI. |
| Playwright says "Cannot find module @playwright/test" | You ran it from the wrong dir or inside the wrong container. Run from `Kadahapola FE` on the host; the FE package.json owns Playwright. |
| Vite shows blank page after big FE changes | Vite HMR staleness. Restart the FE container: `docker restart erp-web-dev`. |
| PDF uploads land but aren't served | `/uploads` is mounted at `/app/uploads` inside the API container. If files exist but `GET /uploads/hr/emp-<id>/cv.pdf` 404s, the Fiber static handler didn't bind — check `internal/infrastructure/http/server/routes.go` for the `s.app.Static("/uploads", ...)` call. |

---

## 9. Reference — what's inside each repo

**Kadahapola BE** (Go, Fiber v2, GORM, ClickHouse client)

```
cmd/
  api/               — main server binary + all cross-module wiring
  migrate/           — CLI: go run ./cmd/migrate up|down|status
  seed/              — CLI: go run ./cmd/seed dev|prod|all
internal/
  modules/
    masterdata/      — currencies, UOMs, products, parties, doc types, HR employees, banks, CoA, tax
    procurement/     — PR / PO / GRN / PI + validators.go
    inventory/       — MR / GT / GI / SA / QC + allocation_*.go
    sales/           — SQ / SO / DO / SI
    manufacturing/   — Plans / MOs (normal + refining)
    finance/         — payments, journal entries, aging, settings
    hr/              — family, emergency, attendance, salary history, CV upload
  infrastructure/
    database/migrations/  — SQL migrations 000001 … 000047
scripts/             — bulk seed helpers (bash + curl)
```

**Kadahapola FE** (React 19, Vite, Tailwind, Playwright)

```
src/
  pages/
    procurement/ inventory/ sales/ manufacturing/ finance/ hr/ master-data/
  components/layout/  — sidebar with module groups + leaves
  components/ui/      — Table, Modal, Tabs, Badge, StatusBadge, Icon, Input, …
  lib/                — api.ts (auth + fetch), lookups (id → label cache)
  router/             — route table
e2e/
  fixtures/           — auth, api, state, selectors, reporter
  procurement/ inventory/ sales/ manufacturing/ finance/ hr/ validations/
    *.spec.ts         — sequential numbered chains + regressions
```

---

## 10. When you're done

Push your work on the `Kadahapola` branch of both repos:

```bash
cd "Kadahapola BE" && git push origin Kadahapola
cd "Kadahapola FE" && git push origin Kadahapola
```

If push errors with 403, jump to §8's Credential Manager fix.
