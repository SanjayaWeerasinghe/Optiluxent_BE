# Batch 3: Authentication & Authorization - Implementation Plan

**Phase**: Batch 3 - Authentication & Authorization
**Start Date**: October 31, 2025
**Developer**: Sanjaya Weerasinghe
**Status**: 🟡 In Progress (40% Complete)
**Current Session**: Paused - See BATCH_3_STATUS.md for details

---

## Overview

Implement comprehensive authentication and authorization system with JWT tokens, user management, role-based access control (RBAC), and session management.

---

## Components to Build

### 1. JWT Token Management (25%)
**Location**: `pkg/jwt/`

**Files to Create**:
- `jwt.go` - JWT token generation and validation
- `claims.go` - Custom JWT claims structure
- `token_store.go` - Token storage interface

**Features**:
- ✅ Access token generation (short-lived)
- ✅ Refresh token generation (long-lived)
- ✅ Token validation and parsing
- ✅ Token expiration handling
- ✅ Custom claims (user ID, role, permissions)
- ✅ Token blacklisting support

---

### 2. User Domain Layer (20%)
**Location**: `internal/domain/user/`

**Files to Create**:
- `entity.go` - User entity definition
- `repository.go` - User repository interface
- `service.go` - User business logic

**Features**:
- ✅ User entity with fields (ID, email, password hash, role, status)
- ✅ Password hashing (bcrypt)
- ✅ Email validation
- ✅ User status (active, inactive, suspended)
- ✅ Timestamps (created_at, updated_at)

---

### 3. User Repository (15%)
**Location**: `internal/infrastructure/persistence/user/`

**Files to Create**:
- `postgres_repository.go` - PostgreSQL implementation
- `migrations.sql` - Database schema

**Features**:
- ✅ CRUD operations for users
- ✅ Find by email
- ✅ Find by ID
- ✅ Update password
- ✅ Database transactions

**Database Schema**:
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
```

---

### 4. Authentication Handlers (20%)
**Location**: `internal/interfaces/http/handlers/auth/`

**Files to Create**:
- `handler.go` - Authentication handler
- `dto.go` - Data transfer objects (request/response)

**Endpoints**:
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/me` - Get current user info

**Request/Response DTOs**:
```go
type RegisterRequest {
    Email     string
    Password  string
    FirstName string
    LastName  string
}

type LoginRequest {
    Email    string
    Password string
}

type LoginResponse {
    AccessToken  string
    RefreshToken string
    User         UserDTO
}
```

---

### 5. Authentication Middleware (10%)
**Location**: `internal/infrastructure/http/middleware/`

**Files to Create**:
- `auth.go` - JWT authentication middleware
- `permission.go` - Permission checking middleware

**Features**:
- ✅ Extract JWT from Authorization header
- ✅ Validate JWT token
- ✅ Set user context
- ✅ Check token expiration
- ✅ Handle authentication errors

---

### 6. Role-Based Access Control (RBAC) (5%)
**Location**: `internal/domain/rbac/`

**Files to Create**:
- `roles.go` - Role definitions
- `permissions.go` - Permission definitions

**Roles**:
- `admin` - Full system access
- `manager` - Module management access
- `user` - Basic user access
- `guest` - Read-only access

**Permissions** (examples):
- `users:create`
- `users:read`
- `users:update`
- `users:delete`
- `modules:*`

---

### 7. Token Blacklist (Redis) (3%)
**Location**: `internal/infrastructure/cache/`

**Files to Create**:
- `token_blacklist.go` - Token blacklist implementation

**Features**:
- ✅ Add token to blacklist on logout
- ✅ Check if token is blacklisted
- ✅ Auto-expire blacklisted tokens (TTL)

---

### 8. Database Migration System (2%)
**Location**: `internal/infrastructure/database/migrations/`

**Files to Create**:
- `000001_create_users_table.up.sql`
- `000001_create_users_table.down.sql`
- `migrate.go` - Migration runner

---

## File Structure

```
erp-system/
├── pkg/
│   └── jwt/
│       ├── jwt.go                  # JWT token management
│       ├── claims.go               # Custom claims
│       └── token_store.go          # Token storage interface
├── internal/
│   ├── domain/
│   │   ├── user/
│   │   │   ├── entity.go           # User entity
│   │   │   ├── repository.go       # Repository interface
│   │   │   └── service.go          # Business logic
│   │   └── rbac/
│   │       ├── roles.go            # Role definitions
│   │       └── permissions.go      # Permission definitions
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   └── user/
│   │   │       └── postgres_repository.go
│   │   ├── http/
│   │   │   └── middleware/
│   │   │       ├── auth.go         # Auth middleware
│   │   │       └── permission.go   # Permission middleware
│   │   ├── cache/
│   │   │   └── token_blacklist.go  # Token blacklist
│   │   └── database/
│   │       └── migrations/
│   │           ├── 000001_create_users_table.up.sql
│   │           └── 000001_create_users_table.down.sql
│   └── interfaces/
│       └── http/
│           └── handlers/
│               └── auth/
│                   ├── handler.go   # Auth handler
│                   └── dto.go       # DTOs
└── cmd/
    └── migrate/
        └── main.go                  # Migration tool
```

---

## Implementation Order

### Phase 1: Database & User Entity (Steps 1-3)
1. Create database migration system
2. Create users table migration
3. Define User entity and repository interface

### Phase 2: JWT Implementation (Steps 4-5)
4. Implement JWT token generation
5. Implement JWT token validation

### Phase 3: User Repository (Steps 6-7)
6. Implement PostgreSQL user repository
7. Test repository operations

### Phase 4: Authentication Handlers (Steps 8-11)
8. Implement registration endpoint
9. Implement login endpoint
10. Implement refresh token endpoint
11. Implement logout endpoint

### Phase 5: Middleware & RBAC (Steps 12-14)
12. Implement authentication middleware
13. Implement permission middleware
14. Define roles and permissions

### Phase 6: Token Blacklist (Step 15)
15. Implement Redis token blacklist

---

## Dependencies to Add

```bash
# JWT library
github.com/golang-jwt/jwt/v5

# Password hashing (already using bcrypt from golang.org/x/crypto)
golang.org/x/crypto/bcrypt

# Database migration tool
github.com/golang-migrate/migrate/v4
github.com/golang-migrate/migrate/v4/database/postgres
github.com/golang-migrate/migrate/v4/source/file
```

---

## Security Considerations

1. **Password Security**:
   - Use bcrypt with cost factor 12
   - Never log or expose passwords
   - Enforce password complexity rules

2. **Token Security**:
   - Short access token expiry (15 minutes - 1 hour)
   - Long refresh token expiry (7-30 days)
   - Secure token storage on client
   - Token rotation on refresh

3. **API Security**:
   - Rate limiting on auth endpoints
   - CORS properly configured
   - HTTPS only in production
   - Secure cookie flags (httpOnly, secure, sameSite)

4. **Session Security**:
   - Token blacklist for logout
   - Automatic token cleanup
   - Session timeout handling

---

## Testing Plan

After implementation, we'll test:
1. User registration with validation
2. User login with correct/incorrect credentials
3. Token generation and validation
4. Token refresh flow
5. Logout and token blacklist
6. Protected endpoint access
7. Permission-based access control
8. Token expiration handling
9. Database migrations
10. Concurrent login sessions

---

## Success Criteria

- ✅ Users can register with email and password
- ✅ Users can login and receive JWT tokens
- ✅ Access token can be validated
- ✅ Refresh token can generate new access token
- ✅ Users can logout (token blacklisted)
- ✅ Protected endpoints require authentication
- ✅ RBAC works for different user roles
- ✅ Password hashing is secure (bcrypt)
- ✅ Database migrations run successfully
- ✅ Token blacklist prevents reuse of logged-out tokens

---

## Integration with Previous Batches

Batch 3 will use:
- ✅ Configuration system (JWT secret, token expiry)
- ✅ Logger (auth events logging)
- ✅ Error handling (auth error responses)
- ✅ PostgreSQL (user storage)
- ✅ Redis (token blacklist)
- ✅ HTTP server (auth endpoints)
- ✅ Middleware (authentication integration)

---

## API Endpoints Summary

```
POST   /api/v1/auth/register      - Register new user
POST   /api/v1/auth/login         - Login user
POST   /api/v1/auth/refresh       - Refresh access token
POST   /api/v1/auth/logout        - Logout user
GET    /api/v1/auth/me            - Get current user info
GET    /api/v1/users              - List users (admin only)
GET    /api/v1/users/:id          - Get user by ID
PUT    /api/v1/users/:id          - Update user
DELETE /api/v1/users/:id          - Delete user (admin only)
```

---

## Estimated Time: 4-5 hours

---

**Next**: Start with Phase 1 - Database Migrations & User Entity

**Author**: Sanjaya Weerasinghe
**Date**: October 31, 2025
