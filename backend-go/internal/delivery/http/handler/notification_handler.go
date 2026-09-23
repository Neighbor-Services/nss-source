package handler

import (
	"net/http"
	"strconv"
	"strings"

	"backend-go/internal/config"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/websocket"
	"backend-go/pkg/auth"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notifUC domainUsecase.NotificationUseCase
	hub     *websocket.Hub
	cfg     *config.Config
}

func NewNotificationHandler(notifUC domainUsecase.NotificationUseCase, hub *websocket.Hub, cfg *config.Config) *NotificationHandler {
	return &NotificationHandler{notifUC: notifUC, hub: hub, cfg: cfg}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	notifications, err := h.notifUC.GetNotifications(c.Request.Context(), userUUID, limit)
	if err != nil {
		response.InternalError(c, "Failed to load notifications")
		return
	}
	response.JSON(c, http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	notifIDStr := c.Param("id")
	notifUUID, err := uuid.Parse(notifIDStr)
	if err != nil {
		response.BadRequest(c, "invalid notification ID")
		return
	}

	if err := h.notifUC.MarkAsRead(c.Request.Context(), notifUUID, userUUID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "notification marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	if err := h.notifUC.MarkAllAsRead(c.Request.Context(), userUUID); err != nil {
		response.InternalError(c, "Failed to mark all notifications as read")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "all notifications marked as read"})
}

type DeviceTokenRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform"`
	DeviceID string `json:"device_id"`
}

func (h *NotificationHandler) RegisterDeviceToken(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var req DeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Token is required")
		return
	}

	if req.Platform == "" {
		req.Platform = "mobile"
	}

	if err := h.notifUC.RegisterDeviceToken(c.Request.Context(), userUUID, req.Token, req.Platform, req.DeviceID); err != nil {
		response.InternalError(c, "Failed to register device token")
		return
	}

	response.Created(c, "Token registered successfully", gin.H{"status": "registered"})
}

func (h *NotificationHandler) GetDeviceTokens(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	tokens, err := h.notifUC.GetDeviceTokens(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to load tokens")
		return
	}
	response.JSON(c, http.StatusOK, tokens)
}

func (h *NotificationHandler) UnregisterDeviceToken(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var req struct {
		Token string `json:"token"`
	}
	_ = c.ShouldBindJSON(&req)
	token := req.Token
	if token == "" {
		token = c.Query("token")
	}

	if token == "" {
		response.BadRequest(c, "Token is required")
		return
	}

	if err := h.notifUC.UnregisterDeviceToken(c.Request.Context(), userUUID, token); err != nil {
		response.InternalError(c, "Failed to unregister device token")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "unregistered"})
}


func (h *NotificationHandler) HandleWebSocket(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		token := c.Query("token")
		if token == "" {
			token = c.GetHeader("Sec-WebSocket-Protocol")
			if token == "" {
				authH := c.GetHeader("Authorization")
				if strings.HasPrefix(authH, "Bearer ") {
					token = strings.TrimPrefix(authH, "Bearer ")
				}
			}
		}

		if token == "" {
			response.Unauthorized(c, "Authentication token required for notifications WebSocket")
			c.Abort()
			return
		}

		if h.cfg != nil {
			claims, err := auth.ValidateToken(token, h.cfg.JWTSecret)
			if err != nil {
				response.Unauthorized(c, "Invalid or expired notification WebSocket token")
				c.Abort()
				return
			}
			userID = claims.UserID
			c.Set("userID", claims.UserID)
		} else {
			userID = token
		}
	}

	if userID == "" {
		response.Unauthorized(c, "Unauthenticated WebSocket connection")
		c.Abort()
		return
	}

	websocket.ServeWs(h.hub, c.Writer, c.Request, userID, "notifications")
}
