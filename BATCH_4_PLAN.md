# Batch 4: Advanced Infrastructure - Implementation Plan

**Phase**: Batch 4 - Advanced Infrastructure
**Depends On**: Batch 3 (Authentication & Authorization) ✅ must be complete
**Developer**: Sanjaya Weerasinghe
**Status**: ✅ Complete (2026-05-22)
**Roadmap Sections**: 0.10 – 0.19

---

## Overview

Extend the authenticated foundation with a full authorization layer (RBAC/Casbin), additional domain models and migrations, complete user/role/permission management APIs, multi-tenancy, audit logging, an event bus (Redis Streams), and a module registry. After this batch the system can enforce permissions, isolate tenants, record all actions, and publish/consume cross-module events.

---

## Components to Build

### 1. RBAC System — Casbin (20%)
**Location**: `internal/infrastructure/authorization/`

**Files to Create**:
- `casbin/enforcer.go` — Initialize Casbin enforcer with GORM adapter
- `casbin/model.go` — RBAC model definition (conf string)
- `casbin/policy_loader.go` — Load policies from PostgreSQL at startup

**Domain Layer**:
- `internal/domain/rbac/roles.go` — Role constants and default role definitions
- `internal/domain/rbac/permissions.go` — Permission constants (resource:action format)
- `internal/domain/rbac/enforcer_interface.go` — Enforcer interface (for mocking in tests)

**Roles to Define**:
- `super_admin` — Full system access, cross-tenant
- `admin` — Full access within tenant
- `manager` — Module management access within tenant
- `user` — Standard user access
- `guest` — Read-only access

**Permission Format**:
```
resource:action
Examples:
  users:create    users:read    users:update    users:delete
  roles:create    roles:read    roles:update    roles:delete
  modules:read    modules:manage
  audit:read
  tenants:create  tenants:read  tenants:update
```

**Features**:
- Role hierarchy support (super_admin > admin > manager > user > guest)
- Resource-level permission checking
- Permission caching in Redis (5 min TTL)
- Policy sync on role/permission changes

---

### 2. Core Domain Models (10%)
**Location**: `internal/domain/`

**Files to Create**:
- `internal/domain/role/entity.go` — Role entity (ID, name, description, permissions)
- `internal/domain/role/repository.go` — Role repository interface
- `internal/domain/permission/entity.go` — Permission entity (ID, resource, action, description)
- `internal/domain/permission/repository.go` — Permission repository interface
- `internal/domain/tenant/entity.go` — Tenant entity (ID, name, slug, plan, status, config JSONB)
- `internal/domain/tenant/repository.go` — Tenant repository interface
- `internal/domain/audit/entity.go` — AuditLog entity (ID, tenant_id, user_id, action, resource, resource_id, old_values, new_values, ip, user_agent, created_at)
- `internal/domain/audit/repository.go` — AuditLog repository interface

**BaseEntity** (`internal/domain/base_entity.go`):
```go
type BaseEntity struct {
    ID        uint           `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

---

### 3. Database Migrations (10%)
**Location**: `internal/infrastructure/database/migrations/`

**Migration Files to Create** (in order):

| File | Creates |
|------|---------|
| `000002_create_tenants_table.up.sql` | tenants (id, name, slug, plan, status, config JSONB, created_at, updated_at, deleted_at) |
| `000002_create_tenants_table.down.sql` | rollback |
| `000003_add_tenant_to_users.up.sql` | ALTER users ADD COLUMN tenant_id + index |
| `000003_add_tenant_to_users.down.sql` | rollback |
| `000004_create_roles_table.up.sql` | roles (id, tenant_id, name, description, is_system, created_at, updated_at) |
| `000004_create_roles_table.down.sql` | rollback |
| `000005_create_permissions_table.up.sql` | permissions (id, resource, action, description) |
| `000005_create_permissions_table.down.sql` | rollback |
| `000006_create_role_permissions_table.up.sql` | role_permissions (role_id, permission_id, PRIMARY KEY composite) |
| `000006_create_role_permissions_table.down.sql` | rollback |
| `000007_create_user_roles_table.up.sql` | user_roles (user_id, role_id, tenant_id, PRIMARY KEY composite) |
| `000007_create_user_roles_table.down.sql` | rollback |
| `000008_create_audit_logs_table.up.sql` | audit_logs (id, tenant_id, user_id, action, resource, resource_id, old_values JSONB, new_values JSONB, ip_address, user_agent, created_at) — partitioned by created_at |
| `000008_create_audit_logs_table.down.sql` | rollback |

**Also create**:
- `cmd/migrate/main.go` — Migration CLI (commands: up, down, version, force)

---

### 4. Repository Implementations (10%)
**Location**: `internal/infrastructure/persistence/`

**Files to Create**:
- `role/postgres_repository.go` — GORM implementation of role.Repository
- `permission/postgres_repository.go` — GORM implementation of permission.Repository
- `tenant/postgres_repository.go` — GORM implementation of tenant.Repository
- `audit/postgres_repository.go` — GORM implementation of audit.Repository (append-only, no updates)

**Role Repository Interface Methods**:
```go
Create(ctx, role) error
GetByID(ctx, id uint) (*Role, error)
GetByName(ctx, tenantID uint, name string) (*Role, error)
List(ctx, tenantID uint) ([]*Role, error)
Update(ctx, role) error
Delete(ctx, id uint) error
AssignPermissions(ctx, roleID uint, permissionIDs []uint) error
GetPermissions(ctx, roleID uint) ([]*Permission, error)
```

**Tenant Repository Interface Methods**:
```go
Create(ctx, tenant) error
GetByID(ctx, id uint) (*Tenant, error)
GetBySlug(ctx, slug string) (*Tenant, error)
List(ctx, limit, offset int) ([]*Tenant, int64, error)
Update(ctx, tenant) error
Delete(ctx, id uint) error
```

---

### 5. Permission Middleware (5%)
**Location**: `internal/infrastructure/http/middleware/`

**Files to Create**:
- `permission.go` — Casbin-based permission enforcement middleware

**Usage**:
```go
// In routes.go:
users := v1.Group("/users", middleware.Auth(jwtManager), middleware.RequirePermission(enforcer, "users:read"))
```

**Features**:
- Extract user role from JWT claims
- Check permission via Casbin enforcer
- Return 403 Forbidden with structured error on denial
- Cache permission decisions in Redis

---

### 6. User Management APIs (10%)
**Location**: `internal/interfaces/http/handlers/user/`

**Files to Create**:
- `handler.go` — User handler (uses user repo + auth middleware)
- `dto.go` — Request/response DTOs

**Endpoints**:
```
GET    /api/v1/users                — List users (admin, paginated, filterable)
GET    /api/v1/users/:id            — Get user by ID (admin)
PUT    /api/v1/users/:id            — Update user (admin)
DELETE /api/v1/users/:id            — Soft-delete user (admin)
PUT    /api/v1/users/profile        — Update own profile (authenticated)
PUT    /api/v1/users/password       — Change own password (authenticated)
POST   /api/v1/users/:id/roles      — Assign role to user (admin)
DELETE /api/v1/users/:id/roles/:role_id — Remove role from user (admin)
```

**DTOs**:
```go
type UpdateUserRequest struct {
    FirstName string `json:"first_name" validate:"omitempty,min=1,max=100"`
    LastName  string `json:"last_name"  validate:"omitempty,min=1,max=100"`
    Status    string `json:"status"     validate:"omitempty,oneof=active inactive"`
}

type ChangePasswordRequest struct {
    CurrentPassword string `json:"current_password" validate:"required"`
    NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

type AssignRoleRequest struct {
    RoleID uint `json:"role_id" validate:"required"`
}
```

---

### 7. Role & Permission Management APIs (10%)
**Location**: `internal/interfaces/http/handlers/role/`

**Files to Create**:
- `handler.go` — Role handler
- `dto.go` — Request/response DTOs

**Endpoints**:
```
GET    /api/v1/roles                — List roles (admin)
POST   /api/v1/roles                — Create role (admin)
GET    /api/v1/roles/:id            — Get role with permissions (admin)
PUT    /api/v1/roles/:id            — Update role (admin)
DELETE /api/v1/roles/:id            — Delete role (admin, non-system only)
POST   /api/v1/roles/:id/permissions — Assign permissions to role (admin)
DELETE /api/v1/roles/:id/permissions/:perm_id — Remove permission from role (admin)
GET    /api/v1/permissions          — List all permissions (admin)
```

**DTOs**:
```go
type CreateRoleRequest struct {
    Name        string `json:"name"        validate:"required,min=2,max=50"`
    Description string `json:"description" validate:"omitempty,max=255"`
}

type AssignPermissionsRequest struct {
    PermissionIDs []uint `json:"permission_ids" validate:"required,min=1"`
}
```

---

### 8. Multi-Tenancy System (10%)
**Location**: `internal/infrastructure/http/middleware/` and `internal/infrastructure/persistence/`

**Files to Create**:
- `internal/infrastructure/http/middleware/tenant.go` — Tenant context middleware
- `internal/infrastructure/persistence/tenant_scope.go` — GORM global scope for tenant isolation

**How It Works**:
1. Tenant middleware extracts `tenant_id` from JWT claims
2. Sets tenant in request context (`ctx.Locals("tenant_id", id)`)
3. GORM global scope automatically adds `WHERE tenant_id = ?` to all queries
4. Super admin can switch tenant via `X-Tenant-ID` header

**Tenant API Endpoints**:
```
POST   /api/v1/tenants              — Create tenant (super_admin)
GET    /api/v1/tenants              — List tenants (super_admin)
GET    /api/v1/tenants/:id          — Get tenant (super_admin)
PUT    /api/v1/tenants/:id          — Update tenant (super_admin)
```

---

### 9. Audit Logging System (5%)
**Location**: `internal/infrastructure/http/middleware/` and `internal/infrastructure/audit/`

**Files to Create**:
- `internal/infrastructure/audit/logger.go` — AuditLogger service
- `internal/infrastructure/http/middleware/audit.go` — HTTP audit middleware

**What Gets Logged**:
- All POST/PUT/PATCH/DELETE requests (resource, resource_id, old/new values, user, IP, user agent)
- Auth events: login (success/fail), logout, token refresh
- Authorization failures (403 responses)
- Password changes

**Audit API Endpoints**:
```
GET    /api/v1/audit-logs           — List audit logs (admin, paginated, filterable by resource/user/date)
```

**AuditLogger Interface**:
```go
type AuditLogger interface {
    Log(ctx context.Context, entry AuditEntry) error
    LogAsync(ctx context.Context, entry AuditEntry)  // fire-and-forget via goroutine
}
```

---

### 10. Event System — Redis Streams (8%)
**Location**: `internal/infrastructure/events/`

**Files to Create**:
- `event.go` — Event interface and base Event struct
- `bus.go` — EventBus interface
- `redis_stream_bus.go` — Redis Streams implementation
- `publisher.go` — Publisher helper
- `consumer.go` — Consumer with handler registry
- `handler_registry.go` — Map of event type → handler functions

**Event Structure**:
```go
type Event struct {
    ID            string          `json:"id"`             // UUID
    Type          string          `json:"type"`           // e.g. "user.created"
    TenantID      uint            `json:"tenant_id"`
    UserID        uint            `json:"user_id"`
    CorrelationID string          `json:"correlation_id"`
    OccurredAt    time.Time       `json:"occurred_at"`
    Payload       json.RawMessage `json:"payload"`
}

type EventBus interface {
    Publish(ctx context.Context, stream string, event Event) error
    Subscribe(stream string, group string, handler EventHandler) error
    Start(ctx context.Context) error
    Stop() error
}
```

**Streams to Set Up**:
- `erp:users` — user.created, user.updated, user.deleted, user.login
- `erp:auth` — auth.login_failed, auth.logout, auth.token_refreshed
- `erp:system` — system.module_loaded, system.module_unloaded

**Features**:
- Consumer groups for reliable delivery
- Dead letter queue stream (`erp:dlq`) for failed events
- Event replay via stream IDs
- Automatic retry on handler failure (max 3 retries)

---

### 11. Module Registry System (2%)
**Location**: `internal/infrastructure/modules/`

**Files to Create**:
- `module.go` — Module interface definition
- `registry.go` — ModuleRegistry with dependency resolution
- `loader.go` — Config-driven module loader

**Module Interface**:
```go
type Module interface {
    Name() string
    Dependencies() []string
    Initialize(deps Dependencies) error
    RegisterRoutes(router fiber.Router)
    RegisterEvents(bus events.EventBus)
    Migrate(db *gorm.DB) error
    Shutdown(ctx context.Context) error
}
```

**Registry Features**:
- Topological sort for dependency ordering
- Module enable/disable via `config/modules.yaml`
- Module status endpoint:
  ```
  GET /api/v1/system/modules — List loaded modules and status (super_admin)
  ```

---

## File Structure

```
internal/
├── domain/
│   ├── base_entity.go                              ← NEW
│   ├── role/
│   │   ├── entity.go                               ← NEW
│   │   └── repository.go                           ← NEW
│   ├── permission/
│   │   ├── entity.go                               ← NEW
│   │   └── repository.go                           ← NEW
│   ├── tenant/
│   │   ├── entity.go                               ← NEW
│   │   └── repository.go                           ← NEW
│   ├── audit/
│   │   ├── entity.go                               ← NEW
│   │   └── repository.go                           ← NEW
│   └── rbac/
│       ├── roles.go                                ← NEW
│       ├── permissions.go                          ← NEW
│       └── enforcer_interface.go                   ← NEW
├── infrastructure/
│   ├── authorization/
│   │   └── casbin/
│   │       ├── enforcer.go                         ← NEW
│   │       ├── model.go                            ← NEW
│   │       └── policy_loader.go                    ← NEW
│   ├── persistence/
│   │   ├── role/
│   │   │   └── postgres_repository.go              ← NEW
│   │   ├── permission/
│   │   │   └── postgres_repository.go              ← NEW
│   │   ├── tenant/
│   │   │   └── postgres_repository.go              ← NEW
│   │   ├── audit/
│   │   │   └── postgres_repository.go              ← NEW
│   │   └── tenant_scope.go                         ← NEW
│   ├── database/
│   │   └── migrations/
│   │       ├── 000002_create_tenants_table.up.sql  ← NEW
│   │       ├── 000002_create_tenants_table.down.sql← NEW
│   │       ├── 000003_add_tenant_to_users.up.sql   ← NEW
│   │       ├── 000003_add_tenant_to_users.down.sql ← NEW
│   │       ├── 000004_create_roles_table.up.sql    ← NEW
│   │       ├── 000004_create_roles_table.down.sql  ← NEW
│   │       ├── 000005_create_permissions_table.up.sql ← NEW
│   │       ├── 000005_create_permissions_table.down.sql ← NEW
│   │       ├── 000006_create_role_permissions_table.up.sql ← NEW
│   │       ├── 000006_create_role_permissions_table.down.sql ← NEW
│   │       ├── 000007_create_user_roles_table.up.sql ← NEW
│   │       ├── 000007_create_user_roles_table.down.sql ← NEW
│   │       ├── 000008_create_audit_logs_table.up.sql ← NEW
│   │       └── 000008_create_audit_logs_table.down.sql ← NEW
│   ├── http/
│   │   └── middleware/
│   │       ├── permission.go                       ← NEW
│   │       ├── tenant.go                           ← NEW
│   │       └── audit.go                            ← NEW
│   ├── audit/
│   │   └── logger.go                               ← NEW
│   └── events/
│       ├── event.go                                ← NEW
│       ├── bus.go                                  ← NEW
│       ├── redis_stream_bus.go                     ← NEW
│       ├── publisher.go                            ← NEW
│       ├── consumer.go                             ← NEW
│       └── handler_registry.go                     ← NEW
├── modules/
│   ├── module.go                                   ← NEW
│   ├── registry.go                                 ← NEW
│   └── loader.go                                   ← NEW
└── interfaces/
    └── http/
        └── handlers/
            ├── user/
            │   ├── handler.go                      ← NEW
            │   └── dto.go                          ← NEW
            ├── role/
            │   ├── handler.go                      ← NEW
            │   └── dto.go                          ← NEW
            └── tenant/
                ├── handler.go                      ← NEW
                └── dto.go                          ← NEW
cmd/
└── migrate/
    └── main.go                                     ← NEW
config/
└── modules.yaml                                    ← NEW
```

---

## Implementation Order

### Phase 1: RBAC Foundation (Steps 1–4)
1. Create domain models: Role, Permission, Tenant, AuditLog entities + repository interfaces
2. Write all database migrations (000002–000008), test up/down
3. Implement PostgreSQL repositories for Role, Permission, Tenant, Audit
4. Initialize Casbin enforcer with GORM adapter + RBAC model + policy loader

### Phase 2: APIs (Steps 5–8)
5. Build User Management handlers (list, get, update, delete, profile, password, assign role)
6. Build Role Management handlers (CRUD + permission assignment)
7. Build Tenant handlers (CRUD, super_admin only)
8. Wire permission middleware (`RequirePermission`) onto all new routes

### Phase 3: Multi-Tenancy & Audit (Steps 9–10)
9. Implement tenant context middleware + GORM global scope for tenant isolation
10. Implement AuditLogger + audit HTTP middleware; add to all state-changing routes

### Phase 4: Event System (Steps 11–12)
11. Implement Redis Streams EventBus (publisher + consumer + DLQ)
12. Publish events from auth handlers (user.created, user.login, auth.login_failed)

### Phase 5: Module Registry (Step 13)
13. Implement Module interface, ModuleRegistry with topological sort, modules.yaml loader, status endpoint

### Phase 6: Integration (Step 14)
14. Update `main.go`: initialize Casbin enforcer, tenant middleware, audit logger, event bus, module registry; run all new migrations; wire all routes

### Phase 7: Testing (Step 15)
15. Test all new endpoints with Postman; test RBAC with different roles; test tenant isolation; test audit logs; document results in `test_results/batch_4/BATCH_4_TEST_RESULTS.md`

---

## Dependencies to Add

```bash
# Casbin RBAC
github.com/casbin/casbin/v2
github.com/casbin/gorm-adapter/v3

# UUID for event IDs
github.com/google/uuid

# No new HTTP framework dependencies needed - Fiber already installed
```

---

## Security Considerations

1. **Permission Caching**: Cache Casbin decisions in Redis with 5-min TTL; invalidate on role/permission change
2. **Tenant Isolation**: GORM global scope ensures no cross-tenant data leakage; super_admin bypass is explicit and logged
3. **Audit Log Immutability**: No UPDATE or DELETE on audit_logs table; append-only; partition by month
4. **Event Payload Sanitization**: Never include password hashes or raw secrets in event payloads

---

## Testing Plan

After implementation, test:
1. Register → Login → access protected endpoint with correct/wrong role
2. Assign permission to role → verify access granted → remove permission → verify 403
3. Create two tenants → create users in each → verify no cross-tenant data leakage
4. Login → check audit log records the event → update user → check old/new values logged
5. Publish event from auth handler → verify consumer receives it → kill consumer → verify DLQ receives it
6. Load module registry with a test module → verify lifecycle methods called in order
7. `cmd/migrate/main.go up` → verify all 8 migrations run; `down` → verify clean rollback
8. Test user management CRUD endpoints as admin vs. non-admin
9. Test role management CRUD and permission assignment

---

## Success Criteria

- [ ] Casbin enforcer loaded and enforcing permissions on all routes
- [ ] Migrations 000002–000008 run and roll back cleanly
- [ ] Users can be listed, updated, deleted by admin; own profile updated by self
- [ ] Roles can be created, assigned permissions, assigned to users
- [ ] Two tenants are fully isolated — no cross-tenant data visible
- [ ] Every state-changing request creates an audit log entry
- [ ] Auth events (login, logout, failed login) appear in audit log
- [ ] Redis Streams event bus publishes and consumes events
- [ ] Failed event handling routes to DLQ after 3 retries
- [ ] Module registry loads modules in dependency order
- [ ] `GET /api/v1/system/modules` returns correct status
- [ ] Migration CLI (`cmd/migrate`) works for up/down/version
- [ ] All new endpoints documented in routes.go with permission annotations
- [ ] Test results documented in `test_results/batch_4/`

---

## Integration with Previous Batches

Batch 4 uses:
- ✅ Batch 1: Config (JWT secret, Redis), Logger (audit events), PostgreSQL, Redis
- ✅ Batch 2: HTTP server, middleware stack, response helpers, health routes
- ✅ Batch 3: JWT token (extracts user_id, role, tenant_id), auth middleware, user entity, bcrypt

---

## Estimated Time: 8–10 hours

---

**Author**: Sanjaya Weerasinghe
**Date**: 2026-05-22
**Next**: Start with Phase 1 — domain models and migrations
