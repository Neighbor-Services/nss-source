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

func TestAnti404Scanner_BlocksAfterThreshold(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Configure blocker with threshold of 5 for faster testing, 1s window, 7 days ban
	blocker := middleware.NewIPBlocker(5, 10*time.Second, 7*24*time.Hour)
	defer blocker.Stop()

	r := gin.New()
	r.Use(middleware.Anti404Scanner(blocker))

	r.GET("/valid", func(c *gin.Context) {
		c.String(http.StatusOK, "valid endpoint")
	})

	attackerIP := "203.0.113.42"

	// 1. Valid request from attacker IP should succeed
	reqValid, _ := http.NewRequest(http.MethodGet, "/valid", nil)
	reqValid.RemoteAddr = attackerIP + ":12345"
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)
	assert.Equal(t, http.StatusOK, wValid.Code)

	// 2. Trigger 404 four times (less than threshold of 5)
	for i := 1; i <= 4; i++ {
		req404, _ := http.NewRequest(http.MethodGet, "/non-existent-probe", nil)
		req404.RemoteAddr = attackerIP + ":12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req404)
		assert.Equal(t, http.StatusNotFound, w.Code)

		banned, _ := blocker.IsBanned(attackerIP)
		assert.False(t, banned, "Attacker should not be banned before hitting threshold")
	}

	// 3. Trigger the 5th 404 (reaching threshold)
	req404Final, _ := http.NewRequest(http.MethodGet, "/wp-login.php", nil)
	req404Final.RemoteAddr = attackerIP + ":12345"
	wFinal := httptest.NewRecorder()
	r.ServeHTTP(wFinal, req404Final)
	assert.Equal(t, http.StatusNotFound, wFinal.Code)

	// Now attacker MUST be banned
	banned, remaining := blocker.IsBanned(attackerIP)
	assert.True(t, banned, "Attacker should be banned after 5th 404")
	assert.True(t, remaining > 6*24*time.Hour, "Ban duration should be ~7 days")

	// 4. Any subsequent request (even to valid endpoints!) must be blocked with HTTP 403 Forbidden
	reqBlocked, _ := http.NewRequest(http.MethodGet, "/valid", nil)
	reqBlocked.RemoteAddr = attackerIP + ":12345"
	wBlocked := httptest.NewRecorder()
	r.ServeHTTP(wBlocked, reqBlocked)

	assert.Equal(t, http.StatusForbidden, wBlocked.Code)
	assert.Contains(t, wBlocked.Body.String(), "IP_BLOCKED_SCANNER")
	assert.NotEmpty(t, wBlocked.Header().Get("Retry-After"))

	// 5. Unban attacker and verify they can access again
	blocker.Unban(attackerIP)
	bannedAfterUnban, _ := blocker.IsBanned(attackerIP)
	assert.False(t, bannedAfterUnban)

	wAllowedAgain := httptest.NewRecorder()
	reqAllowedAgain, _ := http.NewRequest(http.MethodGet, "/valid", nil)
	reqAllowedAgain.RemoteAddr = attackerIP + ":12345"
	r.ServeHTTP(wAllowedAgain, reqAllowedAgain)
	assert.Equal(t, http.StatusOK, wAllowedAgain.Code)
}

func TestAnti404Scanner_Whitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	blocker := middleware.NewIPBlocker(3, 5*time.Second, 7*24*time.Hour)
	defer blocker.Stop()

	whitelistedIP := "192.168.1.50"
	blocker.SetWhitelisted(whitelistedIP, true)

	r := gin.New()
	r.Use(middleware.Anti404Scanner(blocker))

	// Trigger 5 404s (exceeding threshold of 3)
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/missing", nil)
		req.RemoteAddr = whitelistedIP + ":12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	}

	banned, _ := blocker.IsBanned(whitelistedIP)
	assert.False(t, banned, "Whitelisted IP should never be banned")
}
