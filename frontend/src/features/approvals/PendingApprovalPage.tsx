import { Typography, Card, Space } from '@arco-design/web-react';

const { Title, Paragraph } = Typography;

export default function PendingApprovalPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>待审批</Title>
                <Paragraph type="secondary">
                    审查并批准高风险操作
                </Paragraph>
            </div>

            <Card>
                <Paragraph>暂无待审批项。</Paragraph>
                <Paragraph type="secondary">
                    当 AI 智能体请求执行高风险操作时，会在这里展示供你审核。
                </Paragraph>
            </Card>
        </Space>
    );
}
