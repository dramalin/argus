import React, { lazy, Suspense, useContext } from 'react';
import { ThemeProvider, CssBaseline } from '@mui/material';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { getTheme } from './theme/theme';
import type { ColorTone, ThemeMode } from './theme/theme';
import AppProvider from './context/AppProvider';
import ErrorBoundary from './components/ErrorBoundary';
import LoadingFallback from './components/LoadingFallback';
import { routes, RouteComponent } from './routes';
import './App.css';
import { useNotification } from './hooks';
import { useUiContext } from './context/UiContext';

// Lazy-loaded components
const Layout = lazy(() => import('./components/Layout'));

// A component to handle notifications
const NotificationHandler: React.FC = () => {
  useNotification();
  return null;
};

/**
 * Root application component
 * Wraps the entire application with providers and error boundaries
 * Uses React Router for navigation and React.lazy for code splitting
 */
function App(): React.ReactElement {
  return (
    <AppProvider>
      <ErrorBoundary>
        <AppWithContext />
      </ErrorBoundary>
    </AppProvider>
  );
}

function AppWithContext(): React.ReactElement {
  const ui = useUiContext();
  const themeMode: ThemeMode = ui.state.themeMode || 'light';
  const colorTone: ColorTone = ui.state.colorTone || 'morandi';
  const theme = getTheme(colorTone, themeMode);
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <NotificationHandler />
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
    </ThemeProvider>
  );
}

export default App;
