export type Locale = 'zh-CN' | 'en-US';

type Dictionary = Record<string, Record<Locale, string>>;

const dictionary: Dictionary = {
  // Navbar
  'navbar.brand': { 'zh-CN': 'AICenter 控制台', 'en-US': 'AICenter Console' },
  'navbar.profile': { 'zh-CN': '个人资料', 'en-US': 'Profile' },
  'navbar.settings': { 'zh-CN': '设置', 'en-US': 'Settings' },
  'navbar.logout': { 'zh-CN': '退出登录', 'en-US': 'Logout' },
  'navbar.logoutSuccess': { 'zh-CN': '已退出登录', 'en-US': 'Logged out successfully' },

  // Sidebar
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

  // Login
  'login.title': { 'zh-CN': 'AICenter 登录', 'en-US': 'AICenter Login' },
  'login.username': { 'zh-CN': '用户名', 'en-US': 'Username' },
  'login.password': { 'zh-CN': '密码', 'en-US': 'Password' },
  'login.submit': { 'zh-CN': '登录', 'en-US': 'Login' },
  'login.loading': { 'zh-CN': '登录中...', 'en-US': 'Logging in...' },
  'login.error.empty': { 'zh-CN': '请输入用户名和密码', 'en-US': 'Please enter username and password' },
  'login.error.invalid': { 'zh-CN': '用户名或密码错误', 'en-US': 'Invalid username or password' },

  // Dashboard
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

  // Common
  'common.cancel': { 'zh-CN': '取消', 'en-US': 'Cancel' },
  'common.confirm': { 'zh-CN': '确定', 'en-US': 'Confirm' },
  'common.save': { 'zh-CN': '保存', 'en-US': 'Save' },
  'common.delete': { 'zh-CN': '删除', 'en-US': 'Delete' },
  'common.edit': { 'zh-CN': '编辑', 'en-US': 'Edit' },
  'common.create': { 'zh-CN': '创建', 'en-US': 'Create' },
  'common.update': { 'zh-CN': '更新', 'en-US': 'Update' },
  'common.refresh': { 'zh-CN': '刷新', 'en-US': 'Refresh' },
  'common.search': { 'zh-CN': '搜索', 'en-US': 'Search' },
  'common.actions': { 'zh-CN': '操作', 'en-US': 'Actions' },
  'common.status': { 'zh-CN': '状态', 'en-US': 'Status' },
  'common.name': { 'zh-CN': '名称', 'en-US': 'Name' },
  'common.type': { 'zh-CN': '类型', 'en-US': 'Type' },
  'common.description': { 'zh-CN': '描述', 'en-US': 'Description' },
  'common.enabled': { 'zh-CN': '启用', 'en-US': 'Enabled' },
  'common.disabled': { 'zh-CN': '停用', 'en-US': 'Disabled' },
  'common.loading': { 'zh-CN': '加载中...', 'en-US': 'Loading...' },
};

type Translator = {
  t: (key: string, fallback?: string) => string;
  locale: Locale;
  setLocale: (locale: Locale) => void;
  toggleLocale: () => Locale;
};

let currentLocale: Locale = (localStorage.getItem('locale') as Locale) || 'zh-CN';

export const i18n: Translator = {
  t: (key: string, fallback?: string) => {
    const entry = dictionary[key];
    if (!entry) return fallback || key;
    return entry[currentLocale] || fallback || entry['en-US'] || key;
  },
  get locale() {
    return currentLocale;
  },
  setLocale: (locale: Locale) => {
    currentLocale = locale;
    localStorage.setItem('locale', locale);
  },
  toggleLocale: () => {
    const next: Locale = currentLocale === 'zh-CN' ? 'en-US' : 'zh-CN';
    currentLocale = next;
    localStorage.setItem('locale', next);
    return next;
  },
};

export { dictionary };
