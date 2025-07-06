import { useEffect, useRef, useState, useCallback } from 'react';

export interface UseWebSocketProps {
  url: string;
  onMessage: (data: any) => void;
  onError?: (error: Event) => void;
  onOpen?: () => void;
  onClose?: () => void;
  reconnectLimit?: number;
  reconnectInterval?: number; // Initial interval in ms
  maxReconnectInterval?: number; // Max interval in ms
}

const useWebSocket = ({
  url,
  onMessage,
  onError,
  onOpen,
  onClose,
  reconnectLimit = 5,
  reconnectInterval = 1000, // Start with 1 second
  maxReconnectInterval = 30000, // Cap at 30 seconds
}: UseWebSocketProps) => {
  const ws = useRef<WebSocket | null>(null);
  const reconnectAttempts = useRef(0);
  const [isConnected, setIsConnected] = useState(false);
  const reconnectTimeoutId = useRef<NodeJS.Timeout | null>(null);
  const isCleaningUp = useRef(false);

  const clearReconnectTimeout = useCallback(() => {
    if (reconnectTimeoutId.current) {
      clearTimeout(reconnectTimeoutId.current);
      reconnectTimeoutId.current = null;
    }
  }, []);

  const connect = useCallback(() => {
    // Don't attempt to reconnect if we're cleaning up
    if (isCleaningUp.current) return;

    // Clear any existing reconnect timeout
    clearReconnectTimeout();

    // Check reconnect limit before attempting connection
    if (reconnectAttempts.current >= reconnectLimit) {
      console.error('WebSocket reconnect limit reached. Not attempting further reconnections.');
      if (onError) onError(new Event('reconnect_limit_reached'));
      return;
    }

    console.log(`Attempting WebSocket connection to ${url} (Attempt ${reconnectAttempts.current + 1}/${reconnectLimit})`);
    
    try {
      ws.current = new WebSocket(url);

      ws.current.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        reconnectAttempts.current = 0; // Reset on successful connection
        if (onOpen) onOpen();
      };

      ws.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          onMessage(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.current.onerror = (error) => {
        console.error('WebSocket error:', error);
        if (onError) onError(error);
      };

      ws.current.onclose = () => {
        console.log('WebSocket disconnected');
        setIsConnected(false);
        if (onClose) onClose();

        // Don't attempt reconnection if we're cleaning up
        if (isCleaningUp.current) return;

        // Increment attempt counter after a connection fails
        reconnectAttempts.current++;

        // Calculate exponential backoff with min and max limits
        const backoffMs = Math.min(
          reconnectInterval * Math.pow(2, reconnectAttempts.current - 1),
          maxReconnectInterval
        );
        
        // Add jitter (±10% of backoff)
        const jitter = backoffMs * 0.1 * (Math.random() * 2 - 1);
        const delay = Math.max(reconnectInterval, backoffMs + jitter);

        console.log(`Reconnecting in ${Math.round(delay / 1000)} seconds...`);
        reconnectTimeoutId.current = setTimeout(connect, delay);
      };
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
      if (onError) onError(new Event('connection_error'));
    }
  }, [url, onMessage, onError, onOpen, onClose, reconnectLimit, reconnectInterval, maxReconnectInterval, clearReconnectTimeout]);

  useEffect(() => {
    connect();

    return () => {
      isCleaningUp.current = true;
      clearReconnectTimeout();
      
      if (ws.current) {
        // Prevent the onclose handler from triggering a reconnect
        ws.current.onclose = null;
        ws.current.close();
        ws.current = null;
      }
      
      // Reset connection state and attempts
      setIsConnected(false);
      reconnectAttempts.current = 0;
    };
  }, [connect, clearReconnectTimeout]);

  const sendMessage = (data: any) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(data));
    } else {
      console.error('WebSocket is not connected. Message not sent:', data);
    }
  };

  return { isConnected, sendMessage };
};

export default useWebSocket; 