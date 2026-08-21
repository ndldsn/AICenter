// Shared formatting helpers for the Docker feature.

export function formatBytes(bytes: number): string {
    if (!bytes && bytes !== 0) return '-';
    if (bytes === 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    return `${(bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

export function formatUnixTime(unixSec: number): string {
    if (!unixSec) return '-';
    const d = new Date(unixSec * 1000);
    return d.toLocaleString('zh-CN', { hour12: false });
}

export function formatPorts(ports: Array<{ private_port: number; public_port?: number; type: string }>): string {
    if (!ports || ports.length === 0) return '-';
    return ports
        .map((p) => (p.public_port ? `${p.public_port}->${p.private_port}/${p.type}` : `${p.private_port}/${p.type}`))
        .join(', ');
}

export function shortId(id: string): string {
    if (!id) return '-';
    return id.length > 12 ? id.slice(0, 12) : id;
}