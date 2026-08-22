package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/google/uuid"
)

var ErrChannelNotFound = errors.New("notification channel not found")
var ErrTemplateNotFound = errors.New("notification template not found")

type NotificationRepository struct{ db *sql.DB }

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ---- Channels ----

func (r *NotificationRepository) ListChannels(enabledOnly bool) ([]models.NotificationChannel, error) {
	q := `SELECT id, name, type, config, is_enabled, created_at, updated_at FROM notification_channels`
	if enabledOnly {
		q += " WHERE is_enabled=1"
	}
	q += " ORDER BY created_at DESC"
	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.NotificationChannel{}
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *ch)
	}
	return out, rows.Err()
}

func (r *NotificationRepository) GetChannel(id string) (*models.NotificationChannel, error) {
	row := r.db.QueryRow(`SELECT id, name, type, config, is_enabled, created_at, updated_at FROM notification_channels WHERE id=?`, id)
	return scanChannel(row)
}

func (r *NotificationRepository) CreateChannel(ch *models.NotificationChannel) error {
	ch.ID = uuid.New().String()
	if ch.Config == "" {
		ch.Config = "{}"
	}
	now := time.Now().UTC()
	ch.CreatedAt, ch.UpdatedAt = now, now
	_, err := r.db.Exec(`INSERT INTO notification_channels (id, name, type, config, is_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		ch.ID, ch.Name, ch.Type, ch.Config, boolInt(ch.IsEnabled))
	return err
}

func (r *NotificationRepository) UpdateChannel(id string, upd *models.NotificationChannel) (*models.NotificationChannel, error) {
	res, err := r.db.Exec(`UPDATE notification_channels SET name=?, type=?, config=?, is_enabled=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		upd.Name, upd.Type, upd.Config, boolInt(upd.IsEnabled), id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrChannelNotFound
	}
	return r.GetChannel(id)
}

func (r *NotificationRepository) DeleteChannel(id string) error {
	res, err := r.db.Exec(`DELETE FROM notification_channels WHERE id=?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrChannelNotFound
	}
	return nil
}

func scanChannel(sc rowScanner) (*models.NotificationChannel, error) {
	var ch models.NotificationChannel
	var cfg string
	var enabled int
	var created, updated string
	if err := sc.Scan(&ch.ID, &ch.Name, &ch.Type, &cfg, &enabled, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	ch.Config = cfg
	ch.IsEnabled = enabled == 1
	ch.CreatedAt = parseDBTime(created)
	ch.UpdatedAt = parseDBTime(updated)
	return &ch, nil
}

// ---- Templates ----

func (r *NotificationRepository) ListTemplates(eventType string) ([]models.NotificationTemplate, error) {
	q := `SELECT id, name, event_type, subject, body, channels, is_enabled, created_at, updated_at FROM notification_templates`
	var args []interface{}
	if eventType != "" {
		q += " WHERE event_type=?"
		args = append(args, eventType)
	}
	q += " ORDER BY created_at DESC"
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.NotificationTemplate{}
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *NotificationRepository) GetTemplate(id string) (*models.NotificationTemplate, error) {
	row := r.db.QueryRow(`SELECT id, name, event_type, subject, body, channels, is_enabled, created_at, updated_at FROM notification_templates WHERE id=?`, id)
	return scanTemplate(row)
}

func (r *NotificationRepository) CreateTemplate(t *models.NotificationTemplate) error {
	t.ID = uuid.New().String()
	if t.Channels == "" {
		t.Channels = "[]"
	}
	now := time.Now().UTC()
	t.CreatedAt, t.UpdatedAt = now, now
	_, err := r.db.Exec(`INSERT INTO notification_templates (id, name, event_type, subject, body, channels, is_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		t.ID, t.Name, t.EventType, t.Subject, t.Body, t.Channels, boolInt(t.IsEnabled))
	return err
}

func (r *NotificationRepository) UpdateTemplate(id string, upd *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	res, err := r.db.Exec(`UPDATE notification_templates SET name=?, event_type=?, subject=?, body=?, channels=?, is_enabled=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		upd.Name, upd.EventType, upd.Subject, upd.Body, upd.Channels, boolInt(upd.IsEnabled), id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrTemplateNotFound
	}
	return r.GetTemplate(id)
}

func (r *NotificationRepository) DeleteTemplate(id string) error {
	res, err := r.db.Exec(`DELETE FROM notification_templates WHERE id=?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrTemplateNotFound
	}
	return nil
}

func scanTemplate(sc rowScanner) (*models.NotificationTemplate, error) {
	var t models.NotificationTemplate
	var channels string
	var enabled int
	var created, updated string
	if err := sc.Scan(&t.ID, &t.Name, &t.EventType, &t.Subject, &t.Body, &channels, &enabled, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	if channels != "" {
		t.Channels = channels
	}
	t.IsEnabled = enabled == 1
	t.CreatedAt = parseDBTime(created)
	t.UpdatedAt = parseDBTime(updated)
	return &t, nil
}

// ---- Delivery logs ----

func (r *NotificationRepository) CreateDeliveryLog(l *models.NotificationDeliveryLog) error {
	l.ID = uuid.New().String()
	if l.Status == "" {
		l.Status = "pending"
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`INSERT INTO notification_delivery_logs (id, channel_id, channel_type, template_id, event_type, subject, body, status, error_message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		l.ID, nullIfEmpty(l.ChannelID), l.ChannelType, nullIfEmpty(l.TemplateID),
		l.EventType, l.Subject, l.Body, l.Status, nullIfEmpty(l.ErrorMessage))
	return err
}

func (r *NotificationRepository) ListDeliveryLogs(status string, limit int) ([]models.NotificationDeliveryLog, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	q := `SELECT id, COALESCE(channel_id,''), COALESCE(channel_type,''), COALESCE(template_id,''),
	             COALESCE(event_type,''), COALESCE(subject,''), COALESCE(body,''), status,
	             COALESCE(error_message,''), created_at
	      FROM notification_delivery_logs`
	var args []interface{}
	if status != "" {
		q += " WHERE status=?"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.NotificationDeliveryLog{}
	for rows.Next() {
		var l models.NotificationDeliveryLog
		var created string
		if err := rows.Scan(&l.ID, &l.ChannelID, &l.ChannelType, &l.TemplateID, &l.EventType,
			&l.Subject, &l.Body, &l.Status, &l.ErrorMessage, &created); err != nil {
			return nil, err
		}
		l.CreatedAt = parseDBTime(created)
		out = append(out, l)
	}
	return out, rows.Err()
}

// ChannelsByIDs returns channels whose ids are in the provided list.
func (r *NotificationRepository) ChannelsByIDs(ids []string) ([]models.NotificationChannel, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q := `SELECT id, name, type, config, is_enabled, created_at, updated_at FROM notification_channels WHERE id IN (`
	args := make([]interface{}, 0, len(ids))
	for i, id := range ids {
		if i > 0 {
			q += ","
		}
		q += "?"
		args = append(args, id)
	}
	q += ") AND is_enabled=1"
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.NotificationChannel{}
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *ch)
	}
	return out, rows.Err()
}

// ParseChannelIDs decodes the rule's notification_channels JSON column.
func ParseChannelIDs(raw string) []string {
	if raw == "" {
		return nil
	}
	var ids []string
	if json.Unmarshal([]byte(raw), &ids) == nil {
		return ids
	}
	return nil
}
