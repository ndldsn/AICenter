package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/pkg/cache"
	"github.com/aicenter/aicenter/internal/pkg/crypto"
	"github.com/aicenter/aicenter/internal/pkg/ssh"
	"github.com/aicenter/aicenter/internal/repository"
)

// ServerService handles server business logic
type ServerService struct {
	repo      *repository.ServerRepository
	cache     cache.Store
	encryptFn func(string) (string, error)
}

// NewServerService creates a new server service
func NewServerService() *ServerService {
	return NewServerServiceWithEncrypt(nil)
}

// NewServerServiceWithEncrypt creates a server service with a custom encrypt
// function. If encryptFn is nil, crypto.Encrypt (AES-256-GCM) is used.
func NewServerServiceWithEncrypt(encryptFn func(string) (string, error)) *ServerService {
	if encryptFn == nil {
		encryptFn = crypto.Encrypt
	}
	return &ServerService{repo: repository.NewServerRepository(), encryptFn: encryptFn}
}

// NewServerServiceWithCache creates a new server service with a cache store
// for hot read paths (server list, get-by-id, groups).
func NewServerServiceWithCache(c cache.Store) *ServerService {
	s := NewServerServiceWithEncrypt(nil)
	if c == nil {
		c = cache.NewMemory(512)
	}
	s.cache = c
	return s
}

// invalidateServerCache clears the cached server entries on a write.
func (s *ServerService) invalidateServerCache() {
	if s.cache == nil {
		return
	}
	s.cache.Delete(cache.ServerListKey())
	s.cache.Clear()
}

// invalidateServer clears just one server's cache entry (used by Update/Delete).
func (s *ServerService) invalidateServer(id string) {
	if s.cache != nil {
		s.cache.Delete(cache.ServerKey(id))
		s.cache.Delete(cache.ServerListKey())
	}
}

// encrypt applies s.encryptFn; returns the plaintext unchanged on error so
// callers can decide how to handle partial failures.
func (s *ServerService) encrypt(v string) string {
	enc, err := s.encryptFn(v)
	if err != nil {
		return v
	}
	return enc
}

// CreateServer creates a new server
func (s *ServerService) CreateServer(req *CreateServerRequest) (*models.Server, error) {
	server := &models.Server{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		AuthType: req.AuthType,
		PasswordEnc:   s.encrypt(req.Password),
		PrivateKeyEnc: s.encrypt(req.PrivateKey),
		Tags:     req.Tags,
	}
	if req.GroupID != "" {
		server.GroupID = &req.GroupID
	}

	if err := s.repo.Create(server); err != nil {
		return nil, fmt.Errorf("create server: %w", err)
	}
	s.invalidateServerCache()
	return server, nil
}

// GetServer retrieves a server by ID
func (s *ServerService) GetServer(id string) (*models.Server, error) {
	if s.cache != nil {
		key := cache.ServerKey(id)
		if v, ok := s.cache.Get(key); ok {
			return v.(*models.Server), nil
		}
	}
	sv, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Set(cache.ServerKey(id), sv, cache.DefaultTTL)
	}
	return sv, nil
}

// ListServers retrieves all servers
func (s *ServerService) ListServers(page, limit int) ([]*models.Server, int64, error) {
	offset := (page - 1) * limit
	// Only cache the unpaginated full-list reads (dashboard/server list use
	// large page sizes); cache the result by (page,limit) key.
	if s.cache != nil && page <= 1 && limit >= 100 {
		key := "servers:list:all"
		if v, ok := s.cache.Get(key); ok {
			all := v.([]*models.Server)
			return all, int64(len(all)), nil
		}
		all, _, err := s.repo.List(0, 10000)
		if err != nil {
			return nil, 0, err
		}
		s.cache.Set(key, all, cache.DefaultTTL)
		return all, int64(len(all)), nil
	}
	return s.repo.List(offset, limit)
}

// UpdateServer updates a server
func (s *ServerService) UpdateServer(id string, req *UpdateServerRequest) error {
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Host != "" {
		updates["host"] = req.Host
	}
	if req.Port != 0 {
		updates["port"] = req.Port
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.AuthType != "" {
		updates["auth_type"] = req.AuthType
	}
	if req.Password != "" {
		updates["password_enc"] = s.encrypt(req.Password)
	}
	if req.PrivateKey != "" {
		updates["private_key_enc"] = s.encrypt(req.PrivateKey)
	}
	if req.GroupID != "" {
		updates["group_id"] = req.GroupID
	}
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		updates["tags"] = tagsJSON
	}
	if err := s.repo.Update(id, updates); err != nil {
		return err
	}
	s.invalidateServer(id)
	return nil
}

// DeleteServer deletes a server
func (s *ServerService) DeleteServer(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.invalidateServer(id)
	return nil
}

// TestConnection tests SSH connection to a server
func (s *ServerService) TestConnection(id string) (*ConnectionTestResult, error) {
	server, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.testConnection(server)
}

// TestNewConnection tests connection for new server data
func (s *ServerService) TestNewConnection(req *CreateServerRequest) (*ConnectionTestResult, error) {
	server := &models.Server{
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		AuthType: req.AuthType,
		PasswordEnc: req.Password,
		PrivateKeyEnc: req.PrivateKey,
	}
	return s.testConnection(server)
}

func (s *ServerService) testConnection(server *models.Server) (*ConnectionTestResult, error) {
	result := &ConnectionTestResult{
		Success:   false,
		Timestamp: time.Now(),
	}

	// Test SSH version
	banner, err := ssh.GetSSHVersion(server.Host, server.Port)
	if err != nil {
		result.Message = fmt.Sprintf("Cannot reach SSH server: %v", err)
		return result, nil
	}
	result.SSHBanner = banner

	// Test full SSH connection
	password, _ := crypto.Decrypt(server.PasswordEnc)
	privateKey, _ := crypto.Decrypt(server.PrivateKeyEnc)
	cfg := &ssh.Config{
		Host:       server.Host,
		Port:       server.Port,
		Username:   server.Username,
		AuthType:   server.AuthType,
		Password:   password,
		PrivateKey: privateKey,
		Timeout:    10 * time.Second,
	}

	client := ssh.NewClient(cfg)
	if err := client.Connect(); err != nil {
		result.Message = fmt.Sprintf("SSH connection failed: %v", err)
		return result, nil
	}
	defer client.Close()

	// Get system info
	stdout, stderr, err := client.Run("uname -a && cat /etc/os-release 2>/dev/null && nproc && free -m && df -h /")
	if err != nil {
		result.Message = fmt.Sprintf("Connected but command failed: %v\n%s", err, stderr)
		return result, nil
	}

	result.Success = true
	result.Message = "Connection successful"
	result.SystemInfo = s.parseSystemInfo(stdout)
	return result, nil
}

func (s *ServerService) parseSystemInfo(output string) *SystemInfo {
	info := &SystemInfo{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse uname
		if strings.Contains(line, "Linux") && info.Kernel == "" {
			parts := strings.SplitN(line, " ", 4)
			if len(parts) >= 3 {
				info.Hostname = parts[1]
				info.Kernel = parts[2]
			}
		}

		// Parse OS release
		if strings.HasPrefix(line, "NAME=") {
			info.OS = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
		}
		if strings.HasPrefix(line, "VERSION=") {
			info.OSVersion = strings.Trim(strings.TrimPrefix(line, "VERSION="), "\"")
		}

		// Parse CPU cores
		if _, err := fmt.Sscanf(line, "%d", &info.CPUCores); err == nil && info.CPUCores > 0 && info.CPUCores < 10000 {
			// Likely CPU count
		}

		// Parse memory
		if strings.HasPrefix(line, "Mem:") {
			fmt.Sscanf(line, "Mem: %f", &info.MemoryGB)
			info.MemoryGB = info.MemoryGB / 1024 // Convert MB to GB
		}

		// Parse disk usage
		if strings.Contains(line, "/ ") && strings.Contains(line, "%") {
			fmt.Sscanf(line, "%f", &info.DiskUsagePercent)
		}
	}

	return info
}

// Request/Response types

type CreateServerRequest struct {
	Name       string   `json:"name" binding:"required"`
	Host       string   `json:"host" binding:"required"`
	Port       int      `json:"port"`
	Username   string   `json:"username"`
	AuthType   string   `json:"auth_type"`
	Password   string   `json:"password"`
	PrivateKey string   `json:"private_key"`
	GroupID    string   `json:"group_id"`
	Tags       []string `json:"tags"`
}

type UpdateServerRequest struct {
	Name       string   `json:"name"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	Username   string   `json:"username"`
	AuthType   string   `json:"auth_type"`
	Password   string   `json:"password"`
	PrivateKey string   `json:"private_key"`
	GroupID    string   `json:"group_id"`
	Tags       []string `json:"tags"`
}

type ConnectionTestResult struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	SSHBanner  string      `json:"ssh_banner,omitempty"`
	SystemInfo *SystemInfo `json:"system_info,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

type SystemInfo struct {
	OS               string  `json:"os"`
	OSVersion        string  `json:"os_version"`
	Hostname         string  `json:"hostname"`
	Kernel           string  `json:"kernel"`
	CPUCores         int     `json:"cpu_cores"`
	MemoryGB         float64 `json:"memory_gb"`
	DiskUsagePercent float64 `json:"disk_usage_percent"`
}

// GetServerGroups retrieves server groups
func (s *ServerService) GetServerGroups() ([]*models.ServerGroup, error) {
	return s.repo.GetServerGroups()
}

// CreateServerGroup creates a server group
func (s *ServerService) CreateServerGroup(name, description string, parentID *string) (*models.ServerGroup, error) {
	group := &models.ServerGroup{
		Name:        name,
		Description: description,
		ParentID:    parentID,
	}
	if err := s.repo.CreateGroup(group); err != nil {
		return nil, err
	}
	return group, nil
}

// DeleteServerGroup deletes a server group
func (s *ServerService) DeleteServerGroup(id string) error {
	return s.repo.DeleteGroup(id)
}

// GenerateAgentToken generates a new agent registration token
func (s *ServerService) GenerateAgentToken(serverID string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate agent token: %w", err)
	}
	token := fmt.Sprintf("agent-%x", buf)
	return token, nil
}
