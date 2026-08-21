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
