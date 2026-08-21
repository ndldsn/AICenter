import { useState } from 'react';
import { Table, Button, Space, Typography, Modal, Form, Input, Popconfirm, Select, Tag } from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconDelete } from '@arco-design/web-react/icon';
import { useNetworks, useCreateNetwork, useDeleteNetwork } from './hooks';
import { DockerNetwork } from '@/services/docker';
import { shortId, formatUnixTime } from './utils';

const { Text } = Typography;

const DRIVER_OPTIONS = ['bridge', 'host', 'none', 'macvlan', 'overlay'].map((d) => (
    <Select.Option key={d} value={d}>
        {d}
    </Select.Option>
));

export function NetworkTab() {
    const [createOpen, setCreateOpen] = useState(false);
    const [name, setName] = useState('');
    const [driver, setDriver] = useState('bridge');

    const { data: networks, isLoading, refetch } = useNetworks();
    const createMutation = useCreateNetwork();
    const deleteMutation = useDeleteNetwork();

    const handleCreate = async () => {
        if (!name.trim()) return;
        try {
            await createMutation.mutateAsync({ name: name.trim(), driver });
            setCreateOpen(false);
            setName('');
            setDriver('bridge');
        } catch {
            // Error handled by mutation
        }
    };

    const columns = [
        {
            title: 'Name',
            dataIndex: 'name',
            fixed: 'left' as const,
            width: 220,
            render: (n: string, record: DockerNetwork) => (
                <Space direction="vertical" size={0}>
                    <Text bold style={{ fontFamily: 'monospace' }}>{n}</Text>
                    <Text type="secondary" style={{ fontSize: 12, fontFamily: 'monospace' }}>
                        {shortId(record.id)}
                    </Text>
                </Space>
            ),
        },
        {
            title: 'Driver',
            dataIndex: 'driver',
            width: 100,
            render: (d: string) => <Tag>{d}</Tag>,
        },
        {
            title: 'Scope',
            dataIndex: 'scope',
            width: 100,
            render: (s: string) => <Text style={{ fontSize: 12 }}>{s}</Text>,
        },
        {
            title: 'Containers',
            dataIndex: 'containers',
            width: 110,
            render: (count: number) => <Text>{count}</Text>,
        },
        {
            title: 'Flags',
            dataIndex: 'internal',
            width: 160,
            render: (_: unknown, record: DockerNetwork) => (
                <Space size={4}>
                    {record.internal && <Tag color="orange">internal</Tag>}
                    {record.ipv6 && <Tag color="blue">ipv6</Tag>}
                    {record.attachable && <Tag color="cyan">attachable</Tag>}
                    {!record.internal && !record.ipv6 && !record.attachable && (
                        <Text type="secondary" style={{ fontSize: 12 }}>-</Text>
                    )}
                </Space>
            ),
        },
        {
            title: 'Created',
            dataIndex: 'created',
            width: 160,
            render: (created: number) => <Text style={{ fontSize: 12 }}>{formatUnixTime(created)}</Text>,
        },
        {
            title: 'Actions',
            dataIndex: 'actions',
            width: 100,
            render: (_: unknown, record: DockerNetwork) => (
                <Popconfirm
                    title={`Delete network ${record.name}?`}
                    onOk={() => deleteMutation.mutate(record.id)}
                >
                    <Button
                        size="small"
                        type="outline"
                        status="danger"
                        icon={<IconDelete />}
                    />
                </Popconfirm>
            ),
        },
    ];

    return (
        <>
            <Space style={{ marginBottom: 12 }}>
                <Button
                    size="small"
                    type="primary"
                    icon={<IconPlus />}
                    onClick={() => setCreateOpen(true)}
                >
                    Create network
                </Button>
                <Button size="small" icon={<IconRefresh />} onClick={() => refetch()}>
                    Refresh
                </Button>
            </Space>

            <Table
                rowKey="id"
                loading={isLoading}
                data={networks ?? []}
                columns={columns}
                scroll={{ x: 1000 }}
                pagination={false}
            />

            <Modal
                title="Create network"
                visible={createOpen}
                onCancel={() => setCreateOpen(false)}
                onOk={handleCreate}
                confirmLoading={createMutation.isPending}
                okText="Create"
            >
                <Form layout="vertical" style={{ marginTop: 12 }}>
                    <Form.Item label="Name" required>
                        <Input
                            placeholder="e.g. aicenter_backend"
                            value={name}
                            onChange={setName}
                            onPressEnter={handleCreate}
                        />
                    </Form.Item>
                    <Form.Item label="Driver">
                        <Select value={driver} onChange={setDriver}>
                            {DRIVER_OPTIONS}
                        </Select>
                    </Form.Item>
                </Form>
            </Modal>
        </>
    );
}