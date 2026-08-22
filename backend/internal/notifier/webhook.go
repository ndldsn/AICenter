package notifier

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aicenter/aicenter/internal/pkg/logger"
	"go.uber.org/zap"
)

// WebhookProvider posts JSON payloads to an external URL (Slack, DingTalk,
// generic callback, etc.).
type WebhookProvider struct {
	Client *http.Client
}

func NewWebhookProvider() *WebhookProvider {
	return &WebhookProvider{Client: &http.Client{Timeout: 10 * time.Second}}
}

func (p *WebhookProvider) Type() string { return "webhook" }

func (p *WebhookProvider) Send(channelConfig, subject, body string) error {
	cfg := parseConfig(channelConfig)
	if cfg.URL == "" {
		return nil // no-op channel (misconfigured) — treated as success
	}
	payload := map[string]string{
		"subject": subject,
		"text":    body,
	}
	if subject == "" {
		payload["text"] = body
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, cfg.URL, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		logger.Get().Warn("notifier: webhook returned non-2xx", zap.Int("status", resp.StatusCode))
	}
	return nil
}
