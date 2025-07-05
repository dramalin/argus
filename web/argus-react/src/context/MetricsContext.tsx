import React, { createContext, useContext, useReducer, useEffect, ReactNode } from 'react';
import { apiClient } from '../api';
import type { SystemMetrics } from '../types/api';

// Define the state interface
interface MetricsState {
  metrics: SystemMetrics | null;
  cpuHistory: { value: number; timestamp: string }[];
  loading: boolean;
  error: string | null;
  lastUpdated: string | null;
}

// Define action types
type MetricsAction =
  | { type: 'FETCH_METRICS_START' }
  | { type: 'FETCH_METRICS_SUCCESS'; payload: SystemMetrics }
  | { type: 'FETCH_METRICS_ERROR'; payload: string }
  | { type: 'UPDATE_CPU_HISTORY'; payload: { value: number; timestamp: string } }
  | { type: 'RESET_METRICS' };

// Define the context interface
interface MetricsContextType {
  state: MetricsState;
  dispatch: React.Dispatch<MetricsAction>;
  refreshMetrics: () => Promise<void>;
}

// Maximum number of data points to keep in history
const MAX_HISTORY_POINTS = 20;

// Create the context
const MetricsContext = createContext<MetricsContextType | undefined>(undefined);

// Initial state
const initialState: MetricsState = {
  metrics: null,
  cpuHistory: [],
  loading: true,
  error: null,
  lastUpdated: null,
};

// Reducer function
function metricsReducer(state: MetricsState, action: MetricsAction): MetricsState {
  switch (action.type) {
    case 'FETCH_METRICS_START':
      return {
        ...state,
        loading: true,
        error: null,
      };
    case 'FETCH_METRICS_SUCCESS':
      return {
        ...state,
        metrics: action.payload,
        loading: false,
        error: null,
        lastUpdated: new Date().toISOString(),
      };
    case 'FETCH_METRICS_ERROR':
      return {
        ...state,
        loading: false,
        error: action.payload,
      };
    case 'UPDATE_CPU_HISTORY':
      return {
        ...state,
        cpuHistory: [...state.cpuHistory.slice(-MAX_HISTORY_POINTS + 1), action.payload],
      };
    case 'RESET_METRICS':
      return initialState;
    default:
      return state;
  }
}

// Provider component
interface MetricsProviderProps {
  children: ReactNode;
  pollingInterval?: number;
}

export const MetricsProvider: React.FC<MetricsProviderProps> = ({
  children,
  pollingInterval = 5000, // Default to 5 seconds
}) => {
  const [state, dispatch] = useReducer(metricsReducer, initialState);

  // Function to fetch metrics
  const fetchMetrics = async () => {
    dispatch({ type: 'FETCH_METRICS_START' });
    
    try {
      const response = await apiClient.getAllMetrics();
      
      if (response.success && response.data) {
        dispatch({ type: 'FETCH_METRICS_SUCCESS', payload: response.data });
        
        // Update CPU history
        dispatch({
          type: 'UPDATE_CPU_HISTORY',
          payload: {
            timestamp: new Date().toLocaleTimeString(),
            value: response.data.cpu.usage_percent
          }
        });
      } else {
        dispatch({
          type: 'FETCH_METRICS_ERROR',
          payload: response.error || 'Failed to fetch metrics'
        });
      }
    } catch (err) {
      dispatch({
        type: 'FETCH_METRICS_ERROR',
        payload: err instanceof Error ? err.message : 'Unknown error'
      });
    }
  };

  // Expose a refresh function
  const refreshMetrics = async () => {
    await fetchMetrics();
  };

  // Set up polling
  useEffect(() => {
    // Initial fetch
    fetchMetrics();
    
    // Set up interval if polling is enabled
    if (pollingInterval > 0) {
      const intervalId = setInterval(fetchMetrics, pollingInterval);
      
      // Clean up on unmount
      return () => clearInterval(intervalId);
    }
    
    return undefined;
  }, [pollingInterval]);

  return (
    <MetricsContext.Provider value={{ state, dispatch, refreshMetrics }}>
      {children}
    </MetricsContext.Provider>
  );
};

// Custom hook for using the metrics context
export const useMetricsContext = () => {
  const context = useContext(MetricsContext);
  
  if (context === undefined) {
    throw new Error('useMetricsContext must be used within a MetricsProvider');
  }
  
  return context;
};

export default MetricsContext; 