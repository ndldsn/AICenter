import { useState } from 'react';
import { useT } from '@/stores/uiStore';
import {
    Typography, Card, Button, Input, Table, Tag, Spin, Message, Space,
} from '@arco-design/web-react';
import { IconPlayCircle, IconRefresh } from '@arco-design/web-react/icon';
import { useServers } from './hooks';
import { serverApi, Server, BatchResult } from '@/services/servers';

const { Text } = Typography;
const { TextArea } = Input;

export default function BatchCommandPage() {
    const t = useT();
    const [command, setCommand] = useState('echo "batch from $(hostname)"');
    const [timeout, setTimeout] = useState(30);
    const [selected, setSelected] = useState<string[]>([]);
    const [running, setRunning] = useState(false);
    const [results, setResults] = useState<BatchResult[]>([]);
    const [total, setTotal] = useState(0);
    const { data, refetch } = useServers(1, 10000);

    const servers: Server[] = data?.items || [];

    const run = async () => {
        if (!command.trim()) {
            Message.warning(t('batch.enterCommand'));
            return;
        }
        if (selected.length === 0) {
            Message.warning(t('batch.selectServers'));
            return;
        }
        setRunning(true);
        try {
            const r = await serverApi.batchCommand({
                command, server_ids: selected, timeout_seconds: timeout,
            });
            const items: BatchResult[] = (r.data.items || []).map((x) => ({
                ...x, status: (x.status || (x.error ? 'failed' : 'ok')) as 'ok' | 'failed',
            }));
            setResults(items);
            setTotal(items.length);
        } catch {
            Message.error(t('batch.executeFailed'));
        } finally {
            setRunning(false);
        }
    };

    const columns = [
        { title: t('batch.server'), dataIndex: 'server', key: 'server' },
        { title: t('batch.address'), dataIndex: 'host', key: 'host' },
        {
            title: t('batch.status'), dataIndex: 'status', key: 'status',
            render: (v: string) => (
                <Tag color={v === 'ok' ? 'green' : 'red'}>{v === 'ok' ? t('batch.success') : t('batch.failed')}</Tag>
            ),
        },
        { title: t('batch.duration'), dataIndex: 'duration', key: 'duration' },
        { title: t('batch.exitCode'), dataIndex: 'exit_code', key: 'exit_code' },
        {
            title: t('batch.output'), dataIndex: 'stdout', key: 'stdout',
            render: (v: string, r: BatchResult) => (
                <Input.TextArea
                    value={v || r.stderr || r.error || ''}
                    rows={3}
                    readOnly
                    style={{ fontFamily: 'monospace', fontSize: 12 }}
                />
            ),
        },
    ];

    return (
        <div style={{ padding: 24 }}>
            <Space style={{ marginBottom: 16 }}>
                <Button icon={<IconRefresh />} onClick={() => refetch()}>{t('batch.refreshServers')}</Button>
                <Text type="secondary">{t('batch.selected').replace('{n}', String(selected.length)).replace('{t}', String(timeout))}</Text>
            </Space>
            <Card title={t('batch.title')} style={{ marginBottom: 16 }}>
                <TextArea
                    value={command}
                    onChange={setCommand}
                    rows={3}
                    placeholder='例如: df -h /'
                    style={{ fontFamily: 'monospace', marginBottom: 12 }}
                />
                <Space>
                    <Input
                        type="number"
                        min={1}
                        max={600}
                        value={String(timeout)}
                        onChange={(v) => setTimeout(parseInt(v) || 30)}
                        style={{ width: 120 }}
                    />
                    <Button type="primary" icon={<IconPlayCircle />} loading={running} onClick={run}>
                        {running ? t('batch.executing') : t('batch.execute')}
                    </Button>
                </Space>
            </Card>
            <Card title={t('batch.serverList').replace('{n}', String(servers.length))}>
                <Table
                    rowKey="id"
                    columns={[
                        { title: '名称', dataIndex: 'name', key: 'name' },
                        { title: '地址', dataIndex: 'host', key: 'host' },
                        { title: '状态', dataIndex: 'status', key: 'status' },
                    ]}
                    data={servers}
                    rowSelection={{
                        selectedRowKeys: selected,
                        onChange: (keys: (string | number)[]) => setSelected(keys as string[]),
                    }}
                    pagination={false}
                />
            </Card>
            {running && <div style={{ padding: 24 }}><Spin tip={t('batch.running')} /></div>}
            {!running && results.length > 0 && (
                <Card title={t('batch.results').replace('{n}', String(total))} style={{ marginTop: 16 }}>
                    <Table
                        rowKey="server_id"
                        columns={columns}
                        data={results}
                        pagination={false}
                    />
                </Card>
            )}
        </div>
    );
}
