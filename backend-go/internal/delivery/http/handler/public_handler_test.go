package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"backend-go/internal/domain/entity"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockPublicUseCase struct {
	cmsContent map[string]interface{}
	err        error
	savedMsg   *entity.ContactMessage
	savedRep   *entity.ResolutionReport
}

func (m *mockPublicUseCase) GetCMSContent(ctx context.Context) (map[string]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cmsContent, nil
}

func (m *mockPublicUseCase) SubmitContactMessage(ctx context.Context, msg *entity.ContactMessage) error {
	m.savedMsg = msg
	return m.err
}

func (m *mockPublicUseCase) SubmitResolutionReport(ctx context.Context, report *entity.ResolutionReport) error {
	m.savedRep = report
	return m.err
}

func TestPublicHandler_HealthAndReady(t *testing.T) {
	h := NewPublicHandler(nil)

	t.Run("HealthCheck", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		h.HealthCheck(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "healthy")
	})

	t.Run("ReadyCheck", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		h.ReadyCheck(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ready")
	})
}

func TestPublicHandler_CMSContent(t *testing.T) {
	mockUC := &mockPublicUseCase{
		cmsContent: map[string]interface{}{"hero_title": "Welcome to Neighbor Service"},
	}
	h := NewPublicHandler(mockUC)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/cms", nil)
	c.Request = req

	h.GetCMSContent(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Welcome to Neighbor Service")
}

func TestPublicHandler_SubmitContactMessage(t *testing.T) {
	mockUC := &mockPublicUseCase{}
	h := NewPublicHandler(mockUC)

	t.Run("Valid JSON submission", func(t *testing.T) {
		body := `{"first_name":"Alice","last_name":"Smith","email":"alice@example.com","message":"Hello Neighbor!"}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/contact", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.SubmitContactMessage(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, mockUC.savedMsg)
		assert.Equal(t, "Alice", mockUC.savedMsg.FirstName)
		assert.Equal(t, "alice@example.com", mockUC.savedMsg.Email)
	})

	t.Run("Valid Form POST submission", func(t *testing.T) {
		form := url.Values{}
		form.Add("first_name", "Bob")
		form.Add("email", "bob@example.com")
		form.Add("message", "Need help with service")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		h.SubmitContactMessage(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, mockUC.savedMsg)
		assert.Equal(t, "Bob", mockUC.savedMsg.FirstName)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		body := `{"first_name":"","email":"","message":""}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/contact", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.SubmitContactMessage(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPublicHandler_SubmitResolutionReport(t *testing.T) {
	mockUC := &mockPublicUseCase{}
	h := NewPublicHandler(mockUC)

	t.Run("Valid Resolution Report", func(t *testing.T) {
		report := entity.ResolutionReport{
			Role:        "seeker",
			IssueType:   "dispute",
			Description: "Provider did not show up on time.",
		}
		data, _ := json.Marshal(report)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/resolution", bytes.NewBuffer(data))
		c.Request.Header.Set("Content-Type", "application/json")

		h.SubmitResolutionReport(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, mockUC.savedRep)
		assert.Equal(t, "seeker", mockUC.savedRep.Role)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		body := `{"role":"","issue_type":""}`
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/resolution", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.SubmitResolutionReport(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPublicHandler_RenderStaticPage(t *testing.T) {
	h := NewPublicHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/terms", nil)

	handlerFunc := h.RenderStaticPage("terms")
	handlerFunc(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
