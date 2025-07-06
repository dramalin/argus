import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import useWebSocket from './useWebSocket';

// Mock WebSocket
const mockWebSocket = {
  send: vi.fn(),
  close: vi.fn(),
  onopen: () => {},
  onmessage: () => {},
  onerror: () => {},
  onclose: () => {},
  readyState: WebSocket.OPEN,
};

vi.stubGlobal('WebSocket', vi.fn(() => mockWebSocket));

describe('useWebSocket', () => {
  const onMessage = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should call onMessage when a message is received', () => {
    renderHook(() => useWebSocket({ url: 'ws://test.com', onMessage }));

    act(() => {
      mockWebSocket.onmessage({ data: JSON.stringify({ message: 'test' }) });
    });

    expect(onMessage).toHaveBeenCalledWith({ message: 'test' });
  });
}); 