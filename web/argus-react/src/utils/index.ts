/**
 * Barrel export file for utility functions
 * This file re-exports all utility functions from the various utility files
 * to make imports cleaner and more consistent
 */

// Type guards
export * from './typeGuards';

// Error handling
export * from './errorHandling';

// Validation
export * from './validation';

/**
 * Format bytes to a human-readable string
 * @param bytes - The number of bytes
 * @param decimals - The number of decimal places to show
 * @returns A human-readable string representation of the bytes
 */
export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

/**
 * Format a number as a percentage
 * @param value - The value to format
 * @param decimals - The number of decimal places to show
 * @returns A string representation of the value as a percentage
 */
export function formatPercent(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

/**
 * Format a date to a human-readable string
 * @param date - The date to format (Date object or ISO string)
 * @param includeTime - Whether to include the time
 * @returns A human-readable string representation of the date
 */
export function formatDate(date: Date | string, includeTime = false): string {
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  
  const options: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    ...(includeTime ? { hour: '2-digit', minute: '2-digit' } : {})
  };
  
  return dateObj.toLocaleDateString(undefined, options);
}

/**
 * Generate a unique ID
 * @returns A unique ID string
 */
export function generateId(): string {
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}

/**
 * Debounce a function
 * @param fn - The function to debounce
 * @param delay - The delay in milliseconds
 * @returns A debounced version of the function
 */
export function debounce<T extends (...args: any[]) => any>(
  fn: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout> | null = null;
  
  return function(this: any, ...args: Parameters<T>): void {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    
    timeoutId = setTimeout(() => {
      fn.apply(this, args);
      timeoutId = null;
    }, delay);
  };
}

/**
 * Throttle a function
 * @param fn - The function to throttle
 * @param limit - The limit in milliseconds
 * @returns A throttled version of the function
 */
export function throttle<T extends (...args: any[]) => any>(
  fn: T,
  limit: number
): (...args: Parameters<T>) => void {
  let lastCall = 0;
  
  return function(this: any, ...args: Parameters<T>): void {
    const now = Date.now();
    
    if (now - lastCall >= limit) {
      fn.apply(this, args);
      lastCall = now;
    }
  };
} 