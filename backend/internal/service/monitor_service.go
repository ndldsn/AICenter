package service

import (
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/monitor"
	"github.com/aicenter/aicenter/internal/repository"
)

// MonitorService wires the monitor repository and alert engine to handlers.
type MonitorService struct {
	repo   *repository.MonitorRepository
	engine *monitor.Engine
}

func NewMonitorService(repo *repository.MonitorRepository, engine *monitor.Engine) *MonitorService {
	return &MonitorService{repo: repo, engine: engine}
}

// QueryMetrics returns raw samples.
func (s *MonitorService) QueryMetrics(serverID, name string, since *time.Time, limit int) ([]models.Metric, error) {
	var st time.Time
	if since != nil {
		st = *since
	}
	return s.repo.ListMetrics(serverID, name, st, limit)
}

// QueryLatest returns the newest sample per metric/server.
func (s *MonitorService) QueryLatest(serverID string) ([]models.Metric, error) {
	return s.repo.LatestMetrics(serverID)
}

// QueryAggregate returns bucketed avg/min/max for charts.
func (s *MonitorService) QueryAggregate(serverID, name, interval string, since *time.Time, limit int) ([]models.MetricPoint, error) {
	var st time.Time
	if since != nil {
		st = *since
	}
	return s.repo.AggregateMetrics(serverID, name, st, interval, limit)
}

// Ingest accepts external metric pushes (server agents).
func (s *MonitorService) Ingest(metrics []models.Metric) error {
	if s.engine != nil {
		return s.engine.Ingest(metrics)
	}
	for i := range metrics {
		if err := s.repo.InsertMetric(&metrics[i]); err != nil {
			return err
		}
	}
	return nil
}

// ---- Alert rules ----

func (s *MonitorService) ListRules(enabledOnly bool) ([]models.AlertRule, error) {
	return s.repo.ListRules(enabledOnly)
}

func (s *MonitorService) GetRule(id string) (*models.AlertRule, error) {
	return s.repo.GetRule(id)
}

func (s *MonitorService) CreateRule(rule *models.AlertRule) error {
	switch rule.Condition {
	case "gt", "lt", "gte", "lte":
	default:
		rule.Condition = "gt"
	}
	switch rule.Severity {
	case "info", "warning", "critical":
	default:
		rule.Severity = "warning"
	}
	if !rule.IsEnabledSet {
		rule.IsEnabled = true // omitted → enabled by default
	}
	if rule.Cooldown < 0 {
		rule.Cooldown = 0
	}
	return s.repo.CreateRule(rule)
}

func (s *MonitorService) UpdateRule(id string, rule *models.AlertRule) (*models.AlertRule, error) {
	return s.repo.UpdateRule(id, rule)
}

func (s *MonitorService) DeleteRule(id string) error {
	return s.repo.DeleteRule(id)
}

// ---- Alert events ----

func (s *MonitorService) ListEvents(status string, limit int) ([]models.AlertEvent, error) {
	return s.repo.ListEvents(status, limit)
}

func (s *MonitorService) AckEvent(id, by string) error {
	return s.repo.AckEvent(id, by)
}
