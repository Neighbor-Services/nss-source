package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"backend-go/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Mobile apps or non-browser clients may not supply Origin
		if origin == "" {
			return true
		}
		// Allow local development and standard allowed origins
		if strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:") ||
			strings.HasSuffix(origin, "neighborservice.com") ||
			strings.HasSuffix(origin, "neighborservices.com") {
			return true
		}
		// Check against custom ALLOWED_ORIGINS if configured
		if customOrigins := os.Getenv("ALLOWED_ORIGINS"); customOrigins != "" {
			for _, o := range strings.Split(customOrigins, ",") {
				if strings.TrimSpace(o) == origin || strings.TrimSpace(o) == "*" {
					return true
				}
			}
		}
		return true // fallback permits mobile Flutter clients
	},
}

type Client struct {
	UserID string
	RoomID string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

type RedisWSEvent struct {
	InstanceID string          `json:"instance_id"`
	Type       string          `json:"type"` // "broadcast", "room", "user"
	Target     string          `json:"target"`
	Payload    json.RawMessage `json:"payload"`
}

type Hub struct {
	// Registered clients: map[userID]map[*Client]bool
	clients     map[string]map[*Client]bool
	roomClients map[string]map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mu          sync.RWMutex
	cacheClient cache.Cache
	instanceID  string
	OnMessage   func(client *Client, message []byte) bool
}

var GlobalHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]map[*Client]bool),
		roomClients: make(map[string]map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		instanceID:  uuid.New().String(),
	}
}

// AttachCache enables distributed clustering across multiple server instances via Redis Pub/Sub
func (h *Hub) AttachCache(c cache.Cache) {
	if c == nil {
		return
	}
	h.mu.Lock()
	h.cacheClient = c
	h.mu.Unlock()
	go h.listenRedis()
}

func (h *Hub) listenRedis() {
	if h.cacheClient == nil {
		return
	}
	ctx := context.Background()
	pubsub := h.cacheClient.Subscribe(ctx, "ns:ws:events")
	if pubsub == nil {
		return
	}
	defer pubsub.Close()
	log.Printf("⚡ [WebSocket Hub] Redis Pub/Sub cluster listener active (Instance: %s)", h.instanceID)

	ch := pubsub.Channel()
	for msg := range ch {
		var event RedisWSEvent
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			continue
		}
		if event.InstanceID == h.instanceID {
			continue // Skip messages published by this instance
		}
		switch event.Type {
		case "broadcast":
			h.deliverLocalBroadcast(event.Payload)
		case "room":
			h.deliverLocalRoom(event.Target, event.Payload)
		case "user":
			h.deliverLocalUser(event.Target, event.Payload)
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Register user
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true

			// Register room if any
			if client.RoomID != "" {
				if _, ok := h.roomClients[client.RoomID]; !ok {
					h.roomClients[client.RoomID] = make(map[*Client]bool)
				}
				h.roomClients[client.RoomID][client] = true
			}

			// Handle presence notifications
			var otherUserID string
			if client.RoomID != "" && strings.Contains(client.RoomID, "_") {
				parts := strings.Split(client.RoomID, "_")
				if len(parts) == 2 {
					if parts[0] == client.UserID {
						otherUserID = parts[1]
					} else {
						otherUserID = parts[0]
					}
				}
			}

			var isOtherOnline bool
			if otherUserID != "" {
				if conns, ok := h.clients[otherUserID]; ok && len(conns) > 0 {
					isOtherOnline = true
				}
			}
			h.mu.Unlock()

			// Send other user presence status to the newly connected client
			if otherUserID != "" {
				status := "offline"
				if isOtherOnline {
					status = "online"
				}
				client.SendJSON(map[string]interface{}{
					"type":    "presence",
					"user_id": otherUserID,
					"status":  status,
				})

				// Notify the other user (and room) that this client is online
				h.SendToRoom(client.RoomID, map[string]interface{}{
					"type":    "presence",
					"user_id": client.UserID,
					"status":  "online",
				})
			}

			// Broadcast presence to global presence stream
			h.SendToRoom("presence", map[string]interface{}{
				"type":    "presence",
				"user_id": client.UserID,
				"status":  "online",
			})

		case client := <-h.unregister:
			h.mu.Lock()
			isUserFullyOffline := false
			if userConns, ok := h.clients[client.UserID]; ok {
				if _, exists := userConns[client]; exists {
					delete(userConns, client)
					close(client.Send)
					if len(userConns) == 0 {
						delete(h.clients, client.UserID)
						isUserFullyOffline = true
					}
				}
			}
			if client.RoomID != "" {
				if roomConns, ok := h.roomClients[client.RoomID]; ok {
					delete(roomConns, client)
					if len(roomConns) == 0 {
						delete(h.roomClients, client.RoomID)
					}
				}
			}
			h.mu.Unlock()

			if isUserFullyOffline {
				// Broadcast offline status to the room and global presence
				if client.RoomID != "" {
					h.SendToRoom(client.RoomID, map[string]interface{}{
						"type":    "presence",
						"user_id": client.UserID,
						"status":  "offline",
					})
				}
				h.SendToRoom("presence", map[string]interface{}{
					"type":    "presence",
					"user_id": client.UserID,
					"status":  "offline",
				})
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, userConns := range h.clients {
				for client := range userConns {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(userConns, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if conns, ok := h.clients[userID]; ok && len(conns) > 0 {
		return true
	}
	return false
}

func (c *Client) SendJSON(payload interface{}) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return
	}
	select {
	case c.Send <- bytes:
	default:
	}
}

func (h *Hub) deliverLocalUser(userID string, bytes []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if userConns, ok := h.clients[userID]; ok {
		for client := range userConns {
			select {
			case client.Send <- bytes:
			default:
				log.Printf("Failed to send message to user %s", userID)
			}
		}
	}
}

func (h *Hub) deliverLocalRoom(roomID string, bytes []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomConns, ok := h.roomClients[roomID]; ok {
		for client := range roomConns {
			select {
			case client.Send <- bytes:
			default:
				log.Printf("Failed to send message to room %s", roomID)
			}
		}
	}
}

func (h *Hub) deliverLocalBroadcast(bytes []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userConns := range h.clients {
		for client := range userConns {
			select {
			case client.Send <- bytes:
			default:
			}
		}
	}
}

func (h *Hub) SendToUser(userID string, payload interface{}) {
	var bytes []byte
	switch v := payload.(type) {
	case []byte:
		bytes = v
	case json.RawMessage:
		bytes = []byte(v)
	default:
		var err error
		bytes, err = json.Marshal(payload)
		if err != nil {
			return
		}
	}

	h.deliverLocalUser(userID, bytes)

	if h.cacheClient != nil {
		event := RedisWSEvent{
			InstanceID: h.instanceID,
			Type:       "user",
			Target:     userID,
			Payload:    bytes,
		}
		_ = h.cacheClient.Publish(context.Background(), "ns:ws:events", event)
	}
}

func (h *Hub) SendToRoom(roomID string, payload interface{}) {
	var bytes []byte
	switch v := payload.(type) {
	case []byte:
		bytes = v
	case json.RawMessage:
		bytes = []byte(v)
	default:
		var err error
		bytes, err = json.Marshal(payload)
		if err != nil {
			return
		}
	}

	h.deliverLocalRoom(roomID, bytes)

	if h.cacheClient != nil {
		event := RedisWSEvent{
			InstanceID: h.instanceID,
			Type:       "room",
			Target:     roomID,
			Payload:    bytes,
		}
		_ = h.cacheClient.Publish(context.Background(), "ns:ws:events", event)
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(message, &payload); err == nil {
			msgType, _ := payload["type"].(string)

			if msgType == "get_presence" {
				targetUID, _ := payload["user_id"].(string)
				status := "offline"
				if c.Hub.IsUserOnline(targetUID) {
					status = "online"
				}
				c.SendJSON(map[string]interface{}{
					"type":    "presence",
					"user_id": targetUID,
					"status":  status,
				})
				continue
			}

			if msgType == "typing" {
				if _, exists := payload["user_id"]; !exists || payload["user_id"] == "" {
					payload["user_id"] = c.UserID
				}
				if c.RoomID != "" {
					c.Hub.SendToRoom(c.RoomID, payload)
				}
				continue
			}

			if msgType == "delete_message" || msgType == "update_message" || msgType == "message_status" {
				if c.RoomID != "" {
					c.Hub.SendToRoom(c.RoomID, json.RawMessage(message))
				}
				continue
			}

			if c.Hub.OnMessage != nil && c.Hub.OnMessage(c, message) {
				continue
			}

			if c.RoomID != "" {
				c.Hub.SendToRoom(c.RoomID, json.RawMessage(message))
			}
		} else {
			if c.Hub.OnMessage != nil && c.Hub.OnMessage(c, message) {
				continue
			}
			if c.RoomID != "" {
				c.Hub.SendToRoom(c.RoomID, json.RawMessage(message))
			}
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)
		if err := w.Close(); err != nil {
			return
		}
	}
}

func (h *Hub) BroadcastToRoom(roomID string, payload interface{}) {
	h.SendToRoom(roomID, payload)
}

func (h *Hub) BroadcastToUser(userID string, payload interface{}) {
	h.SendToUser(userID, payload)
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID, roomID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		UserID: userID,
		RoomID: roomID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func HandleWebSocket(c *gin.Context, hub *Hub, userID, roomID string) {
	ServeWs(hub, c.Writer, c.Request, userID, roomID)
}

