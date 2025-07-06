import { useCallback, useRef, useEffect } from 'react';

/**
 * A hook that throttles a callback function
 * @param callback The callback function to throttle
 * @param delay The delay in milliseconds
 * @returns The throttled callback function
 */
function useThrottledCallback<T extends (...args: any[]) => any>(
  callback: T,
  delay: number = 200
): (...args: Parameters<T>) => void {
  const lastCall = useRef<number>(0);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);
  const lastArgsRef = useRef<Parameters<T> | null>(null);

  useEffect(() => {
    // Cleanup function to clear any pending timeout when callback or delay changes, or component unmounts
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
    };
  }, [callback, delay]); // Re-run effect when callback or delay changes

  return useCallback(
    (...args: Parameters<T>) => {
      const now = Date.now();
      const timeSinceLastCall = now - lastCall.current;
      
      // Store the latest arguments
      lastArgsRef.current = args;
      
      // If enough time has passed since the last call, execute immediately
      if (timeSinceLastCall >= delay) {
        if (timeoutRef.current) {
          clearTimeout(timeoutRef.current);
          timeoutRef.current = null;
        }
        lastCall.current = now;
        callback(...args);
        return;
      }
      
      // If a timeout is already set, clear it to ensure only the latest call is processed
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
      
      // Set a new timeout to execute after the remaining delay
      timeoutRef.current = setTimeout(() => {
        if (lastArgsRef.current) {
          callback(...lastArgsRef.current);
          lastCall.current = Date.now();
          lastArgsRef.current = null;
          timeoutRef.current = null;
        }
      }, delay - timeSinceLastCall);
    },
    [callback, delay]
  );
}

export default useThrottledCallback; 