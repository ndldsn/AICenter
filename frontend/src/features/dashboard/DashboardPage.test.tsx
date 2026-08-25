import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import DashboardPage from './DashboardPage';
// Mock the uiStore so the page is deterministic.
vi.mock('@/stores/uiStore', () => ({
    useT: vi.fn(() => (key: string) => key),
}));

describe('DashboardPage', () => {
    it('renders the dashboard title and quick actions', () => {
        render(
            <BrowserRouter>
                <DashboardPage />
            </BrowserRouter>
        );
        expect(screen.getByText('dashboard.title')).toBeInTheDocument();
        expect(screen.getByText('dashboard.quickActions')).toBeInTheDocument();
        expect(screen.getByText('dashboard.systemStatus')).toBeInTheDocument();
    });
});

