import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useUIStore } from '@/stores/uiStore';
import App from './App';
import './styles/global.css';
import '@arco-design/web-react/dist/css/arco.css';

function LocaleSync() {
  const setLocale = useUIStore((s) => s.setLocale);
  React.useEffect(() => {
    const stored = localStorage.getItem('locale');
    if (stored === 'zh-CN' || stored === 'en-US') {
      setLocale(stored);
    }
  }, [setLocale]);
  return null;
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 30_000,
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
        <BrowserRouter
            future={{ v7_startTransition: true, v7_relativeSplatPath: true }}
        >
            <QueryClientProvider client={queryClient}>
                <LocaleSync />
                <App />
            </QueryClientProvider>
        </BrowserRouter>
    </React.StrictMode>
);
