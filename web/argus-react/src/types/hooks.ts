/**
 * Type definitions for custom React hooks
 */
import type { SystemMetrics, CPUInfo, MemoryInfo, NetworkInfo, ProcessInfo, ProcessResponse, ProcessQueryParams, AsyncData, ApiResponse } from './index';

/**
 * Return type for useMetrics hook
 */
export interface UseMetricsResult {
  /** System metrics data */
  metrics: AsyncData<SystemMetrics>;
  /** Function to refresh metrics data */
  refreshMetrics: () => Promise<void>;
  /** CPU metrics data */
  cpu: CPUInfo | null;
  /** Memory metrics data */
  memory: MemoryInfo | null;
  /** Network metrics data */
  network: NetworkInfo | null;
  /** Last updated timestamp */
  lastUpdated: string | null;
}

/**
 * Return type for useProcesses hook
 */
export interface UseProcessesResult {
  /** Process data */
  processData: AsyncData<ProcessResponse>;
  /** Function to refresh process data */
  refreshProcesses: () => Promise<void>;
  /** Function to update query parameters */
  updateQuery: (newParams: Partial<ProcessQueryParams>) => void;
  /** Current query parameters */
  queryParams: ProcessQueryParams;
  /** Process list */
  processes: ProcessInfo[];
}

/**
 * Return type for useApiCache hook
 */
export interface UseApiCacheResult<T> {
  /** Cached data */
  data: T | null;
  /** Loading state */
  loading: boolean;
  /** Error message */
  error: string | null;
  /** Function to refresh data */
  refresh: () => Promise<void>;
  /** Last updated timestamp */
  lastUpdated: string | null;
}

/**
 * Options for useApiCache hook
 */
export interface UseApiCacheOptions {
  /** Cache key */
  cacheKey: string;
  /** Cache TTL in milliseconds */
  cacheTTL?: number;
  /** Whether to fetch data on mount */
  fetchOnMount?: boolean;
  /** Retry count for failed requests */
  retryCount?: number;
  /** Retry delay in milliseconds */
  retryDelay?: number;
  /** Request timeout in milliseconds */
  timeoutMs?: number;
}

/**
 * Return type for useLocalStorage hook
 */
export interface UseLocalStorageResult<T> {
  /** Stored value */
  value: T;
  /** Function to update stored value */
  setValue: (value: T | ((val: T) => T)) => void;
  /** Function to remove stored value */
  removeValue: () => void;
}

/**
 * Return type for useDebounce hook
 */
export interface UseDebounceResult<T> {
  /** Debounced value */
  debouncedValue: T;
  /** Function to update value */
  setValue: (value: T) => void;
  /** Function to flush debounce and update immediately */
  flush: () => void;
}

/**
 * Options for usePolling hook
 */
export interface UsePollingOptions {
  /** Polling interval in milliseconds */
  interval: number;
  /** Whether polling is enabled */
  enabled?: boolean;
  /** Maximum number of retries for failed requests */
  maxRetries?: number;
  /** Retry delay in milliseconds */
  retryDelay?: number;
}

/**
 * Return type for usePolling hook
 */
export interface UsePollingResult {
  /** Whether polling is active */
  isPolling: boolean;
  /** Function to start polling */
  startPolling: () => void;
  /** Function to stop polling */
  stopPolling: () => void;
  /** Function to trigger a single poll */
  triggerPoll: () => Promise<void>;
  /** Last poll timestamp */
  lastPollTime: number | null;
  /** Error message from last poll */
  error: string | null;
}

/**
 * Notification severity types
 */
export type NotificationSeverity = 'success' | 'error' | 'info' | 'warning';

/**
 * Return type for useNotification hook
 */
export interface UseNotificationResult {
  /** Function to show a notification */
  showNotification: (message: string, severity: NotificationSeverity) => void;
  /** Function to clear all notifications */
  clearNotifications: () => void;
}

/**
 * Options for useDataFetching hook
 */
export interface DataFetchingOptions {
  /** Initial loading state */
  initialLoading?: boolean;
  /** Cache TTL in milliseconds */
  cacheTTL?: number;
  /** Whether to fetch on mount */
  fetchOnMount?: boolean;
}

/**
 * Return type for useDataFetching hook
 */
export interface UseDataFetchingResult<T> {
  /** The fetched data */
  data: T | null;
  /** Loading state */
  loading: boolean;
  /** Error message */
  error: string | null;
  /** Timestamp of when the data was last updated */
  lastUpdated: string | null;
  /** Function to manually refresh the data */
  refetch: () => Promise<void>;
}

/**
 * Return type for useDialogState hook
 */
export interface UseDialogStateResult<T extends string = string> {
  /** Current state of all dialogs */
  dialogStates: Record<string, boolean>;
  /** Function to open a specific dialog */
  openDialog: (name: T) => void;
  /** Function to close a specific dialog */
  closeDialog: (name: T) => void;
  /** Function to close all dialogs */
  closeAllDialogs: () => void;
  /** Function to check if a specific dialog is open */
  isDialogOpen: (name: T) => boolean;
}

/**
 * Options for useDateFormatter hook
 */
export interface DateFormatterOptions {
  /** Locale for date formatting (default: 'en-US') */
  locale?: string;
  /** Default format for dates (default: full date and time) */
  defaultFormat?: Intl.DateTimeFormatOptions;
  /** Default value for invalid dates (default: 'N/A') */
  invalidDateText?: string;
}

/**
 * Return type for useDateFormatter hook
 */
export interface UseDateFormatterResult {
  /** Format a date string to a localized string */
  formatDate: (dateString?: string, formatOptions?: Intl.DateTimeFormatOptions) => string;
  /** Format a timestamp (milliseconds) to a localized string */
  formatTimestamp: (timestamp: number, formatOptions?: Intl.DateTimeFormatOptions) => string;
  /** Format a date string as a relative time (e.g., "2 hours ago") */
  formatRelativeTime: (dateString?: string) => string;
} 