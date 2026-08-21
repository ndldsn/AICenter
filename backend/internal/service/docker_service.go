package service

import (
	"context"
	"os"

	"github.com/aicenter/aicenter/internal/docker"
)

// DockerService implements Phase 3 container management business logic.
type DockerService struct {
	client docker.Client
}

// NewDockerService creates a DockerService. The client mode is driven by the
// DOCKER_MODE env var ("mock" by default) so the same code path serves both
// development (mock) and production (real daemon) without branching in handlers.
func NewDockerService() *DockerService {
	mode := os.Getenv("DOCKER_MODE")
	if mode == "" {
		mode = "mock"
	}
	return &DockerService{client: docker.NewClient(docker.ClientConfig{Mode: mode})}
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
	return s.client.StartContainer(ctx, id)
}

// StopContainer stops a running container.
func (s *DockerService) StopContainer(ctx context.Context, id string, timeoutSec int) error {
	return s.client.StopContainer(ctx, id, timeoutSec)
}

// DeleteContainer removes a container.
func (s *DockerService) DeleteContainer(ctx context.Context, id string, force bool) error {
	return s.client.DeleteContainer(ctx, id, force)
}

// ListHosts returns the Docker hosts registered in the system. For Phase 3 core
// this is backed by mock data; it maps to managed servers that expose a daemon.
func (s *DockerService) ListHosts(ctx context.Context) ([]docker.Host, error) {
	return []docker.Host{
		{ID: "srv-demo-001", Name: "Demo Web Server", Endpoint: "ssh://ubuntu@192.168.1.10:22", Status: "online"},
		{ID: "srv-demo-002", Name: "Demo DB Server", Endpoint: "ssh://root@192.168.1.11:22", Status: "offline"},
	}, nil
}
