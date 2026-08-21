package docker

import (
	"context"
	"errors"
)

// ErrRealUnavailable is returned by the real client stub until the Docker SDK
// integration is implemented. Selecting mode "real" before that work is done
// fails loudly instead of silently returning empty data.
var ErrRealUnavailable = errors.New("real docker client not implemented yet; set DOCKER_MODE=mock (or provide a docker daemon + SDK integration)")

// RealClient is the placeholder for connecting to an actual Docker daemon
// (via docker socket or SSH tunnel to a remote managed server). It is wired
// into the Client factory so swapping to real data is a one-line change plus
// an SDK implementation here.
type RealClient struct {
	host string
}

// NewRealClient builds a real client stub for the given daemon endpoint.
func NewRealClient(host string) *RealClient {
	return &RealClient{host: host}
}

func (r *RealClient) ListContainers(ctx context.Context, all bool) ([]Container, error) {
	return nil, ErrRealUnavailable
}
func (r *RealClient) GetContainer(ctx context.Context, id string) (*ContainerDetail, error) {
	return nil, ErrRealUnavailable
}
func (r *RealClient) StartContainer(ctx context.Context, id string) error { return ErrRealUnavailable }
func (r *RealClient) StopContainer(ctx context.Context, id string, timeoutSec int) error {
	return ErrRealUnavailable
}
func (r *RealClient) DeleteContainer(ctx context.Context, id string, force bool) error {
	return ErrRealUnavailable
}
func (r *RealClient) ContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	return "", ErrRealUnavailable
}
func (r *RealClient) ListImages(ctx context.Context) ([]Image, error) { return nil, ErrRealUnavailable }
func (r *RealClient) PullImage(ctx context.Context, repository, tag string) error {
	return ErrRealUnavailable
}
func (r *RealClient) DeleteImage(ctx context.Context, id string, force bool) error {
	return ErrRealUnavailable
}
func (r *RealClient) ListVolumes(ctx context.Context) ([]Volume, error) {
	return nil, ErrRealUnavailable
}
func (r *RealClient) CreateVolume(ctx context.Context, name, driver string) (*Volume, error) {
	return nil, ErrRealUnavailable
}
func (r *RealClient) DeleteVolume(ctx context.Context, name string, force bool) error {
	return ErrRealUnavailable
}
func (r *RealClient) ListNetworks(ctx context.Context) ([]Network, error) {
	return nil, ErrRealUnavailable
}
func (r *RealClient) CreateNetwork(ctx context.Context, name, driver string) (*Network, error) {
	return nil, ErrRealUnavailable
}
func (r *RealClient) DeleteNetwork(ctx context.Context, id string) error { return ErrRealUnavailable }
