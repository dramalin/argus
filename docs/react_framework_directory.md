# Argus React Framework Directory Structure

This document provides an overview of the optimized Argus React application structure after the refactoring and optimization process.

## Project Structure

```
argus-react/
├── dist/                  # Compiled output (generated on build)
├── node_modules/          # Dependencies
├── public/                # Static assets
├── src/                   # Source code
│   ├── components/        # Reusable UI components
│   ├── context/           # React context providers
│   ├── hooks/             # Custom React hooks
│   ├── routes/            # Route components
│   ├── tests/             # Test utilities and setup
│   ├── theme/             # Theme configuration
│   ├── types/             # TypeScript type definitions
│   ├── utils/             # Utility functions
│   ├── api.ts             # API service layer
│   ├── App.tsx            # Main application component
│   ├── Dashboard.tsx      # Dashboard page component
│   ├── main.tsx           # Application entry point
│   └── vite-env.d.ts      # Vite type declarations
├── .gitignore             # Git ignore configuration
├── .prettierrc            # Prettier configuration
├── BUILD.md               # Build instructions
├── eslint.config.js       # ESLint configuration
├── index.html             # HTML entry point
├── package.json           # Project dependencies and scripts
├── package-lock.json      # Locked dependencies
├── README.md              # Project documentation
├── tsconfig.app.json      # TypeScript configuration for app
├── tsconfig.json          # Base TypeScript configuration
├── tsconfig.node.json     # TypeScript configuration for Node
├── tsconfig.paths.json    # Path aliases configuration
├── vite.config.ts         # Vite bundler configuration
└── vitest.config.ts       # Vitest test configuration
```

## Key Directories

### Components

The `components/` directory contains reusable UI components, each focused on a specific responsibility:

```
components/
├── ChartWidget.tsx         # Wrapper for chart.js with accessibility features
├── ErrorBoundary.tsx       # Error boundary component for graceful error handling
├── Layout.tsx              # Main application layout
├── LoadingErrorHandler.tsx # Handles loading and error states
├── LoadingFallback.tsx     # Loading indicator component
├── MetricsCharts.tsx       # Charts for system metrics
├── ProcessTable.tsx        # Virtualized table for process data
├── SystemMetricsCard.tsx   # Card component for system metrics
└── SystemOverview.tsx      # System overview component
```

### Context

The `context/` directory contains React context providers for state management:

```
context/
├── AppProvider.tsx        # Root context provider
├── MetricsContext.tsx     # Context for system metrics data
├── ProcessesContext.tsx   # Context for process data
└── UiContext.tsx          # Context for UI state
```

### Hooks

The `hooks/` directory contains custom React hooks:

```
hooks/
├── useApiCache.ts          # Hook for API response caching
├── useDebounce.ts          # Hook for debouncing values
├── useMetrics.ts           # Hook for accessing metrics data
├── useProcesses.ts         # Hook for accessing process data
└── useThrottledCallback.ts # Hook for throttling function calls
```

### Routes

The `routes/` directory contains route components:

```
routes/
├── Alerts.tsx             # Alerts page
├── index.tsx              # Routes configuration
├── NotFound.tsx           # 404 page
├── Settings.tsx           # Settings page
└── Tasks.tsx              # Tasks page
```

### Types

The `types/` directory contains TypeScript type definitions:

```
types/
├── api.ts                 # API response types
├── common.ts              # Common utility types
├── context.ts             # Context types
├── hooks.ts               # Hook types
├── index.ts               # Type exports
└── process.ts             # Process data types
```

### Utils

The `utils/` directory contains utility functions:

```
utils/
├── errorHandling.ts       # Error handling utilities
├── index.ts               # Utility exports
├── lazyImport.ts          # Utilities for lazy loading
├── typeGuards.ts          # TypeScript type guards
└── validation.ts          # Data validation utilities
```

## Build Configuration

The project uses Vite as its build tool with the following configuration files:

- `vite.config.ts`: Main Vite configuration with optimizations for production builds
- `tsconfig.json`: Base TypeScript configuration
- `tsconfig.app.json`: Application-specific TypeScript settings
- `tsconfig.paths.json`: Path alias configurations
- `vitest.config.ts`: Test configuration for Vitest

## Testing

The project uses Vitest and React Testing Library for testing:

```
tests/
├── setup.ts               # Test setup configuration
└── utils/                 # Test utilities
    ├── api-mocks.ts       # API mocking utilities
    └── test-utils.tsx     # Testing utilities
```

## Performance Optimizations

The codebase includes several performance optimizations:

1. **Code Splitting**: Using React.lazy and dynamic imports
2. **Memoization**: Using React.memo, useMemo, and useCallback
3. **Virtualization**: For handling large datasets in tables
4. **API Caching**: To reduce unnecessary network requests
5. **Bundle Optimization**: With Vite's production build features

## Type Safety

The project uses TypeScript with strict type checking and includes:

1. **Utility Types**: Enhanced common types (Nullable<T>, Optional<T>, etc.)
2. **Type Guards**: Specialized functions for runtime type checking
3. **Validation Utilities**: For data validation
4. **JSDoc Comments**: For improved developer experience
