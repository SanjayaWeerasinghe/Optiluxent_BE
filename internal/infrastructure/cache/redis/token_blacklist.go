package redis

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenBlacklistPrefix = "blacklist:"

// TokenBlacklist stores invalidated JWT tokens in Redis until they expire naturally.
type TokenBlacklist struct {
	client *redis.Client
}

// NewTokenBlacklist creates a new TokenBlacklist backed by the given Redis client.
func NewTokenBlacklist(client *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{client: client}
}

// Add marks a token as blacklisted with the given TTL.
// TTL should match the token's remaining lifetime so the entry cleans itself up.
func (b *TokenBlacklist) Add(ctx context.Context, token string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	key := tokenBlacklistPrefix + hashToken(token)
	return b.client.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted returns true if the token has been blacklisted.
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := tokenBlacklistPrefix + hashToken(token)
	count, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// hashToken returns a fixed-length hex string derived from the token,
// keeping Redis keys short and consistent regardless of token length.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}
