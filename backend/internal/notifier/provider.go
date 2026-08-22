// Package notifier delivers NotificationEvents over configured channels.
package notifier

import (
	"encoding/json"

	"github.com/aicenter/aicenter/internal/models"
)

// Provider delivers a rendered message to one channel type.
type Provider interface {
	// Type identifies the provider (webhook, email, sms, im, console).
	Type() string
	// Send dispatches the message. channelConfig is the raw JSON config of the
	// NotificationChannel; subject may be empty for channel types that ignore it.
	Send(channelConfig, subject, body string) error
}

// channelConfig is the parsed form shared by providers.
type channelConfig struct {
	URL      string `json:"url"`      // webhook
	Token    string `json:"token"`    // im / webhook auth
	To       string `json:"to"`       // email / sms recipient(s), comma separated
	From     string `json:"from"`     // email from
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	APIKey   string `json:"api_key"` // sms / im gateway
}

func parseConfig(raw string) channelConfig {
	var c channelConfig
	if raw == "" {
		return c
	}
	_ = json.Unmarshal([]byte(raw), &c)
	return c
}

var _ = models.NotificationEvent{}
