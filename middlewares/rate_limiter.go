package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gotask-backend/utils"

	"github.com/gin-gonic/gin"
)

// RateLimiterConfig holds the configuration for a rate limiter.
type RateLimiterConfig struct {
	RequestsPerWindow int                       // Max requests allowed per window
	WindowDuration    time.Duration             // Time window
	KeyFunc           func(*gin.Context) string // Extract the rate limit key
}

// slidingWindowRateLimiter implements a simple sliding window rate limiter.
type slidingWindowRateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time // key -> timestamps of requests
	limit    int
	window   time.Duration
}

func newSlidingWindowRateLimiter(limit int, window time.Duration) *slidingWindowRateLimiter {
	return &slidingWindowRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// isAllowed checks if a request is allowed and records it.
func (r *slidingWindowRateLimiter) isAllowed(key string) (allowed bool, remaining int, resetAt int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	// Filter out old timestamps
	var validTimes []time.Time
	for _, t := range r.requests[key] {
		if t.After(windowStart) {
			validTimes = append(validTimes, t)
		}
	}

	currentCount := len(validTimes)

	if currentCount >= r.limit {
		// Rate limit exceeded
		nextReset := validTimes[0].Add(r.window).Unix()
		return false, 0, nextReset
	}

	// Allow this request
	validTimes = append(validTimes, now)
	r.requests[key] = validTimes

	remaining = r.limit - len(validTimes)
	nextReset := now.Add(r.window).Unix()

	return true, remaining, nextReset
}

// RateLimiterMiddleware returns a middleware that enforces rate limiting.
func RateLimiterMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	limiter := newSlidingWindowRateLimiter(config.RequestsPerWindow, config.WindowDuration)

	return func(c *gin.Context) {
		// Skip health check endpoints
		path := c.Request.URL.Path
		if path == "/health" || path == "/ready" {
			c.Next()
			return
		}

		key := config.KeyFunc(c)

		allowed, remaining, resetAt := limiter.isAllowed(key)

		// Always set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerWindow))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if !allowed {
			logger := utils.GetLogger()
			logger.Warn("Rate limit exceeded",
				"key", key,
				"path", path,
				"method", c.Request.Method,
			)

			retryAfter := int(time.Until(time.Unix(resetAt, 0)).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}

			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", retryAfter),
			})
			return
		}

		c.Next()
	}
}

// IPKeyFunc extracts the client IP as the rate limit key.
// Use this for unauthenticated requests.
func IPKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// UserKeyFunc extracts the user ID from JWT as the rate limit key.
// Use this for authenticated requests.
func UserKeyFunc(c *gin.Context) string {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(interface{ GetID() uint }); ok {
			return fmt.Sprintf("user:%d", u.GetID())
		}
	}
	// Fallback to IP if no user found
	return "ip:" + c.ClientIP()
}
