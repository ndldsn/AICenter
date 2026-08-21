import { useState } from 'react';
import { Typography, Tabs, Button, Space, Grid } from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useDockerHosts } from './hooks';
import { ContainerTab } from './ContainerTab';
import { ImageTab } from './ImageTab';
import { VolumeTab } from './VolumeTab';
import { NetworkTab } from './NetworkTab';

const { Title, Paragraph } = Typography;

export default function DockerDashboardPage() {
    const [activeTab, setActiveTab] = useState('containers');
    const { data: hosts, refetch: refetchHosts } = useDockerHosts();

    const onlineHosts = (hosts ?? []).filter((h) => h.status === 'online').length;

    const refresh = () => {
        refetchHosts();
    };

    return (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                    <Title heading={4}>Docker</Title>
                    <Paragraph type="secondary">
                        Manage containers, images, volumes, and networks
                    </Paragraph>
                </div>
                <Space>
                    <Button icon={<IconRefresh />} onClick={refresh}>
                        Refresh
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
                        <div style={{ fontSize: 12, color: 'var(--color-text-3)' }}>Hosts</div>
                        <div style={{ fontSize: 20, fontWeight: 600 }}>
                            {onlineHosts} / {hosts?.length ?? 0} online
                        </div>
                    </div>
                </Grid.Col>
            </Grid.Row>

            <Tabs activeTab={activeTab} onChange={setActiveTab}>
                <Tabs.TabPane key="containers" title="Containers">
                    <ContainerTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="images" title="Images">
                    <ImageTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="volumes" title="Volumes">
                    <VolumeTab />
                </Tabs.TabPane>
                <Tabs.TabPane key="networks" title="Networks">
                    <NetworkTab />
                </Tabs.TabPane>
            </Tabs>
        </Space>
    );
}