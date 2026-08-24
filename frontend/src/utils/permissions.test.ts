import { describe, it, expect } from 'vitest';
import { PERM, ROUTE_PERMISSIONS, SENSITIVE_ACTIONS } from './permissions';

describe('PERM', () => {
    it('should contain all expected permission keys as non-empty strings', () => {
        const expectedKeys = [
            'SERVERS_READ',
            'SERVERS_WRITE',
            'SERVERS_DELETE',
            'DOCKER_READ',
            'DOCKER_WRITE',
            'AI_PROVIDERS_READ',
            'AI_PROVIDERS_WRITE',
            'AI_MODELS_READ',
            'AI_CHAT',
            'AGENTS_READ',
            'AGENTS_WRITE',
            'AGENTS_SESSIONS',
            'NOTIFICATIONS_MANAGE',
            'MONITOR_READ',
            'MONITOR_WRITE',
            'TERMINAL_CREATE',
            'TERMINAL_LIST',
            'USERS_MANAGE',
            'ROLES_READ',
            'ROLES_MANAGE',
            'APPROVALS_MANAGE',
            'TASKS_MANAGE',
            'AUDIT_READ',
            'SETTINGS_READ',
            'SETTINGS_WRITE',
        ] as const;

        for (const key of expectedKeys) {
            expect(PERM).toHaveProperty(key);
            expect(typeof PERM[key as keyof typeof PERM]).toBe('string');
            expect((PERM as Record<string, string>)[key]).toBeTruthy();
        }
    });

    it('should derive Permission type from all values', () => {
        // Ensure every value is a valid dot-separated permission string.
        for (const value of Object.values(PERM)) {
            expect(typeof value).toBe('string');
            expect(value).toContain('.');
        }
    });
});

describe('ROUTE_PERMISSIONS', () => {
    it('should map all known routes to a permission or permission array', () => {
        const entries = Object.entries(ROUTE_PERMISSIONS);
        expect(entries.length).toBeGreaterThan(0);

        for (const [, perm] of entries) {
            if (Array.isArray(perm)) {
                expect(perm.length).toBeGreaterThan(0);
                for (const p of perm) {
                    expect(typeof p).toBe('string');
                    expect(p).toContain('.');
                }
            } else {
                expect(typeof perm).toBe('string');
                expect(perm).toContain('.');
            }
        }
    });

    it('should include at minimum the dashboard and servers routes', () => {
        expect(ROUTE_PERMISSIONS).toHaveProperty('/');
        expect(ROUTE_PERMISSIONS).toHaveProperty('/servers');
    });
});

describe('SENSITIVE_ACTIONS', () => {
    it('should be a non-empty array of strings', () => {
        expect(Array.isArray(SENSITIVE_ACTIONS)).toBe(true);
        expect(SENSITIVE_ACTIONS.length).toBeGreaterThan(0);
        for (const action of SENSITIVE_ACTIONS) {
            expect(typeof action).toBe('string');
        }
    });
});

