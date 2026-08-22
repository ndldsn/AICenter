import { apiGet, apiPost, apiPut, apiDelete } from './api';

export interface Metric {
    id: string;
    server_id?: string;
    metric_name: string;
    value: number;
    unit?: string;
    labels?: Record<string, string>;
    collected_at: string;
}

export interface MetricPoint {
    bucket: string;
    avg: number;
    min: number;
    max: number;
    count: number;
}

export interface AlertRule {
    id: string;
    name: string;
    metric_name: string;
    condition: string; // gt / lt / gte / lte
    threshold: number;
    duration: number;
    severity: string; // info / warning / critical
    server_id?: string | null;
    is_enabled: boolean;
    cooldown: number;
    created_at: string;
    updated_at: string;
}

export interface AlertEvent {
    id: string;
    rule_id?: string;
    rule_name?: string;
    server_id?: string;
    metric_name?: string;
    value: number;
    threshold: number;
    condition?: string;
    severity: string;
    message?: string;
    status: string; // firing / acknowledged
    triggered_at: string;
    acknowledged_by?: string;
    acknowledged_at?: string;
}

export const monitorApi = {
    queryMetrics: (params: { server_id?: string; name?: string; since?: string; limit?: number; aggregate?: string }) =>
        apiGet<{ code: number; data: { items: (Metric | MetricPoint)[]; total?: number } }>(
            `/monitor/metrics${toQuery(params)}`
        ).then(r => r.data),
    latestMetrics: (serverId?: string) =>
        apiGet<{ code: number; data: { items: Metric[]; total: number } }>(
            `/monitor/metrics/latest${serverId ? `?server_id=${encodeURIComponent(serverId)}` : ''}`
        ).then(r => r.data),
    ingestMetrics: (metrics: Partial<Metric>[]) =>
        apiPost<{ code: number; data: any }>('/monitor/metrics/ingest', { metrics }).then(r => r.data),

    listRules: () =>
        apiGet<{ code: number; data: { items: AlertRule[]; total: number } }>('/monitor/alert-rules').then(r => r.data),
    createRule: (rule: Partial<AlertRule>) =>
        apiPost<{ code: number; data: AlertRule }>('/monitor/alert-rules', rule).then(r => r.data),
    updateRule: (id: string, rule: Partial<AlertRule>) =>
        apiPut<{ code: number; data: AlertRule }>(`/monitor/alert-rules/${id}`, rule).then(r => r.data),
    deleteRule: (id: string) => apiDelete(`/monitor/alert-rules/${id}`),

    listAlerts: (status?: string) =>
        apiGet<{ code: number; data: { items: AlertEvent[]; total: number } }>(
            `/monitor/alerts${status ? `?status=${status}` : ''}`
        ).then(r => r.data),
    ackAlert: (id: string) => apiPost(`/monitor/alerts/${id}/ack`, {}),
};

function toQuery(params: Record<string, any>): string {
    const qs = Object.entries(params)
        .filter(([, v]) => v !== undefined && v !== null && v !== '')
        .map(([k, v]) => `${k}=${encodeURIComponent(v)}`)
        .join('&');
    return qs ? `?${qs}` : '';
}
