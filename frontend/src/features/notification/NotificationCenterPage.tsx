import { useState, useEffect } from 'react';
import {
    Table, Button, Space, Typography, Tag, Card, Modal, Form, Input,
    Select, Switch, Popconfirm, Spin, Message, Tabs, Radio,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconSend } from '@arco-design/web-react/icon';
import { useT } from '@/stores/uiStore';
import {
    notificationApi, NotificationChannel, NotificationTemplate, DeliveryLog, ChannelType,
} from '@/services/notification';

const { Title, Text } = Typography;

const CHANNEL_TYPES: ChannelType[] = ['webhook', 'email', 'sms', 'im', 'console'];
const EVENT_TYPES = [
    { label: 'alert.fired', value: 'alert.fired' },
    { label: 'approval.requested', value: 'approval.requested' },
    { label: 'approval.resolved', value: 'approval.resolved' },
];

const STATUS_COLORS: Record<string, string> = {
    sent: 'green',
    failed: 'red',
    pending: 'gray',
};

export default function NotificationCenterPage() {
    const t = useT();
    const [activeTab, setActiveTab] = useState('channels');
    return (
        <div style={{ padding: 16 }}>
            <Title heading={3}>{t('notifications.title')}</Title>
            <Text type="secondary">{t('notifications.subtitle')}</Text>
            <Tabs activeTab={activeTab} onChange={setActiveTab} style={{ marginTop: 12 }}>
                <Tabs.TabPane title={t('notifications.channels')} key="channels">
                    <ChannelTab />
                </Tabs.TabPane>
                <Tabs.TabPane title={t('notifications.templates')} key="templates">
                    <TemplateTab />
                </Tabs.TabPane>
                <Tabs.TabPane title={t('notifications.logs')} key="logs">
                    <DeliveryLogTab />
                </Tabs.TabPane>
                <Tabs.TabPane title={t('notifications.testSend')} key="test">
                    <SendTestTab />
                </Tabs.TabPane>
            </Tabs>
        </div>
    );
}

function ChannelTab() {
    const t = useT();
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
            try { cfg = JSON.stringify(JSON.parse(v.config || '{}')); } catch { Message.error(t('notifications.configMustBeJson')); return; }
            if (editing) {
                await notificationApi.updateChannel(editing.id, { name: v.name, type: v.type, is_enabled: v.is_enabled, config: cfg });
            } else {
                await notificationApi.createChannel({ name: v.name, type: v.type, is_enabled: v.is_enabled, config: cfg });
            }
            Message.success(t('notifications.saved'));
            setModalOpen(false);
            refresh();
        } finally { setSaving(false); }
    };

    const remove = async (id: string) => { await notificationApi.deleteChannel(id); Message.success(t('notifications.deleted')); refresh(); };

    const columns = [
        { title: t('notifications.column.name'), dataIndex: 'name' },
        { title: t('notifications.column.type'), dataIndex: 'type', render: (tp: ChannelType) => {
            const labels: Record<ChannelType, string> = {
                webhook: t('notifications.webhook'), email: t('notifications.email'),
                sms: t('notifications.sms'), im: t('notifications.im'), console: t('notifications.console'),
            };
            return <Tag color="arcoblue">{labels[tp]}</Tag>;
        } },
        { title: t('notifications.column.status'), dataIndex: 'is_enabled', render: (e: boolean) => e ? <Tag color="green">{t('notifications.enable')}</Tag> : <Tag>{t('notifications.disable')}</Tag> },
        { title: t('notifications.column.config'), dataIndex: 'config', render: (c?: string) => <Text type="secondary" style={{ fontSize: 12 }}>{c && c.length > 60 ? c.slice(0, 60) + '…' : (c || '-')}</Text> },
        {
            title: t('notifications.column.actions'), render: (_: any, r: NotificationChannel) => (
                <Space>
                    <Button size="small" onClick={() => openEdit(r)}>{t('notifications.edit')}</Button>
                    <Popconfirm title={t('notifications.confirmDeleteChannel')} onOk={() => remove(r.id)}>
                        <Button size="small" status="danger">{t('notifications.delete')}</Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <Card>
            <Space style={{ marginBottom: 12 }}>
                <Button type="primary" icon={<IconPlus />} onClick={openCreate}>{t('notifications.newChannel')}</Button>
                <Button icon={<IconRefresh />} onClick={refresh}>{t('notifications.refresh')}</Button>
            </Space>
            <Spin loading={loading}>
                <Table rowKey="id" columns={columns} data={items} pagination={false} />
            </Spin>
            <Modal
                visible={modalOpen} title={editing ? t('notifications.editChannel') : t('notifications.newChannel')} onOk={save} onCancel={() => setModalOpen(false)}
                confirmLoading={saving}
            >
                <Form form={form} layout="vertical">
                    <Form.Item label={t('notifications.name')} field="name" rules={[{ required: true, message: t('notifications.nameRequired') }]}>
                        <Input placeholder="Slack Webhook" />
                    </Form.Item>
                    <Form.Item label={t('notifications.type')} field="type" rules={[{ required: true }]}>
                        <Select options={CHANNEL_TYPES.map(tp => ({ label: tp, value: tp }))} />
                    </Form.Item>
                    <Form.Item label={t('notifications.configJson')} field="config" extra={t('notifications.configHint')}>
                        <Input.TextArea rows={4} placeholder='{"url":"https://hooks.slack.com/...","token":""}' />
                    </Form.Item>
                    <Form.Item label={t('notifications.enable')} field="is_enabled"><Switch /></Form.Item>
                </Form>
            </Modal>
        </Card>
    );
}

function TemplateTab() {
    const t = useT();
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
    const openEdit = (tp: NotificationTemplate) => {
        setEditing(tp);
        form.setFieldsValue({ name: tp.name, event_type: tp.event_type, subject: tp.subject || '', body: tp.body, channels: tp.channels || '[]', is_enabled: tp.is_enabled });
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
            Message.success(t('notifications.saved'));
            setModalOpen(false);
            refresh();
        } finally { setSaving(false); }
    };
    const remove = async (id: string) => { await notificationApi.deleteTemplate(id); Message.success(t('notifications.deleted')); refresh(); };

    const columns = [
        { title: t('notifications.name'), dataIndex: 'name' },
        { title: t('notifications.column.eventType'), dataIndex: 'event_type', render: (e: string) => <Tag color="purple">{e}</Tag> },
        { title: t('notifications.column.subject'), dataIndex: 'subject', render: (s?: string) => s || <Text type="secondary">-</Text> },
        { title: t('notifications.column.status'), dataIndex: 'is_enabled', render: (e: boolean) => e ? <Tag color="green">{t('notifications.enable')}</Tag> : <Tag>{t('notifications.disable')}</Tag> },
        {
            title: t('notifications.column.actions'), render: (_: any, r: NotificationTemplate) => (
                <Space>
                    <Button size="small" onClick={() => openEdit(r)}>{t('notifications.edit')}</Button>
                    <Popconfirm title={t('notifications.confirmDeleteTemplate')} onOk={() => remove(r.id)}>
                        <Button size="small" status="danger">{t('notifications.delete')}</Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <Card>
            <Space style={{ marginBottom: 12 }}>
                <Button type="primary" icon={<IconPlus />} onClick={openCreate}>{t('notifications.newTemplate')}</Button>
                <Button icon={<IconRefresh />} onClick={refresh}>{t('notifications.refresh')}</Button>
            </Space>
            <Spin loading={loading}>
                <Table rowKey="id" columns={columns} data={items} pagination={false} />
            </Spin>
            <Modal visible={modalOpen} title={editing ? t('notifications.editTemplate') : t('notifications.newTemplate')} onOk={save} onCancel={() => setModalOpen(false)} confirmLoading={saving} style={{ width: 640 }}>
                <Form form={form} layout="vertical">
                    <Form.Item label={t('notifications.name')} field="name" rules={[{ required: true }]}><Input /></Form.Item>
                    <Form.Item label={t('notifications.eventType')} field="event_type" rules={[{ required: true }]}>
                        <Select options={EVENT_TYPES.map(et => ({ label: et.label, value: et.value }))} />
                    </Form.Item>
                    <Form.Item label={t('notifications.templateSubject')} field="subject" extra={t('notifications.titleHint')}>
                        <Input placeholder={t('notifications.titlePlaceholder')} />
                    </Form.Item>
                    <Form.Item label={t('notifications.body')} field="body" rules={[{ required: true }]} extra={t('notifications.bodyHint')}>
                        <Input.TextArea rows={5} />
                    </Form.Item>
                    <Form.Item label={t('notifications.channelTypes')} field="channels">
                        <Input placeholder={t('notifications.channelTypesPlaceholder')} />
                    </Form.Item>
                    <Form.Item label={t('notifications.enable')} field="is_enabled"><Switch /></Form.Item>
                </Form>
            </Modal>
        </Card>
    );
}

function DeliveryLogTab() {
    const t = useT();
    const [items, setItems] = useState<DeliveryLog[]>([]);
    const [loading, setLoading] = useState(false);
    const [status, setStatus] = useState('');

    const refresh = async () => {
        setLoading(true);
        try { setItems((await notificationApi.listDeliveryLogs(status || undefined)).items); } finally { setLoading(false); }
    };
    useEffect(() => { refresh(); }, [status]);

    const columns = [
        { title: t('notifications.column.time'), dataIndex: 'created_at', render: (tp: string) => <Text style={{ fontSize: 12 }}>{tp}</Text> },
        { title: t('notifications.column.event'), dataIndex: 'event_type', render: (e?: string) => <Tag color="purple">{e || '-'}</Tag> },
        { title: t('notifications.column.channel'), dataIndex: 'channel_type', render: (c?: string) => c || '-' },
        { title: t('notifications.column.subject'), dataIndex: 'subject', render: (s?: string) => s || '-' },
        { title: t('notifications.column.status'), dataIndex: 'status', render: (s: string) => <Tag color={STATUS_COLORS[s] || 'gray'}>{s}</Tag> },
        { title: t('notifications.column.error'), dataIndex: 'error_message', render: (e?: string) => e ? <Text type="error" style={{ fontSize: 12 }}>{e}</Text> : '-' },
    ];

    return (
        <Card>
            <Space style={{ marginBottom: 12 }}>
                <Radio.Group type="button" value={status} onChange={(v) => setStatus(v as string)}>
                    <Radio value="">{t('notifications.all')}</Radio>
                    <Radio value="sent">{t('notifications.sent')}</Radio>
                    <Radio value="failed">{t('notifications.failed')}</Radio>
                </Radio.Group>
                <Button icon={<IconRefresh />} onClick={refresh}>{t('notifications.refresh')}</Button>
            </Space>
            <Spin loading={loading}>
                <Table rowKey="id" columns={columns} data={items} pagination={{ pageSize: 20 }} />
            </Spin>
        </Card>
    );
}

function SendTestTab() {
    const t = useT();
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
            if (v.data) { try { data = JSON.parse(v.data); } catch { Message.error(t('notifications.dataMustBeJson')); return; } }
            await notificationApi.sendTest({ event_type: v.event_type, title: v.title, severity: v.severity, message: v.message, data });
            Message.success(t('notifications.testSent'));
        } finally { setSaving(false); }
    };

    return (
        <Card>
            <Form form={form} layout="vertical" style={{ maxWidth: 520 }}>
                <Form.Item label={t('notifications.eventType')} field="event_type" rules={[{ required: true }]}>
                    <Select options={EVENT_TYPES.map(et => ({ label: et.label, value: et.value }))} />
                </Form.Item>
                <Form.Item label={t('notifications.templateSubject')} field="title" rules={[{ required: true }]}><Input /></Form.Item>
                <Form.Item label={t('notifications.level')} field="severity">
                    <Select options={[{ label: 'info', value: 'info' }, { label: 'warning', value: 'warning' }, { label: 'critical', value: 'critical' }]} />
                </Form.Item>
                <Form.Item label={t('notifications.message')} field="message"><Input.TextArea rows={3} /></Form.Item>
                <Form.Item label={t('notifications.data')} field="data" extra={t('notifications.dataHint')}>
                    <Input.TextArea rows={2} />
                </Form.Item>
                <Button type="primary" icon={<IconSend />} loading={saving} onClick={send}>{t('notifications.sendTest')}</Button>
            </Form>
        </Card>
    );
}
