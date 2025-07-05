import { useMemo } from 'react';
import type { DateFormatterOptions, UseDateFormatterResult } from '../types/hooks';

/**
 * Default date formatting options
 */
const DEFAULT_OPTIONS: Required<DateFormatterOptions> = {
  locale: 'en-US',
  defaultFormat: {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
  },
  invalidDateText: 'N/A',
};

/**
 * A custom hook that provides consistent date formatting utilities
 * 
 * @param options Configuration options for date formatting
 * @returns Object with date formatting functions
 */
export function useDateFormatter(options: DateFormatterOptions = {}): UseDateFormatterResult {
  // Merge default options with provided options
  const config = useMemo(() => ({
    ...DEFAULT_OPTIONS,
    ...options,
  }), [options]);

  /**
   * Format a date string to a localized string
   * @param dateString ISO date string or undefined/null
   * @param formatOptions Optional Intl.DateTimeFormatOptions
   * @returns Formatted date string or 'N/A' for invalid dates
   */
  const formatDate = useMemo(() => (
    (dateString?: string, formatOptions?: Intl.DateTimeFormatOptions): string => {
      if (!dateString) return config.invalidDateText;
      
      try {
        const date = new Date(dateString);
        
        // Check if date is valid
        if (isNaN(date.getTime())) {
          return config.invalidDateText;
        }
        
        return date.toLocaleString(
          config.locale, 
          formatOptions || config.defaultFormat
        );
      } catch (error) {
        return config.invalidDateText;
      }
    }
  ), [config]);

  /**
   * Format a timestamp (milliseconds) to a localized string
   * @param timestamp Timestamp in milliseconds
   * @param formatOptions Optional Intl.DateTimeFormatOptions
   * @returns Formatted date string or 'N/A' for invalid timestamps
   */
  const formatTimestamp = useMemo(() => (
    (timestamp: number, formatOptions?: Intl.DateTimeFormatOptions): string => {
      if (!timestamp) return config.invalidDateText;
      
      try {
        const date = new Date(timestamp);
        
        // Check if date is valid
        if (isNaN(date.getTime())) {
          return config.invalidDateText;
        }
        
        return date.toLocaleString(
          config.locale, 
          formatOptions || config.defaultFormat
        );
      } catch (error) {
        return config.invalidDateText;
      }
    }
  ), [config]);

  /**
   * Format a date string as a relative time (e.g., "2 hours ago")
   * @param dateString ISO date string
   * @returns Relative time string or 'N/A' for invalid dates
   */
  const formatRelativeTime = useMemo(() => (
    (dateString?: string): string => {
      if (!dateString) return config.invalidDateText;
      
      try {
        const date = new Date(dateString);
        
        // Check if date is valid
        if (isNaN(date.getTime())) {
          return config.invalidDateText;
        }
        
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffSec = Math.round(diffMs / 1000);
        const diffMin = Math.round(diffSec / 60);
        const diffHour = Math.round(diffMin / 60);
        const diffDay = Math.round(diffHour / 24);
        
        if (diffSec < 60) {
          return `${diffSec} second${diffSec !== 1 ? 's' : ''} ago`;
        } else if (diffMin < 60) {
          return `${diffMin} minute${diffMin !== 1 ? 's' : ''} ago`;
        } else if (diffHour < 24) {
          return `${diffHour} hour${diffHour !== 1 ? 's' : ''} ago`;
        } else if (diffDay < 30) {
          return `${diffDay} day${diffDay !== 1 ? 's' : ''} ago`;
        } else {
          // For older dates, just use the standard format
          return formatDate(dateString);
        }
      } catch (error) {
        return config.invalidDateText;
      }
    }
  ), [config, formatDate]);

  return {
    formatDate,
    formatTimestamp,
    formatRelativeTime,
  };
}

export default useDateFormatter; 