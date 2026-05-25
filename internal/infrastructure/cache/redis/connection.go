package redis

import (
	"context"
	"fmt"
	"time"

	"erp-system/internal/infrastructure/config"
	"erp-system/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

// Connect establishes a connection to Redis
func Connect(cfg *config.Config) (*redis.Client, error) {
	if client != nil {
		return client, nil
	}

	// Create Redis client
	client = redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})

	// Ping Redis to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected successfully",
		logger.String("host", cfg.Redis.Host),
		logger.Int("port", cfg.Redis.Port),
		logger.Int("db", cfg.Redis.DB),
	)

	return client, nil
}

// Get returns the Redis client instance
func Get() *redis.Client {
	if client == nil {
		panic("Redis not initialized, call Connect() first")
	}
	return client
}

// Close closes the Redis connection
func Close() error {
	if client == nil {
		return nil
	}

	logger.Info("Closing Redis connection")
	return client.Close()
}

// HealthCheck performs a health check on Redis
func HealthCheck(ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("Redis not initialized")
	}

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	return nil
}

// GetStats returns Redis statistics
func GetStats() (map[string]interface{}, error) {
	if client == nil {
		return nil, fmt.Errorf("Redis not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	poolStats := client.PoolStats()
	info, err := client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"pool_hits":      poolStats.Hits,
		"pool_misses":    poolStats.Misses,
		"pool_timeouts":  poolStats.Timeouts,
		"total_conns":    poolStats.TotalConns,
		"idle_conns":     poolStats.IdleConns,
		"stale_conns":    poolStats.StaleConns,
		"server_info":    info,
	}, nil
}
