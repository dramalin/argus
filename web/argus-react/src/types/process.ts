/**
 * Process-related type definitions for Argus System Monitor
 */
import type { SortDirection } from './common';

/**
 * Process information returned by the API
 */
export interface ProcessInfo {
  /** Process ID */
  pid: number;
  /** Process name */
  name: string;
  /** CPU usage percentage (0-100) */
  cpu_percent: number;
  /** Memory usage percentage (0-100) */
  mem_percent: number;
}

/**
 * Valid sort fields for process queries
 */
export type ProcessSortField = 'pid' | 'name' | 'cpu' | 'memory' | 'created_at';

/**
 * Query parameters for filtering and sorting processes
 */
export interface ProcessQueryParams {
  /** Maximum number of processes to return */
  limit?: number;
  /** Number of processes to skip (for pagination) */
  offset?: number;
  /** Field to sort by */
  sort_by?: ProcessSortField | string;
  /** Sort direction */
  sort_order?: SortDirection;
  /** Filter processes by name (case-insensitive substring match) */
  name_contains?: string;
  /** Filter processes by minimum CPU usage percentage */
  min_cpu?: number;
  /** Filter processes by minimum memory usage percentage */
  min_memory?: number;
}

/**
 * Pagination information returned by the API
 */
export interface ProcessPagination {
  /** Total number of processes matching the query */
  total_count: number;
  /** Total number of pages */
  total_pages: number;
  /** Current page number (1-based) */
  current_page: number;
  /** Maximum number of processes per page */
  limit: number;
  /** Number of processes skipped */
  offset: number;
  /** Whether there is a next page */
  has_next: boolean;
  /** Whether there is a previous page */
  has_previous: boolean;
}

/**
 * Filter information returned by the API
 */
export interface ProcessFilters {
  /** Field used for sorting */
  sort_by: string;
  /** Sort direction */
  sort_order: string;
  /** Minimum CPU usage filter, if applied */
  min_cpu: number | null;
  /** Minimum memory usage filter, if applied */
  min_memory: number | null;
  /** Name filter, if applied */
  name_contains: string | null;
  /** Top N filter, if applied */
  top_n: number | null;
}

/**
 * Process response returned by the API
 */
export interface ProcessResponse {
  /** List of processes matching the query */
  processes: ProcessInfo[];
  /** Total number of processes matching the query */
  total_count: number;
  /** Pagination information */
  pagination: ProcessPagination;
  /** Filter information */
  filters: ProcessFilters;
  /** Timestamp when the data was updated (ISO format) */
  updated_at: string;
} 