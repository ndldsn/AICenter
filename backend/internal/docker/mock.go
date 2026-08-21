package docker

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"
)

// MockClient is an in-memory Docker client used when no local daemon is available.
// It keeps mutable state so start/stop/delete behave realistically for testing.
type MockClient struct {
	mu         sync.RWMutex
	containers []Container
	details    map[string]*ContainerDetail
	images     []Image
	volumes    []Volume
	networks   []Network
}

// NewMockClient builds a client pre-seeded with realistic container data.
func NewMockClient() *MockClient {
	c := &MockClient{
		details: make(map[string]*ContainerDetail),
		images: []Image{
			{ID: "sha256:nginx1", Repository: "nginx", Tag: "1.27", Size: 187_000_000, Created: 1755600000},
			{ID: "sha256:api142", Repository: "aicenter/api", Tag: "1.4.2", Size: 92_000_000, Created: 1755400000},
			{ID: "sha256:mysql80", Repository: "mysql", Tag: "8.0", Size: 565_000_000, Created: 1755000000},
			{ID: "sha256:redis7", Repository: "redis", Tag: "7.2", Size: 118_000_000, Created: 1754800000},
		},
		volumes: []Volume{
			{Name: "aicenter_pgdata", Driver: "local", Mountpoint: "/var/lib/docker/volumes/aicenter_pgdata/_data", Scope: "local", Created: 1755000000, HostID: "srv-demo-001"},
			{Name: "aicenter_logs", Driver: "local", Mountpoint: "/var/lib/docker/volumes/aicenter_logs/_data", Scope: "local", Created: 1754900000, HostID: "srv-demo-001"},
			{Name: "mysql_data", Driver: "local", Mountpoint: "/var/lib/docker/volumes/mysql_data/_data", Scope: "local", Created: 1754800000, HostID: "srv-demo-002"},
		},
		networks: []Network{
			{ID: "net-bridge-1", Name: "aicenter_default", Driver: "bridge", Scope: "local", Attachable: true, Containers: 2, Created: 1755600000, HostID: "srv-demo-001"},
			{ID: "net-host-1", Name: "host", Driver: "host", Scope: "local", Created: 1755600000, HostID: "srv-demo-001"},
			{ID: "net-bridge-2", Name: "db-net", Driver: "bridge", Scope: "local", Containers: 1, Created: 1755000000, HostID: "srv-demo-002"},
		},
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
			Env:       []string{"TZ=Asia/Shanghai", "ENV=production"},
			Mounts:    []Mount{{Source: "/data/" + s.Name, Destination: "/var/lib/data", Mode: "rw"}},
			Network:   map[string]string{"bridge": "aicenter_default"},
			Logs:      "[" + s.Name + "] mock container log line\n",
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

func (m *MockClient) ContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.details[id]
	if !ok {
		return "", ErrNotFound
	}
	if tail <= 0 || tail > 200 {
		tail = 200
	}
	logs := d.Logs
	if len(logs) > tail*80 {
		logs = logs[len(logs)-tail*80:]
	}
	return logs, nil
}

func (m *MockClient) ListImages(ctx context.Context) ([]Image, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Image, len(m.images))
	copy(out, m.images)
	return out, nil
}

func (m *MockClient) PullImage(ctx context.Context, repository, tag string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if repository == "" {
		return errors.New("repository is required")
	}
	if tag == "" {
		tag = "latest"
	}
	for _, img := range m.images {
		if img.Repository == repository && img.Tag == tag {
			return nil // already present
		}
	}
	m.images = append(m.images, Image{
		ID:         "sha256:pulled-" + sanitize(repository) + "-" + tag,
		Repository: repository,
		Tag:        tag,
		Size:       0,
		Created:    time.Now().Unix(),
	})
	return nil
}

func (m *MockClient) DeleteImage(ctx context.Context, id string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, img := range m.images {
		if img.ID == id {
			m.images = append(m.images[:i], m.images[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockClient) ListVolumes(ctx context.Context) ([]Volume, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Volume, len(m.volumes))
	copy(out, m.volumes)
	return out, nil
}

func (m *MockClient) CreateVolume(ctx context.Context, name, driver string) (*Volume, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if name == "" {
		return nil, errors.New("volume name is required")
	}
	for _, v := range m.volumes {
		if v.Name == name {
			return nil, errors.New("volume already exists: " + name)
		}
	}
	if driver == "" {
		driver = "local"
	}
	vol := Volume{
		Name:       name,
		Driver:     driver,
		Mountpoint: "/var/lib/docker/volumes/" + name + "/_data",
		Scope:      "local",
		Created:    time.Now().Unix(),
		HostID:     "srv-demo-001",
	}
	m.volumes = append(m.volumes, vol)
	return &vol, nil
}

func (m *MockClient) DeleteVolume(ctx context.Context, name string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, v := range m.volumes {
		if v.Name == name {
			m.volumes = append(m.volumes[:i], m.volumes[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockClient) ListNetworks(ctx context.Context) ([]Network, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Network, len(m.networks))
	copy(out, m.networks)
	return out, nil
}

func (m *MockClient) CreateNetwork(ctx context.Context, name, driver string) (*Network, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if name == "" {
		return nil, errors.New("network name is required")
	}
	for _, n := range m.networks {
		if n.Name == name {
			return nil, errors.New("network already exists: " + name)
		}
	}
	if driver == "" {
		driver = "bridge"
	}
	net := Network{
		ID:      "net-" + sanitize(name),
		Name:    name,
		Driver:  driver,
		Scope:   "local",
		Created: time.Now().Unix(),
		HostID:  "srv-demo-001",
	}
	m.networks = append(m.networks, net)
	return &net, nil
}

func (m *MockClient) DeleteNetwork(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, n := range m.networks {
		if n.ID == id {
			if n.Containers > 0 {
				return errors.New("network has connected containers")
			}
			m.networks = append(m.networks[:i], m.networks[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// sanitize makes an arbitrary string safe to embed in a mock identifier.
func sanitize(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	return strings.ToLower(re.ReplaceAllString(s, "-"))
}
