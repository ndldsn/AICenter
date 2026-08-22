import { apiGet, apiPost, apiPut, apiDelete } from './api';

// Server types
export interface Server {
    id: string;
    name: string;
    host: string;
    port: number;
    username: string;
    auth_type: 'password' | 'key';
    agent_connected: boolean;
    group_id?: string;
    tags: string[];
    os_info?: {
        distribution: string;
        kernel: string;
        architecture: string;
        hostname: string;
    };
    hardware_info?: {
        cpu_model: string;
        cpu_cores: number;
        memory_gb: number;
        disk_gb: number;
    };
    status: 'online' | 'offline' | 'unknown';
    last_heartbeat?: string;
    created_at: string;
    updated_at: string;
}

export interface ServerGroup {
    id: string;
    name: string;
    description: string;
    parent_id?: string;
    created_at: string;
}

export interface CreateServerRequest {
    name: string;
    host: string;
    port?: number;
    username: string;
    auth_type: 'password' | 'key';
    password?: string;
    private_key?: string;
    group_id?: string;
    tags?: string[];
}

export interface UpdateServerRequest {
    name?: string;
    host?: string;
    port?: number;
    username?: string;
    auth_type?: 'password' | 'key';
    password?: string;
    private_key?: string;
    group_id?: string;
    tags?: string[];
}

export interface ConnectionTestResult {
    success: boolean;
    message: string;
    ssh_banner?: string;
    system_info?: {
        os: string;
        os_version: string;
        hostname: string;
        kernel: string;
        cpu_cores: number;
        memory_gb: number;
        disk_usage_percent: number;
    };
    timestamp: string;
}

export interface PaginatedResponse<T> {
    items: T[];
    total: number;
    page: number;
    limit: number;
}

export interface ApiResponse<T> {
    code: number;
    message: string;
    data: T;
}

// Server API service
export const serverApi = {
    list: (page = 1, limit = 20) =>
        apiGet<ApiResponse<PaginatedResponse<Server>>>(
            `/servers?page=${page}&limit=${limit}`
        ),

    get: (id: string) =>
        apiGet<ApiResponse<Server>>(`/servers/${id}`),

    create: (data: CreateServerRequest) =>
        apiPost<ApiResponse<Server>>('/servers', data),

    update: (id: string, data: UpdateServerRequest) =>
        apiPut<ApiResponse<Server>>(`/servers/${id}`, data),

    remove: (id: string) =>
        apiDelete<ApiResponse<null>>(`/servers/${id}`),

    testConnection: (id: string, data?: any) =>
        apiPost<ApiResponse<ConnectionTestResult>>(
            `/servers/${id}/connect`,
            data
        ),

    getMetrics: (id: string) =>
        apiGet<ApiResponse<any>>(`/servers/${id}/metrics`),

    generateAgentToken: (id: string) =>
        apiPost<ApiResponse<{ token: string }>>(`/servers/${id}/agent-token`),

    listGroups: () =>
        apiGet<ApiResponse<ServerGroup[]>>('/server-groups'),

    createGroup: (data: { name: string; description?: string; parent_id?: string }) =>
        apiPost<ApiResponse<ServerGroup>>('/server-groups', data),

    deleteGroup: (id: string) =>
        apiDelete<ApiResponse<null>>(`/server-groups/${id}`),

    // Batch operations (Phase 7.2)
    batchCommand: (data: { command: string; server_ids?: string[]; timeout_seconds?: number; group_id?: string }) =>
        apiPost<ApiResponse<{ items: BatchResult[]; total: number }>>(`/servers/batch/command`, data),
};

export interface BatchResult {
    server_id: string;
    server: string;
    host: string;
    status: 'ok' | 'failed';
    stdout: string;
    stderr: string;
    error?: string;
    duration: string;
    exit_code: number;
}
