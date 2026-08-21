import { useState } from 'react';
import { Table, Button, Space, Tag, Drawer, Typography, Popconfirm, Switch, Tooltip } from '@arco-design/web-react';
import {
    IconRefresh,
    IconPlayArrow,
    IconPause,
    IconDelete,
    IconFile,
} from '@arco-design/web-react/icon';
import {
    useContainers,
    useStartContainer,
    useStopContainer,
    useDeleteContainer,
    useContainerLogs,
} from './hooks';
import { DockerContainer, DockerContainerDetail } from '@/services/docker';
import { formatPorts, shortId, formatUnixTime } from './utils';

const { Text } = Typography;

const stateTag: Record<string, { color: string; label: string }> = {
    running: { color: 'green', label: 'RUNNING' },
    exited: { color: 'red', label: 'EXITED' },
    paused: { color: 'orange', label: 'PAUSED' },
    created: { color: 'gray', label: 'CREATED' },
};

export function ContainerTab() {
    const [showAll, setShowAll] = useState(false);
    const [detail, setDetail] = useState<DockerContainerDetail | null>(null);
    const [logsOpen, setLogsOpen] = useState(false);
    const [logsId, setLogsId] = useState('');

    const { data: containers, isLoading, refetch } = useContainers(showAll);
    const startMutation = useStartContainer();
    const stopMutation = useStopContainer();
    const deleteMutation = useDeleteContainer();
    const { data: logs, isFetching: logsLoading } = useContainerLogs(logsId, logsOpen && !!logsId);

    const openLogs = (record: DockerContainer) => {
        setLogsId(record.id);
        setLogsOpen(true);
    };

    const columns = [
        {
            title: 'Name',
            dataIndex: 'name',
            fixed: 'left' as const,
            width: 220,
            render: (name: string, record: DockerContainer) => (
                <Space direction="vertical" size={0}>
                    <Text bold style={{ fontFamily: 'monospace' }}>{name}</Text>
                    <Text type="secondary" style={{ fontSize: 12, fontFamily: 'monospace' }}>
                        {shortId(record.id)}
                    </Text>
                </Space>
            ),
        },
        {
            title: 'Image',
            dataIndex: 'image',
            width: 220,
            render: (image: string) => (
                <Text style={{ fontFamily: 'monospace' }}>{image}</Text>
            ),
        },
        {
            title: 'State',
            dataIndex: 'state',
            width: 110,
            render: (state: string) => {
                const cfg = stateTag[state] || stateTag.created;
                return <Tag color={cfg.color}>{cfg.label}</Tag>;
            },
        },
        {
            title: 'Status',
            dataIndex: 'status',
            width: 160,
            render: (status: string) => (
                <Text style={{ fontSize: 12 }}>{status}</Text>
            ),
        },
        {
            title: 'Ports',
            dataIndex: 'ports',
            width: 200,
            render: (ports: DockerContainer['ports']) => (
                <Text style={{ fontSize: 12, fontFamily: 'monospace' }}>
                    {formatPorts(ports)}
                </Text>
            ),
        },
        {
            title: 'Created',
            dataIndex: 'created',
            width: 150,
            render: (created: number) => (
                <Text style={{ fontSize: 12 }}>{formatUnixTime(created)}</Text>
            ),
        },
        {
            title: 'Actions',
            dataIndex: 'actions',
            width: 220,
            render: (_: unknown, record: DockerContainer) => (
                <Space size={4}>
                    <Tooltip content="Logs">
                        <Button size="small" icon={<IconFile />} onClick={() => openLogs(record)} />
                    </Tooltip>
                    {record.state !== 'running' ? (
                        <Tooltip content="Start">
                            <Button
                                size="small"
                                type="primary"
                                status="success"
                                icon={<IconPlayArrow />}
                                loading={startMutation.isPending}
                                onClick={() => startMutation.mutate(record.id)}
                            />
                        </Tooltip>
                    ) : (
                        <Tooltip content="Stop">
                            <Button
                                size="small"
                                type="primary"
                                status="warning"
                                icon={<IconPause />}
                                loading={stopMutation.isPending}
                                onClick={() => stopMutation.mutate(record.id)}
                            />
                        </Tooltip>
                    )}
                    <Popconfirm
                        title={`Delete container ${record.name}?`}
                        onOk={() => deleteMutation.mutate(record.id)}
                    >
                        <Button
                            size="small"
                            type="outline"
                            status="danger"
                            icon={<IconDelete />}
                        />
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <>
            <Space style={{ marginBottom: 12 }}>
                <Switch
                    checked={showAll}
                    onChange={setShowAll}
                    checkedText="All containers"
                    uncheckedText="Running only"
                />
                <Button size="small" icon={<IconRefresh />} onClick={() => refetch()}>
                    Refresh
                </Button>
            </Space>

            <Table
                rowKey="id"
                loading={isLoading}
                data={containers ?? []}
                columns={columns}
                scroll={{ x: 1280 }}
                pagination={false}
                onRow={(record) => ({
                    onClick: () => setDetail(record as DockerContainerDetail),
                })}
            />

            <Drawer
                width={520}
                title={detail ? `Container: ${detail.name}` : 'Container'}
                visible={logsOpen}
                onCancel={() => setLogsOpen(false)}
                footer={null}
            >
                {detail && (
                    <Space direction="vertical" size={8} style={{ width: '100%', marginBottom: 16 }}>
                        <div>
                            <Text type="secondary" style={{ fontSize: 12 }}>ID: </Text>
                            <Text style={{ fontFamily: 'monospace', fontSize: 12 }}>{detail.id}</Text>
                        </div>
                        <div>
                            <Text type="secondary" style={{ fontSize: 12 }}>Image: </Text>
                            <Text style={{ fontFamily: 'monospace', fontSize: 12 }}>{detail.image}</Text>
                        </div>
                        {detail.env && detail.env.length > 0 && (
                            <div>
                                <Text type="secondary" style={{ fontSize: 12 }}>Env: </Text>
                                <Text style={{ fontFamily: 'monospace', fontSize: 12 }}>
                                    {detail.env.join(', ')}
                                </Text>
                            </div>
                        )}
                    </Space>
                )}
                <Text type="secondary">Logs (tail 200)</Text>
                <pre
                    style={{
                        background: '#1e1e2e',
                        color: '#cdd6f4',
                        padding: 12,
                        borderRadius: 8,
                        fontSize: 12,
                        lineHeight: 1.5,
                        maxHeight: 420,
                        overflow: 'auto',
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-all',
                    }}
                >
                    {logsLoading ? 'Loading logs...' : logs || '(no logs)'}
                </pre>
            </Drawer>
        </>
    );
}