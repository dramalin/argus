import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import AlertDialog from './AlertDialog';

vi.mock('../../hooks/useMetrics', () => ({
  useMetrics: () => ({
    metrics: {},
    loading: false,
    error: null,
    lastUpdated: new Date().toISOString(),
  }),
}));

vi.mock('../../hooks/useProcesses', () => ({
  useProcesses: () => ({
    processes: [],
    total: 0,
    loading: false,
    error: null,
    params: {},
    setParams: () => {},
    resetFilters: () => {},
  }),
}));

describe('AlertDialog', () => {
  const onSave = vi.fn();
  const onClose = vi.fn();

  it('renders email recipient field for email notifications', () => {
    render(<AlertDialog open={true} onClose={onClose} onSave={onSave} />);
    
    // Change notification type to email
    fireEvent.mouseDown(screen.getByLabelText('Type'));
    fireEvent.click(screen.getByText('Email'));

    expect(screen.getByLabelText('Recipient Email')).toBeInTheDocument();
  });

  it('renders process target field for process metrics', () => {
    render(<AlertDialog open={true} onClose={onClose} onSave={onSave} />);
    
    // Change metric type to process
    fireEvent.mouseDown(screen.getByLabelText('Metric Type'));
    fireEvent.click(screen.getByText('Process'));

    expect(screen.getByLabelText('Process Name or PID')).toBeInTheDocument();
  });

  it('submits new fields correctly', async () => {
    render(<AlertDialog open={true} onClose={onClose} onSave={onSave} />);
    
    // Fill in basic info
    fireEvent.change(screen.getByLabelText('Alert Name'), { target: { value: 'Test Alert' } });

    // Select email notification
    fireEvent.mouseDown(screen.getAllByLabelText('Type')[0]);
    fireEvent.click(screen.getByText('Email'));
    fireEvent.change(screen.getByLabelText('Recipient Email'), { target: { value: 'test@example.com' } });

    // Select process metric
    fireEvent.mouseDown(screen.getByLabelText('Metric Type'));
    fireEvent.click(screen.getByText('Process'));
    fireEvent.change(screen.getByLabelText('Process Name or PID'), { target: { value: 'my-process' } });

    // Save
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'Test Alert',
        threshold: expect.objectContaining({
          metric_type: 'process',
          target: 'my-process'
        }),
        notifications: expect.arrayContaining([
          expect.objectContaining({
            type: 'email',
            settings: { recipient: 'test@example.com' }
          })
        ])
      })
    );
  });
}); 