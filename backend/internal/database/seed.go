package database

import (
	"database/sql"
	"time"

	"github.com/aicenter/aicenter/internal/pkg/crypto"
)

// SeedData adds initial data for development
func SeedData(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM servers").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(`
		INSERT INTO servers (id, name, host, port, username, auth_type, status, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "srv-demo-001", "Demo Web Server", "192.168.1.10", 22, "ubuntu", "key", "offline", `["web","production"]`, now, now)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO servers (id, name, host, port, username, auth_type, status, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "srv-demo-002", "Demo DB Server", "192.168.1.11", 22, "root", "password", "offline", `["database","mysql"]`, now, now)
	if err != nil {
		return err
	}

	// Encrypt a dev-only test key so the provider rows have a well-formed
	// encrypted blob; the key is a sentinel string, not a real credential.
	testKey := "dev-test-key-sentinel-not-for-production"
	encKey, err := crypto.Encrypt(testKey)
	if err != nil {
		return err
	}

	providers := []struct {
		id, name, display, url, apiType string
		enabled, isDefault              int
	}{
		{"prov-openai", "OpenAI", "OpenAI", "https://api.openai.com/v1", "openai-compatible", 1, 1},
		{"prov-deepseek", "DeepSeek", "DeepSeek", "https://api.deepseek.com/v1", "openai-compatible", 1, 0},
		{"prov-ollama", "Ollama", "Ollama (local)", "http://localhost:11434/v1", "openai-compatible", 0, 0},
		{"prov-llmstudio", "LLM Studio", "LLM Studio (local)", "http://localhost:1234/v1", "openai-compatible", 0, 0},
		{"prov-mock", "Mock (dev)", "Mock Provider", "http://mock.local/v1", "mock", 1, 0},
	}
	for _, p := range providers {
		_, err = db.Exec(`
			INSERT INTO ai_providers (id, name, display_name, base_url, api_key_enc, api_type, is_enabled, is_default, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, p.id, p.name, p.display, p.url, encKey, p.apiType, p.enabled, p.isDefault, now, now)
		if err != nil {
			return err
		}
	}

	models := []struct {
		prov, name, model string
		def               int
	}{
		{"prov-openai", "GPT-4o", "gpt-4o", 1},
		{"prov-openai", "GPT-4.1", "gpt-4.1", 0},
		{"prov-openai", "GPT-4.1 mini", "gpt-4.1-mini", 0},
		{"prov-deepseek", "DeepSeek-V3", "deepseek-v3", 1},
		{"prov-ollama", "Llama 3.1 8B", "llama3.1", 1},
		{"prov-ollama", "DeepSeek-R1", "deepseek-r1:latest", 0},
		{"prov-llmstudio", "Llama 3.1 8B", "llama-3.1-8b", 1},
		{"prov-mock", "Mock-4o", "gpt-4o", 1},
	}
	for _, m := range models {
		_, err = db.Exec(`
			INSERT INTO ai_models (id, provider_id, name, model_id, model_type, max_tokens, supports_stream, supports_tools, is_enabled, is_default, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, "mod-"+m.model+"-"+m.prov, m.prov, m.name, m.model, "chat", 8192, 1, 0, 1, m.def, now, now)
		if err != nil {
			return err
		}
	}

	return nil
}
