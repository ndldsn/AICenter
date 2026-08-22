package notifier

import (
	"github.com/aicenter/aicenter/internal/pkg/logger"
	"go.uber.org/zap"
)

// EmailProvider / SMSProvider / IMProvider are placeholder implementations that
// record the dispatch intent. Real delivery (SMTP, SMS gateway, IM bot) can be
// wired in later through the channel Config JSON without changing the API. MVP
// keeps them as no-op-with-log so the notification pipeline works end-to-end.
type EmailProvider struct{}

func (p *EmailProvider) Type() string { return "email" }

func (p *EmailProvider) Send(channelConfig, subject, body string) error {
	cfg := parseConfig(channelConfig)
	logger.Get().Info("[notification] email queued",
		zap.String("to", cfg.To),
		zap.String("from", cfg.From),
		zap.String("subject", subject),
		zap.String("body", body))
	return nil
}

type SMSProvider struct{}

func (p *SMSProvider) Type() string { return "sms" }

func (p *SMSProvider) Send(channelConfig, _, body string) error {
	cfg := parseConfig(channelConfig)
	logger.Get().Info("[notification] sms queued",
		zap.String("to", cfg.To),
		zap.String("body", body))
	return nil
}

type IMProvider struct{}

func (p *IMProvider) Type() string { return "im" }

func (p *IMProvider) Send(channelConfig, _, body string) error {
	cfg := parseConfig(channelConfig)
	logger.Get().Info("[notification] im queued",
		zap.String("token", maskToken(cfg.Token)),
		zap.String("body", body))
	return nil
}

func maskToken(t string) string {
	if len(t) <= 4 {
		return "***"
	}
	return t[:2] + "***" + t[len(t)-2:]
}
