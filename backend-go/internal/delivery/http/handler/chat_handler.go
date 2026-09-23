package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/websocket"
	"backend-go/pkg/auth"
	"backend-go/pkg/media"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chatUC domainUsecase.ChatUseCase
	hub    *websocket.Hub
	cfg    *config.Config
}

func NewChatHandler(chatUC domainUsecase.ChatUseCase, hub *websocket.Hub, cfg *config.Config) *ChatHandler {
	h := &ChatHandler{chatUC: chatUC, hub: hub, cfg: cfg}
	if hub != nil {
		hub.OnMessage = h.HandleWsMessage
	}
	return h
}

func (h *ChatHandler) wrapConversations(conversations []entity.Conversation, currentUserID uuid.UUID) []gin.H {
	wrapped := make([]gin.H, 0, len(conversations))
	for _, conv := range conversations {
		var meProfile interface{}
		var otherProfile interface{}
		var otherID string
		var lastMsg interface{}

		if len(conv.Messages) > 0 {
			lastMsg = conv.Messages[len(conv.Messages)-1]
		}

		for _, p := range conv.Participants {
			if p.ID == currentUserID {
				if p.Profile != nil {
					meProfile = p.Profile
				}
			} else {
				otherID = p.ID.String()
				if p.Profile != nil {
					otherProfile = p.Profile
				}
			}
		}

		chatData := gin.H{
			"id":               conv.ID.String(),
			"chat_room":        conv.ID.String(),
			"user1":            currentUserID.String(),
			"user2":            otherID,
			"unread_count":     0,
			"created_at":       conv.CreatedAt,
			"updated_at":       conv.UpdatedAt,
			"last_message":     lastMsg,
			"is_blocked":       false,
			"is_blocked_by_me": false,
		}

		wrapped = append(wrapped, gin.H{
			"chat":  chatData,
			"me":    meProfile,
			"other": otherProfile,
		})
	}
	return wrapped
}

func (h *ChatHandler) wrapMessages(messages []entity.Message, currentUserID uuid.UUID) []gin.H {
	wrapped := make([]gin.H, 0, len(messages))
	for _, m := range messages {
		var senderProfile interface{}
		if m.Sender != nil && m.Sender.Profile != nil {
			senderProfile = m.Sender.Profile
		}

		wrapped = append(wrapped, gin.H{
			"message":  m,
			"sender":   senderProfile,
			"receiver": nil,
		})
	}
	return wrapped
}

func (h *ChatHandler) GetConversations(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	conversations, err := h.chatUC.GetConversations(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to load conversations")
		return
	}
	response.JSON(c, http.StatusOK, h.wrapConversations(conversations, userUUID))
}

func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var req struct {
		RecipientID string `json:"recipient_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "recipient_id is required")
		return
	}

	recipientUUID, err := uuid.Parse(req.RecipientID)
	if err != nil {
		response.BadRequest(c, "invalid recipient ID")
		return
	}

	conv, err := h.chatUC.CreateConversation(c.Request.Context(), userUUID, recipientUUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, conv)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	convIDStr := c.Query("conversation")
	if convIDStr == "" {
		convIDStr = c.Query("conversation_id")
	}
	if convIDStr == "" {
		convIDStr = c.Param("conversation_id")
	}
	convUUID, err := uuid.Parse(convIDStr)
	if err != nil {
		response.BadRequest(c, "valid conversation_id is required")
		return
	}

	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.chatUC.GetMessages(c.Request.Context(), convUUID, userUUID, limit, offset)
	if err != nil || len(messages) == 0 {
		// convUUID might be the other party's user UUID instead of conversation UUID
		conv, err2 := h.chatUC.GetByParticipants(c.Request.Context(), userUUID, convUUID)
		if err2 == nil && conv != nil {
			messages, _ = h.chatUC.GetMessages(c.Request.Context(), conv.ID, userUUID, limit, offset)
		}
	}
	response.JSON(c, http.StatusOK, h.wrapMessages(messages, userUUID))
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var req struct {
		ConversationID string `json:"conversation_id" binding:"required"`
		Content        string `json:"content"`
		Message        string `json:"message"`
		MediaURL       string `json:"media_url"`
		MessageType    string `json:"message_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid message payload")
		return
	}

	convUUID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	text := req.Content
	if text == "" {
		text = req.Message
	}

	msg, err := h.chatUC.SendMessage(c.Request.Context(), convUUID, userUUID, text, req.MediaURL, req.MessageType)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if h.hub != nil {
		h.hub.BroadcastToRoom(req.ConversationID, msg)
	}

	response.JSON(c, http.StatusCreated, msg)
}

func (h *ChatHandler) SetSeen(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var body struct {
		ReceiverID string `json:"receiver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "receiver_id is required")
		return
	}

	receiverUUID, err := uuid.Parse(body.ReceiverID)
	if err != nil {
		response.BadRequest(c, "Invalid receiver_id")
		return
	}

	_ = h.chatUC.SetSeen(c.Request.Context(), userUUID, receiverUUID)
	response.JSON(c, http.StatusOK, gin.H{"status": "success"})
}

func (h *ChatHandler) BlockChat(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var body struct {
		UserID         string  `json:"user_id" binding:"required"`
		ConversationID *string `json:"conversation_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "user_id is required")
		return
	}

	targetUUID, err := uuid.Parse(body.UserID)
	if err != nil {
		response.BadRequest(c, "Invalid user_id")
		return
	}

	var convUUID *uuid.UUID
	if body.ConversationID != nil && *body.ConversationID != "" {
		if cid, err := uuid.Parse(*body.ConversationID); err == nil {
			convUUID = &cid
		}
	}

	_ = h.chatUC.BlockChat(c.Request.Context(), userUUID, targetUUID, convUUID)
	response.JSON(c, http.StatusOK, gin.H{"status": "blocked"})
}

func (h *ChatHandler) UnblockChat(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "user_id is required")
		return
	}

	targetUUID, err := uuid.Parse(body.UserID)
	if err != nil {
		response.BadRequest(c, "Invalid user_id")
		return
	}

	_ = h.chatUC.UnblockChat(c.Request.Context(), userUUID, targetUUID)
	response.JSON(c, http.StatusOK, gin.H{"status": "unblocked"})
}

// UploadChatMedia accepts a multipart/form-data field named "image", "file", "audio", "document", or "media",
// saves it under media/chat/, and returns the relative media_url, file_name, and file_size.
func (h *ChatHandler) UploadChatMedia(c *gin.Context) {
	var fileHeader *multipart.FileHeader
	var err error

	for _, field := range []string{"image", "file", "audio", "document", "media"} {
		fileHeader, err = c.FormFile(field)
		if err == nil && fileHeader != nil {
			break
		}
	}

	if fileHeader == nil {
		response.BadRequest(c, "No media file provided")
		return
	}

	mediaDir := media.ResolveMediaDir("chat")
	cleanBase := filepath.Base(fileHeader.Filename)
	filename := fmt.Sprintf("chat_%d_%s", time.Now().UnixNano(), cleanBase)
	savePath := filepath.Join(mediaDir, filename)
	if err := c.SaveUploadedFile(fileHeader, savePath); err != nil {
		response.InternalError(c, "Failed to save uploaded file")
		return
	}

	mediaURL := "/media/chat/" + filename
	response.JSON(c, http.StatusCreated, gin.H{
		"media_url": mediaURL,
		"file_name": cleanBase,
		"file_size": fileHeader.Size,
	})
}

func (h *ChatHandler) GetBlockedUsers(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	blocks, err := h.chatUC.GetBlockedUsers(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to load blocked users")
		return
	}

	response.JSON(c, http.StatusOK, blocks)
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	roomID := c.Param("conversation_id")
	if roomID == "" {
		roomID = c.Param("appointment_id")
	}

	userID := c.GetString("userID")
	isAdminStream := strings.Contains(c.Request.URL.Path, "/ws/admin/events")

	// If not already set by middleware, authenticate via token param or header
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
			response.Unauthorized(c, "Authentication token required for WebSocket connection")
			c.Abort()
			return
		}

		if h.cfg != nil {
			claims, err := auth.ValidateToken(token, h.cfg.JWTSecret)
			if err != nil {
				response.Unauthorized(c, "Invalid or expired WebSocket authentication token")
				c.Abort()
				return
			}
			userID = claims.UserID
			c.Set("userID", claims.UserID)
			c.Set("userType", claims.UserType)
			c.Set("isStaff", claims.IsStaff)
			c.Set("isSuperuser", claims.IsSuperuser)

			if isAdminStream && !claims.IsStaff && !claims.IsSuperuser {
				response.Forbidden(c, "Admin privileges required for admin event stream")
				c.Abort()
				return
			}
		} else {
			userID = token
		}
	} else if isAdminStream {
		isStaff := c.GetBool("isStaff")
		isSuperuser := c.GetBool("isSuperuser")
		if !isStaff && !isSuperuser {
			response.Forbidden(c, "Admin privileges required for admin event stream")
			c.Abort()
			return
		}
	}

	if userID == "" {
		response.Unauthorized(c, "Unauthenticated WebSocket connection")
		c.Abort()
		return
	}

	websocket.ServeWs(h.hub, c.Writer, c.Request, userID, roomID)
}

func (h *ChatHandler) HandleWsMessage(client *websocket.Client, rawMsg []byte) bool {
	var payload struct {
		Type        string `json:"type"`
		ChatRoomID  string `json:"chat_room_id"`
		Receiver    string `json:"receiver"`
		ReceiverID  string `json:"receiver_id"`
		Sender      string `json:"sender"`
		Message     string `json:"message"`
		Content     string `json:"content"`
		WithImage   bool   `json:"with_image"`
		Image       string `json:"image"`
		MediaURL    string `json:"media_url"`
		FileName    string `json:"filename"`
		FileNameAlt string `json:"file_name"`
	}

	if err := json.Unmarshal(rawMsg, &payload); err != nil {
		return false
	}

	text := payload.Message
	if text == "" {
		text = payload.Content
	}
	media := payload.MediaURL
	if media == "" {
		media = payload.Image
	}

	// Real-time live location event streaming
	if payload.Type == "live_location_update" || payload.Type == "stop_live_location" {
		var genericMap map[string]interface{}
		if err := json.Unmarshal(rawMsg, &genericMap); err == nil {
			if client.RoomID != "" {
				h.hub.BroadcastToRoom(client.RoomID, genericMap)
			}
			targetRecv := payload.Receiver
			if targetRecv == "" {
				targetRecv = payload.ReceiverID
			}
			if targetRecv != "" {
				h.hub.BroadcastToUser(targetRecv, genericMap)
			}
		}
		return true
	}

	// Ignore non-chat payloads (e.g. status/typing/presence which are handled separately)
	if text == "" && media == "" {
		return false
	}

	senderIDStr := client.UserID
	if senderIDStr == "" {
		senderIDStr = payload.Sender
	}
	senderUUID, err := uuid.Parse(senderIDStr)
	if err != nil {
		return false
	}

	receiverIDStr := payload.Receiver
	if receiverIDStr == "" {
		receiverIDStr = payload.ReceiverID
	}

	// If receiver is not in the JSON body, attempt extraction from roomID (format: "uid1_uid2")
	if receiverIDStr == "" && client.RoomID != "" && strings.Contains(client.RoomID, "_") {
		parts := strings.Split(client.RoomID, "_")
		if len(parts) == 2 {
			if parts[0] == senderIDStr {
				receiverIDStr = parts[1]
			} else {
				receiverIDStr = parts[0]
			}
		}
	}

	receiverUUID, err := uuid.Parse(receiverIDStr)
	if err != nil {
		return false
	}

	ctx := context.Background()

	// 1. Ensure conversation exists in DB
	conv, err := h.chatUC.CreateConversation(ctx, senderUUID, receiverUUID)
	if err != nil || conv == nil {
		return false
	}

	fileName := payload.FileName
	if fileName == "" {
		fileName = payload.FileNameAlt
	}

	isImage := payload.WithImage || (media != "" && (strings.HasSuffix(strings.ToLower(media), ".jpg") || strings.HasSuffix(strings.ToLower(media), ".jpeg") || strings.HasSuffix(strings.ToLower(media), ".png") || strings.HasSuffix(strings.ToLower(media), ".webp")))

	// 2. Persist message to database
	msgEntity := entity.Message{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		SenderID:       senderUUID,
		Content:        text,
		Message:        text,
		MediaURL:       media,
		Image:          media,
		FileName:       fileName,
		WithImage:      isImage,
		IsSeen:         false,
		IsDelivered:    true,
		CreatedAt:      time.Now(),
	}

	savedMsg, err := h.chatUC.SaveFullMessage(ctx, &msgEntity)
	if err != nil || savedMsg == nil {
		return false
	}

	// 3. Attach sender and receiver profiles for Flutter parsing
	var senderProfile interface{}
	var receiverProfile interface{}
	for _, p := range conv.Participants {
		if p.ID == senderUUID && p.Profile != nil {
			senderProfile = p.Profile
		} else if p.ID == receiverUUID && p.Profile != nil {
			receiverProfile = p.Profile
		}
	}

	// 4. Build standard ChatMessage envelope that Flutter expects
	broadcastPayload := gin.H{
		"type": "message",
		"message": gin.H{
			"id":           savedMsg.ID.String(),
			"chat_room_id": client.RoomID,
			"sender":       senderUUID.String(),
			"receiver":     receiverUUID.String(),
			"message":      savedMsg.Message,
			"with_image":   savedMsg.WithImage,
			"image":        savedMsg.Image,
			"media_url":    savedMsg.MediaURL,
			"file_name":    savedMsg.FileName,
			"is_delivered": true,
			"read":         false,
			"created_at":   savedMsg.CreatedAt.Format(time.RFC3339Nano),
			"updated_at":   savedMsg.CreatedAt.Format(time.RFC3339Nano),
		},
		"sender":   senderProfile,
		"receiver": receiverProfile,
	}

	// 5. Broadcast to room and deliver directly to both users
	if client.RoomID != "" {
		h.hub.BroadcastToRoom(client.RoomID, broadcastPayload)
	}
	h.hub.BroadcastToUser(receiverUUID.String(), broadcastPayload)
	h.hub.BroadcastToUser(senderUUID.String(), broadcastPayload)

	return true
}
