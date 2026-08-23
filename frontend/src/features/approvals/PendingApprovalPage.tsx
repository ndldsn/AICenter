import { Typography, Card, Space } from '@arco-design/web-react';
import { useT } from '@/stores/uiStore';

const { Title, Paragraph } = Typography;

export default function PendingApprovalPage() {
    const t = useT();
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>{t('approvals.title')}</Title>
                <Paragraph type="secondary">
                    {t('approvals.subtitle')}
                </Paragraph>
            </div>

            <Card>
                <Paragraph>{t('approvals.empty')}</Paragraph>
                <Paragraph type="secondary">
                    {t('approvals.hint')}
                </Paragraph>
            </Card>
        </Space>
    );
}
