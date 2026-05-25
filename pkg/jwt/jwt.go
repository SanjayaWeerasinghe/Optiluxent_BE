package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents JWT custom claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	TenantID uint   `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

// Manager handles JWT operations
type Manager struct {
	secret              string
	accessTokenExpiry   time.Duration
	refreshTokenExpiry  time.Duration
	issuer              string
}

// NewManager creates a new JWT manager
func NewManager(secret string, accessExpiry, refreshExpiry time.Duration, issuer string) *Manager {
	return &Manager{
		secret:              secret,
		accessTokenExpiry:   accessExpiry,
		refreshTokenExpiry:  refreshExpiry,
		issuer:              issuer,
	}
}

// GenerateAccessToken generates a new access token
func (m *Manager) GenerateAccessToken(userID, tenantID uint, email, role string) (string, error) {
	return m.generateToken(userID, tenantID, email, role, AccessToken, m.accessTokenExpiry)
}

// GenerateRefreshToken generates a new refresh token
func (m *Manager) GenerateRefreshToken(userID, tenantID uint, email, role string) (string, error) {
	return m.generateToken(userID, tenantID, email, role, RefreshToken, m.refreshTokenExpiry)
}

// generateToken generates a JWT token
func (m *Manager) generateToken(userID, tenantID uint, email, role string, tokenType TokenType, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		Type:     string(tokenType),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

// ValidateToken validates and parses a JWT token
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ExtractClaims extracts claims from a token without validation
func (m *Manager) ExtractClaims(tokenString string) (*Claims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
