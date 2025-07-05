// API Client for Argus System Monitor
import type {
  CPUInfo,
  MemoryInfo,
  NetworkInfo,
  ProcessInfo,
  TaskInfo,
  TaskExecution,
  AlertInfo,
  HealthStatus,
  ApiResponse,
  SystemMetrics
} from './types/api';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class ArgusApiClient {
  private baseURL: string;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  // TODO: This method is not directly used except by other API methods
  // Consider making it private or refactoring if API design changes
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    
    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        ...options,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      
      // Handle both direct data and wrapped responses
      if (data && typeof data === 'object' && 'success' in data) {
        return data;
      }
      
      return {
        success: true,
        data: data
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  // System Metrics APIs
  // TODO: Not currently used directly in UI, only through getAllMetrics
  // Consider removal if not needed for future development
  async getCPU(): Promise<ApiResponse<CPUInfo>> {
    return this.request<CPUInfo>('/api/cpu');
  }

  // TODO: Not currently used directly in UI, only through getAllMetrics
  // Consider removal if not needed for future development
  async getMemory(): Promise<ApiResponse<MemoryInfo>> {
    return this.request<MemoryInfo>('/api/memory');
  }

  // TODO: Not currently used directly in UI, only through getAllMetrics
  // Consider removal if not needed for future development
  async getNetwork(): Promise<ApiResponse<NetworkInfo>> {
    return this.request<NetworkInfo>('/api/network');
  }

  // TODO: Not currently used directly in UI, only through getAllMetrics
  // Consider removal if not needed for future development
  async getProcesses(): Promise<ApiResponse<ProcessInfo[]>> {
    const response = await this.request<any>('/api/process');
    if (response.success && response.data && Array.isArray(response.data.processes)) {
      return {
        success: true,
        data: response.data.processes
      };
    } else if (!response.success) {
      return response;
    } else {
      return {
        success: false,
        error: 'Invalid process response format',
      };
    }
  }

  async getAllMetrics(): Promise<ApiResponse<SystemMetrics>> {
    try {
      const [cpu, memory, network, processes] = await Promise.all([
        this.getCPU(),
        this.getMemory(),
        this.getNetwork(),
        this.getProcesses()
      ]);

      if (!cpu.success || !memory.success || !network.success || !processes.success) {
        return {
          success: false,
          error: 'Failed to fetch one or more metrics'
        };
      }

      return {
        success: true,
        data: {
          cpu: cpu.data!,
          memory: memory.data!,
          network: network.data!,
          processes: processes.data!,
          timestamp: new Date().toISOString()
        }
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to fetch metrics'
      };
    }
  }

  // Task Management APIs
  // TODO: Not currently used in UI, consider removal if no future Task features planned
  async getTasks(): Promise<ApiResponse<TaskInfo[]>> {
    return this.request<TaskInfo[]>('/api/tasks');
  }

  // TODO: Not currently used in UI, consider removal if no future Task features planned
  async getTask(id: string): Promise<ApiResponse<TaskInfo>> {
    return this.request<TaskInfo>(`/api/tasks/${id}`);
  }

  // TODO: Not currently used in UI, consider removal if no future Task features planned
  async createTask(task: Partial<TaskInfo>): Promise<ApiResponse<TaskInfo>> {
    return this.request<TaskInfo>('/api/tasks', {
      method: 'POST',
      body: JSON.stringify(task),
    });
  }

  // TODO: Not currently used in UI, consider removal if no future Task features planned
  async updateTask(id: string, task: Partial<TaskInfo>): Promise<ApiResponse<TaskInfo>> {
    return this.request<TaskInfo>(`/api/tasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(task),
    });
  }

  // TODO: Not currently used in UI, consider removal if no future Task features planned
  async deleteTask(id: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/api/tasks/${id}`, {
      method: 'DELETE',
    });
  }

  // TODO: Not currently used in UI, consider removal if no future Task features planned
  async runTask(id: string): Promise<ApiResponse<TaskExecution>> {
    return this.request<TaskExecution>(`/api/tasks/${id}/run`, {
      method: 'POST',
    });
  }

  // TODO: Not currently used in UI, consider removal if no future Task features planned
  async getTaskExecutions(id: string): Promise<ApiResponse<TaskExecution[]>> {
    return this.request<TaskExecution[]>(`/api/tasks/${id}/executions`);
  }

  // Alert Management APIs
  // TODO: Not currently used in UI, consider removal if no future Alert features planned
  async getAlerts(): Promise<ApiResponse<AlertInfo[]>> {
    return this.request<AlertInfo[]>('/api/alerts');
  }

  // TODO: Not currently used in UI, consider removal if no future Alert features planned
  async getAlert(id: string): Promise<ApiResponse<AlertInfo>> {
    return this.request<AlertInfo>(`/api/alerts/${id}`);
  }

  // TODO: Not currently used in UI, consider removal if no future Alert features planned
  async createAlert(alert: Partial<AlertInfo>): Promise<ApiResponse<AlertInfo>> {
    return this.request<AlertInfo>('/api/alerts', {
      method: 'POST',
      body: JSON.stringify(alert),
    });
  }

  // TODO: Not currently used in UI, consider removal if no future Alert features planned
  async updateAlert(id: string, alert: Partial<AlertInfo>): Promise<ApiResponse<AlertInfo>> {
    return this.request<AlertInfo>(`/api/alerts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(alert),
    });
  }

  // TODO: Not currently used in UI, consider removal if no future Alert features planned
  async deleteAlert(id: string): Promise<ApiResponse<void>> {
    return this.request<void>(`/api/alerts/${id}`, {
      method: 'DELETE',
    });
  }

  // Health Check API
  // TODO: Not currently used in UI, consider removal if health check not planned for dashboard
  async getHealth(): Promise<ApiResponse<HealthStatus>> {
    return this.request<HealthStatus>('/health');
  }
}

export const apiClient = new ArgusApiClient();
export default apiClient;