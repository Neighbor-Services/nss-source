package handler

import (
	"fmt"
	"net/http"
	"strconv"

	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/response"
	"backend-go/pkg/webhookbuffer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentUC domainUsecase.PaymentUseCase
}

func NewPaymentHandler(paymentUC domainUsecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{paymentUC: paymentUC}
}

func (h *PaymentHandler) GetSubscriptionPlans(c *gin.Context) {
	plans, err := h.paymentUC.GetSubscriptionPlans(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load subscription plans")
		return
	}
	response.JSON(c, http.StatusOK, plans)
}

func (h *PaymentHandler) ValidateAppleReceipt(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var req domainUsecase.AppleValidationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "receipt_data is required")
		return
	}

	sub, err := h.paymentUC.ValidateAppleReceipt(c.Request.Context(), userUUID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"status":       "active",
		"is_active":    sub.IsActive,
		"subscription": sub,
	})
}

func (h *PaymentHandler) ValidateGooglePlay(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var req domainUsecase.GoogleValidationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "purchase_token and product_id are required")
		return
	}

	sub, err := h.paymentUC.ValidateGooglePlay(c.Request.Context(), userUUID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"status":       "active",
		"is_active":    sub.IsActive,
		"subscription": sub,
	})
}

func (h *PaymentHandler) GetUserSubscription(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	sub, err := h.paymentUC.GetUserSubscription(c.Request.Context(), userUUID)
	if err != nil || sub == nil {
		response.JSON(c, http.StatusOK, gin.H{"is_active": false, "subscription": nil})
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"subscription": sub,
		"is_active":    sub.IsActive,
	})
}

func (h *PaymentHandler) DeleteUserSubscription(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	if err := h.paymentUC.DeleteUserSubscription(c.Request.Context(), userUUID); err != nil {
		response.InternalError(c, "Failed to cancel subscription")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "subscription_cancelled", "message": "Subscription deleted"})
}

func (h *PaymentHandler) AppleS2SWebhook(c *gin.Context) {
	var payload map[string]interface{}
	_ = c.ShouldBindJSON(&payload)
	_ = h.paymentUC.ProcessAppleS2SWebhook(c.Request.Context(), payload)
	response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *PaymentHandler) GooglePubSubWebhook(c *gin.Context) {
	var payload map[string]interface{}
	_ = c.ShouldBindJSON(&payload)
	_ = h.paymentUC.ProcessGooglePubSubWebhook(c.Request.Context(), payload)
	response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *PaymentHandler) StripeWebhook(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		response.BadRequest(c, "Failed to read webhook body")
		return
	}

	sig := c.GetHeader("Stripe-Signature")
	processErr := h.paymentUC.ProcessStripeWebhook(c.Request.Context(), body, sig)

	// Capture event for admin webhook log viewer
	snip := string(body)
	eventType := "unknown"
	status := "OK"
	errMsg := ""
	if len(body) > 0 {
		// Quick type extraction without full unmarshal
		for _, kv := range []struct{ key, prefix string }{
			{"\"type\":\"payment_intent", "payment_intent"},
			{"\"type\":\"charge", "charge"},
			{"\"type\":\"payout", "payout"},
			{"\"type\":\"customer", "customer"},
			{"\"type\":\"account", "account"},
			{"\"type\":\"transfer", "transfer"},
		} {
			if idx := indexOfSubstr(snip, kv.key); idx >= 0 {
				eventType = extractStripeType(snip)
				break
			}
		}
	}
	if processErr != nil {
		status = "ERROR"
		errMsg = processErr.Error()
	}
	webbufID := fmt.Sprintf("wh_%d", webhookbuffer.GlobalBuffer.NextID())
	webbuf := webhookbuffer.WebhookEvent{
		ID:          webbufID,
		Source:      "stripe",
		EventType:   eventType,
		Status:      status,
		PayloadSnip: snip,
		ErrorMsg:    errMsg,
	}
	webbuf.PayloadSnip = snip
	webbuf.ErrorMsg = errMsg
	webbuf.Status = status
	webbuf.Source = "stripe"
	webhookbuffer.GlobalBuffer.Push(webbuf)

	if processErr != nil {
		response.BadRequest(c, processErr.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "received"})
}

// indexOfSubstr is a simple indexOf for webhook type extraction.
func indexOfSubstr(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// extractStripeType tries to pluck the "type" field value from a raw JSON string.
func extractStripeType(body string) string {
	key := `"type":"`
	idx := indexOfSubstr(body, key)
	if idx < 0 {
		return "unknown"
	}
	start := idx + len(key)
	end := indexOfSubstr(body[start:], `"`)
	if end < 0 {
		return "unknown"
	}
	return body[start : start+end]
}

func (h *PaymentHandler) GetWallet(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	wallet, err := h.paymentUC.GetWallet(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to load wallet")
		return
	}

	response.JSON(c, http.StatusOK, wallet)
}

func (h *PaymentHandler) MyWallet(c *gin.Context) {
	h.GetWallet(c)
}

func (h *PaymentHandler) Transactions(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	txs, err := h.paymentUC.GetWalletTransactions(c.Request.Context(), userUUID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to load transactions")
		return
	}

	response.JSON(c, http.StatusOK, txs)
}

func (h *PaymentHandler) RequestPayout(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid amount format")
		return
	}

	res, err := h.paymentUC.RequestPayout(c.Request.Context(), userUUID, req.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, res)
}

func (h *PaymentHandler) Onboard(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)
	emailStr := c.GetString("userEmail")

	res, err := h.paymentUC.StripeOnboard(c.Request.Context(), userUUID, emailStr)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, res)
}

func (h *PaymentHandler) OnboardingStatus(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	isOnboarded, err := h.paymentUC.GetOnboardingStatus(c.Request.Context(), userUUID)
	if err != nil {
		response.JSON(c, http.StatusOK, gin.H{"is_onboarded": false})
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"is_onboarded": isOnboarded})
}

func (h *PaymentHandler) StripeDashboard(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	url, err := h.paymentUC.GetStripeDashboardLink(c.Request.Context(), userUUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"url": url})
}

func (h *PaymentHandler) FundBackgroundCheck(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	res, err := h.paymentUC.FundBackgroundCheck(c.Request.Context(), userUUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, res)
}

func (h *PaymentHandler) GetCustomerUser(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	if qUserID := c.Query("user_id"); qUserID != "" {
		if parsed, err := uuid.Parse(qUserID); err == nil {
			userUUID = parsed
		}
	} else {
		var body struct {
			UserID string `json:"user_id"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.UserID != "" {
			if parsed, err := uuid.Parse(body.UserID); err == nil {
				userUUID = parsed
			}
		}
	}

	cust, err := h.paymentUC.GetCustomer(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to load customer")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"customer": cust})
}

func (h *PaymentHandler) CreateCustomer(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	cust, err := h.paymentUC.CreateCustomer(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to create customer")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"customer": cust})
}

func (h *PaymentHandler) CreateEphemeralKey(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	secret, err := h.paymentUC.CreateEphemeralKey(c.Request.Context(), userUUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"customer": gin.H{"ephemeral_secret": secret}})
}

func (h *PaymentHandler) UpdatePaymentMethod(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var body struct {
		ID string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Payment method ID required")
		return
	}

	if err := h.paymentUC.UpdatePaymentMethod(c.Request.Context(), userUUID, body.ID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "updated"})
}

func (h *PaymentHandler) AccountConnect(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	acctID, err := h.paymentUC.AccountConnect(c.Request.Context(), userUUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"message": "Account connected successfully", "account_id": acctID})
}

func (h *PaymentHandler) Transfer(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var body struct {
		Amount float64 `json:"amount" binding:"required"`
		UserID string  `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Amount and user_id required")
		return
	}

	targetUUID, err := uuid.Parse(body.UserID)
	if err != nil {
		response.BadRequest(c, "Invalid user_id")
		return
	}

	txID, err := h.paymentUC.Transfer(c.Request.Context(), userUUID, targetUUID, body.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "Transfer successful", "transfer_id": txID})
}

func (h *PaymentHandler) FundAppointment(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var body struct {
		AppointmentID string  `json:"appointment_id" binding:"required"`
		Amount        float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "appointment_id is required")
		return
	}

	aptUUID, err := uuid.Parse(body.AppointmentID)
	if err != nil {
		response.BadRequest(c, "Invalid appointment_id")
		return
	}

	res, err := h.paymentUC.FundAppointment(c.Request.Context(), userUUID, aptUUID, body.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, res)
}

func (h *PaymentHandler) PaymentSheet(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var body struct {
		Amount float64 `json:"amount"`
	}
	_ = c.ShouldBindJSON(&body)

	res, err := h.paymentUC.PaymentSheet(c.Request.Context(), userUUID, body.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, res)
}

func (h *PaymentHandler) TipProvider(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var body struct {
		ProviderID    string  `json:"provider_id" binding:"required"`
		AppointmentID string  `json:"appointment_id" binding:"required"`
		Amount        float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Valid provider_id, appointment_id, and positive amount are required")
		return
	}

	provUUID, err := uuid.Parse(body.ProviderID)
	if err != nil {
		response.BadRequest(c, "Invalid provider_id")
		return
	}

	aptUUID, err := uuid.Parse(body.AppointmentID)
	if err != nil {
		response.BadRequest(c, "Invalid appointment_id")
		return
	}

	txID, err := h.paymentUC.TipProvider(c.Request.Context(), userUUID, provUUID, aptUUID, body.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"status":         "success",
		"message":        "Tip sent successfully",
		"transaction_id": txID,
		"amount":         body.Amount,
	})
}

