/**
 * Hooks index file
 * This file exports all reusable hooks
 */

// Export hooks
export { useNotification } from './useNotification';
export { useDataFetching } from './useDataFetching';
export { useDialogState } from './useDialogState';
export { useDateFormatter } from './useDateFormatter';

// Re-export existing hooks for convenience
export { default as useApiCache } from './useApiCache';
export { default as useDebounce } from './useDebounce';
export { default as useMetrics } from './useMetrics';
export { default as useProcesses } from './useProcesses';
export { default as useThrottledCallback } from './useThrottledCallback';
export { default as useResourceCRUD } from './useResourceCRUD'; 