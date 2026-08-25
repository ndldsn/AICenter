import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { AddServerModal } from './AddServerModal';
import { useCreateServer, useTestNewConnection } from './hooks';
vi.mock('@arco-design/web-react', async () => {
    const actual = await vi.importActual('@arco-design/web-react') as any;
    return {
        ...actual,
        Message: { success: vi.fn(), warning: vi.fn(), info: vi.fn() },
    };
});

vi.mock('./hooks', () => ({
    useCreateServer: vi.fn(),
    useTestNewConnection: vi.fn(),
    useUpdateServer: vi.fn(),
    useServers: vi.fn(),
    useDeleteServer: vi.fn(),
    useTestConnection: vi.fn(),
}));

const mockMutateAsync = vi.fn();

describe('AddServerModal', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
        (useCreateServer as any).mockReturnValue({
            mutateAsync: mockMutateAsync.mockResolvedValue({ code: 0, data: { id: 'new' } }),
            isSuccess: false,
        });
        (useTestNewConnection as any).mockReturnValue({
            mutateAsync: vi.fn().mockResolvedValue({
                code: 0,
                data: { success: true, message: 'OK', timestamp: new Date().toISOString() },
            }),
        });
    });

    it('does not render when visible is false', () => {
        const { container } = render(<AddServerModal visible={false} onClose={() => {}} onSuccess={() => {}} />);
        expect(container.firstChild).toBeNull();
    });

    it('calls onClose when close button is clicked', () => {
        const onClose = vi.fn();
        render(<AddServerModal visible={true} onClose={onClose} onSuccess={() => {}} />);
        // Modal has a close icon button; locate by aria-label fallback or role.
        const closeBtn = screen.getByRole('button', { name: /close/i });
        fireEvent.click(closeBtn);
        expect(onClose).toHaveBeenCalled();
    });

    it('calls createMutation.mutateAsync on form submit', async () => {
        mockMutateAsync.mockResolvedValueOnce({ code: 0, data: { id: 'new' } });
        render(<AddServerModal visible={true} onClose={() => {}} onSuccess={() => {}} />);
        // Find the submit button inside the modal footer (Arco default is "确定").
        const submitBtn = screen.getByRole('button', { name: /确定/i });
        fireEvent.click(submitBtn);
        await waitFor(() => expect(mockMutateAsync).toHaveBeenCalled());
    });
});
