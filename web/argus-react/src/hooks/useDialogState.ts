import { useState, useCallback } from 'react';
import type { UseDialogStateResult } from '../types/hooks';

/**
 * A custom hook for managing multiple dialog states
 * 
 * @template T Optional type for dialog names to ensure type safety
 * @returns Object with methods to manage dialog states
 */
export function useDialogState<T extends string = string>(): UseDialogStateResult<T> {
  // Track dialog states in a single object for efficiency
  const [dialogStates, setDialogStates] = useState<Record<string, boolean>>({});
  
  /**
   * Open a specific dialog
   * @param name The name/identifier of the dialog to open
   */
  const openDialog = useCallback((name: T) => {
    setDialogStates(prev => ({
      ...prev,
      [name]: true
    }));
  }, []);
  
  /**
   * Close a specific dialog
   * @param name The name/identifier of the dialog to close
   */
  const closeDialog = useCallback((name: T) => {
    setDialogStates(prev => ({
      ...prev,
      [name]: false
    }));
  }, []);
  
  /**
   * Close all dialogs at once
   */
  const closeAllDialogs = useCallback(() => {
    // Create a new object with all values set to false
    const allClosed = Object.keys(dialogStates).reduce((acc, key) => {
      acc[key] = false;
      return acc;
    }, {} as Record<string, boolean>);
    
    setDialogStates(allClosed);
  }, [dialogStates]);
  
  /**
   * Check if a specific dialog is open
   * @param name The name/identifier of the dialog to check
   * @returns Boolean indicating whether the dialog is open
   */
  const isDialogOpen = useCallback((name: T): boolean => {
    return !!dialogStates[name];
  }, [dialogStates]);
  
  return {
    dialogStates,
    openDialog,
    closeDialog,
    closeAllDialogs,
    isDialogOpen
  };
}

export default useDialogState; 