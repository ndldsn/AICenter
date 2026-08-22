import { useState, useEffect, useMemo } from 'react';
import {
    Table, Button, Space, Typography, Tag, Card, Modal, Form, Input,
    Select, InputNumber, Switch, Popconfirm, Spin, Empty, Message, Tabs, Statistic,
} from '@arco-design/web-react';
import { Grid } from '@arco-design/web-react';
const { Row, Col } = Grid;
import { IconPlus, IconRefresh } from '@arco-design/web-react/icon';
import { monitorApi, AlertRule, AlertEvent, Metric } from '@/services/monitor';

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
                Message.success('规则已更新');
            } else {
                await monitorApi.createRule(values);
                Message.success('规则已创建');
            }
            setModalOpen(false);
            refresh();
        } finally {
            setSaving(false);
        }
    };

    const onDelete = async (id: string) => {
        await monitorApi.deleteRule(id);
        Message.success('规则已删除');
        refresh();
    };

    const onAck = async (id: string) => {
        await monitorApi.ackAlert(id);
        Message.success('已确认');
        refresh();
    };

    // latest metrics per server grouped for overview cards
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
                    <Title heading={4}>监控告警</Title>
                    <Paragraph type="secondary">指标采集、阈值告警与事件管理</Paragraph>
                </div>
                <Button icon={<IconRefresh />} onClick={refresh} loading={loading}>刷新</Button>
            </div>

            {firingCount > 0 && (
                <Card style={{ borderColor: 'rgb(var(--red-6))' }}>
                    <Text type="error">⚠ 当前有 {firingCount} 条 firing 告警待处理</Text>
                </Card>
            )}

            <Tabs defaultActiveTab={activeTab} onChange={setActiveTab}>
                <Tabs.TabPane key="overview" title="概览">
                    {loading ? <Spin /> : Object.keys(serverStats).length === 0 ? (
                        <Empty description="暂无指标数据，等待采集循环写入" />
                    ) : (
                        Object.entries(serverStats).map(([sid, metrics]) => (
                            <Card key={sid} title={`服务器: ${sid}`} style={{ marginBottom: 12 }}>
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

                <Tabs.TabPane key="alerts" title={`告警事件 (${alerts.length})`}>
                    <Table
                        loading={loading}
                        data={alerts}
                        rowKey="id"
                        pagination={{ pageSize: 10 }}
                        columns={[
                            { title: '级别', dataIndex: 'severity', width: 90, render: (v) => <Tag color={SEVERITY_COLORS[v] || 'gray'}>{v}</Tag> },
                            { title: '状态', dataIndex: 'status', width: 100, render: (v) => <Tag color={STATUS_COLORS[v] || 'gray'}>{v}</Tag> },
                            { title: '服务器', dataIndex: 'server_id', width: 140 },
                            { title: '信息', dataIndex: 'message' },
                            { title: '触发时间', dataIndex: 'triggered_at', width: 170 },
                            {
                                title: '操作', width: 90,
                                render: (_, record) =>
                                    record.status === 'firing' ? (
                                        <Button size="small" type="primary" onClick={() => onAck(record.id)}>确认</Button>
                                    ) : null,
                            },
                        ]}
                    />
                </Tabs.TabPane>

                <Tabs.TabPane key="rules" title={`告警规则 (${rules.length})`}>
                    <Space style={{ marginBottom: 12 }}>
                        <Button type="primary" icon={<IconPlus />} onClick={openCreate}>新建规则</Button>
                    </Space>
                    <Table
                        loading={loading}
                        data={rules}
                        rowKey="id"
                        pagination={{ pageSize: 10 }}
                        columns={[
                            { title: '名称', dataIndex: 'name' },
                            { title: '指标', dataIndex: 'metric_name', width: 130 },
                            {
                                title: '条件', width: 150,
                                render: (_, r) => `${condSymbol(r.condition)} ${r.threshold}`,
                            },
                            { title: '持续(s)', dataIndex: 'duration', width: 80 },
                            { title: '级别', dataIndex: 'severity', width: 90, render: (v) => <Tag color={SEVERITY_COLORS[v]}>{v}</Tag> },
                            {
                                title: '启用', dataIndex: 'is_enabled', width: 80,
                                render: (v) => v ? <Tag color="green">是</Tag> : <Tag color="gray">否</Tag>,
                            },
                            { title: '冷却(s)', dataIndex: 'cooldown', width: 80 },
                            {
                                title: '操作', width: 130,
                                render: (_, record) => (
                                    <Space>
                                        <Button size="small" onClick={() => openEdit(record)}>编辑</Button>
                                        <Popconfirm title="确认删除该规则？" onOk={() => onDelete(record.id)}>
                                            <Button size="small" status="danger">删除</Button>
                                        </Popconfirm>
                                    </Space>
                                ),
                            },
                        ]}
                    />
                </Tabs.TabPane>
            </Tabs>

            <Modal
                title={editingId ? '编辑告警规则' : '新建告警规则'}
                visible={modalOpen}
                onCancel={() => setModalOpen(false)}
                onOk={onSubmit}
                confirmLoading={saving}
                unmountOnExit
            >
                <Form form={form} layout="vertical">
                    <Form.Item label="规则名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
                        <Input placeholder="如 CPU 过高" />
                    </Form.Item>
                    <Form.Item label="指标" field="metric_name" rules={[{ required: true }]}>
                        <Select allowCreate options={METRIC_NAMES.map(m => ({ label: m, value: m }))} />
                    </Form.Item>
                    <Form.Item label="条件" field="condition" rules={[{ required: true }]}>
                        <Select options={CONDITIONS} />
                    </Form.Item>
                    <Form.Item label="阈值" field="threshold" rules={[{ required: true }]}>
                        <InputNumber style={{ width: '100%' }} step={5} />
                    </Form.Item>
                    <Form.Item label="持续时间（秒，0 = 立即）" field="duration">
                        <InputNumber style={{ width: '100%' }} min={0} step={30} />
                    </Form.Item>
                    <Form.Item label="级别" field="severity">
                        <Select options={['info', 'warning', 'critical']} />
                    </Form.Item>
                    <Form.Item label="冷却时间（秒）" field="cooldown">
                        <InputNumber style={{ width: '100%' }} min={0} step={60} />
                    </Form.Item>
                    <Form.Item label="启用" field="is_enabled" triggerPropName="checked">
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
