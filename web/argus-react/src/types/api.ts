/**
 * API Types for Argus System Monitor
 * This file contains all the type definitions for the API responses and requests
 */
import type { ProcessInfo } from './process';
import type { SortDirection } from './common';

/**
 * CPU information returned by the API
 */
export interface CPUInfo {
  /** Average CPU load over the last 1 minute */
  load1: number;
  /** Average CPU load over the last 5 minutes */
  load5: number;
  /** Average CPU load over the last 15 minutes */
  load15: number;
  /** Current CPU usage percentage (0-100) */
  usage_percent: number;
}

/**
 * Memory information returned by the API
 */
export interface MemoryInfo {
  /** Total memory in bytes */
  total: number;
  /** Used memory in bytes */
  used: number;
  /** Free memory in bytes */
  free: number;
  /** Memory usage percentage (0-100) */
  used_percent: number;
}

/**
 * Network information returned by the API
 */
export interface NetworkInfo {
  /** Total bytes sent since system boot */
  bytes_sent: number;
  /** Total bytes received since system boot */
  bytes_recv: number;
  /** Total packets sent since system boot */
  packets_sent: number;
  /** Total packets received since system boot */
  packets_recv: number;
}

/**
 * Task status type
 */
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed';

/**
 * Task information returned by the API
 */
export interface TaskInfo {
  /** Unique identifier for the task */
  id: string;
  /** Task name */
  name: string;
  /** Task type (e.g., 'cron', 'oneshot') */
  type: string;
  /** Whether the task is enabled */
  enabled: boolean;
  /** Cron expression for scheduled tasks */
  cron_expression?: string;
  /** Current status of the task */
  status: TaskStatus;
  /** Creation timestamp (ISO format) */
  created_at: string;
  /** Last update timestamp (ISO format) */
  updated_at: string;
}

/**
 * Task execution information returned by the API
 */
export interface TaskExecution {
  /** Unique identifier for the execution */
  id: string;
  /** ID of the task that was executed */
  task_id: string;
  /** Execution status */
  status: TaskStatus;
  /** Start time of the execution (ISO format) */
  start_time: string;
  /** End time of the execution (ISO format), if completed */
  end_time?: string;
  /** Execution output, if any */
  output?: string;
  /** Error message, if failed */
  error?: string;
}

/**
 * Alert information returned by the API
 */
export interface AlertInfo {
  /** Unique identifier for the alert */
  id: string;
  /** Alert name */
  name: string;
  /** Alert type (e.g., 'threshold', 'anomaly') */
  type: string;
  /** Whether the alert is enabled */
  enabled: boolean;
  /** Alert conditions */
  conditions: Record<string, unknown>;
  /** Actions to take when alert is triggered */
  actions: Record<string, unknown>;
  /** Creation timestamp (ISO format) */
  created_at: string;
  /** Last triggered timestamp (ISO format), if ever triggered */
  triggered_at?: string;
}

/**
 * Health status type
 */
export type HealthStatusType = 'healthy' | 'unhealthy';

/**
 * Health status information returned by the API
 */
export interface HealthStatus {
  /** Current health status */
  status: HealthStatusType;
  /** Timestamp of the health check (ISO format) */
  timestamp: string;
  /** Application version */
  version?: string;
}

/**
 * Generic API response wrapper
 * @template T The type of data contained in the response
 */
export interface ApiResponse<T> {
  /** Whether the request was successful */
  success: boolean;
  /** Response data, if successful */
  data?: T;
  /** Error message, if unsuccessful */
  error?: string;
  /** Additional message */
  message?: string;
}

/**
 * System metrics information returned by the API
 */
export interface SystemMetrics {
  /** CPU information */
  cpu: CPUInfo;
  /** Memory information */
  memory: MemoryInfo;
  /** Network information */
  network: NetworkInfo;
  /** Process information */
  processes: ProcessInfo[];
  /** Timestamp of the metrics (ISO format) */
  timestamp: string;
}

/**
 * WebSocket message types
 * @deprecated This interface is currently unused in the codebase.
 * Consider removing it if WebSocket functionality is not planned for future releases.
 */
export interface WebSocketMessage {
  /** Message type */
  type: 'metrics' | 'alert' | 'task_update' | 'error';
  /** Message data */
  data: unknown;
  /** Timestamp of the message (ISO format) */
  timestamp: string;
}