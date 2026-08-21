import { Typography, Card, Space, Divider } from '@arco-design/web-react';
import { useUIStore } from '@/stores/uiStore';
import { useAuthStore } from '@/stores/authStore';
import { Button } from '@arco-design/web-react';
import { IconSun, IconMoon } from '@arco-design/web-react/icon';

const { Title, Paragraph } = Typography;

export default function SettingsPage() {
    const { theme, setTheme } = useUIStore();
    const { user } = useAuthStore();

    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>Settings</Title>
                <Paragraph type="secondary">
                    Configure your AICenter preferences
                </Paragraph>
            </div>

            <Card title="Appearance">
                <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                    <span>Theme:</span>
                    <Button
                        icon={theme === 'light' ? <IconSun /> : <IconMoon />}
                        onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
                    >
                        {theme === 'light' ? 'Light Mode' : 'Dark Mode'}
                    </Button>
                </div>
            </Card>

            <Card title="Account">
                <div>Username: {user?.username || 'N/A'}</div>
                <div>Email: {user?.email || 'N/A'}</div>
                <div>Role: {user?.role || 'N/A'}</div>
            </Card>

            <Card title="System">
                <Paragraph type="secondary">
                    AICenter v1.0.0 - AI-powered operations control center
                </Paragraph>
            </Card>
        </Space>
    );
}
