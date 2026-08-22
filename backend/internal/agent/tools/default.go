package tools

import (
	"context"
	"encoding/json"
)

// DefaultTools returns a pre-built Registry with the read-only default tool set.
func DefaultTools() *Registry {
	reg := NewRegistry()

	schemaListServers := json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","default":50}}}`)
	schemaInfo := json.RawMessage(`{"type":"object","properties":{}}`)
	schemaEcho := json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`)
	schemaPing := json.RawMessage(`{"type":"object","properties":{"host":{"type":"string"}}}`)

	reg.Register(&Tool{
		Name:        "list_servers",
		Description: "List managed servers (read-only, no side effects).",
		ArgsSchema:  schemaListServers,
		ReadOnly:    true,
		RiskLevel:   "low",
		Call: func(ctx context.Context, args map[string]any) *Result {
			return &Result{
				Ok:      true,
				Status:  "ok",
				Message: "simulated: 3 servers returned",
				Payload: map[string]any{
					"servers": []map[string]any{
						{"id": "srv-001", "name": "web-01"},
						{"id": "srv-002", "name": "db-01"},
						{"id": "srv-003", "name": "worker-01"},
					},
				},
			}
		},
	})

	reg.Register(&Tool{
		Name:        "get_server_info",
		Description: "Get info for a specific server (read-only).",
		ArgsSchema:  schemaPing,
		ReadOnly:    true,
		RiskLevel:   "low",
		Call: func(ctx context.Context, args map[string]any) *Result {
			host, _ := args["host"].(string)
			return &Result{
				Ok:      true,
				Status:  "ok",
				Message: "server info",
				Payload: map[string]any{
					"host": host, "os": "linux", "up": true,
				},
			}
		},
	})

	reg.Register(&Tool{
		Name:        "list_models",
		Description: "List available AI models (read-only).",
		ArgsSchema:  schemaInfo,
		ReadOnly:    true,
		RiskLevel:   "low",
		Call: func(ctx context.Context, args map[string]any) *Result {
			return &Result{
				Ok:     true,
				Status: "ok",
				Payload: map[string]any{
					"models": []map[string]any{
						{"id": "gpt-4o", "name": "GPT-4o"},
						{"id": "deepseek-v3", "name": "DeepSeek-V3"},
					},
				},
			}
		},
	})

	reg.Register(&Tool{
		Name:        "echo",
		Description: "Echo back the provided text (safe, for testing).",
		ArgsSchema:  schemaEcho,
		ReadOnly:    true,
		RiskLevel:   "low",
		Call: func(ctx context.Context, args map[string]any) *Result {
			text, _ := args["text"].(string)
			return &Result{Ok: true, Status: "ok", Message: text}
		},
	})

	reg.Register(&Tool{
		Name:        "restart_service",
		Description: "Restart a service on a server (requires approval).",
		ArgsSchema:  schemaPing,
		ReadOnly:    false,
		Destructive: true,
		RiskLevel:   "high",
		Call: func(ctx context.Context, args map[string]any) *Result {
			host, _ := args["host"].(string)
			return &Result{
				Ok:      true,
				Status:  "ok",
				Message: "simulated restart",
				Payload: map[string]any{"host": host, "restarted": true},
			}
		},
	})

	return reg
}
