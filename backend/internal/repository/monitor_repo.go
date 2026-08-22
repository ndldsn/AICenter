package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/google/uuid"
)

var ErrAlertRuleNotFound = errors.New("alert rule not found")

type MonitorRepository struct{ db *sql.DB }

func NewMonitorRepository(db *sql.DB) *MonitorRepository {
	return &MonitorRepository{db: db}
}

// ---- Metrics ----

// InsertMetric stores one sample.
func (r *MonitorRepository) InsertMetric(m *models.Metric) error {
	m.ID = uuid.New().String()
	labels := "{}"
	if len(m.Labels) > 0 {
		labels = jsonMarshal(m.Labels)
	}
	if m.CollectedAt.IsZero() {
		m.CollectedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`
		INSERT INTO monitor_metrics (id, server_id, metric_name, value, unit, labels, collected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, nullIfEmpty(m.ServerID), m.MetricName, m.Value, m.Unit, labels,
		m.CollectedAt.UTC().Format(time.RFC3339))
	return err
}

// ListMetrics queries raw samples with optional filters.
func (r *MonitorRepository) ListMetrics(serverID, name string, since time.Time, limit int) ([]models.Metric, error) {
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}
	where := []string{"1=1"}
	var args []interface{}
	if serverID != "" {
		where = append(where, "server_id=?")
		args = append(args, serverID)
	}
	if name != "" {
		where = append(where, "metric_name=?")
		args = append(args, name)
	}
	if !since.IsZero() {
		where = append(where, "collected_at>=?")
		args = append(args, since.UTC().Format(time.RFC3339))
	}
	q := fmt.Sprintf(`SELECT id, COALESCE(server_id,''), metric_name, value, COALESCE(unit,''), COALESCE(labels,'{}'), collected_at
		FROM monitor_metrics WHERE %s ORDER BY collected_at DESC LIMIT ?`,
		strings.Join(where, " AND "))
	args = append(args, limit)

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMetrics(rows)
}

// AggregateMetrics buckets samples by interval ("1m", "5m", "1h").
func (r *MonitorRepository) AggregateMetrics(serverID, name string, since time.Time, interval string, limit int) ([]models.MetricPoint, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	bucketLen := 16 // YYYY-MM-DDTHH:mm
	switch interval {
	case "1h":
		bucketLen = 13
	case "5m", "1m":
		bucketLen = 16
	default:
		interval, bucketLen = "1m", 16
	}

	where := []string{"metric_name=?"}
	var args []interface{}
	args = append(args, name)
	if serverID != "" {
		where = append(where, "server_id=?")
		args = append(args, serverID)
	}
	if !since.IsZero() {
		where = append(where, "collected_at>=?")
		args = append(args, since.UTC().Format(time.RFC3339))
	}

	q := fmt.Sprintf(`
		SELECT substr(collected_at,1,%d) AS bucket,
		       AVG(value), MIN(value), MAX(value), COUNT(*)
		FROM monitor_metrics WHERE %s
		GROUP BY bucket ORDER BY bucket DESC LIMIT ?`,
		bucketLen, strings.Join(where, " AND "))
	args = append(args, limit)

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.MetricPoint{}
	for rows.Next() {
		var p models.MetricPoint
		if err := rows.Scan(&p.Bucket, &p.Avg, &p.Min, &p.Max, &p.Count); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// LatestMetrics returns the most recent sample per metric per server.
func (r *MonitorRepository) LatestMetrics(serverID string) ([]models.Metric, error) {
	where := ""
	var args []interface{}
	if serverID != "" {
		where = "WHERE mm.server_id=?"
		args = append(args, serverID)
	}
	q := fmt.Sprintf(`
		SELECT mm.id, COALESCE(mm.server_id,''), mm.metric_name, mm.value, COALESCE(mm.unit,''), COALESCE(mm.labels,'{}'), mm.collected_at
		FROM monitor_metrics mm
		JOIN (
			SELECT metric_name, COALESCE(server_id,'') sid, MAX(collected_at) mc
			FROM monitor_metrics GROUP BY metric_name, sid
		) latest ON latest.metric_name=mm.metric_name AND latest.sid=COALESCE(mm.server_id,'') AND latest.mc=mm.collected_at
		%s ORDER BY mm.metric_name`, where)
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMetrics(rows)
}

func scanMetrics(rows *sql.Rows) ([]models.Metric, error) {
	out := []models.Metric{}
	for rows.Next() {
		var m models.Metric
		var labels, ts string
		if err := rows.Scan(&m.ID, &m.ServerID, &m.MetricName, &m.Value, &m.Unit, &labels, &ts); err != nil {
			return nil, err
		}
		t, _ := time.Parse(time.RFC3339, ts)
		if t.IsZero() {
			t, _ = time.Parse("2006-01-02T15:04:05", ts)
		}
		m.CollectedAt = t
		if labels != "" && labels != "{}" {
			lbls := map[string]string{}
			if json.Unmarshal([]byte(labels), &lbls) == nil {
				m.Labels = lbls
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// DeleteMetricsBefore prunes old raw samples (retention).
func (r *MonitorRepository) DeleteMetricsBefore(t time.Time) (int64, error) {
	res, err := r.db.Exec("DELETE FROM monitor_metrics WHERE collected_at<?",
		t.UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ---- Alert rules ----

const ruleCols = `id, name, metric_name, condition, threshold, duration, severity, server_id, is_enabled, cooldown, notification_channels, created_at, updated_at`

func (r *MonitorRepository) CreateRule(rule *models.AlertRule) error {
	rule.ID = uuid.New().String()
	now := time.Now().UTC()
	rule.CreatedAt, rule.UpdatedAt = now, now
	if rule.NotificationChannels == "" {
		rule.NotificationChannels = "[]"
	}
	_, err := r.db.Exec(`
		INSERT INTO alert_rules (id, name, metric_name, condition, threshold, duration, severity, server_id, is_enabled, cooldown, notification_channels, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		rule.ID, rule.Name, rule.MetricName, rule.Condition, rule.Threshold,
		rule.Duration, rule.Severity, nullIfEmpty(deref(rule.ServerID)), boolInt(rule.IsEnabled), rule.Cooldown, rule.NotificationChannels)
	return err
}

func (r *MonitorRepository) UpdateRule(id string, upd *models.AlertRule) (*models.AlertRule, error) {
	if upd.NotificationChannels == "" {
		upd.NotificationChannels = "[]"
	}
	res, err := r.db.Exec(`
		UPDATE alert_rules SET name=?, metric_name=?, condition=?, threshold=?, duration=?,
		       severity=?, server_id=?, is_enabled=?, cooldown=?, notification_channels=?, updated_at=datetime('now')
		WHERE id=?`,
		upd.Name, upd.MetricName, upd.Condition, upd.Threshold, upd.Duration,
		upd.Severity, nullIfEmpty(deref(upd.ServerID)), boolInt(upd.IsEnabled), upd.Cooldown, upd.NotificationChannels, id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrAlertRuleNotFound
	}
	return r.GetRule(id)
}

func (r *MonitorRepository) GetRule(id string) (*models.AlertRule, error) {
	row := r.db.QueryRow(`SELECT `+ruleCols+` FROM alert_rules WHERE id=?`, id)
	return scanRule(row)
}

func (r *MonitorRepository) ListRules(enabledOnly bool) ([]models.AlertRule, error) {
	q := `SELECT ` + ruleCols + ` FROM alert_rules`
	if enabledOnly {
		q += " WHERE is_enabled=1"
	}
	q += " ORDER BY created_at DESC"
	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.AlertRule{}
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rule)
	}
	return out, rows.Err()
}

func (r *MonitorRepository) DeleteRule(id string) error {
	res, err := r.db.Exec(`DELETE FROM alert_rules WHERE id=?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrAlertRuleNotFound
	}
	return nil
}

type rowScanner interface{ Scan(...interface{}) error }

func scanRule(sc rowScanner) (*models.AlertRule, error) {
	var rule models.AlertRule
	var serverID sql.NullString
	var enabled int
	var notifCh sql.NullString
	var created, updated string
	if err := sc.Scan(&rule.ID, &rule.Name, &rule.MetricName, &rule.Condition, &rule.Threshold,
		&rule.Duration, &rule.Severity, &serverID, &enabled, &rule.Cooldown, &notifCh, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAlertRuleNotFound
		}
		return nil, err
	}
	rule.IsEnabled = enabled == 1
	if serverID.Valid && serverID.String != "" {
		s := serverID.String
		rule.ServerID = &s
	}
	if notifCh.Valid && notifCh.String != "" {
		rule.NotificationChannels = notifCh.String
	}
	rule.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt, _ = time.Parse(time.RFC3339, created)
	}
	rule.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
	if rule.UpdatedAt.IsZero() {
		rule.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	}
	return &rule, nil
}

// ---- Alert events ----

func (r *MonitorRepository) CreateEvent(e *models.AlertEvent) error {
	e.ID = uuid.New().String()
	if e.Status == "" {
		e.Status = "firing"
	}
	e.TriggeredAt = time.Now().UTC()
	_, err := r.db.Exec(`
		INSERT INTO alert_events (id, rule_id, rule_name, server_id, metric_name, value, threshold, condition, severity, message, status, triggered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		e.ID, nullIfEmpty(e.RuleID), e.RuleName, nullIfEmpty(e.ServerID),
		e.MetricName, e.Value, e.Threshold, e.Condition, e.Severity, e.Message, e.Status)
	return err
}

func (r *MonitorRepository) ListEvents(status string, limit int) ([]models.AlertEvent, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	q := `SELECT id, COALESCE(rule_id,''), COALESCE(rule_name,''), COALESCE(server_id,''),
	             COALESCE(metric_name,''), value, threshold, COALESCE(condition,''), severity,
	             COALESCE(message,''), status, triggered_at, COALESCE(acknowledged_by,''), acknowledged_at
	      FROM alert_events`
	var args []interface{}
	if status != "" {
		q += " WHERE status=?"
		args = append(args, status)
	}
	q += " ORDER BY triggered_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.AlertEvent{}
	for rows.Next() {
		var e models.AlertEvent
		var ackBy, ackAt, triggeredAt sql.NullString
		if err := rows.Scan(&e.ID, &e.RuleID, &e.RuleName, &e.ServerID, &e.MetricName,
			&e.Value, &e.Threshold, &e.Condition, &e.Severity, &e.Message, &e.Status,
			&triggeredAt, &ackBy, &ackAt); err != nil {
			return nil, err
		}
		e.TriggeredAt = parseDBTime(triggeredAt.String)
		if ackBy.Valid {
			e.AcknowledgedBy = ackBy.String
		}
		if ackAt.Valid && ackAt.String != "" {
			t, _ := time.Parse("2006-01-02 15:04:05", ackAt.String)
			if !t.IsZero() {
				e.AcknowledgedAt = &t
			}
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *MonitorRepository) AckEvent(id, by string) error {
	res, err := r.db.Exec(`UPDATE alert_events SET status='acknowledged', acknowledged_by=?, acknowledged_at=datetime('now')
		WHERE id=? AND status='firing'`, by, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("alert event not found or already resolved")
	}
	return nil
}

// LastFiringForRule returns the most recent trigger time of a rule (for cooldown checks).
func (r *MonitorRepository) LastFiringForRule(ruleID, serverID string) (time.Time, bool, error) {
	q := `SELECT triggered_at FROM alert_events WHERE rule_id=?`
	var args []interface{}
	args = append(args, ruleID)
	if serverID != "" {
		q += " AND server_id=?"
		args = append(args, serverID)
	}
	q += " ORDER BY triggered_at DESC LIMIT 1"
	var ts string
	err := r.db.QueryRow(q, args...).Scan(&ts)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	t, _ := time.Parse("2006-01-02 15:04:05", ts)
	if t.IsZero() {
		t, _ = time.Parse(time.RFC3339, ts)
	}
	return t, true, nil
}

// ---- helpers ----

// parseDBTime handles both SQLite "2006-01-02 15:04:05" and RFC3339 strings.
func parseDBTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	if t.IsZero() {
		t, _ = time.Parse(time.RFC3339, s)
	}
	if t.IsZero() {
		t, _ = time.Parse("2006-01-02T15:04:05", s)
	}
	return t
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
