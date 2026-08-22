package middleware

import (
	"net/url"
	"strings"
	"testing"
)

func TestClassifyAction(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   string
	}{
		// servers
		{"GET", "/servers", "servers.list"},
		{"POST", "/servers", "servers.create"},
		{"GET", "/servers/abc", "servers.get"},
		{"PUT", "/servers/abc", "servers.update"},
		{"DELETE", "/servers/abc", "servers.delete"},
		{"POST", "/servers/abc/restart", "servers.restart"},
		{"POST", "/servers/abc/ping", "servers.ping"},
		// docker containers
		{"GET", "/docker/containers", "docker_containers.list"},
		{"POST", "/docker/containers/start/abc", "docker_containers.start"},
		{"POST", "/docker/containers/stop/abc", "docker_containers.stop"},
		{"POST", "/docker/containers/restart/abc", "docker_containers.restart"},
		{"POST", "/docker/containers/prune", "docker_containers.prune"},
		{"POST", "/docker/containers/run", "docker_containers.run"},
		{"POST", "/docker/containers/abc/exec", "docker_containers.exec"},
		{"POST", "/docker/containers/abc/logs", "docker_containers.logs"},
		// docker images
		{"GET", "/docker/images", "docker_images.list"},
		{"GET", "/docker/images?pull=true", "docker_images.pull"},
		{"POST", "/docker/images/abc/prune", "docker_images.prune"},
		{"POST", "/docker/images/abc/build", "docker_images.build"},
		// docker volumes
		{"POST", "/docker/volumes", "docker_volumes.create"},
		{"GET", "/docker/volumes/prune", "docker_volumes.prune"},
		{"DELETE", "/docker/volumes/abc", "docker_volumes.delete"},
		// ai
		{"GET", "/ai/providers", "ai_providers.list"},
		{"GET", "/ai/providers/abc", "ai_providers.get"},
		{"POST", "/ai/providers", "ai_providers.create"},
		{"PUT", "/ai/providers/abc", "ai_providers.update"},
		{"DELETE", "/ai/providers/abc", "ai_providers.delete"},
		{"GET", "/ai/models/provider-abc", "ai_models.list"},
		{"POST", "/ai/chat", "ai_chat.send"},
		// agents
		{"GET", "/agents", "agents.list"},
		{"POST", "/agents", "agents.create"},
		{"GET", "/agents/abc", "agents.get"},
		{"PUT", "/agents/abc", "agents.update"},
		{"DELETE", "/agents/abc", "agents.delete"},
		{"POST", "/agents/abc/sessions", "agents_sessions.create"},
		{"GET", "/agents/sessions", "agents_sessions.list"},
		{"GET", "/agents/sessions/abc", "agents_sessions.get"},
		{"POST", "/agents/sessions/abc/messages", "agents_sessions.send"},
		// audit & approvals
		{"GET", "/audit-logs", "audit.list"},
		{"GET", "/approvals", "approvals.list"},
		{"GET", "/approvals/abc", "approvals.get"},
		{"POST", "/approvals/abc/approve", "approvals.approve"},
		{"POST", "/approvals/abc/reject", "approvals.reject"},
		// tasks
		{"GET", "/tasks", "tasks.list"},
		{"POST", "/tasks", "tasks.create"},
		{"GET", "/tasks/abc", "tasks.get"},
		{"POST", "/tasks/abc/retry", "tasks.retry"},
		{"POST", "/tasks/abc/cancel", "tasks.cancel"},
		// terminal
		{"POST", "/terminal/sessions", "terminal.create"},
		{"GET", "/terminal/sessions", "terminal.list"},
		{"POST", "/terminal/sessions/abc/close", "terminal.close"},
		// roles & permissions
		{"GET", "/roles", "roles.list"},
		{"POST", "/roles", "roles.create"},
		{"GET", "/roles/abc", "roles.get"},
		{"PUT", "/roles/abc", "roles.update"},
		{"DELETE", "/roles/abc", "roles.delete"},
		{"POST", "/roles/abc/permissions", "roles_permissions.assign"},
		{"GET", "/roles/groups", "roles.groups"},
		{"GET", "/permissions", "permissions.list"},
		// users
		{"GET", "/users", "users.list"},
		{"GET", "/users/abc", "users.get"},
		{"PUT", "/users/abc", "users.update"},
		{"PATCH", "/users/abc/role", "users.role"},
		{"DELETE", "/users/abc", "users.delete"},
		// auth
		{"POST", "/auth/login", "auth.login"},
		{"POST", "/auth/register", "auth.register"},
		{"POST", "/auth/refresh", "auth.refresh"},
		{"GET", "/auth/me", "auth.me"},
		// settings
		{"GET", "/settings", "settings.get"},
		{"PUT", "/settings", "settings.update"},
		// others
		{"GET", "/dashboard", "dashboard.get"},
	}

	for _, tt := range tests {
		q := url.Values{}
		if idx := strings.Index(tt.path, "?"); idx >= 0 {
			q, _ = url.ParseQuery(tt.path[idx+1:])
		}
		got := classifyAction(tt.method, tt.path, q)
		if got != tt.want {
			t.Errorf("classifyAction(%q, %q) = %q, want %q", tt.method, tt.path, got, tt.want)
		}
	}
}

func TestSanitizeBody(t *testing.T) {
	body := `{"password":"secret","api_key":"sk-123","username":"alice","role":"admin"}`
	san := sanitizeBody([]byte(body))
	if containsString(string(san), "secret") {
		t.Error("password not redacted")
	}
	if containsString(string(san), "sk-123") {
		t.Error("api_key not redacted")
	}
	if !containsString(string(san), "alice") {
		t.Error("username should remain")
	}
}

func containsString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}