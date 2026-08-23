import { useState } from 'react';
import { Typography, Tabs, Button, Space, Grid, Tag, Message } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useDockerHosts, useDockerEvents } from './hooks';
import { useT } from '@/stores/uiStore';
import { ContainerTab } from './ContainerTab';
import { ImageTab } from './ImageTab';
import { VolumeTab } from './VolumeTab';
import { NetworkTab } from './NetworkTab';
import { ComposeTab } from './ComposeTab';

const { Title, Paragraph } = Typography;

export default function DockerDashboardPage() {
    const t = useT();
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
                    <Title heading={4}>{t('docker.title')}</Title>
                    <Paragraph type="secondary">
                        {t('docker.subtitle')}
                    </Paragraph>
                </div>
                <Space>
                    <Tag color={connected ? 'green' : 'red'} style={{ marginRight: 0 }}>
                        {connected ? t('docker.connected') : t('docker.connecting')}
                    </Tag>
                    <Button icon={<IconRefresh />} onClick={() => refetchHosts()}>
                        {t('docker.refresh')}
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
                        <div style={{ fontSize: 12, color: 'var(--color-text-3)' }}>{t('docker.hosts')}</div>
                        <div style={{ fontSize: 20, fontWeight: 600 }}>
                            {onlineHosts} / {hosts?.length ?? 0} {t('docker.online')}
                        </div>
                    </div>
                </Grid.Col>
            </Grid.Row>

            <Tabs activeTab={activeTab} onChange={setActiveTab}>
                <Tabs.TabPane key="containers" title={t('docker.containers')}>
                    <ContainerTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="images" title={t('docker.images')}>
                    <ImageTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="volumes" title={t('docker.volumes')}>
                    <VolumeTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="networks" title={t('docker.networks')}>
                    <NetworkTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="compose" title={t('docker.compose')}>
                    <ComposeTab />
                </Tabs.TabPane>
            </Tabs>
        </Space>
    );
}