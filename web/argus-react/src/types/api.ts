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
 * Schedule information for tasks
 */
export interface Schedule {
  /** Cron expression for recurring tasks */
  cron_expression: string;
  /** Whether this is a one-time task */
  one_time: boolean;
  /** Next scheduled execution time (ISO format) */
  next_run_time: string;
}

/**
 * Task information returned by the API
 */
export interface TaskInfo {
  /** Unique identifier for the task */
  id: string;
  /** Task name */
  name: string;
  /** Optional task description */
  description?: string;
  /** Task type (e.g., 'log_rotation', 'metrics_aggregation') */
  type: string;
  /** Whether the task is enabled */
  enabled: boolean;
  /** Schedule information */
  schedule: Schedule;
  /** Task-specific parameters */
  parameters?: Record<string, string>;
  /** Current status of the task (not in backend, but needed for UI) */
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
 * Metric type for alert monitoring
 */
export type MetricType = 'cpu' | 'memory' | 'load' | 'network' | 'disk' | 'process';

/**
 * Comparison operator for alert thresholds
 */
export type ComparisonOperator = '>' | '>=' | '<' | '<=' | '==' | '!=';

/**
 * Alert severity levels
 */
export type AlertSeverity = 'info' | 'warning' | 'critical';

/**
 * Notification channel types
 */
export type NotificationType = 'in-app' | 'email';

/**
 * Alert state
 */
export type AlertState = 'active' | 'inactive' | 'pending' | 'resolved';

/**
 * Threshold configuration for alerts
 */
export interface ThresholdConfig {
  /** Type of metric to monitor */
  metric_type: MetricType;
  /** Specific metric name within the type */
  metric_name: string;
  /** Comparison operator for threshold evaluation */
  operator: ComparisonOperator;
  /** Threshold value */
  value: number;
  /** Duration for sustained condition (milliseconds) */
  duration?: number;
  /** Number of consecutive evaluations required */
  sustained_for?: number;
}

/**
 * Notification configuration for alerts
 */
export interface NotificationConfig {
  /** Type of notification */
  type: NotificationType;
  /** Whether this notification is enabled */
  enabled: boolean;
  /** Additional settings for the notification */
  settings?: Record<string, any>;
}

/**
 * Alert configuration
 */
export interface AlertConfig {
  /** Unique identifier for the alert */
  id: string;
  /** Name of the alert */
  name: string;
  /** Optional description of the alert */
  description?: string;
  /** Whether the alert is enabled */
  enabled: boolean;
  /** Alert severity level */
  severity: AlertSeverity;
  /** Timestamp when the alert was created */
  created_at: string;
  /** Timestamp when the alert was last updated */
  updated_at?: string;
  /** Timestamp when the alert was last triggered */
  triggered_at?: string;
  /** Alert threshold configuration */
  threshold: ThresholdConfig;
  /** Alert notification settings */
  notifications: NotificationConfig[];
}

/**
 * Alert status information
 */
export interface AlertStatus {
  /** ID of the alert */
  alert_id: string;
  /** Current state of the alert */
  state: AlertState;
  /** Current value of the monitored metric */
  current_value: number;
  /** When the alert was triggered (ISO format), if applicable */
  triggered_at?: string;
  /** When the alert was resolved (ISO format), if applicable */
  resolved_at?: string;
  /** Status message */
  message?: string;
}

/**
 * Alert notification
 */
export interface AlertNotification {
  /** Unique identifier for the notification */
  id: string;
  /** ID of the alert that triggered this notification */
  alert_id: string;
  /** Alert name */
  alert_name: string;
  /** Notification message */
  message: string;
  /** Notification timestamp (ISO format) */
  timestamp: string;
  /** Whether the notification has been read */
  read: boolean;
  /** Alert severity level */
  severity: AlertSeverity;
}

/**
 * Alert information returned by the API
 * @deprecated Use AlertConfig instead for complete alert information
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
 * Alert test event response
 */
export interface AlertTestEvent {
  /** ID of the alert */
  alert_id: string;
  /** Previous state */
  old_state: AlertState;
  /** New state */
  new_state: AlertState;
  /** Current value of the monitored metric */
  current_value: number;
  /** Threshold value */
  threshold: number;
  /** Event timestamp (ISO format) */
  timestamp: string;
  /** Event message */
  message: string;
  /** Alert configuration */
  alert: AlertConfig;
  /** Alert status */
  status: AlertStatus;
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

// No need to re-export types that are already properly exported
// The issue is likely with TypeScript configuration, not with the exports themselves