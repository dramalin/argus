import { lazy, Suspense } from 'react';
import { ThemeProvider, CssBaseline } from '@mui/material';
import theme from './theme/theme';
import AppProvider from './context/AppProvider';
import ErrorBoundary from './components/ErrorBoundary';
import LoadingFallback from './components/LoadingFallback';
import './App.css';

// Lazy-loaded components
const Layout = lazy(() => import('./components/Layout'));
const Dashboard = lazy(() => import('./Dashboard'));

/**
 * Root application component
 * Wraps the entire application with providers and error boundaries
 * Uses React.lazy and Suspense for code splitting
 */
function App(): JSX.Element {
  return (
    <ErrorBoundary>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <AppProvider>
          <Suspense fallback={<LoadingFallback message="Loading application..." contained={false} />}>
            <Layout>
              <ErrorBoundary>
                <Suspense fallback={<LoadingFallback message="Loading dashboard..." />}>
                  <Dashboard />
                </Suspense>
              </ErrorBoundary>
            </Layout>
          </Suspense>
        </AppProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
