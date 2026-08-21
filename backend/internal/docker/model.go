package docker

// Container is a summary view of a Docker container.
type Container struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	ImageID string            `json:"image_id"`
	Command string            `json:"command"`
	State   string            `json:"state"` // running | exited | paused | created
	Status  string            `json:"status"`
	Ports   []Port            `json:"ports"`
	Labels  map[string]string `json:"labels,omitempty"`
	Created int64             `json:"created"` // unix seconds
	HostID  string            `json:"host_id"`
}

// Port describes a published container port.
type Port struct {
	IP          string `json:"ip"`
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port,omitempty"`
	Type        string `json:"type"` // tcp | udp
}

// ContainerDetail is the full view of a single container.
type ContainerDetail struct {
	Container
	Env     []string          `json:"env,omitempty"`
	Mounts  []Mount           `json:"mounts,omitempty"`
	Network map[string]string `json:"network,omitempty"`
	Logs    string            `json:"logs,omitempty"`
}

// Mount describes a bind/volume mount.
type Mount struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
}

// Image is a Docker image summary.
type Image struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Size       int64  `json:"size"`
	Created    int64  `json:"created"`
}

// Host represents a Docker host (managed server running a Docker daemon).
type Host struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"` // online | offline
}
