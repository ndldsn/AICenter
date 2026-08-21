package models

import "time"

// Server represents a managed Linux server
type Server struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	Username       string    `json:"username"`
	AuthType       string    `json:"auth_type"`
	PasswordEnc    string    `json:"-"`
	PrivateKeyEnc  string    `json:"-"`
	AgentConnected bool      `json:"agent_connected"`
	AgentToken     string    `json:"-"`
	GroupID        *string   `json:"group_id,omitempty"`
	Tags           []string  `json:"tags"`
	OsInfo         *OsInfo   `json:"os_info,omitempty"`
	HardwareInfo   *HardwareInfo `json:"hardware_info,omitempty"`
	Status         string    `json:"status"`
	LastHeartbeat  *time.Time `json:"last_heartbeat,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OsInfo struct {
	Distribution string `json:"distribution"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	Hostname     string `json:"hostname"`
}

type HardwareInfo struct {
	CPUModel   string  `json:"cpu_model"`
	CPUCores   int     `json:"cpu_cores"`
	MemoryGB   float64 `json:"memory_gb"`
	DiskGB     float64 `json:"disk_gb"`
}

// ServerGroup represents a group of servers
type ServerGroup struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ServerMetric represents a collected metric from a server
type ServerMetric struct {
	ID         string    `json:"id"`
	ServerID   string    `json:"server_id"`
	MetricName string    `json:"metric_name"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Labels     map[string]string `json:"labels"`
	CollectedAt time.Time `json:"collected_at"`
}
