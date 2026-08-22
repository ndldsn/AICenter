package service

import (
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/notifier"
	"github.com/aicenter/aicenter/internal/repository"
)

// NotificationService exposes channel/template/log CRUD and event dispatch.
type NotificationService struct {
	repo       *repository.NotificationRepository
	dispatcher *notifier.Dispatcher
}

func NewNotificationService(repo *repository.NotificationRepository, dispatcher *notifier.Dispatcher) *NotificationService {
	return &NotificationService{repo: repo, dispatcher: dispatcher}
}

// ---- Channels ----

func (s *NotificationService) ListChannels(enabledOnly bool) ([]models.NotificationChannel, error) {
	return s.repo.ListChannels(enabledOnly)
}
func (s *NotificationService) GetChannel(id string) (*models.NotificationChannel, error) {
	return s.repo.GetChannel(id)
}
func (s *NotificationService) CreateChannel(c *models.NotificationChannel) error {
	return s.repo.CreateChannel(c)
}
func (s *NotificationService) UpdateChannel(id string, c *models.NotificationChannel) (*models.NotificationChannel, error) {
	return s.repo.UpdateChannel(id, c)
}
func (s *NotificationService) DeleteChannel(id string) error {
	return s.repo.DeleteChannel(id)
}

// ---- Templates ----

func (s *NotificationService) ListTemplates(eventType string) ([]models.NotificationTemplate, error) {
	return s.repo.ListTemplates(eventType)
}
func (s *NotificationService) GetTemplate(id string) (*models.NotificationTemplate, error) {
	return s.repo.GetTemplate(id)
}
func (s *NotificationService) CreateTemplate(t *models.NotificationTemplate) error {
	return s.repo.CreateTemplate(t)
}
func (s *NotificationService) UpdateTemplate(id string, t *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	return s.repo.UpdateTemplate(id, t)
}
func (s *NotificationService) DeleteTemplate(id string) error {
	return s.repo.DeleteTemplate(id)
}

// ---- Delivery logs ----

func (s *NotificationService) ListDeliveryLogs(status string, limit int) ([]models.NotificationDeliveryLog, error) {
	return s.repo.ListDeliveryLogs(status, limit)
}

// ---- Dispatch ----

// Notify dispatches an event, optionally restricted to the given channel ids
// (e.g. alert rule's bound notification_channels). When channelIDs is empty the
// dispatcher selects channels from the matching templates.
func (s *NotificationService) Notify(eventType, title, severity, message string, data map[string]string, channelIDs []string) error {
	ev := &models.NotificationEvent{
		Type:     eventType,
		Title:    title,
		Severity: severity,
		Message:  message,
		Data:     data,
	}
	var channels []models.NotificationChannel
	if len(channelIDs) > 0 {
		chs, err := s.repo.ChannelsByIDs(channelIDs)
		if err != nil {
			return err
		}
		channels = chs
	}
	s.dispatcher.DispatchEvent(ev, channels)
	return nil
}
