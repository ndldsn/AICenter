import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useHasPermission, useCanAccess } from './useHasPermission';
import { useAuthStore } from '@/stores/authStore';

// Reset Zustand store between tests so role changes don't bleed.
beforeEach(() => {
    useAuthStore.setState({
        accessToken: null,
        refreshToken: null,
        user: null,
        isAuthenticated: false,
    });
});

describe('useHasPermission', () => {
    it('returns false when no user is logged in', () => {
        const { result } = renderHook(() => useHasPermission());
        expect(result.current).toBe(false);
    });

    it('returns true when user role is superadmin', () => {
        useAuthStore.setState({
            user: { id: '1', username: 'admin', email: 'admin@test', role: 'superadmin' },
            isAuthenticated: true,
        });
        const { result } = renderHook(() => useHasPermission());
        expect(result.current).toBe(true);
    });

    it('returns true when user role is admin', () => {
        useAuthStore.setState({
            user: { id: '2', username: 'admin2', email: 'admin2@test', role: 'admin' },
            isAuthenticated: true,
        });
        const { result } = renderHook(() => useHasPermission());
        expect(result.current).toBe(true);
    });

    it('returns false for other roles (no full catalog on client side)', () => {
        useAuthStore.setState({
            user: { id: '3', username: 'viewer', email: 'viewer@test', role: 'viewer' },
            isAuthenticated: true,
        });
        const { result } = renderHook(() => useHasPermission());
        expect(result.current).toBe(false);
    });
});

describe('useCanAccess', () => {
    it('grants access for admin regardless of requested permission string', () => {
        useAuthStore.setState({
            user: { id: '4', username: 'admin3', email: 'admin3@test', role: 'admin' },
            isAuthenticated: true,
        });
        const { result } = renderHook(() => useCanAccess('servers.delete' as any));
        expect(result.current).toBe(true);
    });

    it('denies access for non-admin roles', () => {
        useAuthStore.setState({
            user: { id: '5', username: 'user', email: 'user@test', role: 'viewer' },
            isAuthenticated: true,
        });
        const { result } = renderHook(() => useCanAccess('servers.delete' as any));
        expect(result.current).toBe(false);
    });

    it('denies access when given an array of permissions but user is non-admin', () => {
        useAuthStore.setState({
            user: { id: '6', username: 'user2', email: 'user2@test', role: 'viewer' },
            isAuthenticated: true,
        });
        const { result } = renderHook(() =>
            useCanAccess(['servers.read', 'servers.write'] as any)
        );
        expect(result.current).toBe(false);
    });
});

