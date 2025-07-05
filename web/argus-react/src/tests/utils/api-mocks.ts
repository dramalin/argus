import { vi } from 'vitest';
import axios from 'axios';
import { mockSystemMetrics, mockApiResponse } from './test-utils';

/**
 * Mock Axios for testing
 */
export const mockAxios = () => {
  vi.mock('axios');

  // Mock successful response
  const mockGet = vi.fn().mockResolvedValue(mockApiResponse);
  
  // Mock error response
  const mockGetError = vi.fn().mockRejectedValue({
    response: {
      status: 500,
      data: { message: 'Internal Server Error' },
    },
  });

  // Mock timeout error
  const mockGetTimeout = vi.fn().mockRejectedValue({
    code: 'ECONNABORTED',
    message: 'timeout of 5000ms exceeded',
  });

  // Reset mocks before each test
  beforeEach(() => {
    vi.resetAllMocks();
    axios.get = mockGet;
  });

  return {
    mockGet,
    mockGetError,
    mockGetTimeout,
    setMockImplementation: (implementation: 'success' | 'error' | 'timeout') => {
      if (implementation === 'success') {
        axios.get = mockGet;
      } else if (implementation === 'error') {
        axios.get = mockGetError;
      } else if (implementation === 'timeout') {
        axios.get = mockGetTimeout;
      }
    },
  };
};

/**
 * Mock API responses for specific endpoints
 */
export const mockApiEndpoints = () => {
  vi.mock('../api', async (importOriginal) => {
    const actual = await importOriginal() as Record<string, unknown>;
    return {
      ...(actual as object),
      getSystemMetrics: vi.fn().mockResolvedValue(mockSystemMetrics),
      getProcesses: vi.fn().mockResolvedValue({
        processes: mockSystemMetrics.processes,
        total: mockSystemMetrics.processes.length,
      }),
    };
  });
};

/**
 * Create a custom mock for the useMetrics hook
 */
export const mockUseMetrics = (options: {
  loading?: boolean;
  error?: string | null;
  metrics?: typeof mockSystemMetrics | null;
} = {}) => {
  const {
    loading = false,
    error = null,
    metrics = mockSystemMetrics,
  } = options;

  return {
    metrics: {
      data: metrics,
      loading,
      error,
    },
    refreshMetrics: vi.fn(),
    cpu: metrics?.cpu || null,
    memory: metrics?.memory || null,
    network: metrics?.network || null,
    lastUpdated: metrics ? new Date().toISOString() : null,
  };
};

/**
 * Create a custom mock for the useProcesses hook
 */
export const mockUseProcesses = (options: {
  loading?: boolean;
  error?: string | null;
  processes?: typeof mockSystemMetrics.processes;
  total?: number;
} = {}) => {
  const {
    loading = false,
    error = null,
    processes = mockSystemMetrics.processes,
    total = mockSystemMetrics.processes.length,
  } = options;

  return {
    processData: {
      data: {
        processes,
        total,
      },
      loading,
      error,
    },
    processParams: {
      limit: 10,
      offset: 0,
      sort_by: 'cpu',
      sort_order: 'desc',
    },
    setProcessParams: vi.fn(),
    resetProcessParams: vi.fn(),
    lastUpdated: processes ? new Date().toISOString() : null,
  };
}; 