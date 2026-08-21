import { Typography, Card, Grid, Space, Tag } from '@arco-design/web-react';
import { IconDesktop, IconCloud, IconRobot, IconApps } from '@arco-design/web-react/icon';

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
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>Dashboard</Title>
                <Paragraph type="secondary">
                    Welcome to AICenter - Your AI-powered operations control center
                </Paragraph>
            </div>

            {/* Stats */}
            <Row gutter={16}>
                <Col span={6}>
                    <StatCard
                        title="Servers"
                        value={0}
                        icon={<IconDesktop />}
                        color="rgb(var(--primary-6))"
                    />
                </Col>
                <Col span={6}>
                    <StatCard
                        title="Containers"
                        value={0}
                        icon={<IconCloud />}
                        color="rgb(var(--success-6))"
                    />
                </Col>
                <Col span={6}>
                    <StatCard
                        title="AI Agents"
                        value={0}
                        icon={<IconRobot />}
                        color="rgb(var(--warning-6))"
                    />
                </Col>
                <Col span={6}>
                    <StatCard
                        title="AI Models"
                        value={0}
                        icon={<IconApps />}
                        color="rgb(var(--orange-6))"
                    />
                </Col>
            </Row>

            {/* Quick Actions */}
            <Card title="Quick Actions">
                <Space>
                    <Tag color="arcoblue" bordered>
                        Add Server
                    </Tag>
                    <Tag color="green" bordered>
                        Deploy Container
                    </Tag>
                    <Tag color="orange" bordered>
                        Create Agent
                    </Tag>
                    <Tag color="purple" bordered>
                        Add Provider
                    </Tag>
                </Space>
            </Card>

            {/* System Status */}
            <Card title="System Status">
                <Paragraph>
                    AICenter is running. No servers connected yet.
                </Paragraph>
                <Paragraph type="secondary">
                    Add your first server to start monitoring and managing your infrastructure.
                </Paragraph>
            </Card>
        </Space>
    );
}
