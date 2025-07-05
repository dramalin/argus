import { useUiContext } from '../context/UiContext';
import type { NotificationSeverity, UseNotificationResult } from '../types/hooks';

/**
 * Hook for managing notifications
 * Provides a simplified interface for showing notifications using the UiContext
 * 
 * @returns Object with notification management functions
 */
export function useNotification(): UseNotificationResult {
  const { addNotification, removeNotification, clearNotifications } = useUiContext();
  
  /**
   * Show a notification with the specified message and severity
   * @param message The notification message
   * @param severity The severity level (success, error, info, warning)
   */
  const showNotification = (message: string, severity: NotificationSeverity) => {
    addNotification(message, severity);
  };
  
  return {
    showNotification,
    clearNotifications,
  };
} 