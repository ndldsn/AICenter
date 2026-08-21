import { useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Menu, Button } from '@arco-design/web-react';
import {
    IconDashboard,
    IconDesktop,
    IconCloud,
    IconRobot,
    IconApps,
    IconCalendar,
    IconEye,
    IconFile,
    IconCheck,
    IconSettings,
} from '@arco-design/web-react/icon';
import { useUIStore } from '@/stores/uiStore';
import { useAuthStore } from '@/stores/authStore';
import type { RouteMeta } from '@/routes/routeConfig';

const iconMap: Record<string, React.ReactNode> = {
    dashboard: <IconDashboard />,
    server: <IconDesktop />,
    docker: <IconCloud />,
    robot: <IconRobot />,
    apps: <IconApps />,
    calendar: <IconCalendar />,
    eye: <IconEye />,
    file: <IconFile />,
    check: <IconCheck />,
    settings: <IconSettings />,
};

export function Sidebar() {
    const navigate = useNavigate();
    const location = useLocation();
    const { sidebarCollapsed } = useUIStore();
    const { user } = useAuthStore();
    const [selectedKeys, setSelectedKeys] = useState<string[]>([location.pathname]);

    const routeMeta: RouteMeta[] = [
        { path: '/', label: 'Dashboard', icon: 'dashboard' },
        { path: '/servers', label: 'Servers', icon: 'server' },
        { path: '/docker', label: 'Docker', icon: 'docker' },
        { path: '/models', label: 'AI Models', icon: 'apps' },
        { path: '/agents', label: 'Agents', icon: 'robot' },
        { path: '/tasks', label: 'Tasks', icon: 'calendar' },
        { path: '/monitor', label: 'Monitor', icon: 'eye' },
        { path: '/approvals', label: 'Approvals', icon: 'check' },
        { path: '/audit', label: 'Audit Log', icon: 'file' },
        { path: '/settings', label: 'Settings', icon: 'settings' },
    ];

    const handleMenuClick = (key: string) => {
        setSelectedKeys([key]);
        navigate(key);
    };

    return (
        <div
            style={{
                width: sidebarCollapsed ? 48 : 220,
                height: '100vh',
                backgroundColor: 'var(--color-bg-2)',
                borderRight: '1px solid var(--color-border)',
                display: 'flex',
                flexDirection: 'column',
                transition: 'width 0.2s',
                overflow: 'hidden',
            }}
        >
            {/* Logo */}
            <div
                style={{
                    height: 56,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    borderBottom: '1px solid var(--color-border)',
                    padding: '0 12px',
                }}
            >
                <span
                    style={{
                        fontSize: sidebarCollapsed ? 14 : 18,
                        fontWeight: 700,
                        color: 'var(--color-text-1)',
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                    }}
                >
                    {sidebarCollapsed ? 'AC' : 'AICenter'}
                </span>
            </div>

            {/* Menu */}
            <Menu
                style={{ flex: 1, overflow: 'auto' }}
                selectedKeys={selectedKeys}
                onClickMenuItem={handleMenuClick}
                collapsed={sidebarCollapsed}
                autoOpen
            >
                {routeMeta.map((item) => (
                    <Menu.Item key={item.path}>
                        <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                            {iconMap[item.icon]}
                            {!sidebarCollapsed && item.label}
                        </span>
                    </Menu.Item>
                ))}
            </Menu>

            {/* User section */}
            <div
                style={{
                    borderTop: '1px solid var(--color-border)',
                    padding: '12px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                }}
            >
                {!sidebarCollapsed && (
                    <div style={{ overflow: 'hidden' }}>
                        <div
                            style={{
                                fontSize: 13,
                                fontWeight: 500,
                                color: 'var(--color-text-1)',
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                                whiteSpace: 'nowrap',
                            }}
                        >
                            {user?.username || 'Admin'}
                        </div>
                        <div style={{ fontSize: 11, color: 'var(--color-text-3)' }}>
                            {user?.role || 'superadmin'}
                        </div>
                    </div>
                )}
                <Button
                    size="small"
                    icon={<IconSettings />}
                    onClick={() => navigate('/settings')}
                />
            </div>
        </div>
    );
}
