package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrProviderUnreachable = errors.New("provider unreachable")
var ErrProviderAuth = errors.New("provider auth failed")

type openAICompatClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

type openAIModelList struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func NewOpenAICompatClient(cfg Config) Client {
	return &openAICompatClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		http: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (c *openAICompatClient) ListModels(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderUnreachable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToErr(resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	var list openAIModelList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, err
	}
	out := make([]Model, 0, len(list.Data))
	for _, d := range list.Data {
		out = append(out, Model{
			ID:             d.ID,
			ModelID:        d.ID,
			Name:           d.ID,
			ModelType:      "chat",
			SupportsStream: true,
			IsEnabled:      true,
		})
	}
	return out, nil
}

func (c *openAICompatClient) ChatCompletion(ctx context.Context, req ChatRequest) error {
	payload := map[string]interface{}{
		"model":      req.Model,
		"messages":   req.Messages,
		"stream":     req.Stream,
		"max_tokens": req.MaxTokens,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuth(httpReq)
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrProviderUnreachable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return statusToErr(resp.StatusCode)
	}
	if !req.Stream || req.Streamer == nil {
		return nil
	}

	sc := lineScanner{r: resp.Body}
	for {
		line, ok := sc.ScanLine()
		if !ok {
			break
		}
		if line == "data: [DONE]" {
			return req.Streamer.Flush()
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if strings.TrimSpace(data) == "" {
			continue
		}
		if err := req.Streamer.PushDelta(data); err != nil {
			return err
		}
		if err := req.Streamer.Flush(); err != nil {
			return err
		}
	}
	return req.Streamer.Flush()
}

func (c *openAICompatClient) setAuth(r *http.Request) {
	if c.apiKey != "" {
		r.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

func statusToErr(status int) error {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%w (HTTP %d)", ErrProviderAuth, status)
	default:
		return fmt.Errorf("%w (HTTP %d)", ErrProviderUnreachable, status)
	}
}

// lineScanner is a small line-oriented scanner for streaming SSE bodies.
type lineScanner struct {
	r   io.Reader
	buf [4096]byte
	n   int
}

func (s *lineScanner) ScanLine() (string, bool) {
	for {
		idx := bytes.IndexByte(s.buf[:s.n], '\n')
		if idx >= 0 {
			line := string(s.buf[:idx])
			s.n = copy(s.buf[:], s.buf[idx+1:s.n])
			return line, true
		}
		n, err := s.r.Read(s.buf[s.n:])
		if n > 0 {
			s.n += n
			continue
		}
		if err == io.EOF {
			if s.n > 0 {
				line := string(s.buf[:s.n])
				s.n = 0
				return line, true
			}
			return "", false
		}
		return "", false
	}
}
