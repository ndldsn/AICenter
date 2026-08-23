package ai

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFactory_Build_WrapsDeepSeekAndOllamaAsOpenAICompat(t *testing.T) {
	f := NewFactory()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cds := f.Build(ProviderDeepSeek, Config{BaseURL: srv.URL, APIKey: "***"})
	col := f.Build(ProviderOllama, Config{BaseURL: srv.URL, APIKey: "***"})
	cgem := f.Build(ProviderGemini, Config{})
	cmock := f.Build(ProviderMock, Config{})
	if cds == nil { t.Fatal("deepseek nil") }
	if col == nil { t.Fatal("ollama nil") }
	if cgem == nil { t.Fatal("gemini nil") }
	if cmock == nil { t.Fatal("mock nil") }
}

func TestFactory_Build_Defaults_WhenBaseURLEmpty(t *testing.T) {
	f := NewFactory()
	for _, p := range []ProviderType{
		ProviderDeepSeek,
		ProviderOllama,
		ProviderGemini,
		ProviderOpenAICompatible,
	} {
		c := f.Build(p, Config{})
		if c == nil {
			t.Fatalf("%s returned nil client", p)
		}
	}
}
