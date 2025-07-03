// API Types for Argus System Monitor

export interface CPUInfo {
  load1: number;
  load5: number;
  load15: number;
  usage_percent: number;
}

export interface MemoryInfo {
  total: number;
  used: number;
  free: number;
  used_percent: number;
}

export interface NetworkInfo {
  bytes_sent: number;
  bytes_recv: number;
  packets_sent: number;
  packets_recv: number;
}

export interface ProcessInfo {
  pid: number;
  name: string;
  cpu_percent: number;
  mem_percent: number;
}

export interface TaskInfo {
  id: string;
  name: string;
  type: string;
  enabled: boolean;
  cron_expression?: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  created_at: string;
  updated_at: string;
}

export interface TaskExecution {
  id: string;
  task_id: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  start_time: string;
  end_time?: string;
  output?: string;
  error?: string;
}

export interface AlertInfo {
  id: string;
  name: string;
  type: string;
  enabled: boolean;
  conditions: Record<string, any>;
  actions: Record<string, any>;
  created_at: string;
  triggered_at?: string;
}

export interface HealthStatus {
  status: 'healthy' | 'unhealthy';
  timestamp: string;
  version?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface SystemMetrics {
  cpu: CPUInfo;
  memory: MemoryInfo;
  network: NetworkInfo;
  processes: ProcessInfo[];
  timestamp: string;
}

// WebSocket message types
export interface WebSocketMessage {
  type: 'metrics' | 'alert' | 'task_update' | 'error';
  data: any;
  timestamp: string;
} 