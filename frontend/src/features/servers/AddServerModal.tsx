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
                    {isEdit ? '编辑服务器' : '添加服务器'}
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
                <TabPane key="basic" title="基本信息">
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
                            label="服务器名称"
                            field="name"
                            rules={[{ required: true, message: '请输入服务器名称' }]}
                        >
                            <Input placeholder="例如：生产 Web 服务器" />
                        </Item>

                        <div style={{ display: 'flex', gap: 16 }}>
                            <Item
                                label="主机 / IP"
                                field="host"
                                style={{ flex: 1 }}
                                rules={[{ required: true, message: '请输入主机地址' }]}
                            >
                                <Input placeholder="192.168.1.100 或 server.example.com" />
                            </Item>
                            <Item label="端口" field="port" style={{ width: 100 }}>
                                <InputNumber min={1} max={65535} placeholder="22" />
                            </Item>
                        </div>

                        <Item
                            label="用户名"
                            field="username"
                            rules={[{ required: true, message: '请输入用户名' }]}
                        >
                            <Input placeholder="root" />
                        </Item>
                    </Form>
                </TabPane>

                <TabPane key="auth" title="认证">
                    <Form form={form} layout="vertical">
                        <Item label="认证方式" field="auth_type">
                            <Select placeholder="选择认证方式">
                                <Select.Option value="password">密码</Select.Option>
                                <Select.Option value="key">SSH 密钥</Select.Option>
                            </Select>
                        </Item>

                        {authType === 'key' ? (
                            <Item label="私钥" field="private_key">
                                <Input.TextArea
                                    placeholder="粘贴你的 SSH 私钥..."
                                    rows={6}
                                    style={{ fontFamily: 'monospace' }}
                                />
                            </Item>
                        ) : (
                            <Item label="密码" field="password">
                                <Input.Password placeholder="请输入密码" />
                            </Item>
                        )}
                    </Form>
                </TabPane>

                <TabPane key="test" title="测试连接">
                    <div style={{ textAlign: 'center', padding: 20 }}>
                        <Button
                            type="primary"
                            icon={<IconLink />}
                            onClick={handleTestConnection}
                            loading={testing}
                            size="large"
                        >
                            {testing ? '正在测试...' : '测试连接'}
                        </Button>

                        {testResult && (
                            <div style={{ marginTop: 20, textAlign: 'left' }}>
                                <Descriptions
                                    column={2}
                                    title="连接测试结果"
                                    data={[
                                        {
                                            label: '状态',
                                            value: (
                                                <Tag color={testResult.success ? 'green' : 'red'}>
                                                    {testResult.success ? '成功' : '失败'}
                                                </Tag>
                                            ),
                                        },
                                        { label: '信息', value: testResult.message },
                                        { label: 'SSH Banner', value: testResult.ssh_banner || '-' },
                                        {
                                            label: '时间',
                                            value: new Date(testResult.timestamp).toLocaleString(),
                                        },
                                    ]}
                                    style={{ marginTop: 16 }}
                                    labelStyle={{ fontWeight: 600 }}
                                />

                                {testResult.system_info && (
                                    <Descriptions
                                        column={2}
                                        title="系统信息"
                                        data={[
                                            { label: '系统', value: testResult.system_info.os },
                                            {
                                                label: '主机名',
                                                value: testResult.system_info.hostname,
                                            },
                                            {
                                                label: '内核',
                                                value: testResult.system_info.kernel,
                                            },
                                            {
                                                label: 'CPU 核心',
                                                value: testResult.system_info.cpu_cores,
                                            },
                                            {
                                                label: '内存 (GB)',
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
