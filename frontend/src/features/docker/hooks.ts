import { useEffect, useRef, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { dockerApi, DockerEvent } from '@/services/docker';
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
        // Live tail while the logs drawer is open.
        refetchInterval: enabled ? 3000 : false,
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

// ---- Compose ----
export function useComposeProjects() {
    return useQuery({
        queryKey: ['docker', 'compose'],
        queryFn: () => dockerApi.listCompose(),
        select: (res) => res.data?.items ?? [],
    });
}

export function useCreateCompose() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ name, content }: { name: string; content: string }) =>
            dockerApi.createCompose(name, content),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'compose'] });
            Message.success('Compose project created');
        },
    });
}

export function useUpdateCompose() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, content }: { id: string; content: string }) =>
            dockerApi.updateCompose(id, content),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'compose'] });
            Message.success('Compose project updated');
        },
    });
}

export function useDeleteCompose() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => dockerApi.deleteCompose(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'compose'] });
            Message.success('Compose project deleted');
        },
    });
}

export function useDeployCompose() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => dockerApi.deployCompose(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'compose'] });
            Message.success('Compose project deployed');
        },
    });
}

export function useDownCompose() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => dockerApi.downCompose(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['docker', 'compose'] });
            Message.success('Compose project stopped');
        },
    });
}

// ---- Real-time events (WebSocket) ----

const EVENT_LABELS: Record<string, string> = {
    'container.start': 'Container started',
    'container.stop': 'Container stopped',
    'container.delete': 'Container deleted',
    'image.pull': 'Image pulled',
    'image.delete': 'Image deleted',
    'volume.create': 'Volume created',
    'volume.delete': 'Volume deleted',
    'network.create': 'Network created',
    'network.delete': 'Network deleted',
    'compose.create': 'Compose project created',
    'compose.update': 'Compose project updated',
    'compose.delete': 'Compose project deleted',
    'compose.deploy': 'Compose project deployed',
    'compose.down': 'Compose project stopped',
};

/**
 * Subscribes to the WebSocket "docker" room. Every change event invalidates
 * the docker query cache (tables refresh automatically) and calls onEvent so
 * the UI can notify. Auto-reconnects with a 3s backoff.
 */
export function useDockerEvents(onEvent?: (e: DockerEvent) => void) {
    const queryClient = useQueryClient();
    const [connected, setConnected] = useState(false);
    const [lastEvent, setLastEvent] = useState<DockerEvent | null>(null);
    const onEventRef = useRef(onEvent);
    onEventRef.current = onEvent;

    useEffect(() => {
        let ws: WebSocket | null = null;
        let retryTimer: ReturnType<typeof setTimeout> | undefined;
        let disposed = false;

        const connect = () => {
            const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
            ws = new WebSocket(`${proto}://${window.location.host}/ws`);

            ws.onopen = () => {
                setConnected(true);
                ws?.send(
                    JSON.stringify({ type: 'subscribe', data: { channels: ['docker'] } })
                );
            };

            ws.onmessage = (ev) => {
                try {
                    const msg = JSON.parse(ev.data as string);
                    if (msg?.type !== 'docker.event' || !msg.data?.type) return;
                    const event: DockerEvent = {
                        type: msg.data.type,
                        target: msg.data.target ?? '',
                        host_id: msg.data.host_id ?? '',
                    };
                    setLastEvent(event);
                    queryClient.invalidateQueries({ queryKey: ['docker'] });
                    onEventRef.current?.(event);
                } catch {
                    // ignore malformed frames
                }
            };

            ws.onclose = () => {
                if (disposed) return;
                setConnected(false);
                retryTimer = setTimeout(connect, 3000);
            };

            ws.onerror = () => ws?.close();
        };

        connect();
        return () => {
            disposed = true;
            ws?.close();
            clearTimeout(retryTimer);
        };
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const label = lastEvent ? EVENT_LABELS[lastEvent.type] ?? lastEvent.type : null;
    return { connected, lastEvent, label };
}