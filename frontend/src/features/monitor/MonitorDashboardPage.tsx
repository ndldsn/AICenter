import { Typography, Card, Space } from '@arco-design/web-react';

const { Title, Paragraph } = Typography;

export default function MonitorDashboardPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>Monitor</Title>
                <Paragraph type="secondary">
                    Real-time monitoring and alerting
                </Paragraph>
            </div>

            <Card>
                <Paragraph>No monitoring data available.</Paragraph>
                <Paragraph type="secondary">
                    Connect servers to start collecting metrics and monitoring.
                </Paragraph>
            </Card>
        </Space>
    );
}
