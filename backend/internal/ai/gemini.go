package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type geminiClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewGeminiClient(cfg Config) Client {
	base := strings.TrimRight(cfg.BaseURL, "/")
	if base == "" {
		base = "https://ai.google.dev"
	}
	return &geminiClient{
		baseURL: base,
		apiKey:  cfg.APIKey,
		http: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (c *geminiClient) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{
		{ID: "gemini-2.5-flash", ModelID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", ModelType: "chat", SupportsStream: true, IsEnabled: true},
		{ID: "gemini-2.5-pro", ModelID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", ModelType: "chat", SupportsStream: true, IsEnabled: true},
		{ID: "gemini-1.5-pro", ModelID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", ModelType: "chat", SupportsStream: true, IsEnabled: true},
		{ID: "gemini-1.5-flash", ModelID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", ModelType: "chat", SupportsStream: true, IsEnabled: true},
	}, nil
}

func (c *geminiClient) ChatCompletion(ctx context.Context, req ChatRequest) error {
	payload, err := buildGeminiPayload(req)
	if err != nil {
		return err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?key=%s",
		c.baseURL, strings.ReplaceAll(req.Model, "/", "%2F"), c.apiKey)
	if !req.Stream {
		url = fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s",
			c.baseURL, strings.ReplaceAll(req.Model, "/", "%2F"), c.apiKey)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderUnreachable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return statusToErr(resp.StatusCode)
	}
	if !req.Stream {
		return nil
	}
	return c.readGeminiStream(resp.Body, req.Streamer)
}

func buildGeminiPayload(req ChatRequest) (map[string]interface{}, error) {
	geminiMsgs := make([]map[string]interface{}, 0, len(req.Messages))
	var systemInstruction string
	for _, m := range req.Messages {
		switch m.Role {
		case "system":
			systemInstruction = m.Content
		default:
			role := "user"
			if m.Role == "assistant" {
				role = "model"
			}
			geminiMsgs = append(geminiMsgs, map[string]interface{}{
				"role":  role,
				"parts": []map[string]string{{"text": m.Content}},
			})
		}
	}
	if len(geminiMsgs) == 0 {
		return nil, fmt.Errorf("gemini requires at least one user message")
	}
	generationConfig := map[string]interface{}{}
	if req.MaxTokens > 0 {
		generationConfig["maxOutputTokens"] = req.MaxTokens
	}
	payload := map[string]interface{}{
		"contents": geminiMsgs,
	}
	if systemInstruction != "" {
		payload["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]string{{"text": systemInstruction}},
		}
	}
	if len(generationConfig) > 0 {
		payload["generationConfig"] = generationConfig
	}
	return payload, nil
}

func (c *geminiClient) readGeminiStream(body io.Reader, streamer SSEStreamer) error {
	sc := lineScanner{r: body}
	var buf []byte
	for {
		line, ok := sc.ScanLine()
		if !ok {
			break
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		buf = append(buf, []byte(data)...)
	}
	if len(buf) == 0 {
		return streamer.Flush()
	}
	var events []geminiStreamEvent
	if err := json.Unmarshal(buf, &events); err != nil {
		var single geminiStreamEvent
		if singleErr := json.Unmarshal(buf, &single); singleErr == nil {
			events = []geminiStreamEvent{single}
		} else {
			return fmt.Errorf("gemini stream parse: %v", err)
		}
	}
	for _, ev := range events {
		for _, cand := range ev.Candidates {
			for _, part := range cand.Content.Parts {
				if part.Text == "" {
					continue
				}
				if err := streamer.PushDelta(part.Text); err != nil {
					return err
				}
				if err := streamer.Flush(); err != nil {
					return err
				}
			}
		}
	}
	return streamer.Flush()
}

type geminiStreamEvent struct {
	Candidates []geminiCandidate `json:"candidates"`
}
type geminiCandidate struct {
	Content geminiContent `json:"content"`
}
type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}
type geminiPart struct {
	Text string `json:"text"`
}
