import '@testing-library/jest-dom';
import { afterEach, vi } from 'vitest';
import { cleanup } from '@testing-library/react';
import React from 'react';

// Extend matchers
declare global {
  namespace Vi {
    interface Assertion {
      toBeInTheDocument(): void;
      toBeVisible(): void;
      toHaveTextContent(text: string): void;
      toHaveClass(className: string): void;
    }
  }
}

// Mock Chart.js to prevent errors in tests
vi.mock('chart.js', () => ({
  Chart: {
    register: vi.fn(),
  },
  CategoryScale: vi.fn(),
  LinearScale: vi.fn(),
  PointElement: vi.fn(),
  LineElement: vi.fn(),
  BarElement: vi.fn(),
  ArcElement: vi.fn(),
  Title: vi.fn(),
  Tooltip: vi.fn(),
  Legend: vi.fn(),
}));

// Mock react-chartjs-2
vi.mock('react-chartjs-2', () => ({
  Line: () => React.createElement('div', { 'data-testid': 'line-chart' }, 'Line Chart'),
  Bar: () => React.createElement('div', { 'data-testid': 'bar-chart' }, 'Bar Chart'),
  Pie: () => React.createElement('div', { 'data-testid': 'pie-chart' }, 'Pie Chart'),
  Doughnut: () => React.createElement('div', { 'data-testid': 'doughnut-chart' }, 'Doughnut Chart'),
}));

// Mock react-window
vi.mock('react-window', () => ({
  FixedSizeList: ({ children, itemCount, itemData }: any) => {
    const items = [];
    for (let i = 0; i < Math.min(itemCount, 10); i++) {
      items.push(children({ index: i, style: {}, data: itemData }));
    }
    return React.createElement('div', { 'data-testid': 'virtualized-list' }, items);
  },
}));

// Mock react-virtualized-auto-sizer
vi.mock('react-virtualized-auto-sizer', () => ({
  default: ({ children }: any) => children({ width: 1000, height: 500 }),
}));

// Clean up after each test
afterEach(() => {
  cleanup();
}); 