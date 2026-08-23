import { useEffect, useRef, useState } from 'react';
import {
    Drawer,
    Input,
    Button,
    Space,
    Typography,
    Select,
    Spin,
    Empty,
} from '@arco-design/web-react';
import { IconSend } from '@arco-design/web-react/icon';
import { useProviderModels, streamChat } from './hooks';
import { ChatMessage } from '@/services/ai';

interface Props {
    provider: { id: string; name: string } | null;
    onClose: () => void;
}

interface UIMessage extends ChatMessage {
    streaming?: boolean;
}

export default function ChatDrawer({ provider, onClose }: Props) {
    const { data: models } = useProviderModels(provider?.id || '');
    const [modelId, setModelId] = useState<string>('');
    const [messages, setMessages] = useState<UIMessage[]>([]);
    const [input, setInput] = useState('');
    const [loading, setLoading] = useState(false);
    const scrollRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (models && models.length > 0) {
            const def = models.find((m) => m.is_default) || models[0];
            setModelId(def.model_id);
        }
    }, [models]);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [messages]);

    const handleSend = async () => {
        if (!input.trim() || !provider || !modelId || loading) return;
        const userMsg: UIMessage = { role: 'user', content: input };
        const history = [...messages, userMsg];
        setMessages(history);
        setInput('');
        setLoading(true);

        const assistantMsg: UIMessage = { role: 'assistant', content: '', streaming: true };
        setMessages([...history, assistantMsg]);

        try {
            await streamChat(provider.id, modelId, history, (chunk) => {
                setMessages((prev) => {
                    const next = [...prev];
                    const lastIdx = next.length - 1;
                    next[lastIdx] = {
                        ...next[lastIdx],
                        content: next[lastIdx].content + chunk,
                    };
                    return next;
                });
            });
        } catch (e: any) {
            setMessages((prev) => {
                const next = [...prev];
                const lastIdx = next.length - 1;
                next[lastIdx] = {
                    ...next[lastIdx],
                    content: `[Error] ${e?.message || 'stream failed'}`,
                    streaming: false,
                };
                return next;
            });
        } finally {
            setMessages((prev) => {
                const next = [...prev];
                const lastIdx = next.length - 1;
                next[lastIdx] = { ...next[lastIdx], streaming: false };
                return next;
            });
            setLoading(false);
        }
    };

    return (
        <Drawer
            title={`对话 · ${provider?.name || ''}`}
            visible={!!provider}
            onCancel={onClose}
            width={680}
            footer={null}
        >
            {provider && (
                <Space direction="vertical" size={12} style={{ width: '100%' }}>
                    <Select
                        style={{ width: '100%' }}
                        placeholder="选择模型"
                        value={modelId}
                        onChange={setModelId}
                        options={(models || []).map((m: any) => ({
                            label: `${m.name} (${m.model_id})`,
                            value: m.model_id,
                        }))}
                    />

                    <div
                        ref={scrollRef}
                        style={{
                            height: 420,
                            overflowY: 'auto',
                            padding: 12,
                            background: 'var(--color-fill-2)',
                            borderRadius: 8,
                        }}
                    >
                        {messages.length === 0 ? (
                            <Empty description="开始对话" />
                        ) : (
                            messages.map((m, idx) => (
                                <div
                                    key={idx}
                                    style={{
                                        marginBottom: 12,
                                        textAlign: m.role === 'user' ? 'right' : 'left',
                                    }}
                                >
                                    <Typography.Text type={m.role === 'user' ? 'primary' : undefined}>
                                        {m.role === 'user' ? '你: ' : 'AI: '}
                                    </Typography.Text>
                                    <div
                                        style={{
                                            display: 'inline-block',
                                            padding: '8px 12px',
                                            borderRadius: 8,
                                            background: m.role === 'user' ? 'var(--color-primary-light-4)' : '#fff',
                                            maxWidth: '85%',
                                            whiteSpace: 'pre-wrap',
                                        }}
                                    >
                                        {m.content}
                                        {m.streaming && <Spin size={12} style={{ marginLeft: 6 }} />}
                                    </div>
                                </div>
                            ))
                        )}
                    </div>

                    <div style={{ display: 'flex', gap: 8 }}>
                        <Input
                            placeholder="输入消息..."
                            value={input}
                            onChange={setInput}
                            onPressEnter={handleSend}
                            disabled={loading}
                            style={{ flex: 1 }}
                        />
                        <Button
                            type="primary"
                            icon={<IconSend />}
                            onClick={handleSend}
                            loading={loading}
                        >
                            发送
                        </Button>
                    </div>
                </Space>
            )}
        </Drawer>
    );
}