import { useState } from 'react';
import {
    Table,
    Button,
    Space,
    Typography,
    Modal,
    Form,
    Input,
    Popconfirm,
    Tag,
    Drawer,
} from '@arco-design/web-react';
import { IconPlus, IconRefresh, IconDelete, IconEdit, IconPlayArrow, IconPause } from '@arco-design/web-react/icon';
import {
    useComposeProjects,
    useCreateCompose,
    useUpdateCompose,
    useDeleteCompose,
    useDeployCompose,
    useDownCompose,
} from './hooks';
import { ComposeProject } from '@/services/docker';
import { YamlEditor } from '@/components/editors/YamlEditor';
import { formatUnixTime } from './utils';

const { Text } = Typography;

const DEFAULT_COMPOSE = `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
`;

const statusTag: Record<string, { color: string; label: string }> = {
    running: { color: 'green', label: 'RUNNING' },
    stopped: { color: 'gray', label: 'STOPPED' },
    error: { color: 'red', label: 'ERROR' },
};

export function ComposeTab() {
    const [editorOpen, setEditorOpen] = useState(false);
    const [editing, setEditing] = useState<ComposeProject | null>(null);
    const [name, setName] = useState('');
    const [content, setContent] = useState(DEFAULT_COMPOSE);
    const [previewId, setPreviewId] = useState<string | null>(null);

    const { data: projects, isLoading, refetch } = useComposeProjects();
    const createMutation = useCreateCompose();
    const updateMutation = useUpdateCompose();
    const deleteMutation = useDeleteCompose();
    const deployMutation = useDeployCompose();
    const downMutation = useDownCompose();

    const preview = projects?.find((p) => p.id === previewId) ?? null;

    const openCreate = () => {
        setEditing(null);
        setName('');
        setContent(DEFAULT_COMPOSE);
        setEditorOpen(true);
    };

    const openEdit = (p: ComposeProject) => {
        setEditing(p);
        setName(p.name);
        setContent(p.content);
        setEditorOpen(true);
    };

    const handleSave = async () => {
        try {
            if (editing) {
                await updateMutation.mutateAsync({ id: editing.id, content });
            } else {
                await createMutation.mutateAsync({ name: name.trim(), content });
            }
            setEditorOpen(false);
        } catch {
            // Error handled by mutation
        }
    };

    const columns = [
        {
            title: 'Name',
            dataIndex: 'name',
            fixed: 'left' as const,
            width: 200,
            render: (n: string) => <Text bold style={{ fontFamily: 'monospace' }}>{n}</Text>,
        },
        {
            title: 'Services',
            dataIndex: 'services',
            render: (services: string[]) => (
                <Space size={4}>
                    {(services ?? []).map((s) => (
                        <Tag key={s} color="arcoblue">
                            {s}
                        </Tag>
                    ))}
                    {(!services || services.length === 0) && (
                        <Text type="secondary" style={{ fontSize: 12 }}>-</Text>
                    )}
                </Space>
            ),
        },
        {
            title: 'Status',
            dataIndex: 'status',
            width: 110,
            render: (status: string) => {
                const cfg = statusTag[status] || statusTag.stopped;
                return <Tag color={cfg.color}>{cfg.label}</Tag>;
            },
        },
        {
            title: 'Project Dir',
            dataIndex: 'project_dir',
            width: 230,
            render: (dir: string) => (
                <Text style={{ fontSize: 12, fontFamily: 'monospace' }}>{dir || '-'}</Text>
            ),
        },
        {
            title: 'Updated',
            dataIndex: 'updated',
            width: 160,
            render: (t: number) => <Text style={{ fontSize: 12 }}>{formatUnixTime(t)}</Text>,
        },
        {
            title: 'Actions',
            dataIndex: 'actions',
            width: 230,
            render: (_: unknown, record: ComposeProject) => (
                <Space size={4}>
                    <Button size="small" onClick={() => setPreviewId(record.id)}>
                        View
                    </Button>
                    {record.status !== 'running' ? (
                        <Button
                            size="small"
                            type="primary"
                            status="success"
                            icon={<IconPlayArrow />}
                            loading={deployMutation.isPending}
                            onClick={() => deployMutation.mutate(record.id)}
                        >
                            Deploy
                        </Button>
                    ) : (
                        <Button
                            size="small"
                            type="primary"
                            status="warning"
                            icon={<IconPause />}
                            loading={downMutation.isPending}
                            onClick={() => downMutation.mutate(record.id)}
                        >
                            Down
                        </Button>
                    )}
                    <Button size="small" icon={<IconEdit />} onClick={() => openEdit(record)} />
                    <Popconfirm
                        title={`Delete compose project ${record.name}?`}
                        onOk={() => deleteMutation.mutate(record.id)}
                    >
                        <Button size="small" type="outline" status="danger" icon={<IconDelete />} />
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <>
            <Space style={{ marginBottom: 12 }}>
                <Button size="small" type="primary" icon={<IconPlus />} onClick={openCreate}>
                    New project
                </Button>
                <Button size="small" icon={<IconRefresh />} onClick={() => refetch()}>
                    Refresh
                </Button>
            </Space>

            <Table
                rowKey="id"
                loading={isLoading}
                data={projects ?? []}
                columns={columns}
                scroll={{ x: 1100 }}
                pagination={false}
            />

            <Modal
                title={editing ? `Edit ${editing.name}` : 'New compose project'}
                visible={editorOpen}
                onCancel={() => setEditorOpen(false)}
                onOk={handleSave}
                confirmLoading={createMutation.isPending || updateMutation.isPending}
                okText={editing ? 'Save' : 'Create'}
                style={{ width: 720 }}
            >
                <Form layout="vertical" style={{ marginTop: 12 }}>
                    <Form.Item label="Project name" required>
                        <Input
                            placeholder="e.g. my-stack"
                            value={name}
                            onChange={setName}
                            disabled={!!editing}
                        />
                    </Form.Item>
                    <Form.Item label="docker-compose.yml">
                        <YamlEditor value={content} onChange={setContent} height={360} />
                    </Form.Item>
                </Form>
            </Modal>

            <Drawer
                width={640}
                title={preview ? `compose: ${preview.name}` : 'Compose file'}
                visible={!!preview}
                onCancel={() => setPreviewId(null)}
                footer={null}
            >
                <pre
                    style={{
                        background: '#1e1e2e',
                        color: '#cdd6f4',
                        padding: 12,
                        borderRadius: 8,
                        fontSize: 12,
                        lineHeight: 1.5,
                        maxHeight: 520,
                        overflow: 'auto',
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-all',
                    }}
                >
                    {preview?.content ?? ''}
                </pre>
            </Drawer>
        </>
    );
}