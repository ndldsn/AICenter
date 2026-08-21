import { create } from 'zustand';

interface UIState {
    sidebarCollapsed: boolean;
    theme: 'light' | 'dark';
    toggleSidebar: () => void;
    setTheme: (theme: 'light' | 'dark') => void;
    toggleTheme: () => void;
}

export const useUIStore = create<UIState>()((set) => ({
    sidebarCollapsed: false,
    theme: (localStorage.getItem('theme') as 'light' | 'dark') || 'light',
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
}));
