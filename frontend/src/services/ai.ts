import { apiGet, apiPost, apiPut, apiDelete } from './api';

export interface AIProvider {
    id: string;
    name: string;
    display_name: string;
    base_url: string;
    api_key_hint?: string;
    api_type: 'openai-compatible' | 'anthropic' | 'gemini' | 'mock';
    is_enabled: boolean;
    is_default: boolean;
    created_at?: string;
    updated_at?: string;
}

export interface AIModel {
    id: string;
    provider_id: string;
    name: string;
    model_id: string;
    model_type: string;
    max_tokens: number;
    supports_stream: boolean;
    supports_tools: boolean;
    is_enabled: boolean;
    is_default: boolean;
    created_at?: string;
    updated_at?: string;
}

export interface ChatMessage {
    role: 'system' | 'user' | 'assistant';
    content: string;
}

export const aiApi = {
    listProviders: () => apiGet<{ code: number; data: AIProvider[] }>('/ai/providers').then(r => r.data),

    getProvider: (id: string) =>
        apiGet<{ code: number; data: AIProvider }>(`/ai/providers/${id}`).then(r => r.data),

    createProvider: (provider: Partial<AIProvider> & { api_key_enc?: string }) =>
        apiPost<{ code: number; data: AIProvider }>('/ai/providers', provider).then(r => r.data),

    updateProvider: (id: string, provider: Partial<AIProvider> & { api_key_enc?: string }) =>
        apiPut<{ code: number; data: AIProvider }>(`/ai/providers/${id}`, provider).then(r => r.data),

    deleteProvider: (id: string) =>
        apiDelete<{ code: number }>(`/ai/providers/${id}`),

    listModels: (providerId: string) =>
        apiGet<{ code: number; data: AIModel[] }>(`/ai/models/${providerId}`).then(r => r.data),

    chat: async function* (providerId: string, modelId: string, messages: ChatMessage[]): AsyncGenerator<string> {
        const token = localStorage.getItem('token') || 'dev-token-1234';
        const resp = await fetch(`/api/v1/ai/chat`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify({
                provider_id: providerId,
                model_id: modelId,
                messages,
            }),
        });

        if (!resp.ok || !resp.body) {
            throw new Error(`Chat request failed: ${resp.status}`);
        }

        const reader = resp.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            buffer += decoder.decode(value, { stream: true });

            const lines = buffer.split('\n');
            buffer = lines.pop() || '';

            for (const line of lines) {
                const trimmed = line.trim();
                if (trimmed.startsWith('data: ')) {
                    const data = trimmed.slice(6).trim();
                    if (data === '[DONE]') return;
                    if (data.startsWith('{')) {
                        try {
                            const parsed = JSON.parse(data);
                            if (parsed.content) yield parsed.content;
                        } catch {
                            // ignore malformed
                        }
                    }
                }
            }
        }
    },
};