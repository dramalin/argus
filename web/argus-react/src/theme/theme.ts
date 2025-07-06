import { createTheme, Theme } from '@mui/material/styles';
import { morandiPalette } from './morandiPalette';

// Example additional color tone: Cobalt
const cobaltPalette = {
  primary: {
    main: '#0047AB', // Cobalt blue
    light: '#5B8FF9',
    dark: '#002D62',
    contrastText: '#fff',
  },
  secondary: {
    main: '#7F8FA6',
    light: '#B2BEC3',
    dark: '#353B48',
    contrastText: '#fff',
  },
  background: {
    default: '#F4F6F8',
    paper: '#E9ECF1',
  },
  error: { main: '#B97A7A', contrastText: '#fff' },
  info: { main: '#7A9EB9', contrastText: '#fff' },
  success: { main: '#7AA29E', contrastText: '#fff' },
  warning: { main: '#B9A97A', contrastText: '#fff' },
  text: { primary: '#2D3142', secondary: '#0047AB', disabled: '#A3B1C6' },
  divider: '#C7D6D9',
};

// Light and dark mode palettes for each tone
const palettes = {
  morandi: {
    light: morandiPalette,
    dark: {
      ...morandiPalette,
      background: { default: '#23272F', paper: '#2D3142' },
      text: { primary: '#F4F6F8', secondary: '#A3B1C6', disabled: '#49587A' },
    },
  },
  cobalt: {
    light: cobaltPalette,
    dark: {
      ...cobaltPalette,
      background: { default: '#1A2236', paper: '#232B3E' },
      text: { primary: '#F4F6F8', secondary: '#5B8FF9', disabled: '#353B48' },
    },
  },
};

export type ColorTone = 'morandi' | 'cobalt';
export type ThemeMode = 'light' | 'dark';

export function getTheme(tone: ColorTone = 'morandi', mode: ThemeMode = 'light'): Theme {
  const palette = palettes[tone][mode];
  return createTheme({
    palette: {
      ...palette,
      mode,
    },
    typography: {
      fontFamily: [
        '-apple-system',
        'BlinkMacSystemFont',
        '"Segoe UI"',
        'Roboto',
        '"Helvetica Neue"',
        'Arial',
        'sans-serif',
        '"Apple Color Emoji"',
        '"Segoe UI Emoji"',
        '"Segoe UI Symbol"',
      ].join(','),
      h1: { fontSize: '2.5rem', fontWeight: 500, lineHeight: 1.2 },
      h2: { fontSize: '2rem', fontWeight: 500, lineHeight: 1.3 },
      h3: { fontSize: '1.5rem', fontWeight: 500, lineHeight: 1.4 },
      h4: { fontSize: '1.25rem', fontWeight: 500, lineHeight: 1.5 },
      h5: { fontSize: '1rem', fontWeight: 500, lineHeight: 1.5 },
      h6: { fontSize: '0.875rem', fontWeight: 500, lineHeight: 1.5 },
      body1: { fontSize: '1rem', lineHeight: 1.5 },
      body2: { fontSize: '0.875rem', lineHeight: 1.5 },
    },
    components: {
      MuiButton: {
        styleOverrides: {
          root: { borderRadius: 8, textTransform: 'none', fontWeight: 500 },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: { borderRadius: 10, boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)' },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: { boxShadow: '0 2px 10px rgba(0, 0, 0, 0.1)' },
        },
      },
    },
    shape: { borderRadius: 8 },
  });
} 