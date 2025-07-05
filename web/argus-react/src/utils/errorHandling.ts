/**
 * Error handling utilities
 * @module ErrorHandling
 */
import { isObject, isString } from './typeGuards';

/**
 * Custom error class for API errors
 * @class ApiError
 * @extends Error
 */
export class ApiError extends Error {
  /** HTTP status code */
  status: number;
  /** Response data */
  data?: unknown;
  /** Request URL that caused the error */
  url: string | null;
  /** Request method that caused the error */
  method: string | null;

  /**
   * Create a new ApiError
   * @param message - Error message
   * @param status - HTTP status code
   * @param data - Additional error data
   * @param url - Request URL that caused the error
   * @param method - Request method that caused the error
   */
  constructor(
    message: string, 
    status: number, 
    data?: unknown, 
    url: string | null = null, 
    method: string | null = null
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.data = data;
    this.url = url;
    this.method = method;
    
    // Ensure proper prototype chain for instanceof checks
    Object.setPrototypeOf(this, ApiError.prototype);
  }

  /**
   * Get a formatted string representation of the error
   * @returns A formatted string with error details
   */
  toFormattedString(): string {
    return `API Error (${this.status}): ${this.message}${this.url ? ` [${this.method || 'GET'} ${this.url}]` : ''}`;
  }
}

/**
 * Custom error class for timeout errors
 * @class TimeoutError
 * @extends Error
 */
export class TimeoutError extends Error {
  /** Request URL that timed out */
  url: string | null;
  /** Timeout duration in milliseconds */
  timeoutMs: number | null;

  /**
   * Create a new TimeoutError
   * @param message - Error message
   * @param url - Request URL that timed out
   * @param timeoutMs - Timeout duration in milliseconds
   */
  constructor(message = 'Request timed out', url: string | null = null, timeoutMs: number | null = null) {
    super(message);
    this.name = 'TimeoutError';
    this.url = url;
    this.timeoutMs = timeoutMs;
    
    // Ensure proper prototype chain for instanceof checks
    Object.setPrototypeOf(this, TimeoutError.prototype);
  }

  /**
   * Get a formatted string representation of the error
   * @returns A formatted string with error details
   */
  toFormattedString(): string {
    return `Timeout Error: ${this.message}${this.url ? ` [${this.url}]` : ''}${this.timeoutMs ? ` (${this.timeoutMs}ms)` : ''}`;
  }
}

/**
 * Custom error class for validation errors
 * @class ValidationError
 * @extends Error
 */
export class ValidationError extends Error {
  /** Validation errors by field */
  errors: Record<string, string[]>;

  /**
   * Create a new ValidationError
   * @param message - Error message
   * @param errors - Validation errors by field
   */
  constructor(message = 'Validation failed', errors: Record<string, string[]> = {}) {
    super(message);
    this.name = 'ValidationError';
    this.errors = errors;
    
    // Ensure proper prototype chain for instanceof checks
    Object.setPrototypeOf(this, ValidationError.prototype);
  }

  /**
   * Get a formatted string representation of the error
   * @returns A formatted string with error details
   */
  toFormattedString(): string {
    const errorDetails = Object.entries(this.errors)
      .map(([field, messages]) => `${field}: ${messages.join(', ')}`)
      .join('; ');
    
    return `Validation Error: ${this.message}${errorDetails ? ` (${errorDetails})` : ''}`;
  }
}

/**
 * Extract an error message from various error types
 * @param error - The error to extract a message from
 * @param fallbackMessage - Fallback message if no error message can be extracted
 * @returns A user-friendly error message
 */
export function getErrorMessage(error: unknown, fallbackMessage = 'An unknown error occurred'): string {
  // Handle string errors
  if (isString(error)) {
    return error;
  }
  
  // Handle Error objects
  if (error instanceof Error) {
    // Handle custom error classes
    if (error instanceof ApiError) {
      return error.toFormattedString();
    }
    
    if (error instanceof TimeoutError) {
      return error.toFormattedString();
    }
    
    if (error instanceof ValidationError) {
      return error.toFormattedString();
    }
    
    return error.message;
  }
  
  // Handle objects with error or message properties
  if (isObject(error)) {
    if (isString(error.message)) {
      return error.message;
    }
    if (isString(error.error)) {
      return error.error;
    }
    if (isString(error.errorMessage)) {
      return error.errorMessage;
    }
    
    // Try to convert to string
    try {
      return JSON.stringify(error);
    } catch {
      // Ignore JSON stringify errors
    }
  }
  
  // Default error message
  return fallbackMessage;
}

/**
 * Create a function that will timeout after a specified time
 * @param ms - Timeout in milliseconds
 * @param url - Optional URL for context
 * @returns A promise that rejects with a TimeoutError after the specified time
 */
export function createTimeout(ms: number, url: string | null = null): Promise<never> {
  return new Promise((_, reject) => {
    setTimeout(() => {
      reject(new TimeoutError(`Request timed out after ${ms}ms`, url, ms));
    }, ms);
  });
}

/**
 * Execute a promise with a timeout
 * @template T - The promise result type
 * @param promise - The promise to execute
 * @param timeoutMs - Timeout in milliseconds
 * @param url - Optional URL for context
 * @returns The result of the promise, or throws a TimeoutError if the timeout is reached
 */
export async function withTimeout<T>(promise: Promise<T>, timeoutMs: number, url: string | null = null): Promise<T> {
  return Promise.race([
    promise,
    createTimeout(timeoutMs, url)
  ]);
}

/**
 * Options for retry with backoff
 */
export interface RetryOptions {
  /** Maximum number of retries */
  retries?: number;
  /** Initial delay in milliseconds */
  initialDelay?: number;
  /** Maximum delay in milliseconds */
  maxDelay?: number;
  /** Function to determine if a retry should be attempted */
  shouldRetry?: (error: unknown, attempt: number) => boolean;
  /** Function to execute before each retry */
  onRetry?: (error: unknown, attempt: number, delay: number) => void;
}

/**
 * Retry a function with exponential backoff
 * @template T - The function result type
 * @param fn - The function to retry
 * @param options - Retry options
 * @returns The result of the function, or throws the last error
 */
export async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  options: RetryOptions = {}
): Promise<T> {
  const {
    retries = 3,
    initialDelay = 300,
    maxDelay = 30000,
    shouldRetry = () => true,
    onRetry = () => {}
  } = options;
  
  let attempt = 0;
  let lastError: unknown;
  
  while (attempt <= retries) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;
      attempt += 1;
      
      if (attempt > retries || !shouldRetry(error, attempt)) {
        throw error;
      }
      
      // Calculate delay with exponential backoff and jitter
      const exponentialDelay = Math.min(
        initialDelay * Math.pow(2, attempt - 1),
        maxDelay
      );
      const jitter = Math.random() * 0.3 * exponentialDelay;
      const delay = exponentialDelay + jitter;
      
      // Execute onRetry callback
      onRetry(error, attempt, delay);
      
      // Wait with exponential backoff
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
  
  // This should never be reached, but TypeScript needs it
  throw lastError;
}

/**
 * Error severity levels
 */
export const ErrorSeverity = {
  /** Informational errors that don't affect functionality */
  INFO: 'info',
  /** Warning errors that might affect functionality but aren't critical */
  WARNING: 'warning',
  /** Critical errors that affect functionality */
  ERROR: 'error',
  /** Fatal errors that crash the application */
  FATAL: 'fatal'
} as const;

export type ErrorSeverityType = typeof ErrorSeverity[keyof typeof ErrorSeverity];

/**
 * Log context information
 */
export interface LogContext {
  /** Component or module where the error occurred */
  component?: string;
  /** Function where the error occurred */
  function?: string;
  /** Additional data related to the error */
  data?: Record<string, unknown>;
  /** Error severity */
  severity?: ErrorSeverityType;
  /** User action that triggered the error */
  action?: string;
  /** User ID, if available */
  userId?: string;
  /** Session ID, if available */
  sessionId?: string;
}

/**
 * Log an error to the console with additional context
 * @param error - The error to log
 * @param context - Additional context information
 */
export function logError(error: unknown, context: LogContext = {}): void {
  const {
    component,
    function: functionName,
    data,
    severity = ErrorSeverity.ERROR,
    action,
    userId,
    sessionId
  } = context;
  
  const errorInfo = {
    error,
    message: getErrorMessage(error),
    timestamp: new Date().toISOString(),
    severity,
    component,
    function: functionName,
    action,
    userId,
    sessionId,
    ...data
  };
  
  // Log to console with appropriate method based on severity
  switch (severity) {
    case ErrorSeverity.INFO:
      console.info('[Argus Info]', errorInfo);
      break;
    case ErrorSeverity.WARNING:
      console.warn('[Argus Warning]', errorInfo);
      break;
    case ErrorSeverity.FATAL:
      console.error('[Argus Fatal]', errorInfo);
      break;
    case ErrorSeverity.ERROR:
    default:
      console.error('[Argus Error]', errorInfo);
      break;
  }
  
  // TODO: In a production environment, you might want to send errors to a monitoring service
  // sendErrorToMonitoringService(errorInfo);
}

/**
 * Create a safe version of a function that catches errors
 * @template T - The function parameters type
 * @template R - The function return type
 * @param fn - The function to make safe
 * @param errorHandler - Function to handle errors
 * @returns A safe version of the function that never throws
 */
export function makeSafe<T extends unknown[], R>(
  fn: (...args: T) => R,
  errorHandler: (error: unknown, ...args: T) => R
): (...args: T) => R {
  return (...args: T): R => {
    try {
      return fn(...args);
    } catch (error) {
      return errorHandler(error, ...args);
    }
  };
}

/**
 * Create a safe version of an async function that catches errors
 * @template T - The function parameters type
 * @template R - The function return type
 * @param fn - The async function to make safe
 * @param errorHandler - Function to handle errors
 * @returns A safe version of the async function that never rejects
 */
export function makeSafeAsync<T extends unknown[], R>(
  fn: (...args: T) => Promise<R>,
  errorHandler: (error: unknown, ...args: T) => Promise<R> | R
): (...args: T) => Promise<R> {
  return async (...args: T): Promise<R> => {
    try {
      return await fn(...args);
    } catch (error) {
      return errorHandler(error, ...args);
    }
  };
} 