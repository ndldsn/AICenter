package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/aicenter/aicenter/internal/database"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/pkg/utils"
	"github.com/google/uuid"
)

// ServerRepository handles server database operations
type ServerRepository struct {
	db *sql.DB
}

// NewServerRepository creates a new server repository
func NewServerRepository() *ServerRepository {
	return &ServerRepository{db: database.Get()}
}

// Create inserts a new server
func (r *ServerRepository) Create(s *models.Server) error {
	s.ID = uuid.New().String()
	s.Status = "unknown"
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	tagsJSON, _ := json.Marshal(s.Tags)

	_, err := r.db.Exec(`
		INSERT INTO servers (id, name, host, port, username, auth_type, password_enc, private_key_enc,
			agent_connected, agent_token, group_id, tags, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Name, s.Host, s.Port, s.Username, s.AuthType,
		s.PasswordEnc, s.PrivateKeyEnc, false, s.AgentToken,
		s.GroupID, tagsJSON, s.Status, s.CreatedAt.Format("2006-01-02 15:04:05"), s.UpdatedAt.Format("2006-01-02 15:04:05"))
	return err
}

// GetByID retrieves a server by ID
func (r *ServerRepository) GetByID(id string) (*models.Server, error) {
	row := r.db.QueryRow(`
		SELECT id, name, host, port, username, auth_type, password_enc, private_key_enc,
			agent_connected, agent_token, group_id, tags, os_info, hardware_info, status, last_heartbeat, created_at, updated_at
		FROM servers WHERE id = ?`, id)

	return r.scanServer(row)
}

// List retrieves all servers
func (r *ServerRepository) List(offset, limit int) ([]*models.Server, int64, error) {
	var total int64
	if err := r.db.QueryRow("SELECT COUNT(*) FROM servers").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
		SELECT id, name, host, port, username, auth_type, password_enc, private_key_enc,
			agent_connected, agent_token, group_id, tags, os_info, hardware_info, status, last_heartbeat, created_at, updated_at
		FROM servers ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	servers := []*models.Server{}
	for rows.Next() {
		s, err := r.scanServer(rows)
		if err != nil {
			return nil, 0, err
		}
		servers = append(servers, s)
	}
	return servers, total, nil
}

// Update updates a server
func (r *ServerRepository) Update(id string, updates map[string]interface{}) error {
	// Dynamic update
	query := "UPDATE servers SET updated_at = ?"
	args := []interface{}{time.Now()}

	if v, ok := updates["name"]; ok {
		query += ", name = ?"
		args = append(args, v)
	}
	if v, ok := updates["host"]; ok {
		query += ", host = ?"
		args = append(args, v)
	}
	if v, ok := updates["port"]; ok {
		query += ", port = ?"
		args = append(args, v)
	}
	if v, ok := updates["username"]; ok {
		query += ", username = ?"
		args = append(args, v)
	}
	if v, ok := updates["auth_type"]; ok {
		query += ", auth_type = ?"
		args = append(args, v)
	}
	if v, ok := updates["password_enc"]; ok {
		query += ", password_enc = ?"
		args = append(args, v)
	}
	if v, ok := updates["private_key_enc"]; ok {
		query += ", private_key_enc = ?"
		args = append(args, v)
	}
	if v, ok := updates["group_id"]; ok {
		query += ", group_id = ?"
		args = append(args, v)
	}
	if v, ok := updates["tags"]; ok {
		query += ", tags = ?"
		args = append(args, v)
	}
	if v, ok := updates["status"]; ok {
		query += ", status = ?"
		args = append(args, v)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}

// Delete removes a server
func (r *ServerRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM servers WHERE id = ?", id)
	return err
}

// UpdateStatus updates server status
func (r *ServerRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec("UPDATE servers SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now(), id)
	return err
}

// UpdateHeartbeat updates last heartbeat time
func (r *ServerRepository) UpdateHeartbeat(id string) error {
	now := time.Now()
	_, err := r.db.Exec("UPDATE servers SET last_heartbeat = ?, updated_at = ? WHERE id = ?",
		now, now, id)
	return err
}

// UpdateAgentConnected updates agent connection status
func (r *ServerRepository) UpdateAgentConnected(id string, connected bool) error {
	_, err := r.db.Exec("UPDATE servers SET agent_connected = ?, updated_at = ? WHERE id = ?",
		connected, time.Now(), id)
	return err
}

// UpdateSystemInfo updates OS and hardware info
func (r *ServerRepository) UpdateSystemInfo(id string, osInfo, hwInfo string) error {
	_, err := r.db.Exec("UPDATE servers SET os_info = ?, hardware_info = ?, updated_at = ? WHERE id = ?",
		osInfo, hwInfo, time.Now(), id)
	return err
}

func (r *ServerRepository) scanServer(row interface {
	Scan(dest ...interface{}) error
}) (*models.Server, error) {
	var s models.Server
	var tagsJSON []byte
	var groupID sql.NullString
	var osInfo sql.NullString
	var hwInfo sql.NullString
	var lastHB sql.NullString
	var createdAt sql.NullString
	var updatedAt sql.NullString
	var passwordEnc sql.NullString
	var privateKeyEnc sql.NullString
	var agentToken sql.NullString

	err := row.Scan(
		&s.ID, &s.Name, &s.Host, &s.Port, &s.Username, &s.AuthType,
		&passwordEnc, &privateKeyEnc, &s.AgentConnected, &agentToken,
		&groupID, &tagsJSON, &osInfo, &hwInfo, &s.Status, &lastHB,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if passwordEnc.Valid {
		s.PasswordEnc = passwordEnc.String
	}
	if privateKeyEnc.Valid {
		s.PrivateKeyEnc = privateKeyEnc.String
	}
	if agentToken.Valid {
		s.AgentToken = agentToken.String
	}
	if groupID.Valid {
		s.GroupID = &groupID.String
	}
	if osInfo.Valid && osInfo.String != "" {
		s.OsInfo = &models.OsInfo{}
		json.Unmarshal([]byte(osInfo.String), s.OsInfo)
	}
	if hwInfo.Valid && hwInfo.String != "" {
		s.HardwareInfo = &models.HardwareInfo{}
		json.Unmarshal([]byte(hwInfo.String), s.HardwareInfo)
	}
	if lastHB.Valid && lastHB.String != "" {
		t, err := utils.ParseTimestamp(lastHB.String)
		if err == nil && !t.IsZero() {
			s.LastHeartbeat = &t
		}
	}
	if createdAt.Valid && createdAt.String != "" {
		t, err := utils.ParseTimestamp(createdAt.String)
		if err == nil {
			s.CreatedAt = t
		}
	}
	if updatedAt.Valid && updatedAt.String != "" {
		t, err := utils.ParseTimestamp(updatedAt.String)
		if err == nil {
			s.UpdatedAt = t
		}
	}

	json.Unmarshal(tagsJSON, &s.Tags)

	return &s, nil
}

// GetServerGroups retrieves all server groups
func (r *ServerRepository) GetServerGroups() ([]*models.ServerGroup, error) {
	rows, err := r.db.Query("SELECT id, name, description, parent_id, created_at FROM server_groups ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*models.ServerGroup
	for rows.Next() {
		var g models.ServerGroup
		var parentID sql.NullString
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &parentID, &g.CreatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			g.ParentID = &parentID.String
		}
		groups = append(groups, &g)
	}
	return groups, nil
}

// CreateGroup creates a new server group
func (r *ServerRepository) CreateGroup(g *models.ServerGroup) error {
	g.ID = uuid.New().String()
	g.CreatedAt = time.Now()
	_, err := r.db.Exec("INSERT INTO server_groups (id, name, description, parent_id, created_at) VALUES (?, ?, ?, ?, ?)",
		g.ID, g.Name, g.Description, g.ParentID, g.CreatedAt)
	return err
}

// DeleteGroup deletes a server group
func (r *ServerRepository) DeleteGroup(id string) error {
	_, err := r.db.Exec("DELETE FROM server_groups WHERE id = ?", id)
	return err
}
