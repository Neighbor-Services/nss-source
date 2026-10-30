package http_test

import (
	"testing"

	"backend-go/internal/config"
	"backend-go/internal/delivery/http"
	"backend-go/internal/delivery/http/handler"

	"github.com/stretchr/testify/assert"
)

func TestSetupRouter_NoDuplicateRoutes(t *testing.T) {
	assert.NotPanics(t, func() {
		cfg := &config.Config{
			Env: "test",
		}
		deps := http.RouterDependencies{
			Config:          cfg,
			AuthHandler:     &handler.AuthHandler{},
			ProfileHandler:  &handler.ProfileHandler{},
			ServiceHandler:  &handler.ServiceHandler{},
			InterHandler:    &handler.InteractionHandler{},
			ChatHandler:     &handler.ChatHandler{},
			NotifHandler:    &handler.NotificationHandler{},
			PaymentHandler:  &handler.PaymentHandler{},
			ConsultHandler:  &handler.ConsultationHandler{},
			ModerateHandler: &handler.ModerationHandler{},
			AuditHandler:    &handler.AuditHandler{},
			AdminHandler:    &handler.AdminHandler{},
			PublicHandler:   &handler.PublicHandler{},
			DocsHandler:     &handler.DocsHandler{},
			GoStoreHandler:  &handler.GoStoreHandler{},
			LocationHandler: &handler.LocationHandler{},
		}
		router := http.SetupRouter(deps)
		assert.NotNil(t, router)
	})
}
