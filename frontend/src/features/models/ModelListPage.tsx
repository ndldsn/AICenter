import { Typography, Card, Button, Space } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';

const { Title, Paragraph } = Typography;

export default function ModelListPage() {
    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>AI Models</Title>
                    <Paragraph type="secondary">
                        Configure AI providers and models for Agent operations
                    </Paragraph>
                </div>
                <Button type="primary" icon={<IconPlus />}>
                    Add Provider
                </Button>
            </div>

            <Card>
                <Paragraph>No AI providers configured yet.</Paragraph>
                <Paragraph type="secondary">
                    Add an AI provider (OpenAI, Anthropic, Gemini, DeepSeek, Ollama) to enable Agent operations.
                </Paragraph>
            </Card>
        </Space>
    );
}
