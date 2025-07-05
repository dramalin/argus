import { useProcessesContext } from '../context/ProcessesContext';
import type { ProcessQueryParams } from '../types/process';

/**
 * Hook for accessing process data from the ProcessesContext
 * Provides a simplified interface for components to use
 */
export function useProcesses() {
  const { 
    state, 
    refreshProcesses, 
    updateParams, 
    resetParams 
  } = useProcessesContext();
  
  return {
    processes: state.processes,
    total: state.total,
    loading: state.loading,
    error: state.error,
    lastUpdated: state.lastUpdated,
    params: state.params,
    refreshProcesses,
    handleParamChange: updateParams,
    getResetFilters: () => {
      resetParams();
      return state.params;
    }
  };
}

export default useProcesses; 