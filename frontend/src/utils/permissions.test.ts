import { describe, it, expect } from 'vitest';
import { PERM, ROUTE_PERMISSIONS, SENSITIVE_ACTIONS } from './permissions';

describe('PERM', () => {
    it('should contain all expected permission keys as non-empty strings', () => {
        const requiredKeys = [
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
        ];

        for (const key of requiredKeys) {
            expect(PERM).toHaveProperty(key);
            expect(typeof PERM[key as keyof typeof PERM]).toBe('string');
            expect(PERM[key as keyof typeof PERM].length).toBeGreaterThan(0);
        }
    });

    it('should have values in dot-notation format', () => {
        for (const [, value] of Object.entries(PERM)) {
            expect(value).toMatch(/^[a-z]+\.[a-z.]+$/);
        }
    });
});

describe('ROUTE_PERMISSIONS', () => {
    it('should map route paths to permission arrays', () => {
        expect(ROUTE_PERMISSIONS).toBeDefined();
        expect(typeof ROUTE_PERMISSIONS).toBe('object');
    });

    it('should have permissions for key routes', () => {
        expect(ROUTE_PERMISSIONS['/']).toBeDefined();
        expect(ROUTE_PERMISSIONS['/servers']).toBeDefined();
        expect(ROUTE_PERMISSIONS['/agents']).toBeDefined();
        expect(ROUTE_PERMISSIONS['/settings']).toBeDefined();
    });
});

describe('SENSITIVE_ACTIONS', () => {
    it('should define sensitive action constants', () => {
        expect(SENSITIVE_ACTIONS).toBeDefined();
        expect(Array.isArray(SENSITIVE_ACTIONS)).toBe(true);
        expect(SENSITIVE_ACTIONS.length).toBeGreaterThan(0);
    });
});