import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Form, Input, Button } from '@arco-design/web-react';
import { authApi } from '@/services/agent';
import { useAuthStore } from '@/stores/authStore';
import { useUIStore } from '@/stores/uiStore';

export default function LoginPage() {
    const [form] = Form.useForm();
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const setTokens = useAuthStore((s) => s.setTokens);
    const setUser = useAuthStore((s) => s.setUser);
    const t = useUIStore((s) => s.t);

    const submit = async (values: { username: string; password: string }) => {
        setError('');
        setLoading(true);
        try {
            const res = await authApi.login(values);
            if (res.access_token) {
                setTokens(res.access_token, res.refresh_token || '');
                if (res.user) setUser(res.user);
                navigate('/');
            } else {
                setError(t('login.error.invalid'));
            }
        } catch (e: any) {
            setError(e?.message || t('login.error.invalid'));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', background: 'var(--color-fill-2)' }}>
            <Card title={t('login.title')} style={{ width: 360 }}>
                <Form form={form} layout="vertical" onSubmit={submit} autoComplete="off">
                    <Form.Item label={t('login.username')} field="username" rules={[{ required: true, message: t('login.error.empty') }]}>
                        <Input placeholder="admin" />
                    </Form.Item>
                    <Form.Item label={t('login.password')} field="password" rules={[{ required: true, message: t('login.error.empty') }]}>
                        <Input.Password placeholder="Admin@123!" />
                    </Form.Item>
                    {error ? <div style={{ color: 'red', marginBottom: 12 }}>{error}</div> : null}
                    <Button type="primary" htmlType="submit" loading={loading} long>
                        {loading ? t('login.loading') : t('login.submit')}
                    </Button>
                </Form>
            </Card>
        </div>
    );
}
