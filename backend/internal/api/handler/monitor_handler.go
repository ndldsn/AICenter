package handler

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/aicenter/aicenter/internal/api/response"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/service"
	"github.com/gin-gonic/gin"
)

type MonitorHandler struct{ svc *service.MonitorService }

func NewMonitorHandler(svc *service.MonitorService) *MonitorHandler { return &MonitorHandler{svc: svc} }

func (h *MonitorHandler) RegisterRoutes(group gin.IRouter, auth gin.HandlerFunc) {
	g := group.(*gin.RouterGroup)
	g.GET("/monitor/metrics", auth, h.QueryMetrics)
	g.GET("/monitor/metrics/latest", auth, h.LatestMetrics)
	g.POST("/monitor/metrics/ingest", auth, h.IngestMetrics)
	g.GET("/monitor/alert-rules", auth, h.ListRules)
	g.POST("/monitor/alert-rules", auth, h.CreateRule)
	g.PUT("/monitor/alert-rules/:id", auth, h.UpdateRule)
	g.DELETE("/monitor/alert-rules/:id", auth, h.DeleteRule)
	g.GET("/monitor/alerts", auth, h.ListAlerts)
	g.POST("/monitor/alerts/:id/ack", auth, h.AckAlert)
}

func parseSince(c *gin.Context) *time.Time {
	s := c.Query("since")
	if s == "" {
		return nil
	}
	// accept RFC3339 or relative minutes like "30m" / "2h"
	if d, err := time.ParseDuration(s); err == nil {
		t := time.Now().UTC().Add(-d)
		return &t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	return nil
}

// QueryMetrics GET /monitor/metrics?server_id=&name=&since=&limit=&aggregate=
func (h *MonitorHandler) QueryMetrics(c *gin.Context) {
	serverID := c.Query("server_id")
	name := c.Query("name")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "500"))
	since := parseSince(c)

	if agg := c.Query("aggregate"); agg != "" {
		if strings.TrimSpace(name) == "" {
			response.BadRequest(c, "name is required when aggregate is set")
			return
		}
		points, err := h.svc.QueryAggregate(serverID, name, agg, since, limit)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		response.Success(c, gin.H{"items": points, "interval": agg})
		return
	}

	items, err := h.svc.QueryMetrics(serverID, name, since, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

// LatestMetrics GET /monitor/metrics/latest?server_id=
func (h *MonitorHandler) LatestMetrics(c *gin.Context) {
	items, err := h.svc.QueryLatest(c.Query("server_id"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

// IngestMetrics POST /monitor/metrics/ingest  {metrics: [...]}
func (h *MonitorHandler) IngestMetrics(c *gin.Context) {
	var body struct {
		Metrics []models.Metric `json:"metrics"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Metrics) == 0 {
		response.BadRequest(c, "metrics array required")
		return
	}
	if len(body.Metrics) > 1000 {
		response.BadRequest(c, "at most 1000 metrics per batch")
		return
	}
	if err := h.svc.Ingest(body.Metrics); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"accepted": len(body.Metrics)})
}

// ListRules GET /monitor/alert-rules?enabled=
func (h *MonitorHandler) ListRules(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"
	items, err := h.svc.ListRules(enabledOnly)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

func ruleFromBody(c *gin.Context) (*models.AlertRule, bool) {
	body, err := c.GetRawData()
	if err != nil {
		response.BadRequest(c, err.Error())
		return nil, false
	}
	var r models.AlertRule
	if err := json.Unmarshal(body, &r); err != nil {
		response.BadRequest(c, err.Error())
		return nil, false
	}
	// distinguish explicit "is_enabled": false from an omitted field
	var raw map[string]json.RawMessage
	if json.Unmarshal(body, &raw) == nil {
		if _, ok := raw["is_enabled"]; ok {
			r.IsEnabledSet = true
		}
	}
	return &r, true
}

func (h *MonitorHandler) CreateRule(c *gin.Context) {
	rule, ok := ruleFromBody(c)
	if !ok {
		return
	}
	if rule.Name == "" || rule.MetricName == "" {
		response.BadRequest(c, "name and metric_name are required")
		return
	}
	if err := h.svc.CreateRule(rule); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rule)
}

func (h *MonitorHandler) UpdateRule(c *gin.Context) {
	rule, ok := ruleFromBody(c)
	if !ok {
		return
	}
	updated, err := h.svc.UpdateRule(c.Param("id"), rule)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, updated)
}

func (h *MonitorHandler) DeleteRule(c *gin.Context) {
	if err := h.svc.DeleteRule(c.Param("id")); err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, gin.H{"deleted": c.Param("id")})
}

// ListAlerts GET /monitor/alerts?status=&limit=
func (h *MonitorHandler) ListAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	items, err := h.svc.ListEvents(c.Query("status"), limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items)})
}

// AckAlert POST /monitor/alerts/:id/ack
func (h *MonitorHandler) AckAlert(c *gin.Context) {
	by, _ := c.Get("username")
	user, _ := by.(string)
	if user == "" {
		user = "system"
	}
	if err := h.svc.AckEvent(c.Param("id"), user); err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, gin.H{"acknowledged": c.Param("id")})
}
