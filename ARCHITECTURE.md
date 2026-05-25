# Kadahapola ERP Backend — Architecture & Implementation Reference

> Last updated: 2026-05-22  
> Completion: Batches 1–5 (foundation through security hardening). Master Data module and business modules pending.

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Technology Stack](#2-technology-stack)
3. [Development Environment](#3-development-environment)
4. [Project Structure](#4-project-structure)
5. [Architectural Layers](#5-architectural-layers)
6. [Configuration System](#6-configuration-system)
7. [Database Schema](#7-database-schema)
8. [Domain Layer](#8-domain-layer)
9. [Authentication & JWT](#9-authentication--jwt)
10. [RBAC & Authorization](#10-rbac--authorization)
11. [HTTP Server & Middleware Pipeline](#11-http-server--middleware-pipeline)
12. [API Reference](#12-api-reference)
13. [Feature Flags](#13-feature-flags)
14. [Security Hardening](#14-security-hardening)
15. [Event Bus](#15-event-bus)
16. [Audit Logging](#16-audit-logging)
17. [Observability & Metrics](#17-observability--metrics)
18. [Validation](#18-validation)
19. [Error Handling](#19-error-handling)
20. [Module System](#20-module-system)
21. [Database Migrations](#21-database-migrations)
22. [Seeding](#22-seeding)
23. [CLI Tools](#23-cli-tools)
24. [Known Issues & Decisions](#24-known-issues--decisions)

---

## 1. Project Overview

Kadahapola ERP is a multi-tenant, role-based Enterprise Resource Planning backend built in Go for a Sri Lankan manufacturing and sales/distribution business. It is designed as a modular monolith: a single deployable binary with a module registry that allows business modules (inventory, manufacturing, HR, etc.) to be added without touching the core.

**Go module name:** `erp-system`  
**Go version:** 1.24.1  
**Primary language:** Go  
**Database:** PostgreSQL 16  
**Cache/Queue:** Redis 7

---

## 2. Technology Stack

| Concern | Library | Version | Notes |
|---------|---------|---------|-------|
| HTTP framework | `github.com/gofiber/fiber/v2` | 2.52.9 | Fast, Express-inspired |
| ORM | `gorm.io/gorm` | 1.31.1 | PostgreSQL driver via pgx |
| Database driver | `gorm.io/driver/postgres` | 1.6.0 | Uses pgx under the hood |
| Redis client | `github.com/redis/go-redis/v9` | 9.16.0 | Token blacklist, cache, streams |
| JWT | `github.com/golang-jwt/jwt/v5` | 5.3.0 | HS256 signing |
| RBAC | `github.com/casbin/casbin/v2` | 2.135.0 | Custom DB adapter (no gorm-adapter) |
| Migrations | `github.com/golang-migrate/migrate/v4` | 4.19.0 | SQL file-based, forward/back |
| Validation | `github.com/go-playground/validator/v10` | 10.28.0 | Struct tags + custom rules |
| Config | `github.com/spf13/viper` | 1.21.0 | YAML file + env var overrides |
| Logging | `go.uber.org/zap` | 1.27.0 | Structured JSON/console |
| Metrics | `github.com/prometheus/client_golang` | 1.23.2 | Prometheus exposition format |
| Crypto | `golang.org/x/crypto` | 0.46.0 | bcrypt for passwords |
| UUID | `github.com/google/uuid` | 1.6.0 | Event IDs |
| Hot reload | Air | 1.61.5 | **Pinned** — v1.65.3 requires Go ≥ 1.25 |

---

## 3. Development Environment

### Docker Compose Stack

```
docker-compose -f docker-compose.dev.yml up --build
```

| Service | Container name | Host port | Container port |
|---------|---------------|-----------|----------------|
| Go API (Air) | `erp-api-dev` | 3000 | 3000 |
| PostgreSQL 16 | `erp-postgres-dev` | 5433 | 5432 |
| Redis 7 | `erp-redis-dev` | 6380 | 6379 |

> **Why port 6380?** Port 6379 is occupied by a separate Docker project on the development machine. The dev compose maps Redis to 6380 on the host to avoid conflicts.

### Hot Reload

Air watches all `.go` files under `/app` (excluding `tmp/`). Every file save triggers a full recompile inside the container. The compiled binary is placed at `tmp/main`.

**Air version is pinned at v1.61.5** in `Dockerfile.dev`. Upgrading to v1.65.3 fails with:

```
note: module requires Go >= 1.25
```

Go 1.24.1 is the current version; do not upgrade Air until Go is upgraded.

### Running Commands Inside the Container

Since there is no local Go installation, all Go commands must be run through Docker:

```bash
# Build
docker exec erp-api-dev sh -c "go build ./..."

# Run migrations
docker exec erp-api-dev sh -c "go run cmd/migrate/main.go -command=up"

# Run seeders
docker exec erp-api-dev sh -c "SEED_ADMIN_EMAIL=admin@kadahapola.com SEED_ADMIN_PASSWORD=Admin@12345 go run cmd/seed/main.go all"
```

---

## 4. Project Structure

```
Kadahapola BE/
├── cmd/
│   ├── api/main.go          # API server entry point
│   ├── migrate/main.go      # Migration CLI
│   └── seed/main.go         # Seeder CLI
│
├── config/                  # YAML config files (config.yaml)
│
├── docs/
│   └── postman/
│       └── kadahapola_erp_api.json  # Postman collection
│
├── internal/
│   ├── domain/              # Business entities and repository interfaces
│   │   ├── base_entity.go
│   │   ├── audit/
│   │   ├── permission/
│   │   ├── rbac/            # Permission constants, role defaults, Enforcer interface
│   │   ├── role/
│   │   ├── tenant/
│   │   └── user/
│   │
│   ├── infrastructure/      # Technical implementations
│   │   ├── audit/           # Audit logger
│   │   ├── authorization/casbin/   # Casbin enforcer + custom DB adapter
│   │   ├── cache/redis/     # Redis connection, cache, token blacklist
│   │   ├── config/          # Viper-based config loading
│   │   ├── database/
│   │   │   ├── migrations/  # SQL migration files (000001–000009)
│   │   │   ├── postgres/    # GORM connection
│   │   │   └── seeder/      # Production and development seeders
│   │   ├── events/          # Redis Streams event bus
│   │   ├── featureflags/    # Feature flag store (DB + Redis cache)
│   │   ├── http/
│   │   │   ├── handlers/    # Health, Metrics (Prometheus)
│   │   │   ├── middleware/  # All middleware
│   │   │   └── server/      # Fiber server + route registration
│   │   ├── persistence/     # Repository implementations
│   │   │   ├── audit/
│   │   │   ├── permission/
│   │   │   ├── role/
│   │   │   ├── tenant/
│   │   │   └── user/
│   │   └── security/        # Account lockout
│   │
│   ├── interfaces/
│   │   └── http/handlers/   # HTTP request/response handlers
│   │       ├── auth/
│   │       ├── audit/
│   │       ├── featureflags/
│   │       ├── role/
│   │       ├── tenant/
│   │       └── user/
│   │
│   └── modules/             # Module registry + interface
│
├── pkg/                     # Shared, framework-independent packages
│   ├── errors/              # AppError, error codes, HTTP status mapping
│   ├── http/                # Standard response helpers, request helpers
│   ├── jwt/                 # JWT manager (generate + validate)
│   ├── logger/              # Zap singleton
│   └── validation/          # Validator singleton, custom rules
│
├── test_results/
│   ├── batch_2/
│   ├── batch_4/
│   └── batch_5/
│
├── .golangci.yml            # Linter configuration
├── docker-compose.yml       # Production compose
├── docker-compose.dev.yml   # Development compose
├── Dockerfile               # Production image
├── Dockerfile.dev           # Development image (Air hot-reload)
└── go.mod / go.sum
```

---

## 5. Architectural Layers

The codebase uses **Clean Architecture** with three layers:

```
┌─────────────────────────────────────────────────────────────────┐
│  INTERFACES LAYER  (internal/interfaces/)                        │
│  HTTP handlers — parse requests, call domain, serialize response │
├─────────────────────────────────────────────────────────────────┤
│  INFRASTRUCTURE LAYER  (internal/infrastructure/)                │
│  DB repos, Redis, Casbin, event bus, middleware, server config   │
├─────────────────────────────────────────────────────────────────┤
│  DOMAIN LAYER  (internal/domain/)                                │
│  Business entities, repository interfaces, value objects         │
└─────────────────────────────────────────────────────────────────┘
```

**Dependency direction:** Interfaces → Infrastructure → Domain. Domain has zero external imports.

**Repository pattern:** Every domain aggregate (User, Role, Permission, Tenant, Audit) has a `Repository` interface in the domain layer and a `PostgresRepository` implementation in `infrastructure/persistence/`. Handlers receive the domain interface, not the implementation.

**Import aliases** used throughout (set once per file at the top):

| Alias | Package |
|-------|---------|
| `apperrors` | `erp-system/pkg/errors` |
| `httputil` | `erp-system/pkg/http` |
| `domainuser` | `erp-system/internal/domain/user` |
| `rediscache` | `erp-system/internal/infrastructure/cache/redis` |
| `jwtpkg` | `erp-system/pkg/jwt` |
| `casbininfra` | `erp-system/internal/infrastructure/authorization/casbin` |
| `auditinfra` | `erp-system/internal/infrastructure/audit` |
| `ffhandler` | `erp-system/internal/interfaces/http/handlers/featureflags` |

---

## 6. Configuration System

Configuration is loaded by `internal/infrastructure/config/config.go` using **Viper**. Sources are applied in priority order (highest wins):

1. Environment variables (bound explicitly via `viper.BindEnv`)
2. Config file: `config/config.yaml` (YAML, optional)
3. Coded defaults set in `setDefaults()`

The config struct is a singleton. Calling `config.Load()` a second time returns the cached value.

### Config Groups

```go
type Config struct {
    Server   ServerConfig    // Port, timeouts, body limit, CORS toggle
    Database DatabaseConfig  // Host/port/user/pass/dbname, pool settings
    Redis    RedisConfig     // Host/port/pass, pool settings
    JWT      JWTConfig       // Secret, access expiry (1h), refresh expiry (7d), issuer
    CORS     CORSConfig      // Origins, methods, headers, credentials
    Logging  LoggingConfig   // Level, format (json/console), output path
    App      AppConfig       // Name, version, environment, debug flag
}
```

### Key Defaults

| Key | Default | Notes |
|-----|---------|-------|
| `server.port` | 3000 | |
| `server.body_limit` | 4 MB | |
| `database.max_open_conns` | 100 | |
| `database.max_idle_conns` | 10 | |
| `database.conn_max_lifetime` | 1h | |
| `jwt.access_token_expiry` | 1h | |
| `jwt.refresh_token_expiry` | 168h (7d) | |
| `redis.pool_size` | 10 | |
| `logging.level` | info | |
| `logging.format` | json | |

### Environment Variable Bindings

| Env var | Config key |
|---------|-----------|
| `PORT` | `server.port` |
| `DB_HOST` | `database.host` |
| `DB_PORT` | `database.port` |
| `DB_USER` | `database.user` |
| `DB_PASSWORD` | `database.password` |
| `DB_NAME` | `database.database` |
| `DB_SSL_MODE` | `database.ssl_mode` |
| `REDIS_HOST` | `redis.host` |
| `REDIS_PORT` | `redis.port` |
| `REDIS_PASSWORD` | `redis.password` |
| `JWT_SECRET` | `jwt.secret` |
| `ENV` | `app.environment` |

### Validation Rules

- Production environment requires `JWT_SECRET` to not be the default value
- `logging.level` must be one of: `debug`, `info`, `warn`, `error`, `fatal`
- `app.environment` must be one of: `development`, `staging`, `production`

---

## 7. Database Schema

All schema changes are applied through SQL migration files in `internal/infrastructure/database/migrations/`. Migration tool: `golang-migrate/v4`. Migration state is tracked in the `schema_migrations` table.

### Migration History

| # | File | What it creates |
|---|------|----------------|
| 001 | `create_users_table` | `users` table + `update_updated_at_column()` trigger function |
| 002 | `create_tenants_table` | `tenants` table with JSONB `config` column |
| 003 | `add_tenant_to_users` | `tenant_id` FK column on `users` |
| 004 | `create_roles_table` | `roles` table with soft-delete |
| 005 | `create_permissions_table` | `permissions` table with `UNIQUE(resource, action)` |
| 006 | `create_role_permissions_table` | `role_permissions` join table |
| 007 | `create_user_roles_table` | `user_roles` join table |
| 008 | `create_audit_logs_table` | `audit_logs` append-only table |
| 009 | `create_feature_flags_table` | `feature_flags` table with JSONB `rules` |

### Table Definitions

#### `users`
```sql
id            BIGSERIAL PRIMARY KEY
email         VARCHAR(255) UNIQUE NOT NULL
password_hash VARCHAR(255) NOT NULL
first_name    VARCHAR(100)
last_name     VARCHAR(100)
role          VARCHAR(50)  NOT NULL DEFAULT 'user'      -- legacy role field
status        VARCHAR(20)  NOT NULL DEFAULT 'active'    -- active|inactive|suspended
created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()       -- auto-maintained by trigger
last_login_at TIMESTAMP
deleted_at    TIMESTAMP                                 -- soft delete (NULL = not deleted)
```
Indexes: `email`, `role`, `status`, `deleted_at`

#### `tenants`
```sql
id         BIGSERIAL PRIMARY KEY
name       VARCHAR(255) NOT NULL
slug       VARCHAR(100) UNIQUE NOT NULL                 -- URL-safe identifier
plan       VARCHAR(50)  NOT NULL DEFAULT 'standard'    -- standard|professional|enterprise
status     VARCHAR(20)  NOT NULL DEFAULT 'active'
config     JSONB        NOT NULL DEFAULT '{}'           -- arbitrary tenant config
created_at TIMESTAMP    NOT NULL DEFAULT NOW()
updated_at TIMESTAMP    NOT NULL DEFAULT NOW()
```

#### `roles`
```sql
id          BIGSERIAL PRIMARY KEY
tenant_id   BIGINT REFERENCES tenants(id) ON DELETE CASCADE  -- NULL = system-wide
name        VARCHAR(100) NOT NULL
description VARCHAR(255)
is_system   BOOLEAN      NOT NULL DEFAULT FALSE
created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
deleted_at  TIMESTAMP
```
Unique constraint: `COALESCE(tenant_id, 0), name` WHERE `deleted_at IS NULL` — allows same role name across tenants, prevents duplicates within a tenant.

#### `permissions`
```sql
id         BIGSERIAL PRIMARY KEY
resource   VARCHAR(100) NOT NULL   -- e.g. 'users'
action     VARCHAR(100) NOT NULL   -- e.g. 'create'
UNIQUE(resource, action)
```

#### `role_permissions`
```sql
role_id       BIGINT REFERENCES roles(id)       ON DELETE CASCADE
permission_id BIGINT REFERENCES permissions(id) ON DELETE CASCADE
PRIMARY KEY (role_id, permission_id)
```

#### `user_roles`
```sql
user_id BIGINT REFERENCES users(id) ON DELETE CASCADE
role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE
PRIMARY KEY (user_id, role_id)
```

#### `audit_logs` (append-only — no UPDATE/DELETE triggers)
```sql
id          BIGSERIAL PRIMARY KEY
tenant_id   BIGINT REFERENCES tenants(id) ON DELETE SET NULL
user_id     BIGINT REFERENCES users(id)   ON DELETE SET NULL
action      VARCHAR(100) NOT NULL   -- create|update|delete
resource    VARCHAR(100) NOT NULL   -- users|roles|tenants|...
resource_id VARCHAR(100)
old_values  JSONB
new_values  JSONB
ip_address  VARCHAR(45)
user_agent  VARCHAR(500)
created_at  TIMESTAMP NOT NULL DEFAULT NOW()
```
Indexes: `tenant_id`, `user_id`, `resource`, `action`, `created_at`, composite `(resource, resource_id)`

#### `feature_flags`
```sql
id          BIGSERIAL PRIMARY KEY
tenant_id   BIGINT REFERENCES tenants(id) ON DELETE CASCADE  -- NULL = global flag
name        VARCHAR(100) NOT NULL
description TEXT
enabled     BOOLEAN NOT NULL DEFAULT false
rules       JSONB   NOT NULL DEFAULT '{}'  -- {"user_ids":[1,2], "percentage": 50}
created_at  TIMESTAMP NOT NULL DEFAULT NOW()
updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
```
Unique indexes (partial, handles NULL tenant_id):
- `uidx_feature_flags_tenant_name` ON `(tenant_id, name)` WHERE `tenant_id IS NOT NULL`
- `uidx_feature_flags_global_name` ON `(name)` WHERE `tenant_id IS NULL`

> **Why two partial indexes instead of `UNIQUE(COALESCE(tenant_id, 0), name)`?**  
> PostgreSQL does not support function calls inside inline `UNIQUE` table constraints. Two partial unique indexes are the standard PostgreSQL idiom for nullable composite uniqueness.

---

## 8. Domain Layer

The domain layer contains pure Go — no framework imports, no database imports. It defines what data looks like and what operations exist, but not how they are implemented.

### Base Entity

```go
// internal/domain/base_entity.go
type BaseEntity struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### User Entity (`internal/domain/user/entity.go`)

Key behaviors on the entity:

| Method | What it does |
|--------|-------------|
| `SetPassword(password string) error` | bcrypt.GenerateFromPassword(cost=10), stores in PasswordHash |
| `CheckPassword(password string) bool` | bcrypt.CompareHashAndPassword — constant-time |
| `IsActive() bool` | `status == "active" && deleted_at == nil` |
| `SoftDelete()` | Sets `deleted_at = now`, `status = "inactive"` |
| `FullName() string` | Returns `FirstName + " " + LastName`, falls back to email |

Status constants: `StatusActive`, `StatusInactive`, `StatusSuspended`  
Role constants (legacy field): `RoleAdmin`, `RoleManager`, `RoleUser`, `RoleGuest`

### Tenant Entity (`internal/domain/tenant/entity.go`)

Plans: `PlanStandard`, `PlanProfessional`, `PlanEnterprise`  
Statuses: `StatusActive`, `StatusInactive`, `StatusSuspended`  
Config stored as `json.RawMessage` JSONB — arbitrary key-value pairs.

### Role Entity (`internal/domain/role/entity.go`)

Has `TenantID *uint` — nil means it is a system-wide role available to all tenants.  
Has `IsSystem bool` — system roles cannot be deleted by tenants.  
Has a many-to-many slice of `Permission` for preloading.

### Permission Entity (`internal/domain/permission/entity.go`)

```go
type Permission struct {
    ID       uint   `gorm:"primarykey"`
    Resource string `gorm:"size:100;not null"`
    Action   string `gorm:"size:100;not null"`
}

func (p Permission) Key() string { return p.Resource + ":" + p.Action }
```

### Audit Log Entity (`internal/domain/audit/entity.go`)

Two structs:
- `Log` — the persisted record (ID, TenantID, UserID, Action, Resource, ResourceID, OldValues, NewValues, IPAddress, UserAgent, CreatedAt)
- `Entry` — the input struct passed to the audit logger (no ID, no CreatedAt — those are generated on save)

### RBAC Types (`internal/domain/rbac/`)

#### Permission Constants

All 19 permissions follow `resource:action` format:

```
users:create    users:read    users:update    users:delete
roles:create    roles:read    roles:update    roles:delete
permissions:read
tenants:create  tenants:read  tenants:update  tenants:delete
modules:read    modules:manage
audit:read
feature_flags:read   feature_flags:manage
```

`AllPermissions []string` — the canonical list used by the seeder.

#### Role Constants and Default Permissions

| Role | Default permissions |
|------|-------------------|
| `super_admin` | `*` (wildcard — seeder assigns ALL permissions via subquery) |
| `admin` | All user, role, permission, module, audit, feature_flags permissions |
| `manager` | users:read/update, roles:read, permissions:read, modules:read, feature_flags:read |
| `user` | users:read only |
| `guest` | None |

#### Enforcer Interface

```go
type Enforcer interface {
    Enforce(ctx context.Context, subject, resource, action string) (bool, error)
    AddPolicy(role, resource, action string) error
    RemovePolicy(role, resource, action string) error
    AddRoleForUser(user, role string) error
    RemoveRoleForUser(user, role string) error
    GetRolesForUser(user string) ([]string, error)
    LoadPolicy() error
}
```

---

## 9. Authentication & JWT

### JWT Manager (`pkg/jwt/jwt.go`)

Algorithm: **HS256** (HMAC-SHA256)

#### Token Types

| Type | `type` claim | Expiry default | Purpose |
|------|-------------|----------------|---------|
| Access | `"access"` | 1 hour | Sent in `Authorization: Bearer` header |
| Refresh | `"refresh"` | 7 days | Used only at `POST /api/v1/auth/refresh` |

#### JWT Claims

```go
type Claims struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    Type   string `json:"type"`       // "access" or "refresh"
    jwt.RegisteredClaims               // exp, iat, nbf, iss, sub
}
```

The `sub` field is set to the stringified `user_id`. The `iss` field is `"erp-system"`.

### Token Blacklist (`internal/infrastructure/cache/redis/token_blacklist.go`)

Revoked tokens (from logout) are stored in Redis with a TTL equal to the token's remaining lifetime. The token itself is never stored raw — only a **SHA256 hash** of the token is used as the Redis key:

```
Key format:  blacklist:{sha256hex(token)}
Value:       "1"
TTL:         token remaining lifetime
```

This prevents information leakage if Redis is compromised and keeps keys at a fixed 64-character length regardless of token length.

### Auth Flow — Login

```
POST /api/v1/auth/login
  │
  ├─ 1. Check account lockout (Redis key: lockout:{email})
  │      If locked → 429 TOO_MANY_REQUESTS
  │
  ├─ 2. GetByEmail from DB
  │
  ├─ 3. bcrypt.CompareHashAndPassword
  │      If u == nil OR wrong password:
  │        - RecordFailure (increments lockout counter)
  │        → 401 INVALID_CREDENTIALS
  │
  ├─ 4. Check u.IsActive()
  │      If inactive/suspended → 401 UNAUTHORIZED
  │
  ├─ 5. GenerateAccessToken + GenerateRefreshToken
  │
  ├─ 6. UpdateLastLogin (fire-and-forget, error ignored)
  │
  ├─ 7. lockout.Reset (clears failure counter on success)
  │
  └─ 8. Return {access_token, refresh_token, token_type, expires_in, user}
```

### Auth Flow — Token Refresh

```
POST /api/v1/auth/refresh
  │
  ├─ 1. Parse and validate refresh token (signature + expiry)
  ├─ 2. Verify token.Type == "refresh"
  ├─ 3. Check blacklist (refresh token can also be blacklisted)
  ├─ 4. Load user from DB, check IsActive()
  └─ 5. Issue new access token (refresh token is NOT rotated)
```

### Auth Flow — Logout

```
POST /api/v1/auth/logout  (requires Authenticate middleware)
  │
  ├─ 1. Extract Bearer token from Authorization header
  ├─ 2. ValidateToken to get expiry (if already invalid, treat as success)
  ├─ 3. blacklist.Add(token, ttl = time.Until(claims.ExpiresAt))
  └─ 4. Return 200 OK
```

---

## 10. RBAC & Authorization

### Casbin Setup

The system uses **Casbin v2** with a **custom database adapter** (NOT `gorm-adapter`).

#### Why a custom adapter?

`gorm-adapter/v3 v3.32+` switched its internal implementation to use `casbin/v3/persist.Adapter`. When passed to `casbin/v2`'s `NewEnforcer`, a runtime type-assertion panic occurs:

```
panic: interface conversion: *gormadapter.Adapter is not persist.Adapter:
       missing method LoadPolicy
```

The fix: `internal/infrastructure/authorization/casbin/enforcer.go` implements a private `dbAdapter` struct that directly implements `casbin/v2/persist.Adapter` using raw SQL against our own `role_permissions` and `user_roles` tables. The Casbin `casbin_rule` table is never used.

#### Casbin Model

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act || r.sub == "super_admin"
```

The critical part is the matcher: `g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act || r.sub == "super_admin"`.

This means:
- Normal check: user's role must be granted a policy for (resource, action)
- **`super_admin` bypass**: if the subject string is literally `"super_admin"`, ALL checks pass

#### Policy Loading

On startup, `LoadPolicy()` is called, which runs two SQL queries:

```sql
-- Load role → permission policies  (Casbin "p" rules)
SELECT r.name AS role_name, p.resource, p.action
FROM role_permissions rp
JOIN roles r       ON r.id = rp.role_id       AND r.deleted_at IS NULL
JOIN permissions p ON p.id = rp.permission_id

-- Load user → role assignments  (Casbin "g" rules)
SELECT ur.user_id, r.name AS role_name
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL
```

These are loaded into the in-memory policy store as:
- `p, admin, users, create` (policy line)
- `g, 1, super_admin` (grouping line — user ID 1 has role super_admin)

The `CasbinEnforcer.AddPolicy/RemovePolicy` and `AddRoleForUser/RemoveRoleForUser` methods mutate the in-memory policy but rely on the role/user handlers to persist changes to the DB. `LoadPolicy()` can be called to reload from DB at any time.

#### Permission Middleware

```go
// internal/infrastructure/http/middleware/permission.go
func RequirePermission(enforcer rbac.Enforcer, resource, action string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        role, ok := c.Locals("user_role").(string)
        if !ok || role == "" {
            return httputil.Unauthorized(c, "Authentication required")
        }
        allowed, err := enforcer.Enforce(c.Context(), role, resource, action)
        if err != nil {
            return httputil.InternalServerError(c, "Permission check failed")
        }
        if !allowed {
            return httputil.Forbidden(c, "Insufficient permissions")
        }
        return c.Next()
    }
}
```

The `user_role` local is set by the `Authenticate` middleware from the JWT `role` claim.

### Multi-Tenancy in RBAC

- System roles (`tenant_id IS NULL`): `super_admin`, `admin`, `manager`, `user`, `guest`
- Tenant-scoped roles: A tenant admin can create custom roles for their tenant (same `roles` table, `tenant_id` set)
- `TenantContext` middleware: extracts `tenant_id` from JWT; `super_admin` can override via `X-Tenant-ID` header to impersonate any tenant

---

## 11. HTTP Server & Middleware Pipeline

### Server Setup (`internal/infrastructure/http/server/server.go`)

Built on **Fiber v2**. Key configuration:

- Body limit: 4 MB
- Request ID: UUID v4, added to `X-Request-ID` response header
- Graceful shutdown: 10-second timeout, listens for `SIGINT`/`SIGTERM`

### Middleware Execution Order

Every request passes through this chain (registered in `cmd/api/main.go` and `routes.go`):

```
Request arrives
     │
     ▼
1.  RequestID              — Assigns X-Request-ID (UUID)
2.  CORS                   — If cfg.Server.EnableCORS == true
3.  Logger                 — Logs method, path, status, duration
4.  SecurityHeaders        — X-Frame-Options, CSP, etc.
     │
     ├── /health/*  →  HealthHandler  (no auth)
     ├── /metrics/* →  PrometheusHandler  (no auth)
     │
     └── /api/v1/...
           │
           ├── /auth/register     →  RateLimiter(10/min/IP)  → AuthHandler
           ├── /auth/login        →  RateLimiter(10/min/IP)  → AuthHandler
           ├── /auth/refresh      →  RateLimiter(20/min/IP)  → AuthHandler
           ├── /auth/logout       →  Authenticate → AuthHandler
           ├── /auth/me           →  Authenticate → AuthHandler
           │
           └── (all other routes)
                 │
                 ▼
5.  Authenticate           — JWT validation + blacklist check
                             Stores: user_id, user_email, user_role, token in locals
                 │
                 ▼
6.  TenantContext          — Sets tenant_id local from JWT claims
                             super_admin: override with X-Tenant-ID header
                 │
                 ▼
7.  RequirePermission      — Casbin Enforce(role, resource, action)
                 │
                 ▼
8.  AuditLog (optional)   — POST/PUT/PATCH/DELETE on 2xx → async audit entry
                 │
                 ▼
9.  PrometheusMetrics      — Records http_requests_total, duration, active
                 │
                 ▼
        Handler
```

### Middleware Details

#### RequestID (`middleware/request_id.go`)
Generates a UUID v4 and stores it in `X-Request-ID` response header and `c.Locals("request_id")`.

#### Authenticate (`middleware/auth.go`)
1. Reads `Authorization: Bearer <token>` header
2. `jwtManager.ValidateToken(token)` — parses and verifies signature + expiry
3. Verifies `claims.Type == "access"` (rejects refresh tokens)
4. `blacklist.IsBlacklisted(ctx, token)` — checks Redis
5. Stores in fiber locals: `user_id (uint)`, `user_email (string)`, `user_role (string)`, `token (string)`

#### TenantContext (`middleware/tenant.go`)
- For `super_admin`: if `X-Tenant-ID` header is present and parseable, uses that tenant
- For all others: tenant_id is expected to already be in locals (set via JWT claims in the auth handler — note: currently the auth handler does not embed tenant_id in JWT; this is a planned enhancement)

#### AuditLog (`middleware/audit.go`)
Only fires for `POST`, `PUT`, `PATCH`, `DELETE` methods AND only for 2xx status codes. Builds an `audit.Entry` from context locals and calls `auditLogger.LogAsync()` which spawns a goroutine so it never blocks the response.

Resource extraction: strips `/api/v1/` prefix, takes first path segment (e.g., `/api/v1/users/123` → `users`).

#### SecurityHeaders (`middleware/security_headers.go`)

| Header | Value |
|--------|-------|
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` |
| `X-XSS-Protection` | `1; mode=block` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `Permissions-Policy` | `geolocation=(), microphone=(), camera=()` |
| `Content-Security-Policy` | `default-src 'self'` |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` (production only) |

#### RateLimiter (`middleware/rate_limit.go`)

Redis INCR + EXPIRE sliding window implementation:

```go
count, err := r.client.Incr(ctx, key).Result()
if count == 1 {
    _ = r.client.Expire(ctx, key, window)  // set TTL on first hit
}
if count > int64(max) {
    return httputil.Error(c, apperrors.New(apperrors.CodeTooManyRequests, "..."))
}
```

- `PerIP(max, window)` — key: `ratelimit:ip:{client_ip}`
- `PerUser(max, window)` — key: `ratelimit:user:{user_id}` or falls back to IP if not authenticated
- **Fail-open**: Redis errors let the request through to avoid blocking legitimate traffic
- Sets response headers: `X-RateLimit-Limit` and `X-RateLimit-Remaining`

Applied rates:
- `POST /api/v1/auth/login` — 10 requests / 1 minute / IP
- `POST /api/v1/auth/register` — 10 requests / 1 minute / IP
- `POST /api/v1/auth/refresh` — 20 requests / 1 minute / IP

#### PrometheusMetrics (`middleware/metrics.go`)

Registered on the shared Prometheus registry (same registry as `/metrics/` endpoint):

```
http_requests_total{method, path, status}   -- counter
http_request_duration_seconds{method, path} -- histogram (default buckets)
http_active_requests                        -- gauge (in-flight requests)
```

---

## 12. API Reference

**Base URL:** `http://localhost:3000/api/v1`  
**Auth:** All protected routes require `Authorization: Bearer <access_token>`

### Standard Response Envelope

All responses use this envelope:

```json
// Success
{
  "success": true,
  "message": "Human-readable message",
  "data": { ... },      // omitted on no-content responses
  "meta": { ... }       // omitted unless paginated
}

// Error
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "fields": { "field_name": "error_tag" }   // validation errors only
  }
}
```

### Pagination Meta

```json
"meta": {
  "page": 1,
  "per_page": 20,
  "total": 150,
  "total_pages": 8
}
```

---

### Auth Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | None | Create a new user account |
| POST | `/auth/login` | None | Login, receive tokens |
| POST | `/auth/refresh` | None | Exchange refresh token for new access token |
| POST | `/auth/logout` | Bearer | Blacklist current access token |
| GET | `/auth/me` | Bearer | Get current user profile |

#### POST /auth/register
```json
// Request
{
  "email": "user@example.com",    // required, valid email
  "password": "Test@12345",        // required, strongpassword
  "first_name": "John",            // required
  "last_name": "Doe"               // required
}

// Response 201
{
  "success": true,
  "message": "Account created successfully",
  "data": { "id": 5, "email": "user@example.com", "role": "user", "status": "active", ... }
}
```

#### POST /auth/login
```json
// Request
{
  "email": "admin@kadahapola.com",
  "password": "Admin@12345"
}

// Response 200
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": { "id": 1, "email": "...", "role": "super_admin", ... }
  }
}

// 401 — wrong credentials
{ "success": false, "error": { "code": "INVALID_CREDENTIALS", "message": "Invalid email or password" } }

// 429 — account locked
{ "success": false, "error": { "code": "TOO_MANY_REQUESTS", "message": "Account temporarily locked..." } }
```

---

### User Management Endpoints

| Method | Path | Permission | Description |
|--------|------|-----------|-------------|
| GET | `/users` | `users:read` | List all users |
| GET | `/users/:id` | `users:read` | Get specific user |
| PUT | `/users/:id` | `users:update` | Update user |
| DELETE | `/users/:id` | `users:delete` | Soft-delete user |
| PUT | `/profile` | (authenticated) | Update own profile |
| PUT | `/password` | (authenticated) | Change own password |

---

### Role & Permission Management

| Method | Path | Permission | Description |
|--------|------|-----------|-------------|
| GET | `/roles` | `roles:read` | List all roles |
| POST | `/roles` | `roles:create` | Create role |
| GET | `/roles/:id` | `roles:read` | Get role |
| PUT | `/roles/:id` | `roles:update` | Update role |
| DELETE | `/roles/:id` | `roles:delete` | Delete role |
| POST | `/roles/:id/permissions` | `roles:update` | Assign permissions to role |
| DELETE | `/roles/:id/permissions/:perm_id` | `roles:update` | Remove permission from role |
| GET | `/permissions` | `permissions:read` | List all permissions |

---

### Tenant Management

| Method | Path | Permission | Description |
|--------|------|-----------|-------------|
| POST | `/tenants` | `tenants:create` | Create tenant |
| GET | `/tenants` | `tenants:read` | List tenants |
| GET | `/tenants/:id` | `tenants:read` | Get tenant |
| PUT | `/tenants/:id` | `tenants:update` | Update tenant |

---

### Feature Flag Endpoints

| Method | Path | Permission | Description |
|--------|------|-----------|-------------|
| GET | `/feature-flags` | `feature_flags:read` | List flags for current tenant |
| POST | `/feature-flags` | `feature_flags:manage` | Create flag |
| GET | `/feature-flags/:name` | `feature_flags:read` | Get flag by name |
| PUT | `/feature-flags/:name/enabled` | `feature_flags:manage` | Toggle enabled state |
| DELETE | `/feature-flags/:name` | `feature_flags:manage` | Delete flag |
| POST | `/feature-flags/:name/invalidate` | `feature_flags:manage` | Bust Redis cache entry |

---

### Audit Log Endpoint

| Method | Path | Permission | Query params |
|--------|------|-----------|-------------|
| GET | `/audit-logs` | `audit:read` | `resource`, `action`, `user_id`, `date_from`, `date_to`, `limit`, `offset` |

---

### Infrastructure Endpoints (no auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health/` | Health check — returns DB + Redis status |
| GET | `/health/live` | Kubernetes liveness probe |
| GET | `/health/ready` | Kubernetes readiness probe |
| GET | `/metrics/` | Prometheus metrics (OpenMetrics format) |

---

## 13. Feature Flags

### Architecture

Feature flags live in `internal/infrastructure/featureflags/`:

```
FeatureFlag entity  (flag.go)
Store interface     (store.go)
DBStore impl        (db_store.go)   ← PostgreSQL + Redis 5-min cache
```

### Flag Entity

```go
type FeatureFlag struct {
    ID          uint
    TenantID    *uint           // nil = global flag (applies to all tenants)
    Name        string          // unique within tenant (or globally)
    Description string
    Enabled     bool
    Rules       json.RawMessage // {"user_ids": [1,2,3], "percentage": 50.0}
}

type FlagRules struct {
    UserIDs    []uint  `json:"user_ids,omitempty"`   // explicit allow-list
    Percentage float64 `json:"percentage,omitempty"` // 0-100 rollout %
}
```

### Caching Strategy

- **Cache key:** `ff:{tenantID}:{name}` for tenant flags, `ff:global:{name}` for global flags
- **TTL:** 5 minutes
- **Read path:** Check Redis → if miss, query PostgreSQL → write to Redis
- **Write path (Set/Delete):** Write to PostgreSQL → call `Invalidate()` to delete Redis key

### `IsEnabledForUser`

Used for gradual rollouts and per-user targeting:

```go
func (s *DBStore) IsEnabledForUser(ctx, tenantID, userID, name) (bool, error):
    flag := Get(ctx, tenantID, name)
    if !flag.Enabled: return false
    if len(flag.Rules) == 0 or rules == {}: return flag.Enabled
    unmarshal rules
    for uid in rules.UserIDs:
        if uid == userID: return true  // explicit allow
    return flag.Enabled  // fallback to global flag state
```

> Percentage-based rollout is stored in `rules.Percentage` but the actual sampling logic (hash(userID) % 100 < percentage) is planned for a future implementation.

---

## 14. Security Hardening

### Account Lockout (`internal/infrastructure/security/account_lockout.go`)

Redis-backed progressive lockout:

| Failures | Lockout duration |
|----------|-----------------|
| 5–9 | 15 minutes |
| 10–19 | 1 hour |
| 20+ | 24 hours |

Redis key: `lockout:{email}` — stores JSON `{"attempts": N, "locked_until": "2026-..."}` with a 24-hour TTL.

**Flow in Login handler:**
1. `IsLocked(ctx, email)` → if locked, return 429 immediately (before DB query)
2. If DB lookup or password check fails → `RecordFailure(ctx, email)`
3. If login succeeds → `Reset(ctx, email)` (deletes the Redis key)

### Rate Limiting

See [Middleware Details](#middleware-details) above. Applied only to authentication endpoints. Not applied to authenticated API endpoints (by design — users are already trusted after login).

### Security Headers

Applied globally to every response by the `SecurityHeaders` middleware registered before route setup. See the [SecurityHeaders section](#securityheaders-middlewaresecurity_headers.go) for the full header list.

### Password Policy

Enforced at registration by the `strongpassword` custom validator:
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one punctuation or symbol character

### Token Blacklisting

Described in the [Auth section](#token-blacklist). All logouts immediately invalidate the access token. Tokens are stored by SHA256 hash, never by raw value.

---

## 15. Event Bus

### Architecture (`internal/infrastructure/events/`)

**Redis Streams** implementation with consumer groups. The bus is started at application boot and runs consumer goroutines in the background.

```go
type EventBus interface {
    Publish(ctx context.Context, stream string, event Event) error
    Subscribe(stream, group, consumer string, handler EventHandler)
    Start(ctx context.Context) error
    Stop()
}
```

### Event Structure

```go
type Event struct {
    ID            string          // UUID v4, auto-generated if empty
    Type          string          // e.g., "user.created"
    TenantID      uint
    UserID        uint
    CorrelationID string
    OccurredAt    time.Time       // auto-set if zero
    Payload       json.RawMessage // domain-specific data
}
```

### Well-Known Streams

| Constant | Stream name | Purpose |
|----------|-------------|---------|
| `StreamUsers` | `erp:users` | User lifecycle events |
| `StreamAuth` | `erp:auth` | Login/logout/token events |
| `StreamSystem` | `erp:system` | System-level events |
| `StreamDLQ` | `erp:dlq` | Dead-letter queue |

### Well-Known Event Types

```
user.created      user.updated      user.deleted      user.login
auth.login_failed auth.logout       auth.token_refreshed
```

### Retry & DLQ

Each message is attempted up to **3 times** (`maxRetries = 3`). Between attempts, a warning is logged with attempt number and error. After 3 failures, the event is published to `erp:dlq` and the original message is acknowledged (preventing infinite redelivery).

### Consumer Group Configuration

- Poll interval: 100ms
- Idle claim timeout: 30 seconds (for crash recovery — messages idle >30s are re-claimed)
- Consumer groups are auto-created with `XGroupCreateMkStream` when `Subscribe` is called

### Current Subscribers

No production subscribers are registered yet (the event bus is wired and running but nothing subscribes). Future modules (e.g., email notifications, analytics) will subscribe at module `Initialize()` time.

---

## 16. Audit Logging

### Architecture (`internal/infrastructure/audit/`)

`Logger` wraps the `audit.Repository` interface. It provides two methods:

```go
func (l *Logger) Log(ctx context.Context, entry audit.Entry) error
func (l *Logger) LogAsync(ctx context.Context, entry audit.Entry)
```

`LogAsync` spawns a fire-and-forget goroutine — the HTTP response is sent before the DB write. If the write fails, the error is logged but the HTTP request is unaffected.

### What Gets Logged

The `AuditLog` middleware fires on `POST`, `PUT`, `PATCH`, `DELETE` for `2xx` responses only. It captures:

| Field | Source |
|-------|--------|
| `action` | HTTP method → create/update/delete |
| `resource` | First path segment after `/api/v1/` |
| `resource_id` | `:id` route parameter if present |
| `user_id` | `c.Locals("user_id")` (from JWT) |
| `tenant_id` | `c.Locals("tenant_id")` (from TenantContext) |
| `ip_address` | `c.IP()` |
| `user_agent` | `User-Agent` request header |

`old_values` and `new_values` are not automatically captured by the middleware — handlers that need them can call `auditLogger.Log` directly with those fields populated.

### Querying Audit Logs

```
GET /api/v1/audit-logs
  ?resource=users
  &action=delete
  &user_id=5
  &date_from=2026-05-01T00:00:00Z
  &date_to=2026-05-31T23:59:59Z
  &limit=50
  &offset=0
```

---

## 17. Observability & Metrics

### Structured Logging

`pkg/logger/` provides a Zap-based singleton. Zap is initialized once in `main.go` and used everywhere via the package-level functions:

```go
logger.Info("message", logger.String("key", "value"), logger.Err(err))
logger.Error("failed", logger.Uint("user_id", 5), logger.Err(err))
```

Field constructors: `logger.String`, `logger.Int`, `logger.Uint`, `logger.Bool`, `logger.Err`, `logger.Duration`, `logger.Any`

Format: `json` in production, `console` in development (configurable via `LOG_FORMAT` env).

### Health Endpoints

```
GET /health/       — Summary: {"status":"healthy","database":"ok","redis":"ok"}
GET /health/live   — Kubernetes liveness: always 200 if process is up
GET /health/ready  — Kubernetes readiness: 200 if DB + Redis reachable
```

### Prometheus Metrics

**Endpoint:** `GET /metrics/` (OpenMetrics format)

**Go runtime metrics** (from `prometheus.NewGoCollector`):
- `go_goroutines`, `go_memstats_*`, GC stats, etc.

**Process metrics** (from `prometheus.NewProcessCollector`):
- `process_cpu_seconds_total`, `process_resident_memory_bytes`, etc.

**HTTP metrics** (from `middleware.PrometheusMetrics`):

| Metric | Type | Labels |
|--------|------|--------|
| `http_requests_total` | Counter | `method`, `path`, `status` |
| `http_request_duration_seconds` | Histogram | `method`, `path` |
| `http_active_requests` | Gauge | — |

All metrics are registered on a **custom `prometheus.Registry`** (not the default global registry) to avoid conflicts with any imported library that self-registers.

---

## 18. Validation

### Validator Singleton (`pkg/validation/validator.go`)

`validation.Get()` returns a singleton `*validator.Validate` with all custom rules pre-registered.

`validation.ValidateStruct(v, s)` returns `map[string]string` of field name → error tag, or nil if valid.

### Custom Validation Rules

| Tag | Rule | Example |
|-----|------|---------|
| `slug` | Lowercase alphanumeric segments joined by single hyphens: `^[a-z0-9]+(?:-[a-z0-9]+)*$` | `kadahapola-branch-2` |
| `strongpassword` | ≥8 chars, must contain: uppercase + lowercase + digit + special character | `Admin@12345` |
| `phone` | E.164 international format: `^\+[1-9]\d{6,14}$` | `+94771234567` |
| `currency_code` | ISO 4217 3-letter code (subset: AED, AUD, LKR, USD, GBP, EUR, ...) | `LKR` |
| `country_code` | ISO 3166-1 alpha-2 (subset: LK, US, GB, IN, ...) | `LK` |

### How Handlers Use Validation

Each handler has a local `validateStruct` helper (not the pkg-level one — they're per-handler):

```go
func validateStruct(v *validator.Validate, s interface{}) map[string]interface{} {
    err := v.Struct(s)
    if err == nil { return nil }
    fields := make(map[string]interface{})
    for _, e := range err.(validator.ValidationErrors) {
        fields[strings.ToLower(e.Field())] = e.Tag()
    }
    return fields
}
```

On failure:
```go
if fields := validateStruct(h.validate, req); fields != nil {
    return httputil.ValidationError(c, "Validation failed", fields)
    // → HTTP 422 with {"error":{"code":"VALIDATION_ERROR","fields":{...}}}
}
```

---

## 19. Error Handling

### AppError (`pkg/errors/errors.go`)

```go
type AppError struct {
    Code       ErrorCode              // machine-readable string constant
    Message    string                 // human-readable
    Details    string                 // optional extra details
    HTTPStatus int                    // derived from Code
    Err        error                  // wrapped underlying error
    Meta       map[string]interface{} // arbitrary extra data
}
```

### Error Codes → HTTP Status Mapping

| Code | HTTP | Usage |
|------|------|-------|
| `INTERNAL_ERROR` | 500 | Unexpected errors |
| `BAD_REQUEST` | 400 | Malformed input |
| `NOT_FOUND` | 404 | Resource not found |
| `UNAUTHORIZED` | 401 | Missing or expired auth |
| `FORBIDDEN` | 403 | Authenticated but no permission |
| `CONFLICT` | 409 | Duplicate entity |
| `VALIDATION_ERROR` | 422 | Field-level validation failure |
| `TOO_MANY_REQUESTS` | 429 | Rate limit or lockout |
| `INVALID_CREDENTIALS` | 401 | Wrong email/password |
| `TOKEN_EXPIRED` | 401 | JWT past expiry |
| `TOKEN_INVALID` | 401 | Invalid JWT |
| `TOKEN_MISSING` | 401 | No Authorization header |
| `DUPLICATE_ENTRY` | 409 | DB unique constraint |

### httputil Response Functions

```go
httputil.Success(c, message, data)              // 200 JSON
httputil.Created(c, message, data)              // 201 JSON
httputil.BadRequest(c, message)                 // 400 JSON
httputil.Unauthorized(c, message)               // 401 JSON
httputil.Forbidden(c, message)                  // 403 JSON
httputil.NotFound(c, message)                   // 404 JSON
httputil.Conflict(c, message)                   // 409 JSON
httputil.ValidationError(c, message, fields)    // 422 JSON
httputil.InternalServerError(c, message)        // 500 JSON
httputil.Error(c, *AppError)                    // uses AppError.HTTPStatus
```

> **Critical API rule:** All `httputil.*` functions except `httputil.Error` take **plain strings**, NOT `*AppError`. `httputil.Error` is the only function that accepts `*AppError`.

---

## 20. Module System

### Module Interface (`internal/modules/module.go`)

```go
type Module interface {
    Name() string
    Dependencies() []string   // other module names this depends on
    Initialize(deps Dependencies) error
    RegisterRoutes(router fiber.Router)
    RegisterEvents(bus events.EventBus)
    Shutdown(ctx context.Context) error
}

type Dependencies struct {
    DB          *gorm.DB
    Redis       *redis.Client
    Config      *config.Config
    Logger      *zap.Logger
    EventBus    events.EventBus
    Enforcer    rbac.Enforcer
}
```

### Registry (`internal/modules/registry.go`)

The registry resolves initialization order using **Kahn's topological sort** (DFS with cycle detection). If module A depends on module B, B is initialized first. A circular dependency returns an error at startup.

```go
registry := modules.NewRegistry()
registry.Register(inventoryModule)
registry.Register(salesModule)  // depends on inventory
registry.Initialize(deps)       // initializes: inventory → sales

// Shutdown is in reverse order: sales → inventory
registry.Shutdown(ctx)
```

**No modules are registered yet** — the registry and interface are ready for future Master Data and business modules.

---

## 21. Database Migrations

### Running Migrations

```bash
# Apply all pending migrations
docker exec erp-api-dev sh -c "go run cmd/migrate/main.go -command=up"

# Roll back one step
docker exec erp-api-dev sh -c "go run cmd/migrate/main.go -command=down -steps=1"

# Check current version
docker exec erp-api-dev sh -c "go run cmd/migrate/main.go -command=version"

# Force version (fix dirty state)
docker exec erp-api-dev sh -c "go run cmd/migrate/main.go -command=force -version=8"
```

### Adding a New Migration

1. Create two files in `internal/infrastructure/database/migrations/`:
   - `000010_description.up.sql`
   - `000010_description.down.sql`
2. The `.up.sql` must be fully idempotent (use `IF NOT EXISTS`, `IF EXISTS`)
3. The `.down.sql` must exactly reverse the `.up.sql`
4. Never modify an already-applied migration file

### `update_updated_at_column()` Trigger

Defined in migration 001, reused by all tables that have an `updated_at` column:

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';
```

Tables with this trigger: `users`, `tenants`, `roles`, `feature_flags`

---

## 22. Seeding

### Seeder Framework (`internal/infrastructure/database/seeder/seeder.go`)

```go
type Seeder interface {
    Run(ctx context.Context, db *gorm.DB) error
}

func Run(ctx, db, ...seeders) error   // runs each seeder in order, logs name
func Production(ctx, db, adminEmail, adminPassword) error
func Development(ctx, db, adminEmail, adminPassword) error
```

`Production` runs: `SeedDefaultTenant` → `SeedPermissions` → `SeedRoles` → `SeedAdminUser`  
`Development` adds: `SeedDevelopmentData` (test tenant + sample feature flags)

All seed operations use `ON CONFLICT DO NOTHING` — safe to run multiple times.

### Individual Seeders

#### `SeedDefaultTenant`
Inserts the "Kadahapola" enterprise tenant with slug `kadahapola`.

#### `SeedPermissions`
Iterates `rbac.AllPermissions` (19 entries), splits each `resource:action` string, inserts into `permissions` table.

#### `SeedRoles`
Inserts 5 system roles. For each role, assigns permissions:
- `super_admin`: `INSERT INTO role_permissions SELECT ?, id FROM permissions` — all permissions via subquery
- Others: iterates `DefaultRolePermissions[roleName]`, splits each permission key, inserts one by one

#### `SeedAdminUser`
Requires `SEED_ADMIN_EMAIL` and `SEED_ADMIN_PASSWORD` env vars. Creates the admin user with `bcrypt(cost=10)` hash. Also inserts into `user_roles` to assign the `super_admin` role.

#### `SeedDevelopmentData`
Inserts a "test-corp" tenant and 3 sample feature flags for local testing.

### Running the Seeder

```bash
# All (production + development data)
docker exec erp-api-dev sh -c "SEED_ADMIN_EMAIL=admin@kadahapola.com SEED_ADMIN_PASSWORD=Admin@12345 go run cmd/seed/main.go all"

# Production only
docker exec erp-api-dev sh -c "SEED_ADMIN_EMAIL=admin@kadahapola.com SEED_ADMIN_PASSWORD=Admin@12345 go run cmd/seed/main.go prod"

# Development data only
docker exec erp-api-dev sh -c "SEED_ADMIN_EMAIL=admin@kadahapola.com SEED_ADMIN_PASSWORD=Admin@12345 go run cmd/seed/main.go dev"
```

---

## 23. CLI Tools

### `cmd/api/main.go` — API Server

Initialization sequence:

1. `config.Load()` — load Viper config
2. `logger.Initialize()` — start Zap
3. `postgres.Connect(cfg)` + `HealthCheck`
4. `rediscache.Connect(cfg)` + `HealthCheck`
5. `jwtpkg.NewManager` — JWT config
6. `rediscache.NewTokenBlacklist` — logout blacklist
7. `userrepo.NewPostgresRepository`
8. `security.NewAccountLockout` — Redis lockout
9. `authhandler.NewHandler(userRepo, jwtMgr, blacklist, lockout, expiry)`
10. `casbininfra.New(db)` — enforcer with policy load
11. `rolerepo`, `permrepo`, `tenantrepo`, `auditrepo` — all repos
12. `auditinfra.NewLogger(auditRepo)` — audit logger
13. `events.NewRedisStreamBus(redisClient)` + `Start(ctx)`
14. `featureflags.NewDBStore(db, redisClient)` — ff store
15. All HTTP handlers
16. `server.New(cfg, logger)` — Fiber server
17. Global middleware: RequestID, CORS (if enabled), Logger
18. Health + Metrics routes
19. Prometheus metrics middleware
20. `srv.SetupRoutes(...)` — all API routes with middleware
21. `srv.StartWithGracefulShutdown()` — listen + SIGTERM handler

### `cmd/migrate/main.go` — Migration CLI

```
Flags:
  -command   up | down | version | force   (default: up)
  -steps     N                             (for up/down N steps)
  -version   N                             (for force command)
```

Connects to PostgreSQL using the same DSN as the main app. Migration files are read from `internal/infrastructure/database/migrations/`.

### `cmd/seed/main.go` — Seed CLI

```
Usage: go run cmd/seed/main.go [all|prod|dev]

Environment variables:
  SEED_ADMIN_EMAIL     admin email address (defaults to admin@kadahapola.com with warning)
  SEED_ADMIN_PASSWORD  admin password      (defaults to Admin@12345 with warning)
```

---

## 24. Known Issues & Decisions

### Casbin gorm-adapter Incompatibility

**Problem:** `gorm-adapter/v3 v3.32+` internally uses `casbin/v3/persist.Adapter`. Passing it to `casbin/v2.NewEnforcer` causes a runtime panic at the type assertion.

**Decision:** Remove `gorm-adapter` entirely. Implement `persist.Adapter` directly in `casbin/enforcer.go` using two raw SQL queries on our own `role_permissions` and `user_roles` tables.

**Impact:** Casbin's `SavePolicy`, `AddPolicy`, `RemovePolicy`, `RemoveFilteredPolicy` methods on the adapter are no-ops. Actual permission changes are persisted through the role HTTP handler which writes to the DB directly. Calling `enforcer.LoadPolicy()` will reload from DB.

### Air Hot-Reload Version Pin

`Dockerfile.dev` pins Air at `v1.61.5`. Do not upgrade to v1.65.3 — it requires Go 1.25. The version will be unpinned when Go is upgraded to 1.25+.

### `users.role` Column (Legacy)

The `users` table has a `role VARCHAR(50)` column that was added in Batch 1 before RBAC was designed. It is still populated (set to `"super_admin"` for the admin user, `"user"` for regular users). The actual RBAC check uses the Casbin policy loaded from `user_roles` → `roles`, not this column. The column is used only when generating the JWT `role` claim. This is a known inconsistency that will be resolved when user management is refactored.

### Tenant ID in JWT

The JWT claims currently contain `user_id`, `email`, and `role`. `tenant_id` is NOT embedded in the token. The `TenantContext` middleware therefore sets `tenant_id` in locals only for `super_admin` via the `X-Tenant-ID` header. Regular users' tenant context is not automatically injected.

This is a planned enhancement: embed `tenant_id` in JWT during login (requires the user to be associated with a tenant at login time).

### Feature Flag Percentage Rollout

`FlagRules.Percentage` is stored but not evaluated. `IsEnabledForUser` only checks `UserIDs`. The percentage sampling logic (deterministic hash-based bucketing) is to be implemented when gradual rollouts are needed.

### Audit `old_values` / `new_values`

The audit middleware captures `action`, `resource`, `resource_id`, `user_id`, `tenant_id`, `ip_address`, `user_agent`. It does NOT capture `old_values` or `new_values` automatically — that requires handlers to call `auditLogger.Log` explicitly with those fields, which no handler currently does.

### No Swagger/OpenAPI Docs

Swagger documentation was in the original Batch 5 plan but was deprioritized in favor of the Postman collection (`docs/postman/kadahapola_erp_api.json`). Swag annotations on handlers and `make docs` generation remain as a future task.

### No Integration Tests

The `tests/` directory exists but contains no Go test files yet. Integration test infrastructure (`tests/setup/`, `tests/mocks/`) is planned for a future batch.

---

## Appendix: Redis Key Namespaces

| Key Pattern | Package | Purpose |
|------------|---------|---------|
| `blacklist:{sha256}` | `cache/redis` | JWT token blacklist |
| `ratelimit:ip:{ip}` | `middleware` | Per-IP rate limit counter |
| `ratelimit:user:{id}` | `middleware` | Per-user rate limit counter |
| `lockout:{email}` | `security` | Account lockout state |
| `ff:{tenantID}:{name}` | `featureflags` | Feature flag cache |
| `ff:global:{name}` | `featureflags` | Global feature flag cache |
| `erp:users` | `events` | Redis Stream — user events |
| `erp:auth` | `events` | Redis Stream — auth events |
| `erp:system` | `events` | Redis Stream — system events |
| `erp:dlq` | `events` | Redis Stream — dead-letter queue |

## Appendix: Import Alias Conventions

```go
import (
    apperrors  "erp-system/pkg/errors"
    httputil   "erp-system/pkg/http"
    jwtpkg     "erp-system/pkg/jwt"
    domainuser "erp-system/internal/domain/user"
    domainrole "erp-system/internal/domain/role"
    domainperm "erp-system/internal/domain/permission"
    rediscache "erp-system/internal/infrastructure/cache/redis"
    casbininfra "erp-system/internal/infrastructure/authorization/casbin"
    auditinfra  "erp-system/internal/infrastructure/audit"
    ffhandler   "erp-system/internal/interfaces/http/handlers/featureflags"
)
```
