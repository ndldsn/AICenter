package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"


	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

var ErrRealUnavailable = errors.New("real docker client not implemented yet")

type RealClient struct {
	cli     *client.Client
	mu      sync.RWMutex
	compose map[string]*ComposeProject
}

func NewRealClient(host string) (*RealClient, error) {
	var opts []client.Opt
	if host != "" {
		opts = append(opts, client.WithHost(host))
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	rc := &RealClient{cli: cli, compose: make(map[string]*ComposeProject)}
	now := time.Now().Unix()
	rc.compose["compose-demo-stack"] = &ComposeProject{
		ID:       "compose-demo-stack",
		Name:     "demo-stack",
		HostID:   "srv-demo-001",
		Status:   "stopped",
		Content:  "services:\n  web:\n    image: nginx:1.27\n    ports:\n      - \"8080:80\"\n  api:\n    image: aicenter/api:1.4.2\n    environment:\n      - DB_HOST=db\n  db:\n    image: mysql:8.0\n",
		Services: []string{"api", "db", "web"},
		Created:  now - 86400,
		Updated:  now - 3600,
	}
	return rc, nil
}

func (r *RealClient) ensureCli() error {
	if r.cli == nil { return ErrRealUnavailable }
	return nil
}

// --- Containers ---

func (r *RealClient) ListContainers(ctx context.Context, all bool) ([]Container, error) {
	if err := r.ensureCli(); err != nil { return nil, err }
	l, err := r.cli.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil { return nil, fmt.Errorf("container list: %w", err) }
	out := make([]Container, 0, len(l))
	for _, c := range l { out = append(out, toContainer(c)) }
	return out, nil
}

func (r *RealClient) GetContainer(ctx context.Context, id string) (*ContainerDetail, error) {
	if err := r.ensureCli(); err != nil { return nil, err }
	inspect, err := r.cli.ContainerInspect(ctx, id)
	if err != nil { return nil, fmt.Errorf("container inspect: %w", err) }
	return toContainerDetail(inspect), nil
}

func (r *RealClient) StartContainer(ctx context.Context, id string) error {
	if err := r.ensureCli(); err != nil { return err }
	if err := r.cli.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return fmt.Errorf("container start: %w", err)
	}
	return nil
}

func (r *RealClient) StopContainer(ctx context.Context, id string, timeoutSec int) error {
	if err := r.ensureCli(); err != nil { return err }
	to := &timeoutSec
	if err := r.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: to}); err != nil {
		return fmt.Errorf("container stop: %w", err)
	}
	return nil
}

func (r *RealClient) DeleteContainer(ctx context.Context, id string, force bool) error {
	if err := r.ensureCli(); err != nil { return err }
	opts := container.RemoveOptions{RemoveVolumes: force}
	if err := r.cli.ContainerRemove(ctx, id, opts); err != nil {
		return fmt.Errorf("container remove: %w", err)
	}
	return nil
}

func (r *RealClient) ContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	if err := r.ensureCli(); err != nil { return "", err }
	opts := container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: strconv.Itoa(tail)}
	rc, err := r.cli.ContainerLogs(ctx, id, opts)
	if err != nil { return "", fmt.Errorf("container logs: %w", err) }
	defer rc.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return "", fmt.Errorf("read container logs: %w", err)
	}
	return buf.String(), nil
}

// --- Images ---

func (r *RealClient) ListImages(ctx context.Context) ([]Image, error) {
	if err := r.ensureCli(); err != nil { return nil, err }
	l, err := r.cli.ImageList(ctx, image.ListOptions{})
	if err != nil { return nil, fmt.Errorf("image list: %w", err) }
	out := make([]Image, 0, len(l))
	for _, img := range l { out = append(out, toImage(img)) }
	return out, nil
}

func (r *RealClient) PullImage(ctx context.Context, repository, tag string) error {
	if err := r.ensureCli(); err != nil { return err }
	if repository == "" { return errors.New("repository is required") }
	ref := repository
	if tag != "" { ref = ref + ":" + tag }
	rc, err := r.cli.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil { return fmt.Errorf("image pull %s: %w", ref, err) }
	rc.Close()
	return nil
}

func (r *RealClient) DeleteImage(ctx context.Context, id string, force bool) error {
	if err := r.ensureCli(); err != nil { return err }
	_, err := r.cli.ImageRemove(ctx, id, image.RemoveOptions{Force: force})
	if err != nil { return fmt.Errorf("image remove: %w", err) }
	return nil
}

// --- Volumes ---

func (r *RealClient) ListVolumes(ctx context.Context) ([]Volume, error) {
	if err := r.ensureCli(); err != nil { return nil, err }
	resp, err := r.cli.VolumeList(ctx, volume.ListOptions{Filters: filters.Args{}})
	if err != nil { return nil, fmt.Errorf("volume list: %w", err) }
	out := make([]Volume, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		if v != nil { out = append(out, *toVolumePtr(v)) }
	}
	return out, nil
}

func (r *RealClient) CreateVolume(ctx context.Context, name, driver string) (*Volume, error) {
	if err := r.ensureCli(); err != nil { return nil, err }
	if name == "" { return nil, errors.New("volume name is required") }
	if driver == "" { driver = "local" }
	opts := volume.CreateOptions{Driver: driver}
	v, err := r.cli.VolumeCreate(ctx, opts)
	if err != nil { return nil, fmt.Errorf("volume create: %w", err) }
	return toVolumePtr(&v), nil
}

func (r *RealClient) DeleteVolume(ctx context.Context, name string, force bool) error {
	if err := r.ensureCli(); err != nil { return err }
	if err := r.cli.VolumeRemove(ctx, name, force); err != nil {
		return fmt.Errorf("volume remove: %w", err)
	}
	return nil
}

// --- Networks ---

func (r *RealClient) ListNetworks(ctx context.Context) ([]Network, error) {
	if err := r.ensureCli(); err != nil { return nil, err }
	l, err := r.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil { return nil, fmt.Errorf("network list: %w", err) }
	out := make([]Network, 0, len(l))
	for _, n := range l { out = append(out, toNetwork(n)) }
	return out, nil
}

func (r *RealClient) CreateNetwork(ctx context.Context, name, driver string) (*Network, error) {
	if err := r.ensureCli(); err != nil { return nil, err }
	if name == "" { return nil, errors.New("network name is required") }
	if driver == "" { driver = "bridge" }
	opts := network.CreateOptions{Driver: driver}
	net, err := r.cli.NetworkCreate(ctx, name, opts)
	if err != nil { return nil, fmt.Errorf("network create: %w", err) }
	inspected, _ := r.cli.NetworkInspect(ctx, net.ID, network.InspectOptions{})
	return toNetworkPtr(&inspected), nil
}

func (r *RealClient) DeleteNetwork(ctx context.Context, id string) error {
	if err := r.ensureCli(); err != nil { return err }
	if err := r.cli.NetworkRemove(ctx, id); err != nil {
		return fmt.Errorf("network remove: %w", err)
	}
	return nil
}

// --- Compose (local-state; Docker SDK has no compose API) ---

func (r *RealClient) ListComposeProjects(ctx context.Context) ([]ComposeProject, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ComposeProject, 0, len(r.compose))
	for _, p := range r.compose { cp := *p; out = append(out, cp) }
	return out, nil
}

func (r *RealClient) GetComposeProject(ctx context.Context, id string) (*ComposeProject, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.compose[id]
	if !ok { return nil, ErrNotFound }
	cp := *p
	return &cp, nil
}

func (r *RealClient) CreateComposeProject(ctx context.Context, p ComposeProject) (*ComposeProject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p.Name == "" { return nil, errors.New("compose project name is required") }
	if strings.TrimSpace(p.Content) == "" { return nil, errors.New("compose file content is required") }
	svcs, err := parseComposeServices(p.Content)
	if err != nil { return nil, err }
	cp := &ComposeProject{
		ID: "compose-" + sanitize(p.Name),
		Name: p.Name, HostID: "srv-demo-001",
		Content: p.Content, Services: svcs,
		Status: "stopped", ProjectDir: "/opt/aicenter/compose/" + sanitize(p.Name),
		Created: time.Now().Unix(), Updated: time.Now().Unix(),
	}
	r.compose[cp.ID] = cp
	return cp, nil
}

func (r *RealClient) UpdateComposeProject(ctx context.Context, id string, p ComposeProject) (*ComposeProject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp, ok := r.compose[id]
	if !ok { return nil, ErrNotFound }
	if p.Name != "" { cp.Name = p.Name }
	if strings.TrimSpace(p.Content) != "" {
		svcs, err := parseComposeServices(p.Content)
		if err != nil { return nil, err }
		cp.Content, cp.Services, cp.Status = p.Content, svcs, "stopped"
	}
	cp.Updated = time.Now().Unix()
	out := *cp
	return &out, nil
}

func (r *RealClient) DeleteComposeProject(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.compose[id]; !ok { return ErrNotFound }
	delete(r.compose, id)
	return nil
}

func (r *RealClient) DeployComposeProject(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp, ok := r.compose[id]
	if !ok { return ErrNotFound }
	cp.Status = "running"
	cp.Updated = time.Now().Unix()
	return nil
}

func (r *RealClient) DownComposeProject(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp, ok := r.compose[id]
	if !ok { return ErrNotFound }
	cp.Status = "stopped"
	cp.Updated = time.Now().Unix()
	return nil
}

// --- Docker SDK to local model helpers ---

func toContainer(c container.Summary) Container {
	state := string(c.State)
	status := c.Status
	if state == "" { state = strings.ToLower(strings.SplitN(status, " ", 2)[0]) }
	name := ""
	if len(c.Names) > 0 { name = strings.TrimPrefix(c.Names[0], "/") }
	return Container{
		ID: c.ID, Name: name, Image: c.Image, ImageID: c.ImageID,
		Command: c.Command, State: state, Status: status,
		Ports: toPorts(c.Ports), Labels: c.Labels, Created: c.Created,
	}
}

func toPorts(ports []container.Port) []Port {
	out := make([]Port, 0, len(ports))
	for _, p := range ports {
		out = append(out, Port{IP: p.IP, PrivatePort: int(p.PrivatePort), PublicPort: int(p.PublicPort), Type: p.Type})
	}
	return out
}

func toContainerDetail(inspect container.InspectResponse) *ContainerDetail {
	base := inspect.ContainerJSONBase
	state := ""
	if base.State != nil { state = string(base.State.Status) }
	name := strings.TrimPrefix(base.Name, "/")
	created := int64(0)
	if base.Created != "" {
		t, err := time.Parse(time.RFC3339, base.Created)
		if err == nil { created = t.Unix() }
	}
	c := Container{
		ID: base.ID, Name: name, Image: base.Image,
		Command: base.Path, State: state, Status: state,
		Labels: func() map[string]string { if inspect.Config != nil { return inspect.Config.Labels }; return nil }(),
		Created: created,
	}
	env := make([]string, 0, len(inspect.Config.Env))
	env = append(env, inspect.Config.Env...)
	mounts := make([]Mount, 0, len(inspect.Mounts))
	for _, m := range inspect.Mounts {
		mounts = append(mounts, Mount{Source: m.Source, Destination: m.Destination, Mode: string(m.Type)})
	}
	networks := map[string]string{}
	if inspect.NetworkSettings != nil {
		for name, settings := range inspect.NetworkSettings.Networks {
			networks[name] = settings.IPAddress
		}
	}
	return &ContainerDetail{Container: c, Env: env, Mounts: mounts, Network: networks}
}

func toImage(img image.Summary) Image {
	repo, tag := "", ""
	if len(img.RepoTags) > 0 {
		parts := strings.SplitN(img.RepoTags[0], ":", 2)
		repo = parts[0]
		if len(parts) > 1 { tag = parts[1] }
	}
	return Image{ID: img.ID, Repository: repo, Tag: tag, Size: img.Size, Created: img.Created}
}

func toVolumePtr(v *volume.Volume) *Volume {
	if v == nil { return nil }
	return &Volume{Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint, Scope: v.Scope, Created: 0}
}

func toVolume(v volume.Volume) Volume { return *toVolumePtr(&v) }

func toNetwork(n network.Inspect) Network {
	return Network{
		ID: n.ID, Name: n.Name, Driver: n.Driver, Scope: n.Scope,
		Attachable: n.Attachable, Containers: len(n.Containers), Created: n.Created.Unix(),
	}
}

func toNetworkPtr(n *network.Inspect) *Network {
	if n == nil { return nil }
	nw := toNetwork(*n)
	return &nw
}
