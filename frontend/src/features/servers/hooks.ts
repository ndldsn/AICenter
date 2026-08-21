import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { serverApi, Server, CreateServerRequest, ConnectionTestResult } from '@/services/servers';
import { Message } from '@arco-design/web-react';

// Hook for server list
export function useServers(page = 1, limit = 20) {
    return useQuery({
        queryKey: ['servers', page, limit],
        queryFn: () => serverApi.list(page, limit),
        select: (res) => res.data,
    });
}

// Hook for single server
export function useServer(id: string) {
    return useQuery({
        queryKey: ['server', id],
        queryFn: () => serverApi.get(id),
        select: (res) => res.data,
        enabled: !!id,
    });
}

// Hook for creating server
export function useCreateServer() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateServerRequest) => serverApi.create(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['servers'] });
            Message.success('Server created successfully');
        },
    });
}

// Hook for updating server
export function useUpdateServer() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: string; data: any }) => serverApi.update(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['servers'] });
            Message.success('Server updated successfully');
        },
    });
}

// Hook for deleting server
export function useDeleteServer() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => serverApi.remove(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['servers'] });
            Message.success('Server deleted successfully');
        },
    });
}

// Hook for testing connection
export function useTestConnection() {
    return useMutation({
        mutationFn: (id: string) => serverApi.testConnection(id),
        select: (res) => res.data,
    });
}

// Hook for testing new server connection (before saving)
export function useTestNewConnection() {
    return useMutation({
        mutationFn: (data: CreateServerRequest) =>
            serverApi.testConnection('test', data),
        select: (res) => res.data,
    });
}
