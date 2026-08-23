import { Typography, Card, Grid, Space, Button } from '@arco-design/web-react';
import { IconDesktop, IconCloud, IconRobot, IconApps } from '@arco-design/web-react/icon';
import { useNavigate } from 'react-router-dom';

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
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>仪表盘</Title>
                <Paragraph type="secondary">
                    欢迎使用 AICenter - 你的 AI 运维控制中心
                </Paragraph>
            </div>

            {/* Stats */}
            <Row gutter={16}>
                <Col span={6}>
                    <StatCard
                        title="服务器"
                        value={0}
                        icon={<IconDesktop />}
                        color="rgb(var(--primary-6))"
                    />
                </Col>
                <Col span={6}>
                    <StatCard
                        title="容器"
                        value={0}
                        icon={<IconCloud />}
                        color="rgb(var(--success-6))"
                    />
                </Col>
                <Col span={6}>
                    <StatCard
                        title="智能体"
                        value={0}
                        icon={<IconRobot />}
                        color="rgb(var(--warning-6))"
                    />
                </Col>
                <Col span={6}>
                    <StatCard
                        title="AI 模型"
                        value={0}
                        icon={<IconApps />}
                        color="rgb(var(--orange-6))"
                    />
                </Col>
            </Row>

            {/* Quick Actions */}
            <Card title="快速操作">
                <Space>
                    <Button size="small" type="outline" onClick={() => navigate('/servers')}>
                        添加服务器
                    </Button>
                    <Button size="small" type="outline" onClick={() => navigate('/docker')}>
                        部署容器
                    </Button>
                    <Button size="small" type="outline" onClick={() => navigate('/agents')}>
                        创建智能体
                    </Button>
                    <Button size="small" type="outline" onClick={() => navigate('/ai')}>
                        添加 Provider
                    </Button>
                </Space>
            </Card>

            {/* System Status */}
            <Card title="系统状态">
                <Paragraph>
                    AICenter 运行中。尚未连接任何服务器。
                </Paragraph>
                <Paragraph type="secondary">
                    添加第一台服务器即可开始监控和管理基础设施。
                </Paragraph>
            </Card>
        </Space>
    );
}
