/**
 * Type definitions for custom React hooks
 */
import { SystemMetrics, CPUInfo, MemoryInfo, NetworkInfo } from './api';
import { ProcessInfo, ProcessResponse, ProcessQueryParams } from './process';
import { AsyncData } from './common';

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