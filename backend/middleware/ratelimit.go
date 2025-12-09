package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu       sync.Mutex
	tokens   map[string]int
	lastTime map[string]time.Time
	rate     int           // tokens per interval
	interval time.Duration // refill interval
	burst    int           // max tokens
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, interval time.Duration, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:   make(map[string]int),
		lastTime: make(map[string]time.Time),
		rate:     rate,
		interval: interval,
		burst:    burst,
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	last, exists := rl.lastTime[key]

	if !exists {
		rl.tokens[key] = rl.burst - 1
		rl.lastTime[key] = now
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(last)
	refill := int(elapsed / rl.interval) * rl.rate
	tokens := rl.tokens[key] + refill
	if tokens > rl.burst {
		tokens = rl.burst
	}

	if tokens > 0 {
		rl.tokens[key] = tokens - 1
		rl.lastTime[key] = now
		return true
	}

	return false
}

// RateLimitMiddleware creates a rate limiting middleware
// Default: 10 requests per second with burst of 20
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many requests, please try again later",
				"code":    "RATE_LIMITED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
