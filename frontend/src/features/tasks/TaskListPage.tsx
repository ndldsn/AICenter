import { Typography, Card, Button, Space } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';

const { Title, Paragraph } = Typography;

export default function TaskListPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>Tasks</Title>
                    <Paragraph type="secondary">
                        Schedule and manage automated tasks
                    </Paragraph>
                </div>
                <Button type="primary" icon={<IconPlus />}>
                    Create Task
                </Button>
            </div>

            <Card>
                <Paragraph>No tasks yet.</Paragraph>
                <Paragraph type="secondary">
                    Create scheduled tasks for automated operations.
                </Paragraph>
            </Card>
        </Space>
    );
}
