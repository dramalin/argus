# Argus UI Style Guide - Morandi Blue Theme

This style guide documents the UI design system for Argus, based on Material Design with a custom Morandi blue palette. It provides guidelines for consistent application of colors, typography, spacing, and components.

## Color System

Our color system is based on a Morandi blue palette - muted, elegant, and harmonious blues with complementary shades. These colors are designed to create a calm, professional interface that reduces eye strain during extended monitoring sessions.

### Color Tokens

| Role         | Main      | Light     | Dark      | Contrast Text |
|--------------|-----------|-----------|-----------|---------------|
| Primary      | #6A7BA2   | #A3B1C6   | #49587A   | #fff          |
| Secondary    | #A2B6B9   | #C7D6D9   | #6B7C7E   | #fff          |
| Background   | #F4F6F8   |           |           |               |
| Paper/Surface| #E9ECF1   |           |           |               |
| Error        | #B97A7A   |           |           | #fff          |
| Info         | #7A9EB9   |           |           | #fff          |
| Success      | #7AA29E   |           |           | #fff          |
| Warning      | #B9A97A   |           |           | #fff          |
| Text Primary | #2D3142   |           |           |               |
| Text Second. | #6A7BA2   |           |           |               |
| Text Disabld | #A3B1C6   |           |           |               |
| Divider      | #C7D6D9   |           |           |               |

### Color Usage Guidelines

- **Primary**: Use for main actions, navigation elements, and key UI components
- **Secondary**: Use for complementary actions, less prominent UI elements
- **Background**: Use for page backgrounds and large surface areas
- **Paper/Surface**: Use for cards, dialogs, and elevated surfaces
- **Status Colors**: Use error, warning, info, and success colors for their respective states
- **Text**: Follow the hierarchy of text colors for different levels of importance

## Typography

Our typography system is based on the standard Material Design type scale, with some customizations for better readability in monitoring contexts.

### Font Family

```
'-apple-system',
'BlinkMacSystemFont',
'"Segoe UI"',
'Roboto',
'"Helvetica Neue"',
'Arial',
'sans-serif',
'"Apple Color Emoji"',
'"Segoe UI Emoji"',
'"Segoe UI Symbol"'
```

### Type Scale

| Element | Size      | Weight | Line Height |
|---------|-----------|--------|-------------|
| h1      | 2.5rem    | 500    | 1.2         |
| h2      | 2rem      | 500    | 1.3         |
| h3      | 1.5rem    | 500    | 1.4         |
| h4      | 1.25rem   | 500    | 1.5         |
| h5      | 1rem      | 500    | 1.5         |
| h6      | 0.875rem  | 500    | 1.5         |
| body1   | 1rem      | 400    | 1.5         |
| body2   | 0.875rem  | 400    | 1.5         |

### Typography Usage Guidelines

- Use appropriate heading levels to maintain semantic hierarchy
- Maintain consistent text styles across similar UI elements
- Avoid using too many different text styles in a single view
- For data-dense displays, prefer body2 to maintain readability

## Component Styling

Our components follow Material Design principles with custom styling to align with our Morandi blue theme.

### Buttons

- Rounded corners (8px border radius)
- No text transformation (preserve case as written)
- Medium font weight (500)
- Use primary color for primary actions
- Use secondary color for secondary actions

```tsx
// Primary button example
<Button variant="contained" color="primary">
  Submit
</Button>

// Secondary button example
<Button variant="outlined" color="secondary">
  Cancel
</Button>
```

### Cards

- Slightly more rounded corners (10px border radius)
- Subtle shadow for depth (0 4px 6px rgba(0, 0, 0, 0.1))
- Use Paper/Surface color for background
- Consistent padding (16px recommended)

```tsx
<Card>
  <CardContent>
    <Typography variant="h5">Card Title</Typography>
    <Typography variant="body2">Card content goes here</Typography>
  </CardContent>
  <CardActions>
    <Button size="small">Action</Button>
  </CardActions>
</Card>
```

### App Bar

- Subtle shadow (0 2px 10px rgba(0, 0, 0, 0.1))
- Use primary color
- Consistent height and padding

```tsx
<AppBar position="static">
  <Toolbar>
    <Typography variant="h6">Argus Monitor</Typography>
  </Toolbar>
</AppBar>
```

## Layout & Spacing

- Use the MUI spacing system (theme.spacing()) for consistent spacing
- Base spacing unit is 8px (theme.spacing(1) = 8px)
- Use responsive breakpoints for adaptive layouts
- Maintain consistent padding and margins across similar components

```tsx
// Example of consistent spacing
<Box sx={{ p: 2, m: 1 }}>
  <Typography variant="body1" sx={{ mb: 2 }}>
    Content with consistent spacing
  </Typography>
</Box>
```

## Accessibility Guidelines

- All UI components meet WCAG 2.1 AA standards
- Color contrast ratios meet or exceed 4.5:1 for normal text
- Interactive elements have clear focus indicators
- All interactive elements are keyboard accessible
- Use semantic HTML elements and appropriate ARIA attributes

## Best Practices

1. **Theme Consistency**
   - Always use theme tokens instead of hardcoded values
   - Access colors through the theme: `theme.palette.primary.main`
   - Access typography through the theme: `theme.typography.body1`

2. **Component Usage**
   - Use MUI components whenever possible for consistency
   - Extend components using the `sx` prop or `styled` API rather than creating custom components
   - Follow the component API documentation for proper usage

3. **Responsive Design**
   - Use responsive breakpoints for layout adjustments
   - Test all UI components across different screen sizes
   - Use the MUI Grid system for complex layouts

4. **Dark Mode Compatibility**
   - The theme is designed to work with both light and dark modes
   - Test components in both modes when implementing new features

## Code Examples

### Theme Usage

```tsx
// Accessing theme in styled components
import { styled } from '@mui/material/styles';

const StyledComponent = styled('div')(({ theme }) => ({
  backgroundColor: theme.palette.background.paper,
  padding: theme.spacing(2),
  borderRadius: theme.shape.borderRadius,
}));
```

### Using the sx Prop

```tsx
// Using the sx prop for styling
<Box
  sx={{
    bgcolor: 'background.paper',
    p: 2,
    borderRadius: 1,
    boxShadow: 1,
  }}
>
  Content
</Box>
```

### Custom Theme Extension

```tsx
// Extending the theme for a specific component
import { createTheme } from '@mui/material/styles';
import theme from '../theme/theme';

const extendedTheme = createTheme({
  ...theme,
  components: {
    ...theme.components,
    MuiButton: {
      ...theme.components.MuiButton,
      styleOverrides: {
        root: {
          ...theme.components.MuiButton?.styleOverrides?.root,
          // Additional custom styles
        },
      },
    },
  },
});
```

---

_Argus UI Modernization, 2024-07-04_
