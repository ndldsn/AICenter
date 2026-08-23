import { Typography, Card, Grid, Space, Button } from '@arco-design/web-react';
import { IconDesktop, IconCloud, IconRobot, IconApps } from '@arco-design/web-react/icon';
import { useNavigate } from 'react-router-dom';
import { useUIStore } from '@/stores/uiStore';

const { Title, Paragraph } = Typography;
const { Row, Col } = Grid;

interface StatCardProps {
    title: string;
    value: string | number;
    icon: React.ReactNode;
    color: string;
}

function StatCard({ title, value, icon, color }: StatCardProps) {
    return (
        <Card>
            <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                <div
                    style={{
                        width: 48,
                        height: 48,
                        borderRadius: 8,
                        backgroundColor: color,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#fff',
                        fontSize: 24,
                    }}
                >
                    {icon}
                </div>
                <div>
                    <div style={{ fontSize: 14, color: 'var(--color-text-3)' }}>{title}</div>
                    <div style={{ fontSize: 24, fontWeight: 600, color: 'var(--color-text-1)' }}>
                        {value}
                    </div>
                </div>
            </div>
        </Card>
    );
}

export default function DashboardPage() {
    const navigate = useNavigate();
    const t = useUIStore((s) => s.t);
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>{t('dashboard.title')}</Title>
                <Paragraph type="secondary">
                    {t('dashboard.welcome')}
                </Paragraph>
            </div>

            {/* Stats */}
            <Row gutter={16}>
                <Col span={6}>
                    <StatCard
                        title={t('sidebar.servers')}
                        value={0}
                        icon={<IconDesktop />}
                        color="rgb(var(--primary-6))"
                    />
                </Col>
                <Col span={6}>
                    <StatCard
                        title={t('sidebar.docker')}
                        value={0}
                        icon={<IconCloud />}
                        color="rgb(var(--success-6))"
                    />
                </Col>
                <Col span={6}>
                    <StatCard
                        title={t('sidebar.agents')}
                        value={0}
                        icon={<IconRobot />}
                        color="rgb(var(--warning-6))"
                    />
                </Col>
                <Col span={6}>
                    <StatCard
                        title={t('sidebar.ai')}
                        value={0}
                        icon={<IconApps />}
                        color="rgb(var(--orange-6))"
                    />
                </Col>
            </Row>

            {/* Quick Actions */}
            <Card title={t('dashboard.quickActions')}>
                <Space>
                    <Button size="small" type="outline" onClick={() => navigate('/servers')}>
                        {t('dashboard.addServer')}
                    </Button>
                    <Button size="small" type="outline" onClick={() => navigate('/docker')}>
                        {t('dashboard.deployContainer')}
                    </Button>
                    <Button size="small" type="outline" onClick={() => navigate('/agents')}>
                        {t('dashboard.createAgent')}
                    </Button>
                    <Button size="small" type="outline" onClick={() => navigate('/ai')}>
                        {t('dashboard.addProvider')}
                    </Button>
                </Space>
            </Card>

            {/* System Status */}
            <Card title={t('dashboard.systemStatus')}>
                <Paragraph>
                    {t('dashboard.systemOnline')}
                </Paragraph>
                <Paragraph type="secondary">
                    {t('dashboard.systemHint')}
                </Paragraph>
            </Card>
        </Space>
    );
}
