import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { ROUTE_PERMISSIONS, type Permission } from '@/utils/permissions';
import { useCanAccess } from '@/hooks/useHasPermission';

interface ProtectedRouteProps {
  children: React.ReactNode;
  path: string;
}

export function ProtectedRoute({ children, path }: ProtectedRouteProps) {
  const { isAuthenticated, user } = useAuthStore();
  const required = ROUTE_PERMISSIONS[path];
  const can = useCanAccess((required || '') as Permission);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (required && !can) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '60vh',
          flexDirection: 'column',
          gap: 16,
        }}
      >
        <h2>403 — Forbidden</h2>
        <p style={{ color: 'var(--muted-foreground)' }}>
          Your role ({user?.role || 'unknown'}) does not have permission to
          access this page.
        </p>
        <button onClick={() => window.history.back()}>Go Back</button>
      </div>
    );
  }

  return <>{children}</>;
}
