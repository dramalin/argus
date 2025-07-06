import { describe, it, expect, vi } from 'vitest';
import { screen, within, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ProcessTable from './ProcessTable';
import { renderWithProviders } from '../tests/utils/test-utils';
import { mockSystemMetrics } from '../tests/utils/test-utils';

describe('ProcessTable', () => {
  const defaultProps = {
    processes: mockSystemMetrics.processes,
    processParams: {
      limit: 10,
      offset: 0,
      sort_by: 'cpu',
      sort_order: 'desc' as const,
    },
    processTotal: mockSystemMetrics.processes.length,
    processLoading: false,
    processError: null,
    lastUpdated: new Date().toISOString(),
    onParamChange: vi.fn(),
    onResetFilters: vi.fn(),
  };

  it('renders the process table with headers', () => {
    renderWithProviders(<ProcessTable {...defaultProps} />);
    
    expect(screen.getByText('PID')).toBeInTheDocument();
    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('CPU %')).toBeInTheDocument();
    expect(screen.getByText('Memory %')).toBeInTheDocument();
  });

  it('renders process data correctly', () => {
    renderWithProviders(<ProcessTable {...defaultProps} />);
    
    // Check if virtualized list is rendered
    expect(screen.getByTestId('virtualized-list')).toBeInTheDocument();
    
    // Check if process data is displayed
    expect(screen.getByText('process1')).toBeInTheDocument();
    expect(screen.getByText('10.5')).toBeInTheDocument();
  });

  it('shows loading state when loading', () => {
    renderWithProviders(<ProcessTable {...defaultProps} processLoading={true} processes={[]} />);
    
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
    expect(screen.queryByText('PID')).not.toBeInTheDocument();
  });

  it('shows error state when there is an error', () => {
    renderWithProviders(
      <ProcessTable {...defaultProps} processError="Failed to load processes" />
    );
    
    expect(screen.getByText('Failed to load processes')).toBeInTheDocument();
    expect(screen.queryByText('PID')).not.toBeInTheDocument();
  });

  it('calls onParamChange when filter inputs change', async () => {
    const onParamChange = vi.fn();
    
    renderWithProviders(
      <ProcessTable {...defaultProps} onParamChange={onParamChange} />
    );
    
    // Filter by name
    const nameInput = screen.getByLabelText('Filter by name');
    fireEvent.change(nameInput, { target: { value: 'test' } });
    expect(onParamChange).toHaveBeenCalledWith('name_contains', 'test');
    
    // Filter by CPU
    const cpuInput = screen.getByLabelText('Min CPU %');
    fireEvent.change(cpuInput, { target: { value: '50' } });
    expect(onParamChange).toHaveBeenCalledWith('min_cpu', 50);
  });

  it('calls onResetFilters when reset button is clicked', async () => {
    const onResetFilters = vi.fn();
    const user = userEvent.setup();
    
    renderWithProviders(
      <ProcessTable {...defaultProps} onResetFilters={onResetFilters} />
    );
    
    const resetButton = screen.getByText('Reset Filters');
    await user.click(resetButton);
    expect(onResetFilters).toHaveBeenCalled();
  });

  it('displays pagination correctly', () => {
    renderWithProviders(
      <ProcessTable {...defaultProps} processTotal={30} />
    );
    
    // With 30 total items and 10 per page, we should have 3 pages
    const pagination = screen.getByRole('navigation');
    expect(within(pagination).getAllByRole('button').length).toBeGreaterThan(3); // 3 pages + navigation buttons
  });

  it('displays total process count', () => {
    renderWithProviders(<ProcessTable {...defaultProps} />);
    
    expect(screen.getByText(`${defaultProps.processTotal} total processes`)).toBeInTheDocument();
  });

  it('displays last updated time', () => {
    renderWithProviders(<ProcessTable {...defaultProps} />);
    
    expect(screen.getByText(/Updated:/)).toBeInTheDocument();
  });
}); 