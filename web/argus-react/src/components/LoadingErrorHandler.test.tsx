import { describe, it, expect } from 'vitest';
import { screen } from '@testing-library/react';
import LoadingErrorHandler from './LoadingErrorHandler';
import { renderWithProviders } from '../tests/utils/test-utils';

describe('LoadingErrorHandler', () => {
  it('renders loading state correctly', () => {
    renderWithProviders(
      <LoadingErrorHandler loading={true} error={null}>
        <div>Content</div>
      </LoadingErrorHandler>
    );
    
    expect(screen.getByText('Loading system metrics...')).toBeInTheDocument();
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.queryByText('Content')).not.toBeInTheDocument();
  });

  it('renders custom loading message', () => {
    renderWithProviders(
      <LoadingErrorHandler loading={true} error={null} loadingMessage="Custom loading message">
        <div>Content</div>
      </LoadingErrorHandler>
    );
    
    expect(screen.getByText('Custom loading message')).toBeInTheDocument();
  });

  it('renders error state correctly', () => {
    const errorMessage = 'Failed to fetch data';
    
    renderWithProviders(
      <LoadingErrorHandler loading={false} error={errorMessage}>
        <div>Content</div>
      </LoadingErrorHandler>
    );
    
    expect(screen.getByText('Error loading metrics')).toBeInTheDocument();
    expect(screen.getByText(errorMessage)).toBeInTheDocument();
    const alerts = screen.getAllByRole('alert');
    const liveAlert = alerts.find(alert => alert.getAttribute('aria-live') === 'assertive');
    expect(liveAlert).toBeInTheDocument();
    expect(screen.getByText('Retry')).toBeInTheDocument();
    expect(screen.queryByText('Content')).not.toBeInTheDocument();
  });

  it('renders children when not loading and no error', () => {
    renderWithProviders(
      <LoadingErrorHandler loading={false} error={null}>
        <div>Content</div>
      </LoadingErrorHandler>
    );
    
    expect(screen.getByText('Content')).toBeInTheDocument();
    expect(screen.queryByText('Loading system metrics...')).not.toBeInTheDocument();
    expect(screen.queryByText('Error loading metrics')).not.toBeInTheDocument();
  });

  it('has proper accessibility attributes', () => {
    renderWithProviders(
      <LoadingErrorHandler loading={true} error={null}>
        <div>Content</div>
      </LoadingErrorHandler>
    );
    
    const loadingElement = screen.getByRole('status');
    expect(loadingElement).toHaveAttribute('aria-live', 'polite');
    expect(loadingElement).toHaveAttribute('aria-busy', 'true');
  });

  it('retry button has accessible label', () => {
    renderWithProviders(
      <LoadingErrorHandler loading={false} error="Error message">
        <div>Content</div>
      </LoadingErrorHandler>
    );
    
    const retryButton = screen.getByText('Retry');
    expect(retryButton).toHaveAttribute('aria-label', 'Retry loading metrics');
  });
}); 