import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
    useDockerHosts,
    useContainers,
    useContainerLogs,
    useStartContainer,
    useStopContainer,
    useDeleteContainer,
    useImages,
    usePullImage,
    useDeleteImage,
    useVolumes,
    useCreateVolume,
    useDeleteVolume,
    useNetworks,
    useCreateNetwork,
    useDeleteNetwork,
    useComposeProjects,
    useCreateCompose,
    useUpdateCompose,
    useDeleteCompose,
    useDeployCompose,
    useDownCompose,
} from './hooks';
import { dockerApi } from '@/services/docker';

vi.mock('@arco-design/web-react', () => ({
    Message: { success: vi.fn() },
}));

vi.mock('@/services/docker', () => ({
    dockerApi: {
        listHosts: vi.fn(),
        listContainers: vi.fn(),
        getContainerLogs: vi.fn(),
        startContainer: vi.fn(),
        stopContainer: vi.fn(),
        deleteContainer: vi.fn(),
        listImages: vi.fn(),
        pullImage: vi.fn(),
        deleteImage: vi.fn(),
        listVolumes: vi.fn(),
        createVolume: vi.fn(),
        deleteVolume: vi.fn(),
        listNetworks: vi.fn(),
        createNetwork: vi.fn(),
        deleteNetwork: vi.fn(),
        listCompose: vi.fn(),
        createCompose: vi.fn(),
        updateCompose: vi.fn(),
        deleteCompose: vi.fn(),
        deployCompose: vi.fn(),
        downCompose: vi.fn(),
    },
}));

function wrapper({ children }: { children: React.ReactNode }) {
    return React.createElement(QueryClientProvider, { client: new QueryClient(), children });
}

const mkRes = (data: any) => ({ code: 0, data });

describe('docker hooks', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });

    // ---- Containers ----
    it('useDockerHosts calls dockerApi.listHosts and selects items', async () => {
        (dockerApi.listHosts as any).mockResolvedValueOnce(mkRes({ items: [{ id: 'h1' }] }));
        const { result } = renderHook(() => useDockerHosts(), { wrapper });
        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(dockerApi.listHosts).toHaveBeenCalled();
        expect(result.current.data).toEqual([{ id: 'h1' }]);
    });

    it('useContainers calls dockerApi.listContainers and selects items', async () => {
        (dockerApi.listContainers as any).mockResolvedValueOnce(mkRes({ items: [] }));
        const { result } = renderHook(() => useContainers(), { wrapper });
        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(dockerApi.listContainers).toHaveBeenCalledWith(true);
    });

    it('useStartContainer mutate calls dockerApi.startContainer', async () => {
        (dockerApi.startContainer as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => useStartContainer(), { wrapper });
        await result.current.mutateAsync('c1');
        expect(dockerApi.startContainer).toHaveBeenCalledWith('c1');
    });

    it('useStopContainer mutate calls dockerApi.stopContainer', async () => {
        (dockerApi.stopContainer as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => useStopContainer(), { wrapper });
        await result.current.mutateAsync('c1');
        expect(dockerApi.stopContainer).toHaveBeenCalledWith('c1');
    });

    it('useDeleteContainer mutate calls dockerApi.deleteContainer', async () => {
        (dockerApi.deleteContainer as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => useDeleteContainer(), { wrapper });
        await result.current.mutateAsync('c1');
        expect(dockerApi.deleteContainer).toHaveBeenCalledWith('c1');
    });

    // ---- Images ----
    it('useImages calls dockerApi.listImages', async () => {
        (dockerApi.listImages as any).mockResolvedValueOnce(mkRes({ items: [] }));
        const { result } = renderHook(() => useImages(), { wrapper });
        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(dockerApi.listImages).toHaveBeenCalled();
    });

    it('usePullImage mutate calls dockerApi.pullImage', async () => {
        (dockerApi.pullImage as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => usePullImage(), { wrapper });
        await result.current.mutateAsync({ repository: 'nginx', tag: 'latest' });
        expect(dockerApi.pullImage).toHaveBeenCalledWith('nginx', 'latest');
    });

    it('useDeleteImage mutate calls dockerApi.deleteImage', async () => {
        (dockerApi.deleteImage as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => useDeleteImage(), { wrapper });
        await result.current.mutateAsync('img1');
        expect(dockerApi.deleteImage).toHaveBeenCalledWith('img1');
    });

    // ---- Volumes ----
    it('useVolumes calls dockerApi.listVolumes', async () => {
        (dockerApi.listVolumes as any).mockResolvedValueOnce(mkRes({ items: [] }));
        const { result } = renderHook(() => useVolumes(), { wrapper });
        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(dockerApi.listVolumes).toHaveBeenCalled();
    });

    it('useCreateVolume mutate calls dockerApi.createVolume', async () => {
        (dockerApi.createVolume as any).mockResolvedValueOnce(mkRes({ name: 'v1' }));
        const { result } = renderHook(() => useCreateVolume(), { wrapper });
        await result.current.mutateAsync({ name: 'v1', driver: 'local' });
        expect(dockerApi.createVolume).toHaveBeenCalledWith('v1', 'local');
    });

    it('useDeleteVolume mutate calls dockerApi.deleteVolume', async () => {
        (dockerApi.deleteVolume as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => useDeleteVolume(), { wrapper });
        await result.current.mutateAsync('v1');
        expect(dockerApi.deleteVolume).toHaveBeenCalledWith('v1');
    });

    // ---- Networks ----
    it('useNetworks calls dockerApi.listNetworks', async () => {
        (dockerApi.listNetworks as any).mockResolvedValueOnce(mkRes({ items: [] }));
        const { result } = renderHook(() => useNetworks(), { wrapper });
        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(dockerApi.listNetworks).toHaveBeenCalled();
    });

    it('useCreateNetwork mutate calls dockerApi.createNetwork', async () => {
        (dockerApi.createNetwork as any).mockResolvedValueOnce(mkRes({ name: 'n1' }));
        const { result } = renderHook(() => useCreateNetwork(), { wrapper });
        await result.current.mutateAsync({ name: 'n1', driver: 'bridge' });
        expect(dockerApi.createNetwork).toHaveBeenCalledWith('n1', 'bridge');
    });

    it('useDeleteNetwork mutate calls dockerApi.deleteNetwork', async () => {
        (dockerApi.deleteNetwork as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => useDeleteNetwork(), { wrapper });
        await result.current.mutateAsync('n1');
        expect(dockerApi.deleteNetwork).toHaveBeenCalledWith('n1');
    });

    // ---- Compose ----
    it('useComposeProjects calls dockerApi.listCompose', async () => {
        (dockerApi.listCompose as any).mockResolvedValueOnce(mkRes({ items: [] }));
        const { result } = renderHook(() => useComposeProjects(), { wrapper });
        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(dockerApi.listCompose).toHaveBeenCalled();
    });

    it('useCreateCompose mutate calls dockerApi.createCompose', async () => {
        (dockerApi.createCompose as any).mockResolvedValueOnce(mkRes({ name: 'p1' }));
        const { result } = renderHook(() => useCreateCompose(), { wrapper });
        await result.current.mutateAsync({ name: 'p1', content: 'version: 3' });
        expect(dockerApi.createCompose).toHaveBeenCalledWith('p1', 'version: 3');
    });

    it('useUpdateCompose mutate calls dockerApi.updateCompose', async () => {
        (dockerApi.updateCompose as any).mockResolvedValueOnce(mkRes({ name: 'p1' }));
        const { result } = renderHook(() => useUpdateCompose(), { wrapper });
        await result.current.mutateAsync({ id: 'p1', content: 'v2' });
        expect(dockerApi.updateCompose).toHaveBeenCalledWith('p1', 'v2');
    });

    it('useDeleteCompose mutate calls dockerApi.deleteCompose', async () => {
        (dockerApi.deleteCompose as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => useDeleteCompose(), { wrapper });
        await result.current.mutateAsync('p1');
        expect(dockerApi.deleteCompose).toHaveBeenCalledWith('p1');
    });

    it('useDeployCompose mutate calls dockerApi.deployCompose', async () => {
        (dockerApi.deployCompose as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => useDeployCompose(), { wrapper });
        await result.current.mutateAsync('p1');
        expect(dockerApi.deployCompose).toHaveBeenCalledWith('p1');
    });

    it('useDownCompose mutate calls dockerApi.downCompose', async () => {
        (dockerApi.downCompose as any).mockResolvedValueOnce(mkRes(null));
        const { result } = renderHook(() => useDownCompose(), { wrapper });
        await result.current.mutateAsync('p1');
        expect(dockerApi.downCompose).toHaveBeenCalledWith('p1');
    });
});

