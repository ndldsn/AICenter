import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Typography, Card, Descriptions, Tabs, Button, Spin, Message, Tag, Space,
} from '@arco-design/web-react';
import { IconArrowLeft, IconRefresh } from '@arco-design/web-react/icon';
import { serverApi, Server } from '@/services/servers';
import { apiPost } from '@/services/api';
import { useT } from '@/stores/uiStore';
import Terminal from './Terminal';

const { Title, Text } = Typography;

export default function ServerDetailPage() {
    const t = useT();
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [server, setServer] = useState<Server | null>(null);
    const [loading, setLoading] = useState(true);
    const [sessionId, setSessionId] = useState<string | null>(null);
    const [opening, setOpening] = useState(false);
    const [activeTab, setActiveTab] = useState('overview');

    const load = async () => {
        if (!id) return;
        setLoading(true);
        try {
            const r = await serverApi.get(id);
            setServer(r.data);
        } catch {
            Message.error(t('server.detail.loading'));
        } finally {
            setLoading(false);
        }
    };
    useEffect(() => { load(); }, [id]);

    const openTerminal = async () => {
        setOpening(true);
        try {
            const r = await apiPost<{ code: number; data: { session_id: string } }>(
                '/terminal/sessions', { server_id: id, cols: 80, rows: 24 }
            );
            setSessionId(r.data.session_id);
            setActiveTab('terminal');
        } catch {
            Message.error(t('server.detail.openTerminalFailed'));
        } finally {
            setOpening(false);
        }
    };

    if (loading) {
        return <div style={{ padding: 24 }}><Spin /></div>;
    }
    if (!server) {
        return <div style={{ padding: 24 }}><Text>{t('server.detail.notFound')}</Text></div>;
    }

    return (
        <div style={{ padding: 16 }}>
            <Space style={{ marginBottom: 12 }}>
                <Button icon={<IconArrowLeft />} onClick={() => navigate('/servers')}>{t('server.detail.back')}</Button>
                <Button icon={<IconRefresh />} onClick={load}>{t('server.detail.refresh')}</Button>
                <Button type="primary" loading={opening} onClick={openTerminal}>{t('server.detail.openTerminal')}</Button>
            </Space>
            <Title heading={4}>{server.name}</Title>
            <Tabs activeTab={activeTab} onChange={setActiveTab}>
                <Tabs.TabPane title={t('server.detail.overview')} key="overview">
                    <Card>
                        <Descriptions
                            column={2}
                            data={[
                                { label: t('server.detail.host'), value: `${server.host}:${server.port}` },
                                { label: t('server.detail.username'), value: server.username },
                                { label: t('server.detail.authType'), value: server.auth_type },
                                { label: t('server.detail.status'), value: <Tag color={server.status === 'online' ? 'green' : 'gray'}>{server.status}</Tag> },
                                { label: t('server.detail.agent'), value: <Tag color={server.agent_connected ? 'green' : 'gray'}>{server.agent_connected ? t('server.detail.connected') : t('server.detail.disconnected')}</Tag> },
                                { label: t('server.detail.tags'), value: (server.tags || []).join(', ') || '-' },
                            ]}
                        />
                        {server.os_info && (
                            <Descriptions
                                column={2} style={{ marginTop: 12 }}
                                data={[
                                    { label: t('server.detail.distribution'), value: server.os_info.distribution },
                                    { label: t('server.detail.kernel'), value: server.os_info.kernel },
                                    { label: t('server.detail.architecture'), value: server.os_info.architecture },
                                    { label: t('server.detail.hostname'), value: server.os_info.hostname },
                                ]}
                            />
                        )}
                    </Card>
                </Tabs.TabPane>
                <Tabs.TabPane title={t('server.detail.terminal')} key="terminal">
                    <Card>
                        {sessionId ? (
                            <div style={{ height: '60vh' }}>
                                <Terminal sessionId={sessionId} />
                            </div>
                        ) : (
                            <Text type="secondary">{t('server.detail.startPtyHint')}</Text>
                        )}
                    </Card>
                </Tabs.TabPane>
            </Tabs>
        </div>
    );
}
