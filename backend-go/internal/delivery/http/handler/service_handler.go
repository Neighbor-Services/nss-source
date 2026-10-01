package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/media"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ServiceHandler struct {
	serviceUC domainUsecase.ServiceUseCase
}

func NewServiceHandler(serviceUC domainUsecase.ServiceUseCase) *ServiceHandler {
	return &ServiceHandler{serviceUC: serviceUC}
}

func (h *ServiceHandler) GetCategories(c *gin.Context) {
	categories, err := h.serviceUC.GetCategories(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load categories")
		return
	}
	response.JSON(c, http.StatusOK, categories)
}

func (h *ServiceHandler) GetCatalogServices(c *gin.Context) {
	category := c.Query("category")
	search := c.Query("search")

	services, err := h.serviceUC.GetCatalogServices(c.Request.Context(), category, search)
	if err != nil {
		response.InternalError(c, "Failed to load catalog services")
		return
	}
	response.JSON(c, http.StatusOK, services)
}

func (h *ServiceHandler) MatchProviders(c *gin.Context) {
	var req domainUsecase.MatchProvidersInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	if userIDStr := c.GetString("userID"); userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			req.UserID = uid
		}
	}

	providers, err := h.serviceUC.MatchProviders(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "Failed to match providers")
		return
	}
	response.JSON(c, http.StatusOK, providers)
}

func (h *ServiceHandler) GetRequests(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userType := c.GetString("userType")
	userUUID, _ := uuid.Parse(userIDStr)
	status := c.Query("status")
	targeted := c.Query("targeted") == "true"

	userMe := c.Query("user_me") == "true" || c.Query("seeker_me") == "true" || c.Query("my_requests") == "true"
	if userMe {
		userType = "SEEKER"
	} else if queryUserType := c.Query("user_type"); queryUserType != "" {
		userType = queryUserType
	}

	var optParams repository.ServiceRequestFilterParams
	if latStr := c.Query("lat"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			optParams.Latitude = &lat
		}
	} else if latStr := c.Query("latitude"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			optParams.Latitude = &lat
		}
	}
	if lngStr := c.Query("lng"); lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
			optParams.Longitude = &lng
		}
	} else if lngStr := c.Query("longitude"); lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
			optParams.Longitude = &lng
		}
	}
	if radStr := c.Query("radius"); radStr != "" {
		if rad, err := strconv.ParseFloat(radStr, 64); err == nil && rad > 0 {
			optParams.RadiusKm = &rad
		}
	} else if radStr := c.Query("radius_km"); radStr != "" {
		if rad, err := strconv.ParseFloat(radStr, 64); err == nil && rad > 0 {
			optParams.RadiusKm = &rad
		}
	} else if radStr := c.Query("distance"); radStr != "" {
		if rad, err := strconv.ParseFloat(radStr, 64); err == nil && rad > 0 {
			optParams.RadiusKm = &rad
		}
	}
	if catSrv := c.Query("catalog_service"); catSrv != "" {
		if catUUID, err := uuid.Parse(catSrv); err == nil {
			optParams.CatalogServiceID = &catUUID
		} else {
			optParams.CatalogServiceName = catSrv
		}
	} else if catSrv := c.Query("catalog_service_id"); catSrv != "" {
		if catUUID, err := uuid.Parse(catSrv); err == nil {
			optParams.CatalogServiceID = &catUUID
		}
	}

	requests, err := h.serviceUC.GetRequests(c.Request.Context(), userUUID, userType, status, targeted, optParams)
	if err != nil {
		response.InternalError(c, "Failed to load requests")
		return
	}
	response.JSON(c, http.StatusOK, requests)
}

func (h *ServiceHandler) GetRequestByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}

	req, err := h.serviceUC.GetRequestByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Request not found")
		return
	}
	response.JSON(c, http.StatusOK, req)
}

type CreateRequestPayload struct {
	Title                string      `json:"title"`
	Description          string      `json:"description"`
	Service              interface{} `json:"service"`
	ServiceID            interface{} `json:"service_id"`
	CatalogService       interface{} `json:"catalog_service"`
	CatalogServiceID     interface{} `json:"catalog_service_id"`
	TargetProvider       interface{} `json:"target_provider"`
	TargetProviderID     interface{} `json:"target_provider_id"`
	Price                interface{} `json:"price"`
	Status               string      `json:"status"`
	PreferredPaymentMode string      `json:"preferred_payment_mode"`
	PaymentMode          string      `json:"payment_mode"`
	ServiceType          string      `json:"service_type"`
	WithImage            *bool       `json:"with_image"`
	Image                string      `json:"image"`
	Longitude            interface{} `json:"longitude"`
	Latitude             interface{} `json:"latitude"`
	ScheduledTime        interface{} `json:"scheduled_time"`
}

func parseUUIDField(v interface{}) *uuid.UUID {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		val = strings.TrimSpace(val)
		if val == "" || val == "null" || val == "<nil>" {
			return nil
		}
		if u, err := uuid.Parse(val); err == nil && u != uuid.Nil {
			return &u
		}
	case map[string]interface{}:
		if idVal, ok := val["id"]; ok {
			return parseUUIDField(idVal)
		}
	}
	return nil
}

func parseFloatField(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case float64:
		return &val
	case float32:
		f := float64(val)
		return &f
	case int:
		f := float64(val)
		return &f
	case int64:
		f := float64(val)
		return &f
	case string:
		val = strings.TrimSpace(val)
		if val == "" || val == "null" {
			return nil
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return &f
		}
	}
	return nil
}

func parseTimeField(v interface{}) *time.Time {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		val = strings.TrimSpace(val)
		if val == "" || val == "null" {
			return nil
		}
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05.999999999",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, f := range formats {
			if t, err := time.Parse(f, val); err == nil {
				return &t
			}
		}
	}
	return nil
}

func (h *ServiceHandler) CreateRequest(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var payload CreateRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "Invalid service request payload: "+err.Error())
		return
	}

	title := strings.TrimSpace(payload.Title)
	if title == "" {
		response.BadRequest(c, "Title is required")
		return
	}

	description := strings.TrimSpace(payload.Description)

	catID := parseUUIDField(payload.CatalogService)
	if catID == nil {
		catID = parseUUIDField(payload.CatalogServiceID)
	}
	if catID == nil {
		catID = parseUUIDField(payload.Service)
	}
	if catID == nil {
		catID = parseUUIDField(payload.ServiceID)
	}

	targetProvID := parseUUIDField(payload.TargetProvider)
	if targetProvID == nil {
		targetProvID = parseUUIDField(payload.TargetProviderID)
	}

	payMode := payload.PreferredPaymentMode
	if payMode == "" {
		payMode = payload.PaymentMode
	}
	if payMode == "" {
		payMode = "IN_APP"
	}

	status := payload.Status
	if status == "" {
		status = "OPEN"
	}

	withImg := false
	if payload.WithImage != nil {
		withImg = *payload.WithImage
	}

	srvType := payload.ServiceType
	if srvType == "" {
		if str, ok := payload.Service.(string); ok && catID == nil {
			srvType = str
		}
	}

	req := entity.ServiceRequest{
		UserID:               userUUID,
		CatalogServiceID:     catID,
		TargetProviderID:     targetProvID,
		Title:                title,
		Description:          description,
		Price:                parseFloatField(payload.Price),
		Status:               status,
		PreferredPaymentMode: payMode,
		ServiceType:          srvType,
		WithImage:            withImg,
		Image:                payload.Image,
		Longitude:            parseFloatField(payload.Longitude),
		Latitude:             parseFloatField(payload.Latitude),
		ScheduledTime:        parseTimeField(payload.ScheduledTime),
	}

	created, err := h.serviceUC.CreateRequest(c.Request.Context(), userUUID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, created)
}

func (h *ServiceHandler) UpdateRequest(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "Invalid JSON payload")
		return
	}

	updated, err := h.serviceUC.UpdateRequest(c.Request.Context(), userUUID, id, updates)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, updated)
}

func (h *ServiceHandler) DeleteRequest(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}

	if err := h.serviceUC.DeleteRequest(c.Request.Context(), userUUID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "deleted"})
}

func (h *ServiceHandler) ApproveProposal(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	requestID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}

	var body struct {
		ProposalID string `json:"proposal_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "proposal_id is required")
		return
	}

	proposalUUID, err := uuid.Parse(body.ProposalID)
	if err != nil {
		response.BadRequest(c, "Invalid proposal_id")
		return
	}

	req, err := h.serviceUC.ApproveProposal(c.Request.Context(), userUUID, requestID, proposalUUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, req)
}

func (h *ServiceHandler) CancelApproval(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	requestID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}

	if err := h.serviceUC.CancelApproval(c.Request.Context(), userUUID, requestID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "approval cancelled"})
}

func (h *ServiceHandler) UploadRequestImage(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var reqIDStr string
	dataStr := c.PostForm("data")
	if dataStr != "" {
		var d struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(dataStr), &d); err == nil {
			reqIDStr = d.ID
		}
	}
	if reqIDStr == "" {
		reqIDStr = c.PostForm("id")
	}
	if reqIDStr == "" {
		reqIDStr = c.Param("id")
	}

	reqUUID, err := uuid.Parse(reqIDStr)
	if err != nil {
		response.BadRequest(c, "Request ID required")
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		response.BadRequest(c, "No image provided")
		return
	}

	mediaDir := media.ResolveMediaDir("requests")
	filename := fmt.Sprintf("request_%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	savePath := filepath.Join(mediaDir, filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.InternalError(c, "Failed to save uploaded image")
		return
	}

	imageURL := "/media/requests/" + filename
	updatedReq, err := h.serviceUC.UploadRequestImage(c.Request.Context(), userUUID, reqUUID, imageURL)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, updatedReq)
}

func (h *ServiceHandler) wrapProposals(proposals []entity.Proposal) gin.H {
	wrapped := make([]gin.H, 0, len(proposals))
	for _, p := range proposals {
		var providerProfile interface{}
		var seekerProfile interface{}

		if p.Provider != nil && p.Provider.Profile != nil {
			p.Provider.Profile.UserID = p.Provider.ID
			providerProfile = p.Provider.Profile
		}
		if p.Request != nil {
			if p.Request.User != nil && p.Request.User.Profile != nil {
				p.Request.User.Profile.UserID = p.Request.User.ID
				seekerProfile = p.Request.User.Profile
			}
			if p.IsApproved {
				p.Request.Approved = true
				if p.Provider != nil && p.Provider.Profile != nil {
					p.Request.ApprovedUser = p.Provider.Profile
				} else {
					p.Request.ApprovedUser = p.ProviderID.String()
				}
			}
		}

		proposalData := gin.H{
			"id":              p.ID.String(),
			"request":         p.Request,
			"request_id":      p.RequestID.String(),
			"request_details": p.Request,
			"provider":        p.ProviderID.String(),
			"is_approved":     p.IsApproved,
			"created_at":      p.CreatedAt,
			"updated_at":      p.UpdatedAt,
		}

		wrapped = append(wrapped, gin.H{
			"acceptance": proposalData,
			"provider":   providerProfile,
			"user":       seekerProfile,
		})
	}
	return gin.H{"requests_acceptance": wrapped}
}

func (h *ServiceHandler) GetProposals(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userType := c.GetString("userType")
	userUUID, _ := uuid.Parse(userIDStr)

	var requestUUID *uuid.UUID
	reqIDStr := c.Query("request_id")
	if reqIDStr == "" {
		reqIDStr = c.Query("service_request")
	}
	if reqIDStr != "" {
		if uid, err := uuid.Parse(reqIDStr); err == nil {
			requestUUID = &uid
		}
	}

	proposals, err := h.serviceUC.GetProposals(c.Request.Context(), userUUID, userType, requestUUID)
	if err != nil {
		response.InternalError(c, "Failed to load proposals")
		return
	}

	response.JSON(c, http.StatusOK, h.wrapProposals(proposals))
}

type CreateProposalPayload struct {
	RequestID  uuid.UUID
	IsApproved bool
}

func (p *CreateProposalPayload) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for _, key := range []string{"request", "service_request", "request_id", "service_request_id", "service"} {
		if val, exists := raw[key]; exists && val != nil {
			if parsed := parseUUIDField(val); parsed != nil {
				p.RequestID = *parsed
				break
			}
		}
	}

	if approved, ok := raw["is_approved"].(bool); ok {
		p.IsApproved = approved
	}

	return nil
}

func (h *ServiceHandler) CreateProposal(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var payload CreateProposalPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "Invalid proposal payload")
		return
	}

	if payload.RequestID == uuid.Nil {
		response.BadRequest(c, "Valid service request ID is required")
		return
	}

	prop := entity.Proposal{
		RequestID:  payload.RequestID,
		IsApproved: payload.IsApproved,
	}

	created, err := h.serviceUC.CreateProposal(c.Request.Context(), userUUID, &prop)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	wrapped := h.wrapProposals([]entity.Proposal{*created})["requests_acceptance"].([]gin.H)
	if len(wrapped) > 0 {
		response.JSON(c, http.StatusCreated, wrapped[0])
		return
	}
	response.Created(c, "Proposal created successfully", created)
}

func (h *ServiceHandler) DeleteProposal(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid proposal ID")
		return
	}

	if err := h.serviceUC.DeleteProposal(c.Request.Context(), userUUID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusNoContent, nil)
}

// ─── EMERGENCY FLASH DISPATCH HANDLERS ───────────────────────────────────────

func (h *ServiceHandler) CreateFlashDispatch(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var dispatch entity.FlashDispatch
	if err := c.ShouldBindJSON(&dispatch); err != nil {
		response.BadRequest(c, "Invalid emergency flash dispatch payload")
		return
	}

	created, err := h.serviceUC.CreateFlashDispatch(c.Request.Context(), userUUID, &dispatch)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, "Emergency flash dispatch broadcasted successfully", created)
}

func (h *ServiceHandler) GetFlashDispatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid dispatch ID")
		return
	}

	dispatch, err := h.serviceUC.GetFlashDispatch(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Flash dispatch not found")
		return
	}

	response.JSON(c, http.StatusOK, dispatch)
}

func (h *ServiceHandler) AcceptFlashDispatch(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid dispatch ID")
		return
	}

	dispatch, apt, err := h.serviceUC.AcceptFlashDispatch(c.Request.Context(), userUUID, id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"message":     "Emergency flash dispatch accepted successfully",
		"dispatch":    dispatch,
		"appointment": apt,
	})
}

func (h *ServiceHandler) CancelFlashDispatch(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid dispatch ID")
		return
	}

	if err := h.serviceUC.CancelFlashDispatch(c.Request.Context(), userUUID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"message": "Emergency flash dispatch cancelled",
	})
}

// GetAISuggestions returns real-time dynamic search suggestions and prompt patterns learned from the database catalog.
func (h *ServiceHandler) GetAISuggestions(c *gin.Context) {
	suggestions, err := h.serviceUC.GetAISuggestions(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load AI suggestions")
		return
	}

	type suggestionDTO struct {
		Title           string   `json:"title"`
		TradeConcept    string   `json:"trade_concept"`
		CategoryName    string   `json:"category_name"`
		QueryPattern    string   `json:"query_pattern"`
		SentimentIntent string   `json:"sentiment_intent"`
		PraiseSignals   []string `json:"praise_signals"`
	}

	var dtos []suggestionDTO
	for _, s := range suggestions {
		title := s.SearchQueryPattern
		if title == "" {
			title = s.Keyword
		}
		dtos = append(dtos, suggestionDTO{
			Title:           title,
			TradeConcept:    s.TradeConcept,
			CategoryName:    s.CategoryName,
			QueryPattern:    s.SearchQueryPattern,
			SentimentIntent: s.SentimentIntent,
			PraiseSignals:   s.PraiseSignals,
		})
	}

	response.JSON(c, http.StatusOK, gin.H{
		"suggestions": dtos,
		"total":       len(dtos),
	})
}

// TriggerCatalogReindex triggers background knowledge reindexing from the database catalog.
func (h *ServiceHandler) TriggerCatalogReindex(c *gin.Context) {
	if err := h.serviceUC.TriggerCatalogReindex(c.Request.Context()); err != nil {
		response.InternalError(c, "Failed to reindex catalog knowledge")
		return
	}
	response.JSON(c, http.StatusOK, gin.H{
		"message": "AI catalog knowledge reindexed successfully",
	})
}

// ParseVoiceSpeech parses spoken speech into an AI-refined request draft with sentiment, category match, and budget.
func (h *ServiceHandler) ParseVoiceSpeech(c *gin.Context) {
	var body struct {
		Speech string `json:"speech" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "speech field is required")
		return
	}

	var userUUID *uuid.UUID
	if userIDStr := c.GetString("userID"); userIDStr != "" {
		if u, err := uuid.Parse(userIDStr); err == nil {
			userUUID = &u
		}
	}

	parsedLog, err := h.serviceUC.ParseAndRefineVoiceSpeech(c.Request.Context(), userUUID, body.Speech)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"id":                 parsedLog.ID,
		"raw_transcript":     parsedLog.RawTranscript,
		"refined_transcript": parsedLog.RefinedTranscript,
		"parsed_title":       parsedLog.ParsedTitle,
		"parsed_description": parsedLog.ParsedDescription,
		"service_type":       parsedLog.ServiceType,
		"catalog_service_id": parsedLog.CatalogServiceID,
		"suggested_budget":   parsedLog.SuggestedBudget,
		"budget_type":        parsedLog.BudgetType,
		"urgency":            parsedLog.Urgency,
		"sentiment":          parsedLog.Sentiment,
		"sentiment_score":    parsedLog.SentimentScore,
		"keywords":           parsedLog.Keywords,
		"confidence_score":   parsedLog.ConfidenceScore,
		"catalog_service":    parsedLog.CatalogService,
	})
}


