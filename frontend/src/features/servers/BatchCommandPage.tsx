import { useState } from 'react';
import {
    Typography, Card, Button, Input, Table, Tag, Spin, Message, Space,
} from '@arco-design/web-react';
import { IconPlayCircle, IconRefresh } from '@arco-design/web-react/icon';
import { useServers } from './hooks';
import { serverApi, Server, BatchResult } from '@/services/servers';

const { Text } = Typography;
const { TextArea } = Input;

export default function BatchCommandPage() {
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
            Message.warning('请输入命令');
            return;
        }
        if (selected.length === 0) {
            Message.warning('请选择服务器');
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
            Message.error('批量执行失败');
        } finally {
            setRunning(false);
        }
    };

    const columns = [
        { title: '服务器', dataIndex: 'server', key: 'server' },
        { title: '地址', dataIndex: 'host', key: 'host' },
        {
            title: '状态', dataIndex: 'status', key: 'status',
            render: (v: string) => (
                <Tag color={v === 'ok' ? 'green' : 'red'}>{v === 'ok' ? '成功' : '失败'}</Tag>
            ),
        },
        { title: '耗时', dataIndex: 'duration', key: 'duration' },
        { title: '退出码', dataIndex: 'exit_code', key: 'exit_code' },
        {
            title: '输出', dataIndex: 'stdout', key: 'stdout',
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
                <Button icon={<IconRefresh />} onClick={() => refetch()}>刷新服务器</Button>
                <Text type="secondary">已选 {selected.length} 台 · 命令超时 {timeout}s</Text>
            </Space>
            <Card title="批量执行命令" style={{ marginBottom: 16 }}>
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
                        {running ? '执行中...' : '执行'}
                    </Button>
                </Space>
            </Card>
            <Card title={`服务器列表 (${servers.length})`}>
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
            {running && <div style={{ padding: 24 }}><Spin tip="正在执行…" /></div>}
            {!running && results.length > 0 && (
                <Card title={`结果 (${total})`} style={{ marginTop: 16 }}>
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
