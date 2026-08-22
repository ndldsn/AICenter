package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/google/uuid"
)

var (
	ErrRoleNotFound   = errors.New("role not found")
	ErrRoleIsSystem   = errors.New("role is a system role and cannot be modified")
	ErrRoleNameTaken  = errors.New("role name already exists")
)

// RoleRepository persists the role catalog and role→group grants.
type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) InitTables() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS permissions (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			resource TEXT NOT NULL,
			action TEXT NOT NULL,
			group_id TEXT,
			group_name TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
		CREATE TABLE IF NOT EXISTS role_permissions (
			role_id TEXT NOT NULL REFERENCES roles(id),
			permission_id TEXT NOT NULL REFERENCES permissions(id),
			PRIMARY KEY (role_id, permission_id),
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func (r *RoleRepository) EnsurePermission(name, resource, action, groupID, groupName string) error {
	_, err := r.db.Exec(`
		INSERT OR IGNORE INTO permissions (id, name, resource, action, group_id, group_name)
		VALUES (?, ?, ?, ?, ?, ?)
	`, uuid.New().String(), name, resource, action, groupID, groupName)
	return err
}

func (r *RoleRepository) UpsertRole(id, name, description string, isSystem bool) error {
	_, err := r.db.Exec(`
		INSERT OR IGNORE INTO roles (id, name, description, is_system) VALUES (?, ?, ?, ?)
	`, id, name, description, intB(isSystem))
	if err != nil {
		return err
	}
	return r.UpdateRoleDescription(id, description)
}

func (r *RoleRepository) UpdateRoleDescription(id, description string) error {
	_, err := r.db.Exec("UPDATE roles SET description = ? WHERE id = ?", description, id)
	return err
}

func (r *RoleRepository) GrantGroupToRole(roleID string, groupPermissions []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, p := range groupPermissions {
		_, err = tx.Exec(`
			INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
			SELECT ?, name FROM permissions WHERE name = ?
		`, roleID, p)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *RoleRepository) RevokeGroupFromRole(roleID string, groupPermissions []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, p := range groupPermissions {
		_, err = tx.Exec(`
			DELETE FROM role_permissions
			WHERE role_id = ? AND permission_id IN (SELECT id FROM permissions WHERE name = ?)
		`, roleID, p)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *RoleRepository) GetRole(name string) (*models.Role, error) {
	role := &models.Role{}
	err := r.db.QueryRow("SELECT id, name, description, is_system FROM roles WHERE name = ?", name).
		Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) ListRoles() ([]*models.Role, error) {
	rows, err := r.db.Query("SELECT id, name, description, is_system FROM roles ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *RoleRepository) ListPermissions() ([]*models.Permission, error) {
	rows, err := r.db.Query("SELECT id, name, resource, action, group_id FROM permissions ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []*models.Permission
	for rows.Next() {
		p := &models.Permission{}
		err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Group)
		if err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (r *RoleRepository) GrantedPermissions(roleName string) ([]*models.Permission, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.name, p.resource, p.action, p.group_id
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN roles r ON r.id = rp.role_id
		WHERE r.name = ?
		ORDER BY p.name
	`, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []*models.Permission
	for rows.Next() {
		p := &models.Permission{}
		err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Group)
		if err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (r *RoleRepository) RoleGrantedGroups(roleName string) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT p.group_id
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		JOIN roles r ON r.id = rp.role_id
		WHERE r.name = ? AND p.group_id IS NOT NULL AND p.group_id != ''
		ORDER BY p.group_id
	`, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *RoleRepository) Exists(name string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM roles WHERE name = ?)", name).Scan(&exists)
	return exists, err
}

func (r *RoleRepository) UserRoles(userID string) ([]string, error) {
	rows, err := r.db.Query("SELECT role FROM users WHERE id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

// Scan helpers (kept local to this package to avoid cross-import churn).
func intB(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (r *RoleRepository) _now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}