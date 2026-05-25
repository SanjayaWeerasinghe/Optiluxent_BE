# Batch 2: HTTP Server Layer - Implementation Plan

**Phase**: Batch 2 - HTTP Server Layer
**Start Date**: October 31, 2025
**Developer**: Sanjaya Weerasinghe
**Status**: 🟡 In Progress

---

## Overview

Build a production-ready HTTP server layer with Fiber framework, comprehensive middleware stack, request/response utilities, and observability endpoints.

---

## Components to Build

### 1. HTTP Server Foundation (30%)
**Location**: `internal/infrastructure/http/server/`

**Files to Create**:
- `server.go` - Fiber server initialization and configuration
- `routes.go` - Route registration and grouping
- `shutdown.go` - Graceful shutdown handler

**Features**:
- ✅ Fiber v2 server setup
- ✅ Configurable host/port from config
- ✅ Graceful shutdown with context
- ✅ Request timeout configuration
- ✅ Body size limits
- ✅ JSON serialization settings

---

### 2. Middleware Stack (30%)
**Location**: `internal/infrastructure/http/middleware/`

**Files to Create**:
- `cors.go` - CORS middleware
- `logger.go` - HTTP request logger
- `recovery.go` - Panic recovery
- `request_id.go` - Request ID tracking
- `rate_limit.go` - Rate limiting
- `timeout.go` - Request timeout
- `compress.go` - Response compression

**Features**:
- ✅ CORS with configurable origins
- ✅ Request/response logging with Zap
- ✅ Panic recovery with stack traces
- ✅ Request ID generation and tracking
- ✅ Rate limiting (per IP, per route)
- ✅ Request timeout handling
- ✅ Gzip/Brotli compression

---

### 3. Request/Response Utilities (15%)
**Location**: `pkg/http/`

**Files to Create**:
- `response.go` - Response helpers
- `request.go` - Request parsing helpers
- `pagination.go` - Pagination utilities
- `binding.go` - Request binding

**Features**:
- ✅ Success/Error response formatters
- ✅ JSON response helpers
- ✅ Pagination (offset/limit, cursor-based)
- ✅ Request body binding with validation
- ✅ Query parameter parsing
- ✅ Header helpers

---

### 4. Validation System (10%)
**Location**: `internal/infrastructure/http/middleware/`

**Files to Create**:
- `validator.go` - Validation middleware
- `custom_validators.go` - Custom validation rules

**Features**:
- ✅ Integration with go-playground/validator
- ✅ Custom validation rules
- ✅ Field-level error messages
- ✅ Error translation
- ✅ Validation middleware

---

### 5. Health & Observability Endpoints (10%)
**Location**: `internal/infrastructure/http/handlers/`

**Files to Create**:
- `health.go` - Health check endpoints
- `metrics.go` - Metrics endpoints
- `debug.go` - Debug endpoints (dev only)

**Endpoints**:
- `GET /health` - Basic health check
- `GET /health/ready` - Readiness probe (checks DB, Redis)
- `GET /health/live` - Liveness probe
- `GET /metrics` - Prometheus metrics
- `GET /debug/pprof/*` - Go pprof (dev only)

---

### 6. API Versioning (3%)
**Location**: `internal/infrastructure/http/server/`

**Features**:
- ✅ Route grouping by version
- ✅ `/api/v1/` prefix
- ✅ Version deprecation handling
- ✅ Version header support

---

### 7. Swagger Documentation (2%)
**Location**: `docs/swagger/`

**Files to Create**:
- `swagger.yaml` - OpenAPI 3.0 specification
- Integration with swagger-ui

**Features**:
- ✅ Auto-generated API docs
- ✅ Swagger UI at `/docs`
- ✅ ReDoc alternative at `/redoc`

---

## File Structure

```
erp-system/
├── internal/
│   └── infrastructure/
│       └── http/
│           ├── server/
│           │   ├── server.go          # Main HTTP server
│           │   ├── routes.go          # Route registration
│           │   └── shutdown.go        # Graceful shutdown
│           ├── middleware/
│           │   ├── cors.go            # CORS middleware
│           │   ├── logger.go          # Request logging
│           │   ├── recovery.go        # Panic recovery
│           │   ├── request_id.go      # Request ID
│           │   ├── rate_limit.go      # Rate limiting
│           │   ├── timeout.go         # Request timeout
│           │   ├── compress.go        # Compression
│           │   └── validator.go       # Validation
│           └── handlers/
│               ├── health.go          # Health checks
│               ├── metrics.go         # Prometheus metrics
│               └── debug.go           # Debug endpoints
├── pkg/
│   └── http/
│       ├── response.go                # Response helpers
│       ├── request.go                 # Request helpers
│       ├── pagination.go              # Pagination
│       └── binding.go                 # Request binding
├── docs/
│   └── swagger/
│       └── swagger.yaml               # OpenAPI spec
└── cmd/
    └── api/
        └── main.go                    # Updated with HTTP server
```

---

## Implementation Order

### Phase 1: Basic Server (Steps 1-3)
1. ✅ Create server foundation
2. ✅ Set up basic routes
3. ✅ Test server startup

### Phase 2: Middleware Stack (Steps 4-7)
4. ✅ Add CORS middleware
5. ✅ Add logger middleware
6. ✅ Add recovery middleware
7. ✅ Add request ID middleware

### Phase 3: Request/Response (Steps 8-9)
8. ✅ Create response helpers
9. ✅ Create request helpers

### Phase 4: Health & Metrics (Steps 10-11)
10. ✅ Implement health check endpoints
11. ✅ Implement metrics endpoint

### Phase 5: Advanced Features (Steps 12-14)
12. ✅ Add rate limiting
13. ✅ Add validation middleware
14. ✅ Set up Swagger docs

---

## Dependencies to Add

```bash
# Already have fiber
github.com/gofiber/fiber/v2          # HTTP framework

# New dependencies needed
github.com/gofiber/swagger            # Swagger integration
github.com/gofiber/adaptor/v2         # Prometheus adapter
github.com/prometheus/client_golang   # Prometheus client
github.com/swaggo/swag                # Swagger generator
```

---

## Testing Plan

After implementation, we'll test:
1. Server startup and shutdown
2. All middleware functioning
3. Health check endpoints
4. Request/response utilities
5. Rate limiting
6. Validation
7. Metrics collection
8. API documentation

---

## Success Criteria

- ✅ HTTP server starts successfully on configured port
- ✅ All middleware applied in correct order
- ✅ Health check endpoints return correct status
- ✅ Metrics endpoint exposes Prometheus metrics
- ✅ Request ID tracked through logs
- ✅ Rate limiting works correctly
- ✅ Validation catches invalid requests
- ✅ Swagger docs accessible and accurate
- ✅ Graceful shutdown works
- ✅ No memory leaks or panics

---

## Integration with Batch 1

Batch 2 will use:
- ✅ Configuration system (server port, CORS settings)
- ✅ Logger (request logging)
- ✅ Error handling (HTTP error responses)
- ✅ PostgreSQL (health checks)
- ✅ Redis (health checks)

---

## Estimated Time: 3-4 hours

---

**Next**: Start with Phase 1 - Basic Server Setup

**Author**: Sanjaya Weerasinghe
**Date**: October 31, 2025
