/**
 * Type guard utilities for runtime type checking
 * @module TypeGuards
 */
import type { ProcessInfo, SystemMetrics } from '../types';

/**
 * Type guard to check if a value is defined (not null or undefined)
 * @param value - The value to check
 * @returns True if the value is defined
 * @example
 * ```ts
 * const value: string | null = getData();
 * if (isDefined(value)) {
 *   // value is string here
 *   console.log(value.toUpperCase());
 * }
 * ```
 */
export function isDefined<T>(value: T | null | undefined): value is T {
  return value !== null && value !== undefined;
}

/**
 * Type guard to check if a value is a string
 * @param value - The value to check
 * @returns True if the value is a string
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isString(value)) {
 *   // value is string here
 *   console.log(value.toUpperCase());
 * }
 * ```
 */
export function isString(value: unknown): value is string {
  return typeof value === 'string';
}

/**
 * Type guard to check if a value is a non-empty string
 * @param value - The value to check
 * @returns True if the value is a non-empty string
 * @example
 * ```ts
 * const value: string | null = getData();
 * if (isNonEmptyString(value)) {
 *   // value is a non-empty string here
 *   console.log(value.toUpperCase());
 * }
 * ```
 */
export function isNonEmptyString(value: unknown): value is string {
  return isString(value) && value.trim().length > 0;
}

/**
 * Type guard to check if a value is a number
 * @param value - The value to check
 * @returns True if the value is a number
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isNumber(value)) {
 *   // value is number here
 *   console.log(value.toFixed(2));
 * }
 * ```
 */
export function isNumber(value: unknown): value is number {
  return typeof value === 'number' && !isNaN(value);
}

/**
 * Type guard to check if a value is a finite number
 * @param value - The value to check
 * @returns True if the value is a finite number
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isFiniteNumber(value)) {
 *   // value is a finite number here
 *   console.log(value.toFixed(2));
 * }
 * ```
 */
export function isFiniteNumber(value: unknown): value is number {
  return isNumber(value) && isFinite(value);
}

/**
 * Type guard to check if a value is a positive number
 * @param value - The value to check
 * @returns True if the value is a positive number
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isPositiveNumber(value)) {
 *   // value is a positive number here
 *   console.log(value.toFixed(2));
 * }
 * ```
 */
export function isPositiveNumber(value: unknown): value is number {
  return isNumber(value) && value > 0;
}

/**
 * Type guard to check if a value is a boolean
 * @param value - The value to check
 * @returns True if the value is a boolean
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isBoolean(value)) {
 *   // value is boolean here
 *   console.log(value ? 'Yes' : 'No');
 * }
 * ```
 */
export function isBoolean(value: unknown): value is boolean {
  return typeof value === 'boolean';
}

/**
 * Type guard to check if a value is an array
 * @template T - The expected type of array elements
 * @param value - The value to check
 * @returns True if the value is an array
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isArray<string>(value)) {
 *   // value is string[] here
 *   value.forEach(item => console.log(item.toUpperCase()));
 * }
 * ```
 */
export function isArray<T>(value: unknown): value is Array<T> {
  return Array.isArray(value);
}

/**
 * Type guard to check if a value is a non-empty array
 * @template T - The expected type of array elements
 * @param value - The value to check
 * @returns True if the value is a non-empty array
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isNonEmptyArray<string>(value)) {
 *   // value is a non-empty string[] here
 *   console.log(`First item: ${value[0]}`);
 * }
 * ```
 */
export function isNonEmptyArray<T>(value: unknown): value is Array<T> {
  return isArray<T>(value) && value.length > 0;
}

/**
 * Type guard to check if a value is an object (not null, not an array)
 * @param value - The value to check
 * @returns True if the value is an object
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isObject(value)) {
 *   // value is Record<string, unknown> here
 *   console.log(Object.keys(value));
 * }
 * ```
 */
export function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/**
 * Type guard to check if a value is a non-empty object
 * @param value - The value to check
 * @returns True if the value is a non-empty object
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isNonEmptyObject(value)) {
 *   // value is a non-empty Record<string, unknown> here
 *   console.log(`First key: ${Object.keys(value)[0]}`);
 * }
 * ```
 */
export function isNonEmptyObject(value: unknown): value is Record<string, unknown> {
  return isObject(value) && Object.keys(value).length > 0;
}

/**
 * Type guard to check if a value is a function
 * @param value - The value to check
 * @returns True if the value is a function
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isFunction(value)) {
 *   // value is Function here
 *   value();
 * }
 * ```
 */
export function isFunction(value: unknown): value is Function {
  return typeof value === 'function';
}

/**
 * Type guard to check if a value is a date
 * @param value - The value to check
 * @returns True if the value is a date
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isDate(value)) {
 *   // value is Date here
 *   console.log(value.toISOString());
 * }
 * ```
 */
export function isDate(value: unknown): value is Date {
  return value instanceof Date && !isNaN(value.getTime());
}

/**
 * Type guard to check if a value is a valid ISO date string
 * @param value - The value to check
 * @returns True if the value is a valid ISO date string
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isISODateString(value)) {
 *   // value is a valid ISO date string here
 *   console.log(new Date(value));
 * }
 * ```
 */
export function isISODateString(value: unknown): value is string {
  if (!isString(value)) return false;
  
  try {
    const date = new Date(value);
    return !isNaN(date.getTime()) && value.includes('T');
  } catch {
    return false;
  }
}

/**
 * Type guard to check if a value is a valid ProcessInfo object
 * @param value - Value to check
 * @returns True if the value is a valid ProcessInfo object
 */
export function isProcessInfo(value: unknown): value is ProcessInfo {
  return (
    isObject(value) &&
    'pid' in value && typeof value.pid === 'number' &&
    'name' in value && typeof value.name === 'string' &&
    'cpu_percent' in value && typeof value.cpu_percent === 'number' &&
    'mem_percent' in value && typeof value.mem_percent === 'number'
  );
}

/**
 * Type guard to check if a value is a valid SystemMetrics object
 * @param value - Value to check
 * @returns True if the value is a valid SystemMetrics object
 */
export function isSystemMetrics(value: unknown): value is SystemMetrics {
  return (
    isObject(value) &&
    'cpu' in value && isObject(value.cpu) &&
    'memory' in value && isObject(value.memory) &&
    'network' in value && isObject(value.network) &&
    'processes' in value && isArray(value.processes) &&
    'timestamp' in value && isString(value.timestamp)
  );
}

/**
 * Type guard to check if a value has a specific property
 * @template K - The property key type
 * @param value - The value to check
 * @param prop - The property to check for
 * @returns True if the value has the specified property
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (hasProperty(value, 'id')) {
 *   // value has an 'id' property here
 *   console.log(value.id);
 * }
 * ```
 */
export function hasProperty<K extends string>(
  value: unknown,
  prop: K
): value is { [key in K]: unknown } {
  return isObject(value) && prop in value;
}

/**
 * Safely asserts a value to a specific type
 * @template T - The expected type
 * @param value - The value to assert
 * @param typeGuard - The type guard function to use
 * @param defaultValue - The default value to return if the assertion fails
 * @returns The value cast to type T, or the default value
 * @example
 * ```ts
 * const value: unknown = getData();
 * const safeString = assertType(value, isString, '');
 * console.log(safeString.toUpperCase());
 * ```
 */
export function assertType<T>(
  value: unknown,
  typeGuard: (val: unknown) => val is T,
  defaultValue: T
): T {
  return typeGuard(value) ? value : defaultValue;
}

/**
 * Safely get a property from an object with type checking
 * @template T - The expected property type
 * @param obj - The object to get the property from
 * @param key - The property key
 * @param typeGuard - The type guard function to use
 * @param defaultValue - The default value to return if the property doesn't exist or fails type check
 * @returns The property value cast to type T, or the default value
 * @example
 * ```ts
 * const obj: Record<string, unknown> = getData();
 * const name = getTypedProperty(obj, 'name', isString, 'Unknown');
 * console.log(name.toUpperCase());
 * ```
 */
export function getTypedProperty<T>(
  obj: Record<string, unknown> | null | undefined,
  key: string,
  typeGuard: (val: unknown) => val is T,
  defaultValue: T
): T {
  if (!obj || !Object.prototype.hasOwnProperty.call(obj, key)) {
    return defaultValue;
  }
  
  return typeGuard(obj[key]) ? obj[key] as T : defaultValue;
}

/**
 * Type guard to check if all elements in an array satisfy a predicate
 * @template T - The expected element type
 * @param value - The array to check
 * @param predicate - The predicate function to test each element
 * @returns True if all elements satisfy the predicate
 * @example
 * ```ts
 * const values: unknown[] = getData();
 * if (isArrayOf(values, isString)) {
 *   // values is string[] here
 *   values.forEach(v => console.log(v.toUpperCase()));
 * }
 * ```
 */
export function isArrayOf<T>(
  value: unknown,
  predicate: (val: unknown) => val is T
): value is T[] {
  return isArray(value) && value.every(predicate);
}

/**
 * Type guard to check if a value is a Record with specific key and value types
 * @template K - The key type
 * @template V - The value type
 * @param value - The value to check
 * @param keyPredicate - The predicate function to test each key
 * @param valuePredicate - The predicate function to test each value
 * @returns True if the value is a Record with the specified key and value types
 * @example
 * ```ts
 * const value: unknown = getData();
 * if (isRecordOf(value, isString, isNumber)) {
 *   // value is Record<string, number> here
 *   Object.entries(value).forEach(([k, v]) => console.log(`${k}: ${v.toFixed(2)}`));
 * }
 * ```
 */
export function isRecordOf<K extends string | number | symbol, V>(
  value: unknown,
  keyPredicate: (key: unknown) => key is K,
  valuePredicate: (val: unknown) => val is V
): value is Record<K, V> {
  if (!isObject(value)) return false;
  
  return Object.entries(value).every(
    ([key, val]) => keyPredicate(key) && valuePredicate(val)
  );
} 