import { Typography, Card, Space } from '@arco-design/web-react';

const { Title, Paragraph } = Typography;

export default function AuditLogPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>Audit Log</Title>
                <Paragraph type="secondary">
                    Complete history of all operations and changes
                </Paragraph>
            </div>

            <Card>
                <Paragraph>No audit logs yet.</Paragraph>
                <Paragraph type="secondary">
                    All operations will be recorded here for compliance and security.
                </Paragraph>
            </Card>
        </Space>
    );
}
