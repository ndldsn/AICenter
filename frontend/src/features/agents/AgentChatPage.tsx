import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import {
    Card, Space, Typography, Input, Button, Tag, Empty, Spin,
} from '@arco-design/web-react';
import {
    IconRobot, IconSend, IconHistory, IconCheckCircle, IconCloseCircle, IconClockCircle,
} from '@arco-design/web-react/icon';
import { agentApi, AgentSession } from '@/services/agent';
import './AgentChatPage.css';

const { Text } = Typography;

type ChatRole = 'user' | 'assistant' | 'tool';

interface ChatItem {
    id: string;
    role: ChatRole;
    content?: string;
    tool_name?: string;
    tool_args?: any;
    tool_result?: any;
    meta?: any;
    created_at?: string;
}

type RunEvent =
    | { type: 'plan'; data: { text: string; tool_calls: any[] } }
    | { type: 'tool_run'; data: { name: string; args: any } }
    | { type: 'tool_result'; data: { name: string; args: any; result: any } }
    | { type: 'approval_required'; data: any }
    | { type: 'final'; data: { text: string; final: boolean } }
    | { type: 'error'; data: { message: string } }
    | { type: 'done' };

function appendRunEvents(items: ChatItem[], events: RunEvent[]): ChatItem[] {
    const next = [...items];
    for (const evt of events) {
        if (evt.type === 'plan') {
            next.push({ id: crypto.randomUUID(), role: 'assistant', content: evt.data.text, meta: { phase: 'plan', tool_calls: evt.data.tool_calls }, created_at: new Date().toISOString() });
        } else if (evt.type === 'tool_run') {
            next.push({ id: crypto.randomUUID(), role: 'tool', tool_name: evt.data.name, tool_args: evt.data.args, meta: { status: 'running' }, created_at: new Date().toISOString() });
        } else if (evt.type === 'tool_result') {
            const idx = [...next].reverse().findIndex(i => i.tool_name === evt.data.name && i.meta?.status === 'running');
            if (idx >= 0) {
                const pos = next.length - 1 - idx;
                next[pos] = { ...next[pos], meta: { status: evt.data.result?.status || 'done', result: evt.data.result }, created_at: new Date().toISOString() };
            } else {
                next.push({ id: crypto.randomUUID(), role: 'tool', tool_name: evt.data.name, tool_args: evt.data.args, tool_result: evt.data.result, meta: { status: evt.data.result?.status || 'done' }, created_at: new Date().toISOString() });
            }
        } else if (evt.type === 'approval_required') {
            next.push({ id: crypto.randomUUID(), role: 'tool', tool_name: evt.data.tool_name, tool_args: evt.data.tool_args, meta: { status: 'pending_approval', approval_id: evt.data.id }, created_at: new Date().toISOString() });
        } else if (evt.type === 'final') {
            next.push({ id: crypto.randomUUID(), role: 'assistant', content: evt.data.text, meta: { phase: 'final' }, created_at: new Date().toISOString() });
        } else if (evt.type === 'error') {
            next.push({ id: crypto.randomUUID(), role: 'assistant', content: '错误: ' + evt.data.message, meta: { phase: 'error' }, created_at: new Date().toISOString() });
        }
    }
    return next;
}

export default function AgentChatPage() {
    const { id } = useParams<{ id: string }>();
    const [agent, setAgent] = useState<any>(null);
    const [sessions, setSessions] = useState<AgentSession[]>([]);
    const [current, setCurrent] = useState<AgentSession | null>(null);
    const [messages, setMessages] = useState<ChatItem[]>([]);
    const [input, setInput] = useState('');
    const [loading, setLoading] = useState(false);
    const [running, setRunning] = useState(false);
    const bottomRef = useRef<HTMLDivElement>(null);
    const eventSourceRef = useRef<EventSource | null>(null);

    const refreshAgents = useCallback(async () => {
        const list = await agentApi.listAgents();
        const found = list.items.find((a: any) => a.id === id);
        setAgent(found || null);
    }, [id]);

    const refreshSessions = useCallback(async () => {
        const list = await agentApi.listSessions(id);
        setSessions(list.items || []);
    }, [id]);

    const openSession = async (sid: string) => {
        const s = await agentApi.getSession(sid);
        setCurrent(s.session);
        const hist: ChatItem[] = [];
        for (const m of s.messages || []) {
            hist.push({
                id: m.id,
                role: m.role as ChatRole,
                content: m.content,
                tool_name: m.tool_name,
                tool_args: m.tool_args,
                tool_result: m.tool_result,
                meta: m.metadata,
                created_at: m.created_at,
            });
        }
        setMessages(hist);
    };

    const createSession = async (): Promise<string> => {
        if (!id) return '';
        const body = { agent_id: id, query: input || 'new session' };
        const s = await agentApi.createSession(body);
        await refreshSessions();
        setInput('');
        await openSession(s.id);
        return s.id;
    };

    const startRun = async (sessionId: string, query: string) => {
        if (eventSourceRef.current) {
            eventSourceRef.current.close();
        }
        const token = localStorage.getItem('access_token') || 'dev-token-1234';
        const base = `/api/v1/agents/sessions/${encodeURIComponent(sessionId)}/run`;
        const es = new EventSource(`${base}?token=${encodeURIComponent(token)}`, { withCredentials: true } as any);
        eventSourceRef.current = es;

        const buffer: RunEvent[] = [];

        es.addEventListener('open', () => {
            fetch(`${base}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify({ query }),
            }).catch(() => {});
        });

        const flush = () => {
            if (!buffer.length) return;
            setMessages(prev => appendRunEvents(prev, buffer.splice(0, buffer.length)));
        };
        const tick = setInterval(flush, 50);

        es.addEventListener('error', () => {
            clearInterval(tick);
            es.close();
            flush();
            setRunning(false);
        });

        es.addEventListener('message', (ev: MessageEvent) => {
            try {
                const parsed = JSON.parse(ev.data);
                if (parsed.type === 'plan') buffer.push({ type: 'plan', data: parsed.data });
                else if (parsed.type === 'tool_run') buffer.push({ type: 'tool_run', data: parsed.data });
                else if (parsed.type === 'tool_result') buffer.push({ type: 'tool_result', data: parsed.data });
                else if (parsed.type === 'approval_required') buffer.push({ type: 'approval_required', data: parsed.data });
                else if (parsed.type === 'final') buffer.push({ type: 'final', data: parsed.data });
                else if (parsed.type === 'error') buffer.push({ type: 'error', data: parsed.data });
                else if (parsed.type === 'done') { es.close(); clearInterval(tick); setRunning(false); flush(); }
            } catch { /* ignore */ }
        });
    };

    const send = async () => {
        const text = input.trim();
        if (!text || running) return;
        setLoading(true);
        try {
            let sid = current?.id;
            if (!sid) sid = await createSession();
            setMessages(prev => [...prev, { id: crypto.randomUUID(), role: 'user', content: text, created_at: new Date().toISOString() }]);
            setRunning(true);
            await startRun(sid, text);
        } finally {
            setLoading(false);
            setInput('');
        }
    };

    useEffect(() => {
        (async () => {
            await refreshAgents();
            await refreshSessions();
        })();
        return () => { if (eventSourceRef.current) eventSourceRef.current.close(); };
    }, [refreshAgents, refreshSessions]);

    useEffect(() => { bottomRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages, running]);

    useEffect(() => {
        const timer = setInterval(refreshSessions, 10_000);
        return () => clearInterval(timer);
    }, [refreshSessions]);

    if (!agent) {
        return <Card style={{ padding: 60 }}><Empty description="未找到智能体" /></Card>;
    }

    return (
        <div className="agent-chat-page">
            <div className="agent-chat-sidebar">
                <Card title={<Space><IconHistory /> 会话</Space>} extra={
                    <Button size="mini" type="primary" icon={<IconSend />} onClick={async () => {
                        if (eventSourceRef.current) eventSourceRef.current.close();
                        const sid = await createSession();
                        await openSession(sid);
                    }}>新会话</Button>
                }>
                    <Space direction="vertical" style={{ width: '100%' }} size={4}>
                        {sessions.map(s => (
                            <Card key={s.id} size="small" hoverable
                                style={{ width: '100%', background: current?.id === s.id ? 'var(--color-fill-2)' : undefined }}
                                onClick={async () => {
                                    if (eventSourceRef.current) eventSourceRef.current.close();
                                    await openSession(s.id);
                                }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                    <Text ellipsis style={{ maxWidth: 140 }}>{s.title || '未命名'}</Text>
                                    <Tag size="small" color={s.status === 'active' ? 'green' : 'gray'}>{s.status === 'active' ? '进行中' : s.status}</Tag>
                                </div>
                                <Text type="secondary" style={{ fontSize: 12 }}>{new Date(s.started_at).toLocaleString()}</Text>
                            </Card>
                        ))}
                        {!sessions.length && <Empty description="暂无会话" style={{ padding: 20 }} />}
                    </Space>
                </Card>
            </div>

            <div className="agent-chat-main">
                <Card
                    title={<Space><IconRobot /> {agent.name} <Tag size="small">{agent.tool_permission_mode}</Tag></Space>}
                    extra={<Space><Text>模型: {agent.model_id}</Text><Text>温度: {agent.temperature}</Text><Text>迭代: {agent.max_iterations}</Text></Space>}
                >
                    <div className="agent-chat-messages">
                        {messages.length === 0 && (
                            <Empty description="输入问题开始 Agent 运行，结果会展示在这里。" />
                        )}
                        {messages.map(m => (
                            <div key={m.id} className={`chat-bubble ${m.role}`}>
                                <div className="chat-bubble-header">
                                    <Tag size="small" color={m.role === 'user' ? 'blue' : m.role === 'assistant' ? 'purple' : 'orange'}>
                                        {m.role === 'user' ? '用户' : m.role === 'assistant' ? '助手' : '工具'}
                                    </Tag>
                                    {m.tool_name && <Tag size="small">{m.tool_name}</Tag>}
                                    {m.meta?.phase && <Text type="secondary" style={{ fontSize: 12 }}>{m.meta.phase}</Text>}
                                </div>
                                {m.content && <div className="chat-bubble-body">{m.content}</div>}
                                {m.tool_args && <div className="chat-bubble-code">参数: {JSON.stringify(m.tool_args)}</div>}
                                {m.tool_result && <div className="chat-bubble-code">结果: {JSON.stringify(m.tool_result)}</div>}
                                {m.meta?.status && (
                                    <div className="chat-bubble-footer">
                                        <Tag size="small" icon={m.meta.status === 'done' || m.meta.status === 'ok' ? <IconCheckCircle /> : m.meta.status === 'denied' ? <IconCloseCircle /> : <IconClockCircle />}>
                                            {m.meta.status}
                                        </Tag>
                                        {m.meta.approval_id && <Tag color="red" size="small">待审批</Tag>}
                                    </div>
                                )}
                            </div>
                        ))}
                        {running && <Spin size={16} />}
                        <div ref={bottomRef} />
                    </div>

                    <div className="agent-chat-input">
                        <Space>
                            <Input.Search
                                placeholder="输入问题，例如：'list servers' 或 'restart nginx'"
                                value={input}
                                onChange={setInput}
                                onSearch={send}
                                loading={loading || running}
                                disabled={loading || running}
                                style={{ flex: 1 }}
                            />
                            <Button type="primary" loading={loading || running} onClick={send}>运行</Button>
                        </Space>
                    </div>
                </Card>
            </div>
        </div>
    );
}
