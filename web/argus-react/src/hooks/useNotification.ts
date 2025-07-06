import { useUiContext } from '../context/UiContext';
import type { NotificationSeverity, UseNotificationResult } from '../types/hooks';
import type { InAppNotification } from '../types/api';
import useWebSocket from './useWebSocket';

/**
 * Hook for managing notifications
 * Provides a simplified interface for showing notifications using the UiContext
 * 
 * @returns Object with notification management functions
 */
export function useNotification(): UseNotificationResult {
  const { addNotification } = useUiContext();

  const handleWebSocketMessage = (data: InAppNotification) => {
    showNotification(data.subject, data.severity);
  };
  
  const wsUrl = `ws://${window.location.host}/ws`;
  useWebSocket({
    url: wsUrl,
    onMessage: handleWebSocketMessage,
  });

  const showNotification = (message: string, severity: NotificationSeverity) => {
    addNotification(message, severity);
  };
  
  return {
    showNotification,
  };
} 