import React, { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { ThemeProvider } from '@mui/material/styles';
import { CssBaseline } from '@mui/material';
import userEvent from '@testing-library/user-event';
import theme from '../../theme/theme';
import AppProvider from '../../context/AppProvider';

/**
 * Custom render function that includes providers
 */
interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  withTheme?: boolean;
  withAppProvider?: boolean;
}

/**
 * Custom render function that includes providers
 * @param ui - Component to render
 * @param options - Render options
 * @returns Rendered component with testing utilities
 */
export function renderWithProviders(
  ui: ReactElement,
  {
    withTheme = true,
    withAppProvider = true,
    ...renderOptions
  }: CustomRenderOptions = {}
) {
  const AllProviders = ({ children }: { children: React.ReactNode }) => {
    let wrappedChildren = children;

    // Wrap with AppProvider if requested
    if (withAppProvider) {
      wrappedChildren = <AppProvider>{wrappedChildren}</AppProvider>;
    }

    // Wrap with ThemeProvider if requested
    if (withTheme) {
      wrappedChildren = (
        <ThemeProvider theme={theme}>
          <CssBaseline />
          {wrappedChildren}
        </ThemeProvider>
      );
    }

    return <>{wrappedChildren}</>;
  };

  return {
    user: userEvent.setup(),
    ...render(ui, { wrapper: AllProviders, ...renderOptions }),
  };
}

/**
 * Mock system metrics data for testing
 */
export const mockSystemMetrics = {
  cpu: {
    load1: 1.2,
    load5: 1.5,
    load15: 1.8,
    usage_percent: 25.5,
  },
  memory: {
    total: 16000000000,
    used: 8000000000,
    free: 8000000000,
    used_percent: 50.0,
  },
  network: {
    bytes_sent: 1000000,
    bytes_recv: 2000000,
    packets_sent: 1000,
    packets_recv: 2000,
  },
  processes: [
    { pid: 1, name: 'process1', cpu_percent: 10.5, mem_percent: 5.2 },
    { pid: 2, name: 'process2', cpu_percent: 5.3, mem_percent: 2.1 },
    { pid: 3, name: 'process3', cpu_percent: 15.7, mem_percent: 7.8 },
  ],
};

/**
 * Mock API response for testing
 */
export const mockApiResponse = {
  data: mockSystemMetrics,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {},
};

/**
 * Wait for a specified time
 * @param ms - Milliseconds to wait
 * @returns Promise that resolves after the specified time
 */
export const wait = (ms: number) => new Promise(resolve => setTimeout(resolve, ms)); 