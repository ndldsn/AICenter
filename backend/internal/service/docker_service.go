package service

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aicenter/aicenter/internal/docker"
)

// DockerEventRoom is the WebSocket room docker change events are broadcast to.
const DockerEventRoom = "docker"

// EventPublisher is the minimal interface DockerService needs to push change
// events to connected frontends. *websocket.Hub satisfies it without the
// service layer depending on the websocket package.
type EventPublisher interface {
	BroadcastToRoom(room string, message []byte)
}

// DockerEvent is the payload of a docker change notification.
type DockerEvent struct {
	Type   string `json:"type"`   // e.g. "container.start", "image.pull", "compose.deploy"
	Target string `json:"target"` // entity id or name
	HostID string `json:"host_id"`
}

// DockerService implements Phase 3 container management business logic.
type DockerService struct {
	client docker.Client
	bus    EventPublisher
}

// NewDockerService creates a DockerService. The client mode is driven by the
// DOCKER_MODE env var ("mock" by default) so the same code path serves both
// development (mock) and production (real daemon) without branching in handlers.
// bus may be nil; events are then dropped silently (unit tests, standalone use).
func NewDockerService(bus EventPublisher) *DockerService {
	mode := os.Getenv("DOCKER_MODE")
	if mode == "" {
		mode = "mock"
	}
	return &DockerService{client: docker.NewClient(docker.ClientConfig{Mode: mode}), bus: bus}
}

// emit publishes a docker change event to the WebSocket room (nil-safe).
func (s *DockerService) emit(eventType, target, hostID string) {
	if s.bus == nil {
		return
	}
	data, err := json.Marshal(DockerEvent{Type: eventType, Target: target, HostID: hostID})
	if err != nil {
		return
	}
	msg := map[string]interface{}{
		"type":      "docker.event",
		"channel":   DockerEventRoom,
		"timestamp": time.Now(),
		"data":      json.RawMessage(data),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.bus.BroadcastToRoom(DockerEventRoom, payload)
}

// ListContainers returns containers across all known hosts.
func (s *DockerService) ListContainers(ctx context.Context, all bool) ([]docker.Container, error) {
	return s.client.ListContainers(ctx, all)
}

// GetContainer returns full detail for a single container.
func (s *DockerService) GetContainer(ctx context.Context, id string) (*docker.ContainerDetail, error) {
	return s.client.GetContainer(ctx, id)
}

// StartContainer starts a stopped container.
func (s *DockerService) StartContainer(ctx context.Context, id string) error {
	if err := s.client.StartContainer(ctx, id); err != nil {
		return err
	}
	s.emit("container.start", id, "")
	return nil
}

// StopContainer stops a running container.
func (s *DockerService) StopContainer(ctx context.Context, id string, timeoutSec int) error {
	if err := s.client.StopContainer(ctx, id, timeoutSec); err != nil {
		return err
	}
	s.emit("container.stop", id, "")
	return nil
}

// DeleteContainer removes a container.
func (s *DockerService) DeleteContainer(ctx context.Context, id string, force bool) error {
	if err := s.client.DeleteContainer(ctx, id, force); err != nil {
		return err
	}
	s.emit("container.delete", id, "")
	return nil
}

// ContainerLogs returns the tail of a container's log output.
func (s *DockerService) ContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	return s.client.ContainerLogs(ctx, id, tail)
}

// ListImages returns all images on the managed hosts.
func (s *DockerService) ListImages(ctx context.Context) ([]docker.Image, error) {
	return s.client.ListImages(ctx)
}

// PullImage pulls an image by repository and tag.
func (s *DockerService) PullImage(ctx context.Context, repository, tag string) error {
	if err := s.client.PullImage(ctx, repository, tag); err != nil {
		return err
	}
	s.emit("image.pull", repository+":"+tag, "")
	return nil
}

// DeleteImage removes an image by id.
func (s *DockerService) DeleteImage(ctx context.Context, id string, force bool) error {
	if err := s.client.DeleteImage(ctx, id, force); err != nil {
		return err
	}
	s.emit("image.delete", id, "")
	return nil
}

// ListVolumes returns all named volumes.
func (s *DockerService) ListVolumes(ctx context.Context) ([]docker.Volume, error) {
	return s.client.ListVolumes(ctx)
}

// CreateVolume creates a named volume.
func (s *DockerService) CreateVolume(ctx context.Context, name, driver string) (*docker.Volume, error) {
	vol, err := s.client.CreateVolume(ctx, name, driver)
	if err != nil {
		return nil, err
	}
	s.emit("volume.create", name, vol.HostID)
	return vol, nil
}

// DeleteVolume removes a named volume.
func (s *DockerService) DeleteVolume(ctx context.Context, name string, force bool) error {
	if err := s.client.DeleteVolume(ctx, name, force); err != nil {
		return err
	}
	s.emit("volume.delete", name, "")
	return nil
}

// ListNetworks returns all networks.
func (s *DockerService) ListNetworks(ctx context.Context) ([]docker.Network, error) {
	return s.client.ListNetworks(ctx)
}

// CreateNetwork creates a network.
func (s *DockerService) CreateNetwork(ctx context.Context, name, driver string) (*docker.Network, error) {
	net, err := s.client.CreateNetwork(ctx, name, driver)
	if err != nil {
		return nil, err
	}
	s.emit("network.create", name, net.HostID)
	return net, nil
}

// DeleteNetwork removes a network by id.
func (s *DockerService) DeleteNetwork(ctx context.Context, id string) error {
	if err := s.client.DeleteNetwork(ctx, id); err != nil {
		return err
	}
	s.emit("network.delete", id, "")
	return nil
}

// ListComposeProjects returns all compose projects.
func (s *DockerService) ListComposeProjects(ctx context.Context) ([]docker.ComposeProject, error) {
	return s.client.ListComposeProjects(ctx)
}

// GetComposeProject returns a single compose project.
func (s *DockerService) GetComposeProject(ctx context.Context, id string) (*docker.ComposeProject, error) {
	return s.client.GetComposeProject(ctx, id)
}

// CreateComposeProject saves a new compose project.
func (s *DockerService) CreateComposeProject(ctx context.Context, p docker.ComposeProject) (*docker.ComposeProject, error) {
	cp, err := s.client.CreateComposeProject(ctx, p)
	if err != nil {
		return nil, err
	}
	s.emit("compose.create", cp.Name, cp.HostID)
	return cp, nil
}

// UpdateComposeProject saves a compose project. Compose YAML edits are treated
// as write operations: they are persisted but never auto-deployed.
func (s *DockerService) UpdateComposeProject(ctx context.Context, id string, p docker.ComposeProject) (*docker.ComposeProject, error) {
	cp, err := s.client.UpdateComposeProject(ctx, id, p)
	if err != nil {
		return nil, err
	}
	s.emit("compose.update", cp.Name, cp.HostID)
	return cp, nil
}

// DeleteComposeProject removes a compose project.
func (s *DockerService) DeleteComposeProject(ctx context.Context, id string) error {
	if err := s.client.DeleteComposeProject(ctx, id); err != nil {
		return err
	}
	s.emit("compose.delete", id, "")
	return nil
}

// DeployComposeProject runs `docker compose up -d` for a project.
func (s *DockerService) DeployComposeProject(ctx context.Context, id string) error {
	if err := s.client.DeployComposeProject(ctx, id); err != nil {
		return err
	}
	s.emit("compose.deploy", id, "")
	return nil
}

// DownComposeProject runs `docker compose down` for a project.
func (s *DockerService) DownComposeProject(ctx context.Context, id string) error {
	if err := s.client.DownComposeProject(ctx, id); err != nil {
		return err
	}
	s.emit("compose.down", id, "")
	return nil
}

// ListHosts returns the Docker hosts registered in the system. For Phase 3 core
// this is backed by mock data; it maps to managed servers that expose a daemon.
func (s *DockerService) ListHosts(ctx context.Context) ([]docker.Host, error) {
	return []docker.Host{
		{ID: "srv-demo-001", Name: "Demo Web Server", Endpoint: "ssh://ubuntu@192.168.1.10:22", Status: "online"},
		{ID: "srv-demo-002", Name: "Demo DB Server", Endpoint: "ssh://root@192.168.1.11:22", Status: "offline"},
	}, nil
}
