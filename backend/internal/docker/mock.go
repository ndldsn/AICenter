package docker

import (
	"context"
	"sync"
)

// MockClient is an in-memory Docker client used when no local daemon is available.
// It keeps mutable state so start/stop/delete behave realistically for testing.
type MockClient struct {
	mu         sync.RWMutex
	containers []Container
	details    map[string]*ContainerDetail
}

// NewMockClient builds a client pre-seeded with realistic container data.
func NewMockClient() *MockClient {
	c := &MockClient{
		details: make(map[string]*ContainerDetail),
	}
	seed := []Container{
		{
			ID:      "abc123web01",
			Name:    "web-nginx",
			Image:   "nginx:1.27",
			ImageID: "sha256:nginx1",
			Command: "nginx -g 'daemon off;'",
			State:   "running",
			Status:  "Up 3 hours",
			Ports:   []Port{{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8080, Type: "tcp"}},
			Labels:  map[string]string{"app": "web", "tier": "frontend"},
			Created: 1755600000,
			HostID:  "srv-demo-001",
		},
		{
			ID:      "def456api02",
			Name:    "api-server",
			Image:   "aicenter/api:1.4.2",
			ImageID: "sha256:api142",
			Command: "node server.js",
			State:   "running",
			Status:  "Up 2 days",
			Ports:   []Port{{IP: "0.0.0.0", PrivatePort: 3000, PublicPort: 3000, Type: "tcp"}},
			Labels:  map[string]string{"app": "api"},
			Created: 1755400000,
			HostID:  "srv-demo-001",
		},
		{
			ID:      "ghi789db03",
			Name:    "mysql-db",
			Image:   "mysql:8.0",
			ImageID: "sha256:mysql80",
			Command: "mysqld",
			State:   "exited",
			Status:  "Exited (0) 12 hours ago",
			Ports:   []Port{{IP: "127.0.0.1", PrivatePort: 3306, PublicPort: 3306, Type: "tcp"}},
			Labels:  map[string]string{"app": "db"},
			Created: 1755000000,
			HostID:  "srv-demo-002",
		},
	}
	for _, s := range seed {
		s := s
		c.containers = append(c.containers, s)
		c.details[s.ID] = &ContainerDetail{
			Container: s,
			Env:      []string{"TZ=Asia/Shanghai", "ENV=production"},
			Mounts:   []Mount{{Source: "/data/" + s.Name, Destination: "/var/lib/data", Mode: "rw"}},
			Network:  map[string]string{"bridge": "aicenter_default"},
			Logs:     "[" + s.Name + "] mock container log line\n",
		}
	}
	return c
}

func (m *MockClient) ListContainers(ctx context.Context, all bool) ([]Container, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Container, 0, len(m.containers))
	for _, c := range m.containers {
		if !all && c.State != "running" {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (m *MockClient) GetContainer(ctx context.Context, id string) (*ContainerDetail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.details[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (m *MockClient) StartContainer(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.details[id]
	if !ok {
		return ErrNotFound
	}
	d.State = "running"
	d.Status = "Up 1 second"
	return nil
}

func (m *MockClient) StopContainer(ctx context.Context, id string, timeoutSec int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.details[id]
	if !ok {
		return ErrNotFound
	}
	d.State = "exited"
	d.Status = "Exited (0) 1 second ago"
	return nil
}

func (m *MockClient) DeleteContainer(ctx context.Context, id string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.details[id]; !ok {
		return ErrNotFound
	}
	delete(m.details, id)
	for i, c := range m.containers {
		if c.ID == id {
			m.containers = append(m.containers[:i], m.containers[i+1:]...)
			break
		}
	}
	return nil
}
