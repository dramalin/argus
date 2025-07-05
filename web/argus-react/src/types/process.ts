export interface ProcessInfo {
  pid: number;
  name: string;
  cpu_percent: number;
  mem_percent: number;
}

export interface ProcessQueryParams {
  limit?: number;
  offset?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
  name_contains?: string;
  min_cpu?: number;
  min_memory?: number;
}

export interface ProcessPagination {
  total_count: number;
  total_pages: number;
  current_page: number;
  limit: number;
  offset: number;
  has_next: boolean;
  has_previous: boolean;
}

export interface ProcessFilters {
  sort_by: string;
  sort_order: string;
  min_cpu: number | null;
  min_memory: number | null;
  name_contains: string | null;
  top_n: number | null;
}

export interface ProcessResponse {
  processes: ProcessInfo[];
  total_count: number;
  pagination: ProcessPagination;
  filters: ProcessFilters;
  updated_at: string;
} 