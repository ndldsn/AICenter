import { useState } from 'react';
import { Table, Button, Space, Typography, Modal, Form, Input, Popconfirm } from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconDelete } from '@arco-design/web-react/icon';
import { useImages, usePullImage, useDeleteImage } from './hooks';
import { DockerImage } from '@/services/docker';
import { formatBytes, shortId, formatUnixTime } from './utils';

const { Text } = Typography;

export function ImageTab() {
    const [pullOpen, setPullOpen] = useState(false);
    const [repository, setRepository] = useState('');
    const [tag, setTag] = useState('latest');

    const { data: images, isLoading, refetch } = useImages();
    const pullMutation = usePullImage();
    const deleteMutation = useDeleteImage();

    const handlePull = async () => {
        if (!repository.trim()) return;
        try {
            await pullMutation.mutateAsync({ repository: repository.trim(), tag: tag.trim() || 'latest' });
            setPullOpen(false);
            setRepository('');
            setTag('latest');
        } catch {
            // Error handled by mutation
        }
    };

    const columns = [
        {
            title: 'Repository',
            dataIndex: 'repository',
            fixed: 'left' as const,
            width: 260,
            render: (repo: string, record: DockerImage) => (
                <Space direction="vertical" size={0}>
                    <Text bold style={{ fontFamily: 'monospace' }}>
                        {repo}:{record.tag}
                    </Text>
                    <Text type="secondary" style={{ fontSize: 12, fontFamily: 'monospace' }}>
                        {shortId(record.id)}
                    </Text>
                </Space>
            ),
        },
        {
            title: 'Size',
            dataIndex: 'size',
            width: 120,
            render: (size: number) => <Text style={{ fontFamily: 'monospace' }}>{formatBytes(size)}</Text>,
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
            render: (_: unknown, record: DockerImage) => (
                <Popconfirm
                    title={`Delete image ${record.repository}:${record.tag}?`}
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
                    onClick={() => setPullOpen(true)}
                >
                    Pull image
                </Button>
                <Button size="small" icon={<IconRefresh />} onClick={() => refetch()}>
                    Refresh
                </Button>
            </Space>

            <Table
                rowKey="id"
                loading={isLoading}
                data={images ?? []}
                columns={columns}
                scroll={{ x: 700 }}
                pagination={false}
            />

            <Modal
                title="Pull image"
                visible={pullOpen}
                onCancel={() => setPullOpen(false)}
                onOk={handlePull}
                confirmLoading={pullMutation.isPending}
                okText="Pull"
            >
                <Form layout="vertical" style={{ marginTop: 12 }}>
                    <Form.Item label="Repository" required>
                        <Input
                            placeholder="e.g. nginx, postgres, ghcr.io/org/app"
                            value={repository}
                            onChange={setRepository}
                            onPressEnter={handlePull}
                        />
                    </Form.Item>
                    <Form.Item label="Tag">
                        <Input
                            placeholder="latest"
                            value={tag}
                            onChange={setTag}
                            onPressEnter={handlePull}
                        />
                    </Form.Item>
                </Form>
            </Modal>
        </>
    );
}