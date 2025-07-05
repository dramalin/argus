import { useState, useEffect } from 'react';
import useApiCache from './useApiCache';
import type { ApiResponse } from '../types/api';
import type { DataFetchingOptions, UseDataFetchingResult } from '../types/hooks';

/**
 * A simplified wrapper around useApiCache that provides a common data fetching pattern
 * with automatic timestamp handling
 * 
 * @template T The type of data to fetch
 * @param key Cache key for the data
 * @param fetchFn Function that returns a Promise with the data
 * @param options Optional configuration options
 * @returns Object with data, loading state, error, lastUpdated timestamp, and refetch function
 */
export function useDataFetching<T>(
  key: string,
  fetchFn: () => Promise<ApiResponse<T>>,
  options: DataFetchingOptions = {}
): UseDataFetchingResult<T> {
  const {
    initialLoading = true,
    cacheTTL = 30000, // 30 seconds default
    fetchOnMount = true,
  } = options;

  const [lastUpdated, setLastUpdated] = useState<string | null>(null);
  
  // Use the existing useApiCache hook
  const { data, error, loading, refetch } = useApiCache<T>(key, fetchFn, {
    ttl: cacheTTL,
  });
  
  // Update lastUpdated timestamp whenever data changes
  useEffect(() => {
    if (data) {
      setLastUpdated(new Date().toISOString());
    }
  }, [data]);
  
  // Enhanced refetch function that updates the timestamp
  const refetchWithTimestamp = async (): Promise<void> => {
    await refetch();
    setLastUpdated(new Date().toISOString());
  };
  
  return {
    data,
    loading,
    error,
    lastUpdated,
    refetch: refetchWithTimestamp,
  };
}

export default useDataFetching; 