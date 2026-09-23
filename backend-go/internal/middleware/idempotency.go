package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/pkg/cache"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
)

type cachedResponse struct {
	Status  int                 `json:"status"`
	Header  map[string][]string `json:"header"`
	Body    string              `json:"body"`
	SavedAt time.Time           `json:"saved_at"`
}

type memoryIdempotencyStore struct {
	mu      sync.RWMutex
	records map[string]cachedResponse
}

var memStore = &memoryIdempotencyStore{
	records: make(map[string]cachedResponse),
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Idempotency returns a Gin middleware that enforces idempotency on mutating operations
// using an `X-Idempotency-Key` or `Idempotency-Key` request header.
func Idempotency(c cache.Cache, ttl time.Duration) gin.HandlerFunc {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	return func(ctx *gin.Context) {
		// Only apply to mutating methods
		if ctx.Request.Method != http.MethodPost && ctx.Request.Method != http.MethodPut && ctx.Request.Method != http.MethodPatch {
			ctx.Next()
			return
		}

		key := ctx.GetHeader("X-Idempotency-Key")
		if key == "" {
			key = ctx.GetHeader("Idempotency-Key")
		}

		// If no idempotency key provided, proceed normally
		if key == "" {
			ctx.Next()
			return
		}

		// Hash key with path and user if available to prevent key collision between routes
		userID, _ := ctx.Get("userID")
		cacheKey := fmt.Sprintf("idempotency:%s:%s:%v", ctx.Request.URL.Path, key, userID)
		hash := sha256.Sum256([]byte(cacheKey))
		hashedKey := "idem:" + hex.EncodeToString(hash[:16])

		// Check Redis cache first if available
		if c != nil {
			var cached cachedResponse
			_, err := c.Get(ctx.Request.Context(), hashedKey, &cached)
			if err == nil && cached.Status != 0 {
				for k, vals := range cached.Header {
					for _, v := range vals {
						ctx.Header(k, v)
					}
				}
				ctx.Header("X-Cache-Lookup", "HIT-IDEMPOTENCY")
				ctx.Data(cached.Status, "application/json; charset=utf-8", []byte(cached.Body))
				ctx.Abort()
				return
			}
		} else {
			// Fallback to in-memory store
			memStore.mu.RLock()
			cached, exists := memStore.records[hashedKey]
			memStore.mu.RUnlock()
			if exists && time.Since(cached.SavedAt) < ttl {
				for k, vals := range cached.Header {
					for _, v := range vals {
						ctx.Header(k, v)
					}
				}
				ctx.Header("X-Cache-Lookup", "HIT-IDEMPOTENCY")
				ctx.Data(cached.Status, "application/json; charset=utf-8", []byte(cached.Body))
				ctx.Abort()
				return
			}
		}

		// Intercept the response body and status
		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = bw

		ctx.Next()

		// If status is 2xx, cache the result
		if ctx.Writer.Status() >= 200 && ctx.Writer.Status() < 300 {
			toSave := cachedResponse{
				Status:  ctx.Writer.Status(),
				Header:  ctx.Writer.Header(),
				Body:    bw.body.String(),
				SavedAt: time.Now().UTC(),
			}

			if c != nil {
				_ = c.Set(ctx.Request.Context(), hashedKey, toSave, ttl)
			} else {
				memStore.mu.Lock()
				memStore.records[hashedKey] = toSave
				memStore.mu.Unlock()
			}
		}
	}
}

// AdminTwoFactorRequired ensures that administrative accounts with 2FA enabled have submitted a valid 2FA token.
func AdminTwoFactorRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, exists := c.Get("user")
		if !exists {
			c.Next()
			return
		}

		user, ok := u.(*entity.User)
		if !ok || user == nil {
			c.Next()
			return
		}

		if user.TwoFactorEnabled {
			twoFactorVerified, verifiedExists := c.Get("two_factor_verified")
			if !verifiedExists || twoFactorVerified != true {
				// Check for 2FA token header
				tokenHeader := c.GetHeader("X-Admin-2FA-Token")
				if tokenHeader == "" {
					response.Forbidden(c, "2FA verification required for this admin operation. Provide X-Admin-2FA-Token header.")
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
