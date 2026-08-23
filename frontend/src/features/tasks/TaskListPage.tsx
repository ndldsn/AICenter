import { Typography, Card, Button, Space } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { useT } from '@/stores/uiStore';

const { Title, Paragraph } = Typography;

export default function TaskListPage() {
    const t = useT();
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>{t('tasks.title')}</Title>
                    <Paragraph type="secondary">
                        {t('tasks.subtitle')}
                    </Paragraph>
                </div>
                <Button type="primary" icon={<IconPlus />}>
                    {t('tasks.create')}
                </Button>
            </div>

            <Card>
                <Paragraph>{t('tasks.empty')}</Paragraph>
                <Paragraph type="secondary">
                    {t('tasks.hint')}
                </Paragraph>
            </Card>
        </Space>
    );
}
