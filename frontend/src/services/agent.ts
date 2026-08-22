import { apiGet, apiPost, apiPut, apiDelete } from './api';

export interface Agent {
    id: string;
    name: string;
    description?: string;
    model_id: string;
    system_prompt?: string;
    temperature: number;
    max_tokens: number;
    max_iterations: number;
    tools: string[];
    tool_permission_mode: string;
    require_approval_for: string[];
    is_enabled: boolean;
    created_by?: string;
    created_at: string;
    updated_at: string;
}

export interface AgentSession {
    id: string;
    agent_id: string;
    user_id: string;
    server_id?: string;
    title: string;
    status: string;
    context_summary?: string;
    token_input: number;
    token_output: number;
    started_at: string;
    ended_at?: string;
    created_at: string;
}

export interface AgentMessage {
    id: string;
    session_id: string;
    role: string;
    content?: string;
    tool_call_id?: string;
    tool_name?: string;
    tool_args?: any;
    tool_result?: any;
    metadata?: any;
    created_at: string;
}

export interface ApprovalRequest {
    id: string;
    request_type: string;
    status: string;
    requested_by?: string;
    tool_name: string;
    tool_args?: any;
    risk_level: string;
    dry_run_result?: any;
    approved_by?: string;
    created_at: string;
}

const api = {
    listAgents: (enabled?: boolean) =>
        apiGet<{ code: number; data: { items: Agent[]; total: number } }>(
            `/agents${enabled != null ? `?enabled=${enabled}` : ''}`
        ).then(r => r.data),
    getAgent: (id: string) =>
        apiGet<{ code: number; data: Agent }>(`/agents/${id}`).then(r => r.data),
    createAgent: (a: Partial<Agent>) =>
        apiPost<{ code: number; data: Agent }>('/agents', a).then(r => r.data),
    updateAgent: (id: string, a: Partial<Agent>) =>
        apiPut<{ code: number; data: Agent }>(`/agents/${id}`, a).then(r => r.data),
    deleteAgent: (id: string) =>
        apiDelete<{ code: number; data: any }>(`/agents/${id}`),

    createSession: (body: { agent_id: string; query: string; server_id?: string }) =>
        apiPost<{ code: number; data: AgentSession }>(`/agents/${body.agent_id}/sessions`, body).then(r => r.data),
    listSessions: (agentId?: string) =>
        apiGet<{ code: number; data: { items: AgentSession[]; total: number } }>(
            `/agents/sessions${agentId ? `?agent_id=${agentId}` : ''}`
        ).then(r => r.data),
    getSession: (id: string) =>
        apiGet<{ code: number; data: { session: AgentSession; messages: AgentMessage[] } }>(
            `/agents/sessions/${id}`
        ).then(r => r.data),
    sendMessage: (sessionId: string, message: string) =>
        apiPost<{ code: number; data: any }>(`/agents/sessions/${sessionId}/messages`, { message }),

    listApprovals: (status?: string) =>
        apiGet<{ code: number; data: { items: ApprovalRequest[]; total: number } }>(
            `/approvals${status ? `?status=${status}` : ''}`
        ).then(r => r.data),
    getApproval: (id: string) =>
        apiGet<{ code: number; data: ApprovalRequest }>(`/approvals/${id}`).then(r => r.data),
    approve: (id: string) =>
        apiPost(`/approvals/${id}/approve`, {}).then(r => r),
    reject: (id: string) =>
        apiPost(`/approvals/${id}/reject`, {}).then(r => r),
};

export { api as agentApi };
