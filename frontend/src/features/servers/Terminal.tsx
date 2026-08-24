import { useEffect, useRef } from 'react';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { Message } from '@arco-design/web-react';

interface TerminalProps {
    sessionId: string;
    wsBaseUrl?: string; // overrides auto-derived base
    onClose?: () => void;
}

// Terminal bridges an xterm.js instance to the backend PTY over WebSocket.
export default function Terminal({ sessionId, wsBaseUrl, onClose }: TerminalProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const termRef = useRef<XTerm | null>(null);
    const fitRef = useRef<FitAddon | null>(null);
    const wsRef = useRef<WebSocket | null>(null);

    useEffect(() => {
        if (!containerRef.current) return;
        const term = new XTerm({
            cursorBlink: true,
            fontSize: 13,
            fontFamily: 'Menlo, Consolas, "Courier New", monospace',
            theme: { background: '#1e1e1e' },
        });
        const fit = new FitAddon();
        term.loadAddon(fit);
        term.open(containerRef.current);
        try { fit.fit(); } catch { /* container not sized yet */ }
        termRef.current = term;
        fitRef.current = fit;

        const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
        const base = wsBaseUrl || `${proto}://${window.location.host}`;
        const token = localStorage.getItem('access_token') || '';
        const ws = new WebSocket(`${base}/ws/terminal?session=${encodeURIComponent(sessionId)}&token=${encodeURIComponent(token)}`);
        wsRef.current = ws;
        ws.onopen = () => {
            term.focus();
        };
        ws.onmessage = (ev) => {
            term.write(ev.data as string);
        };
        ws.onerror = () => Message.error('终端连接出错');
        ws.onclose = () => {
            term.write('\r\n\x1b[31m[连接已关闭]\x1b[0m\r\n');
            onClose?.();
        };

        term.onData((data) => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'input', data }));
            }
        });

        const onResize = () => {
            try {
                fit.fit();
                const dims = fit.proposeDimensions();
                if (dims && ws.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify({ type: 'resize', cols: dims.cols, rows: dims.rows }));
                }
            } catch { /* ignore */ }
        };
        window.addEventListener('resize', onResize);

        return () => {
            window.removeEventListener('resize', onResize);
            ws.close();
            term.dispose();
            termRef.current = null;
        };
    }, [sessionId, wsBaseUrl, onClose]);

    return <div ref={containerRef} style={{ width: '100%', height: '100%', background: '#1e1e1e' }} />;
}
