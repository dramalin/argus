/**
 * Common type definitions shared across the application
 */

/**
 * Status types used throughout the application
 * @description Represents the different states of an async operation
 */
export type Status = 'idle' | 'loading' | 'success' | 'error';

/**
 * Sort direction type
 * @description Represents the direction of sorting (ascending or descending)
 */
export type SortDirection = 'asc' | 'desc';

/**
 * Time period for data filtering
 * @description Represents different time periods for filtering data
 */
export type TimePeriod = '1h' | '6h' | '12h' | '24h' | '7d' | '30d' | 'custom';

/**
 * Theme mode type
 * @description Represents the different theme modes available
 */
export type ThemeMode = 'light' | 'dark' | 'system';

/**
 * Notification type
 * @description Represents the different types of notifications
 */
export type NotificationType = 'info' | 'success' | 'warning' | 'error';

/**
 * Notification interface
 * @description Represents a notification message
 */
export interface Notification {
  /** Unique identifier for the notification */
  id: string;
  /** Type of notification */
  type: NotificationType;
  /** Notification message content */
  message: string;
  /** Timestamp when the notification was created */
  timestamp: number;
  /** Optional duration after which the notification should auto-hide (in milliseconds) */
  autoHideDuration?: number;
}

/**
 * Pagination parameters
 * @description Represents pagination state and metadata
 */
export interface PaginationParams {
  /** Current page number (1-based) */
  page: number;
  /** Number of items per page */
  pageSize: number;
  /** Total number of items across all pages */
  totalItems: number;
  /** Total number of pages */
  totalPages: number;
}

/**
 * Generic async data state
 * @description Represents the state of asynchronously loaded data
 * @template T The type of data being loaded
 */
export interface AsyncData<T> {
  /** The data, if loaded successfully */
  data: T | null;
  /** Whether the data is currently loading */
  loading: boolean;
  /** Error message if loading failed, null otherwise */
  error: string | null;
  /** ISO timestamp when the data was last updated, null if never updated */
  lastUpdated: string | null;
}

/**
 * Generic key-value record
 * @description A record with string keys and values of type T
 * @template T The type of values in the record (defaults to string)
 */
export type KeyValueRecord<T = string> = Record<string, T>;

/**
 * Function type with no parameters and no return value
 * @description A function that takes no arguments and returns nothing
 */
export type VoidFunction = () => void;

/**
 * Function type with generic parameter and no return value
 * @description A function that takes a parameter and returns nothing
 * @template T The type of the parameter
 */
export type ParameterizedVoidFunction<T> = (param: T) => void;

/**
 * Nullable type
 * @description Makes a type nullable (can be null)
 * @template T The type to make nullable
 */
export type Nullable<T> = T | null;

/**
 * Optional type
 * @description Makes a type optional (can be undefined)
 * @template T The type to make optional
 */
export type Optional<T> = T | undefined;

/**
 * Maybe type
 * @description Makes a type nullable or optional (can be null or undefined)
 * @template T The type to make nullable or optional
 */
export type Maybe<T> = T | null | undefined;

/**
 * DeepPartial type
 * @description Makes all properties in an object optional recursively
 * @template T The type to make deeply partial
 */
export type DeepPartial<T> = T extends object ? {
  [P in keyof T]?: DeepPartial<T[P]>;
} : T;

/**
 * NonNullable type
 * @description Removes null and undefined from a type
 * @template T The type to remove null and undefined from
 */
export type NonNullableFields<T> = {
  [P in keyof T]: NonNullable<T[P]>;
};

/**
 * Awaited type
 * @description Unwraps the type inside a Promise
 * @template T The Promise type to unwrap
 */
export type Awaited<T> = T extends Promise<infer U> ? U : T;

/**
 * ActionHandler type
 * @description A function that handles an action with a payload
 * @template T The type of the payload
 */
export type ActionHandler<T> = (payload: T) => void;

/**
 * ErrorHandler type
 * @description A function that handles an error
 */
export type ErrorHandler = (error: Error | string) => void;

/**
 * LoadingState type
 * @description Represents the loading state of a component or operation
 */
export interface LoadingState {
  /** Whether the component or operation is loading */
  loading: boolean;
  /** Error message if loading failed, null otherwise */
  error: string | null;
}

/**
 * Size type
 * @description Represents different size options
 */
export type Size = 'small' | 'medium' | 'large';

/**
 * Position type
 * @description Represents different position options
 */
export type Position = 'top' | 'right' | 'bottom' | 'left';

/**
 * Alignment type
 * @description Represents different alignment options
 */
export type Alignment = 'start' | 'center' | 'end';

/**
 * ValidationResult type
 * @description Represents the result of a validation operation
 */
export interface ValidationResult {
  /** Whether the validation passed */
  valid: boolean;
  /** Error message if validation failed, null otherwise */
  error: string | null;
} 