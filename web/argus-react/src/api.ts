// API Client for Argus System Monitor
import type {
  CPUInfo,
  MemoryInfo,
  NetworkInfo,
  TaskInfo,
  TaskExecution,
  AlertInfo,
  HealthStatus,
  ApiResponse,
  SystemMetrics,
} from './types/api';
import type {
  ProcessInfo,
  ProcessQueryParams,
  ProcessResponse
} from './types/process';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const API_TIMEOUT = 10000; // 10 seconds timeout

/**
 * Custom error class for API request timeouts
 */
class RequestTimeoutError extends Error {
  constructor(message = 'Request timed out') {
    super(message);
    this.name = 'RequestTimeoutError';
  }
}

/**
 * Argus API Client for interacting with the backend services
 */
class ArgusApiClient {
  private baseURL: string;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  /**
   * Generic request method with timeout and error handling
   */
  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    timeout: number = API_TIMEOUT
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    
    // Create abort controller for timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);
    
    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        signal: controller.signal,
        ...options,
      });

      // Clear timeout since request completed
      clearTimeout(timeoutId);

      if (!response.ok) {
        // Enhanced error handling with status codes
        const errorText = await response.text().catch(() => 'No error details available');
        let errorMessage = `HTTP error! status: ${response.status}`;
        
        switch (response.status) {
          case 400:
            errorMessage = `Bad request: ${errorText}`;
            break;
          case 401:
            errorMessage = 'Authentication required';
            break;
          case 403:
            errorMessage = 'Access forbidden';
            break;
          case 404:
            errorMessage = `Resource not found: ${endpoint}`;
            break;
          case 429:
            errorMessage = 'Too many requests, please try again later';
            break;
          case 500:
            errorMessage = 'Internal server error';
            break;
          case 503:
            errorMessage = 'Service unavailable, please try again later';
            break;
        }
        
        throw new Error(errorMessage);
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
      // Handle abort error as timeout
      if (error instanceof DOMException && error.name === 'AbortError') {
        return {
          success: false,
          error: 'Request timed out. Please try again.',
        };
      }
      
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    } finally {
      clearTimeout(timeoutId);
    }
  }

  /**
   * Get processes with filtering and pagination
   */
  async getProcesses(params?: ProcessQueryParams): Promise<ApiResponse<ProcessResponse>> {
    let query = '';
    if (params) {
      const searchParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          searchParams.append(key, String(value));
        }
      });
      query = '?' + searchParams.toString();
    }
    return this.request<ProcessResponse>(`/api/process${query}`);
  }

  /**
   * Get all system metrics in a single call
   * First tries the unified endpoint, falls back to individual calls if needed
   */
  async getAllMetrics(): Promise<ApiResponse<SystemMetrics>> {
    try {
      // Try to use the unified endpoint first
      const unifiedResponse = await this.request<SystemMetrics>('/api/metrics');
      
      // If unified endpoint works, return the data
      if (unifiedResponse.success && unifiedResponse.data) {
        return unifiedResponse;
      }
      
      // Fall back to individual calls if unified endpoint fails or doesn't exist
      const [cpu, memory, network, processes] = await Promise.all([
        this.request<CPUInfo>('/api/cpu'),
        this.request<MemoryInfo>('/api/memory'),
        this.request<NetworkInfo>('/api/network'),
        this.getProcesses()
      ]);

      if (!cpu.success || !memory.success || !network.success || !processes.success) {
        const errors = [];
        if (!cpu.success) errors.push(`CPU: ${cpu.error}`);
        if (!memory.success) errors.push(`Memory: ${memory.error}`);
        if (!network.success) errors.push(`Network: ${network.error}`);
        if (!processes.success) errors.push(`Processes: ${processes.error}`);
        
        return {
          success: false,
          error: `Failed to fetch metrics: ${errors.join(', ')}`,
        };
      }

      return {
        success: true,
        data: {
          cpu: cpu.data!,
          memory: memory.data!,
          network: network.data!,
          processes: processes.data!.processes,
          timestamp: new Date().toISOString(),
        },
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error 
          ? `Failed to fetch metrics: ${error.message}` 
          : 'Failed to fetch metrics: Unknown error',
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