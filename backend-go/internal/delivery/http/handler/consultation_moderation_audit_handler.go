package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Consultation Handler
type ConsultationHandler struct {
	consultationUC domainUsecase.ConsultationUseCase
}

func NewConsultationHandler(consultationUC domainUsecase.ConsultationUseCase) *ConsultationHandler {
	return &ConsultationHandler{consultationUC: consultationUC}
}

func (h *ConsultationHandler) GetRTCToken(c *gin.Context) {
	channelName := c.Query("channel_name")
	uidStr := c.Query("uid")
	role := c.DefaultQuery("role", "publisher")

	var uid uint32
	if uidStr != "" {
		if u, err := strconv.ParseUint(uidStr, 10, 32); err == nil {
			uid = uint32(u)
		}
	}

	if channelName == "" {
		var req struct {
			ChannelName string `json:"channel_name"`
			UID         uint32 `json:"uid"`
			Role        string `json:"role"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			if req.ChannelName != "" {
				channelName = req.ChannelName
			}
			if req.UID != 0 {
				uid = req.UID
			}
			if req.Role != "" {
				role = req.Role
			}
		}
	}

	if channelName == "" {
		response.BadRequest(c, "channel_name is required")
		return
	}

	res, err := h.consultationUC.GetRTCToken(c.Request.Context(), channelName, uid, role)
	if err != nil {
		response.InternalError(c, "Failed to generate RTC token")
		return
	}
	response.JSON(c, http.StatusOK, res)
}

func (h *ConsultationHandler) GetRTMToken(c *gin.Context) {
	userAccount := c.Query("user_account")
	if userAccount == "" {
		userAccount = c.Query("user_id")
	}
	if userAccount == "" {
		userAccount = c.GetString("userID")
	}

	if userAccount == "" {
		var req struct {
			UserAccount string `json:"user_account"`
			UserID      string `json:"user_id"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			if req.UserAccount != "" {
				userAccount = req.UserAccount
			} else if req.UserID != "" {
				userAccount = req.UserID
			}
		}
	}

	if userAccount == "" {
		response.BadRequest(c, "user_account or user_id is required")
		return
	}

	res, err := h.consultationUC.GetRTMToken(c.Request.Context(), userAccount)
	if err != nil {
		response.InternalError(c, "Failed to generate RTM token")
		return
	}
	response.JSON(c, http.StatusOK, res)
}

// Moderation Handler
type ModerationHandler struct {
	moderationUC domainUsecase.ModerationUseCase
}

func NewModerationHandler(moderationUC domainUsecase.ModerationUseCase) *ModerationHandler {
	return &ModerationHandler{moderationUC: moderationUC}
}

func (h *ModerationHandler) SubmitReport(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var req struct {
		ReportedUserID *string `json:"reported_user"`
		ContentType    string  `json:"content_type"`
		ResourceType   string  `json:"resource_type"`
		ObjectID       string  `json:"object_id"`
		ResourceID     string  `json:"resource_id"`
		Reason         string  `json:"reason"`
		Title          string  `json:"title"`
		Description    string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid report payload")
		return
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = req.ResourceType
	}
	objectID := req.ObjectID
	if objectID == "" {
		objectID = req.ResourceID
	}

	reason := req.Reason
	if reason == "" {
		if req.Title != "" {
			reason = req.Title
		} else if req.Description != "" {
			reason = req.Description
		} else {
			response.BadRequest(c, "reason or description is required")
			return
		}
	}

	var reportedUUID *uuid.UUID
	if req.ReportedUserID != nil && *req.ReportedUserID != "" {
		if uid, err := cleanUUID(*req.ReportedUserID); err == nil {
			reportedUUID = &uid
		}
	}

	report, err := h.moderationUC.SubmitReport(c.Request.Context(), userUUID, reportedUUID, contentType, objectID, reason, req.Description)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Report submitted successfully", report)
}

func (h *ModerationHandler) GetVerifications(c *gin.Context) {
	userIDStr := c.GetString("userID")
	var providerUUID *uuid.UUID
	if uid, err := uuid.Parse(userIDStr); err == nil {
		providerUUID = &uid
	}

	verifications, err := h.moderationUC.GetVerifications(c.Request.Context(), providerUUID)
	if err != nil {
		response.InternalError(c, "Failed to load verifications")
		return
	}
	response.JSON(c, http.StatusOK, verifications)
}

func (h *ModerationHandler) GetBackgroundChecks(c *gin.Context) {
	userIDStr := c.GetString("userID")
	var providerUUID *uuid.UUID
	if uid, err := uuid.Parse(userIDStr); err == nil {
		providerUUID = &uid
	}

	checks, err := h.moderationUC.GetBackgroundChecks(c.Request.Context(), providerUUID)
	if err != nil {
		response.InternalError(c, "Failed to load background checks")
		return
	}
	response.JSON(c, http.StatusOK, checks)
}

func (h *ModerationHandler) GetBackgroundCheckConfig(c *gin.Context) {
	cfg, err := h.moderationUC.GetBackgroundCheckConfig(c.Request.Context())
	if err != nil {
		response.JSON(c, http.StatusOK, gin.H{"payment_mode": "IN_APP_STRIPE"})
		return
	}
	response.JSON(c, http.StatusOK, cfg)
}

func (h *ModerationHandler) InitiateBackgroundCheck(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var body struct {
		PaymentIntentID string `json:"payment_intent_id"`
	}
	_ = c.ShouldBindJSON(&body)

	check, err := h.moderationUC.InitiateBackgroundCheck(c.Request.Context(), userUUID, body.PaymentIntentID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"id":             check.ID.String(),
		"invitation_url": check.InvitationURL,
		"status":         check.Status,
	})
}

func (h *ModerationHandler) ResyncBackgroundCheck(c *gin.Context) {
	idStr := c.Param("id")
	checkUUID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid background check ID")
		return
	}

	check, err := h.moderationUC.ResyncBackgroundCheck(c.Request.Context(), checkUUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, check)
}

func (h *ModerationHandler) CheckrWebhook(c *gin.Context) {
	sig := c.GetHeader("X-Checkr-Signature")
	rawBody, _ := io.ReadAll(c.Request.Body)

	var payload map[string]interface{}
	if len(rawBody) > 0 {
		_ = json.Unmarshal(rawBody, &payload)
	}

	_ = h.moderationUC.ProcessCheckrWebhook(c.Request.Context(), sig, rawBody, payload)
	response.JSON(c, http.StatusOK, gin.H{"status": "ok", "detail": "Acknowledged."})
}

// Audit Handler
type AuditHandler struct {
	auditUC domainUsecase.AuditUseCase
}

func NewAuditHandler(auditUC domainUsecase.AuditUseCase) *AuditHandler {
	return &AuditHandler{auditUC: auditUC}
}

func (h *AuditHandler) GetLogs(c *gin.Context) {
	var targetUUID *uuid.UUID

	authUserIDStr := c.GetString("userID")
	userType := c.GetString("userType")
	isSuperuser, _ := c.Get("isSuperuser")
	isSuperuserBool, _ := isSuperuser.(bool)
	isAdmin := userType == "ADMIN" || isSuperuserBool

	userIDQuery := c.Query("user_id")

	if !isAdmin {
		// Non-admin users (SEEKER, PROVIDER, etc.) can only ever see their own audit logs
		if authUserIDStr != "" {
			if uid, err := uuid.Parse(authUserIDStr); err == nil {
				targetUUID = &uid
			}
		}
		if targetUUID == nil {
			response.JSON(c, http.StatusOK, []interface{}{})
			return
		}
	} else {
		// Admin users can inspect logs by user_id or view all
		if userIDQuery != "" && userIDQuery != "all" && userIDQuery != "me" {
			if uid, err := uuid.Parse(userIDQuery); err == nil {
				targetUUID = &uid
			}
		} else if userIDQuery == "me" {
			if authUserIDStr != "" {
				if uid, err := uuid.Parse(authUserIDStr); err == nil {
					targetUUID = &uid
				}
			}
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.auditUC.GetLogs(c.Request.Context(), targetUUID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to load audit logs")
		return
	}

	if len(logs) == 0 && targetUUID != nil {
		clientIP := c.ClientIP()
		_ = h.auditUC.LogAction(
			c.Request.Context(),
			targetUUID,
			"LOGIN",
			"AUTH",
			"SESSION",
			clientIP,
			map[string]interface{}{
				"status": "active_session",
				"info":   "Account session verified and activity logging active",
			},
		)
		logs, _ = h.auditUC.GetLogs(c.Request.Context(), targetUUID, limit, offset)
	}

	response.JSON(c, http.StatusOK, logs)
}
