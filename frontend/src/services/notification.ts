import { apiGet, apiPost, apiPut, apiDelete } from './api';

export type ChannelType = 'webhook' | 'email' | 'sms' | 'im' | 'console';

export interface NotificationChannel {
    id: string;
    name: string;
    type: ChannelType;
    config?: string;
    is_enabled: boolean;
    created_at: string;
    updated_at: string;
}

export interface NotificationTemplate {
    id: string;
    name: string;
    event_type: string;
    subject?: string;
    body: string;
    channels?: string;
    is_enabled: boolean;
    created_at: string;
    updated_at: string;
}

export interface DeliveryLog {
    id: string;
    channel_id?: string;
    channel_type?: string;
    template_id?: string;
    event_type?: string;
    subject?: string;
    body?: string;
    status: string; // pending / sent / failed
    error_message?: string;
    created_at: string;
}

export const notificationApi = {
    listChannels: (enabled?: boolean) =>
        apiGet<{ code: number; data: { items: NotificationChannel[]; total: number } }>(
            `/notification/channels${enabled ? '?enabled=true' : ''}`
        ).then(r => r.data),
    createChannel: (c: Partial<NotificationChannel>) =>
        apiPost<{ code: number; data: NotificationChannel }>('/notification/channels', c).then(r => r.data),
    updateChannel: (id: string, c: Partial<NotificationChannel>) =>
        apiPut<{ code: number; data: NotificationChannel }>(`/notification/channels/${id}`, c).then(r => r.data),
    deleteChannel: (id: string) => apiDelete(`/notification/channels/${id}`),

    listTemplates: (eventType?: string) =>
        apiGet<{ code: number; data: { items: NotificationTemplate[]; total: number } }>(
            `/notification/templates${eventType ? `?event_type=${encodeURIComponent(eventType)}` : ''}`
        ).then(r => r.data),
    createTemplate: (t: Partial<NotificationTemplate>) =>
        apiPost<{ code: number; data: NotificationTemplate }>('/notification/templates', t).then(r => r.data),
    updateTemplate: (id: string, t: Partial<NotificationTemplate>) =>
        apiPut<{ code: number; data: NotificationTemplate }>(`/notification/templates/${id}`, t).then(r => r.data),
    deleteTemplate: (id: string) => apiDelete(`/notification/templates/${id}`),

    listDeliveryLogs: (status?: string, limit = 200) =>
        apiGet<{ code: number; data: { items: DeliveryLog[]; total: number } }>(
            `/notification/delivery-logs${status ? `?status=${status}` : ''}&limit=${limit}`
        ).then(r => r.data),

    sendTest: (payload: {
        event_type: string;
        title: string;
        severity?: string;
        message?: string;
        data?: Record<string, string>;
        channel_ids?: string[];
    }) => apiPost<{ code: number; data: any }>('/notification/send', payload).then(r => r.data),
};
