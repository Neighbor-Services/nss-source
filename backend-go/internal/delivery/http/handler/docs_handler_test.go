package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-go/internal/delivery/http/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDocsHandler_SwaggerUI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	docsH := handler.NewDocsHandler()
	r.GET("/docs", docsH.GetSwaggerUI)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/docs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "SwaggerUIBundle")
	assert.Contains(t, w.Body.String(), "/openapi.json")
}

func TestDocsHandler_OpenAPISpec(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	docsH := handler.NewDocsHandler()
	r.GET("/openapi.json", docsH.GetOpenAPISpec)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/openapi.json", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var spec map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &spec)
	assert.NoError(t, err)

	assert.Equal(t, "3.0.3", spec["openapi"])
	info, ok := spec["info"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, info["title"], "Neighbor Services")

	paths, ok := spec["paths"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, paths, "/api/v1/accounts/register/")
	assert.Contains(t, paths, "/api/v1/accounts/login/")
	assert.Contains(t, paths, "/api/v1/services/requests/")
	assert.Contains(t, paths, "/api/v1/interactions/appointments/")
	assert.Contains(t, paths, "/api/v1/chat/conversations/")
	assert.Contains(t, paths, "/api/v1/payments/wallet/my_wallet/")
	assert.Contains(t, paths, "/api/v1/admin/dashboard/stats")
	assert.Contains(t, paths, "/api/v1/admin/users")
	assert.Contains(t, paths, "/api/v1/admin/verifications")
	assert.Contains(t, paths, "/api/v1/admin/background-checks")
	assert.Contains(t, paths, "/api/v1/admin/reports")
	assert.Contains(t, paths, "/api/v1/admin/disputes")
	assert.Contains(t, paths, "/api/v1/admin/payouts")
	assert.Contains(t, paths, "/api/v1/admin/wallets")
	assert.Contains(t, paths, "/api/v1/admin/subscriptions")
	assert.Contains(t, paths, "/api/v1/admin/categories")
	assert.Contains(t, paths, "/api/v1/admin/catalog-services")
	assert.Contains(t, paths, "/api/v1/admin/settings")
	assert.Contains(t, paths, "/api/v1/admin/audit-logs")
	assert.Contains(t, paths, "/api/v1/admin/feature-flags")
	assert.Contains(t, paths, "/api/v1/admin/2fa/verify")
	assert.Contains(t, paths, "/api/v1/admin/roles")
	assert.Contains(t, paths, "/api/v1/admin/fraud/risk-alerts")
	assert.Contains(t, paths, "/api/v1/admin/templates/emails")
	assert.Contains(t, paths, "/api/v1/admin/system/backups")
	assert.Contains(t, paths, "/ws/presence")
	assert.Contains(t, paths, "/ws/chat/{conversation_id}")
}
