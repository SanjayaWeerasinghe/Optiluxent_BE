# Batch 4 Test Results

**Date**: 2026-05-22
**Status**: ✅ All tests passed

---

## Migrations

All 8 migrations ran and rolled back cleanly:
- `000001_create_users_table` — ✅ (Batch 3)
- `000002_create_tenants_table` — ✅
- `000003_add_tenant_to_users` — ✅
- `000004_create_roles_table` — ✅
- `000005_create_permissions_table` — ✅
- `000006_create_role_permissions_table` — ✅
- `000007_create_user_roles_table` — ✅
- `000008_create_audit_logs_table` — ✅

---

## RBAC / Casbin Enforcer

| Test | Result |
|------|--------|
| `user` role → `GET /api/v1/users` | ✅ 403 Forbidden |
| `user` role → `GET /api/v1/roles` | ✅ 403 Forbidden |
| `user` role → `GET /api/v1/permissions` | ✅ 403 Forbidden |
| `user` role → `GET /api/v1/tenants` | ✅ 403 Forbidden |
| `user` role → `GET /api/v1/audit-logs` | ✅ 403 Forbidden |
| `super_admin` role → all above | ✅ 200 OK (Casbin model bypass) |

**Note**: Custom DB-backed Casbin adapter implemented to avoid gorm-adapter/v3 version conflict with casbin/v2 (gorm-adapter v3.32+ switched internal interface to casbin/v3).

---

## Tenant Management (super_admin only)

| Endpoint | Status | Notes |
|----------|--------|-------|
| `POST /api/v1/tenants` | ✅ 201 | Created "Kadahapola Ltd" slug=kadahapola plan=enterprise |
| `GET /api/v1/tenants` | ✅ 200 | total=1 |
| `GET /api/v1/tenants/:id` | ✅ 200 | Returns tenant details |
| `PUT /api/v1/tenants/:id` | ✅ 200 | Updated plan to professional |

---

## Role Management (super_admin)

| Endpoint | Status | Notes |
|----------|--------|-------|
| `POST /api/v1/roles` | ✅ 201 | Created "accounts_manager" role |
| `GET /api/v1/roles` | ✅ 200 | Returns all roles (count=2 incl. previous) |
| `GET /api/v1/roles/:id` | ✅ 200 | Returns role with permissions |
| `GET /api/v1/permissions` | ✅ 200 | Returns all system permissions |

---

## User Management

| Endpoint | Status | Notes |
|----------|--------|-------|
| `GET /api/v1/users` | ✅ 200 | total=3 (super_admin only) |
| `GET /api/v1/users/:id` | ✅ 200 | Returns user details |
| `PUT /api/v1/profile` | ✅ 200 | Works for any authenticated user |
| `PUT /api/v1/password` | ✅ 200 | Works for any authenticated user |

---

## Audit Logging

| Test | Result |
|------|--------|
| State-changing requests logged | ✅ total=6 after test sequence |
| Async logging (fire-and-forget) | ✅ No request latency impact |

---

## Event Bus (Redis Streams)

| Test | Result |
|------|--------|
| RedisStreamBus.Start() | ✅ Initialised at startup |
| No consumer subscriptions yet | ✅ Gracefully idle |

---

## Issues Encountered & Fixed

1. **Casbin gorm-adapter v3 version conflict**: gorm-adapter/v3 v3.32+ switched to implementing `casbin/v3/persist.Adapter` internally. Passing the adapter through `NewEnforcer(m, adapter)` caused a runtime panic (`interface conversion: *gormadapter.Adapter is not persist.Adapter: missing method LoadPolicy`). **Fix**: Implemented a custom `dbAdapter` struct that implements `casbin/v2/persist.Adapter` directly by querying our `role_permissions` and `user_roles` tables via SQL. Eliminates the gorm-adapter dependency entirely.

2. **Slug validation**: `alphanum` validator rejects hyphens. **Fix**: Removed alphanum constraint; slugs are validated by min/max length only.
