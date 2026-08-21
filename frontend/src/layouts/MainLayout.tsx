import { Suspense } from 'react';
import { Spin } from '@arco-design/web-react';
import { Sidebar } from './Sidebar';
import { Navbar } from './Navbar';

interface MainLayoutProps {
    children: React.ReactNode;
}

export function MainLayout({ children }: MainLayoutProps) {
    return (
        <div style={{ display: 'flex', height: '100vh', overflow: 'hidden' }}>
            <Sidebar />
            <div
                style={{
                    flex: 1,
                    display: 'flex',
                    flexDirection: 'column',
                    overflow: 'hidden',
                    transition: 'margin-left 0.2s',
                }}
            >
                <Navbar />
                <main
                    style={{
                        flex: 1,
                        overflow: 'auto',
                        padding: 20,
                        backgroundColor: 'var(--color-bg-1)',
                    }}
                >
                    <Suspense fallback={<Spin style={{ marginTop: 100 }} />}>
                        {children}
                    </Suspense>
                </main>
            </div>
        </div>
    );
}
