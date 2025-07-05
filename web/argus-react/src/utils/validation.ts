/**
 * Validation utilities
 * @module Validation
 */
import { ValidationResult } from '../types/common';
import { isString, isNumber, isObject, isArray, isNonEmptyString } from './typeGuards';
import { ValidationError } from './errorHandling';

/**
 * Validate a string value
 * @param value - The value to validate
 * @param options - Validation options
 * @returns Validation result
 */
export function validateString(
  value: unknown,
  options: {
    required?: boolean;
    minLength?: number;
    maxLength?: number;
    pattern?: RegExp;
    allowEmpty?: boolean;
    fieldName?: string;
  } = {}
): ValidationResult {
  const {
    required = false,
    minLength,
    maxLength,
    pattern,
    allowEmpty = false,
    fieldName = 'Value'
  } = options;

  // Check if value is required
  if (required && (value === undefined || value === null)) {
    return { valid: false, error: `${fieldName} is required` };
  }

  // If not required and not provided, it's valid
  if (!required && (value === undefined || value === null)) {
    return { valid: true, error: null };
  }

  // Check if value is a string
  if (!isString(value)) {
    return { valid: false, error: `${fieldName} must be a string` };
  }

  // Check if empty string is allowed
  if (!allowEmpty && value.trim() === '') {
    return { valid: false, error: `${fieldName} cannot be empty` };
  }

  // Check minimum length
  if (minLength !== undefined && value.length < minLength) {
    return {
      valid: false,
      error: `${fieldName} must be at least ${minLength} characters`
    };
  }

  // Check maximum length
  if (maxLength !== undefined && value.length > maxLength) {
    return {
      valid: false,
      error: `${fieldName} must be no more than ${maxLength} characters`
    };
  }

  // Check pattern
  if (pattern && !pattern.test(value)) {
    return { valid: false, error: `${fieldName} has an invalid format` };
  }

  return { valid: true, error: null };
}

/**
 * Validate a number value
 * @param value - The value to validate
 * @param options - Validation options
 * @returns Validation result
 */
export function validateNumber(
  value: unknown,
  options: {
    required?: boolean;
    min?: number;
    max?: number;
    integer?: boolean;
    positive?: boolean;
    fieldName?: string;
  } = {}
): ValidationResult {
  const {
    required = false,
    min,
    max,
    integer = false,
    positive = false,
    fieldName = 'Value'
  } = options;

  // Check if value is required
  if (required && (value === undefined || value === null)) {
    return { valid: false, error: `${fieldName} is required` };
  }

  // If not required and not provided, it's valid
  if (!required && (value === undefined || value === null)) {
    return { valid: true, error: null };
  }

  // Check if value is a number
  if (!isNumber(value)) {
    return { valid: false, error: `${fieldName} must be a number` };
  }

  // Check if value is an integer
  if (integer && !Number.isInteger(value)) {
    return { valid: false, error: `${fieldName} must be an integer` };
  }

  // Check if value is positive
  if (positive && value <= 0) {
    return { valid: false, error: `${fieldName} must be positive` };
  }

  // Check minimum value
  if (min !== undefined && value < min) {
    return {
      valid: false,
      error: `${fieldName} must be at least ${min}`
    };
  }

  // Check maximum value
  if (max !== undefined && value > max) {
    return {
      valid: false,
      error: `${fieldName} must be no more than ${max}`
    };
  }

  return { valid: true, error: null };
}

/**
 * Validate an array value
 * @param value - The value to validate
 * @param options - Validation options
 * @returns Validation result
 */
export function validateArray<T>(
  value: unknown,
  options: {
    required?: boolean;
    minLength?: number;
    maxLength?: number;
    itemValidator?: (item: unknown, index: number) => ValidationResult;
    fieldName?: string;
  } = {}
): ValidationResult {
  const {
    required = false,
    minLength,
    maxLength,
    itemValidator,
    fieldName = 'Array'
  } = options;

  // Check if value is required
  if (required && (value === undefined || value === null)) {
    return { valid: false, error: `${fieldName} is required` };
  }

  // If not required and not provided, it's valid
  if (!required && (value === undefined || value === null)) {
    return { valid: true, error: null };
  }

  // Check if value is an array
  if (!isArray(value)) {
    return { valid: false, error: `${fieldName} must be an array` };
  }

  // Check minimum length
  if (minLength !== undefined && value.length < minLength) {
    return {
      valid: false,
      error: `${fieldName} must have at least ${minLength} items`
    };
  }

  // Check maximum length
  if (maxLength !== undefined && value.length > maxLength) {
    return {
      valid: false,
      error: `${fieldName} must have no more than ${maxLength} items`
    };
  }

  // Validate each item
  if (itemValidator) {
    const errors: string[] = [];
    
    for (let i = 0; i < value.length; i++) {
      const result = itemValidator(value[i], i);
      if (!result.valid && result.error) {
        errors.push(`Item ${i + 1}: ${result.error}`);
      }
    }
    
    if (errors.length > 0) {
      return {
        valid: false,
        error: `${fieldName} contains invalid items: ${errors.join('; ')}`
      };
    }
  }

  return { valid: true, error: null };
}

/**
 * Validate an object value
 * @param value - The value to validate
 * @param options - Validation options
 * @returns Validation result
 */
export function validateObject<T extends Record<string, unknown>>(
  value: unknown,
  options: {
    required?: boolean;
    schema?: Record<string, (value: unknown) => ValidationResult>;
    allowUnknownProperties?: boolean;
    fieldName?: string;
  } = {}
): ValidationResult {
  const {
    required = false,
    schema,
    allowUnknownProperties = true,
    fieldName = 'Object'
  } = options;

  // Check if value is required
  if (required && (value === undefined || value === null)) {
    return { valid: false, error: `${fieldName} is required` };
  }

  // If not required and not provided, it's valid
  if (!required && (value === undefined || value === null)) {
    return { valid: true, error: null };
  }

  // Check if value is an object
  if (!isObject(value)) {
    return { valid: false, error: `${fieldName} must be an object` };
  }

  // Validate against schema
  if (schema) {
    const errors: Record<string, string[]> = {};
    
    // Check required properties
    for (const [key, validator] of Object.entries(schema)) {
      const result = validator(value[key]);
      if (!result.valid && result.error) {
        if (!errors[key]) {
          errors[key] = [];
        }
        errors[key].push(result.error);
      }
    }
    
    // Check for unknown properties
    if (!allowUnknownProperties) {
      for (const key of Object.keys(value)) {
        if (!schema[key]) {
          if (!errors.unknownProperties) {
            errors.unknownProperties = [];
          }
          errors.unknownProperties.push(`Unknown property: ${key}`);
        }
      }
    }
    
    // If there are errors, return them
    if (Object.keys(errors).length > 0) {
      return {
        valid: false,
        error: `${fieldName} has validation errors`
      };
    }
  }

  return { valid: true, error: null };
}

/**
 * Validate an email address
 * @param value - The email address to validate
 * @param options - Validation options
 * @returns Validation result
 */
export function validateEmail(
  value: unknown,
  options: {
    required?: boolean;
    fieldName?: string;
  } = {}
): ValidationResult {
  const { required = false, fieldName = 'Email' } = options;

  // Check if value is required
  if (required && (value === undefined || value === null)) {
    return { valid: false, error: `${fieldName} is required` };
  }

  // If not required and not provided, it's valid
  if (!required && (value === undefined || value === null || value === '')) {
    return { valid: true, error: null };
  }

  // Check if value is a string
  if (!isString(value)) {
    return { valid: false, error: `${fieldName} must be a string` };
  }

  // Simple email validation regex
  // This is not perfect but catches most common errors
  const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
  if (!emailRegex.test(value)) {
    return { valid: false, error: `${fieldName} is not a valid email address` };
  }

  return { valid: true, error: null };
}

/**
 * Validate a URL
 * @param value - The URL to validate
 * @param options - Validation options
 * @returns Validation result
 */
export function validateUrl(
  value: unknown,
  options: {
    required?: boolean;
    protocols?: string[];
    fieldName?: string;
  } = {}
): ValidationResult {
  const {
    required = false,
    protocols = ['http', 'https'],
    fieldName = 'URL'
  } = options;

  // Check if value is required
  if (required && (value === undefined || value === null)) {
    return { valid: false, error: `${fieldName} is required` };
  }

  // If not required and not provided, it's valid
  if (!required && (value === undefined || value === null || value === '')) {
    return { valid: true, error: null };
  }

  // Check if value is a string
  if (!isString(value)) {
    return { valid: false, error: `${fieldName} must be a string` };
  }

  try {
    const url = new URL(value);
    
    // Check protocol
    if (protocols.length > 0 && !protocols.includes(url.protocol.replace(':', ''))) {
      return {
        valid: false,
        error: `${fieldName} must use one of these protocols: ${protocols.join(', ')}`
      };
    }
    
    return { valid: true, error: null };
  } catch {
    return { valid: false, error: `${fieldName} is not a valid URL` };
  }
}

/**
 * Validate a date string
 * @param value - The date string to validate
 * @param options - Validation options
 * @returns Validation result
 */
export function validateDate(
  value: unknown,
  options: {
    required?: boolean;
    minDate?: Date;
    maxDate?: Date;
    iso?: boolean;
    fieldName?: string;
  } = {}
): ValidationResult {
  const {
    required = false,
    minDate,
    maxDate,
    iso = false,
    fieldName = 'Date'
  } = options;

  // Check if value is required
  if (required && (value === undefined || value === null)) {
    return { valid: false, error: `${fieldName} is required` };
  }

  // If not required and not provided, it's valid
  if (!required && (value === undefined || value === null || value === '')) {
    return { valid: true, error: null };
  }

  // Check if value is a string
  if (!isString(value) && !(value instanceof Date)) {
    return { valid: false, error: `${fieldName} must be a string or Date object` };
  }

  // Parse the date
  let date: Date;
  if (value instanceof Date) {
    date = value;
  } else {
    date = new Date(value);
  }

  // Check if date is valid
  if (isNaN(date.getTime())) {
    return { valid: false, error: `${fieldName} is not a valid date` };
  }

  // Check ISO format
  if (iso && isString(value) && !/^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}:\d{2}(\.\d{3})?(Z|[+-]\d{2}:\d{2})?)?$/.test(value)) {
    return { valid: false, error: `${fieldName} must be in ISO format (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)` };
  }

  // Check minimum date
  if (minDate && date < minDate) {
    return {
      valid: false,
      error: `${fieldName} must be on or after ${minDate.toISOString().split('T')[0]}`
    };
  }

  // Check maximum date
  if (maxDate && date > maxDate) {
    return {
      valid: false,
      error: `${fieldName} must be on or before ${maxDate.toISOString().split('T')[0]}`
    };
  }

  return { valid: true, error: null };
}

/**
 * Validate multiple values and collect all errors
 * @param validations - Object with validation results
 * @returns Combined validation result and errors by field
 */
export function validateAll(
  validations: Record<string, ValidationResult>
): { valid: boolean; errors: Record<string, string> } {
  const errors: Record<string, string> = {};
  let valid = true;

  for (const [field, result] of Object.entries(validations)) {
    if (!result.valid && result.error) {
      errors[field] = result.error;
      valid = false;
    }
  }

  return { valid, errors };
}

/**
 * Throw a ValidationError if validation fails
 * @param validations - Object with validation results
 * @throws ValidationError if validation fails
 */
export function validateAndThrow(
  validations: Record<string, ValidationResult>
): void {
  const { valid, errors } = validateAll(validations);
  
  if (!valid) {
    const errorsByField: Record<string, string[]> = {};
    
    for (const [field, error] of Object.entries(errors)) {
      errorsByField[field] = [error];
    }
    
    throw new ValidationError('Validation failed', errorsByField);
  }
}

/**
 * Create a validator function for a specific schema
 * @param schema - Validation schema
 * @returns A validator function
 */
export function createValidator<T extends Record<string, unknown>>(
  schema: Record<string, (value: unknown) => ValidationResult>
): (data: unknown) => { valid: boolean; errors: Record<string, string>; data: T | null } {
  return (data: unknown) => {
    // Check if data is an object
    if (!isObject(data)) {
      return {
        valid: false,
        errors: { _error: 'Data must be an object' },
        data: null
      };
    }

    // Validate each field
    const validations: Record<string, ValidationResult> = {};
    for (const [field, validator] of Object.entries(schema)) {
      validations[field] = validator(data[field]);
    }

    // Collect errors
    const { valid, errors } = validateAll(validations);

    return {
      valid,
      errors,
      data: valid ? (data as T) : null
    };
  };
} 