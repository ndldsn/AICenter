package models

import "time"

// NotificationChannel is a configured delivery destination.
type NotificationChannel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // webhook, email, sms, im, console
	Config    string    `json:"config,omitempty"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NotificationTemplate renders a message for a given event type.
type NotificationTemplate struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	EventType  string    `json:"event_type"` // alert.fired, approval.requested, approval.resolved
	Subject    string    `json:"subject,omitempty"`
	Body       string    `json:"body"`
	Channels   string    `json:"channels,omitempty"` // JSON array of channel types
	IsEnabled  bool      `json:"is_enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NotificationDeliveryLog records a single dispatch attempt.
type NotificationDeliveryLog struct {
	ID           string    `json:"id"`
	ChannelID    string    `json:"channel_id,omitempty"`
	ChannelType  string    `json:"channel_type,omitempty"`
	TemplateID   string    `json:"template_id,omitempty"`
	EventType    string    `json:"event_type,omitempty"`
	Subject      string    `json:"subject,omitempty"`
	Body         string    `json:"body,omitempty"`
	Status       string    `json:"status"` // pending, sent, failed
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// NotificationEvent is the payload passed to templates/providers.
type NotificationEvent struct {
	Type      string            `json:"type"`
	Title     string            `json:"title"`
	Severity  string            `json:"severity,omitempty"`
	Message   string            `json:"message,omitempty"`
	Data      map[string]string `json:"data,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}
