import { useState, useEffect, useMemo } from 'react';
import {
    Table, Button, Space, Typography, Tag, Card, Modal, Form, Input,
    Select, InputNumber, Switch, Popconfirm, Spin, Empty, Message, Tabs, Statistic,
} from '@arco-design/web-react';
import { Grid } from '@arco-design/web-react';
const { Row, Col } = Grid;
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { monitorApi, AlertRule, AlertEvent, Metric } from '@/services/monitor';
import { useT } from '@/stores/uiStore';

const { Title, Paragraph, Text } = Typography;

const SEVERITY_COLORS: Record<string, string> = {
    info: 'blue', warning: 'orange', critical: 'red',
};
const STATUS_COLORS: Record<string, string> = {
    firing: 'red', acknowledged: 'gray',
};
const METRIC_NAMES = ['cpu.usage', 'memory.usage', 'disk.usage', 'load.1'];
const CONDITIONS = [
    { label: '> (gt)', value: 'gt' },
    { label: '>= (gte)', value: 'gte' },
    { label: '< (lt)', value: 'lt' },
    { label: '<= (lte)', value: 'lte' },
];

export default function MonitorDashboardPage() {
    const t = useT();
    const [activeTab, setActiveTab] = useState('overview');
    const [latest, setLatest] = useState<Metric[]>([]);
    const [alerts, setAlerts] = useState<AlertEvent[]>([]);
    const [rules, setRules] = useState<AlertRule[]>([]);
    const [loading, setLoading] = useState(false);
    const [modalOpen, setModalOpen] = useState(false);
    const [editingId, setEditingId] = useState<string | null>(null);
    const [saving, setSaving] = useState(false);
    const [form] = Form.useForm();

    const refresh = async () => {
        setLoading(true);
        try {
            const [l, a, r] = await Promise.all([
                monitorApi.latestMetrics(),
                monitorApi.listAlerts(),
                monitorApi.listRules(),
            ]);
            setLatest((l as any).items || []);
            setAlerts((a as any).items || []);
            setRules((r as any).items || []);
        } finally {
            setLoading(false);
        }
    };
    useEffect(() => { refresh(); }, []);

    const firingCount = alerts.filter(a => a.status === 'firing').length;

    const openCreate = () => {
        setEditingId(null);
        form.resetFields();
        form.setFieldsValue({
            condition: 'gt', severity: 'warning',
            duration: 0, cooldown: 300, is_enabled: true, threshold: 90,
            metric_name: 'cpu.usage',
        });
        setModalOpen(true);
    };

    const openEdit = (rule: AlertRule) => {
        setEditingId(rule.id);
        form.setFieldsValue({ ...rule });
        setModalOpen(true);
    };

    const onSubmit = async () => {
        const values = await form.validate();
        setSaving(true);
        try {
            if (editingId) {
                await monitorApi.updateRule(editingId, values);
                Message.success(t('monitor.ruleUpdated'));
            } else {
                await monitorApi.createRule(values);
                Message.success(t('monitor.ruleCreated'));
            }
            setModalOpen(false);
            refresh();
        } finally {
            setSaving(false);
        }
    };

    const onDelete = async (id: string) => {
        await monitorApi.deleteRule(id);
        Message.success(t('monitor.ruleDeleted'));
        refresh();
    };

    const onAck = async (id: string) => {
        await monitorApi.ackAlert(id);
        Message.success(t('monitor.acked'));
        refresh();
    };

    const serverStats = useMemo(() => {
        const byServer: Record<string, Record<string, Metric>> = {};
        for (const m of latest) {
            const sid = m.server_id || '-';
            (byServer[sid] ||= {})[m.metric_name] = m;
        }
        return byServer;
    }, [latest]);

    return (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <div>
                    <Title heading={4}>{t('monitor.title')}</Title>
                    <Paragraph type="secondary">{t('monitor.subtitle')}</Paragraph>
                </div>
                <Button icon={<IconRefresh />} onClick={refresh} loading={loading}>{t('monitor.refresh')}</Button>
            </div>

            {firingCount > 0 && (
                <Card style={{ borderColor: 'rgb(var(--red-6))' }}>
                    <Text type="error">{t('monitor.firingAlerts').replace('{n}', firingCount.toString())}</Text>
                </Card>
            )}

            <Tabs defaultActiveTab={activeTab} onChange={setActiveTab}>
                <Tabs.TabPane key="overview" title={t('monitor.overview')}>
                    {loading ? <Spin /> : Object.keys(serverStats).length === 0 ? (
                        <Empty description={t('monitor.noData')} />
                    ) : (
                        Object.entries(serverStats).map(([sid, metrics]) => (
                            <Card key={sid} title={t('monitor.server').replace('{id}', sid)} style={{ marginBottom: 12 }}>
                                <Row gutter={16}>
                                    {['cpu.usage', 'memory.usage', 'disk.usage'].map(name => {
                                        const m = metrics[name];
                                        return (
                                            <Col span={8} key={name}>
                                                <Statistic
                                                    title={name}
                                                    value={m ? m.value.toFixed(1) : '--'}
                                                    suffix={m?.unit || '%'}
                                                />
                                            </Col>
                                        );
                                    })}
                                </Row>
                            </Card>
                        ))
                    )}
                </Tabs.TabPane>

                <Tabs.TabPane key="alerts" title={t('monitor.alertEvents').replace('{n}', alerts.length.toString())}>
                    <Table
                        loading={loading}
                        data={alerts}
                        rowKey="id"
                        pagination={{ pageSize: 10 }}
                        columns={[
                            { title: t('monitor.column.severity'), dataIndex: 'severity', width: 90, render: (v) => <Tag color={SEVERITY_COLORS[v] || 'gray'}>{v}</Tag> },
                            { title: t('monitor.column.status'), dataIndex: 'status', width: 100, render: (v) => <Tag color={STATUS_COLORS[v] || 'gray'}>{v}</Tag> },
                            { title: t('monitor.column.server'), dataIndex: 'server_id', width: 140 },
                            { title: t('monitor.column.message'), dataIndex: 'message' },
                            { title: t('monitor.column.triggeredAt'), dataIndex: 'triggered_at', width: 170 },
                            {
                                title: t('monitor.column.actions'), width: 90,
                                render: (_, record) =>
                                    record.status === 'firing' ? (
                                        <Button size="small" type="primary" onClick={() => onAck(record.id)}>{t('monitor.ack')}</Button>
                                    ) : null,
                            },
                        ]}
                    />
                </Tabs.TabPane>

                <Tabs.TabPane key="rules" title={t('monitor.alertRules').replace('{n}', rules.length.toString())}>
                    <Space style={{ marginBottom: 12 }}>
                        <Button type="primary" icon={<IconPlus />} onClick={openCreate}>{t('monitor.createRule')}</Button>
                    </Space>
                    <Table
                        loading={loading}
                        data={rules}
                        rowKey="id"
                        pagination={{ pageSize: 10 }}
                        columns={[
                            { title: t('monitor.column.name'), dataIndex: 'name' },
                            { title: t('monitor.column.metric'), dataIndex: 'metric_name', width: 130 },
                            {
                                title: t('monitor.column.condition'), width: 150,
                                render: (_, r) => `${condSymbol(r.condition)} ${r.threshold}`,
                            },
                            { title: t('monitor.column.duration'), dataIndex: 'duration', width: 80 },
                            { title: t('monitor.column.severity'), dataIndex: 'severity', width: 90, render: (v) => <Tag color={SEVERITY_COLORS[v]}>{v}</Tag> },
                            {
                                title: t('monitor.column.enabled'), dataIndex: 'is_enabled', width: 80,
                                render: (v) => v ? <Tag color="green">{t('monitor.column.yes')}</Tag> : <Tag color="gray">{t('monitor.column.no')}</Tag>,
                            },
                            { title: t('monitor.column.cooldown'), dataIndex: 'cooldown', width: 80 },
                            {
                                title: t('monitor.column.actions'), width: 130,
                                render: (_, record) => (
                                    <Space>
                                        <Button size="small" onClick={() => openEdit(record)}>{t('monitor.edit')}</Button>
                                        <Popconfirm title={t('monitor.confirmDeleteRule')} onOk={() => onDelete(record.id)}>
                                            <Button size="small" status="danger">{t('monitor.delete')}</Button>
                                        </Popconfirm>
                                    </Space>
                                ),
                            },
                        ]}
                    />
                </Tabs.TabPane>
            </Tabs>

            <Modal
                title={editingId ? t('monitor.editRule') : t('monitor.createRuleTitle')}
                visible={modalOpen}
                onCancel={() => setModalOpen(false)}
                onOk={onSubmit}
                confirmLoading={saving}
                unmountOnExit
            >
                <Form form={form} layout="vertical">
                    <Form.Item label={t('monitor.ruleName')} field="name" rules={[{ required: true, message: t('monitor.enterName') }]}>
                        <Input placeholder={t('monitor.ruleNamePlaceholder')} />
                    </Form.Item>
                    <Form.Item label={t('monitor.metric')} field="metric_name" rules={[{ required: true }]}>
                        <Select allowCreate options={METRIC_NAMES.map(m => ({ label: m, value: m }))} />
                    </Form.Item>
                    <Form.Item label={t('monitor.condition')} field="condition" rules={[{ required: true }]}>
                        <Select options={CONDITIONS} />
                    </Form.Item>
                    <Form.Item label={t('monitor.threshold')} field="threshold" rules={[{ required: true }]}>
                        <InputNumber style={{ width: '100%' }} step={5} />
                    </Form.Item>
                    <Form.Item label={t('monitor.durationSeconds')} field="duration">
                        <InputNumber style={{ width: '100%' }} min={0} step={30} />
                    </Form.Item>
                    <Form.Item label={t('monitor.severity')} field="severity">
                        <Select options={['info', 'warning', 'critical']} />
                    </Form.Item>
                    <Form.Item label={t('monitor.cooldownSeconds')} field="cooldown">
                        <InputNumber style={{ width: '100%' }} min={0} step={60} />
                    </Form.Item>
                    <Form.Item label={t('monitor.enable')} field="is_enabled" triggerPropName="checked">
                        <Switch />
                    </Form.Item>
                </Form>
            </Modal>
        </Space>
    );
}

function condSymbol(c: string): string {
    switch (c) {
        case 'gt': return '>';
        case 'gte': return '>=';
        case 'lt': return '<';
        case 'lte': return '<=';
        default: return c;
    }
}
