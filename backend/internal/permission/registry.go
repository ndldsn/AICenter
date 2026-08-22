// Package permission provides a registration-time RBAC engine.
//
// Model
// =====
// A Permission is a single named grant (e.g. "servers.read"). Permissions are
// grouped into Groups (e.g. "servers_admin" = {servers.read, servers.write,
// servers.delete}). Roles are granted Groups, persisted in the database
// (role_permissions), and looked up at request time by the middleware.
//
// superadmin is implicitly granted every permission. Other roles get their
// declared groups at startup (seeded into role_permissions) and admins can
// reassign via /roles endpoints without redeploying.
//
// Everything is in code + database — no external policy file.
package permission

import (
	"sort"
	"sync"

	"github.com/aicenter/aicenter/internal/models"
)

// Permission is a single, granular grant.
type Permission struct {
	ID        string // stable id; we use the same string as Name for simplicity
	Name      string // slug: "servers.read"
	Resource  string // "servers"
	Action    string // "read" | "write" | "delete" | "admin"
	Group     string // group this permission belongs to
}

// Group bundles a set of permissions.
type Group struct {
	ID        string
	Name      string
	Desc      string
	PermNames []string // permission names this group grants
}

// Registry holds the permission catalog and group definitions.
type Registry struct {
	mu          sync.Mutex
	perms       map[string]*Permission   // name -> permission
	groups      map[string]*Group        // group id -> group
	permInGroup map[string]map[string]bool // group id -> set of perm names
	defaults    map[string]map[string]bool   // role name -> set of group ids
}

var singleton = &Registry{
	perms:       make(map[string]*Permission),
	groups:      make(map[string]*Group),
	permInGroup: make(map[string]map[string]bool),
	defaults:    make(map[string]map[string]bool),
}

func RegistryInstance() *Registry {
	return singleton
}

// Reset clears the registry. Exists solely for tests.
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.perms = make(map[string]*Permission)
	r.groups = make(map[string]*Group)
	r.permInGroup = make(map[string]map[string]bool)
	r.defaults = make(map[string]map[string]bool)
}

// RegisterPermission declares a single grant.
func (r *Registry) RegisterPermission(p Permission) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.perms[p.Name]; dup {
		return
	}
	r.perms[p.Name] = &p
	if p.Group != "" {
		if r.permInGroup[p.Group] == nil {
			r.permInGroup[p.Group] = make(map[string]bool)
		}
		r.permInGroup[p.Group][p.Name] = true
	}
}

// RegisterGroup declares a named bundle of permissions.
func (r *Registry) RegisterGroup(g Group) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.groups[g.ID]; dup {
		return
	}
	r.groups[g.ID] = &g
	for _, n := range g.PermNames {
		if r.permInGroup[g.ID] == nil {
			r.permInGroup[g.ID] = make(map[string]bool)
		}
		r.permInGroup[g.ID][n] = true
	}
}

// RegisterDefaultGrants declares which groups a role has by default.
func (r *Registry) RegisterDefaultGrants(role string, groupIDs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	grants := make(map[string]bool, len(groupIDs))
	for _, gid := range groupIDs {
		grants[gid] = true
	}
	r.defaults[role] = grants
}

// HasPermission reports whether the role is granted the named permission.
// superadmin always returns true.
func (r *Registry) HasPermission(role string, permName string) bool {
	if role == "superadmin" {
		return true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if gs, ok := r.defaults[role]; ok {
		for gid := range gs {
			if members, ok := r.permInGroup[gid]; ok && members[permName] {
				return true
			}
		}
	}
	return false
}

// GrantedPermissionsNames returns the flat list of permission names a role has.
func (r *Registry) GrantedPermissionsNames(role string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]bool)
	if role == "superadmin" {
		for n := range r.perms {
			out[n] = true
		}
	}
	if gs, ok := r.defaults[role]; ok {
		for gid := range gs {
			for n := range r.permInGroup[gid] {
				out[n] = true
			}
		}
	}
	var names []string
	for n := range out {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// AllPermissions returns the catalog sorted by name.
func (r *Registry) AllPermissions() []models.Permission {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]models.Permission, 0, len(r.perms))
	for _, p := range r.perms {
		out = append(out, models.Permission{
			ID:       p.Name,
			Name:     p.Name,
			Resource: p.Resource,
			Action:   p.Action,
			Group:    p.Group,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// AllGroups returns the group catalog sorted by id.
func (r *Registry) AllGroups() []models.GroupView {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]models.GroupView, 0, len(r.groups))
	for _, g := range r.groups {
		out = append(out, models.GroupView{
			ID:        g.ID,
			Name:      g.Name,
			Desc:      g.Desc,
			PermNames: g.PermNames,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// DefaultRoleGrants returns the baseline role -> group assignment.
func (r *Registry) DefaultRoleGrants() map[string][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string][]string, len(r.defaults))
	for role, gs := range r.defaults {
		var ids []string
		for gid := range gs {
			ids = append(ids, gid)
		}
		sort.Strings(ids)
		out[role] = ids
	}
	return out
}

// AllPermissionNames returns every registered permission name (for seeding).
func (r *Registry) AllPermissionNames() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []string
	for n := range r.perms {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}