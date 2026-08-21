package docker

import (
	"context"
	"errors"
	"testing"
)

func TestMockClient_ListContainers(t *testing.T) {
	c := NewMockClient()
	ctx := context.Background()

	all, err := c.ListContainers(ctx, true)
	if err != nil {
		t.Fatalf("ListContainers(all=true): %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 seeded containers, got %d", len(all))
	}

	// all=false must only return running ones
	running, err := c.ListContainers(ctx, false)
	if err != nil {
		t.Fatalf("ListContainers(all=false): %v", err)
	}
	for _, cn := range running {
		if cn.State != "running" {
			t.Fatalf("exited container %s returned when all=false", cn.Name)
		}
	}
}

func TestMockClient_ContainerLifecycle(t *testing.T) {
	c := NewMockClient()
	ctx := context.Background()

	// Stop a running container
	if err := c.StopContainer(ctx, "abc123web01", 10); err != nil {
		t.Fatalf("StopContainer: %v", err)
	}
	d, err := c.GetContainer(ctx, "abc123web01")
	if err != nil {
		t.Fatalf("GetContainer after stop: %v", err)
	}
	if d.State != "exited" {
		t.Fatalf("expected state exited, got %s", d.State)
	}

	// Start it again
	if err := c.StartContainer(ctx, "abc123web01"); err != nil {
		t.Fatalf("StartContainer: %v", err)
	}
	d, err = c.GetContainer(ctx, "abc123web01")
	if err != nil {
		t.Fatalf("GetContainer after start: %v", err)
	}
	if d.State != "running" {
		t.Fatalf("expected state running, got %s", d.State)
	}

	// Delete removes it
	if err := c.DeleteContainer(ctx, "abc123web01", false); err != nil {
		t.Fatalf("DeleteContainer: %v", err)
	}
	if _, err := c.GetContainer(ctx, "abc123web01"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMockClient_ContainerLogs(t *testing.T) {
	c := NewMockClient()
	ctx := context.Background()

	logs, err := c.ContainerLogs(ctx, "abc123web01", 200)
	if err != nil {
		t.Fatalf("ContainerLogs: %v", err)
	}
	if logs == "" {
		t.Fatal("expected non-empty logs")
	}

	if _, err := c.ContainerLogs(ctx, "does-not-exist", 200); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMockClient_Images(t *testing.T) {
	c := NewMockClient()
	ctx := context.Background()

	imgs, err := c.ListImages(ctx)
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}
	if len(imgs) != 4 {
		t.Fatalf("expected 4 seeded images, got %d", len(imgs))
	}

	// Pull new image
	if err := c.PullImage(ctx, "postgres", "16"); err != nil {
		t.Fatalf("PullImage: %v", err)
	}
	imgs, _ = c.ListImages(ctx)
	if len(imgs) != 5 {
		t.Fatalf("expected 5 images after pull, got %d", len(imgs))
	}

	// Pulling same image again is idempotent
	if err := c.PullImage(ctx, "postgres", "16"); err != nil {
		t.Fatalf("PullImage (dup): %v", err)
	}
	imgs, _ = c.ListImages(ctx)
	if len(imgs) != 5 {
		t.Fatalf("expected idempotent pull to keep 5 images, got %d", len(imgs))
	}

	// Delete
	imgs, _ = c.ListImages(ctx)
	var pgID string
	for _, img := range imgs {
		if img.Repository == "postgres" && img.Tag == "16" {
			pgID = img.ID
		}
	}
	if pgID == "" {
		t.Fatal("pulled image not found in list")
	}
	if err := c.DeleteImage(ctx, pgID, false); err != nil {
		t.Fatalf("DeleteImage: %v", err)
	}
	if _, err := c.ListImages(ctx); err != nil {
		t.Fatalf("ListImages after delete: %v", err)
	}
}

func TestMockClient_Volumes(t *testing.T) {
	c := NewMockClient()
	ctx := context.Background()

	vols, err := c.ListVolumes(ctx)
	if err != nil {
		t.Fatalf("ListVolumes: %v", err)
	}
	if len(vols) != 3 {
		t.Fatalf("expected 3 seeded volumes, got %d", len(vols))
	}

	// Create
	vol, err := c.CreateVolume(ctx, "backup_data", "local")
	if err != nil {
		t.Fatalf("CreateVolume: %v", err)
	}
	if vol.Name != "backup_data" || vol.Driver != "local" {
		t.Fatalf("unexpected volume: %+v", vol)
	}

	// Duplicate name rejected
	if _, err := c.CreateVolume(ctx, "backup_data", ""); err == nil {
		t.Fatal("expected error creating duplicate volume")
	}

	// Delete + not found
	if err := c.DeleteVolume(ctx, "backup_data", false); err != nil {
		t.Fatalf("DeleteVolume: %v", err)
	}
	if err := c.DeleteVolume(ctx, "backup_data", false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMockClient_Networks(t *testing.T) {
	c := NewMockClient()
	ctx := context.Background()

	nets, err := c.ListNetworks(ctx)
	if err != nil {
		t.Fatalf("ListNetworks: %v", err)
	}
	if len(nets) != 3 {
		t.Fatalf("expected 3 seeded networks, got %d", len(nets))
	}

	net, err := c.CreateNetwork(ctx, "monitor-net", "bridge")
	if err != nil {
		t.Fatalf("CreateNetwork: %v", err)
	}
	if net.Name != "monitor-net" || net.Driver != "bridge" {
		t.Fatalf("unexpected network: %+v", net)
	}

	// Deleting a network with connected containers is refused
	if err := c.DeleteNetwork(ctx, "net-bridge-1"); err == nil {
		t.Fatal("expected error deleting network with containers")
	}

	if err := c.DeleteNetwork(ctx, net.ID); err != nil {
		t.Fatalf("DeleteNetwork: %v", err)
	}
}
