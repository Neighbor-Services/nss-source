package recovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSafeRun(t *testing.T) {
	t.Run("SafeRun executes normal task", func(t *testing.T) {
		executed := false
		SafeRun("test-normal", func() {
			executed = true
		})
		assert.True(t, executed)
	})

	t.Run("SafeRun catches panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SafeRun("test-panic", func() {
				panic("simulated panic in background task")
			})
		})
	})

	t.Run("SafeRunWithContext catches panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SafeRunWithContext(context.Background(), "test-ctx-panic", func(ctx context.Context) {
				panic("simulated panic with context")
			})
		})
	})
}

func TestMiddlewareRecovery(t *testing.T) {
	r := gin.New()
	r.Use(Middleware())

	r.GET("/panic", func(c *gin.Context) {
		panic("fatal handler crash")
	})

	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	wPanic := httptest.NewRecorder()
	reqPanic, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(wPanic, reqPanic)

	assert.Equal(t, http.StatusInternalServerError, wPanic.Code)
	assert.Contains(t, wPanic.Body.String(), "An unexpected server error occurred")

	wOK := httptest.NewRecorder()
	reqOK, _ := http.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(wOK, reqOK)

	assert.Equal(t, http.StatusOK, wOK.Code)
}
