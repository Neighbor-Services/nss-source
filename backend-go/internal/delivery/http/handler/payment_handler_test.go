package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domainUsecase "backend-go/internal/domain/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPaymentUseCase struct {
	domainUsecase.PaymentUseCase
	gotUserID        uuid.UUID
	gotAppointmentID uuid.UUID
	gotAmount        float64
	resp             map[string]string
	err              error
}

func (s *stubPaymentUseCase) FundAppointment(_ context.Context, userID, appointmentID uuid.UUID, amount float64) (map[string]string, error) {
	s.gotUserID = userID
	s.gotAppointmentID = appointmentID
	s.gotAmount = amount
	return s.resp, s.err
}

func TestPaymentHandler_FundAppointment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	aptID := uuid.New()

	tests := []struct {
		name           string
		payload        string
		wantStatus     int
		wantAmount     float64
		wantErrorSub   string
		expectUCCalled bool
	}{
		{
			name:           "numeric amount",
			payload:        `{"appointment_id":"` + aptID.String() + `","amount":500.00}`,
			wantStatus:     http.StatusOK,
			wantAmount:     500.00,
			expectUCCalled: true,
		},
		{
			name:           "string amount from legacy mobile client",
			payload:        `{"appointment_id":"` + aptID.String() + `","amount":"500.00"}`,
			wantStatus:     http.StatusOK,
			wantAmount:     500.00,
			expectUCCalled: true,
		},
		{
			name:           "prefixed appointment_id with apt_",
			payload:        `{"appointment_id":"apt_` + aptID.String() + `","amount":500.00}`,
			wantStatus:     http.StatusOK,
			wantAmount:     500.00,
			expectUCCalled: true,
		},
		{
			name:         "missing appointment_id",
			payload:      `{"amount":500.00}`,
			wantStatus:   http.StatusBadRequest,
			wantErrorSub: "appointment_id is required",
		},
		{
			name:         "invalid non-numeric amount reports invalid amount instead of appointment_id",
			payload:      `{"appointment_id":"` + aptID.String() + `","amount":"abc"}`,
			wantStatus:   http.StatusBadRequest,
			wantErrorSub: "invalid amount",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := &stubPaymentUseCase{
				resp: map[string]string{"paymentIntent": "pi_test_123"},
			}
			h := NewPaymentHandler(uc)

			r := gin.New()
			r.POST("/customer/fund-appointment/", func(c *gin.Context) {
				c.Set("userID", userID.String())
				h.FundAppointment(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/customer/fund-appointment/", bytes.NewBufferString(tc.payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
			if tc.expectUCCalled {
				assert.Equal(t, userID, uc.gotUserID)
				assert.Equal(t, aptID, uc.gotAppointmentID)
				assert.InDelta(t, tc.wantAmount, uc.gotAmount, 0.0001)

				var body map[string]string
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
				assert.Equal(t, "pi_test_123", body["paymentIntent"])
			} else if tc.wantErrorSub != "" {
				assert.Contains(t, w.Body.String(), tc.wantErrorSub)
			}
		})
	}
}

