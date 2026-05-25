package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"erp-system/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// Cache interface defines caching operations
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	FlushAll(ctx context.Context) error
}

// RedisCache implements the Cache interface
type RedisCache struct {
	client *redis.Client
	prefix string
}

// NewCache creates a new Redis cache instance
func NewCache(client *redis.Client, prefix string) Cache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// getKey returns the prefixed key
func (c *RedisCache) getKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// Get retrieves a value from cache and unmarshals it into dest
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := c.getKey(key)

	val, err := c.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		logger.Debug("Cache miss", logger.String("key", fullKey))
		return fmt.Errorf("cache key not found: %s", key)
	}
	if err != nil {
		logger.Error("Cache get error", logger.Err(err), logger.String("key", fullKey))
		return err
	}

	logger.Debug("Cache hit", logger.String("key", fullKey))

	// Unmarshal JSON into dest
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		logger.Error("Cache unmarshal error", logger.Err(err), logger.String("key", fullKey))
		return err
	}

	return nil
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.getKey(key)

	// Marshal value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		logger.Error("Cache marshal error", logger.Err(err), logger.String("key", fullKey))
		return err
	}

	// Set in Redis
	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		logger.Error("Cache set error", logger.Err(err), logger.String("key", fullKey))
		return err
	}

	logger.Debug("Cache set",
		logger.String("key", fullKey),
		logger.Duration("ttl", ttl),
	)

	return nil
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.getKey(key)

	if err := c.client.Del(ctx, fullKey).Err(); err != nil {
		logger.Error("Cache delete error", logger.Err(err), logger.String("key", fullKey))
		return err
	}

	logger.Debug("Cache delete", logger.String("key", fullKey))
	return nil
}

// Exists checks if a key exists in cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.getKey(key)

	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		logger.Error("Cache exists error", logger.Err(err), logger.String("key", fullKey))
		return false, err
	}

	return count > 0, nil
}

// Expire sets a TTL on an existing key
func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := c.getKey(key)

	if err := c.client.Expire(ctx, fullKey, ttl).Err(); err != nil {
		logger.Error("Cache expire error", logger.Err(err), logger.String("key", fullKey))
		return err
	}

	logger.Debug("Cache expire set",
		logger.String("key", fullKey),
		logger.Duration("ttl", ttl),
	)

	return nil
}

// GetTTL returns the remaining TTL for a key
func (c *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := c.getKey(key)

	ttl, err := c.client.TTL(ctx, fullKey).Result()
	if err != nil {
		logger.Error("Cache get TTL error", logger.Err(err), logger.String("key", fullKey))
		return 0, err
	}

	return ttl, nil
}

// Increment increments a numeric key
func (c *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := c.getKey(key)

	val, err := c.client.Incr(ctx, fullKey).Result()
	if err != nil {
		logger.Error("Cache increment error", logger.Err(err), logger.String("key", fullKey))
		return 0, err
	}

	logger.Debug("Cache increment", logger.String("key", fullKey), logger.Int64("value", val))
	return val, nil
}

// Decrement decrements a numeric key
func (c *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	fullKey := c.getKey(key)

	val, err := c.client.Decr(ctx, fullKey).Result()
	if err != nil {
		logger.Error("Cache decrement error", logger.Err(err), logger.String("key", fullKey))
		return 0, err
	}

	logger.Debug("Cache decrement", logger.String("key", fullKey), logger.Int64("value", val))
	return val, nil
}

// FlushAll removes all keys from cache (use with caution)
func (c *RedisCache) FlushAll(ctx context.Context) error {
	if err := c.client.FlushAll(ctx).Err(); err != nil {
		logger.Error("Cache flush all error", logger.Err(err))
		return err
	}

	logger.Warn("Cache flushed - all keys deleted")
	return nil
}

// GetString retrieves a string value from cache
func (c *RedisCache) GetString(ctx context.Context, key string) (string, error) {
	fullKey := c.getKey(key)

	val, err := c.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("cache key not found: %s", key)
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

// SetString stores a string value in cache
func (c *RedisCache) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	fullKey := c.getKey(key)
	return c.client.Set(ctx, fullKey, value, ttl).Err()
}

// DeletePattern deletes all keys matching a pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := c.getKey(pattern)

	iter := c.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			logger.Error("Error deleting key", logger.Err(err), logger.String("key", iter.Val()))
		}
	}

	if err := iter.Err(); err != nil {
		logger.Error("Error scanning keys", logger.Err(err))
		return err
	}

	logger.Debug("Cache pattern delete", logger.String("pattern", fullPattern))
	return nil
}

// Remember retrieves a value from cache or executes a function to populate it
func (c *RedisCache) Remember(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	var result interface{}
	err := c.Get(ctx, key, &result)
	if err == nil {
		return result, nil
	}

	// Cache miss, execute function
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.Set(ctx, key, value, ttl); err != nil {
		// Log error but don't fail the request
		logger.Error("Failed to cache result", logger.Err(err), logger.String("key", key))
	}

	return value, nil
}
