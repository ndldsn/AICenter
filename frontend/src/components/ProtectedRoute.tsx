import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { ROUTE_PERMISSIONS, type Permission } from '@/utils/permissions';
import { useCanAccess } from '@/hooks/useHasPermission';
import { useT } from '@/stores/uiStore';

interface ProtectedRouteProps {
  children: React.ReactNode;
  path: string;
}

export function ProtectedRoute({ children, path }: ProtectedRouteProps) {
  const { isAuthenticated } = useAuthStore();
  const required = ROUTE_PERMISSIONS[path];
  const can = useCanAccess((required || '') as Permission);
  const t = useT();

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
        <h2>{t('auth.403')}</h2>
        <p style={{ color: 'var(--muted-foreground)' }}>
          {t('auth.403Hint')}
        </p>
        <button onClick={() => window.history.back()}>{t('auth.back')}</button>
      </div>
    );
  }

  return <>{children}</>;
}
