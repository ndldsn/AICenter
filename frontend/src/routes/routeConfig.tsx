import { lazy } from 'react';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { ROUTE_PERMISSIONS } from '@/utils/permissions';
import LoginPage from '@/features/auth/LoginPage';
import AgentChatPage from '@/features/agents/AgentChatPage';

const DashboardPage = lazy(() => import('@/features/dashboard/DashboardPage'));
const ServerListPage = lazy(() => import('@/features/servers/ServerListPage'));
const ServerDetailPage = lazy(() => import('@/features/servers/ServerDetailPage'));
const BatchCommandPage = lazy(() => import('@/features/servers/BatchCommandPage'));
const DockerDashboardPage = lazy(() => import('@/features/docker/DockerDashboardPage'));
const ModelListPage = lazy(() => import('@/features/ai/ModelListPage'));
const AgentListPage = lazy(() => import('@/features/agents/AgentListPage'));
const TaskListPage = lazy(() => import('@/features/tasks/TaskListPage'));
const MonitorDashboardPage = lazy(() => import('@/features/monitor/MonitorDashboardPage'));
const PendingApprovalPage = lazy(() => import('@/features/approvals/PendingApprovalPage'));
const AuditLogPage = lazy(() => import('@/features/audit/AuditLogPage'));
const NotificationCenterPage = lazy(() => import('@/features/notification/NotificationCenterPage'));
const SettingsPage = lazy(() => import('@/features/settings/SettingsPage'));

export interface RouteMeta {
    path: string;
    label: string;
    icon: string;
    labelKey: string;
    permission?: string | string[];
}

export const routeMeta: RouteMeta[] = [
    { path: '/', label: '仪表盘', icon: 'dashboard', labelKey: 'dashboard', permission: ROUTE_PERMISSIONS['/'] },
    { path: '/servers', label: '服务器', icon: 'server', labelKey: 'servers', permission: ROUTE_PERMISSIONS['/servers'] },
    { path: '/servers/batch', label: '批量操作', icon: 'cluster', labelKey: 'servers', permission: ROUTE_PERMISSIONS['/servers/batch'] },
    { path: '/servers/:id', label: '服务器详情', icon: 'server', labelKey: 'servers', permission: ROUTE_PERMISSIONS['/servers'] },
    { path: '/docker', label: '容器', icon: 'docker', labelKey: 'docker', permission: ROUTE_PERMISSIONS['/docker'] },
    { path: '/models', label: 'AI 模型', icon: 'apps', labelKey: 'ai', permission: ROUTE_PERMISSIONS['/models'] },
    { path: '/agents', label: '智能体', icon: 'robot', labelKey: 'agents', permission: ROUTE_PERMISSIONS['/agents'] },
    { path: '/agents/:id/chat', label: '智能体对话', icon: 'robot', labelKey: 'agents', permission: ROUTE_PERMISSIONS['/agents'] },
    { path: '/tasks', label: '任务', icon: 'calendar', labelKey: 'tasks', permission: ROUTE_PERMISSIONS['/tasks'] },
    { path: '/monitor', label: '监控', icon: 'eye', labelKey: 'monitor', permission: ROUTE_PERMISSIONS['/monitor'] },
    { path: '/notifications', label: '通知', icon: 'notification', labelKey: 'notifications', permission: ROUTE_PERMISSIONS['/notifications'] },
    { path: '/approvals', label: '待审批', icon: 'check', labelKey: 'approvals', permission: ROUTE_PERMISSIONS['/approvals'] },
    { path: '/audit', label: '审计日志', icon: 'file', labelKey: 'audit', permission: ROUTE_PERMISSIONS['/audit'] },
    { path: '/settings', label: '设置', icon: 'settings', labelKey: 'settings', permission: ROUTE_PERMISSIONS['/settings'] },
];

export const routesConfig = [
    { path: '/login', element: <LoginPage /> },
    { path: '/', element: <ProtectedRoute path="/"><DashboardPage /></ProtectedRoute> },
    { path: '/servers', element: <ProtectedRoute path="/servers"><ServerListPage /></ProtectedRoute> },
    { path: '/servers/batch', element: <ProtectedRoute path="/servers/batch"><BatchCommandPage /></ProtectedRoute> },
    { path: '/servers/:id', element: <ProtectedRoute path="/servers"><ServerDetailPage /></ProtectedRoute> },
    { path: '/docker', element: <ProtectedRoute path="/docker"><DockerDashboardPage /></ProtectedRoute> },
    { path: '/models', element: <ProtectedRoute path="/models"><ModelListPage /></ProtectedRoute> },
    { path: '/agents', element: <ProtectedRoute path="/agents"><AgentListPage /></ProtectedRoute> },
    { path: '/agents/:id/chat', element: <ProtectedRoute path="/agents"><AgentChatPage /></ProtectedRoute> },
    { path: '/tasks', element: <ProtectedRoute path="/tasks"><TaskListPage /></ProtectedRoute> },
    { path: '/monitor', element: <ProtectedRoute path="/monitor"><MonitorDashboardPage /></ProtectedRoute> },
    { path: '/notifications', element: <ProtectedRoute path="/notifications"><NotificationCenterPage /></ProtectedRoute> },
    { path: '/approvals', element: <ProtectedRoute path="/approvals"><PendingApprovalPage /></ProtectedRoute> },
    { path: '/audit', element: <ProtectedRoute path="/audit"><AuditLogPage /></ProtectedRoute> },
    { path: '/settings', element: <ProtectedRoute path="/settings"><SettingsPage /></ProtectedRoute> },
];
