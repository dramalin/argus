import React, { Component, ErrorInfo, ReactNode } from 'react';
import { logError } from '../utils/errorHandling';

/**
 * Props for the ErrorBoundary component
 */
interface ErrorBoundaryProps {
  /** Child elements to render */
  children: ReactNode;
  /** Optional fallback component to render when an error occurs */
  fallback?: ReactNode;
  /** Optional callback function called when an error is caught */
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  /** Whether to reset the error state when the children prop changes */
  resetOnPropsChange?: boolean;
}

/**
 * State for the ErrorBoundary component
 */
interface ErrorBoundaryState {
  /** Whether an error has occurred */
  hasError: boolean;
  /** The error that occurred, if any */
  error: Error | null;
}

/**
 * ErrorBoundary component
 * 
 * Catches JavaScript errors anywhere in its child component tree,
 * logs those errors, and displays a fallback UI instead of the component tree that crashed.
 * 
 * @example
 * ```tsx
 * <ErrorBoundary fallback={<ErrorFallback message="Something went wrong" />}>
 *   <ComponentThatMightError />
 * </ErrorBoundary>
 * ```
 */
class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  /**
   * Default props for the ErrorBoundary component
   */
  static defaultProps = {
    resetOnPropsChange: false,
    fallback: (
      <div className="error-boundary-fallback">
        <h2>Something went wrong.</h2>
        <p>Please try refreshing the page or contact support if the problem persists.</p>
      </div>
    ),
  };

  /**
   * Initialize the component state
   */
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  /**
   * Update state when an error occurs during rendering
   */
  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  /**
   * Handle component errors
   */
  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    // Log the error to our error reporting service
    logError(error, { 
      component: 'ErrorBoundary',
      errorInfo,
      stack: error.stack 
    });
    
    // Call the onError callback if provided
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
  }

  /**
   * Reset error state when props change if resetOnPropsChange is true
   */
  componentDidUpdate(prevProps: ErrorBoundaryProps): void {
    if (
      this.state.hasError &&
      this.props.resetOnPropsChange &&
      prevProps.children !== this.props.children
    ) {
      this.setState({ hasError: false, error: null });
    }
  }

  /**
   * Render the component
   */
  render(): ReactNode {
    if (this.state.hasError) {
      return this.props.fallback;
    }

    return this.props.children;
  }
}

export default ErrorBoundary; 