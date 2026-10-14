package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-go/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoginBruteForceBlocker_LocksAfterThreshold(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Threshold: 3 attempts within 10s, lock for 1 hour
	blocker := middleware.NewLoginBruteForceBlocker(3, 10*time.Second, 1*time.Hour)
	defer blocker.Stop()

	ip := "198.51.100.25"
	email := "victim@example.com"

	// 1. Initially not locked
	locked, _, _ := blocker.IsLocked(ip, email)
	assert.False(t, locked)

	// 2. First 2 failures: not locked
	newlyLocked, _ := blocker.RecordFailure(ip, email)
	assert.False(t, newlyLocked)
	locked, _, _ = blocker.IsLocked(ip, email)
	assert.False(t, locked)

	newlyLocked, _ = blocker.RecordFailure(ip, email)
	assert.False(t, newlyLocked)
	locked, _, _ = blocker.IsLocked(ip, email)
	assert.False(t, locked)

	// 3. 3rd failure hits threshold: MUST lock
	newlyLocked, rem := blocker.RecordFailure(ip, email)
	assert.True(t, newlyLocked)
	assert.True(t, rem > 50*time.Minute)

	locked, rem2, reason := blocker.IsLocked(ip, email)
	assert.True(t, locked)
	assert.True(t, rem2 > 50*time.Minute)
	assert.NotEmpty(t, reason)

	// Even if checking with a different IP against the same email, it should be locked
	lockedOtherIP, _, _ := blocker.IsLocked("198.51.100.99", email)
	assert.True(t, lockedOtherIP, "Account should be locked against other IPs as well")

	// Even if checking with the same IP against a different email, the IP is locked
	lockedSameIP, _, _ := blocker.IsLocked(ip, "other@example.com")
	assert.True(t, lockedSameIP, "Attacker IP should be locked regardless of target email")

	// 4. RecordSuccess clears locks
	blocker.RecordSuccess(ip, email)
	lockedAfterSuccess, _, _ := blocker.IsLocked(ip, email)
	assert.False(t, lockedAfterSuccess)
}

func TestLoginBruteForceBlocker_Whitelist(t *testing.T) {
	blocker := middleware.NewLoginBruteForceBlocker(2, 5*time.Second, 1*time.Hour)
	defer blocker.Stop()

	whitelistedIP := "127.0.0.1"
	blocker.SetWhitelisted(whitelistedIP, true)

	// Fail 5 times
	for i := 0; i < 5; i++ {
		blocker.RecordFailure(whitelistedIP, "")
	}

	locked, _, _ := blocker.IsLocked(whitelistedIP, "")
	assert.False(t, locked, "Whitelisted IP should not be locked")
}

func TestRespondBruteForceLocked(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/login", func(c *gin.Context) {
		middleware.RespondBruteForceLocked(c, 15*time.Minute, "ip")
	})

	req, _ := http.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "TOO_MANY_FAILED_ATTEMPTS")
	assert.Equal(t, "900", w.Header().Get("Retry-After"))
}
