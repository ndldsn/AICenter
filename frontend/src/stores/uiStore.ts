import { create } from 'zustand';

type Locale = 'zh-CN' | 'en-US';

type Dictionary = Record<string, Record<Locale, string>>;

const dictionary: Dictionary = {
  'navbar.brand': { 'zh-CN': 'AICenter 控制台', 'en-US': 'AICenter Console' },
  'navbar.profile': { 'zh-CN': '个人资料', 'en-US': 'Profile' },
  'navbar.settings': { 'zh-CN': '设置', 'en-US': 'Settings' },
  'navbar.logout': { 'zh-CN': '退出登录', 'en-US': 'Logout' },
  'navbar.logoutSuccess': { 'zh-CN': '已退出登录', 'en-US': 'Logged out successfully' },

  'sidebar.dashboard': { 'zh-CN': '仪表盘', 'en-US': 'Dashboard' },
  'sidebar.servers': { 'zh-CN': '服务器', 'en-US': 'Servers' },
  'sidebar.docker': { 'zh-CN': 'Docker', 'en-US': 'Docker' },
  'sidebar.agents': { 'zh-CN': '智能体', 'en-US': 'Agents' },
  'sidebar.tasks': { 'zh-CN': '任务', 'en-US': 'Tasks' },
  'sidebar.monitor': { 'zh-CN': '监控', 'en-US': 'Monitor' },
  'sidebar.notifications': { 'zh-CN': '通知', 'en-US': 'Notifications' },
  'sidebar.approvals': { 'zh-CN': '待审批', 'en-US': 'Approvals' },
  'sidebar.audit': { 'zh-CN': '审计日志', 'en-US': 'Audit Log' },
  'sidebar.ai': { 'zh-CN': 'AI 模型', 'en-US': 'AI Models' },
  'sidebar.settings': { 'zh-CN': '设置', 'en-US': 'Settings' },

  'login.title': { 'zh-CN': 'AICenter 登录', 'en-US': 'AICenter Login' },
  'login.username': { 'zh-CN': '用户名', 'en-US': 'Username' },
  'login.password': { 'zh-CN': '密码', 'en-US': 'Password' },
  'login.submit': { 'zh-CN': '登录', 'en-US': 'Login' },
  'login.loading': { 'zh-CN': '登录中...', 'en-US': 'Logging in...' },
  'login.error.empty': { 'zh-CN': '请输入用户名和密码', 'en-US': 'Please enter username and password' },
  'login.error.invalid': { 'zh-CN': '用户名或密码错误', 'en-US': 'Invalid username or password' },

  'dashboard.title': { 'zh-CN': '仪表盘', 'en-US': 'Dashboard' },
  'dashboard.welcome': { 'zh-CN': '欢迎使用 AICenter - 你的 AI 运维控制中心', 'en-US': 'Welcome to AICenter - Your AI operations control center' },
  'dashboard.quickActions': { 'zh-CN': '快速操作', 'en-US': 'Quick Actions' },
  'dashboard.addServer': { 'zh-CN': '添加服务器', 'en-US': 'Add Server' },
  'dashboard.deployContainer': { 'zh-CN': '部署容器', 'en-US': 'Deploy Container' },
  'dashboard.createAgent': { 'zh-CN': '创建智能体', 'en-US': 'Create Agent' },
  'dashboard.addProvider': { 'zh-CN': '添加 Provider', 'en-US': 'Add Provider' },
  'dashboard.systemStatus': { 'zh-CN': '系统状态', 'en-US': 'System Status' },
  'dashboard.systemOnline': { 'zh-CN': 'AICenter 运行中。尚未连接任何服务器。', 'en-US': 'AICenter is running. No servers connected yet.' },
  'dashboard.systemHint': { 'zh-CN': '添加第一台服务器即可开始监控和管理基础设施。', 'en-US': 'Add your first server to start monitoring and managing infrastructure.' },

  'servers.title': { 'zh-CN': '服务器', 'en-US': 'Servers' },
  'servers.subtitle': { 'zh-CN': '管理你的 Linux 服务器并监控其状态', 'en-US': 'Manage your Linux servers and monitor their status' },
  'servers.refresh': { 'zh-CN': '刷新', 'en-US': 'Refresh' },
  'servers.add': { 'zh-CN': '添加服务器', 'en-US': 'Add Server' },
  'servers.empty': { 'zh-CN': '暂无服务器，点击“添加服务器”开始。', 'en-US': 'No servers yet. Click "Add Server" to get started.' },
  'servers.status.online': { 'zh-CN': '在线', 'en-US': 'Online' },
  'servers.status.offline': { 'zh-CN': '离线', 'en-US': 'Offline' },
  'servers.status.unknown': { 'zh-CN': '未知', 'en-US': 'Unknown' },
  'servers.connected': { 'zh-CN': '已连接', 'en-US': 'Connected' },
  'servers.disconnected': { 'zh-CN': '未连接', 'en-US': 'Disconnected' },
  'servers.never': { 'zh-CN': '从未', 'en-US': 'Never' },
  'servers.testConnection': { 'zh-CN': '测试连接', 'en-US': 'Test Connection' },
  'servers.testing': { 'zh-CN': '正在测试...', 'en-US': 'Testing...' },
  'servers.testSuccess': { 'zh-CN': '连接测试成功', 'en-US': 'Connection test successful' },
  'servers.testFailed': { 'zh-CN': '连接测试失败', 'en-US': 'Connection test failed' },
  'servers.copied': { 'zh-CN': 'SSH 命令已复制', 'en-US': 'SSH command copied' },
  'servers.confirmDelete': { 'zh-CN': '确定删除该服务器？', 'en-US': 'Delete this server?' },
  'servers.cancel': { 'zh-CN': '取消', 'en-US': 'Cancel' },
  'servers.delete': { 'zh-CN': '删除', 'en-US': 'Delete' },
  'servers.actions': { 'zh-CN': '操作', 'en-US': 'Actions' },
  'servers.edit': { 'zh-CN': '编辑', 'en-US': 'Edit' },
  'servers.addTitle': { 'zh-CN': '添加服务器', 'en-US': 'Add Server' },
  'servers.editTitle': { 'zh-CN': '编辑服务器', 'en-US': 'Edit Server' },
  'servers.basicInfo': { 'zh-CN': '基本信息', 'en-US': 'Basic Info' },
  'servers.name': { 'zh-CN': '服务器名称', 'en-US': 'Server Name' },
  'servers.namePlaceholder': { 'zh-CN': '例如：生产 Web 服务器', 'en-US': 'e.g., Production Web Server' },
  'servers.host': { 'zh-CN': '主机 / IP', 'en-US': 'Host / IP' },
  'servers.hostPlaceholder': { 'zh-CN': '192.168.1.100 或 server.example.com', 'en-US': '192.168.1.100 or server.example.com' },
  'servers.port': { 'zh-CN': '端口', 'en-US': 'Port' },
  'servers.username': { 'zh-CN': '用户名', 'en-US': 'Username' },
  'servers.auth': { 'zh-CN': '认证', 'en-US': 'Authentication' },
  'servers.authType': { 'zh-CN': '认证方式', 'en-US': 'Auth Type' },
  'servers.authSelect': { 'zh-CN': '选择认证方式', 'en-US': 'Select auth type' },
  'servers.password': { 'zh-CN': '密码', 'en-US': 'Password' },
  'servers.passwordPlaceholder': { 'zh-CN': '请输入密码', 'en-US': 'Enter password' },
  'servers.privateKey': { 'zh-CN': '私钥', 'en-US': 'Private Key' },
  'servers.privateKeyPlaceholder': { 'zh-CN': '粘贴你的 SSH 私钥...', 'en-US': 'Paste your SSH private key here...' },
  'servers.testResultTitle': { 'zh-CN': '连接测试结果', 'en-US': 'Connection Test Result' },
  'servers.systemInfoTitle': { 'zh-CN': '系统信息', 'en-US': 'System Info' },
  'servers.os': { 'zh-CN': '系统', 'en-US': 'OS' },
  'servers.hostname': { 'zh-CN': '主机名', 'en-US': 'Hostname' },
  'servers.kernel': { 'zh-CN': '内核', 'en-US': 'Kernel' },
  'servers.cpuCores': { 'zh-CN': 'CPU 核心', 'en-US': 'CPU Cores' },
  'servers.memoryGB': { 'zh-CN': '内存 (GB)', 'en-US': 'Memory (GB)' },
  'servers.timestamp': { 'zh-CN': '时间', 'en-US': 'Timestamp' },
  'servers.message': { 'zh-CN': '信息', 'en-US': 'Message' },
  'servers.sshBanner': { 'zh-CN': 'SSH Banner', 'en-US': 'SSH Banner' },

  'docker.title': { 'zh-CN': 'Docker', 'en-US': 'Docker' },
  'docker.subtitle': { 'zh-CN': '管理容器、镜像、卷、网络和 Compose 项目', 'en-US': 'Manage containers, images, volumes, networks, and compose projects' },

  'agents.title': { 'zh-CN': '智能体', 'en-US': 'AI Agents' },
  'agents.subtitle': { 'zh-CN': '配置智能体、选择工具和权限策略，然后在对话中运行', 'en-US': 'Configure agents, choose tools and permission policy, then run them in a chat.' },
  'agents.refresh': { 'zh-CN': '刷新', 'en-US': 'Refresh' },
  'agents.create': { 'zh-CN': '新建智能体', 'en-US': 'Create Agent' },
  'agents.edit': { 'zh-CN': '编辑智能体', 'en-US': 'Edit Agent' },
  'agents.empty': { 'zh-CN': '暂无智能体，请先创建。', 'en-US': 'No agents yet. Create one to start.' },
  'agents.name': { 'zh-CN': '名称', 'en-US': 'Name' },
  'agents.description': { 'zh-CN': '描述', 'en-US': 'Description' },
  'agents.model': { 'zh-CN': '模型 (provider id)', 'en-US': 'Model (provider id)' },
  'agents.systemPrompt': { 'zh-CN': '系统提示词', 'en-US': 'System prompt' },
  'agents.permissionMode': { 'zh-CN': '权限模式', 'en-US': 'Permission mode' },
  'agents.availableTools': { 'zh-CN': '可用工具', 'en-US': 'Available tools' },
  'agents.requireApproval': { 'zh-CN': '需要审批的工具', 'en-US': 'Require approval for' },
  'agents.temperature': { 'zh-CN': '温度', 'en-US': 'Temperature' },
  'agents.maxTokens': { 'zh-CN': '最大令牌数', 'en-US': 'Max tokens' },
  'agents.maxIterations': { 'zh-CN': '最大迭代次数', 'en-US': 'Max iterations' },
  'agents.enabled': { 'zh-CN': '启用', 'en-US': 'Enabled' },
  'agents.notFound': { 'zh-CN': '未找到智能体', 'en-US': 'Agent not found' },
  'agents.untitled': { 'zh-CN': '未命名', 'en-US': 'untitled' },
  'agents.modelLabel': { 'zh-CN': '模型', 'en-US': 'Model' },
  'agents.tempLabel': { 'zh-CN': '温度', 'en-US': 'Temp' },
  'agents.iterLabel': { 'zh-CN': '迭代', 'en-US': 'Iter' },
  'agents.user': { 'zh-CN': '用户', 'en-US': 'User' },
  'agents.assistant': { 'zh-CN': '助手', 'en-US': 'Assistant' },
  'agents.tool': { 'zh-CN': '工具', 'en-US': 'Tool' },
  'agents.args': { 'zh-CN': '参数', 'en-US': 'args' },
  'agents.result': { 'zh-CN': '结果', 'en-US': 'result' },
  'agents.pendingApproval': { 'zh-CN': '待审批', 'en-US': 'pending approval' },
  'agents.askPlaceholder': { 'zh-CN': '输入问题，例如：\'list servers\' 或 \'restart nginx\'', 'en-US': 'Ask the agent... e.g. \'list servers\' or \'restart nginx\'' },
  'agents.run': { 'zh-CN': '运行', 'en-US': 'Run' },

  'approvals.title': { 'zh-CN': '待审批', 'en-US': 'Approvals' },
  'approvals.subtitle': { 'zh-CN': '审查并批准高风险操作', 'en-US': 'Review and approve high-risk operations' },
  'approvals.empty': { 'zh-CN': '暂无待审批项。', 'en-US': 'No pending approvals.' },
  'approvals.hint': { 'zh-CN': '当 AI 智能体请求执行高风险操作时，会在这里展示供你审核。', 'en-US': 'When AI agents request high-risk operations, they will appear here for your review.' },

  'audit.title': { 'zh-CN': '审计日志', 'en-US': 'Audit Log' },
  'audit.subtitle': { 'zh-CN': '所有操作与变更的完整历史记录', 'en-US': 'Complete history of all operations and changes' },
  'audit.empty': { 'zh-CN': '暂无审计日志。', 'en-US': 'No audit logs yet.' },
  'audit.hint': { 'zh-CN': '所有操作都会被记录在这里，用于合规与安全审计。', 'en-US': 'All operations will be recorded here for compliance and security.' },

  'ai.title': { 'zh-CN': 'AI 模型', 'en-US': 'AI Providers' },
  'ai.subtitle': { 'zh-CN': '配置 AI Provider，供智能体和其他功能使用', 'en-US': 'Configure AI providers and models for Agent operations' },
  'ai.addProvider': { 'zh-CN': '添加 Provider', 'en-US': 'Add Provider' },
  'ai.editProvider': { 'zh-CN': '编辑 Provider', 'en-US': 'Edit Provider' },
  'ai.create': { 'zh-CN': '创建', 'en-US': 'Create' },
  'ai.update': { 'zh-CN': '更新', 'en-US': 'Update' },
  'ai.name': { 'zh-CN': '名称', 'en-US': 'Name' },
  'ai.nameRequired': { 'zh-CN': '名称为必填', 'en-US': 'Name is required' },
  'ai.displayName': { 'zh-CN': '显示名称', 'en-US': 'Display Name' },
  'ai.type': { 'zh-CN': '类型', 'en-US': 'Type' },
  'ai.baseUrl': { 'zh-CN': '地址', 'en-US': 'Base URL' },
  'ai.apiKey': { 'zh-CN': 'API Key', 'en-US': 'API Key' },
  'ai.apiKeyExtra': { 'zh-CN': '编辑时留空则保留现有密钥', 'en-US': 'Leave empty to keep existing key when editing' },
  'ai.enabled': { 'zh-CN': '启用', 'en-US': 'Enabled' },
  'ai.default': { 'zh-CN': '默认', 'en-US': 'Default' },
  'ai.chatTitle': { 'zh-CN': '对话', 'en-US': 'Chat' },
  'ai.selectModel': { 'zh-CN': '选择模型', 'en-US': 'Select model' },
  'ai.startConversation': { 'zh-CN': '开始对话', 'en-US': 'Start a conversation' },
  'ai.you': { 'zh-CN': '你', 'en-US': 'You' },
  'ai.ai': { 'zh-CN': 'AI', 'en-US': 'AI' },
  'ai.send': { 'zh-CN': '发送', 'en-US': 'Send' },
  'ai.typeMessage': { 'zh-CN': '输入消息...', 'en-US': 'Type a message...' },
  'ai.namePlaceholder': { 'zh-CN': 'openai', 'en-US': 'openai' },
  'ai.displayNamePlaceholder': { 'zh-CN': 'OpenAI', 'en-US': 'OpenAI' },
  'ai.baseUrlPlaceholder': { 'zh-CN': 'https://api.openai.com/v1', 'en-US': 'https://api.openai.com/v1' },
  'ai.keyPlaceholder': { 'zh-CN': 'sk-...', 'en-US': 'sk-...' },

  'common.confirm': { 'zh-CN': '确定', 'en-US': 'Confirm' },
  'common.cancel': { 'zh-CN': '取消', 'en-US': 'Cancel' },
  'common.save': { 'zh-CN': '保存', 'en-US': 'Save' },
  'common.delete': { 'zh-CN': '删除', 'en-US': 'Delete' },
  'common.edit': { 'zh-CN': '编辑', 'en-US': 'Edit' },
  'common.create': { 'zh-CN': '创建', 'en-US': 'Create' },
  'common.update': { 'zh-CN': '更新', 'en-US': 'Update' },
  'common.refresh': { 'zh-CN': '刷新', 'en-US': 'Refresh' },
  'common.actions': { 'zh-CN': '操作', 'en-US': 'Actions' },
  'common.status': { 'zh-CN': '状态', 'en-US': 'Status' },
  'common.name': { 'zh-CN': '名称', 'en-US': 'Name' },
  'common.type': { 'zh-CN': '类型', 'en-US': 'Type' },
  'common.description': { 'zh-CN': '描述', 'en-US': 'Description' },
  'common.enabled': { 'zh-CN': '启用', 'en-US': 'Enabled' },
  'common.disabled': { 'zh-CN': '停用', 'en-US': 'Disabled' },
  'common.loading': { 'zh-CN': '加载中...', 'en-US': 'Loading...' },
  'common.success': { 'zh-CN': '成功', 'en-US': 'Success' },
  'common.failed': { 'zh-CN': '失败', 'en-US': 'Failed' },

  'auth.403': { 'zh-CN': '403 — 无权访问', 'en-US': '403 — Forbidden' },
  'auth.403Hint': { 'zh-CN': '你的角色无权访问该页面。', 'en-US': 'Your role does not have permission to access this page.' },
  'auth.back': { 'zh-CN': '返回', 'en-US': 'Go Back' },
};

const getInitialLocale = (): Locale =>
  (localStorage.getItem('locale') as Locale) || 'zh-CN';

export interface UIState {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark';
  locale: Locale;
  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  toggleTheme: () => void;
  setLocale: (locale: Locale) => void;
  toggleLocale: () => void;
  t: (key: string, fallback?: string) => string;
}

export const useUIStore = create<UIState>()((set, get) => ({
  sidebarCollapsed: false,
  theme: (localStorage.getItem('theme') as 'light' | 'dark') || 'light',
  locale: getInitialLocale(),
  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
  setTheme: (theme) => {
    localStorage.setItem('theme', theme);
    document.body.setAttribute('arco-theme', theme);
    set({ theme });
  },
  toggleTheme: () =>
    set((state) => {
      const theme = state.theme === 'light' ? 'dark' : 'light';
      localStorage.setItem('theme', theme);
      document.body.setAttribute('arco-theme', theme);
      return { theme };
    }),
  setLocale: (locale) => {
    localStorage.setItem('locale', locale);
    set({ locale });
  },
  toggleLocale: () =>
    set((state) => {
      const locale = state.locale === 'zh-CN' ? 'en-US' : 'zh-CN';
      localStorage.setItem('locale', locale);
      return { locale };
    }),
  t: (key: string, fallback?: string) => {
    const entry = dictionary[key];
    if (!entry) return fallback || key;
    return entry[get().locale] || fallback || entry['en-US'] || key;
  },
}));
