// Permission constants synced with backend/internal/permission/definitions.go.
// Keep in sync when adding new routes/actions.
export const PERM = {
  SERVERS_READ: 'servers.read',
  SERVERS_WRITE: 'servers.write',
  SERVERS_DELETE: 'servers.delete',
  DOCKER_READ: 'docker.read',
  DOCKER_WRITE: 'docker.write',
  AI_PROVIDERS_READ: 'ai.providers.read',
  AI_PROVIDERS_WRITE: 'ai.providers.write',
  AI_MODELS_READ: 'ai.models.read',
  AI_CHAT: 'ai.chat',
  AGENTS_READ: 'agents.read',
  AGENTS_WRITE: 'agents.write',
  AGENTS_SESSIONS: 'agents.sessions',
  NOTIFICATIONS_MANAGE: 'notifications.manage',
  MONITOR_READ: 'monitor.read',
  MONITOR_WRITE: 'monitor.write',
  TERMINAL_CREATE: 'terminal.create',
  TERMINAL_LIST: 'terminal.list',
  USERS_MANAGE: 'users.manage',
  ROLES_READ: 'roles.read',
  ROLES_MANAGE: 'roles.manage',
  APPROVALS_MANAGE: 'approvals.manage',
  TASKS_MANAGE: 'tasks.manage',
  AUDIT_READ: 'audit.read',
  SETTINGS_READ: 'settings.read',
  SETTINGS_WRITE: 'settings.write',
} as const;

export type Permission = (typeof PERM)[keyof typeof PERM];

// Route-to-permission mapping.
export const ROUTE_PERMISSIONS: Record<string, Permission | Permission[]> = {
  '/': PERM.SERVERS_READ, // dashboard needs at least view access
  '/servers': PERM.SERVERS_READ,
  '/servers/batch': PERM.SERVERS_WRITE,
  '/docker': PERM.DOCKER_READ,
  '/models': PERM.AI_MODELS_READ,
  '/agents': PERM.AGENTS_READ,
  '/tasks': PERM.TASKS_MANAGE,
  '/monitor': PERM.MONITOR_READ,
  '/notifications': PERM.NOTIFICATIONS_MANAGE,
  '/approvals': PERM.APPROVALS_MANAGE,
  '/audit': PERM.AUDIT_READ,
  '/settings': PERM.SETTINGS_READ,
};

// Sensitive actions that require a second confirmation dialog.
export const SENSITIVE_ACTIONS = [
  '/servers', // delete/restart server
  '/servers/batch', // batch commands
  '/users', // delete/update user
  '/roles', // delete/update role
  '/docker', // prune/remove containers
  '/settings', // update settings
] as const;
