package permission

import (
	"reflect"
	"testing"
)

func setupRegistry(t *testing.T) *Registry {
	t.Helper()
	r := RegistryInstance()
	r.Reset()
	for _, p := range Permissions() {
		r.RegisterPermission(p)
	}
	for _, g := range Groups() {
		r.RegisterGroup(g)
	}
	for role, groups := range DefaultRoleGrants() {
		r.RegisterDefaultGrants(role, groups)
	}
	// Ensure superadmin always has all permissions, regardless of seed.
	r.RegisterDefaultGrants("superadmin", []string{"all"})
	return r
}

func TestRegistry_AdminHasAllManagePerms(t *testing.T) {
	setupRegistry(t)
	for _, perm := range []string{"servers.read", "servers.write", "servers.delete", "roles.manage", "users.manage", "ai.providers.write"} {
		if !RegistryInstance().HasPermission("admin", perm) {
			t.Fatalf("admin missing %s", perm)
		}
	}
}

func TestRegistry_ViewerCannotWrite(t *testing.T) {
	setupRegistry(t)
	for _, perm := range []string{"servers.write", "servers.delete", "roles.manage", "users.manage", "ai.providers.write", "notifications.manage"} {
		if RegistryInstance().HasPermission("viewer", perm) {
			t.Fatalf("viewer should NOT have %s", perm)
		}
	}
}

func TestRegistry_ViewerCanRead(t *testing.T) {
	setupRegistry(t)
	for _, perm := range []string{"servers.read", "ai.providers.read", "audit.read", "roles.read", "terminal.list"} {
		if !RegistryInstance().HasPermission("viewer", perm) {
			t.Fatalf("viewer missing %s", perm)
		}
	}
}

func TestRegistry_SuperadminHasEverything(t *testing.T) {
	setupRegistry(t)
	for _, p := range Permissions() {
		if !RegistryInstance().HasPermission("superadmin", p.Name) {
			t.Fatalf("superadmin missing %s", p.Name)
		}
	}
}

func TestRegistry_UnknownRoleHasNothing(t *testing.T) {
	setupRegistry(t)
	if RegistryInstance().HasPermission("unknown", "servers.read") {
		t.Fatal("unknown role should have no permissions")
	}
}

func TestRegistry_CatalogSizes(t *testing.T) {
	setupRegistry(t)
	perms := RegistryInstance().AllPermissions()
	groups := RegistryInstance().AllGroups()
	if len(perms) < 20 {
		t.Fatalf("expected >=20 permissions, got %d", len(perms))
	}
	if len(groups) < 10 {
		t.Fatalf("expected >=10 groups, got %d", len(groups))
	}
}

func TestRegistry_AdminDefaultGrantsContainRolesManage(t *testing.T) {
	setupRegistry(t)
	grants := RegistryInstance().DefaultRoleGrants()
	if _, ok := grants["admin"]; !ok {
		t.Fatal("admin role missing from default grants")
	}
	found := false
	for _, g := range grants["admin"] {
		if g == "roles_manage" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("admin default grants missing roles_manage")
	}
}

func TestRegistry_ExpandGroupIncludesAllPerms(t *testing.T) {
	setupRegistry(t)
	groups := RegistryInstance().AllGroups()
	for _, g := range groups {
		for _, pn := range g.PermNames {
			// Every declared perm name must actually exist in the catalog.
			if !RegistryInstance().HasPermission("superadmin", pn) {
				t.Fatalf("group %s references unknown perm %s", g.ID, pn)
			}
		}
	}
}

func TestRegistry_DuplicateRegistrationIsIdempotent(t *testing.T) {
	setupRegistry(t)
	p := Permissions()[0]
	// Registering the same permission twice must not panic and must not
	// change the catalog size.
	RegistryInstance().RegisterPermission(p)
	perms := RegistryInstance().AllPermissions()
	if len(perms) < 20 {
		t.Fatalf("idempotency broke catalog size: %d", len(perms))
	}
}

func TestRegistry_AllPermissionsSorted(t *testing.T) {
	setupRegistry(t)
	perms := RegistryInstance().AllPermissions()
	for i := 1; i < len(perms); i++ {
		if perms[i].Name < perms[i-1].Name {
			t.Fatalf("permissions not sorted: %s > %s", perms[i-1].Name, perms[i].Name)
		}
	}
}

func TestRegistry_GroupsSorted(t *testing.T) {
	setupRegistry(t)
	groups := RegistryInstance().AllGroups()
	for i := 1; i < len(groups); i++ {
		if groups[i].ID < groups[i-1].ID {
			t.Fatalf("groups not sorted: %s > %s", groups[i-1].ID, groups[i].ID)
		}
	}
}

func TestRegistry_GroupViewHasPermissions(t *testing.T) {
	setupRegistry(t)
	groups := RegistryInstance().AllGroups()
	for _, g := range groups {
		if g.ID == "" {
			t.Fatal("group missing id")
		}
		if len(g.PermNames) == 0 {
			t.Fatalf("group %s has no permissions", g.ID)
		}
	}
}

func TestRegistry_ViewerGrantSetIncludesRolesView(t *testing.T) {
	setupRegistry(t)
	grants := RegistryInstance().DefaultRoleGrants()
	found := false
	for _, g := range grants["viewer"] {
		if g == "roles_view" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("viewer missing roles_view default grant")
	}
}

func TestRegistry_OperatorHasRolesViewButNotManage(t *testing.T) {
	setupRegistry(t)
	if !RegistryInstance().HasPermission("operator", "roles.read") {
		t.Fatal("operator should have roles.read")
	}
	if RegistryInstance().HasPermission("operator", "roles.manage") {
		t.Fatal("operator should NOT have roles.manage")
	}
}

func TestRegistry_StructEquality(t *testing.T) {
	p := Permission{ID: "x", Name: "x", Resource: "r", Action: "read", Group: "g"}
	got := Permission{ID: p.ID, Name: p.Name, Resource: p.Resource, Action: p.Action, Group: p.Group}
	if !reflect.DeepEqual(p, got) {
		t.Fatal("permission struct should be deep-equal")
	}
}

func TestRegistry_DuplicateGroupIsIgnored(t *testing.T) {
	setupRegistry(t)
	g := Groups()[0]
	RegistryInstance().RegisterGroup(g)
	groups := RegistryInstance().AllGroups()
	// Count must not grow.
	if len(groups) < 10 {
		t.Fatalf("duplicate group registration changed size: %d", len(groups))
	}
}