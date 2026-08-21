package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/google/uuid"
)

var ErrAIProviderNotFound = errors.New("ai provider not found")
var ErrAIModelNotFound = errors.New("ai model not found")

// AIProviderRepository persists ai_providers rows.
type AIProviderRepository struct {
	db *sql.DB
}

func NewAIProviderRepository(db *sql.DB) *AIProviderRepository {
	return &AIProviderRepository{db: db}
}

func (r *AIProviderRepository) Create(p *models.AIProvider) error {
	p.ID = uuid.New().String()
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := r.db.Exec(`
		INSERT INTO ai_providers (id, name, display_name, base_url, api_key_enc, api_type, is_enabled, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.DisplayName, p.BaseURL, p.APIKeyEnc,
		string(p.APIType), p.IsEnabled, p.IsDefault, now, now)
	return err
}

func (r *AIProviderRepository) Update(p *models.AIProvider) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := r.db.Exec(`
		UPDATE ai_providers SET display_name=?, base_url=?, api_key_enc=?, api_type=?, is_enabled=?, updated_at=?
		WHERE id=?`,
		p.DisplayName, p.BaseURL, p.APIKeyEnc, string(p.APIType), p.IsEnabled, now, p.ID)
	return err
}

func (r *AIProviderRepository) Get(id string) (*models.AIProvider, error) {
	p, err := r.getByKey(id)
	if err != nil {
		return nil, err
	}
	p.APIKeyEnc = "" // never return the encrypted key from the store; caller re-fetches via a separate path if needed
	return p, nil
}

func (r *AIProviderRepository) GetRawKey(id string) (string, error) {
	var key string
	err := r.db.QueryRow("SELECT api_key_enc FROM ai_providers WHERE id=? AND is_enabled=1", id).Scan(&key)
	if err == sql.ErrNoRows {
		return "", ErrAIProviderNotFound
	}
	if err != nil {
		return "", err
	}
	return key, nil
}

func (r *AIProviderRepository) Delete(id string) error {
	r.db.Exec("DELETE FROM ai_models WHERE provider_id=?", id)
	result, err := r.db.Exec("DELETE FROM ai_providers WHERE id=?", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrAIProviderNotFound
	}
	return nil
}

func (r *AIProviderRepository) List() ([]models.AIProvider, error) {
	rows, err := r.db.Query(`SELECT id, name, display_name, base_url, api_key_enc, api_type, is_enabled, is_default, created_at, updated_at FROM ai_providers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAll(rows)
}

func (r *AIProviderRepository) getByKey(id string) (*models.AIProvider, error) {
	p, err := r.scanOne("SELECT id, name, display_name, base_url, api_key_enc, api_type, is_enabled, is_default, created_at, updated_at FROM ai_providers WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *AIProviderRepository) scanAll(rows *sql.Rows) ([]models.AIProvider, error) {
	var out []models.AIProvider
	for rows.Next() {
		p, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *AIProviderRepository) scanOne(query string, id string) (*models.AIProvider, error) {
	p := new(models.AIProvider)
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.DisplayName, &p.BaseURL, &p.APIKeyEnc, &p.APIType, &p.IsEnabled, &p.IsDefault, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrAIProviderNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *AIProviderRepository) scanRow(rows *sql.Rows) (*models.AIProvider, error) {
	p := new(models.AIProvider)
	err := rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.BaseURL, &p.APIKeyEnc, &p.APIType, &p.IsEnabled, &p.IsDefault, &p.CreatedAt, &p.UpdatedAt)
	if err == nil {
		p.APIKeyEnc = "" // never leak the encrypted blob in list results
	}
	return p, err
}

// AIModelRepository persists ai_models rows.
type AIModelRepository struct {
	db *sql.DB
}

func NewAIModelRepository(db *sql.DB) *AIModelRepository {
	return &AIModelRepository{db: db}
}

func (r *AIModelRepository) Create(m *models.AIModel) error {
	m.ID = uuid.New().String()
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := r.db.Exec(`
		INSERT INTO ai_models (id, provider_id, name, model_id, model_type, max_tokens, supports_stream, supports_tools, is_enabled, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.ProviderID, m.Name, m.ModelID, m.ModelType, m.MaxTokens,
		m.SupportsStream, m.SupportsTools, m.IsEnabled, m.IsDefault, now, now)
	return err
}

func (r *AIModelRepository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM ai_models WHERE id=?", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrAIModelNotFound
	}
	return nil
}

func (r *AIModelRepository) SetDefault(providerID, modelID string) error {
	_, err := r.db.Exec(`UPDATE ai_models SET is_default=0 WHERE provider_id=?`, providerID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`UPDATE ai_models SET is_default=1 WHERE provider_id=? AND model_id=? AND is_enabled=1`, providerID, modelID)
	return err
}

func (r *AIModelRepository) ToggleEnabled(modelID string, enabled bool) error {
	_, err := r.db.Exec("UPDATE ai_models SET is_enabled=? WHERE model_id=? AND provider_id IS NOT NULL", enabled, modelID)
	return err
}

func (r *AIModelRepository) ListByProvider(providerID string) ([]models.AIModel, error) {
	rows, err := r.db.Query(`SELECT id, provider_id, name, model_id, model_type, max_tokens, supports_stream, supports_tools, is_enabled, is_default, created_at, updated_at
		FROM ai_models WHERE provider_id=? AND is_enabled=1 ORDER BY name`, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AIModel
	for rows.Next() {
		m := new(models.AIModel)
		err := rows.Scan(&m.ID, &m.ProviderID, &m.Name, &m.ModelID, &m.ModelType, &m.MaxTokens, &m.SupportsStream, &m.SupportsTools, &m.IsEnabled, &m.IsDefault, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

// ConfigJSON unmarshals a provider's config JSON into v (used for extra dial
// options like model caps, custom headers, etc.).
func ConfigJSON(raw string, v interface{}) error {
	if raw == "" {
		return nil
	}
	return json.Unmarshal([]byte(raw), v)
}
