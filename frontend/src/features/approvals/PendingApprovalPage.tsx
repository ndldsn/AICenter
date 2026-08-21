import { Typography, Card, Space } from '@arco-design/web-react';

const { Title, Paragraph } = Typography;

export default function PendingApprovalPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>Approvals</Title>
                <Paragraph type="secondary">
                    Review and approve high-risk operations
                </Paragraph>
            </div>

            <Card>
                <Paragraph>No pending approvals.</Paragraph>
                <Paragraph type="secondary">
                    When AI agents request high-risk operations, they will appear here for your review.
                </Paragraph>
            </Card>
        </Space>
    );
}
