import { useState } from 'react';
import { Table, Button, Space, Typography, Modal, Form, Input, Popconfirm, Select } from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconDelete } from '@arco-design/web-react/icon';
import { useVolumes, useCreateVolume, useDeleteVolume } from './hooks';
import { DockerVolume } from '@/services/docker';
import { formatUnixTime } from './utils';

const { Text } = Typography;

const DRIVER_OPTIONS = ['local', 'nfs'].map((d) => (
    <Select.Option key={d} value={d}>
        {d}
    </Select.Option>
));

export function VolumeTab() {
    const [createOpen, setCreateOpen] = useState(false);
    const [name, setName] = useState('');
    const [driver, setDriver] = useState('local');

    const { data: volumes, isLoading, refetch } = useVolumes();
    const createMutation = useCreateVolume();
    const deleteMutation = useDeleteVolume();

    const handleCreate = async () => {
        if (!name.trim()) return;
        try {
            await createMutation.mutateAsync({ name: name.trim(), driver });
            setCreateOpen(false);
            setName('');
            setDriver('local');
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
            render: (n: string) => <Text bold style={{ fontFamily: 'monospace' }}>{n}</Text>,
        },
        {
            title: 'Driver',
            dataIndex: 'driver',
            width: 100,
            render: (d: string) => <Text style={{ fontFamily: 'monospace' }}>{d}</Text>,
        },
        {
            title: 'Mountpoint',
            dataIndex: 'mountpoint',
            render: (m: string) => (
                <Text style={{ fontFamily: 'monospace', fontSize: 12 }}>{m}</Text>
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
            render: (_: unknown, record: DockerVolume) => (
                <Popconfirm
                    title={`Delete volume ${record.name}?`}
                    onOk={() => deleteMutation.mutate(record.name)}
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
                    Create volume
                </Button>
                <Button size="small" icon={<IconRefresh />} onClick={() => refetch()}>
                    Refresh
                </Button>
            </Space>

            <Table
                rowKey="name"
                loading={isLoading}
                data={volumes ?? []}
                columns={columns}
                scroll={{ x: 800 }}
                pagination={false}
            />

            <Modal
                title="Create volume"
                visible={createOpen}
                onCancel={() => setCreateOpen(false)}
                onOk={handleCreate}
                confirmLoading={createMutation.isPending}
                okText="Create"
            >
                <Form layout="vertical" style={{ marginTop: 12 }}>
                    <Form.Item label="Name" required>
                        <Input
                            placeholder="e.g. pgdata, logs, backup_data"
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