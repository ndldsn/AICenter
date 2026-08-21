import { lazy } from 'react';

const DashboardPage = lazy(() => import('@/features/dashboard/DashboardPage'));
const ServerListPage = lazy(() => import('@/features/servers/ServerListPage'));
const DockerDashboardPage = lazy(() => import('@/features/docker/DockerDashboardPage'));
const ModelListPage = lazy(() => import('@/features/models/ModelListPage'));
const AgentListPage = lazy(() => import('@/features/agents/AgentListPage'));
const TaskListPage = lazy(() => import('@/features/tasks/TaskListPage'));
const MonitorDashboardPage = lazy(() => import('@/features/monitor/MonitorDashboardPage'));
const PendingApprovalPage = lazy(() => import('@/features/approvals/PendingApprovalPage'));
const AuditLogPage = lazy(() => import('@/features/audit/AuditLogPage'));
const SettingsPage = lazy(() => import('@/features/settings/SettingsPage'));

export interface RouteMeta {
    path: string;
    label: string;
    icon: string;
}

export const routeMeta: RouteMeta[] = [
    { path: '/', label: 'Dashboard', icon: 'dashboard' },
    { path: '/servers', label: 'Servers', icon: 'server' },
    { path: '/docker', label: 'Docker', icon: 'docker' },
    { path: '/models', label: 'AI Models', icon: 'apps' },
    { path: '/agents', label: 'Agents', icon: 'robot' },
    { path: '/tasks', label: 'Tasks', icon: 'calendar' },
    { path: '/monitor', label: 'Monitor', icon: 'eye' },
    { path: '/approvals', label: 'Approvals', icon: 'check' },
    { path: '/audit', label: 'Audit Log', icon: 'file' },
    { path: '/settings', label: 'Settings', icon: 'settings' },
];

export const routesConfig = [
    { path: '/', element: <DashboardPage /> },
    { path: '/servers', element: <ServerListPage /> },
    { path: '/docker', element: <DockerDashboardPage /> },
    { path: '/models', element: <ModelListPage /> },
    { path: '/agents', element: <AgentListPage /> },
    { path: '/tasks', element: <TaskListPage /> },
    { path: '/monitor', element: <MonitorDashboardPage /> },
    { path: '/approvals', element: <PendingApprovalPage /> },
    { path: '/audit', element: <AuditLogPage /> },
    { path: '/settings', element: <SettingsPage /> },
];
