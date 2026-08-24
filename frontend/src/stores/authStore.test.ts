import { describe, it, expect, beforeEach } from 'vitest';
import { useAuthStore } from './authStore';

beforeEach(() => {
    // Clear localStorage before each test to isolate state.
    localStorage.clear();
    useAuthStore.setState({
        accessToken: null,
        refreshToken: null,
        user: null,
        isAuthenticated: false,
    });
});

describe('useAuthStore', () => {
    it('starts unauthenticated', () => {
        const state = useAuthStore.getState();
        expect(state.isAuthenticated).toBe(false);
        expect(state.user).toBeNull();
        expect(state.accessToken).toBeNull();
        expect(state.refreshToken).toBeNull();
    });

    it('sets tokens and marks authenticated', () => {
        useAuthStore.getState().setTokens('access-123', 'refresh-456');
        const state = useAuthStore.getState();
        expect(state.accessToken).toBe('access-123');
        expect(state.refreshToken).toBe('refresh-456');
        expect(state.isAuthenticated).toBe(true);
    });

    it('persists tokens to localStorage on setTokens', () => {
        useAuthStore.getState().setTokens('tok', 'ref');
        expect(localStorage.getItem('access_token')).toBe('tok');
        expect(localStorage.getItem('refresh_token')).toBe('ref');
    });

    it('sets user info', () => {
        const user = { id: '1', username: 'admin', email: 'admin@test', role: 'admin' };
        useAuthStore.getState().setUser(user);
        expect(useAuthStore.getState().user).toEqual(user);
    });

    it('clears all state on logout', () => {
        useAuthStore.getState().setTokens('a', 'r');
        useAuthStore.getState().setUser({ id: '1', username: 'u', email: 'e', role: 'admin' });
        useAuthStore.getState().logout();

        const state = useAuthStore.getState();
        expect(state.isAuthenticated).toBe(false);
        expect(state.user).toBeNull();
        expect(state.accessToken).toBeNull();
        expect(state.refreshToken).toBeNull();
    });

    it('removes tokens from localStorage on logout', () => {
        useAuthStore.getState().setTokens('a', 'r');
        useAuthStore.getState().logout();
        expect(localStorage.getItem('access_token')).toBeNull();
        expect(localStorage.getItem('refresh_token')).toBeNull();
    });
});

