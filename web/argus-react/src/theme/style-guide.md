# Morandi Blue Material Design Palette

This palette is inspired by Morandi blue tones: muted, elegant, and harmonious blues with complementary shades. It is designed for use with Material Design and Material UI (MUI) theming.

## Color Tokens

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

## Usage

- Import the palette:
  ```ts
  import { morandiPalette } from './morandiPalette';
  ```
- Use these tokens in your Material UI theme or custom theming solution.
- Assign colors to Material Design roles (primary, secondary, background, etc.) as shown above.

## Integration Notes

- For Material UI, use the palette in the `createTheme` function.
- Ensure all UI components reference theme tokens, not hardcoded colors.
- For custom CSS, use these colors as CSS variables or in style definitions.

## Accessibility

- All colors have been selected for sufficient contrast and visual harmony.
- Test UI components for accessibility (WCAG) compliance after applying the palette.

---
_Argus UI Modernization, 2024-07-03_ 