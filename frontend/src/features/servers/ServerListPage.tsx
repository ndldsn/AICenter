import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Typography,
    Card,
    Button,
    Table,
    Space,
    Tag,
    Message,
    Popconfirm,
    Tooltip,
    Avatar,
} from '@arco-design/web-react';
import {
    IconPlus,
    IconRefresh,
    IconPlayCircle,
    IconEdit,
    IconDelete,
    IconCopy,
} from '@arco-design/web-react/icon';
import { useServers, useDeleteServer } from './hooks';
import { AddServerModal } from './AddServerModal';
import { Server } from '@/services/servers';

const { Title, Paragraph } = Typography;

export default function ServerListPage() {
    const [page, setPage] = useState(1);
    const [limit] = useState(20);
    const [addModalVisible, setAddModalVisible] = useState(false);
    const [editingServer, setEditingServer] = useState<Server | null>(null);
    const navigate = useNavigate();

    const { data, isLoading, refetch } = useServers(page, limit);
    const deleteMutation = useDeleteServer();

    const handleDelete = async (id: string) => {
        try {
            await deleteMutation.mutateAsync(id);
        } catch {
            // Error handled by mutation
        }
    };

    const columns = [
        {
            title: '名称',
            dataIndex: 'name',
            render: (name: string, record: Server) => (
                <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <Avatar size={32} style={{ backgroundColor: 'rgb(var(--primary-6))' }}>
                        {name.charAt(0).toUpperCase()}
                    </Avatar>
                    <div>
                        <div style={{ fontWeight: 500 }}>{name}</div>
                        <div style={{ fontSize: 12, color: 'var(--color-text-3)' }}>
                            {record.host}
                        </div>
                    </div>
                </div>
            ),
        },
        {
            title: '状态',
            dataIndex: 'status',
            render: (status: string, _record: Server) => {
                const statusConfig: Record<string, { color: string; label: string }> = {
                    online: { color: 'green', label: 'Online' },
                    offline: { color: 'red', label: 'Offline' },
                    unknown: { color: 'gray', label: 'Unknown' },
                };
                const config = statusConfig[status] || statusConfig.unknown;
                return (
                    <Tag color={config.color}>
                        <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                            <span
                                style={{
                                    width: 6,
                                    height: 6,
                                    borderRadius: '50%',
                                    backgroundColor:
                                        config.color === 'green'
                                            ? 'rgb(var(--success-6))'
                                            : config.color === 'red'
                                            ? 'rgb(var(--danger-6))'
                                            : 'var(--color-text-3)',
                                }}
                            />
                            {config.label}
                        </span>
                    </Tag>
                );
            },
        },
        {
            title: 'Agent',
            dataIndex: 'agent_connected',
            render: (connected: boolean) => (
                <Tag color={connected ? 'green' : 'gray'}>
                    {connected ? '已连接' : '未连接'}
                </Tag>
            ),
        },
        {
            title: 'OS',
            dataIndex: 'os_info',
            render: (osInfo: Server['os_info']) => {
                if (!osInfo) return <span style={{ color: 'var(--color-text-3)' }}>-</span>;
                return (
                    <span>
                        {osInfo.distribution} {osInfo.architecture}
                    </span>
                );
            },
        },
        {
            title: 'CPU / Memory',
            dataIndex: 'hardware_info',
            render: (hwInfo: Server['hardware_info']) => {
                if (!hwInfo) return <span style={{ color: 'var(--color-text-3)' }}>-</span>;
                return (
                    <span>
                        {hwInfo.cpu_cores} cores / {hwInfo.memory_gb.toFixed(1)} GB
                    </span>
                );
            },
        },
        {
            title: '最近在线',
            dataIndex: 'last_heartbeat',
            render: (heartbeat: string | undefined) => {
                if (!heartbeat) return <span style={{ color: 'var(--color-text-3)' }}>从未</span>;
                return (
                    <span>
                        {new Date(heartbeat).toLocaleString('zh-CN', {
                            month: '2-digit',
                            day: '2-digit',
                            hour: '2-digit',
                            minute: '2-digit',
                        })}
                    </span>
                );
            },
        },
        {
            title: '操作',
            dataIndex: 'id',
            render: (id: string, record: Server) => (
                <Space>
                    <Tooltip content="测试连接">
                        <Button
                            size="small"
                            icon={<IconPlayCircle />}
                            onClick={() => Message.info('正在测试连接...')}
                        />
                    </Tooltip>
                    <Tooltip content="编辑">
                        <Button
                            size="small"
                            icon={<IconEdit />}
                            onClick={() => {
                                setEditingServer(record);
                                setAddModalVisible(true);
                            }}
                        />
                    </Tooltip>
                    <Tooltip content="复制 SSH 命令">
                        <Button
                            size="small"
                            icon={<IconCopy />}
                            onClick={() => {
                                const cmd = `ssh ${record.username}@${record.host} -p ${record.port}`;
                                navigator.clipboard.writeText(cmd);
                                Message.success('SSH 命令已复制');
                            }}
                        />
                    </Tooltip>
                    <Popconfirm
                        title="确定删除该服务器？"
                        onOk={() => handleDelete(id)}
                        okText="删除"
                        cancelText="取消"
                    >
                        <Button size="small" icon={<IconDelete />} status="danger" />
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>服务器</Title>
                    <Paragraph type="secondary">
                        管理你的 Linux 服务器并监控其状态
                    </Paragraph>
                </div>
                <Space>
                    <Button icon={<IconRefresh />} onClick={() => refetch()}>
                        刷新
                    </Button>
                    <Button
                        type="primary"
                        icon={<IconPlus />}
                        onClick={() => {
                            setEditingServer(null);
                            setAddModalVisible(true);
                        }}
                    >
                        添加服务器
                    </Button>
                </Space>
            </div>

            <Card>
                <Table
                    columns={columns}
                    data={data?.items || []}
                    loading={isLoading}
                    pagination={{
                        total: data?.total || 0,
                        current: page,
                        pageSize: limit,
                        onChange: (p) => setPage(p),
                        showTotal: (total) => `共 ${total} 台服务器`,
                        showJumper: true,
                        sizeCanChange: true,
                    }}
                    noDataElement={
                        <div style={{ padding: 40, textAlign: 'center', color: 'var(--color-text-3)' }}>
                            暂无服务器，点击“添加服务器”开始。
                        </div>
                    }
                    rowKey="id"
                    onRow={(record) => ({
                        style: { cursor: 'pointer' },
                        onClick: () => navigate(`/servers/${record.id}`),
                    })}
                />
            </Card>

            <AddServerModal
                visible={addModalVisible}
                onClose={() => {
                    setAddModalVisible(false);
                    setEditingServer(null);
                }}
                server={editingServer}
                onSuccess={() => {
                    setAddModalVisible(false);
                    setEditingServer(null);
                    refetch();
                }}
            />
        </Space>
    );
}
