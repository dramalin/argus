import { useState, useCallback } from 'react';
import { useDataFetching, useNotification, useDialogState } from '.'; // Assuming index.ts exports these

interface UseResourceCRUDProps<T, CreateT = T, UpdateT = T> {
  resourceName: string;
  fetchFn: () => Promise<{ success: boolean; data?: T[]; error?: string }>;
  createFn?: (item: CreateT) => Promise<{ success: boolean; data?: T; error?: string }>;
  updateFn?: (id: string, item: UpdateT) => Promise<{ success: boolean; data?: T; error?: string }>;
  deleteFn?: (id: string) => Promise<{ success: boolean; error?: string }>;
  cacheTTL?: number; // Optional cache TTL for useDataFetching
}

interface UseResourceCRUDResult<T> {
  items: T[];
  loading: boolean;
  error: string | null;
  lastUpdated: Date | null;
  refetch: () => Promise<void>;
  actionLoading: boolean;
  selectedItem: T | null;
  setSelectedItem: (item: T | null) => void;
  isDialogOpen: (dialogType: string) => boolean;
  openDialog: (dialogType: string) => void;
  closeDialog: (dialogType: string) => void;
  handleCreate: (item: any) => Promise<void>;
  handleUpdate: (id: string, item: any) => Promise<void>;
  handleDelete: (id: string) => Promise<void>;
}

function useResourceCRUD<T, CreateT = T, UpdateT = T>(
  {
    resourceName,
    fetchFn,
    createFn,
    updateFn,
    deleteFn,
    cacheTTL,
  }: UseResourceCRUDProps<T, CreateT, UpdateT>
): UseResourceCRUDResult<T> {
  const { data, loading, error, lastUpdated, refetch } = useDataFetching<T[]>(resourceName, fetchFn, { cacheTTL });
  const { showNotification } = useNotification();
  const { openDialog, closeDialog, isDialogOpen } = useDialogState();
  const [actionLoading, setActionLoading] = useState<boolean>(false);
  const [selectedItem, setSelectedItem] = useState<T | null>(null);

  const items = data || [];

  const handleCreate = useCallback(async (item: CreateT) => {
    if (!createFn) {
      showNotification(`Create operation not supported for ${resourceName}`, 'warning');
      return;
    }
    setActionLoading(true);
    try {
      const response = await createFn(item);
      if (response.success) {
        showNotification(`${resourceName} created successfully`, 'success');
        await refetch();
        closeDialog('create');
      } else {
        throw new Error(response.error || `Failed to create ${resourceName}`);
      }
    } catch (err) {
      showNotification(err instanceof Error ? err.message : `Failed to create ${resourceName}`, 'error');
    } finally {
      setActionLoading(false);
    }
  }, [createFn, refetch, showNotification, closeDialog, resourceName]);

  const handleUpdate = useCallback(async (id: string, item: UpdateT) => {
    if (!updateFn) {
      showNotification(`Update operation not supported for ${resourceName}`, 'warning');
      return;
    }
    setActionLoading(true);
    try {
      const response = await updateFn(id, item);
      if (response.success) {
        showNotification(`${resourceName} updated successfully`, 'success');
        await refetch();
        closeDialog('edit');
      } else {
        throw new Error(response.error || `Failed to update ${resourceName}`);
      }
    } catch (err) {
      showNotification(err instanceof Error ? err.message : `Failed to update ${resourceName}`, 'error');
    } finally {
      setActionLoading(false);
    }
  }, [updateFn, refetch, showNotification, closeDialog, resourceName]);

  const handleDelete = useCallback(async (id: string) => {
    if (!deleteFn) {
      showNotification(`Delete operation not supported for ${resourceName}`, 'warning');
      return;
    }
    setActionLoading(true);
    try {
      const response = await deleteFn(id);
      if (response.success) {
        showNotification(`${resourceName} deleted successfully`, 'success');
        await refetch();
        closeDialog('delete');
      } else {
        throw new Error(response.error || `Failed to delete ${resourceName}`);
      }
    } catch (err) {
      showNotification(err instanceof Error ? err.message : `Failed to delete ${resourceName}`, 'error');
    } finally {
      setActionLoading(false);
    }
  }, [deleteFn, refetch, showNotification, closeDialog, resourceName]);

  return {
    items,
    loading,
    error,
    lastUpdated,
    refetch,
    actionLoading,
    selectedItem,
    setSelectedItem,
    isDialogOpen,
    openDialog,
    closeDialog,
    handleCreate,
    handleUpdate,
    handleDelete,
  };
}

export default useResourceCRUD; 