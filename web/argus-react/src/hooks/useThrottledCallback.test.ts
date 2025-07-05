import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import useThrottledCallback from './useThrottledCallback';

describe('useThrottledCallback', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  it('should call the callback immediately on first call', () => {
    const callback = vi.fn();
    const { result } = renderHook(() => useThrottledCallback(callback, 200));
    
    result.current('test');
    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith('test');
  });

  it('should not call the callback again before the delay has passed', () => {
    const callback = vi.fn();
    const { result } = renderHook(() => useThrottledCallback(callback, 200));
    
    // First call - immediate
    result.current('test1');
    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith('test1');
    
    // Second call before delay - should not trigger callback
    result.current('test2');
    expect(callback).toHaveBeenCalledTimes(1);
    
    // Third call before delay - should not trigger callback
    result.current('test3');
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('should call the callback with the latest arguments after the delay', () => {
    const callback = vi.fn();
    const { result } = renderHook(() => useThrottledCallback(callback, 200));
    
    // First call - immediate
    result.current('test1');
    expect(callback).toHaveBeenCalledTimes(1);
    
    // Second call before delay
    result.current('test2');
    
    // Third call before delay
    result.current('test3');
    
    // Advance timers
    vi.advanceTimersByTime(200);
    
    // Should have called with the latest arguments
    expect(callback).toHaveBeenCalledTimes(2);
    expect(callback).toHaveBeenCalledWith('test3');
  });

  it('should work correctly with multiple parameters', () => {
    const callback = vi.fn();
    const { result } = renderHook(() => useThrottledCallback(callback, 200));
    
    // First call - immediate
    result.current('test', 123, { key: 'value' });
    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback).toHaveBeenCalledWith('test', 123, { key: 'value' });
    
    // Second call before delay
    result.current('updated', 456, { key: 'new value' });
    
    // Advance timers
    vi.advanceTimersByTime(200);
    
    // Should have called with the latest arguments
    expect(callback).toHaveBeenCalledTimes(2);
    expect(callback).toHaveBeenCalledWith('updated', 456, { key: 'new value' });
  });

  it('should handle multiple throttled calls correctly', () => {
    const callback = vi.fn();
    const { result } = renderHook(() => useThrottledCallback(callback, 200));
    
    // First call - immediate
    result.current('test1');
    expect(callback).toHaveBeenCalledTimes(1);
    
    // Second call before delay
    result.current('test2');
    
    // Advance timers partially
    vi.advanceTimersByTime(100);
    expect(callback).toHaveBeenCalledTimes(1);
    
    // Third call
    result.current('test3');
    
    // Advance timers to complete first throttle
    vi.advanceTimersByTime(100);
    expect(callback).toHaveBeenCalledTimes(2);
    expect(callback).toHaveBeenCalledWith('test3');
    
    // Fourth call
    result.current('test4');
    
    // Advance timers to complete second throttle
    vi.advanceTimersByTime(200);
    expect(callback).toHaveBeenCalledTimes(3);
    expect(callback).toHaveBeenCalledWith('test4');
  });

  it('should update when the callback changes', () => {
    const callback1 = vi.fn();
    const callback2 = vi.fn();
    
    const { result, rerender } = renderHook(
      ({ callback }) => useThrottledCallback(callback, 200),
      { initialProps: { callback: callback1 } }
    );
    
    // Call with first callback
    result.current('test1');
    expect(callback1).toHaveBeenCalledTimes(1);
    expect(callback2).toHaveBeenCalledTimes(0);
    
    // Update the callback
    rerender({ callback: callback2 });
    
    // Call with second callback
    result.current('test2');
    expect(callback1).toHaveBeenCalledTimes(1);
    expect(callback2).toHaveBeenCalledTimes(1);
  });

  it('should update when the delay changes', () => {
    const callback = vi.fn();
    
    const { result, rerender } = renderHook(
      ({ delay }) => useThrottledCallback(callback, delay),
      { initialProps: { delay: 200 } }
    );
    
    // First call - immediate
    result.current('test1');
    expect(callback).toHaveBeenCalledTimes(1);
    
    // Second call before delay
    result.current('test2');
    
    // Update the delay
    rerender({ delay: 500 });
    
    // Third call
    result.current('test3');
    
    // Advance timers to what would have been the first delay
    vi.advanceTimersByTime(200);
    
    // Should not have called yet due to new longer delay
    expect(callback).toHaveBeenCalledTimes(1);
    
    // Advance timers to complete the new delay
    vi.advanceTimersByTime(300);
    expect(callback).toHaveBeenCalledTimes(2);
    expect(callback).toHaveBeenCalledWith('test3');
  });
}); 