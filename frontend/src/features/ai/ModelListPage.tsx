import { useState } from 'react';
import {
    Table,
    Button,
    Space,
    Typography,
    Tag,
    Card,
    Popconfirm,
    Empty,
    Spin,
} from '@arco-design/web-react';
import { IconPlus, IconDelete, IconEdit, IconRefresh } from '@arco-design/web-react/icon';
import { useProviders, useDeleteProvider, useProviderModels } from './hooks';
import AddProviderModal from './AddProviderModal';
import ChatDrawer from './ChatDrawer';
import { useT } from '@/stores/uiStore';

const { Title, Paragraph } = Typography;

export default function ModelListPage() {
    const t = useT();
    const { data: providers, isLoading, refetch } = useProviders();
    const deleteProvider = useDeleteProvider();
    const [modalOpen, setModalOpen] = useState(false);
    const [editing, setEditing] = useState<string | null>(null);
    const [expanded, setExpanded] = useState<string | null>(null);
    const [chatProvider, setChatProvider] = useState<{ id: string; name: string } | null>(null);

    const columns = [
        {
            title: t('ai.column.name'),
            dataIndex: 'display_name',
            render: (_: any, p: any) => (
                <Space>
                    <Typography.Text bold>{p.display_name}</Typography.Text>
                    <Tag size="small">{p.api_type}</Tag>
                </Space>
            ),
        },
        {
            title: t('ai.column.address'),
            dataIndex: 'base_url',
            render: (v: string) => <Typography.Text code>{v}</Typography.Text>,
        },
        {
            title: t('ai.column.apiKey'),
            dataIndex: 'api_key_hint',
            render: (v: string) => (v ? <Tag color="green">{v}</Tag> : <Tag>{t('ai.notSet')}</Tag>),
        },
        {
            title: t('ai.column.enabled'),
            dataIndex: 'is_enabled',
            render: (v: boolean) => (v ? <Tag color="green">{t('ai.column.enabled')}</Tag> : <Tag color="gray">{t('common.disabled')}</Tag>),
        },
        {
            title: t('ai.column.actions'),
            render: (_: any, p: any) => (
                <Space>
                    <Button
                        size="mini"
                        type="text"
                        icon={<IconEdit />}
                        onClick={() => {
                            setEditing(p.id);
                            setModalOpen(true);
                        }}
                    >
                        {t('ai.action.edit')}
                    </Button>
                    <Button
                        size="mini"
                        type="text"
                        icon={<IconPlus />}
                        onClick={() => setChatProvider({ id: p.id, name: p.display_name })}
                    >
                        {t('ai.action.chat')}
                    </Button>
                    <Popconfirm
                        title={t('ai.confirmDelete')}
                        content={t('ai.confirmDeleteHint')}
                        onOk={() => deleteProvider.mutate(p.id)}
                    >
                        <Button size="mini" type="text" status="danger" icon={<IconDelete />}>
                            {t('ai.action.delete')}
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    return (
        <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>{t('ai.title')}</Title>
                    <Paragraph type="secondary">
                        {t('ai.subtitle')}
                    </Paragraph>
                </div>
                <Space>
                    <Button icon={<IconRefresh />} onClick={() => refetch()}>
                        {t('ai.refresh')}
                    </Button>
                    <Button
                        type="primary"
                        icon={<IconPlus />}
                        onClick={() => {
                            setEditing(null);
                            setModalOpen(true);
                        }}
                    >
                        {t('ai.addProvider')}
                    </Button>
                </Space>
            </div>

            <Card>
                {isLoading ? (
                    <div style={{ textAlign: 'center', padding: 40 }}>
                        <Spin />
                    </div>
                ) : providers && providers.length > 0 ? (
                    <Table<any>
                        rowKey="id"
                        columns={columns}
                        data={providers}
                        expandedRowRender={(record: any) => <ProviderModels providerId={record.id} />}
                        expandedRowKeys={expanded ? [expanded] : ([] as string[])}
                        onExpand={(expandedRows: string[]) =>
                            setExpanded(expandedRows.length ? expandedRows[expandedRows.length - 1] : null)
                        }
                        pagination={false}
                    />
                ) : (
                    <Empty description={t('ai.empty')} />
                )}
            </Card>

            <AddProviderModal
                open={modalOpen}
                editingId={editing}
                onClose={() => {
                    setModalOpen(false);
                    setEditing(null);
                }}
            />

            <ChatDrawer provider={chatProvider} onClose={() => setChatProvider(null)} />
        </Space>
    );
}

function ProviderModels({ providerId }: { providerId: string }) {
    const { data: models, isLoading } = useProviderModels(providerId);
    if (isLoading) return <Spin />;
    if (!models || models.length === 0) return <Empty description="No models" />;
    return (
        <Space wrap>
            {models.map((m: any) => (
                <Tag key={m.id} color={m.is_default ? 'arcoblue' : 'gray'}>
                    {m.name} ({m.model_id})
                </Tag>
            ))}
        </Space>
    );
}
