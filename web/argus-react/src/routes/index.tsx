/**
 * Routes configuration for the application
 * This file defines all routes and their corresponding lazy-loaded components
 */
import { lazy, Suspense } from 'react';
import LoadingFallback from '../components/LoadingFallback';

// Lazy-loaded route components
const Dashboard = lazy(() => import('../Dashboard'));
const Settings = lazy(() => import('./Settings'));
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
 * Currently not used, but prepared for future multi-page functionality
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
    path: '/settings',
    component: Settings,
    label: 'Settings',
    showInNav: true,
    icon: 'settings',
  },
  {
    path: '/alerts',
    component: Alerts,
    label: 'Alerts',
    showInNav: true,
    icon: 'notifications',
  },
  {
    path: '/tasks',
    component: Tasks,
    label: 'Tasks',
    showInNav: true,
    icon: 'assignment',
  },
  {
    path: '*',
    component: NotFound,
    label: 'Not Found',
    showInNav: false,
  },
]; 