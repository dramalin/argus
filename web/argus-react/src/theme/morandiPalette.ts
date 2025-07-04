// Morandi Blue Material Design Palette
// This palette is inspired by Morandi blue tones: muted, elegant, and harmonious blues with complementary shades.
// Use these tokens for Material Design theme configuration.
// Author: Argus UI Modernization
// Date: 2024-07-03

export const morandiPalette = {
    primary: {
        main: '#6A7BA2', // Morandi blue (primary)
        light: '#A3B1C6',
        dark: '#49587A',
        contrastText: '#fff',
    },
    secondary: {
        main: '#A2B6B9', // Muted blue-green
        light: '#C7D6D9',
        dark: '#6B7C7E',
        contrastText: '#fff',
    },
    background: {
        default: '#F4F6F8', // Soft gray
        paper: '#E9ECF1',   // Slightly deeper for surfaces
    },
    surface: {
        main: '#E9ECF1', // For cards, sheets, etc.
    },
    error: {
        main: '#B97A7A', // Muted rose
        contrastText: '#fff',
    },
    text: {
        primary: '#2D3142', // Deep blue-gray
        secondary: '#6A7BA2', // Morandi blue
        disabled: '#A3B1C6',
    },
    divider: '#C7D6D9',
    info: {
        main: '#7A9EB9', // Muted blue
        contrastText: '#fff',
    },
    success: {
        main: '#7AA29E', // Muted teal
        contrastText: '#fff',
    },
    warning: {
        main: '#B9A97A', // Muted gold
        contrastText: '#fff',
    },
};

// Usage: import { morandiPalette } from './morandiPalette';
// Use these tokens in your Material UI theme or custom theming solution. 