import { Typography, Card, Button, Space } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';

const { Title, Paragraph } = Typography;

export default function AgentListPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>AI Agents</Title>
                    <Paragraph type="secondary">
                        Create and manage AI agents for automated operations
                    </Paragraph>
                </div>
                <Button type="primary" icon={<IconPlus />}>
                    Create Agent
                </Button>
            </div>

            <Card>
                <Paragraph>No agents configured yet.</Paragraph>
                <Paragraph type="secondary">
                    Create an AI agent to automate server diagnostics and operations.
                </Paragraph>
            </Card>
        </Space>
    );
}
