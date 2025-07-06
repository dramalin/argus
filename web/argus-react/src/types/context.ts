/**
 * Type definitions for React contexts
 */
import type { ReactNode } from 'react';
import type { SystemMetrics, ProcessInfo, ProcessResponse, ProcessQueryParams, AsyncData, ThemeMode, Notification } from './index';
import type { ColorTone } from '../theme/theme';

/**
 * State for the metrics context
 */
export interface MetricsState {
  /** System metrics data */
  metrics: AsyncData<SystemMetrics>;
  /** Polling interval in milliseconds */
  pollingInterval: number;
  /** Whether metrics polling is enabled */
  pollingEnabled: boolean;
}

/**
 * Action types for metrics context
 */
export type MetricsAction =
  | { type: 'FETCH_METRICS_START' }
  | { type: 'FETCH_METRICS_SUCCESS'; payload: SystemMetrics }
  | { type: 'FETCH_METRICS_ERROR'; payload: string }
  | { type: 'SET_POLLING_INTERVAL'; payload: number }
  | { type: 'SET_POLLING_ENABLED'; payload: boolean };

/**
 * Context value for metrics context
 */
export interface MetricsContextValue {
  /** Metrics state */
  state: MetricsState;
  /** Function to refresh metrics */
  refreshMetrics: () => Promise<void>;
  /** Function to set polling interval */
  setPollingInterval: (interval: number) => void;
  /** Function to enable/disable polling */
  setPollingEnabled: (enabled: boolean) => void;
}

/**
 * State for the processes context
 */
export interface ProcessesState {
  /** Process data */
  processData: AsyncData<ProcessResponse>;
  /** Current query parameters */
  queryParams: ProcessQueryParams;
  /** Polling interval in milliseconds */
  pollingInterval: number;
  /** Whether process polling is enabled */
  pollingEnabled: boolean;
}

/**
 * Action types for processes context
 */
export type ProcessesAction =
  | { type: 'FETCH_PROCESSES_START' }
  | { type: 'FETCH_PROCESSES_SUCCESS'; payload: ProcessResponse }
  | { type: 'FETCH_PROCESSES_ERROR'; payload: string }
  | { type: 'UPDATE_QUERY_PARAMS'; payload: Partial<ProcessQueryParams> }
  | { type: 'SET_POLLING_INTERVAL'; payload: number }
  | { type: 'SET_POLLING_ENABLED'; payload: boolean };

/**
 * Context value for processes context
 */
export interface ProcessesContextValue {
  /** Processes state */
  state: ProcessesState;
  /** Function to refresh processes */
  refreshProcesses: () => Promise<void>;
  /** Function to update query parameters */
  updateQuery: (params: Partial<ProcessQueryParams>) => void;
  /** Function to set polling interval */
  setPollingInterval: (interval: number) => void;
  /** Function to enable/disable polling */
  setPollingEnabled: (enabled: boolean) => void;
}

/**
 * State for the UI context
 */
export interface UiState {
  /** Current theme mode */
  themeMode: ThemeMode;
  /** Current color tone */
  colorTone: ColorTone;
  /** Whether the sidebar is open */
  sidebarOpen: boolean;
  /** Current notifications */
  notifications: Notification[];
}

/**
 * Action types for UI context
 */
export type UiAction =
  | { type: 'SET_THEME_MODE'; payload: ThemeMode }
  | { type: 'TOGGLE_SIDEBAR' }
  | { type: 'SET_SIDEBAR_OPEN'; payload: boolean }
  | { type: 'ADD_NOTIFICATION'; payload: Notification }
  | { type: 'REMOVE_NOTIFICATION'; payload: string }
  | { type: 'CLEAR_NOTIFICATIONS' };

/**
 * Context value for UI context
 */
export interface UiContextValue {
  /** UI state */
  state: UiState;
  /** Function to set theme mode */
  setThemeMode: (mode: ThemeMode) => void;
  /** Function to set color tone */
  setColorTone: (tone: ColorTone) => void;
  /** Function to toggle sidebar */
  toggleSidebar: () => void;
  /** Function to set sidebar open state */
  setSidebarOpen: (open: boolean) => void;
  /** Function to add a notification */
  addNotification: (notification: Omit<Notification, 'id' | 'timestamp'>) => void;
  /** Function to remove a notification */
  removeNotification: (id: string) => void;
  /** Function to clear all notifications */
  clearNotifications: () => void;
}

/**
 * Props for context providers
 */
export interface ProviderProps {
  /** Child components */
  children: ReactNode;
} 