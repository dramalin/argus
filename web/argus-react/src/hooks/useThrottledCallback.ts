import { useCallback, useRef } from 'react';

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

  return useCallback(
    (...args: Parameters<T>) => {
      const now = Date.now();
      const timeSinceLastCall = now - lastCall.current;
      
      // Store the latest arguments
      lastArgsRef.current = args;
      
      // If enough time has passed since the last call, execute immediately
      if (timeSinceLastCall >= delay) {
        lastCall.current = now;
        callback(...args);
        return;
      }
      
      // Otherwise, set a timeout to execute after the remaining delay
      if (timeoutRef.current === null) {
        timeoutRef.current = setTimeout(() => {
          if (lastArgsRef.current) {
            callback(...lastArgsRef.current);
            lastCall.current = Date.now();
            lastArgsRef.current = null;
            timeoutRef.current = null;
          }
        }, delay - timeSinceLastCall);
      }
    },
    [callback, delay]
  );
}

export default useThrottledCallback; 