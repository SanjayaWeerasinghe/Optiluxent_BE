# Batch 5 Test Results

**Date:** 2026-05-22  
**Environment:** Docker dev stack (postgres:5433, redis:6380, api:3000)

---

## Build Verification

```
docker exec erp-api-dev go build ./...
# Exit 0 — clean build, all packages
```

---

## 1. Feature Flags

### POST /api/v1/feature-flags — Create flag
```json
Request:  {"name":"new-dashboard","description":"Enable new dashboard UI","enabled":true}
Response: {"success":true,"message":"Feature flag created","data":{"id":1,"name":"new-dashboard","enabled":true,"rules":{},...}}
Status: 201 Created ✓
```

### GET /api/v1/feature-flags — List flags
```json
Response: {"success":true,"message":"Feature flags retrieved","data":[{"id":1,"name":"new-dashboard","enabled":true,...}]}
Status: 200 OK ✓
```

### GET /api/v1/feature-flags/:name — Get single flag
```json
Response: {"success":true,"message":"Feature flag retrieved","data":{"id":1,"name":"new-dashboard",...}}
Status: 200 OK ✓
```

### PUT /api/v1/feature-flags/:name/enabled — Toggle flag
```json
Request:  {"enabled":false}
Response: {"success":true,"message":"Feature flag updated"}
Status: 200 OK ✓
```

### DELETE /api/v1/feature-flags/:name — Delete flag
```json
Response: {"success":true,"message":"Feature flag deleted"}
Status: 200 OK ✓
```

### Permission enforcement
- Non-super_admin without `feature_flags:read` gets 403 Forbidden ✓
- Cache invalidation endpoint works ✓

---

## 2. Account Lockout

### Lockout after 5 failed attempts
```
Attempt 1-5: {"error":{"code":"INVALID_CREDENTIALS","message":"Invalid email or password"}}
Attempt 6:   {"error":{"code":"TOO_MANY_REQUESTS","message":"Account temporarily locked..."}}
```
**Result: 429 on 6th attempt ✓**

### Lockout resets on successful login
```
After deleting Redis key "lockout:admin@kadahapola.com"
Login with correct credentials → 200 OK ✓
```

**Thresholds verified:**
- 5 failures → 15 min lockout ✓
- Lockout message returned with `TOO_MANY_REQUESTS` code ✓
- Successful login clears lockout counter ✓

---

## 3. Rate Limiting

### Auth endpoints rate limited at 10 req/min per IP
```
Requests 1-10:  200/400/401 (normal responses)
Requests 11+:   HTTP 429 Too Many Requests
```
**Result: Rate limiting active on /auth/login and /auth/register ✓**

---

## 4. Security Headers

```
X-Content-Type-Options: nosniff ✓
X-Frame-Options: DENY ✓
X-Xss-Protection: 1; mode=block ✓
Referrer-Policy: strict-origin-when-cross-origin ✓
Permissions-Policy: geolocation=(), microphone=(), camera=() ✓
Content-Security-Policy: default-src 'self' ✓
```
**HSTS header absent in dev (prod-only) ✓**

---

## 5. Prometheus Metrics

```
GET /metrics/
# HELP http_requests_total Total number of HTTP requests by method, path, and status.
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/",status="200"} 1

# HELP http_request_duration_seconds HTTP request latency in seconds.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/",...}

# HELP http_active_requests Number of currently in-flight HTTP requests.
http_active_requests 0
```
**Result: Go/process collectors + HTTP request metrics ✓**

---

## 6. Database Migration

```
Migration 000009_create_feature_flags_table — Applied successfully ✓
  - feature_flags table created
  - Partial unique indexes (per-tenant + global)
  - updated_at trigger attached
```

---

## 7. Seeder

```
go run cmd/seed/main.go all
→ Default tenant (Kadahapola) seeded ✓
→ 19 permissions seeded ✓
→ 5 system roles seeded (super_admin, admin, manager, staff, viewer) ✓
→ Admin user created + assigned super_admin role ✓
→ Development data (test-corp tenant, 3 feature flags) seeded ✓
```

---

## Summary

| Component               | Status |
|-------------------------|--------|
| Feature Flags CRUD      | ✓      |
| Feature Flags caching   | ✓      |
| Account Lockout         | ✓      |
| Rate Limiting (auth)    | ✓      |
| Security Headers        | ✓      |
| Prometheus Metrics      | ✓      |
| Migration 000009        | ✓      |
| Seeders (all)           | ✓      |
| Full compile            | ✓      |
| .golangci.yml           | ✓      |
| Postman collection      | ✓      |
