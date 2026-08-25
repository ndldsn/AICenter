import { describe, it, expect, vi } from 'vitest';
import { dockerApi } from './docker';
import { apiGet, apiPost, apiPut, apiDelete } from './api';

vi.mock('./api', () => ({
    apiGet: vi.fn(),
    apiPost: vi.fn(),
    apiPut: vi.fn(),
    apiDelete: vi.fn(),
}));

const mockedApiGet = apiGet as ReturnType<typeof vi.fn>;
const mockedApiPost = apiPost as ReturnType<typeof vi.fn>;
const mockedApiPut = apiPut as ReturnType<typeof vi.fn>;
const mockedApiDelete = apiDelete as ReturnType<typeof vi.fn>;

describe('docker.ts', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });

    // ---- Hosts ----
    it('dockerApi.listHosts calls apiGet', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: { items: [] } });
        await dockerApi.listHosts();
        expect(mockedApiGet).toHaveBeenCalledWith('/docker/hosts');
    });

    // ---- Containers ----
    it('dockerApi.listContainers calls apiGet with all flag', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: { items: [] } });
        await dockerApi.listContainers(false);
        expect(mockedApiGet).toHaveBeenCalledWith('/docker/containers', { all: false });
    });

    it('dockerApi.getContainer calls apiGet', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: {} });
        await dockerApi.getContainer('c1');
        expect(mockedApiGet).toHaveBeenCalledWith('/docker/containers/c1');
    });

    it('dockerApi.getContainerLogs calls apiGet with tail', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: { logs: '' } });
        await dockerApi.getContainerLogs('c1', 500);
        expect(mockedApiGet).toHaveBeenCalledWith('/docker/containers/c1/logs', { tail: 500 });
    });

    it('dockerApi.startContainer calls apiPost', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.startContainer('c1');
        expect(mockedApiPost).toHaveBeenCalledWith('/docker/containers/c1/start');
    });

    it('dockerApi.stopContainer calls apiPost', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.stopContainer('c1');
        expect(mockedApiPost).toHaveBeenCalledWith('/docker/containers/c1/stop');
    });

    it('dockerApi.deleteContainer calls apiDelete with force flag', async () => {
        mockedApiDelete.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.deleteContainer('c1', true);
        expect(mockedApiDelete).toHaveBeenCalledWith('/docker/containers/c1?force=true');
    });

    // ---- Images ----
    it('dockerApi.listImages calls apiGet', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: { items: [] } });
        await dockerApi.listImages();
        expect(mockedApiGet).toHaveBeenCalledWith('/docker/images');
    });

    it('dockerApi.pullImage calls apiPost', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.pullImage('nginx', 'latest');
        expect(mockedApiPost).toHaveBeenCalledWith('/docker/images/pull', { repository: 'nginx', tag: 'latest' });
    });

    it('dockerApi.deleteImage calls apiDelete with force flag', async () => {
        mockedApiDelete.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.deleteImage('img1', true);
        expect(mockedApiDelete).toHaveBeenCalledWith('/docker/images/img1?force=true');
    });

    // ---- Volumes ----
    it('dockerApi.listVolumes calls apiGet', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: { items: [] } });
        await dockerApi.listVolumes();
        expect(mockedApiGet).toHaveBeenCalledWith('/docker/volumes');
    });

    it('dockerApi.createVolume calls apiPost', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: { name: 'v1' } });
        await dockerApi.createVolume('v1', 'local');
        expect(mockedApiPost).toHaveBeenCalledWith('/docker/volumes', { name: 'v1', driver: 'local' });
    });

    it('dockerApi.deleteVolume calls apiDelete with force=false', async () => {
        mockedApiDelete.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.deleteVolume('v1');
        expect(mockedApiDelete).toHaveBeenCalledWith('/docker/volumes/v1?force=false');
    });

    // ---- Networks ----
    it('dockerApi.listNetworks calls apiGet', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: { items: [] } });
        await dockerApi.listNetworks();
        expect(mockedApiGet).toHaveBeenCalledWith('/docker/networks');
    });

    it('dockerApi.createNetwork calls apiPost', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: { name: 'n1' } });
        await dockerApi.createNetwork('n1', 'bridge');
        expect(mockedApiPost).toHaveBeenCalledWith('/docker/networks', { name: 'n1', driver: 'bridge' });
    });

    it('dockerApi.deleteNetwork calls apiDelete', async () => {
        mockedApiDelete.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.deleteNetwork('n1');
        expect(mockedApiDelete).toHaveBeenCalledWith('/docker/networks/n1');
    });

    // ---- Compose ----
    it('dockerApi.listCompose calls apiGet', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: { items: [] } });
        await dockerApi.listCompose();
        expect(mockedApiGet).toHaveBeenCalledWith('/docker/compose');
    });

    it('dockerApi.getCompose calls apiGet', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: {} });
        await dockerApi.getCompose('p1');
        expect(mockedApiGet).toHaveBeenCalledWith('/docker/compose/p1');
    });

    it('dockerApi.createCompose calls apiPost', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: { name: 'p1' } });
        await dockerApi.createCompose('p1', 'version: 3');
        expect(mockedApiPost).toHaveBeenCalledWith('/docker/compose', { name: 'p1', content: 'version: 3' });
    });

    it('dockerApi.updateCompose calls apiPut', async () => {
        mockedApiPut.mockResolvedValueOnce({ code: 0, data: { name: 'p1' } });
        await dockerApi.updateCompose('p1', 'v2');
        expect(mockedApiPut).toHaveBeenCalledWith('/docker/compose/p1', { content: 'v2' });
    });

    it('dockerApi.deleteCompose calls apiDelete', async () => {
        mockedApiDelete.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.deleteCompose('p1');
        expect(mockedApiDelete).toHaveBeenCalledWith('/docker/compose/p1');
    });

    it('dockerApi.deployCompose calls apiPost', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.deployCompose('p1');
        expect(mockedApiPost).toHaveBeenCalledWith('/docker/compose/p1/deploy');
    });

    it('dockerApi.downCompose calls apiPost', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: null });
        await dockerApi.downCompose('p1');
        expect(mockedApiPost).toHaveBeenCalledWith('/docker/compose/p1/down');
    });
});
