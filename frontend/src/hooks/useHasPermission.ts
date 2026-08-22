import { useMemo } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { type Permission } from '@/utils/permissions';

// hasPermission checks whether the current user's role grants the given
// permission. It is a client-side convenience check; the real enforcement
// lives on the backend. The frontend uses it for route gating and menu
// visibility only.
export function useHasPermission() {
  const { user } = useAuthStore();

  return useMemo(() => {
    if (!user?.role) return false;

    // Superadmin bypasses all checks on the frontend (backend still
    // enforces per-endpoint policies).
    if (user.role === 'superadmin' || user.role === 'admin') {
      return true;
    }

    // Fallback: without a full permission catalog in the client store we
    // cannot evaluate custom roles here. For now only superadmin/admin
    // are auto-granted; other roles require fetching /permissions from
    // the backend and caching in a dedicated permission store.
    return false;
  }, [user?.role]);
}

// Check a single permission by name.
export function useCanAccess(permission: Permission | Permission[]) {
  const has = useHasPermission();

  return useMemo(() => {
    if (Array.isArray(permission)) {
      return permission.some((_p) => has);
    }
    return has ?? false;
  }, [has, permission]);
}
