import { Typography, Card, Space } from '@arco-design/web-react';
import { useT } from '@/stores/uiStore';

const { Title, Paragraph } = Typography;

export default function AuditLogPage() {
    const t = useT();
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>{t('audit.title')}</Title>
                <Paragraph type="secondary">
                    {t('audit.subtitle')}
                </Paragraph>
            </div>

            <Card>
                <Paragraph>{t('audit.empty')}</Paragraph>
                <Paragraph type="secondary">
                    {t('audit.hint')}
                </Paragraph>
            </Card>
        </Space>
    );
}
