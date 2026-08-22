package permission

// All registered permissions (single source of truth for group membership
// and /permissions endpoint). Kept as functions so the order is stable
// regardless of future additions.
func Permissions() []Permission {
	return []Permission{
		{ID: "servers.read", Name: "servers.read", Resource: "servers", Action: "read", Group: "servers_view"},
		{ID: "servers.write", Name: "servers.write", Resource: "servers", Action: "write", Group: "servers_admin"},
		{ID: "servers.delete", Name: "servers.delete", Resource: "servers", Action: "delete", Group: "servers_admin"},
		{ID: "ai.providers.read", Name: "ai.providers.read", Resource: "ai.providers", Action: "read", Group: "ai_view"},
		{ID: "ai.providers.write", Name: "ai.providers.write", Resource: "ai.providers", Action: "write", Group: "ai_admin"},
		{ID: "ai.models.read", Name: "ai.models.read", Resource: "ai.models", Action: "read", Group: "ai_view"},
		{ID: "ai.chat", Name: "ai.chat", Resource: "ai", Action: "write", Group: "ai_admin"},
		{ID: "agents.read", Name: "agents.read", Resource: "agents", Action: "read", Group: "agents_manage"},
		{ID: "agents.write", Name: "agents.write", Resource: "agents", Action: "write", Group: "agents_manage"},
		{ID: "agents.sessions", Name: "agents.sessions", Resource: "agents.sessions", Action: "write", Group: "agents_manage"},
		{ID: "notifications.manage", Name: "notifications.manage", Resource: "notifications", Action: "write", Group: "notifications_admin"},
		{ID: "monitor.read", Name: "monitor.read", Resource: "monitor", Action: "read", Group: "monitor_view"},
		{ID: "monitor.write", Name: "monitor.write", Resource: "monitor", Action: "write", Group: "monitor_admin"},
		{ID: "terminal.create", Name: "terminal.create", Resource: "terminal", Action: "write", Group: "terminal"},
		{ID: "terminal.list", Name: "terminal.list", Resource: "terminal", Action: "read", Group: "terminal"},
		{ID: "users.manage", Name: "users.manage", Resource: "users", Action: "write", Group: "users_manage"},
		{ID: "roles.read", Name: "roles.read", Resource: "roles", Action: "read", Group: "roles_view"},
		{ID: "roles.manage", Name: "roles.manage", Resource: "roles", Action: "write", Group: "roles_manage"},
		{ID: "approvals.manage", Name: "approvals.manage", Resource: "approvals", Action: "write", Group: "approvals_manage"},
		{ID: "tasks.manage", Name: "tasks.manage", Resource: "tasks", Action: "write", Group: "tasks_manage"},
		{ID: "audit.read", Name: "audit.read", Resource: "audit", Action: "read", Group: "audit_view"},
		{ID: "settings.read", Name: "settings.read", Resource: "settings", Action: "read", Group: "settings_view"},
		{ID: "settings.write", Name: "settings.write", Resource: "settings", Action: "write", Group: "settings_admin"},
	}
}

func Groups() []Group {
	return []Group{
		{ID: "servers_view", Name: "servers_view", Desc: "View servers", PermNames: []string{"servers.read"}},
		{ID: "servers_admin", Name: "servers_admin", Desc: "Manage servers (read/write/delete)", PermNames: []string{"servers.read", "servers.write", "servers.delete"}},
		{ID: "ai_view", Name: "ai_view", Desc: "View AI providers/models", PermNames: []string{"ai.providers.read", "ai.models.read"}},
		{ID: "ai_admin", Name: "ai_admin", Desc: "Manage AI providers/models + chat", PermNames: []string{"ai.providers.read", "ai.providers.write", "ai.models.read", "ai.chat"}},
		{ID: "agents_manage", Name: "agents_manage", Desc: "Manage agents and sessions", PermNames: []string{"agents.read", "agents.write", "agents.sessions"}},
		{ID: "notifications_admin", Name: "notifications_admin", Desc: "Manage notification channels/templates", PermNames: []string{"notifications.manage"}},
		{ID: "monitor_view", Name: "monitor_view", Desc: "View metrics/alerts", PermNames: []string{"monitor.read"}},
		{ID: "monitor_admin", Name: "monitor_admin", Desc: "Manage alert rules", PermNames: []string{"monitor.read", "monitor.write"}},
		{ID: "terminal", Name: "terminal", Desc: "Terminal sessions", PermNames: []string{"terminal.create", "terminal.list"}},
		{ID: "users_manage", Name: "users_manage", Desc: "Manage users", PermNames: []string{"users.manage"}},
		{ID: "roles_view", Name: "roles_view", Desc: "View roles catalog", PermNames: []string{"roles.read"}},
		{ID: "roles_manage", Name: "roles_manage", Desc: "Manage roles", PermNames: []string{"roles.read", "roles.manage"}},
		{ID: "approvals_manage", Name: "approvals_manage", Desc: "Approve/reject tool calls", PermNames: []string{"approvals.manage"}},
		{ID: "tasks_manage", Name: "tasks_manage", Desc: "Manage tasks", PermNames: []string{"tasks.manage"}},
		{ID: "audit_view", Name: "audit_view", Desc: "View audit logs", PermNames: []string{"audit.read"}},
		{ID: "settings_view", Name: "settings_view", Desc: "View settings", PermNames: []string{"settings.read"}},
		{ID: "settings_admin", Name: "settings_admin", Desc: "Manage settings", PermNames: []string{"settings.read", "settings.write"}},
	}
}

func DefaultRoleGrants() map[string][]string {
	return map[string][]string{
		"admin":     {"servers_admin", "ai_admin", "notifications_admin", "users_manage", "roles_manage", "approvals_manage", "tasks_manage", "audit_view", "settings_admin"},
		"operator":  {"servers_admin", "ai_view", "agents_manage", "monitor_admin", "terminal", "approvals_manage", "tasks_manage", "audit_view", "roles_view"},
		"viewer":    {"servers_view", "ai_view", "agents_manage", "monitor_view", "terminal", "audit_view", "settings_view", "roles_view"},
	}
}