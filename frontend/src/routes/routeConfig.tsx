import { lazy } from 'react';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { ROUTE_PERMISSIONS } from '@/utils/permissions';
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
    permission?: string | string[];
}

export const routeMeta: RouteMeta[] = [
    { path: '/', label: 'Dashboard', icon: 'dashboard', permission: ROUTE_PERMISSIONS['/'] },
    { path: '/servers', label: 'Servers', icon: 'server', permission: ROUTE_PERMISSIONS['/servers'] },
    { path: '/servers/batch', label: 'Batch', icon: 'cluster', permission: ROUTE_PERMISSIONS['/servers/batch'] },
    { path: '/servers/:id', label: 'Server Detail', icon: 'server', permission: ROUTE_PERMISSIONS['/servers'] },
    { path: '/docker', label: 'Docker', icon: 'docker', permission: ROUTE_PERMISSIONS['/docker'] },
    { path: '/models', label: 'AI Models', icon: 'apps', permission: ROUTE_PERMISSIONS['/models'] },
    { path: '/agents', label: 'Agents', icon: 'robot', permission: ROUTE_PERMISSIONS['/agents'] },
    { path: '/tasks', label: 'Tasks', icon: 'calendar', permission: ROUTE_PERMISSIONS['/tasks'] },
    { path: '/monitor', label: 'Monitor', icon: 'eye', permission: ROUTE_PERMISSIONS['/monitor'] },
    { path: '/notifications', label: 'Notifications', icon: 'notification', permission: ROUTE_PERMISSIONS['/notifications'] },
    { path: '/approvals', label: 'Approvals', icon: 'check', permission: ROUTE_PERMISSIONS['/approvals'] },
    { path: '/audit', label: 'Audit Log', icon: 'file', permission: ROUTE_PERMISSIONS['/audit'] },
    { path: '/settings', label: 'Settings', icon: 'settings', permission: ROUTE_PERMISSIONS['/settings'] },
];

export const routesConfig = [
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
