package middleware

import (
	"fmt"
	"time"

	apperrors "erp-system/pkg/errors"
	httputil "erp-system/pkg/http"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimiter is a Redis-backed sliding-window rate limiter.
type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// PerIP limits requests from a single IP to max per window.
func (r *RateLimiter) PerIP(max int, window time.Duration) fiber.Handler {
	return r.limit(max, window, func(c *fiber.Ctx) string {
		return "ratelimit:ip:" + c.IP()
	})
}

// PerUser limits requests from a single authenticated user.
func (r *RateLimiter) PerUser(max int, window time.Duration) fiber.Handler {
	return r.limit(max, window, func(c *fiber.Ctx) string {
		if uid, ok := c.Locals("user_id").(uint); ok && uid != 0 {
			return fmt.Sprintf("ratelimit:user:%d", uid)
		}
		return "ratelimit:ip:" + c.IP()
	})
}

func (r *RateLimiter) limit(max int, window time.Duration, keyFn func(*fiber.Ctx) string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := keyFn(c)
		ctx := c.UserContext()

		count, err := r.client.Incr(ctx, key).Result()
		if err != nil {
			// Fail open — don't block requests on Redis errors
			return c.Next()
		}

		if count == 1 {
			_ = r.client.Expire(ctx, key, window)
		}

		if count > int64(max) {
			return httputil.Error(c, apperrors.New(apperrors.CodeTooManyRequests, "Rate limit exceeded. Please try again later."))
		}

		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", max))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max-int(count)))

		return c.Next()
	}
}
