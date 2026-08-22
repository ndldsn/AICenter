package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/repository"
	"go.uber.org/zap"
)

// Dispatcher resolves templates and delivers NotificationEvents to channels.
type Dispatcher struct {
	repo      *repository.NotificationRepository
	providers map[string]Provider
	log       *zap.Logger
}

func NewDispatcher(repo *repository.NotificationRepository, log *zap.Logger) *Dispatcher {
	d := &Dispatcher{
		repo:      repo,
		providers: map[string]Provider{},
		log:       log,
	}
	d.Register(NewWebhookProvider(), NewConsoleProvider(),
		&EmailProvider{}, &SMSProvider{}, &IMProvider{})
	return d
}

func (d *Dispatcher) Register(provs ...Provider) {
	for _, p := range provs {
		d.providers[p.Type()] = p
	}
}

// DispatchEvent finds templates for the event type, renders each, and sends it
// through the chosen channels. If no template is configured it falls back to a
// built-in default message delivered via console + any channels passed in.
func (d *Dispatcher) DispatchEvent(ev *models.NotificationEvent, channels []models.NotificationChannel) {
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now().UTC()
	}
	tmpls, err := d.repo.ListTemplates(ev.Type)
	if err != nil {
		d.log.Warn("notifier: list templates failed", zap.Error(err))
	}

	// Determine the target channels: explicit channels win; otherwise use the
	// channel types declared by each template.
	if len(channels) == 0 {
		channels = d.channelsForTemplates(tmpls)
	}

	if len(tmpls) == 0 {
		// No template configured → emit a sensible default through console.
		d.dispatchToChannels(ev.Type, channels, "", defaultMessage(ev), "")
		return
	}

	for _, t := range tmpls {
		if !t.IsEnabled {
			continue
		}
		subject, body, err := render(t, ev)
		if err != nil {
			d.log.Warn("notifier: render template failed", zap.String("tpl", t.ID), zap.Error(err))
			continue
		}
		d.dispatchToChannels(ev.Type, channels, subject, body, t.ID)
	}
}

func (d *Dispatcher) channelsForTemplates(tmpls []models.NotificationTemplate) []models.NotificationChannel {
	// collect declared channel types
	typeSet := map[string]bool{}
	for _, t := range tmpls {
		var types []string
		if t.Channels != "" && t.Channels != "[]" {
			_ = jsonUnmarshalTypes(t.Channels, &types)
		}
		for _, ty := range types {
			typeSet[ty] = true
		}
	}
	all, err := d.repo.ListChannels(true)
	if err != nil {
		return nil
	}
	out := []models.NotificationChannel{}
	for _, ch := range all {
		if typeSet[ch.Type] {
			out = append(out, ch)
		}
	}
	return out
}

func (d *Dispatcher) dispatchToChannels(eventType string, channels []models.NotificationChannel, subject, body, templateID string) {
	if len(channels) == 0 {
		// Always log to console so notifications are visible in dev.
		_ = d.providers["console"].Send("", subject, body)
		return
	}
	for _, ch := range channels {
		if !ch.IsEnabled {
			continue
		}
		prov, ok := d.providers[ch.Type]
		if !ok {
			d.log.Warn("notifier: no provider for channel type", zap.String("type", ch.Type))
			continue
		}
		err := prov.Send(ch.Config, subject, body)
		status := "sent"
		errMsg := ""
		if err != nil {
			status = "failed"
			errMsg = err.Error()
			d.log.Error("notifier: send failed", zap.String("channel", ch.ID), zap.String("type", ch.Type), zap.Error(err))
		}
		_ = d.repo.CreateDeliveryLog(&models.NotificationDeliveryLog{
			ChannelID:    ch.ID,
			ChannelType:  ch.Type,
			TemplateID:   templateID,
			EventType:    eventType,
			Subject:      subject,
			Body:         body,
			Status:       status,
			ErrorMessage: errMsg,
		})
	}
}

func render(t models.NotificationTemplate, ev *models.NotificationEvent) (subject, body string, err error) {
	funcs := template.FuncMap{
		"upper": func(s string) string { return fmt.Sprintf("%s", s) },
	}
	bt, err := template.New("body").Funcs(funcs).Parse(t.Body)
	if err != nil {
		return "", "", err
	}
	var bodyBuf bytes.Buffer
	if err := bt.Execute(&bodyBuf, ev); err != nil {
		return "", "", err
	}
	body = bodyBuf.String()

	if t.Subject != "" {
		st, err := template.New("subject").Parse(t.Subject)
		if err != nil {
			return "", "", err
		}
		var subjBuf bytes.Buffer
		if err := st.Execute(&subjBuf, ev); err != nil {
			return "", "", err
		}
		subject = subjBuf.String()
	}
	return subject, body, nil
}

func defaultMessage(ev *models.NotificationEvent) string {
	msg := ev.Title
	if ev.Message != "" {
		msg += ": " + ev.Message
	}
	if ev.Severity != "" {
		msg = "[" + ev.Severity + "] " + msg
	}
	return msg
}

func jsonUnmarshalTypes(raw string, out *[]string) error {
	return json.Unmarshal([]byte(raw), out)
}
