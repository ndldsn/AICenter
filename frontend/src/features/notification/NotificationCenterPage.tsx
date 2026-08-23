import { useState, useEffect } from 'react';
import {
    Table, Button, Space, Typography, Tag, Card, Modal, Form, Input,
    Select, Switch, Popconfirm, Spin, Message, Tabs, Radio,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconSend } from '@arco-design/web-react/icon';
import {
    notificationApi, NotificationChannel, NotificationTemplate, DeliveryLog, ChannelType,
} from '@/services/notification';

const { Title, Text } = Typography;

const CHANNEL_TYPE_LABELS: Record<ChannelType, string> = {
    webhook: 'Webhook',
    email: '邮件',
    sms: '短信',
    im: 'IM',
    console: '控制台 (开发)',
};

const CHANNEL_TYPES: ChannelType[] = ['webhook', 'email', 'sms', 'im', 'console'];
const EVENT_TYPES = [
    { label: '告警触发', value: 'alert.fired' },
    { label: '审批请求', value: 'approval.requested' },
    { label: '审批结果', value: 'approval.resolved' },
];

const STATUS_COLORS: Record<string, string> = {
    sent: 'green',
    failed: 'red',
    pending: 'gray',
};

export default function NotificationCenterPage() {
    const [activeTab, setActiveTab] = useState('channels');
    return (
        <div style={{ padding: 16 }}>
            <Title heading={3}>通知中心</Title>
            <Text type="secondary">管理通知渠道、模板与投递日志；告警与审批事件将按模板自动推送。</Text>
            <Tabs activeTab={activeTab} onChange={setActiveTab} style={{ marginTop: 12 }}>
                <Tabs.TabPane title="通知渠道" key="channels">
                    <ChannelTab />
                </Tabs.TabPane>
                <Tabs.TabPane title="通知模板" key="templates">
                    <TemplateTab />
                </Tabs.TabPane>
                <Tabs.TabPane title="投递日志" key="logs">
                    <DeliveryLogTab />
                </Tabs.TabPane>
                <Tabs.TabPane title="测试发送" key="test">
                    <SendTestTab />
                </Tabs.TabPane>
            </Tabs>
        </div>
    );
}

function ChannelTab() {
    const [items, setItems] = useState<NotificationChannel[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalOpen, setModalOpen] = useState(false);
    const [editing, setEditing] = useState<NotificationChannel | null>(null);
    const [saving, setSaving] = useState(false);
    const [form] = Form.useForm();

    const refresh = async () => {
        setLoading(true);
        try {
            const data = await notificationApi.listChannels();
            setItems(data.items);
        } finally {
            setLoading(false);
        }
    };
    useEffect(() => { refresh(); }, []);

    const openCreate = () => { setEditing(null); form.resetFields(); form.setFieldValue('type', 'webhook'); form.setFieldValue('is_enabled', true); setModalOpen(true); };
    const openEdit = (ch: NotificationChannel) => {
        setEditing(ch);
        form.setFieldsValue({ name: ch.name, type: ch.type, is_enabled: ch.is_enabled, config: ch.config || '{}' });
        setModalOpen(true);
    };

    const save = async () => {
        const v = await form.validate();
        setSaving(true);
        try {
            let cfg = '{}';
            try { cfg = JSON.stringify(JSON.parse(v.config || '{}')); } catch { Message.error('配置必须是合法 JSON'); return; }
            if (editing) {
                await notificationApi.updateChannel(editing.id, { name: v.name, type: v.type, is_enabled: v.is_enabled, config: cfg });
            } else {
                await notificationApi.createChannel({ name: v.name, type: v.type, is_enabled: v.is_enabled, config: cfg });
            }
            Message.success('已保存');
            setModalOpen(false);
            refresh();
        } finally { setSaving(false); }
    };

    const remove = async (id: string) => { await notificationApi.deleteChannel(id); Message.success('已删除'); refresh(); };

    const columns = [
        { title: '名称', dataIndex: 'name' },
        { title: '类型', dataIndex: 'type', render: (t: ChannelType) => <Tag color="arcoblue">{CHANNEL_TYPE_LABELS[t]}</Tag> },
        { title: '状态', dataIndex: 'is_enabled', render: (e: boolean) => e ? <Tag color="green">启用</Tag> : <Tag>禁用</Tag> },
        { title: '配置预览', dataIndex: 'config', render: (c?: string) => <Text type="secondary" style={{ fontSize: 12 }}>{c && c.length > 60 ? c.slice(0, 60) + '…' : (c || '-')}</Text> },
        {
            title: '操作', render: (_: any, r: NotificationChannel) => (
                <Space>
                    <Button size="small" onClick={() => openEdit(r)}>编辑</Button>
                    <Popconfirm title="确认删除该渠道？" onOk={() => remove(r.id)}>
                        <Button size="small" status="danger">删除</Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <Card>
            <Space style={{ marginBottom: 12 }}>
                <Button type="primary" icon={<IconPlus />} onClick={openCreate}>新建渠道</Button>
                <Button icon={<IconRefresh />} onClick={refresh}>刷新</Button>
            </Space>
            <Spin loading={loading}>
                <Table rowKey="id" columns={columns} data={items} pagination={false} />
            </Spin>
            <Modal
                visible={modalOpen} title={editing ? '编辑渠道' : '新建渠道'} onOk={save} onCancel={() => setModalOpen(false)}
                confirmLoading={saving}
            >
                <Form form={form} layout="vertical">
                    <Form.Item label="名称" field="name" rules={[{ required: true, message: '必填' }]}>
                        <Input placeholder="如 Slack Webhook" />
                    </Form.Item>
                    <Form.Item label="类型" field="type" rules={[{ required: true }]}>
                        <Select options={CHANNEL_TYPES.map(t => ({ label: CHANNEL_TYPE_LABELS[t], value: t }))} />
                    </Form.Item>
                    <Form.Item label="配置 (JSON)" field="config" extra="webhook: {url, token}; email/sms: {to}; im: {token}">
                        <Input.TextArea rows={4} placeholder='{"url":"https://hooks.slack.com/...","token":""}' />
                    </Form.Item>
                    <Form.Item label="启用" field="is_enabled"><Switch /></Form.Item>
                </Form>
            </Modal>
        </Card>
    );
}

function TemplateTab() {
    const [items, setItems] = useState<NotificationTemplate[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalOpen, setModalOpen] = useState(false);
    const [editing, setEditing] = useState<NotificationTemplate | null>(null);
    const [saving, setSaving] = useState(false);
    const [form] = Form.useForm();

    const refresh = async () => {
        setLoading(true);
        try { setItems((await notificationApi.listTemplates()).items); } finally { setLoading(false); }
    };
    useEffect(() => { refresh(); }, []);

    const openCreate = () => { setEditing(null); form.resetFields(); form.setFieldValue('event_type', 'alert.fired'); form.setFieldValue('is_enabled', true); form.setFieldValue('channels', '["console"]'); setModalOpen(true); };
    const openEdit = (t: NotificationTemplate) => {
        setEditing(t);
        form.setFieldsValue({ name: t.name, event_type: t.event_type, subject: t.subject || '', body: t.body, channels: t.channels || '[]', is_enabled: t.is_enabled });
        setModalOpen(true);
    };

    const save = async () => {
        const v = await form.validate();
        setSaving(true);
        try {
            if (editing) {
                await notificationApi.updateTemplate(editing.id, v);
            } else {
                await notificationApi.createTemplate(v);
            }
            Message.success('已保存');
            setModalOpen(false);
            refresh();
        } finally { setSaving(false); }
    };
    const remove = async (id: string) => { await notificationApi.deleteTemplate(id); Message.success('已删除'); refresh(); };

    const columns = [
        { title: '名称', dataIndex: 'name' },
        { title: '事件类型', dataIndex: 'event_type', render: (e: string) => <Tag color="purple">{e}</Tag> },
        { title: '标题', dataIndex: 'subject', render: (s?: string) => s || <Text type="secondary">-</Text> },
        { title: '状态', dataIndex: 'is_enabled', render: (e: boolean) => e ? <Tag color="green">启用</Tag> : <Tag>禁用</Tag> },
        {
            title: '操作', render: (_: any, r: NotificationTemplate) => (
                <Space>
                    <Button size="small" onClick={() => openEdit(r)}>编辑</Button>
                    <Popconfirm title="确认删除？" onOk={() => remove(r.id)}>
                        <Button size="small" status="danger">删除</Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <Card>
            <Space style={{ marginBottom: 12 }}>
                <Button type="primary" icon={<IconPlus />} onClick={openCreate}>新建模板</Button>
                <Button icon={<IconRefresh />} onClick={refresh}>刷新</Button>
            </Space>
            <Spin loading={loading}>
                <Table rowKey="id" columns={columns} data={items} pagination={false} />
            </Spin>
            <Modal visible={modalOpen} title={editing ? '编辑模板' : '新建模板'} onOk={save} onCancel={() => setModalOpen(false)} confirmLoading={saving} style={{ width: 640 }}>
                <Form form={form} layout="vertical">
                    <Form.Item label="名称" field="name" rules={[{ required: true }]}><Input /></Form.Item>
                    <Form.Item label="事件类型" field="event_type" rules={[{ required: true }]}>
                        <Select options={EVENT_TYPES} />
                    </Form.Item>
                    <Form.Item label="标题" field="subject" extra="支持 {{.Title}} 等占位符">
                        <Input placeholder="[AICenter 告警] {{.Title}}" />
                    </Form.Item>
                    <Form.Item label="正文" field="body" rules={[{ required: true }]} extra="可用变量: {{.Title}} {{.Severity}} {{.Message}} {{.Data.xxx}}">
                        <Input.TextArea rows={5} />
                    </Form.Item>
                    <Form.Item label="渠道类型 (JSON 数组)" field="channels">
                        <Input placeholder='["console","webhook"]' />
                    </Form.Item>
                    <Form.Item label="启用" field="is_enabled"><Switch /></Form.Item>
                </Form>
            </Modal>
        </Card>
    );
}

function DeliveryLogTab() {
    const [items, setItems] = useState<DeliveryLog[]>([]);
    const [loading, setLoading] = useState(false);
    const [status, setStatus] = useState('');

    const refresh = async () => {
        setLoading(true);
        try { setItems((await notificationApi.listDeliveryLogs(status || undefined)).items); } finally { setLoading(false); }
    };
    useEffect(() => { refresh(); }, [status]);

    const columns = [
        { title: '时间', dataIndex: 'created_at', render: (t: string) => <Text style={{ fontSize: 12 }}>{t}</Text> },
        { title: '事件', dataIndex: 'event_type', render: (e?: string) => <Tag color="purple">{e || '-'}</Tag> },
        { title: '渠道', dataIndex: 'channel_type', render: (c?: string) => c || '-' },
        { title: '标题', dataIndex: 'subject', render: (s?: string) => s || '-' },
        { title: '状态', dataIndex: 'status', render: (s: string) => <Tag color={STATUS_COLORS[s] || 'gray'}>{s}</Tag> },
        { title: '错误', dataIndex: 'error_message', render: (e?: string) => e ? <Text type="error" style={{ fontSize: 12 }}>{e}</Text> : '-' },
    ];

    return (
        <Card>
            <Space style={{ marginBottom: 12 }}>
                <Radio.Group type="button" value={status} onChange={(v) => setStatus(v as string)}>
                    <Radio value="">全部</Radio>
                    <Radio value="sent">已发送</Radio>
                    <Radio value="failed">失败</Radio>
                </Radio.Group>
                <Button icon={<IconRefresh />} onClick={refresh}>刷新</Button>
            </Space>
            <Spin loading={loading}>
                <Table rowKey="id" columns={columns} data={items} pagination={{ pageSize: 20 }} />
            </Spin>
        </Card>
    );
}

function SendTestTab() {
    const [saving, setSaving] = useState(false);
    const [form] = Form.useForm();
    useEffect(() => {
        form.setFieldsValue({ event_type: 'alert.fired', title: '测试告警', severity: 'warning', message: '这是一条测试通知', data: '' });
    }, []);

    const send = async () => {
        const v = await form.validate();
        setSaving(true);
        try {
            let data: Record<string, string> | undefined;
            if (v.data) { try { data = JSON.parse(v.data); } catch { Message.error('data 必须是合法 JSON'); return; } }
            await notificationApi.sendTest({ event_type: v.event_type, title: v.title, severity: v.severity, message: v.message, data });
            Message.success('已触发发送，请到投递日志查看结果');
        } finally { setSaving(false); }
    };

    return (
        <Card>
            <Form form={form} layout="vertical" style={{ maxWidth: 520 }}>
                <Form.Item label="事件类型" field="event_type" rules={[{ required: true }]}><Select options={EVENT_TYPES} /></Form.Item>
                <Form.Item label="标题" field="title" rules={[{ required: true }]}><Input /></Form.Item>
                <Form.Item label="级别" field="severity">
                    <Select options={[{ label: 'info', value: 'info' }, { label: 'warning', value: 'warning' }, { label: 'critical', value: 'critical' }]} />
                </Form.Item>
                <Form.Item label="正文" field="message"><Input.TextArea rows={3} /></Form.Item>
                <Form.Item label="附加数据 (JSON)" field="data" extra='如 {"server_id":"srv-1","value":"96.5"}'>
                    <Input.TextArea rows={2} />
                </Form.Item>
                <Button type="primary" icon={<IconSend />} loading={saving} onClick={send}>发送测试</Button>
            </Form>
        </Card>
    );
}
