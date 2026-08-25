import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import ServerListPage from './ServerListPage';
import { useServers, useDeleteServer } from './hooks';

vi.mock('@arco-design/web-react', async () => {
    const actual = await vi.importActual('@arco-design/web-react') as any;
    return {
        ...actual,
        Message: { success: vi.fn(), info: vi.fn() },
    };
});

vi.mock('@/stores/uiStore', () => ({
    useT: vi.fn(() => (key: string) => key),
}));

vi.mock('./hooks', () => ({
    useServers: vi.fn(),
    useDeleteServer: vi.fn(),
    useCreateServer: vi.fn(),
    useUpdateServer: vi.fn(),
    useTestConnection: vi.fn(),
    useTestNewConnection: vi.fn(),
}));

vi.mock('./AddServerModal', () => ({
    AddServerModal: vi.fn(({ visible }: any) =>
        visible ? <div data-testid="add-modal">Add Server Modal</div> : null
    ),
}));

const mockServers = {
    items: [
        {
            id: '1',
            name: 'web-01',
            host: '10.0.0.1',
            port: 22,
            status: 'online',
            agent_connected: true,
            os_info: { distribution: 'Ubuntu', architecture: 'amd64' },
            hardware_info: { cpu_cores: 4, memory_gb: 8, disk_gb: 100 },
            last_heartbeat: '2026-01-01T00:00:00Z',
        },
    ],
    total: 1,
    page: 1,
    limit: 20,
};

describe('ServerListPage', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
        (useServers as any).mockReturnValue({
            data: mockServers,
            isLoading: false,
            refetch: vi.fn(),
        });
        (useDeleteServer as any).mockReturnValue({
            mutateAsync: vi.fn().mockResolvedValue(undefined),
        });
    });

    it('renders the server list page title', () => {
        render(
            <BrowserRouter>
                <ServerListPage />
            </BrowserRouter>
        );
        expect(screen.getByText('servers.title')).toBeInTheDocument();
    });

    it('renders server rows from query data', () => {
        render(
            <BrowserRouter>
                <ServerListPage />
            </BrowserRouter>
        );
        expect(screen.getByText('web-01')).toBeInTheDocument();
        expect(screen.getByText('10.0.0.1')).toBeInTheDocument();
    });

    it('shows empty state when no servers', () => {
        (useServers as any).mockReturnValueOnce({
            data: { items: [], total: 0, page: 1, limit: 20 },
            isLoading: false,
            refetch: vi.fn(),
        });
        render(
            <BrowserRouter>
                <ServerListPage />
            </BrowserRouter>
        );
        expect(screen.getByText('servers.empty')).toBeInTheDocument();
    });

    it('opens AddServerModal when add button is clicked', () => {
        render(
            <BrowserRouter>
                <ServerListPage />
            </BrowserRouter>
        );
        // The "Add" button is the second header button (after refresh).
        const buttons = screen.getAllByRole('button');
        const addButton = buttons.find((b) => b.textContent?.includes('servers.add')) || buttons[1];
        fireEvent.click(addButton);
        expect(screen.getByTestId('add-modal')).toBeInTheDocument();
    });

    it('has a delete (danger) icon button in the actions column', () => {
        render(
            <BrowserRouter>
                <ServerListPage />
            </BrowserRouter>
        );
        // There are 4 icon-only action buttons (test/edit/copy/delete); the
        // last one is the delete button with status-danger class.
        const iconButtons = screen.getAllByRole('button', { name: '' });
        const deleteBtn = iconButtons.find((b) =>
            b.classList.contains('arco-btn-status-danger')
        );
        expect(deleteBtn).toBeTruthy();
    });
});
