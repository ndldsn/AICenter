import { useState } from 'react';
import {
    Modal,
    Form,
    Input,
    Select,
    InputNumber,
    Button,
    Message,
    Tabs,
    Descriptions,
    Tag,
} from '@arco-design/web-react';
import { IconDesktop, IconLink } from '@arco-design/web-react/icon';
import { useCreateServer, useTestNewConnection } from './hooks';
import { ConnectionTestResult, CreateServerRequest } from '@/services/servers';

const { Item } = Form;
const { TabPane } = Tabs;

interface AddServerModalProps {
    visible: boolean;
    onClose: () => void;
    server?: any;
    onSuccess: () => void;
}

export function AddServerModal({ visible, onClose, server, onSuccess }: AddServerModalProps) {
    const [form] = Form.useForm();
    const [testResult, setTestResult] = useState<ConnectionTestResult | null>(null);
    const [testing, setTesting] = useState(false);
    const authType = Form.useWatch('auth_type', form) || 'password';

    const createMutation = useCreateServer();
    const testNewConnection = useTestNewConnection();

    const isEdit = !!server;

    const handleTestConnection = async () => {
        try {
            const values = form.getFieldsValue();
            setTesting(true);

            const result = await testNewConnection.mutateAsync({
                name: values.name || 'Test Server',
                host: values.host,
                port: values.port || 22,
                username: values.username || 'root',
                auth_type: values.auth_type || 'password',
                password: values.password,
                private_key: values.private_key,
            });

            setTestResult(result.data);
            if (result.data?.success) {
                Message.success('Connection test successful');
            } else {
                Message.warning('Connection test failed');
            }
        } catch (error: any) {
            Message.error(error?.response?.data?.message || 'Connection test failed');
        } finally {
            setTesting(false);
        }
    };

    const handleSubmit = async () => {
        try {
            const values = form.getFieldsValue();
            await createMutation.mutateAsync(values as CreateServerRequest);
            onSuccess();
        } catch (error: any) {
            Message.error(error?.response?.data?.message || 'Failed to create server');
        }
    };

    return (
        <Modal
            title={
                <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <IconDesktop />
                    {isEdit ? 'Edit Server' : 'Add New Server'}
                </span>
            }
            visible={visible}
            onCancel={onClose}
            onOk={handleSubmit}
            confirmLoading={createMutation.isPending}
            style={{ width: 700 }}
            maskClosable={false}
        >
            <Tabs defaultActiveTab="basic">
                <TabPane key="basic" title="Basic Info">
                    <Form
                        form={form}
                        layout="vertical"
                        initialValues={{
                            port: 22,
                            username: 'root',
                            auth_type: 'password',
                            ...server,
                        }}
                    >
                        <Item
                            label="Server Name"
                            field="name"
                            rules={[{ required: true, message: 'Please enter server name' }]}
                        >
                            <Input placeholder="e.g., Production Web Server" />
                        </Item>

                        <div style={{ display: 'flex', gap: 16 }}>
                            <Item
                                label="Host / IP"
                                field="host"
                                style={{ flex: 1 }}
                                rules={[{ required: true, message: 'Please enter host' }]}
                            >
                                <Input placeholder="192.168.1.100 or server.example.com" />
                            </Item>
                            <Item label="Port" field="port" style={{ width: 100 }}>
                                <InputNumber min={1} max={65535} placeholder="22" />
                            </Item>
                        </div>

                        <Item
                            label="Username"
                            field="username"
                            rules={[{ required: true, message: 'Please enter username' }]}
                        >
                            <Input placeholder="root" />
                        </Item>
                    </Form>
                </TabPane>

                <TabPane key="auth" title="Authentication">
                    <Form form={form} layout="vertical">
                        <Item label="Auth Type" field="auth_type">
                            <Select placeholder="Select auth type">
                                <Select.Option value="password">Password</Select.Option>
                                <Select.Option value="key">SSH Key</Select.Option>
                            </Select>
                        </Item>

                        {authType === 'key' ? (
                            <Item label="Private Key" field="private_key">
                                <Input.TextArea
                                    placeholder="Paste your SSH private key here..."
                                    rows={6}
                                    style={{ fontFamily: 'monospace' }}
                                />
                            </Item>
                        ) : (
                            <Item label="Password" field="password">
                                <Input.Password placeholder="Enter password" />
                            </Item>
                        )}
                    </Form>
                </TabPane>

                <TabPane key="test" title="Test Connection">
                    <div style={{ textAlign: 'center', padding: 20 }}>
                        <Button
                            type="primary"
                            icon={<IconLink />}
                            onClick={handleTestConnection}
                            loading={testing}
                            size="large"
                        >
                            {testing ? 'Testing...' : 'Test Connection'}
                        </Button>

                        {testResult && (
                            <div style={{ marginTop: 20, textAlign: 'left' }}>
                                <Descriptions
                                    column={2}
                                    title="Connection Test Result"
                                    data={[
                                        {
                                            label: 'Status',
                                            value: (
                                                <Tag color={testResult.success ? 'green' : 'red'}>
                                                    {testResult.success ? 'Success' : 'Failed'}
                                                </Tag>
                                            ),
                                        },
                                        { label: 'Message', value: testResult.message },
                                        { label: 'SSH Banner', value: testResult.ssh_banner || '-' },
                                        {
                                            label: 'Timestamp',
                                            value: new Date(testResult.timestamp).toLocaleString(),
                                        },
                                    ]}
                                    style={{ marginTop: 16 }}
                                    labelStyle={{ fontWeight: 600 }}
                                />

                                {testResult.system_info && (
                                    <Descriptions
                                        column={2}
                                        title="System Info"
                                        data={[
                                            { label: 'OS', value: testResult.system_info.os },
                                            {
                                                label: 'Hostname',
                                                value: testResult.system_info.hostname,
                                            },
                                            {
                                                label: 'Kernel',
                                                value: testResult.system_info.kernel,
                                            },
                                            {
                                                label: 'CPU Cores',
                                                value: testResult.system_info.cpu_cores,
                                            },
                                            {
                                                label: 'Memory (GB)',
                                                value: testResult.system_info.memory_gb.toFixed(1),
                                            },
                                        ]}
                                        style={{ marginTop: 16 }}
                                        labelStyle={{ fontWeight: 600 }}
                                    />
                                )}
                            </div>
                        )}
                    </div>
                </TabPane>
            </Tabs>
        </Modal>
    );
}
