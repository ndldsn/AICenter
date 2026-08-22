package notifier

import (
	"github.com/aicenter/aicenter/internal/pkg/logger"
	"go.uber.org/zap"
)

// ConsoleProvider logs the rendered message. It is the default safe channel
// used in development and when no external provider is configured, so the
// notification pipeline is always demonstrable without credentials.
type ConsoleProvider struct{}

func NewConsoleProvider() *ConsoleProvider { return &ConsoleProvider{} }

func (p *ConsoleProvider) Type() string { return "console" }

func (p *ConsoleProvider) Send(_, subject, body string) error {
	if subject != "" {
		logger.Get().Info("[notification] "+subject, zap.String("body", body))
	} else {
		logger.Get().Info("[notification] " + body)
	}
	return nil
}
