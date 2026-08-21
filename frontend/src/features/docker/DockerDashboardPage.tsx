import { Typography, Card, Button, Space, Tag } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';

const { Title, Paragraph } = Typography;

export default function DockerDashboardPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>Docker</Title>
                    <Paragraph type="secondary">
                        Manage containers, images, volumes, and compose projects
                    </Paragraph>
                </div>
                <Button type="primary" icon={<IconPlus />}>
                    Deploy
                </Button>
            </div>

            <Card>
                <Paragraph>No Docker hosts configured yet.</Paragraph>
                <Paragraph type="secondary">
                    Add a server with Docker to start managing containers.
                </Paragraph>
            </Card>
        </Space>
    );
}
