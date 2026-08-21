package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aicenter/aicenter/internal/ai"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/pkg/crypto"
)

var ErrDecryptionFailed = errors.New("decryption failed")

// AIProviderStore abstracts the provider DB access so tests can inject a mock.
type AIProviderStore interface {
	Create(p *models.AIProvider) error
	Update(p *models.AIProvider) error
	Get(id string) (*models.AIProvider, error)
	GetRawKey(id string) (string, error)
	Delete(id string) error
	List() ([]models.AIProvider, error)
}

// AIModelStore abstracts the model DB access.
type AIModelStore interface {
	Create(m *models.AIModel) error
	Delete(id string) error
	SetDefault(providerID, modelID string) error
	ToggleEnabled(modelID string, enabled bool) error
	ListByProvider(providerID string) ([]models.AIModel, error)
}

type AIService struct {
	providerStore AIProviderStore
	modelStore    AIModelStore

	mu       sync.RWMutex
	caches   map[string]ai.Client
	lastKeys map[string]string // cache invalidation key per provider
}

func NewAIService(ps AIProviderStore, ms AIModelStore) *AIService {
	return &AIService{providerStore: ps, modelStore: ms, caches: make(map[string]ai.Client), lastKeys: make(map[string]string)}
}

func (s *AIService) ListProviders() ([]models.AIProvider, error) {
	list, err := s.providerStore.List()
	if err != nil {
		return nil, err
	}
	// Enrich each row with an api_key_hint derived from the (decrypted) key,
	// then drop the encrypted blob from the response so callers never see it.
	for i := range list {
		raw, err := s.providerStore.GetRawKey(list[i].ID)
		if err != nil {
			list[i].APIKeyHint = ""
			continue
		}
		plaintext, err := crypto.Decrypt(raw)
		if err != nil {
			list[i].APIKeyHint = ""
			continue
		}
		list[i].APIKeyHint = crypto.Hint(plaintext)
		list[i].APIKeyEnc = ""
	}
	return list, nil
}

func (s *AIService) GetProvider(id string) (*models.AIProvider, error) {
	p, err := s.providerStore.Get(id)
	if err != nil {
		return nil, err
	}
	// Enrich with the key hint (never expose the encrypted blob).
	raw, err := s.providerStore.GetRawKey(id)
	if err == nil {
		if plaintext, derr := crypto.Decrypt(raw); derr == nil {
			p.APIKeyHint = crypto.Hint(plaintext)
		}
	}
	p.APIKeyEnc = ""
	return p, nil
}

func (s *AIService) CreateProvider(req models.AIProvider) (*models.AIProvider, error) {
	// Encrypt API key before persisting. If none was supplied, still write an
	// empty encrypted blob so the schema remains consistent.
	if req.APIKeyEnc != "" {
		enc, err := crypto.Encrypt(req.APIKeyEnc)
		if err != nil {
			return nil, err
		}
		req.APIKeyEnc = enc
		req.APIKeyHint = crypto.Hint(req.APIKeyEnc)
	}
	if err := s.providerStore.Create(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *AIService) UpdateProvider(id string, req models.AIProvider) (*models.AIProvider, error) {
	existing, err := s.providerStore.Get(id)
	if err != nil {
		return nil, err
	}
	// If a new API key was submitted, encrypt it and refresh the hint.
	if req.APIKeyEnc != "" {
		enc, err := crypto.Encrypt(req.APIKeyEnc)
		if err != nil {
			return nil, err
		}
		existing.APIKeyEnc = enc
		existing.APIKeyHint = crypto.Hint(req.APIKeyEnc)
	} else {
		// Keep the existing encrypted blob for the update SQL.
		raw, err := s.providerStore.GetRawKey(id)
		if err != nil {
			return nil, ErrDecryptionFailed
		}
		existing.APIKeyEnc = raw
	}
	existing.DisplayName = req.DisplayName
	existing.BaseURL = req.BaseURL
	existing.APIType = req.APIType
	existing.IsEnabled = req.IsEnabled
	if err := s.providerStore.Update(existing); err != nil {
		return nil, err
	}
	// Rebuild the provider dialer so the next chat call uses the new key/config.
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastKeys[id] = existing.APIKeyEnc
	s.caches[id] = s.buildClient(existing)
	return existing, nil
}

func (s *AIService) DeleteProvider(id string) error {
	if err := s.providerStore.Delete(id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.caches, id)
	delete(s.lastKeys, id)
	return nil
}

// DecryptProviderKey returns the plaintext API key for the given provider.
// The caller (the chat handler) uses it once and must not persist it.
func (s *AIService) DecryptProviderKey(id string) (string, error) {
	raw, err := s.providerStore.GetRawKey(id)
	if err != nil {
		return "", err
	}
	plaintext, err := crypto.Decrypt(raw)
	if err != nil {
		return "", ErrDecryptionFailed
	}
	return plaintext, nil
}

func (s *AIService) GetProviderClient(id string) (ai.Client, error) {
	s.mu.RLock()
	client, ok := s.caches[id]
	lastKey := s.lastKeys[id]
	s.mu.RUnlock()
	if ok && lastKey != "" {
		return client, nil
	}
	provider, err := s.providerStore.Get(id)
	if err != nil {
		return nil, err
	}
	return s.loadClient(provider)
}

func (s *AIService) loadClient(provider *models.AIProvider) (ai.Client, error) {
	key, err := s.providerStore.GetRawKey(provider.ID)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	plaintext, err := crypto.Decrypt(key)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	cfg := ai.Config{
		BaseURL: strings.TrimRight(provider.BaseURL, "/"),
		APIKey:  plaintext,
	}
	client := ai.NewFactory().Build(ai.ProviderType(provider.APIType), cfg)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.caches[provider.ID] = client
	s.lastKeys[provider.ID] = key
	return client, nil
}

func (s *AIService) buildClient(p *models.AIProvider) ai.Client {
	return ai.NewFactory().Build(ai.ProviderType(p.APIType), ai.Config{BaseURL: strings.TrimRight(p.BaseURL, "/")})
}

// ListModels fetches provider models (provider-side /models call) or returns
// the configured defaults from DB when streaming is disabled.
func toModelsModel(m ai.Model) models.AIModel {
	return models.AIModel{
		ID:             m.ID,
		ProviderID:     m.ProviderID,
		Name:           m.Name,
		ModelID:        m.ModelID,
		ModelType:      m.ModelType,
		MaxTokens:      m.MaxTokens,
		SupportsStream: m.SupportsStream,
		SupportsTools:  m.SupportsTools,
		IsEnabled:      m.IsEnabled,
		IsDefault:      m.IsDefault,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func (s *AIService) ListModels(providerID string) ([]models.AIModel, error) {
	dbList, err := s.modelStore.ListByProvider(providerID)
	if err != nil {
		return nil, err
	}
	if len(dbList) > 0 {
		return dbList, nil
	}
	client, err := s.GetProviderClient(providerID)
	if err != nil {
		return nil, err
	}
	remoteModels, err := client.ListModels(context.Background())
	if err != nil {
		return []models.AIModel{{
			ModelID: fmt.Sprintf("[provider unavailable: %v]", err),
			Name:    "Unavailable",
		}}, nil
	}
	out := make([]models.AIModel, 0, len(remoteModels))
	for _, m := range remoteModels {
		out = append(out, toModelsModel(m))
	}
	return out, nil
}

// Chat calls the provider with streaming; it writes SSE events into the
// streamer and returns only on fatal error.
func (s *AIService) Chat(providerID, modelID string, messages []ai.Message, streamer ai.SSEStreamer) error {
	client, err := s.GetProviderClient(providerID)
	if err != nil {
		return err
	}
	return client.ChatCompletion(context.Background(), ai.ChatRequest{
		Model:     modelID,
		Messages:  messages,
		Stream:    true,
		MaxTokens: 2048,
		Streamer:  streamer,
	})
}
