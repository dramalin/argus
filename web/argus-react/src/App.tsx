import React, { lazy, Suspense } from 'react';
import { ThemeProvider, CssBaseline } from '@mui/material';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import theme from './theme/theme';
import AppProvider from './context/AppProvider';
import ErrorBoundary from './components/ErrorBoundary';
import LoadingFallback from './components/LoadingFallback';
import { routes, RouteComponent } from './routes';
import './App.css';

// Lazy-loaded components
const Layout = lazy(() => import('./components/Layout'));

/**
 * Root application component
 * Wraps the entire application with providers and error boundaries
 * Uses React Router for navigation and React.lazy for code splitting
 */
function App(): React.ReactElement {
  return (
    <ErrorBoundary>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <AppProvider>
          <BrowserRouter>
            <Suspense fallback={<LoadingFallback message="Loading application..." contained={false} />}>
              <Layout>
                <ErrorBoundary>
                  <Routes>
                    {routes.map((route) => (
                      <Route
                        key={route.path}
                        path={route.path}
                        element={<RouteComponent component={route.component} />}
                      />
                    ))}
                  </Routes>
                </ErrorBoundary>
              </Layout>
            </Suspense>
          </BrowserRouter>
        </AppProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
