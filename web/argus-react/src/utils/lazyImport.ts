/**
 * Utility for creating lazy-loaded components with TypeScript support
 */
import { ComponentType, lazy } from 'react';

/**
 * Creates a lazy-loaded component with proper TypeScript typing
 * @param factory - Import function that returns a promise resolving to a module
 * @param exportName - Name of the exported component (default if not specified)
 * @returns Lazy-loaded component with proper TypeScript typing
 */
export function lazyImport<
  T extends ComponentType<any>,
  I extends { [K2 in K]: T },
  K extends keyof I
>(factory: () => Promise<I>, name: K): T {
  return lazy(() => factory().then(module => ({ default: module[name] }))) as T;
}

/**
 * Creates a lazy-loaded component with proper TypeScript typing for default exports
 * @param factory - Import function that returns a promise resolving to a module with default export
 * @returns Lazy-loaded component with proper TypeScript typing
 */
export function lazyLoad<T extends ComponentType<any>>(
  factory: () => Promise<{ default: T }>
): T {
  return lazy(factory) as T;
}

/**
 * Prefetches a component to improve perceived performance
 * @param factory - Import function that returns a promise resolving to a module
 */
export function prefetchComponent(factory: () => Promise<any>): void {
  // Start loading the component in the background
  factory();
}

/**
 * Prefetches multiple components to improve perceived performance
 * @param factories - Array of import functions that return promises resolving to modules
 */
export function prefetchComponents(factories: Array<() => Promise<any>>): void {
  factories.forEach(factory => prefetchComponent(factory));
}

/**
 * Dynamically imports a module with a timeout
 * @param factory - Import function that returns a promise resolving to a module
 * @param timeoutMs - Timeout in milliseconds
 * @returns Promise resolving to the module or rejecting with an error
 */
export function importWithTimeout<T>(
  factory: () => Promise<T>,
  timeoutMs: number = 10000
): Promise<T> {
  return Promise.race([
    factory(),
    new Promise<never>((_, reject) => 
      setTimeout(() => reject(new Error(`Import timed out after ${timeoutMs}ms`)), timeoutMs)
    )
  ]);
} 