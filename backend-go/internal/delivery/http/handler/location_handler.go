package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"backend-go/internal/config"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/websocket"
	"backend-go/pkg/auth"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LocationHandler handles the dedicated WebSocket endpoint for real-time location streaming.
// Clients connect once and send location_update frames as their position changes.
type LocationHandler struct {
	profileUC domainUsecase.ProfileUseCase
	hub       *websocket.Hub
	cfg       *config.Config
}

func NewLocationHandler(profileUC domainUsecase.ProfileUseCase, hub *websocket.Hub, cfg *config.Config) *LocationHandler {
	return &LocationHandler{profileUC: profileUC, hub: hub, cfg: cfg}
}

// locationUpdateMsg is the JSON frame sent by the mobile client.
//
//	{"type":"location_update","lat":37.77,"lng":-122.41,"city":"San Francisco","state":"CA","country":"US","address":"Market St"}
type locationUpdateMsg struct {
	Type    string  `json:"type"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	City    string  `json:"city"`
	State   string  `json:"state"`
	Country string  `json:"country"`
	Address string  `json:"address"`
}

// ServeWsLocation upgrades the connection to WebSocket and streams location updates.
// Authentication is done via ?token= query parameter because browser/Flutter WebSocket
// clients cannot attach custom Authorization headers during the initial HTTP Upgrade.
//
// Route: GET /api/v1/accounts/location/ws/
func (h *LocationHandler) ServeWsLocation(c *gin.Context) {
	// ── 1. Authenticate via query-string token ────────────────────────────────
	tokenStr := c.Query("token")
	if tokenStr == "" {
		// Also accept Bearer in the header if the client sends it before upgrade
		if hdr := c.GetHeader("Authorization"); strings.HasPrefix(hdr, "Bearer ") {
			tokenStr = strings.TrimPrefix(hdr, "Bearer ")
		}
	}
	if tokenStr == "" {
		response.Unauthorized(c, "Missing token")
		return
	}

	claims, err := auth.ValidateToken(tokenStr, h.cfg.JWTSecret)
	if err != nil {
		response.Unauthorized(c, "Invalid or expired token")
		return
	}

	userUUID, err := uuid.Parse(claims.UserID)
	if err != nil || userUUID == uuid.Nil {
		response.Unauthorized(c, "Invalid user ID in token")
		return
	}

	// ── 2. Upgrade to WebSocket ───────────────────────────────────────────────
	conn, err := websocket.UpgradeConn(c.Writer, c.Request)
	if err != nil {
		log.Printf("[LocationWS] Upgrade failed for user %s: %v", userUUID, err)
		return
	}
	defer conn.Close()

	log.Printf("[LocationWS] Client connected: userID=%s", userUUID)
	conn.SetReadDeadline(time.Time{}) // no global read deadline — location updates are periodic

	// ── 3. Read pump — process location frames ────────────────────────────────
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			// Normal closure or network error
			log.Printf("[LocationWS] Disconnected: userID=%s err=%v", userUUID, err)
			break
		}

		var msg locationUpdateMsg
		if jsonErr := json.Unmarshal(raw, &msg); jsonErr != nil {
			log.Printf("[LocationWS] Bad JSON from %s: %v", userUUID, jsonErr)
			continue
		}

		if msg.Type != "location_update" {
			// Ignore unknown frame types (heartbeat, ping, etc.)
			continue
		}

		// ── 4. Persist location asynchronously so we don't block the read loop ──
		go func(uid uuid.UUID, m locationUpdateMsg) {
			updates := map[string]interface{}{
				"latitude":  m.Lat,
				"longitude": m.Lng,
			}
			if m.City != "" {
				updates["city"] = m.City
			}
			if m.State != "" {
				updates["state"] = m.State
			}
			if m.Country != "" {
				updates["country"] = m.Country
			}
			if m.Address != "" {
				updates["address"] = m.Address
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if _, err := h.profileUC.UpdateProfile(ctx, uid, updates); err != nil {
				log.Printf("[LocationWS] UpdateProfile failed for %s: %v", uid, err)
			}
		}(userUUID, msg)

		// Acknowledge receipt back to the client
		_ = conn.WriteMessage(1, []byte(`{"type":"location_ack"}`))
	}
}

// HealthWs is a lightweight endpoint the client can ping to verify the WS server is reachable.
// Route: GET /api/v1/accounts/location/ws/health
func (h *LocationHandler) HealthWs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "endpoint": "location_ws"})
}
