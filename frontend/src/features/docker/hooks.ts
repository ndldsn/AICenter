import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { dockerApi } from '@/services/docker';
import { Message } from '@arco-design/web-react';

// ---- Containers ----
export function useDockerHosts() {
    return useQuery({
        queryKey: ['docker', 'hosts'],
        queryFn: () => dockerApi.listHosts(),
        select: (res) => res.data?.items ?? [],
    });
}

export function useContainers(all = true) {
    return useQuery({
        queryKey: ['docker', 'containers', all],
        queryFn: () => dockerApi.listContainers(all),
        select: (res) => res.data?.items ?? [],
    });
}

export function useContainerLogs(id: string, enabled = false) {
    return useQuery({
        queryKey: ['docker', 'container-logs', id],
        queryFn: () => dockerApi.getContainerLogs(id),
        select: (res) => res.data?.logs ?? '',
        enabled,
    });
}

export function useStartContainer() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => dockerApi.startContainer(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'containers'] });
            Message.success('Container started');
        },
    });
}

export function useStopContainer() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => dockerApi.stopContainer(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'containers'] });
            Message.success('Container stopped');
        },
    });
}

export function useDeleteContainer() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => dockerApi.deleteContainer(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'containers'] });
            Message.success('Container deleted');
        },
    });
}

// ---- Images ----
export function useImages() {
    return useQuery({
        queryKey: ['docker', 'images'],
        queryFn: () => dockerApi.listImages(),
        select: (res) => res.data?.items ?? [],
    });
}

export function usePullImage() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ repository, tag }: { repository: string; tag: string }) =>
            dockerApi.pullImage(repository, tag),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'images'] });
            Message.success('Image pulled');
        },
    });
}

export function useDeleteImage() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => dockerApi.deleteImage(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'images'] });
            Message.success('Image deleted');
        },
    });
}

// ---- Volumes ----
export function useVolumes() {
    return useQuery({
        queryKey: ['docker', 'volumes'],
        queryFn: () => dockerApi.listVolumes(),
        select: (res) => res.data?.items ?? [],
    });
}

export function useCreateVolume() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ name, driver }: { name: string; driver: string }) =>
            dockerApi.createVolume(name, driver),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'volumes'] });
            Message.success('Volume created');
        },
    });
}

export function useDeleteVolume() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (name: string) => dockerApi.deleteVolume(name),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'volumes'] });
            Message.success('Volume deleted');
        },
    });
}

// ---- Networks ----
export function useNetworks() {
    return useQuery({
        queryKey: ['docker', 'networks'],
        queryFn: () => dockerApi.listNetworks(),
        select: (res) => res.data?.items ?? [],
    });
}

export function useCreateNetwork() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ name, driver }: { name: string; driver: string }) =>
            dockerApi.createNetwork(name, driver),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'networks'] });
            Message.success('Network created');
        },
    });
}

export function useDeleteNetwork() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => dockerApi.deleteNetwork(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'networks'] });
            Message.success('Network deleted');
        },
    });
}