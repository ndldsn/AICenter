import { useState } from 'react';
import { Typography, Tabs, Button, Space, Grid, Tag, Message } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useDockerHosts, useDockerEvents } from './hooks';
import { ContainerTab } from './ContainerTab';
import { ImageTab } from './ImageTab';
import { VolumeTab } from './VolumeTab';
import { NetworkTab } from './NetworkTab';
import { ComposeTab } from './ComposeTab';

const { Title, Paragraph } = Typography;

export default function DockerDashboardPage() {
    const [activeTab, setActiveTab] = useState('containers');
    const { data: hosts, refetch: refetchHosts } = useDockerHosts();
    const { connected } = useDockerEvents((event) => {
        Message.info(`Docker: ${event.type} — ${event.target}`);
    });

    const onlineHosts = (hosts ?? []).filter((h) => h.status === 'online').length;

    return (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>Docker</Title>
                    <Paragraph type="secondary">
                        管理容器、镜像、卷、网络和 Compose 项目
                    </Paragraph>
                </div>
                <Space>
                    <Tag color={connected ? 'green' : 'red'} style={{ marginRight: 0 }}>
                        {connected ? '已连接' : '连接中…'}
                    </Tag>
                    <Button icon={<IconRefresh />} onClick={() => refetchHosts()}>
                        刷新
                    </Button>
                </Space>
            </div>

            <Grid.Row gutter={16}>
                <Grid.Col span={6}>
                    <div
                        style={{
                            background: 'var(--color-fill-2)',
                            borderRadius: 8,
                            padding: '12px 16px',
                        }}
                    >
                        <div style={{ fontSize: 12, color: 'var(--color-text-3)' }}>主机</div>
                        <div style={{ fontSize: 20, fontWeight: 600 }}>
                            {onlineHosts} / {hosts?.length ?? 0} 在线
                        </div>
                    </div>
                </Grid.Col>
            </Grid.Row>

            <Tabs activeTab={activeTab} onChange={setActiveTab}>
                <Tabs.TabPane key="containers" title="容器">
                    <ContainerTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="images" title="镜像">
                    <ImageTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="volumes" title="卷">
                    <VolumeTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="networks" title="网络">
                    <NetworkTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="compose" title="Compose">
                    <ComposeTab />
                </Tabs.TabPane>
            </Tabs>
        </Space>
    );
}