package docker

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a container does not exist.
var ErrNotFound = errors.New("container not found")

// Client abstracts a Docker daemon so the service layer is transport-agnostic.
// A mock implementation is used for development without a local daemon; a real
// implementation (docker SDK / SSH tunnel to a remote daemon) can be dropped in
// by implementing this interface and selecting it via ClientConfig.Mode.
type Client interface {
	ListContainers(ctx context.Context, all bool) ([]Container, error)
	GetContainer(ctx context.Context, id string) (*ContainerDetail, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string, timeoutSec int) error
	DeleteContainer(ctx context.Context, id string, force bool) error
	ContainerLogs(ctx context.Context, id string, tail int) (string, error)

	// Images
	ListImages(ctx context.Context) ([]Image, error)
	PullImage(ctx context.Context, repository, tag string) error
	DeleteImage(ctx context.Context, id string, force bool) error

	// Volumes
	ListVolumes(ctx context.Context) ([]Volume, error)
	CreateVolume(ctx context.Context, name, driver string) (*Volume, error)
	DeleteVolume(ctx context.Context, name string, force bool) error

	// Networks
	ListNetworks(ctx context.Context) ([]Network, error)
	CreateNetwork(ctx context.Context, name, driver string) (*Network, error)
	DeleteNetwork(ctx context.Context, id string) error

	// Compose
	ListComposeProjects(ctx context.Context) ([]ComposeProject, error)
	GetComposeProject(ctx context.Context, id string) (*ComposeProject, error)
	CreateComposeProject(ctx context.Context, p ComposeProject) (*ComposeProject, error)
	UpdateComposeProject(ctx context.Context, id string, p ComposeProject) (*ComposeProject, error)
	DeleteComposeProject(ctx context.Context, id string) error
	DeployComposeProject(ctx context.Context, id string) error
	DownComposeProject(ctx context.Context, id string) error
}

// ClientConfig selects and configures a Client implementation.
type ClientConfig struct {
	Mode string // "mock" (default) or "real"
	Host string // daemon endpoint, e.g. unix:///var/run/docker.sock or tcp://host:2375
}

// NewClient returns a Client based on the config. Unknown modes fall back to mock
// so the API always works during development.
func NewClient(cfg ClientConfig) Client {
	if cfg.Mode == "real" {
		return NewRealClient(cfg.Host)
	}
	return NewMockClient()
}
