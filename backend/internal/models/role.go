package models

// Role represents a named role with an associated permission set.
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

// Permission is a single, granular grant shown to admins.
type Permission struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Resource  string `json:"resource"`
	Action    string `json:"action"`
	Group     string `json:"group"`
}

// GroupView is the human-facing bundle of permissions (e.g. "servers_admin").
type GroupView struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Desc      string   `json:"description"`
	PermNames []string `json:"permissions"`
}

// RolePermissionView is what GET /roles/:id returns: role + its granted
// groups, each expanded into the underlying permissions.
type RolePermissionView struct {
	Role           *Role      `json:"role"`
	GrantedGroups  []GroupView `json:"granted_groups"`
}

// RoleCreateRequest is the writable surface for POST /roles.
type RoleCreateRequest struct {
	Name   string `json:"name"`
	Desc   string `json:"description"`
	Groups []string `json:"groups"` // group ids to grant
}

// RoleUpdateRequest is the writable surface for PUT /roles/:id.
// System roles cannot be deleted or renamed; only description and group grants change.
type RoleUpdateRequest struct {
	Desc   string `json:"description"`
	Groups []string `json:"groups"`
}