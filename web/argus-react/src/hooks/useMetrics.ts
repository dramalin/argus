import { useMetricsContext } from '../context/MetricsContext';

/**
 * Hook for accessing metrics data from the MetricsContext
 * Provides a simplified interface for components to use
 */
export function useMetrics() {
  const { state, refreshMetrics } = useMetricsContext();
  
  return {
    metrics: state.metrics,
    cpuHistory: state.cpuHistory,
    loading: state.loading,
    error: state.error,
    lastUpdated: state.lastUpdated,
    refreshMetrics,
  };
}

export default useMetrics; 