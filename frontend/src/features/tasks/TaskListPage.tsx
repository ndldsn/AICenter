import { Typography, Card, Button, Space } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';

const { Title, Paragraph } = Typography;

export default function TaskListPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>任务</Title>
                    <Paragraph type="secondary">
                        定时安排和管理自动化运维任务
                    </Paragraph>
                </div>
                <Button type="primary" icon={<IconPlus />}>
                    创建任务
                </Button>
            </div>

            <Card>
                <Paragraph>暂无任务。</Paragraph>
                <Paragraph type="secondary">
                    创建定时任务来实现自动化运维。
                </Paragraph>
            </Card>
        </Space>
    );
}
