import { apiGet, apiPost, apiDelete } from './api';

// Docker types (mirror of backend/internal/docker/model.go)
export interface DockerHost {
    id: string;
    name: string;
    endpoint: string;
    status: 'online' | 'offline';
}

export interface DockerPort {
    ip: string;
    private_port: number;
    public_port?: number;
    type: 'tcp' | 'udp';
}

export interface DockerContainer {
    id: string;
    name: string;
    image: string;
    image_id: string;
    command: string;
    state: 'running' | 'exited' | 'paused' | 'created';
    status: string;
    ports: DockerPort[];
    labels?: Record<string, string>;
    created: number;
    host_id: string;
}

export interface DockerContainerDetail extends DockerContainer {
    env?: string[];
    mounts?: Array<{ source: string; destination: string; mode: string }>;
    network?: Record<string, string>;
    logs?: string;
}

export interface DockerImage {
    id: string;
    repository: string;
    tag: string;
    size: number;
    created: number;
}

export interface DockerVolume {
    name: string;
    driver: string;
    mountpoint: string;
    labels?: Record<string, string>;
    scope: string;
    created: number;
    host_id: string;
}

export interface DockerNetwork {
    id: string;
    name: string;
    driver: string;
    scope: string;
    internal: boolean;
    attachable: boolean;
    ipv6: boolean;
    containers: number;
    labels?: Record<string, string>;
    created: number;
    host_id: string;
}

export interface ApiResponse<T> {
    code: number;
    message: string;
    data: T;
}

export interface PaginatedResponse<T> {
    items: T[];
    total: number;
}

export interface ContainerLogsData {
    id: string;
    logs: string;
}

// Docker API service
export const dockerApi = {
    listHosts: () => apiGet<ApiResponse<PaginatedResponse<DockerHost>>>('/docker/hosts'),
    listContainers: (all = true) =>
        apiGet<ApiResponse<PaginatedResponse<DockerContainer>>>('/docker/containers', { all }),

    getContainer: (id: string) =>
        apiGet<ApiResponse<DockerContainerDetail>>(`/docker/containers/${id}`),

    getContainerLogs: (id: string, tail = 200) =>
        apiGet<ApiResponse<ContainerLogsData>>(`/docker/containers/${id}/logs`, { tail }),

    startContainer: (id: string) =>
        apiPost<ApiResponse<null>>(`/docker/containers/${id}/start`),

    stopContainer: (id: string) =>
        apiPost<ApiResponse<null>>(`/docker/containers/${id}/stop`),

    deleteContainer: (id: string, force = false) =>
        apiDelete<ApiResponse<null>>(`/docker/containers/${id}?force=${force}`),

    listImages: () => apiGet<ApiResponse<PaginatedResponse<DockerImage>>>('/docker/images'),

    pullImage: (repository: string, tag: string) =>
        apiPost<ApiResponse<null>>('/docker/images/pull', { repository, tag }),

    deleteImage: (id: string, force = false) =>
        apiDelete<ApiResponse<null>>(`/docker/images/${id}?force=${force}`),

    listVolumes: () => apiGet<ApiResponse<PaginatedResponse<DockerVolume>>>('/docker/volumes'),

    createVolume: (name: string, driver = 'local') =>
        apiPost<ApiResponse<DockerVolume>>('/docker/volumes', { name, driver }),

    deleteVolume: (name: string, force = false) =>
        apiDelete<ApiResponse<null>>(`/docker/volumes/${name}?force=${force}`),

    listNetworks: () => apiGet<ApiResponse<PaginatedResponse<DockerNetwork>>>('/docker/networks'),

    createNetwork: (name: string, driver = 'bridge') =>
        apiPost<ApiResponse<DockerNetwork>>('/docker/networks', { name, driver }),

    deleteNetwork: (id: string) =>
        apiDelete<ApiResponse<null>>(`/docker/networks/${id}`),
};