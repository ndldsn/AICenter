import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Typography, Card, Descriptions, Tabs, Button, Spin, Message, Tag, Space,
} from '@arco-design/web-react';
import { IconArrowLeft, IconRefresh } from '@arco-design/web-react/icon';
import { serverApi, Server } from '@/services/servers';
import { apiPost } from '@/services/api';
import Terminal from './Terminal';

const { Title, Text } = Typography;

export default function ServerDetailPage() {
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
            Message.error('加载服务器失败');
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
            Message.error('打开终端失败');
        } finally {
            setOpening(false);
        }
    };

    if (loading) {
        return <div style={{ padding: 24 }}><Spin /></div>;
    }
    if (!server) {
        return <div style={{ padding: 24 }}><Text>服务器不存在</Text></div>;
    }

    return (
        <div style={{ padding: 16 }}>
            <Space style={{ marginBottom: 12 }}>
                <Button icon={<IconArrowLeft />} onClick={() => navigate('/servers')}>返回</Button>
                <Button icon={<IconRefresh />} onClick={load}>刷新</Button>
                <Button type="primary" loading={opening} onClick={openTerminal}>打开终端</Button>
            </Space>
            <Title heading={4}>{server.name}</Title>
            <Tabs activeTab={activeTab} onChange={setActiveTab}>
                <Tabs.TabPane title="概览" key="overview">
                    <Card>
                        <Descriptions
                            column={2}
                            data={[
                                { label: '主机', value: `${server.host}:${server.port}` },
                                { label: '用户名', value: server.username },
                                { label: '认证方式', value: server.auth_type },
                                { label: '状态', value: <Tag color={server.status === 'online' ? 'green' : 'gray'}>{server.status}</Tag> },
                                { label: 'Agent', value: <Tag color={server.agent_connected ? 'green' : 'gray'}>{server.agent_connected ? '已连接' : '未连接'}</Tag> },
                                { label: '标签', value: (server.tags || []).join(', ') || '-' },
                            ]}
                        />
                        {server.os_info && (
                            <Descriptions
                                column={2} style={{ marginTop: 12 }}
                                data={[
                                    { label: '发行版', value: server.os_info.distribution },
                                    { label: '内核', value: server.os_info.kernel },
                                    { label: '架构', value: server.os_info.architecture },
                                    { label: '主机名', value: server.os_info.hostname },
                                ]}
                            />
                        )}
                    </Card>
                </Tabs.TabPane>
                <Tabs.TabPane title="终端" key="terminal">
                    <Card>
                        {sessionId ? (
                            <div style={{ height: '60vh' }}>
                                <Terminal sessionId={sessionId} />
                            </div>
                        ) : (
                            <Text type="secondary">点击「打开终端」启动一个 PTY 会话。</Text>
                        )}
                    </Card>
                </Tabs.TabPane>
            </Tabs>
        </div>
    );
}
