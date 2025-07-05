import { useState, useEffect, useCallback, useRef } from 'react';
import type { ApiResponse } from '../types/api';

interface CacheEntry<T> {
  data: ApiResponse<T>;
  timestamp: number;
  key: string;
}

interface CacheOptions {
  ttl?: number; // Time to live in milliseconds
  dedupingInterval?: number; // Minimum time between requests in milliseconds
  retries?: number; // Number of retries on failure
  retryDelay?: number; // Delay between retries in milliseconds
  onSuccess?: (data: any) => void;
  onError?: (error: string) => void;
}

const DEFAULT_OPTIONS: CacheOptions = {
  ttl: 30000, // 30 seconds cache
  dedupingInterval: 2000, // 2 seconds deduping
  retries: 2,
  retryDelay: 1000,
};

// In-memory cache store
const cacheStore = new Map<string, CacheEntry<any>>();

export function useApiCache<T>(
  key: string,
  fetchFn: () => Promise<ApiResponse<T>>,
  options: CacheOptions = {}
): {
  data: T | null;
  error: string | null;
  loading: boolean;
  refetch: () => Promise<void>;
} {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  
  // Merge default options with provided options
  const mergedOptions = { ...DEFAULT_OPTIONS, ...options };
  
  // Use refs for values that shouldn't trigger re-renders
  const fetchingRef = useRef<boolean>(false);
  const lastFetchTimeRef = useRef<number>(0);
  const retryCountRef = useRef<number>(0);
  
  // Function to check if cache is valid
  const isCacheValid = useCallback((entry: CacheEntry<T>): boolean => {
    return Date.now() - entry.timestamp < (mergedOptions.ttl || 0);
  }, [mergedOptions.ttl]);

  // Main fetch function
  const fetchData = useCallback(async (force = false): Promise<void> => {
    // Skip if already fetching (deduping)
    if (fetchingRef.current) {
      return;
    }
    
    // Check deduping interval
    const now = Date.now();
    if (!force && now - lastFetchTimeRef.current < (mergedOptions.dedupingInterval || 0)) {
      return;
    }
    
    // Check cache first
    const cachedEntry = cacheStore.get(key) as CacheEntry<T> | undefined;
    if (!force && cachedEntry && isCacheValid(cachedEntry)) {
      if (cachedEntry.data.success && cachedEntry.data.data) {
        setData(cachedEntry.data.data);
        setError(null);
        setLoading(false);
        return;
      }
    }
    
    // Set loading state and mark as fetching
    setLoading(true);
    fetchingRef.current = true;
    lastFetchTimeRef.current = now;
    
    try {
      const response = await fetchFn();
      
      if (response.success && response.data) {
        // Cache successful response
        cacheStore.set(key, {
          data: response,
          timestamp: Date.now(),
          key,
        });
        
        setData(response.data);
        setError(null);
        retryCountRef.current = 0;
        
        if (mergedOptions.onSuccess) {
          mergedOptions.onSuccess(response.data);
        }
      } else {
        const errorMessage = response.error || 'Unknown error';
        setError(errorMessage);
        
        // Retry logic
        if (retryCountRef.current < (mergedOptions.retries || 0)) {
          retryCountRef.current++;
          setTimeout(() => {
            fetchingRef.current = false;
            fetchData();
          }, mergedOptions.retryDelay || 1000);
          return;
        }
        
        if (mergedOptions.onError) {
          mergedOptions.onError(errorMessage);
        }
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      
      if (mergedOptions.onError) {
        mergedOptions.onError(errorMessage);
      }
    } finally {
      setLoading(false);
      fetchingRef.current = false;
    }
  }, [key, fetchFn, isCacheValid, mergedOptions]);
  
  // Expose refetch function for manual refetching
  const refetch = useCallback(async (): Promise<void> => {
    await fetchData(true);
  }, [fetchData]);
  
  // Initial fetch on mount or when key changes
  useEffect(() => {
    fetchData();
    
    // Optional cleanup
    return () => {
      fetchingRef.current = false;
    };
  }, [key, fetchData]);
  
  return { data, error, loading, refetch };
}

// Utility function to clear entire cache or specific keys
export function clearApiCache(specificKeys?: string[]): void {
  if (!specificKeys || specificKeys.length === 0) {
    cacheStore.clear();
  } else {
    specificKeys.forEach(key => cacheStore.delete(key));
  }
}

// Utility function to get cache stats
export function getApiCacheStats(): { size: number; keys: string[] } {
  return {
    size: cacheStore.size,
    keys: Array.from(cacheStore.keys()),
  };
}

export default useApiCache; 