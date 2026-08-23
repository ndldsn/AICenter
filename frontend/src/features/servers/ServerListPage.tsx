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
import { useT } from '@/stores/uiStore';

const { Title, Paragraph } = Typography;

export default function ServerListPage() {
    const [page, setPage] = useState(1);
    const [limit] = useState(20);
    const [addModalVisible, setAddModalVisible] = useState(false);
    const [editingServer, setEditingServer] = useState<Server | null>(null);
    const navigate = useNavigate();
    const t = useT();

    const { data, isLoading, refetch } = useServers(page, limit);
    const deleteMutation = useDeleteServer();

    const handleDelete = async (id: string) => {
        try {
            await deleteMutation.mutateAsync(id);
        } catch {
            // Error handled by mutation
        }
    };

    const statusConfig: Record<string, { color: string; label: string }> = {
        online: { color: 'green', label: t('servers.status.online') },
        offline: { color: 'red', label: t('servers.status.offline') },
        unknown: { color: 'gray', label: t('servers.status.unknown') },
    };

    const columns = [
        {
            title: t('common.name'),
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
            title: t('common.status'),
            dataIndex: 'status',
            render: (status: string, _record: Server) => {
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
                    {connected ? t('servers.connected') : t('servers.disconnected')}
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
            title: t('servers.lastSeen', '最近在线'),
            dataIndex: 'last_heartbeat',
            render: (heartbeat: string | undefined) => {
                if (!heartbeat) return <span style={{ color: 'var(--color-text-3)' }}>{t('servers.never')}</span>;
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
            title: t('common.actions'),
            dataIndex: 'id',
            render: (id: string, record: Server) => (
                <Space>
                    <Tooltip content={t('servers.testConnection')}>
                        <Button
                            size="small"
                            icon={<IconPlayCircle />}
                            onClick={() => Message.info(t('servers.testing'))}
                        />
                    </Tooltip>
                    <Tooltip content={t('common.edit')}>
                        <Button
                            size="small"
                            icon={<IconEdit />}
                            onClick={() => {
                                setEditingServer(record);
                                setAddModalVisible(true);
                            }}
                        />
                    </Tooltip>
                    <Tooltip content={t('servers.copySSH', '复制 SSH 命令')}>
                        <Button
                            size="small"
                            icon={<IconCopy />}
                            onClick={() => {
                                const cmd = `ssh ${record.username}@${record.host} -p ${record.port}`;
                                navigator.clipboard.writeText(cmd);
                                Message.success(t('servers.copied'));
                            }}
                        />
                    </Tooltip>
                    <Popconfirm
                        title={t('servers.confirmDelete')}
                        onOk={() => handleDelete(id)}
                        okText={t('servers.delete')}
                        cancelText={t('servers.cancel')}
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
                    <Title heading={4}>{t('servers.title')}</Title>
                    <Paragraph type="secondary">
                        {t('servers.subtitle')}
                    </Paragraph>
                </div>
                <Space>
                    <Button icon={<IconRefresh />} onClick={() => refetch()}>
                        {t('common.refresh')}
                    </Button>
                    <Button
                        type="primary"
                        icon={<IconPlus />}
                        onClick={() => {
                            setEditingServer(null);
                            setAddModalVisible(true);
                        }}
                    >
                        {t('servers.add')}
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
                            {t('servers.empty')}
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
