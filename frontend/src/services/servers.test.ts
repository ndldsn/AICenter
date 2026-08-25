import { describe, it, expect, vi } from 'vitest';
import { serverApi } from './servers';
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

describe('servers.ts', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });

    it('serverApi.list calls apiGet with correct URL', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: { items: [], total: 0, page: 1, limit: 20 } });
        await serverApi.list(1, 20);
        expect(mockedApiGet).toHaveBeenCalledWith('/servers?page=1&limit=20');
    });

    it('serverApi.get calls apiGet with correct URL', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: { id: '1' } });
        await serverApi.get('1');
        expect(mockedApiGet).toHaveBeenCalledWith('/servers/1');
    });

    it('serverApi.create calls apiPost with correct data', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: { id: 'new' } });
        await serverApi.create({ name: 'srv1', host: '1.2.3.4', username: 'root', auth_type: 'password' });
        expect(mockedApiPost).toHaveBeenCalledWith('/servers', { name: 'srv1', host: '1.2.3.4', username: 'root', auth_type: 'password' });
    });

    it('serverApi.update calls apiPut with correct URL and data', async () => {
        mockedApiPut.mockResolvedValueOnce({ code: 0, data: { id: '1' } });
        await serverApi.update('1', { name: 'renamed' });
        expect(mockedApiPut).toHaveBeenCalledWith('/servers/1', { name: 'renamed' });
    });

    it('serverApi.remove calls apiDelete with correct URL', async () => {
        mockedApiDelete.mockResolvedValueOnce({ code: 0, data: null });
        await serverApi.remove('1');
        expect(mockedApiDelete).toHaveBeenCalledWith('/servers/1');
    });

    it('serverApi.testConnection calls apiPost with correct URL', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: { success: true } });
        await serverApi.testConnection('1', { timeout: 5 });
        expect(mockedApiPost).toHaveBeenCalledWith('/servers/1/connect', { timeout: 5 });
    });

    it('serverApi.getMetrics calls apiGet with correct URL', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: {} });
        await serverApi.getMetrics('1');
        expect(mockedApiGet).toHaveBeenCalledWith('/servers/1/metrics');
    });

    it('serverApi.generateAgentToken calls apiPost with correct URL', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: { token: 'tok' } });
        await serverApi.generateAgentToken('1');
        expect(mockedApiPost).toHaveBeenCalledWith('/servers/1/agent-token');
    });

    it('serverApi.listGroups calls apiGet', async () => {
        mockedApiGet.mockResolvedValueOnce({ code: 0, data: [] });
        await serverApi.listGroups();
        expect(mockedApiGet).toHaveBeenCalledWith('/server-groups');
    });

    it('serverApi.batchCommand calls apiPost with correct payload', async () => {
        mockedApiPost.mockResolvedValueOnce({ code: 0, data: { items: [], total: 0 } });
        await serverApi.batchCommand({ command: 'uptime', server_ids: ['1'], timeout_seconds: 5 });
        expect(mockedApiPost).toHaveBeenCalledWith('/servers/batch/command', {
            command: 'uptime',
            server_ids: ['1'],
            timeout_seconds: 5,
            group_id: undefined,
        });
    });
});

