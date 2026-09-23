package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type ipRateLimiter struct {
	tokens     float64
	lastAccess time.Time
}

type RateLimiter struct {
	mu           sync.Mutex
	visitors     map[string]*ipRateLimiter
	rate         float64 // tokens added per second
	capacity     float64 // max bucket capacity
	cleanupTimer *time.Ticker
}

func NewRateLimiter(maxRequests int, perWindow time.Duration) *RateLimiter {
	rate := float64(maxRequests) / perWindow.Seconds()
	capacity := float64(maxRequests)

	rl := &RateLimiter{
		visitors:     make(map[string]*ipRateLimiter),
		rate:         rate,
		capacity:     capacity,
		cleanupTimer: time.NewTicker(2 * perWindow),
	}

	go func() {
		for range rl.cleanupTimer.C {
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.Sub(v.lastAccess) > (3 * perWindow) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *RateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &ipRateLimiter{
			tokens:     rl.capacity - 1,
			lastAccess: now,
		}
		return true, 0
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(v.lastAccess).Seconds()
	v.tokens += elapsed * rl.rate
	if v.tokens > rl.capacity {
		v.tokens = rl.capacity
	}
	v.lastAccess = now

	if v.tokens >= 1 {
		v.tokens -= 1
		return true, 0
	}

	// Calculate wait time
	missing := 1.0 - v.tokens
	retryAfter := time.Duration(missing/rl.rate) * time.Second
	return false, retryAfter
}

// RateLimit returns a Gin middleware that rate limits requests per client IP.
func RateLimit(maxRequests int, perWindow time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(maxRequests, perWindow)
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		allowed, retryAfter := limiter.Allow(clientIP)
		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			response.JSON(c, http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthRateLimit restricts authentication and OTP endpoints to prevent brute-force attacks.
func AuthRateLimit() gin.HandlerFunc {
	return RateLimit(10, 1*time.Minute)
}

// TwoFactorRateLimit restricts 2FA verification attempts.
func TwoFactorRateLimit() gin.HandlerFunc {
	return RateLimit(5, 1*time.Minute)
}
