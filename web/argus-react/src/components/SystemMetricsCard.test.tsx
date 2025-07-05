import { describe, it, expect } from 'vitest';
import { screen } from '@testing-library/react';
import SystemMetricsCard, { MetricCardSkeleton } from './SystemMetricsCard';
import { renderWithProviders } from '../tests/utils/test-utils';

describe('SystemMetricsCard', () => {
  it('renders the title correctly', () => {
    renderWithProviders(
      <SystemMetricsCard title="CPU Usage" value={25.5} unit="%" loading={false} />
    );
    
    expect(screen.getByText('CPU Usage')).toBeInTheDocument();
  });

  it('renders the value and unit correctly', () => {
    renderWithProviders(
      <SystemMetricsCard title="CPU Usage" value={25.5} unit="%" loading={false} />
    );
    
    expect(screen.getByText('25.5%')).toBeInTheDocument();
  });

  it('renders details correctly', () => {
    const details = [
      { label: 'Load 1m', value: '1.20' },
      { label: 'Load 5m', value: '1.50' },
      { label: 'Load 15m', value: '1.80' }
    ];
    
    renderWithProviders(
      <SystemMetricsCard 
        title="CPU Usage" 
        value={25.5} 
        unit="%" 
        loading={false} 
        details={details} 
      />
    );
    
    expect(screen.getByText('Load 1m: 1.20')).toBeInTheDocument();
    expect(screen.getByText('Load 5m: 1.50')).toBeInTheDocument();
    expect(screen.getByText('Load 15m: 1.80')).toBeInTheDocument();
  });

  it('renders a skeleton when loading', () => {
    renderWithProviders(
      <SystemMetricsCard title="CPU Usage" value={25.5} unit="%" loading={true} />
    );
    
    // Skeletons don't have text content, so we check for absence of the title
    expect(screen.queryByText('CPU Usage')).not.toBeInTheDocument();
    
    // Check for skeleton elements
    const skeletons = document.querySelectorAll('.MuiSkeleton-root');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('renders with proper accessibility attributes', () => {
    renderWithProviders(
      <SystemMetricsCard 
        title="CPU Usage" 
        value={25.5} 
        unit="%" 
        loading={false} 
        titleId="cpu-title"
      />
    );
    
    const title = screen.getByText('CPU Usage');
    expect(title).toHaveAttribute('id', 'cpu-title');
    
    const section = document.querySelector('section');
    expect(section).toHaveAttribute('aria-labelledby', 'cpu-title');
  });

  it('MetricCardSkeleton renders correctly', () => {
    renderWithProviders(<MetricCardSkeleton />);
    
    const skeletons = document.querySelectorAll('.MuiSkeleton-root');
    expect(skeletons.length).toBeGreaterThan(0);
  });
}); 