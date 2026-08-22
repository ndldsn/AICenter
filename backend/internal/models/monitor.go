package models

import "time"

// Metric is a single collected monitoring sample.
type Metric struct {
	ID          string            `json:"id"`
	ServerID    string            `json:"server_id,omitempty"`
	MetricName  string            `json:"metric_name"`
	Value       float64           `json:"value"`
	Unit        string            `json:"unit,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	CollectedAt time.Time         `json:"collected_at"`
}

// AlertRule defines a threshold condition on a metric.
type AlertRule struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	MetricName string   `json:"metric_name"`
	Condition string    `json:"condition"` // gt, lt, gte, lte
	Threshold float64   `json:"threshold"`
	Duration  int       `json:"duration"` // seconds; 0 = immediate
	Severity  string    `json:"severity"` // info, warning, critical
	ServerID  *string   `json:"server_id,omitempty"`
	IsEnabled bool      `json:"is_enabled"`
	// IsEnabledSet distinguishes "explicitly false" from "omitted" (JSON
	// cannot express both on a plain bool). When omitted the rule is enabled.
	IsEnabledSet bool `json:"-"`
	Cooldown  int       `json:"cooldown"` // seconds between repeat alerts
	// NotificationChannels is a JSON array of notification_channel ids bound to
	// the rule (Phase 7). Empty → dispatcher selects channels by template.
	NotificationChannels string `json:"notification_channels,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AlertEvent is a fired alert instance.
type AlertEvent struct {
	ID             string     `json:"id"`
	RuleID         string     `json:"rule_id,omitempty"`
	RuleName       string     `json:"rule_name,omitempty"`
	ServerID       string     `json:"server_id,omitempty"`
	MetricName     string     `json:"metric_name,omitempty"`
	Value          float64    `json:"value"`
	Threshold      float64    `json:"threshold"`
	Condition      string     `json:"condition,omitempty"`
	Severity       string     `json:"severity"`
	Message        string     `json:"message,omitempty"`
	Status         string     `json:"status"` // firing, acknowledged, resolved
	TriggeredAt    time.Time  `json:"triggered_at"`
	AcknowledgedBy string     `json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
}

// MetricPoint is an aggregated datapoint for charts.
type MetricPoint struct {
	Bucket string  `json:"bucket"` // ISO timestamp of bucket start
	Avg    float64 `json:"avg"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  int     `json:"count"`
}
