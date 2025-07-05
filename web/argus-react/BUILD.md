# Argus React Build Configuration

This document provides information about the build configuration and development tooling for the Argus React application.

## Build Scripts

The following npm scripts are available:

- `npm run dev`: Start the development server
- `npm run build`: Build the application for production
- `npm run build:analyze`: Build the application and generate a bundle analysis report
- `npm run preview`: Preview the production build locally
- `npm run start`: Start a production-like server for the built application
- `npm run analyze`: Build the application and open the bundle analysis report

## Code Quality Tools

The following tools are configured for code quality:

- **ESLint**: Lints JavaScript and TypeScript code
  - `npm run lint`: Check for linting issues
  - `npm run lint:fix`: Fix linting issues automatically
  - `npm run lint:report`: Generate a linting report

- **Prettier**: Formats code consistently
  - `npm run format`: Format all code files
  - `npm run format:check`: Check if files are formatted correctly

- **TypeScript**: Type checks the codebase
  - `npm run typecheck`: Run TypeScript type checking without emitting files

## Pre-commit Hooks

Husky is configured with pre-commit hooks to ensure code quality:

- Runs ESLint and Prettier on staged files
- Runs TypeScript type checking

## Environment Variables

The application uses the following environment variables:

- `VITE_API_BASE_URL`: Base URL for API requests
- `VITE_API_TIMEOUT`: Timeout for API requests in milliseconds
- `VITE_ENABLE_WEBSOCKETS`: Whether to enable WebSocket connections
- `VITE_ENABLE_ANALYTICS`: Whether to enable analytics
- `VITE_DEFAULT_THEME`: Default theme (light or dark)
- `VITE_DEFAULT_REFRESH_INTERVAL`: Default data refresh interval in milliseconds
- `VITE_APP_TITLE`: Application title
- `VITE_APP_DESCRIPTION`: Application description

Environment variables can be configured in the following files:

- `.env`: Default environment variables
- `.env.development`: Development-specific variables
- `.env.production`: Production-specific variables

## Build Optimization

The build is optimized in the following ways:

- **Code Splitting**: Automatically splits vendor code and application code
- **Bundle Analysis**: Visualizes bundle size with `npm run analyze`
- **Compression**: Automatically compresses assets with Brotli and Gzip
- **Tree Shaking**: Removes unused code
- **Minification**: Minifies JavaScript, CSS, and HTML
- **Source Maps**: Generates source maps for debugging (disabled in production)

## Path Aliases

The following path aliases are configured for cleaner imports:

- `@/*`: Points to `src/*`
- `@components/*`: Points to `src/components/*`
- `@hooks/*`: Points to `src/hooks/*`
- `@utils/*`: Points to `src/utils/*`
- `@contexts/*`: Points to `src/contexts/*`
- `@types`: Points to `src/types`

Example usage:

```typescript
import { Button } from '@components/Button';
import { useMetrics } from '@hooks/useMetrics';
import { formatBytes } from '@utils/formatters';
```
