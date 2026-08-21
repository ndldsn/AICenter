import { useEffect } from 'react';
import {
    Modal,
    Form,
    Input,
    Select,
    Switch,
} from '@arco-design/web-react';
import { useProvider, useCreateProvider, useUpdateProvider } from './hooks';

const PROVIDER_TYPES = [
    { label: 'OpenAI Compatible', value: 'openai-compatible' },
    { label: 'Anthropic', value: 'anthropic' },
    { label: 'Gemini', value: 'gemini' },
    { label: 'Mock (dev)', value: 'mock' },
];

interface Props {
    open: boolean;
    editingId: string | null;
    onClose: () => void;
}

export default function AddProviderModal({ open, editingId, onClose }: Props) {
    const [form] = Form.useForm();
    const createProvider = useCreateProvider();
    const updateProvider = useUpdateProvider();
    const { data: existing } = useProvider(editingId || '');

    useEffect(() => {
        if (open && editingId && existing) {
            form.setFieldsValue({
                name: existing.name,
                display_name: existing.display_name,
                base_url: existing.base_url,
                api_type: existing.api_type,
                is_enabled: existing.is_enabled,
                is_default: existing.is_default,
            });
        } else if (open && !editingId) {
            form.resetFields();
            form.setFieldValue('api_type', 'openai-compatible');
            form.setFieldValue('is_enabled', true);
            form.setFieldValue('is_default', false);
        }
    }, [open, editingId, existing, form]);

    const handleSubmit = async () => {
        const values = await form.validate();
        if (editingId) {
            await updateProvider.mutateAsync({ id: editingId, provider: values });
        } else {
            await createProvider.mutateAsync(values);
        }
        onClose();
    };

    const pending = createProvider.isPending || updateProvider.isPending;

    return (
        <Modal
            title={editingId ? 'Edit Provider' : 'Add Provider'}
            visible={open}
            onOk={handleSubmit}
            onCancel={onClose}
            okText={editingId ? 'Update' : 'Create'}
            confirmLoading={pending}
            style={{ width: 520 }}
        >
            <Form form={form} layout="vertical">
                <Form.Item label="Name" field="name" rules={[{ required: true, message: 'Name is required' }]}>
                    <Input placeholder="openai" />
                </Form.Item>
                <Form.Item label="Display Name" field="display_name" rules={[{ required: true }]}>
                    <Input placeholder="OpenAI" />
                </Form.Item>
                <Form.Item label="Type" field="api_type" rules={[{ required: true }]}>
                    <Select options={PROVIDER_TYPES} />
                </Form.Item>
                <Form.Item label="Base URL" field="base_url" rules={[{ required: true }]}>
                    <Input placeholder="https://api.openai.com/v1" />
                </Form.Item>
                <Form.Item
                    label="API Key"
                    field="api_key_enc"
                    extra="Leave empty to keep existing key when editing"
                >
                    <Input.Password placeholder="sk-..." />
                </Form.Item>
                <Form.Item label="Enabled" field="is_enabled">
                    <Switch />
                </Form.Item>
                <Form.Item label="Default" field="is_default">
                    <Switch />
                </Form.Item>
            </Form>
        </Modal>
    );
}