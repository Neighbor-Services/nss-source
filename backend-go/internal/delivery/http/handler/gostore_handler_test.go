package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"backend-go/internal/config"
)

func TestGoStoreHandler_WebhookSignatureVerification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test_gostore_secret_key_123"
	_ = os.Setenv("GOSTORE_WEBHOOK_SECRET", secret)
	defer os.Unsetenv("GOSTORE_WEBHOOK_SECRET")

	cfg := &config.Config{
		GoStoreURL:    "https://file.proleadsolutions.co",
		GoStoreBucket: "neighborservice",
		GoStoreAPIKey: secret,
	}
	h := NewGoStoreHandler(cfg)

	payload := GoStoreWebhookPayload{
		EventID: "evt_123456",
		Event:   "file.uploaded",
		Bucket:  "neighborservice",
		Path:    "profiles/avatar.png",
		Size:    1024,
	}
	payloadBytes, _ := json.Marshal(payload)

	// Calculate HMAC signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payloadBytes)
	signature := hex.EncodeToString(mac.Sum(nil))

	// 1. Valid Signature
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/gostore", bytes.NewReader(payloadBytes))
	c.Request.Header.Set("X-GoStore-Signature", signature)

	h.HandleWebhook(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Invalid Signature
	wInvalid := httptest.NewRecorder()
	cInvalid, _ := gin.CreateTestContext(wInvalid)
	cInvalid.Request = httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/gostore", bytes.NewReader(payloadBytes))
	cInvalid.Request.Header.Set("X-GoStore-Signature", "wrong_signature")

	h.HandleWebhook(cInvalid)
	assert.Equal(t, http.StatusUnauthorized, wInvalid.Code)
}
