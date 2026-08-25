import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd(), '');

    return {
        plugins: [react()],
        resolve: {
            alias: {
                '@': path.resolve(__dirname, './src'),
            },
        },
        server: {
            host: '0.0.0.0',
            port: 5173,
            strictPort: true,
            proxy: {
                '/api': {
                    target: env.VITE_API_URL || 'http://127.0.0.1:8081',
                    changeOrigin: true,
                    configure: (proxy) => {
                        proxy.on('proxyReq', (proxyReq) => {
                            proxyReq.setHeader('x-forwarded-proto', 'http');
                        });
                    },
                },
                '/ws': {
                    target: env.VITE_WS_URL || 'ws://127.0.0.1:8081',
                    ws: true,
                    changeOrigin: true,
                },
            },
        },
        build: {
            outDir: 'dist',
            sourcemap: true,
            chunkSizeWarningLimit: 1000,
            rollupOptions: {
                output: {
                    manualChunks(id: string) {
                        if (
                            id.includes('node_modules/react') ||
                            id.includes('node_modules/react-dom') ||
                            id.includes('node_modules/react-router')
                        ) {
                            return 'react-vendor';
                        }
                        if (id.includes('node_modules/@arco-design')) {
                            return 'arco-vendor';
                        }
                        if (id.includes('node_modules/echarts')) {
                            return 'echarts-vendor';
                        }
                    },
                },
            },
        },
    };
});