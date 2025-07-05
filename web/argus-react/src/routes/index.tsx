/**
 * Routes configuration for the application
 * This file defines all routes and their corresponding lazy-loaded components
 */
import { lazy, Suspense } from 'react';
import LoadingFallback from '../components/LoadingFallback';

// Lazy-loaded route components
const Dashboard = lazy(() => import('../Dashboard'));
const Processes = lazy(() => import('./Processes'));
const Alerts = lazy(() => import('./Alerts'));
const Tasks = lazy(() => import('./Tasks'));
const NotFound = lazy(() => import('./NotFound'));

/**
 * Route configuration type
 */
export interface RouteConfig {
  /** Route path */
  path: string;
  /** Route component */
  component: React.ComponentType;
  /** Route label for navigation */
  label: string;
  /** Whether to show in navigation */
  showInNav: boolean;
  /** Icon name for navigation */
  icon?: string;
}

/**
 * Route component props
 */
interface RouteComponentProps {
  /** Component to render */
  component: React.ComponentType;
}

/**
 * Route component that wraps the component with Suspense
 */
export const RouteComponent: React.FC<RouteComponentProps> = ({ component: Component }) => {
  return (
    <Suspense fallback={<LoadingFallback message="Loading page..." />}>
      <Component />
    </Suspense>
  );
};

/**
 * Routes configuration
 * Used for React Router navigation
 */
export const routes: RouteConfig[] = [
  {
    path: '/',
    component: Dashboard,
    label: 'Dashboard',
    showInNav: true,
    icon: 'dashboard',
  },
  {
    path: '/tasks',
    component: Tasks,
    label: 'Tasks',
    showInNav: true,
    icon: 'task',
  },
  {
    path: '/alerts',
    component: Alerts,
    label: 'Alerts',
    showInNav: true,
    icon: 'notifications',
  },
  {
    path: '/processes',
    component: Processes,
    label: 'Processes',
    showInNav: true,
    icon: 'terminal',
  },
  {
    path: '*',
    component: NotFound,
    label: 'Not Found',
    showInNav: false,
  },
]; 