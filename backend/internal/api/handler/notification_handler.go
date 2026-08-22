package handler

import (
	"encoding/json"

	"github.com/aicenter/aicenter/internal/api/response"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/service"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct{ svc *service.NotificationService }

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) RegisterRoutes(group gin.IRouter, auth gin.HandlerFunc) {
	g := group.(*gin.RouterGroup)
	g.GET("/notification/channels", auth, h.ListChannels)
	g.POST("/notification/channels", auth, h.CreateChannel)
	g.PUT("/notification/channels/:id", auth, h.UpdateChannel)
	g.DELETE("/notification/channels/:id", auth, h.DeleteChannel)

	g.GET("/notification/templates", auth, h.ListTemplates)
	g.GET("/notification/templates/:id", auth, h.GetTemplate)
	g.POST("/notification/templates", auth, h.CreateTemplate)
	g.PUT("/notification/templates/:id", auth, h.UpdateTemplate)
	g.DELETE("/notification/templates/:id", auth, h.DeleteTemplate)

	g.GET("/notification/delivery-logs", auth, h.ListDeliveryLogs)

	g.POST("/notification/send", auth, h.SendTest)
}

// ---- channels ----

func (h *NotificationHandler) ListChannels(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"
	items, err := h.svc.ListChannels(enabledOnly)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

func (h *NotificationHandler) CreateChannel(c *gin.Context) {
	var ch models.NotificationChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if ch.Name == "" || ch.Type == "" {
		response.BadRequest(c, "name and type are required")
		return
	}
	if ch.Config == "" {
		ch.Config = "{}"
	}
	if err := h.svc.CreateChannel(&ch); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, ch)
}

func (h *NotificationHandler) UpdateChannel(c *gin.Context) {
	var ch models.NotificationChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updated, err := h.svc.UpdateChannel(c.Param("id"), &ch)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, updated)
}

func (h *NotificationHandler) DeleteChannel(c *gin.Context) {
	if err := h.svc.DeleteChannel(c.Param("id")); err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, gin.H{"deleted": c.Param("id")})
}

// ---- templates ----

func (h *NotificationHandler) ListTemplates(c *gin.Context) {
	items, err := h.svc.ListTemplates(c.Query("event_type"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

func (h *NotificationHandler) GetTemplate(c *gin.Context) {
	t, err := h.svc.GetTemplate(c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, t)
}

func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	var t models.NotificationTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if t.Name == "" || t.EventType == "" || t.Body == "" {
		response.BadRequest(c, "name, event_type and body are required")
		return
	}
	if t.Channels == "" {
		t.Channels = "[]"
	}
	if err := h.svc.CreateTemplate(&t); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, t)
}

func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	var t models.NotificationTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	updated, err := h.svc.UpdateTemplate(c.Param("id"), &t)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, updated)
}

func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	if err := h.svc.DeleteTemplate(c.Param("id")); err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, gin.H{"deleted": c.Param("id")})
}

// ---- delivery logs ----

func (h *NotificationHandler) ListDeliveryLogs(c *gin.Context) {
	limit := queryInt(c, "limit", 200)
	items, err := h.svc.ListDeliveryLogs(c.Query("status"), limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

// SendTest POST /notification/send  {event_type, title, severity, message, data?, channel_ids?}
func (h *NotificationHandler) SendTest(c *gin.Context) {
	var body struct {
		EventType string            `json:"event_type"`
		Title     string            `json:"title"`
		Severity  string            `json:"severity"`
		Message   string            `json:"message"`
		Data      map[string]string `json:"data"`
		ChannelIDs []string         `json:"channel_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if body.EventType == "" || body.Title == "" {
		response.BadRequest(c, "event_type and title are required")
		return
	}
	if err := h.svc.Notify(body.EventType, body.Title, body.Severity, body.Message, body.Data, body.ChannelIDs); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"dispatched": true})
}

func queryInt(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	var n int
	if json.Unmarshal([]byte(v), &n) == nil {
		return n
	}
	return def
}
