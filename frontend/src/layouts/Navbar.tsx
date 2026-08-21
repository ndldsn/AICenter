import { useNavigate } from 'react-router-dom';
import { Button, Avatar, Dropdown, Menu, Badge, Message } from '@arco-design/web-react';
import {
    IconSun,
    IconMoon,
    IconNotification,
    IconUser,
    IconPoweroff,
    IconSettings,
} from '@arco-design/web-react/icon';
import { useUIStore } from '@/stores/uiStore';
import { useAuthStore } from '@/stores/authStore';

export function Navbar() {
    const navigate = useNavigate();
    const { theme, toggleTheme } = useUIStore();
    const { user, logout } = useAuthStore();

    const handleLogout = () => {
        logout();
        Message.success('Logged out successfully');
        navigate('/login');
    };

    const dropList = (
        <Menu>
            <Menu.Item key="profile" onClick={() => navigate('/settings')}>
                <IconUser /> Profile
            </Menu.Item>
            <Menu.Item key="settings" onClick={() => navigate('/settings')}>
                <IconSettings /> Settings
            </Menu.Item>
            <Menu.Item key="logout" onClick={handleLogout}>
            <IconPoweroff /> Logout
            </Menu.Item>
        </Menu>
    );

    return (
        <div
            style={{
                height: 56,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '0 20px',
                backgroundColor: 'var(--color-bg-2)',
                borderBottom: '1px solid var(--color-border)',
            }}
        >
            <div style={{ fontSize: 16, fontWeight: 600, color: 'var(--color-text-1)' }}>
                AICenter Dashboard
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                {/* Theme toggle */}
                <Button
                    icon={theme === 'light' ? <IconMoon /> : <IconSun />}
                    shape="circle"
                    onClick={toggleTheme}
                />

                {/* Notifications */}
                <Badge count={0} dot>
                    <Button icon={<IconNotification />} shape="circle" />
                </Badge>

                {/* User menu */}
                <Dropdown droplist={dropList} position="br">
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 8,
                            cursor: 'pointer',
                            padding: '4px 8px',
                            borderRadius: 4,
                        }}
                    >
                        <Avatar size={32} style={{ backgroundColor: 'rgb(var(--primary-6))' }}>
                            {user?.username?.charAt(0).toUpperCase() || 'A'}
                        </Avatar>
                        <span style={{ fontSize: 14, color: 'var(--color-text-1)' }}>
                            {user?.username || 'Admin'}
                        </span>
                    </div>
                </Dropdown>
            </div>
        </div>
    );
}
