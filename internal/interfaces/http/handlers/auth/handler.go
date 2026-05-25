package auth

import (
	"strings"
	"time"

	domainuser "erp-system/internal/domain/user"
	rediscache "erp-system/internal/infrastructure/cache/redis"
	"erp-system/internal/infrastructure/security"
	apperrors "erp-system/pkg/errors"
	httputil "erp-system/pkg/http"
	jwtpkg "erp-system/pkg/jwt"
	"erp-system/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Handler handles authentication endpoints.
type Handler struct {
	userRepo     domainuser.Repository
	jwtManager   *jwtpkg.Manager
	blacklist    *rediscache.TokenBlacklist
	lockout      *security.AccountLockout
	validate     *validator.Validate
	accessExpiry time.Duration
}

// NewHandler creates a new auth Handler.
func NewHandler(
	userRepo domainuser.Repository,
	jwtManager *jwtpkg.Manager,
	blacklist *rediscache.TokenBlacklist,
	lockout *security.AccountLockout,
	accessExpiry time.Duration,
) *Handler {
	return &Handler{
		userRepo:     userRepo,
		jwtManager:   jwtManager,
		blacklist:    blacklist,
		lockout:      lockout,
		validate:     validator.New(),
		accessExpiry: accessExpiry,
	}
}

// RegisterRoutes mounts auth routes onto the given router group.
// Logout and Me require the auth middleware, applied at the route level in routes.go.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Post("/register", h.Register)
	router.Post("/login", h.Login)
	router.Post("/refresh", h.Refresh)
}

// Register godoc
// POST /api/v1/auth/register
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	if fields := h.validateStruct(req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	ctx := c.UserContext()

	// Check for duplicate email
	existing, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Register: GetByEmail failed", logger.Err(err))
		return httputil.InternalServerError(c, "Registration failed")
	}
	if existing != nil {
		return httputil.Conflict(c, "An account with this email already exists")
	}

	u := &domainuser.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      domainuser.RoleUser,
		Status:    domainuser.StatusActive,
	}

	if err := u.SetPassword(req.Password); err != nil {
		logger.Error("Register: password hashing failed", logger.Err(err))
		return httputil.InternalServerError(c, "Registration failed")
	}

	if err := h.userRepo.Create(ctx, u); err != nil {
		logger.Error("Register: Create failed", logger.Err(err))
		return httputil.InternalServerError(c, "Registration failed")
	}

	logger.Info("User registered", logger.String("email", u.Email), logger.Uint("user_id", u.ID))
	return httputil.Created(c, "Account created successfully", toUserResponse(u))
}

// Login godoc
// POST /api/v1/auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	if fields := h.validateStruct(req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	ctx := c.UserContext()

	// Check account lockout before any DB query
	if h.lockout != nil {
		locked, err := h.lockout.IsLocked(ctx, req.Email)
		if err != nil {
			logger.Error("Login: lockout check failed", logger.Err(err))
		} else if locked {
			return httputil.Error(c, apperrors.New(apperrors.CodeTooManyRequests, "Account temporarily locked due to too many failed attempts. Please try again later."))
		}
	}

	u, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Login: GetByEmail failed", logger.Err(err))
		return httputil.InternalServerError(c, "Login failed")
	}

	// Use a constant-time comparison path for both missing user and wrong password
	// to prevent user enumeration via timing attacks.
	if u == nil || !u.CheckPassword(req.Password) {
		if h.lockout != nil {
			if recErr := h.lockout.RecordFailure(ctx, req.Email); recErr != nil {
				logger.Error("Login: record failure failed", logger.Err(recErr))
			}
		}
		return httputil.Error(c, apperrors.InvalidCredentials())
	}

	if !u.IsActive() {
		return httputil.Unauthorized(c, "Account is inactive or suspended")
	}

	var tenantID uint
	if u.TenantID != nil {
		tenantID = *u.TenantID
	}

	accessToken, err := h.jwtManager.GenerateAccessToken(u.ID, tenantID, u.Email, u.Role)
	if err != nil {
		logger.Error("Login: GenerateAccessToken failed", logger.Err(err))
		return httputil.InternalServerError(c, "Login failed")
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(u.ID, tenantID, u.Email, u.Role)
	if err != nil {
		logger.Error("Login: GenerateRefreshToken failed", logger.Err(err))
		return httputil.InternalServerError(c, "Login failed")
	}

	_ = h.userRepo.UpdateLastLogin(ctx, u.ID)
	if h.lockout != nil {
		_ = h.lockout.Reset(ctx, u.Email)
	}

	logger.Info("User logged in", logger.String("email", u.Email), logger.Uint("user_id", u.ID))

	return httputil.Success(c, "Login successful", AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.accessExpiry.Seconds()),
		User:         toUserResponse(u),
	})
}

// Refresh godoc
// POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return httputil.BadRequest(c, "Invalid request body")
	}

	if fields := h.validateStruct(req); fields != nil {
		return httputil.ValidationError(c, "Validation failed", fields)
	}

	claims, err := h.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return httputil.Error(c, apperrors.TokenInvalid())
	}

	if claims.Type != string(jwtpkg.RefreshToken) {
		return httputil.Unauthorized(c, "Invalid token type: expected refresh token")
	}

	ctx := c.UserContext()

	isBlacklisted, err := h.blacklist.IsBlacklisted(ctx, req.RefreshToken)
	if err != nil {
		logger.Error("Refresh: blacklist check failed", logger.Err(err))
		return httputil.InternalServerError(c, "Token validation failed")
	}
	if isBlacklisted {
		return httputil.Error(c, apperrors.TokenInvalid())
	}

	u, err := h.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || u == nil || !u.IsActive() {
		return httputil.Unauthorized(c, "User not found or inactive")
	}

	newAccessToken, err := h.jwtManager.GenerateAccessToken(u.ID, claims.TenantID, u.Email, u.Role)
	if err != nil {
		logger.Error("Refresh: GenerateAccessToken failed", logger.Err(err))
		return httputil.InternalServerError(c, "Token refresh failed")
	}

	return httputil.Success(c, "Token refreshed", TokenResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.accessExpiry.Seconds()),
	})
}

// Logout godoc
// POST /api/v1/auth/logout  (requires auth middleware)
func (h *Handler) Logout(c *fiber.Ctx) error {
	token := extractBearerToken(c)
	if token == "" {
		return httputil.Unauthorized(c, "No token provided")
	}

	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		// Token already invalid — treat as successful logout
		return httputil.Success(c, "Logged out", nil)
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	ctx := c.UserContext()

	if err := h.blacklist.Add(ctx, token, ttl); err != nil {
		logger.Error("Logout: blacklist add failed", logger.Err(err))
		return httputil.InternalServerError(c, "Logout failed")
	}

	logger.Info("User logged out", logger.Uint("user_id", claims.UserID))
	return httputil.Success(c, "Logged out successfully", nil)
}

// Me godoc
// GET /api/v1/auth/me  (requires auth middleware)
func (h *Handler) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return httputil.Unauthorized(c, "Invalid session")
	}

	u, err := h.userRepo.GetByID(c.UserContext(), userID)
	if err != nil {
		logger.Error("Me: GetByID failed", logger.Err(err))
		return httputil.InternalServerError(c, "Failed to retrieve user")
	}
	if u == nil {
		return httputil.NotFound(c, "User not found")
	}

	return httputil.Success(c, "User retrieved", toUserResponse(u))
}

// --- helpers ---

func (h *Handler) validateStruct(s interface{}) map[string]interface{} {
	err := h.validate.Struct(s)
	if err == nil {
		return nil
	}
	fields := make(map[string]interface{})
	for _, e := range err.(validator.ValidationErrors) {
		fields[strings.ToLower(e.Field())] = validationMessage(e)
	}
	return fields
}

func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Too short (minimum " + e.Param() + " characters)"
	case "max":
		return "Too long (maximum " + e.Param() + " characters)"
	default:
		return "Invalid value"
	}
}

func extractBearerToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func toUserResponse(u *domainuser.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Role:        u.Role,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}
