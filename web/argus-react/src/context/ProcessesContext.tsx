import React, { createContext, useContext, useReducer, useEffect, ReactNode } from 'react';
import { apiClient } from '../api';
import type { ProcessInfo, ProcessQueryParams, ProcessResponse } from '../types/process';

// Define the state interface
interface ProcessesState {
  processes: ProcessInfo[];
  total: number;
  loading: boolean;
  error: string | null;
  lastUpdated: string | null;
  params: ProcessQueryParams;
}

// Define action types
type ProcessesAction =
  | { type: 'FETCH_PROCESSES_START' }
  | { type: 'FETCH_PROCESSES_SUCCESS'; payload: { processes: ProcessInfo[]; total: number; updated: string } }
  | { type: 'FETCH_PROCESSES_ERROR'; payload: string }
  | { type: 'UPDATE_PARAMS'; payload: ProcessQueryParams }
  | { type: 'RESET_PARAMS' };

// Define the context interface
interface ProcessesContextType {
  state: ProcessesState;
  dispatch: React.Dispatch<ProcessesAction>;
  refreshProcesses: () => Promise<void>;
  updateParams: (key: keyof ProcessQueryParams, value: any) => void;
  resetParams: () => void;
}

// Default query parameters
const defaultParams: ProcessQueryParams = {
  limit: 10,
  offset: 0,
  sort_by: 'cpu',
  sort_order: 'desc',
  name_contains: '',
  min_cpu: undefined,
  min_memory: undefined,
};

// Create the context
const ProcessesContext = createContext<ProcessesContextType | undefined>(undefined);

// Initial state
const initialState: ProcessesState = {
  processes: [],
  total: 0,
  loading: true,
  error: null,
  lastUpdated: null,
  params: defaultParams,
};

// Reducer function
function processesReducer(state: ProcessesState, action: ProcessesAction): ProcessesState {
  switch (action.type) {
    case 'FETCH_PROCESSES_START':
      return {
        ...state,
        loading: true,
        error: null,
      };
    case 'FETCH_PROCESSES_SUCCESS':
      return {
        ...state,
        processes: action.payload.processes,
        total: action.payload.total,
        loading: false,
        error: null,
        lastUpdated: action.payload.updated,
      };
    case 'FETCH_PROCESSES_ERROR':
      return {
        ...state,
        loading: false,
        error: action.payload,
      };
    case 'UPDATE_PARAMS':
      return {
        ...state,
        params: action.payload,
      };
    case 'RESET_PARAMS':
      return {
        ...state,
        params: defaultParams,
      };
    default:
      return state;
  }
}

// Provider component
interface ProcessesProviderProps {
  children: ReactNode;
  pollingInterval?: number;
  initialParams?: ProcessQueryParams;
}

export const ProcessesProvider: React.FC<ProcessesProviderProps> = ({
  children,
  pollingInterval = 5000, // Default to 5 seconds
  initialParams = defaultParams,
}) => {
  const [state, dispatch] = useReducer(processesReducer, {
    ...initialState,
    params: initialParams,
  });

  // Function to fetch processes
  const fetchProcesses = async () => {
    dispatch({ type: 'FETCH_PROCESSES_START' });
    
    try {
      const response = await apiClient.getProcesses(state.params);
      
      if (response.success && response.data) {
        const { processes, total_count, updated_at } = response.data;
        
        dispatch({
          type: 'FETCH_PROCESSES_SUCCESS',
          payload: {
            processes,
            total: total_count,
            updated: updated_at,
          },
        });
      } else {
        dispatch({
          type: 'FETCH_PROCESSES_ERROR',
          payload: response.error || 'Failed to fetch processes',
        });
      }
    } catch (err) {
      dispatch({
        type: 'FETCH_PROCESSES_ERROR',
        payload: err instanceof Error ? err.message : 'Unknown error',
      });
    }
  };

  // Expose a refresh function
  const refreshProcesses = async () => {
    await fetchProcesses();
  };

  // Function to update params
  const updateParams = (key: keyof ProcessQueryParams, value: any) => {
    const newParams = {
      ...state.params,
      [key]: value,
      // Reset offset when changing filters
      offset: key !== 'offset' ? 0 : value,
    };
    
    dispatch({ type: 'UPDATE_PARAMS', payload: newParams });
  };

  // Function to reset params
  const resetParams = () => {
    dispatch({ type: 'RESET_PARAMS' });
  };

  // Fetch processes when params change or on initial load
  useEffect(() => {
    fetchProcesses();
    
    // Set up interval if polling is enabled
    if (pollingInterval > 0) {
      const intervalId = setInterval(fetchProcesses, pollingInterval);
      
      // Clean up on unmount
      return () => clearInterval(intervalId);
    }
    
    return undefined;
  }, [state.params, pollingInterval]);

  return (
    <ProcessesContext.Provider value={{ state, dispatch, refreshProcesses, updateParams, resetParams }}>
      {children}
    </ProcessesContext.Provider>
  );
};

// Custom hook for using the processes context
export const useProcessesContext = () => {
  const context = useContext(ProcessesContext);
  
  if (context === undefined) {
    throw new Error('useProcessesContext must be used within a ProcessesProvider');
  }
  
  return context;
};

export default ProcessesContext; 