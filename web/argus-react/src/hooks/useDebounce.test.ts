import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import useDebounce from './useDebounce';

describe('useDebounce', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  it('should return the initial value immediately', () => {
    const { result } = renderHook(() => useDebounce('initial value', 500));
    expect(result.current).toBe('initial value');
  });

  it('should not update the debounced value before the delay has passed', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial value', delay: 500 } }
    );

    // Change the value
    rerender({ value: 'updated value', delay: 500 });

    // Value shouldn't have changed yet
    expect(result.current).toBe('initial value');

    // Fast-forward time by 300ms (less than the 500ms delay)
    vi.advanceTimersByTime(300);
    expect(result.current).toBe('initial value');
  });

  it('should update the debounced value after the delay has passed', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial value', delay: 500 } }
    );

    // Change the value
    act(() => {
      rerender({ value: 'updated value', delay: 500 });
    });

    // Fast-forward time by 500ms (equal to the delay)
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(result.current).toBe('updated value');
  });

  it('should handle multiple updates correctly', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial value', delay: 500 } }
    );

    // First update
    act(() => {
      rerender({ value: 'update 1', delay: 500 });
      vi.advanceTimersByTime(200);
    });

    // Second update before the first one completes
    act(() => {
      rerender({ value: 'update 2', delay: 500 });
      vi.advanceTimersByTime(200);
    });

    // Third update before the second one completes
    act(() => {
      rerender({ value: 'update 3', delay: 500 });
    });
    
    // Value should still be the initial one before final advance
    expect(result.current).toBe('initial value');

    // Fast-forward time to complete the last update
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(result.current).toBe('update 3');
  });

  it('should handle delay changes correctly', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial value', delay: 500 } }
    );

    // Change the value and the delay
    act(() => {
      rerender({ value: 'updated value', delay: 1000 });
    });

    // Fast-forward time by 500ms (equal to the original delay)
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(result.current).toBe('initial value');

    // Fast-forward time by another 500ms (to reach the new delay)
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(result.current).toBe('updated value');
  });

  it('should clean up timeout on unmount', () => {
    const clearTimeoutSpy = vi.spyOn(global, 'clearTimeout');
    
    const { unmount } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial value', delay: 500 } }
    );

    unmount();
    expect(clearTimeoutSpy).toHaveBeenCalled();
  });
}); 