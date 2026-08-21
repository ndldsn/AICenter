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
                    target: env.VITE_API_URL || 'http://localhost:8080',
                    changeOrigin: true,
                },
                '/ws': {
                    target: env.VITE_WS_URL || 'ws://localhost:8080',
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
                    manualChunks: {
                        'react-vendor': ['react', 'react-dom', 'react-router-dom'],
                        'arco-vendor': ['@arco-design/web-react'],
                        'echarts-vendor': ['echarts', 'echarts-for-react'],
                    },
                },
            },
        },
    };
});
