# Batch 5: Production Readiness - Implementation Plan

**Phase**: Batch 5 - Production Readiness
**Depends On**: Batch 4 (Advanced Infrastructure) ✅ must be complete
**Developer**: Sanjaya Weerasinghe
**Status**: ⏳ Not Started
**Roadmap Sections**: 0.20 – 0.30

---

## Overview

Transform the working backend into a production-hardened system. This batch adds feature flags, enhanced validation, full Swagger documentation, complete Prometheus observability, deep security hardening, database seeding, a proper integration test framework, and a final end-to-end integration pass. After this batch Phase 0 is complete and business module development (Phase 1+) can begin.

---

## Components to Build

### 1. Feature Flags System (10%)
**Location**: `internal/infrastructure/featureflags/`

**Files to Create**:
- `flag.go` — FeatureFlag entity and interface
- `store.go` — FlagStore interface
- `redis_store.go` — Redis-backed store (fast reads, ~1ms latency)
- `db_store.go` — PostgreSQL-backed store (source of truth)
- `evaluator.go` — Flag evaluator (user-level, tenant-level, global)
- `middleware.go` — HTTP middleware to expose flag state to handlers

**Database Migration**:
- `000009_create_feature_flags_table.up.sql`
  ```sql
  CREATE TABLE feature_flags (
      id          BIGSERIAL PRIMARY KEY,
      tenant_id   BIGINT REFERENCES tenants(id),  -- NULL = global
      name        VARCHAR(100) NOT NULL,
      description TEXT,
      enabled     BOOLEAN NOT NULL DEFAULT false,
      rules       JSONB,   -- user_ids[], percentage rollout, etc.
      created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
      UNIQUE(tenant_id, name)
  );
  ```
- `000009_create_feature_flags_table.down.sql`

**FlagStore Interface**:
```go
type FlagStore interface {
    IsEnabled(ctx context.Context, tenantID uint, flag string) (bool, error)
    IsEnabledForUser(ctx context.Context, tenantID, userID uint, flag string) (bool, error)
    Set(ctx context.Context, tenantID uint, flag string, enabled bool) error
    Get(ctx context.Context, tenantID uint, flag string) (*FeatureFlag, error)
    List(ctx context.Context, tenantID uint) ([]*FeatureFlag, error)
    Delete(ctx context.Context, tenantID uint, flag string) error
    Invalidate(ctx context.Context, tenantID uint, flag string) error  // clears Redis cache
}
```

**Evaluation Priority**: User-level override → Tenant-level → Global → default false

**API Endpoints**:
```
GET    /api/v1/feature-flags            — List flags (admin)
POST   /api/v1/feature-flags            — Create flag (admin)
GET    /api/v1/feature-flags/:name      — Get flag (admin)
PUT    /api/v1/feature-flags/:name      — Enable/disable flag (admin)
DELETE /api/v1/feature-flags/:name      — Delete flag (admin)
GET    /api/v1/feature-flags/evaluate   — Evaluate flags for current user (authenticated)
```

---

### 2. Enhanced Validation System (8%)
**Location**: `internal/infrastructure/http/middleware/` and `pkg/validation/`

**Files to Create**:
- `pkg/validation/validator.go` — Singleton validator instance with custom rules registered
- `pkg/validation/custom_rules.go` — Custom validation rules
- `internal/infrastructure/http/middleware/validate.go` — Generic request validation middleware

**Custom Validation Rules to Register**:
```go
// Usage: validate:"slug"
func validateSlug(fl validator.FieldLevel) bool  // lowercase, alphanumeric, hyphens only

// Usage: validate:"strongpassword"
func validateStrongPassword(fl validator.FieldLevel) bool  // min 8, upper+lower+digit+special

// Usage: validate:"phone"
func validatePhone(fl validator.FieldLevel) bool  // E.164 format

// Usage: validate:"future_date"
func validateFutureDate(fl validator.FieldLevel) bool

// Usage: validate:"past_date"
func validatePastDate(fl validator.FieldLevel) bool

// Usage: validate:"currency_code"
func validateCurrencyCode(fl validator.FieldLevel) bool  // ISO 4217

// Usage: validate:"country_code"
func validateCountryCode(fl validator.FieldLevel) bool   // ISO 3166-1 alpha-2
```

**Validation Error Response Format** (consistent across all endpoints):
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      { "field": "email", "message": "must be a valid email address" },
      { "field": "password", "message": "must be at least 8 characters with upper, lower, digit, and special character" }
    ]
  },
  "request_id": "abc-123"
}
```

**Query Parameter Validation**:
- `pkg/validation/query_params.go` — Validate and parse common query params (page, limit, sort, order, filter)
- `pkg/validation/path_params.go` — Parse and validate :id path parameters (positive integer check)

---

### 3. Full Swagger / OpenAPI Documentation (10%)
**Location**: `docs/` and annotations in all handler files

**Files to Create**:
- `docs/swagger.go` — Swag main doc annotations (title, version, description, contact, license, basePath, securityDefinitions)
- `docs/docs.go` — Auto-generated by `swag init` (do not edit manually)

**Setup**:
```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go -o docs/

# Regenerate after any handler annotation change
make docs
```

**Add Swagger Annotations to ALL Handlers**:

Every handler function needs:
```go
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Registration details"
// @Success      201 {object} response.Success{data=dto.UserResponse}
// @Failure      400 {object} response.Error
// @Failure      409 {object} response.Error "Email already exists"
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
```

**Handlers needing annotation** (all of them):
- auth: Register, Login, Refresh, Logout, Me
- user: List, GetByID, Update, Delete, UpdateProfile, ChangePassword, AssignRole, RemoveRole
- role: List, Create, GetByID, Update, Delete, AssignPermissions, RemovePermission
- permission: List
- tenant: List, Create, GetByID, Update
- audit: List
- featureflags: List, Create, Get, Update, Delete, Evaluate
- health: Health, Live, Ready
- system: Modules

**Swagger UI Endpoint**:
```
GET /docs/*  — Swagger UI (dev + staging only; disabled in production via config)
```

**Add to Makefile**:
```makefile
docs:
    swag init -g cmd/api/main.go -o docs/ --parseDependency --parseInternal
```

---

### 4. Complete Monitoring & Observability (10%)
**Location**: `internal/infrastructure/metrics/` and `internal/infrastructure/tracing/`

**Files to Create**:
- `internal/infrastructure/metrics/metrics.go` — All Prometheus metric definitions
- `internal/infrastructure/metrics/collector.go` — Custom collector registration
- `internal/infrastructure/http/middleware/metrics.go` — Per-route HTTP metrics middleware
- `internal/infrastructure/tracing/correlation.go` — Correlation ID propagation helpers

**Prometheus Metrics to Define**:

```go
// HTTP metrics
http_requests_total         CounterVec   {method, route, status_code}
http_request_duration_seconds HistogramVec {method, route}
http_request_size_bytes     HistogramVec {method, route}
http_response_size_bytes    HistogramVec {method, route}
http_active_requests        GaugeVec     {method, route}

// Auth metrics
auth_login_attempts_total   CounterVec   {tenant_id, status}  // status: success|failure
auth_token_validations_total CounterVec  {result}             // valid|expired|blacklisted

// Database metrics
db_connections_open         Gauge
db_connections_idle         Gauge
db_connections_in_use       Gauge
db_query_duration_seconds   HistogramVec {operation}         // select|insert|update|delete

// Cache metrics
cache_operations_total      CounterVec   {operation, result} // operation: get|set|del; result: hit|miss|error
cache_operation_duration_seconds HistogramVec {operation}

// Business metrics
active_tenants_total        Gauge
active_users_total          GaugeVec     {tenant_id}
events_published_total      CounterVec   {stream}
events_consumed_total       CounterVec   {stream, status}    // status: success|retry|dlq
```

**Slow Query Logging**:
- Add GORM callback that logs any query taking > 200ms at WARN level with the SQL and duration
- Location: `internal/infrastructure/database/postgres/slow_query_logger.go`

**Go Runtime Metrics** (auto-exposed by prometheus/client_golang):
- goroutines count, GC pause duration, heap allocations — these are free via `prometheus.MustRegister(collectors.NewGoCollector())`

---

### 5. Security Hardening (12%)
**Location**: `internal/infrastructure/http/middleware/` and `internal/infrastructure/security/`

**Files to Create**:
- `internal/infrastructure/http/middleware/security_headers.go` — Security HTTP headers
- `internal/infrastructure/http/middleware/rate_limit.go` — Enhanced rate limiting (per-IP + per-user + per-route)
- `internal/infrastructure/http/middleware/csrf.go` — CSRF protection (for cookie-based sessions)
- `internal/infrastructure/security/account_lockout.go` — Account lockout after failed logins
- `internal/infrastructure/security/input_sanitizer.go` — Strip dangerous HTML from string inputs
- `internal/infrastructure/security/ip_list.go` — IP allowlist/blocklist (Redis-backed)

**Security Headers Middleware**:
```go
// Applied globally in server setup
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains (production only)
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

**Account Lockout Logic** (stored in Redis):
```
Key: lockout:{email}
Value: { attempts: int, locked_until: time.Time }
TTL: 24 hours

Rules:
- After 5 failed logins: lock for 15 minutes
- After 10 failed logins: lock for 1 hour
- After 20 failed logins: lock for 24 hours
- Successful login: reset counter
- Lockout events: written to audit log
```

**Rate Limiting Tiers**:
```
Auth endpoints (login, register, refresh):
  - Per-IP:   10 requests/minute
  - Per-email: 5 requests/minute

General API (authenticated):
  - Per-user:  300 requests/minute
  - Per-tenant: 1000 requests/minute

Admin endpoints:
  - Per-user:  60 requests/minute

Public endpoints (health, metrics):
  - Per-IP:   60 requests/minute (no auth needed)
```

**Input Sanitization**:
- Strip `<script>`, `<iframe>`, event handlers from any string field
- Applied at repository layer before write, not at HTTP layer
- Location: `internal/infrastructure/security/input_sanitizer.go`

**SQL Injection Prevention**:
- GORM parameterized queries — already enforced by using GORM (never raw string interpolation in queries)
- Add golangci-lint rule to catch raw `db.Raw(fmt.Sprintf(...))` usage

---

### 6. Database Seeding (8%)
**Location**: `internal/infrastructure/database/seeder/`

**Files to Create**:
- `seeder.go` — Seeder interface and runner
- `default_tenant.go` — Creates the default tenant
- `default_permissions.go` — Seeds all permission constants from `domain/rbac/permissions.go`
- `default_roles.go` — Seeds SuperAdmin, Admin, Manager, User, Guest roles with correct permissions
- `default_users.go` — Creates the default super_admin user (credentials from ENV)
- `development_data.go` — Additional dev/test data (only runs in non-production)

**Seeder CLI** (`cmd/seed/main.go`):
```
Commands:
  seed all        — Run all seeders (idempotent)
  seed prod       — Run production seeders only (tenant, permissions, roles, admin user)
  seed dev        — Run prod + dev data seeders
  seed reset      — Drop and re-seed dev data (non-production only)
```

**Default Data Created**:
```
Tenant:      "Kadahapola" (slug: "kadahapola")
Permissions: All resource:action pairs from permissions.go (~30 permissions)
Roles:
  super_admin — all permissions
  admin       — all permissions except tenants:*
  manager     — read/create/update on most resources, no delete/admin
  user        — read own profile, read basic resources
  guest       — read-only on public resources

Super Admin User:
  email:    from ENV SEED_ADMIN_EMAIL    (default: admin@kadahapola.com)
  password: from ENV SEED_ADMIN_PASSWORD (must be set; no hardcoded default)
  role:     super_admin
```

**Idempotency**: All seeders use `INSERT ... ON CONFLICT DO NOTHING` or check existence before inserting.

**Add to Makefile**:
```makefile
seed:
    go run cmd/seed/main.go all

seed-dev:
    go run cmd/seed/main.go dev

seed-reset:
    go run cmd/seed/main.go reset
```

---

### 7. Testing Framework (15%)
**Location**: `tests/`

**Files to Create**:

**Test Infrastructure**:
- `tests/setup/db.go` — Test database setup/teardown (uses a separate `erp_test` database)
- `tests/setup/redis.go` — Test Redis setup (uses DB 15 to avoid collision)
- `tests/setup/server.go` — Test Fiber server setup with all middleware and routes
- `tests/setup/fixtures.go` — Shared fixture data (test users, roles, tokens)
- `tests/setup/helpers.go` — HTTP test helpers (MakeRequest, AssertSuccess, AssertError, etc.)

**Mock Interfaces**:
- `tests/mocks/user_repository.go` — Mock user.Repository
- `tests/mocks/role_repository.go` — Mock role.Repository
- `tests/mocks/event_bus.go` — Mock events.EventBus
- `tests/mocks/audit_logger.go` — Mock audit.AuditLogger
- `tests/mocks/feature_flag_store.go` — Mock featureflags.FlagStore

**Integration Test Files**:
- `tests/integration/auth_test.go` — Auth flow tests (register, login, refresh, logout, me)
- `tests/integration/user_test.go` — User CRUD + role assignment tests
- `tests/integration/role_test.go` — Role + permission assignment tests
- `tests/integration/rbac_test.go` — Permission enforcement tests for all roles
- `tests/integration/tenant_test.go` — Multi-tenant isolation tests
- `tests/integration/audit_test.go` — Audit log generation tests
- `tests/integration/events_test.go` — Event publish/consume/DLQ tests
- `tests/integration/featureflags_test.go` — Feature flag evaluation tests
- `tests/integration/health_test.go` — Health check endpoint tests

**Unit Test Files**:
- `pkg/jwt/jwt_test.go` — Token generation, validation, expiry, blacklist
- `internal/domain/user/entity_test.go` — Password hashing, status checks
- `internal/infrastructure/security/account_lockout_test.go` — Lockout logic

**Test Helpers API**:
```go
// In tests/setup/helpers.go
func MakeRequest(t *testing.T, app *fiber.App, method, path string, body interface{}, headers map[string]string) *http.Response
func AssertSuccess(t *testing.T, resp *http.Response, expectedStatus int)
func AssertError(t *testing.T, resp *http.Response, expectedStatus int, expectedCode string)
func LoginAs(t *testing.T, app *fiber.App, email, password string) string  // returns JWT token
func CreateTestUser(t *testing.T, db *gorm.DB, role string) *user.User
func CreateTestTenant(t *testing.T, db *gorm.DB) *tenant.Tenant
```

**Coverage Target**: 80%+ for `internal/domain/`, `internal/infrastructure/`, `pkg/`

**Add to Makefile**:
```makefile
test:
    go test ./... -v -race

test-integration:
    go test ./tests/integration/... -v -race -tags=integration

test-coverage:
    go test ./... -coverprofile=coverage.out -covermode=atomic
    go tool cover -html=coverage.out -o coverage.html

test-unit:
    go test ./internal/... ./pkg/... -v -race
```

---

### 8. Development Tooling Polish (5%)
**Location**: `Makefile`, `.golangci.yml`, `.air.toml`

**Files to Create/Update**:
- `.golangci.yml` — golangci-lint configuration
- `.air.toml` — Air hot-reload configuration (if not already present)
- `Makefile` — Add missing targets (documented below)

**golangci-lint Rules** (`.golangci.yml`):
```yaml
linters:
  enable:
    - errcheck       # check all errors are handled
    - gosimple
    - govet
    - staticcheck
    - unused
    - gocritic
    - gofmt
    - goimports
    - misspell
    - noctx          # ensure context is passed to DB/HTTP calls
    - exhaustive     # all switch cases covered
```

**Add to Makefile**:
```makefile
lint:
    golangci-lint run ./...

lint-fix:
    golangci-lint run --fix ./...

docs:
    swag init -g cmd/api/main.go -o docs/ --parseDependency --parseInternal

generate:
    go generate ./...

build-all:
    GOOS=linux   GOARCH=amd64 go build -o bin/erp-linux-amd64    ./cmd/api
    GOOS=darwin  GOARCH=arm64 go build -o bin/erp-darwin-arm64   ./cmd/api
    GOOS=windows GOARCH=amd64 go build -o bin/erp-windows-amd64.exe ./cmd/api

benchmark:
    go test -bench=. -benchmem ./...

vet:
    go vet ./...
```

---

### 9. Final Integration & End-to-End Testing (12%)
**Location**: `tests/e2e/` and `docs/postman/`

**Files to Create**:
- `tests/e2e/full_flow_test.go` — Complete user journey from registration to RBAC-protected resource access
- `docs/postman/Kadahapola_ERP.postman_collection.json` — Full Postman collection
- `docs/postman/Kadahapola_ERP.postman_environment.json` — Dev + staging environment variables

**Postman Collection Structure**:
```
Kadahapola ERP API
├── 00. Health
│   ├── Health Check
│   ├── Liveness Probe
│   └── Readiness Probe
├── 01. Auth
│   ├── Register
│   ├── Login (save tokens to env vars)
│   ├── Refresh Token
│   ├── Logout
│   └── Get Current User
├── 02. User Management
│   ├── List Users
│   ├── Get User by ID
│   ├── Update User
│   ├── Delete User
│   ├── Update Profile
│   ├── Change Password
│   └── Assign / Remove Role
├── 03. Role Management
│   ├── List Roles
│   ├── Create Role
│   ├── Assign Permissions to Role
│   └── Delete Role
├── 04. Permissions
│   └── List Permissions
├── 05. Tenants (super_admin)
│   ├── List Tenants
│   ├── Create Tenant
│   └── Update Tenant
├── 06. Audit Logs
│   └── List Audit Logs
├── 07. Feature Flags
│   ├── List Flags
│   ├── Create Flag
│   ├── Enable / Disable Flag
│   └── Evaluate Flags for Current User
├── 08. System
│   └── List Modules
└── 09. RBAC Verification
    ├── Access as super_admin → 200
    ├── Access as admin → 200
    ├── Access as user (no permission) → 403
    └── Access unauthenticated → 401
```

**Load Testing** (using `hey` or `k6`):
- Target: 500 concurrent requests, 10,000 total
- Measure: p50, p95, p99 latency; error rate; throughput
- Document results in `docs/PERFORMANCE_BENCHMARKS.md`
- Minimum pass criteria: p95 < 200ms, error rate < 0.1%

---

## File Structure

```
internal/
├── infrastructure/
│   ├── featureflags/
│   │   ├── flag.go                         ← NEW
│   │   ├── store.go                        ← NEW
│   │   ├── redis_store.go                  ← NEW
│   │   ├── db_store.go                     ← NEW
│   │   ├── evaluator.go                    ← NEW
│   │   └── middleware.go                   ← NEW
│   ├── metrics/
│   │   ├── metrics.go                      ← NEW
│   │   └── collector.go                    ← NEW
│   ├── http/
│   │   └── middleware/
│   │       ├── security_headers.go         ← NEW
│   │       ├── rate_limit.go               ← NEW (enhanced)
│   │       ├── csrf.go                     ← NEW
│   │       └── metrics.go                  ← NEW
│   ├── security/
│   │   ├── account_lockout.go              ← NEW
│   │   ├── input_sanitizer.go              ← NEW
│   │   └── ip_list.go                      ← NEW
│   └── database/
│       ├── seeder/
│       │   ├── seeder.go                   ← NEW
│       │   ├── default_tenant.go           ← NEW
│       │   ├── default_permissions.go      ← NEW
│       │   ├── default_roles.go            ← NEW
│       │   ├── default_users.go            ← NEW
│       │   └── development_data.go         ← NEW
│       └── postgres/
│           └── slow_query_logger.go        ← NEW
├── interfaces/
│   └── http/
│       └── handlers/
│           └── featureflags/
│               ├── handler.go              ← NEW
│               └── dto.go                  ← NEW
pkg/
└── validation/
    ├── validator.go                         ← NEW
    ├── custom_rules.go                     ← NEW
    ├── query_params.go                     ← NEW
    └── path_params.go                      ← NEW
cmd/
└── seed/
    └── main.go                             ← NEW
docs/
├── swagger.go                              ← NEW
├── docs.go                                 ← AUTO-GENERATED
└── postman/
    ├── Kadahapola_ERP.postman_collection.json   ← NEW
    └── Kadahapola_ERP.postman_environment.json  ← NEW
tests/
├── setup/
│   ├── db.go                               ← NEW
│   ├── redis.go                            ← NEW
│   ├── server.go                           ← NEW
│   ├── fixtures.go                         ← NEW
│   └── helpers.go                          ← NEW
├── mocks/
│   ├── user_repository.go                  ← NEW
│   ├── role_repository.go                  ← NEW
│   ├── event_bus.go                        ← NEW
│   ├── audit_logger.go                     ← NEW
│   └── feature_flag_store.go               ← NEW
├── integration/
│   ├── auth_test.go                        ← NEW
│   ├── user_test.go                        ← NEW
│   ├── role_test.go                        ← NEW
│   ├── rbac_test.go                        ← NEW
│   ├── tenant_test.go                      ← NEW
│   ├── audit_test.go                       ← NEW
│   ├── events_test.go                      ← NEW
│   ├── featureflags_test.go               ← NEW
│   └── health_test.go                      ← NEW
├── e2e/
│   └── full_flow_test.go                   ← NEW
└── batch_5/
    └── TEST_BATCH_5_PRODUCTION_READINESS.md ← NEW (test spec)
.golangci.yml                               ← NEW
```

---

## Implementation Order

### Phase 1: Feature Flags (Steps 1–3)
1. Write migration 000009 (feature_flags table), run it
2. Implement FlagStore (Redis + DB), Evaluator, Middleware
3. Build feature flag API endpoints; test with Postman

### Phase 2: Validation & Security Headers (Steps 4–6)
4. Build enhanced validation: custom rules, query param + path param validators
5. Implement security headers middleware; wire globally in server setup
6. Implement account lockout service; integrate into login handler in auth (Batch 3)

### Phase 3: Rate Limiting Enhancement (Step 7)
7. Replace basic rate limiter with tiered per-IP/per-user/per-route config; test lockout triggers

### Phase 4: Observability (Steps 8–9)
8. Define all Prometheus metric vars; wire HTTP metrics middleware; add slow query GORM callback
9. Add Go runtime metrics collector; verify `/metrics` exposes all expected metric names

### Phase 5: Swagger (Steps 10–11)
10. Install swag, add `docs/swagger.go` main annotation, add `swag init` to Makefile
11. Add `// @Summary` annotations to every handler; run `make docs`; verify Swagger UI loads and all endpoints appear

### Phase 6: Seeding (Steps 12–13)
12. Implement all seeder files; create `cmd/seed/main.go` CLI
13. Run `make seed` against dev DB; verify admin user can log in and has all permissions

### Phase 7: Testing Framework (Steps 14–16)
14. Set up test infrastructure (separate test DB, Redis DB 15, test server, fixtures)
15. Write all mock interfaces
16. Write and run integration tests; fix any failing cases; achieve 80%+ coverage

### Phase 8: E2E & Final Polish (Steps 17–19)
17. Write Postman collection covering all endpoints; test RBAC scenarios manually
18. Write `tests/e2e/full_flow_test.go`; run load test, document results in `PERFORMANCE_BENCHMARKS.md`
19. Run `make lint`; fix all lint errors; run full test suite one final time; update `PROJECT_ROADMAP.md` Phase 0 checkboxes

---

## Dependencies to Add

```bash
# Swagger
github.com/swaggo/swag                       # swag CLI (install separately)
github.com/gofiber/swagger                   # Swagger UI middleware for Fiber

# Validation
github.com/go-playground/validator/v10       # likely already present from Batch 2 plan

# HTML sanitization
github.com/microcosm-cc/bluemonday           # HTML sanitizer for input sanitization

# golangci-lint (install separately, not a go dep)
# hey or k6 for load testing (install separately)
```

---

## Testing Plan

After implementation, verify:

1. Feature flags: create a flag, disable it, verify feature returns 404/disabled; enable, verify accessible
2. Validation: send malformed requests to every endpoint; confirm consistent VALIDATION_ERROR response with field-level messages
3. Security headers: check all response headers with `curl -I`; verify HSTS present in production config
4. Account lockout: attempt login 5 times with wrong password; verify 429 or 423 with lockout message; wait 15 min or use Redis DEL to clear; verify login works again
5. Rate limiting: hammer `/api/v1/auth/login` beyond 10 req/min; verify 429 returned
6. Prometheus metrics: hit several endpoints; call `/metrics`; verify `http_requests_total`, `http_request_duration_seconds`, `auth_login_attempts_total` present with correct labels
7. Swagger UI: open `/docs/`; verify all endpoints listed; test authentication flow through Swagger UI
8. Seeding: fresh DB, `make migrate-up && make seed`; verify admin user login works; verify all roles/permissions exist
9. Integration tests: `make test-integration`; all tests must pass
10. Coverage: `make test-coverage`; must be ≥ 80% across domain + infrastructure + pkg
11. Lint: `make lint`; zero errors
12. Load test: 500 concurrent users; p95 < 200ms; error rate < 0.1%

---

## Success Criteria

- [ ] Feature flags work at user, tenant, and global level; Redis caching reduces DB calls to near-zero on repeated reads
- [ ] All validation errors return consistent JSON with field-level messages
- [ ] Custom validators (slug, strongpassword, phone, currency_code) all working
- [ ] All security headers present on every response
- [ ] Account locked after 5 failed logins; unlocks after TTL
- [ ] Rate limiting enforced per tier (IP, user, route)
- [ ] All Prometheus metrics defined and populated after API calls
- [ ] Slow queries (>200ms) logged at WARN level with SQL
- [ ] Swagger UI loads at `/docs/`; all endpoints documented with request/response models
- [ ] `make seed` creates default tenant, all permissions, all roles, super_admin user
- [ ] Integration test suite passes 100%
- [ ] Code coverage ≥ 80% for core packages
- [ ] `make lint` returns zero errors
- [ ] Load test: p95 latency < 200ms at 500 concurrent users
- [ ] Performance benchmarks documented in `docs/PERFORMANCE_BENCHMARKS.md`
- [ ] Postman collection covers all endpoints and RBAC scenarios
- [ ] All Phase 0 checkboxes in `PROJECT_ROADMAP.md` checked off

---

## Integration with Previous Batches

Batch 5 uses and completes:
- ✅ Batch 1: Config, Logger, PostgreSQL, Redis, Docker
- ✅ Batch 2: HTTP server, middleware stack, response helpers, health + metrics endpoints
- ✅ Batch 3: Auth handlers (add account lockout integration), auth middleware, JWT
- ✅ Batch 4: RBAC enforcer (permission middleware on feature flag API), tenant context, audit logger, event bus, module registry

---

## What Comes Next (Phase 1)

After Batch 5 is complete, Phase 0 is **done**. The next step is Phase 1: the Sales Module.

The Sales module will be the template for all subsequent business modules. It will follow the Module interface defined in Batch 4 and use all infrastructure built in Batches 1–5.

Reference: `PROJECT_ROADMAP.md` → Phase 1 (sections 1.1–1.7)

---

## Estimated Time: 10–12 hours

---

**Author**: Sanjaya Weerasinghe
**Date**: 2026-05-22
**Next**: Start with Phase 1 — Feature Flags (migration + store implementation)
