import { useState, useEffect } from 'react';
import {
    Table,
    Button,
    Space,
    Typography,
    Tag,
    Card,
    Modal,
    Form,
    Input,
    Select,
    Switch,
    InputNumber,
    Popconfirm,
    Tag as TagOld,
    Spin,
    Empty,
    Message,
} from '@arco-design/web-react';
import { IconPlus, IconDelete, IconRobot, IconMessage } from '@arco-design/web-react/icon';
import { agentApi, Agent } from '@/services/agent';
import { aiApi, AIProvider } from '@/services/ai';
import { useT } from '@/stores/uiStore';

const { Title, Paragraph } = Typography;

const MODES = [
    { label: '全部允许 (allow_all)', value: 'allow_all' },
    { label: '全部拒绝 (deny_all)', value: 'deny_all' },
    { label: '人工审批 (manual)', value: 'manual' },
];

const DEFAULT_TOOLS = ['list_servers', 'get_server_info', 'list_models', 'echo', 'restart_service'];

export default function AgentListPage() {
    const t = useT();
    const [agents, setAgents] = useState<Agent[]>([]);
    const [providers, setProviders] = useState<AIProvider[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalOpen, setModalOpen] = useState(false);
    const [form] = Form.useForm();
    const [saving, setSaving] = useState(false);
    const [editingId, setEditingId] = useState<string | null>(null);
    const [selected, setSelected] = useState<Agent | null>(null);
    const [chatOpen, setChatOpen] = useState(false);
    const [chatLoading, setChatLoading] = useState(false);
    const [chatInput, setChatInput] = useState('');
    const [chatOutput, setChatOutput] = useState<string>('');
    const [chatPlans, setChatPlans] = useState<any[]>([]);
    const [showApprovals, setShowApprovals] = useState(false);
    const [approvals, setApprovals] = useState<any[]>([]);

    const refresh = async () => {
        setLoading(true);
        try {
            const [a, p] = await Promise.all([agentApi.listAgents(), aiApi.listProviders()]);
            setAgents((a as any).items || []);
            setProviders((p as any).data || []);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { refresh(); }, []);

    const onProviderChange = (id: string) => {
        const prov = providers.find(p => p.id === id);
        if (prov) {
            (aiApi.listModels(prov.id) as any).then((r: any) => (void r)).catch(() => {});
        }
    };

    const openCreate = () => {
        setEditingId(null);
        setSelected(null);
        form.resetFields();
        form.setFieldValue('tool_permission_mode', 'manual');
        form.setFieldValue('is_enabled', true);
        form.setFieldValue('temperature', 0.7);
        form.setFieldValue('max_tokens', 4096);
        form.setFieldValue('max_iterations', 10);
        form.setFieldValue('tools', ['list_servers', 'echo']);
        setModalOpen(true);
    };

    const openEdit = (a: Agent) => {
        setEditingId(a.id);
        setSelected(a);
        form.setFieldsValue({
            name: a.name,
            description: a.description,
            model_id: a.model_id,
            system_prompt: a.system_prompt,
            temperature: a.temperature,
            max_tokens: a.max_tokens,
            max_iterations: a.max_iterations,
            tools: a.tools,
            tool_permission_mode: a.tool_permission_mode,
            require_approval_for: a.require_approval_for,
            is_enabled: a.is_enabled,
        });
        onProviderChange(a.model_id);
        setModalOpen(true);
    };

    const onSubmit = async () => {
        const values = await form.validate();
        setSaving(true);
        try {
            if (editingId) {
                await agentApi.updateAgent(editingId, values);
                Message.success(t('agents.updated'));
            } else {
                await agentApi.createAgent(values);
                Message.success(t('agents.created'));
            }
            setModalOpen(false);
            refresh();
        } finally {
            setSaving(false);
        }
    };

    const onDelete = async (id: string) => {
        await agentApi.deleteAgent(id);
        Message.success(t('agents.deleted'));
        refresh();
    };

    const openChat = async (a: Agent) => {
        setSelected(a);
        setChatOpen(true);
        setChatOutput('');
        setChatPlans([]);
        setChatInput('');
    };

    const sendChat = async () => {
        if (!selected || !chatInput.trim()) return;
        setChatLoading(true);
        setChatOutput('正在创建会话...');
        try {
            const sess = await agentApi.createSession({
                agent_id: selected.id,
                query: chatInput,
            });
            const result = await agentApi.sendMessage(sess.id, chatInput);
            setChatOutput(result.data?.planned_text || JSON.stringify(result.data, null, 2));
            const runs = result.data?.tool_runs || [];
            setChatPlans(runs);
            if (result.data?.approval) {
                Message.warning(t('agents.approval.required').replace('{tool}', result.data.approval.tool_name));
                setShowApprovals(true);
                loadApprovals();
            }
        } catch (e: any) {
            setChatOutput(`错误: ${e.message}`);
        } finally {
            setChatLoading(false);
        }
    };

    const loadApprovals = async () => {
        try {
            const r = await agentApi.listApprovals('pending') as any;
            setApprovals(r.data?.items || []);
        } catch {
            setApprovals([]);
        }
    };

    const handleApprove = async (id: string) => {
        await agentApi.approve(id);
        Message.success(t('agents.approval.approved'));
        loadApprovals();
    };

    const handleReject = async (id: string) => {
        await agentApi.reject(id);
        Message.success(t('agents.approval.rejected'));
        loadApprovals();
    };

    const columns = [
        {
            title: t('agents.column.name'),
            dataIndex: 'name',
            render: (_: any, r: Agent) => (
                <Space>
                    <IconRobot />
                    <strong>{r.name}</strong>
                    <Tag size="small">{r.tool_permission_mode}</Tag>
                    {r.is_enabled ? <Tag color="green" size="small">{t('agents.enabled')}</Tag> : <Tag color="gray" size="small">{t('agents.disabled')}</Tag>}
                </Space>
            ),
        },
        {
            title: t('agents.column.model'),
            dataIndex: 'model_id',
            render: (v: string) => <Tag size="small">{v}</Tag>,
        },
        {
            title: t('agents.column.tools'),
            render: (_: any, r: Agent) => (
                <Space wrap>
                    {(r.tools || []).map(tool => <TagOld key={tool} size="small">{tool}</TagOld>)}
                </Space>
            ),
        },
        {
            title: t('agents.column.tempIter'),
            render: (_: any, r: Agent) => <span>{r.temperature} / {r.max_iterations}</span>,
        },
        {
            title: t('agents.column.actions'),
            render: (_: any, r: Agent) => (
                <Space>
                    <Button size="mini" type="text" icon={<IconMessage />} onClick={() => openChat(r)}>{t('agents.action.chat')}</Button>
                    <Button size="mini" type="text" onClick={() => openEdit(r)}>{t('agents.action.edit')}</Button>
                    <Popconfirm title={t('agents.confirmDelete')} onOk={() => onDelete(r.id)}>
                        <Button size="mini" type="text" status="danger" icon={<IconDelete />}>{t('agents.action.delete')}</Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>{t('agents.title')}</Title>
                    <Paragraph type="secondary">
                        {t('agents.subtitle')}
                    </Paragraph>
                </div>
                <Space>
                    <Button type="secondary" onClick={() => { setShowApprovals(!showApprovals); loadApprovals(); }}>
                        {t('agents.approvals').replace('{n}', approvals.length.toString())}
                    </Button>
                    <Button icon={<IconPlus />} onClick={refresh}>{t('agents.refresh')}</Button>
                    <Button type="primary" icon={<IconPlus />} onClick={openCreate}>{t('agents.create')}</Button>
                </Space>
            </div>

            <Card>
                {loading ? <Spin /> : (
                    agents.length ? (
                        <Table<Agent> rowKey="id" data={agents} columns={columns} pagination={false} />
                    ) : (
                        <Empty description={t('agents.empty')} />
                    )
                )}
            </Card>

            <Modal
                title={editingId ? t('agents.edit') : t('agents.create')}
                visible={modalOpen}
                onOk={onSubmit}
                onCancel={() => setModalOpen(false)}
                okText={editingId ? t('common.update') : t('common.create')}
                confirmLoading={saving}
                style={{ width: 620 }}
            >
                <Form form={form} layout="vertical">
                    <Form.Item label={t('agents.name')} field="name" rules={[{ required: true }]}>
                        <Input placeholder="web-maintenance" />
                    </Form.Item>
                    <Form.Item label={t('agents.description')} field="description">
                        <Input.TextArea rows={2} />
                    </Form.Item>
                    <Form.Item label={t('agents.model')} field="model_id" rules={[{ required: true }]}>
                        <Select
                            options={providers.map(p => ({
                                label: `${p.display_name || p.name} (${p.id})`,
                                value: p.id,
                            }))}
                            onChange={onProviderChange}
                        />
                    </Form.Item>
                    <Form.Item label={t('agents.systemPrompt')} field="system_prompt">
                        <Input.TextArea rows={2} />
                    </Form.Item>
                    <Form.Item label={t('agents.permissionMode')} field="tool_permission_mode" rules={[{ required: true }]}>
                        <Select options={MODES.map(m => ({ label: m.label, value: m.value }))} />
                    </Form.Item>
                    <Form.Item label={t('agents.availableTools')} field="tools">
                        <Select
                            mode="multiple"
                            options={DEFAULT_TOOLS.map(tool => ({ label: tool, value: tool }))}
                        />
                    </Form.Item>
                    <Form.Item label={t('agents.requireApproval')} field="require_approval_for">
                        <Select
                            mode="multiple"
                            options={DEFAULT_TOOLS.map(tool => ({ label: tool, value: tool }))}
                        />
                    </Form.Item>
                    <Form.Item label={t('agents.temperature')} field="temperature">
                        <InputNumber min={0} max={2} step={0.1} style={{ width: 120 }} />
                    </Form.Item>
                    <Form.Item label={t('agents.maxTokens')} field="max_tokens">
                        <InputNumber min={1} max={16384} style={{ width: 120 }} />
                    </Form.Item>
                    <Form.Item label={t('agents.maxIterations')} field="max_iterations">
                        <InputNumber min={1} max={50} style={{ width: 120 }} />
                    </Form.Item>
                    <Form.Item label={t('agents.enable')} field="is_enabled">
                        <Switch checked={(form.getFieldValue('is_enabled') as boolean)}
                                onChange={(v: boolean) => form.setFieldValue('is_enabled', v)} />
                    </Form.Item>
                </Form>
            </Modal>

            <AgentChatDrawer
                open={chatOpen}
                onClose={() => setChatOpen(false)}
                loading={chatLoading}
                input={chatInput}
                onInput={setChatInput}
                onSend={sendChat}
                output={chatOutput}
                plans={chatPlans}
            />

            <ApprovalPanel
                visible={showApprovals}
                approvals={approvals}
                onApprove={handleApprove}
                onReject={handleReject}
                onClose={() => setShowApprovals(false)}
            />
        </Space>
    );
}

function AgentChatDrawer(props: {
    open: boolean; onClose: () => void; loading: boolean;
    input: string; onInput: (v: string) => void; onSend: () => void;
    output: string; plans: any[];
}) {
    const t = useT();
    return (
        <Card
            style={{ width: '100%', display: props.open ? 'block' : 'none' }}
            title={t('agents.session.title')}
            extra={<Button size="mini" onClick={props.onClose}>{t('agents.session.close')}</Button>}
        >
            <Space direction="vertical" size={10} style={{ width: '100%' }}>
                <div style={{ display: 'flex', gap: 8 }}>
                    <Input
                        value={props.input}
                        onChange={props.onInput}
                        onPressEnter={props.onSend}
                        placeholder={t('agents.session.placeholder')}
                        disabled={props.loading}
                        style={{ flex: 1 }}
                    />
                    <Button type="primary" loading={props.loading} onClick={props.onSend}>{t('agents.session.run')}</Button>
                </div>
                <div
                    style={{
                        minHeight: 120,
                        maxHeight: 260,
                        overflowY: 'auto',
                        padding: 10,
                        background: 'var(--color-fill-2)',
                        borderRadius: 8,
                        whiteSpace: 'pre-wrap',
                        fontFamily: 'monospace',
                        fontSize: 12,
                    }}
                >
                    {props.output || <span style={{ color: 'var(--color-text-3)' }}>{t('agents.session.outputHint')}</span>}
                </div>
                {props.plans.length > 0 && (
                    <Space direction="vertical" size={6}>
                        <Typography.Text type="secondary" style={{ fontSize: 12 }}>{t('agents.session.toolPlan')}</Typography.Text>
                        <Space wrap>
                            {props.plans.map((r, i) => (
                                <Tag key={i} color={r.result?.ok ? 'green' : 'red'}>
                                    {r.name} → {r.result?.status || '…'}
                                </Tag>
                            ))}
                        </Space>
                    </Space>
                )}
            </Space>
        </Card>
    );
}

function ApprovalPanel(props: {
    visible: boolean;
    approvals: any[];
    onApprove: (id: string) => void;
    onReject: (id: string) => void;
    onClose: () => void;
}) {
    const t = useT();
    if (!props.visible) return null;
    return (
        <Card title={t('agents.approval.title')} style={{ width: '100%' }} extra={<Button size="mini" onClick={props.onClose}>{t('agents.session.close')}</Button>}>
            {props.approvals.length === 0 ? (
                <Empty description={t('agents.approval.noPending')} />
            ) : (
                <Table<any>
                    rowKey="id"
                    data={props.approvals}
                    pagination={false}
                    columns={[
                        { title: t('agents.approval.tool'), dataIndex: 'tool_name', render: (v: string) => <Tag>{v}</Tag> },
                        { title: t('agents.approval.risk'), dataIndex: 'risk_level' },
                        { title: t('agents.approval.requester'), dataIndex: 'requested_by' },
                        { title: t('agents.approval.created'), dataIndex: 'created_at' },
                        {
                            title: t('agents.approval.actions'),
                            render: (_: any, r: any) => (
                                <Space>
                                    <Button size="mini" type="primary" onClick={() => props.onApprove(r.id)}>{t('agents.approval.approve')}</Button>
                                    <Button size="mini" status="danger" onClick={() => props.onReject(r.id)}>{t('agents.approval.reject')}</Button>
                                </Space>
                            ),
                        },
                    ]}
                />
            )}
        </Card>
    );
}
