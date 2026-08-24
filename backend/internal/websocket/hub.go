package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/aicenter/aicenter/internal/auth"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 65536
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	rooms      map[string]map[*Client]bool
	mu         sync.RWMutex
	log        *zap.Logger
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	rooms    map[string]bool
	mu       sync.Mutex
	clientID string
	userID   string
}

// Message is the WebSocket message envelope
type Message struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Channel   string          `json:"channel"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// NewHub creates a new Hub
func NewHub(log *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
		log:        log,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.log.Info("WebSocket client connected", zap.String("clientID", client.clientID))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				// Remove from all rooms
				for room := range client.rooms {
					if clients, ok := h.rooms[room]; ok {
						delete(clients, client)
					}
				}
				h.log.Info("WebSocket client disconnected", zap.String("clientID", client.clientID))
			}
		}
	}
}

// JoinRoom adds a client to a room
func (h *Hub) JoinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true

	client.mu.Lock()
	client.rooms[room] = true
	client.mu.Unlock()
}

// LeaveRoom removes a client from a room
func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[room]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}

	client.mu.Lock()
	delete(client.rooms, room)
	client.mu.Unlock()
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(room string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[room]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(clients, client)
			}
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// ServeWs handles websocket requests from the client, authenticating via
// a JWT access token passed as the ?token= query parameter.
func ServeWs(hub *Hub, jwtSecret string, w http.ResponseWriter, r *http.Request) {
	// Require a valid JWT access token via query param.
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ValidateAccessToken(tokenStr, jwtSecret)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.log.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		rooms:    make(map[string]bool),
		clientID: generateClientID(),
		userID:   claims.UserID,
	}

	hub.log.Info("WebSocket client connected",
		zap.String("clientID", client.clientID),
		zap.String("userID", claims.UserID),
		zap.String("username", claims.Username),
	)

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.log.Warn("WebSocket error", zap.Error(err))
			}
			break
		}

		// Handle incoming message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.hub.log.Warn("Invalid WebSocket message", zap.Error(err))
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case "subscribe":
		var data struct {
			Channels []string `json:"channels"`
		}
		if err := json.Unmarshal(msg.Data, &data); err == nil {
			for _, ch := range data.Channels {
				c.hub.JoinRoom(c, ch)
			}
		}

	case "unsubscribe":
		var data struct {
			Channels []string `json:"channels"`
		}
		if err := json.Unmarshal(msg.Data, &data); err == nil {
			for _, ch := range data.Channels {
				c.hub.LeaveRoom(c, ch)
			}
		}

	case "ping":
		resp, _ := json.Marshal(Message{
			Type:      "pong",
			Timestamp: time.Now(),
		})
		c.send <- resp

	default:
		c.hub.log.Debug("Unknown message type", zap.String("type", msg.Type))
	}
}

func generateClientID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
