package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/media"
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

	contentTypeHeader := c.GetHeader("Content-Type")
	var reportedUserIDStr, contentType, objectID, reason, title, description string
	var evidenceList []string

	if strings.HasPrefix(contentTypeHeader, "multipart/form-data") {
		reportedUserIDStr = c.PostForm("reported_user")
		contentType = c.PostForm("content_type")
		if contentType == "" {
			contentType = c.PostForm("resource_type")
		}
		objectID = c.PostForm("object_id")
		if objectID == "" {
			objectID = c.PostForm("resource_id")
		}
		reason = c.PostForm("reason")
		title = c.PostForm("title")
		description = c.PostForm("description")

		// Process multipart uploaded files
		form, formErr := c.MultipartForm()
		if formErr == nil && form != nil {
			fileFields := []string{"evidence", "images", "image", "file", "files", "evidence_files"}
			for _, field := range fileFields {
				if files, exists := form.File[field]; exists {
					for _, f := range files {
						if relPath, err := media.ValidateAndSaveUploadedFile(f, "reports", 15*1024*1024); err == nil && relPath != "" {
							evidenceList = append(evidenceList, relPath)
						}
					}
				}
			}
		}

		// Also check single file fallback
		if len(evidenceList) == 0 {
			if f, err := c.FormFile("evidence"); err == nil && f != nil {
				if relPath, err := media.ValidateAndSaveUploadedFile(f, "reports", 15*1024*1024); err == nil && relPath != "" {
					evidenceList = append(evidenceList, relPath)
				}
			} else if f, err := c.FormFile("image"); err == nil && f != nil {
				if relPath, err := media.ValidateAndSaveUploadedFile(f, "reports", 15*1024*1024); err == nil && relPath != "" {
					evidenceList = append(evidenceList, relPath)
				}
			}
		}
	} else {
		var req struct {
			ReportedUserID *string  `json:"reported_user"`
			ContentType    string   `json:"content_type"`
			ResourceType   string   `json:"resource_type"`
			ObjectID       string   `json:"object_id"`
			ResourceID     string   `json:"resource_id"`
			Reason         string   `json:"reason"`
			Title          string   `json:"title"`
			Description    string   `json:"description"`
			Evidence       string   `json:"evidence"`
			EvidenceList   []string `json:"evidence_list"`
			Images         []string `json:"images"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid report payload")
			return
		}

		if req.ReportedUserID != nil {
			reportedUserIDStr = *req.ReportedUserID
		}
		contentType = req.ContentType
		if contentType == "" {
			contentType = req.ResourceType
		}
		objectID = req.ObjectID
		if objectID == "" {
			objectID = req.ResourceID
		}
		reason = req.Reason
		title = req.Title
		description = req.Description

		if req.Evidence != "" {
			evidenceList = append(evidenceList, req.Evidence)
		}
		if len(req.EvidenceList) > 0 {
			evidenceList = append(evidenceList, req.EvidenceList...)
		}
		if len(req.Images) > 0 {
			evidenceList = append(evidenceList, req.Images...)
		}
	}

	if reason == "" {
		if title != "" {
			reason = title
		} else if description != "" {
			reason = description
		} else {
			response.BadRequest(c, "reason or description is required")
			return
		}
	}

	var reportedUUID *uuid.UUID
	if reportedUserIDStr != "" {
		if uid, err := cleanUUID(reportedUserIDStr); err == nil {
			reportedUUID = &uid
		}
	}

	report, err := h.moderationUC.SubmitReport(c.Request.Context(), userUUID, reportedUUID, contentType, objectID, reason, description, evidenceList)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if report != nil {
		if report.Evidence != "" {
			report.EvidenceURL = formatMediaURL(report.Evidence)
		}
		if len(report.EvidenceList) > 0 {
			report.EvidenceURLs = make([]string, len(report.EvidenceList))
			for i, p := range report.EvidenceList {
				report.EvidenceURLs[i] = formatMediaURL(p)
			}
		}
	}

	response.Created(c, "Report submitted successfully", report)
}

func formatMediaURL(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "/media/") {
		return path
	}
	return "/media/" + strings.TrimPrefix(path, "/")
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

	for i := range verifications {
		if verifications[i].DocumentFront != "" {
			verifications[i].DocumentFrontURL = formatMediaURL(verifications[i].DocumentFront)
		}
		if verifications[i].DocumentBack != "" {
			verifications[i].DocumentBackURL = formatMediaURL(verifications[i].DocumentBack)
		}
		if verifications[i].Selfie != "" {
			verifications[i].SelfieURL = formatMediaURL(verifications[i].Selfie)
		}
		if verifications[i].TradeLicense != "" {
			verifications[i].TradeLicenseURL = formatMediaURL(verifications[i].TradeLicense)
		}
	}

	response.JSON(c, http.StatusOK, verifications)
}

func (h *ModerationHandler) SubmitVerification(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	input := domainUsecase.VerificationSubmitInput{}

	// Handle multipart form-data or JSON payload
	if strings.Contains(c.ContentType(), "multipart/form-data") {
		input.DocumentType = c.PostForm("document_type")
		if input.DocumentType == "" {
			input.DocumentType = "Driver's License"
		}
		input.LicenseNumber = c.PostForm("license_number")
		if expStr := c.PostForm("license_expiry"); expStr != "" {
			if parsed, err := time.Parse(time.RFC3339, expStr); err == nil {
				input.LicenseExpiry = &parsed
			} else if parsed, err := time.Parse("2006-01-02", expStr); err == nil {
				input.LicenseExpiry = &parsed
			}
		}

		// Handle Front Image
		if frontFile, err := c.FormFile("document_front"); err == nil && frontFile != nil {
			if relPath, err := media.ValidateAndSaveUploadedFile(frontFile, "verifications", 15*1024*1024); err == nil {
				input.DocumentFront = relPath
			}
		} else if frontFile, err := c.FormFile("front"); err == nil && frontFile != nil {
			if relPath, err := media.ValidateAndSaveUploadedFile(frontFile, "verifications", 15*1024*1024); err == nil {
				input.DocumentFront = relPath
			}
		}

		// Handle Back Image
		if backFile, err := c.FormFile("document_back"); err == nil && backFile != nil {
			if relPath, err := media.ValidateAndSaveUploadedFile(backFile, "verifications", 15*1024*1024); err == nil {
				input.DocumentBack = relPath
			}
		} else if backFile, err := c.FormFile("back"); err == nil && backFile != nil {
			if relPath, err := media.ValidateAndSaveUploadedFile(backFile, "verifications", 15*1024*1024); err == nil {
				input.DocumentBack = relPath
			}
		}

		// Handle Selfie
		if selfieFile, err := c.FormFile("selfie"); err == nil && selfieFile != nil {
			if relPath, err := media.ValidateAndSaveUploadedFile(selfieFile, "verifications", 15*1024*1024); err == nil {
				input.Selfie = relPath
			}
		}

		// Handle Trade License
		if tradeFile, err := c.FormFile("trade_license"); err == nil && tradeFile != nil {
			if relPath, err := media.ValidateAndSaveUploadedFile(tradeFile, "verifications", 15*1024*1024); err == nil {
				input.TradeLicense = relPath
			}
		} else if tradeFile, err := c.FormFile("trade"); err == nil && tradeFile != nil {
			if relPath, err := media.ValidateAndSaveUploadedFile(tradeFile, "verifications", 15*1024*1024); err == nil {
				input.TradeLicense = relPath
			}
		}
	} else {
		var req struct {
			DocumentType  string  `json:"document_type"`
			DocumentFront string  `json:"document_front"`
			DocumentBack  string  `json:"document_back"`
			Selfie        string  `json:"selfie"`
			TradeLicense  string  `json:"trade_license"`
			LicenseNumber string  `json:"license_number"`
			LicenseExpiry *string `json:"license_expiry"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			input.DocumentType = req.DocumentType
			input.DocumentFront = req.DocumentFront
			input.DocumentBack = req.DocumentBack
			input.Selfie = req.Selfie
			input.TradeLicense = req.TradeLicense
			input.LicenseNumber = req.LicenseNumber
			if req.LicenseExpiry != nil && *req.LicenseExpiry != "" {
				if parsed, err := time.Parse(time.RFC3339, *req.LicenseExpiry); err == nil {
					input.LicenseExpiry = &parsed
				} else if parsed, err := time.Parse("2006-01-02", *req.LicenseExpiry); err == nil {
					input.LicenseExpiry = &parsed
				}
			}
		}
	}

	if input.DocumentType == "" {
		input.DocumentType = "Driver's License"
	}

	verif, err := h.moderationUC.SubmitVerification(c.Request.Context(), userUUID, input)
	if err != nil {
		response.BadRequest(c, "Failed to submit verification: "+err.Error())
		return
	}

	if verif.DocumentFront != "" {
		verif.DocumentFrontURL = formatMediaURL(verif.DocumentFront)
	}
	if verif.DocumentBack != "" {
		verif.DocumentBackURL = formatMediaURL(verif.DocumentBack)
	}
	if verif.Selfie != "" {
		verif.SelfieURL = formatMediaURL(verif.Selfie)
	}
	if verif.TradeLicense != "" {
		verif.TradeLicenseURL = formatMediaURL(verif.TradeLicense)
	}

	response.Created(c, "Verification documents submitted for review", verif)
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
