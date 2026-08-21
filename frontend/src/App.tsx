import { useRoutes } from 'react-router-dom';
import { MainLayout } from '@/layouts';
import { routesConfig } from '@/routes';

function App() {
    const routes = useRoutes(routesConfig);
    return <MainLayout>{routes}</MainLayout>;
}

export default App;
