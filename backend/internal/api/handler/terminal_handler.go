package handler

import (
	"encoding/json"
	"net/http"

	"github.com/aicenter/aicenter/internal/auth"
	"github.com/aicenter/aicenter/internal/terminal"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var terminalUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// TerminalHandler exposes REST session control + the WebSocket bridge.
type TerminalHandler struct {
	mgr       *terminal.Manager
	log       *zap.Logger
	jwtSecret string
}

func NewTerminalHandler(mgr *terminal.Manager, log *zap.Logger, jwtSecret string) *TerminalHandler {
	return &TerminalHandler{mgr: mgr, log: log, jwtSecret: jwtSecret}
}

// CreateSession POST /terminal/sessions  {server_id?, shell?, cols?, rows?}
func (h *TerminalHandler) CreateSession(c *gin.Context) {
	var body struct {
		ServerID string `json:"server_id"`
		Shell    string `json:"shell"`
		Cols     int    `json:"cols"`
		Rows     int    `json:"rows"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respBad(c, "invalid body: "+err.Error())
		return
	}
	sess, err := h.mgr.Create(body.ServerID, body.Shell, body.Cols, body.Rows)
	if err != nil {
		respBad(c, "failed to start terminal: "+err.Error())
		return
	}
	respOK(c, gin.H{"session_id": sess.ID, "command": sess.Command})
}

// ListSessions GET /terminal/sessions
func (h *TerminalHandler) ListSessions(c *gin.Context) {
	respOK(c, gin.H{"items": h.mgr.List()})
}

// CloseSession POST /terminal/sessions/:id/close
func (h *TerminalHandler) CloseSession(c *gin.Context) {
	h.mgr.Remove(c.Param("id"))
	respOK(c, gin.H{"closed": c.Param("id")})
}

// Bridge ws://.../ws/terminal?session=<id>&token=<jwt>  streams PTY in/out.
// Requires a valid JWT access token via the ?token= query parameter.
func (h *TerminalHandler) Bridge(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ValidateAccessToken(tokenStr, h.jwtSecret)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "session required", http.StatusBadRequest)
		return
	}
	sess, ok := h.mgr.Get(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	h.log.Info("terminal ws bridge opened",
		zap.String("session", sessionID),
		zap.String("userID", claims.UserID),
		zap.String("username", claims.Username),
	)

	conn, err := terminalUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Warn("terminal ws upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	// PTY output → WS
	go sess.ReadLoop(func(b []byte) {
		_ = conn.WriteMessage(websocket.TextMessage, b)
	})
	// WS input → PTY, and control messages (resize/close)
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if mt == websocket.TextMessage {
			var m terminal.Message
			if json.Unmarshal(msg, &m) == nil {
				switch m.Type {
				case "input":
					_, _ = sess.Write([]byte(m.Data))
				case "resize":
					_ = sess.Resize(m.Cols, m.Rows)
				case "close":
					sess.Close()
					return
				}
			} else {
				// raw text treated as input
				_, _ = sess.Write(msg)
			}
		}
	}
	// Graceful cleanup after the client disconnects.
	h.mgr.Remove(sessionID)
}

func respOK(c *gin.Context, v interface{})  { c.JSON(http.StatusOK, v) }
func respBad(c *gin.Context, msg string)    { c.JSON(http.StatusBadRequest, gin.H{"error": msg}) }
