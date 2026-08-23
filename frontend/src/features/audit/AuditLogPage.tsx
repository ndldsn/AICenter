import { Typography, Card, Space } from '@arco-design/web-react';

const { Title, Paragraph } = Typography;

export default function AuditLogPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div>
                <Title heading={4}>审计日志</Title>
                <Paragraph type="secondary">
                    所有操作与变更的完整历史记录
                </Paragraph>
            </div>

            <Card>
                <Paragraph>暂无审计日志。</Paragraph>
                <Paragraph type="secondary">
                    所有操作都会被记录在这里，用于合规与安全审计。
                </Paragraph>
            </Card>
        </Space>
    );
}
