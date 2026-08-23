import { Typography, Card, Space } from '@arco-design/web-react';
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
                <Title heading={4}>设置</Title>
                <Paragraph type="secondary">
                    配置 AICenter 偏好设置
                </Paragraph>
                </div>

                <Card title="外观">
                <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                    <span>主题:</span>
                    <Button
                        icon={theme === 'light' ? <IconSun /> : <IconMoon />}
                        onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
                    >
                        {theme === 'light' ? '浅色模式' : '深色模式'}
                    </Button>
                </div>
                </Card>

                <Card title="账户">
                <div>用户名: {user?.username || 'N/A'}</div>
                <div>邮箱: {user?.email || 'N/A'}</div>
                <div>角色: {user?.role || 'N/A'}</div>
                </Card>

                <Card title="系统">
                <Paragraph type="secondary">
                    AICenter v1.0.0 - AI 运维控制中心
                </Paragraph>
                </Card>
        </Space>
    );
}
