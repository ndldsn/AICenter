import { useState } from 'react';
import { Card, Form, Input, Button } from '@arco-design/web-react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '@/services/agent';
import { useAuthStore } from '@/stores/authStore';

export default function LoginPage() {
    const [form] = Form.useForm();
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const setTokens = useAuthStore((s) => s.setTokens);
    const setUser = useAuthStore((s) => s.setUser);

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
                setError('登录失败：未返回 token');
            }
        } catch (e: any) {
            setError(e?.message || '登录失败');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', background: 'var(--color-fill-2)' }}>
            <Card title="AICenter 登录" style={{ width: 360 }}>
                <Form form={form} layout="vertical" onSubmit={submit} autoComplete="off">
                    <Form.Item label="用户名" field="username" rules={[{ required: true, message: '请输入用户名' }]}>
                        <Input placeholder="admin" />
                    </Form.Item>
                    <Form.Item label="密码" field="password" rules={[{ required: true, message: '请输入密码' }]}>
                        <Input.Password placeholder="Admin@123!" />
                    </Form.Item>
                    {error ? <div style={{ color: 'red', marginBottom: 12 }}>{error}</div> : null}
                    <Button type="primary" htmlType="submit" loading={loading} long>
                        登录
                    </Button>
                </Form>
            </Card>
        </div>
    );
}
