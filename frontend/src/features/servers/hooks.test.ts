import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useServers, useServer, useCreateServer, useDeleteServer } from './hooks';
import { serverApi } from '@/services/servers';

vi.mock('@arco-design/web-react', () => ({
    Message: { success: vi.fn() },
}));

vi.mock('@/services/servers', () => ({
    serverApi: {
        list: vi.fn(),
        get: vi.fn(),
        create: vi.fn(),
        remove: vi.fn(),
        update: vi.fn(),
    },
}));

const mockedList = serverApi.list as ReturnType<typeof vi.fn>;
const mockedGet = serverApi.get as ReturnType<typeof vi.fn>;
const mockedCreate = serverApi.create as ReturnType<typeof vi.fn>;
const mockedRemove = serverApi.remove as ReturnType<typeof vi.fn>;

function wrapper({ children }: { children: React.ReactNode }) {
    const client = new QueryClient();
    return React.createElement(QueryClientProvider, { client, children });
}

describe('servers hooks', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });

    it('useServers calls serverApi.list and exposes data', async () => {
        mockedList.mockResolvedValueOnce({ code: 0, data: { items: [{ id: '1' }], total: 1, page: 1, limit: 20 } });
        const { result } = renderHook(() => useServers(1, 20), { wrapper });

        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(mockedList).toHaveBeenCalledWith(1, 20);
        expect(result.current.data).toEqual({ items: [{ id: '1' }], total: 1, page: 1, limit: 20 });
    });

    it('useServer calls serverApi.get when id is present', async () => {
        mockedGet.mockResolvedValueOnce({ code: 0, data: { id: '1', name: 's1' } });
        const { result } = renderHook(() => useServer('1'), { wrapper });

        await waitFor(() => expect(result.current.isSuccess).toBe(true));
        expect(mockedGet).toHaveBeenCalledWith('1');
        expect(result.current.data).toEqual({ id: '1', name: 's1' });
    });

    it('useServer does not call serverApi.get when id is empty', () => {
        const { result } = renderHook(() => useServer(''), { wrapper });
        // isFetching stays false because the query is disabled.
        expect(result.current.isFetching).toBe(false);
    });

    it('useDeleteServer calls serverApi.remove on mutate', async () => {
        mockedRemove.mockResolvedValueOnce({ code: 0, data: null });
        const { result } = renderHook(() => useDeleteServer(), { wrapper });

        await result.current.mutateAsync('1');
        expect(mockedRemove).toHaveBeenCalledWith('1');
    });
});
