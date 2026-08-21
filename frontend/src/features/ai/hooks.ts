import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { aiApi, AIProvider, ChatMessage } from '@/services/ai';
import { Message } from '@arco-design/web-react';

export function useProviders() {
    return useQuery({
        queryKey: ['ai', 'providers'],
        queryFn: () => aiApi.listProviders(),
    });
}

export function useProvider(id: string) {
    return useQuery({
        queryKey: ['ai', 'provider', id],
        queryFn: () => aiApi.getProvider(id),
        enabled: !!id,
    });
}

export function useCreateProvider() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (provider: Partial<AIProvider> & { api_key_enc?: string }) =>
            aiApi.createProvider(provider),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['ai', 'providers'] });
            Message.success('Provider created');
        },
        onError: (err: any) => Message.error(err?.message || 'Create failed'),
    });
}

export function useUpdateProvider() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, provider }: { id: string; provider: Partial<AIProvider> & { api_key_enc?: string } }) =>
            aiApi.updateProvider(id, provider),
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['ai', 'providers'] });
            queryClient.invalidateQueries({ queryKey: ['ai', 'provider', data.id] });
            Message.success('Provider updated');
        },
        onError: (err: any) => Message.error(err?.message || 'Update failed'),
    });
}

export function useDeleteProvider() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => aiApi.deleteProvider(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['ai', 'providers'] });
            Message.success('Provider deleted');
        },
        onError: (err: any) => Message.error(err?.message || 'Delete failed'),
    });
}

export function useProviderModels(providerId: string) {
    return useQuery({
        queryKey: ['ai', 'models', providerId],
        queryFn: () => aiApi.listModels(providerId),
        enabled: !!providerId,
    });
}

// Send a chat message and stream the response
export async function streamChat(
    providerId: string,
    modelId: string,
    messages: ChatMessage[],
    onChunk: (chunk: string) => void,
): Promise<void> {
    for await (const chunk of aiApi.chat(providerId, modelId, messages)) {
        onChunk(chunk);
    }
}