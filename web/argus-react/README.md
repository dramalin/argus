# Argus React Frontend

This is the React frontend for the Argus System Monitor. It provides a modern, responsive UI for monitoring system resources and managing tasks.

## Technology Stack

Built with:
- React
- TypeScript
- Vite
- Material UI (MUI)
- Chart.js

The application uses a component-based architecture with responsive design to provide a seamless monitoring experience.

## UI Design System

Argus uses a custom Material Design theme based on a Morandi blue color palette. The theme provides a calm, professional interface that reduces eye strain during extended monitoring sessions.

### Theme Features

- Custom Morandi blue color palette
- Material Design components and styling
- Responsive layout system
- Accessibility-compliant design
- Dark mode support

For detailed information about the design system, color usage, typography, and component styling guidelines, refer to the [Style Guide](./src/theme/style-guide.md).

## Theme Usage

The application uses Material UI's theming system. The theme is defined in `src/theme/theme.ts` and uses the Morandi blue palette from `src/theme/morandiPalette.ts`.

To use the theme in your components:

```tsx
// Access theme in styled components
import { styled } from '@mui/material/styles';

const StyledComponent = styled('div')(({ theme }) => ({
  backgroundColor: theme.palette.background.paper,
  padding: theme.spacing(2),
}));

// Use the sx prop for inline styling
<Box
  sx={{
    bgcolor: 'primary.main',
    color: 'primary.contrastText',
    p: 2,
  }}
>
  Content
</Box>
```

## Development

To start the development server:

```bash
npm run dev
```

This will start a local development server with hot module replacement.

## Building for Production

To build the application for production:

```bash
npm run build
```

This will create an optimized production build in the `dist` directory. The build artifacts will be copied to the `web/release` directory by the project's Makefile.

## Integration with Go Backend

The application is integrated with the Go backend through API calls. The Go server is configured to:

1. Serve static assets from the `/assets` path
2. Serve the `index.html` file for all non-API routes (SPA fallback)
3. Serve API endpoints under the `/api` path

## ESLint Configuration

```js
// eslint.config.js
export default tseslint.config([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Remove tseslint.configs.recommended and replace with this
      ...tseslint.configs.recommendedTypeChecked,
      // Alternatively, use this for stricter rules
      ...tseslint.configs.strictTypeChecked,
      // Optionally, add this for stylistic rules
      ...tseslint.configs.stylisticTypeChecked,

      // Other configs...
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```

You can also install [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) for React-specific lint rules:

```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default tseslint.config([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...
      // Enable lint rules for React
      reactX.configs['recommended-typescript'],
      // Enable lint rules for React DOM
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```
