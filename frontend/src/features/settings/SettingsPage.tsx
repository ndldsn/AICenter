import { Typography, Card, Space } from '@arco-design/web-react';
import { useT, useUIStore } from '@/stores/uiStore';
import { useAuthStore } from '@/stores/authStore';
import { Button } from '@arco-design/web-react';
import { IconSun, IconMoon } from '@arco-design/web-react/icon';

const { Title, Paragraph } = Typography;

export default function SettingsPage() {
    const t = useT();
    const { theme, setTheme } = useUIStore();
    const { user } = useAuthStore();

    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>{t('settings.title')}</Title>
                <Paragraph type="secondary">
                    {t('settings.subtitle')}
                </Paragraph>
                </div>

                <Card title={t('settings.appearance')}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                    <span>{t('settings.theme')}</span>
                    <Button
                        icon={theme === 'light' ? <IconSun /> : <IconMoon />}
                        onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
                    >
                        {theme === 'light' ? t('settings.lightMode') : t('settings.darkMode')}
                    </Button>
                </div>
                </Card>

                <Card title={t('settings.account')}>
                <div>{t('settings.username')} {user?.username || 'N/A'}</div>
                <div>{t('settings.email')} {user?.email || 'N/A'}</div>
                <div>{t('settings.role')} {user?.role || 'N/A'}</div>
                </Card>

                <Card title={t('settings.system')}>
                <Paragraph type="secondary">
                    {t('settings.version')}
                </Paragraph>
                </Card>
        </Space>
    );
}
