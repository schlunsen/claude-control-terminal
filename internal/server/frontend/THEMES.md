# Theme System Documentation

## Overview

The Claude Control Terminal analytics dashboard now features a comprehensive theming system with 6 built-in themes across different styles and color modes.

## Available Themes

### Default Themes
- **Default Dark** - Classic dark theme with purple accents (default)
  - Font: **Inter** - Modern, clean sans-serif for excellent readability
- **Default Light** - Clean light theme with subtle colors
  - Font: **Inter** - Modern, clean sans-serif for excellent readability

### Neon Themes (Cyberpunk-inspired)
- **Neon Dark** - Deep purple/black background with vibrant cyan, magenta, and neon green accents
  - Font: **Orbitron** - Futuristic, geometric display font for that cyberpunk vibe
- **Neon Light** - Bright white/purple background with bold neon accent colors
  - Font: **Orbitron** - Futuristic, geometric display font for that cyberpunk vibe

### Nord Themes (Arctic-inspired)
- **Nord Dark** - Cool blues and muted tones inspired by the Arctic wilderness
  - Font: **Fira Code** - Developer-focused monospaced font with programming ligatures
- **Nord Light** - Bright Nordic palette with subtle blues and elegant contrast
  - Font: **Fira Code** - Developer-focused monospaced font with programming ligatures

### Dracula Themes (Vibrant & Popular)
- **Dracula Dark** - Vibrant purple and pink tones on dark background
  - Font: **JetBrains Mono** - Professional coding font with excellent clarity
- **Dracula Light** - Soft pastels with Dracula's iconic accent colors
  - Font: **JetBrains Mono** - Professional coding font with excellent clarity

## How to Use

### Theme Selection

**Via Themes Page:**
1. Click "Themes" in the sidebar navigation
2. Browse available themes in the horizontal carousel
3. Use arrow buttons to navigate through themes if they don't all fit on screen
4. Click on any theme card to activate it
5. View the color palette for the current theme

**Via Dropdown Selector:**
1. Click the theme selector icon in the navbar (top right)
2. Select any theme from the dropdown menu
3. Click "Manage Themes" to go to the full themes page

**Via Toggle Button:**
1. Use the sun/moon toggle button in the navbar
2. Toggles between dark/light variants of the current theme family
3. Preserves theme family (e.g., Neon Dark ↔ Neon Light)

### Persistent Storage

Theme preferences are automatically saved to `localStorage` under the key `cct-theme` and persist across sessions.

## Technical Implementation

### File Structure

```
internal/server/frontend/
├── app/
│   ├── composables/
│   │   ├── useTheme.ts          # Theme state management composable
│   │   └── useDarkMode.ts       # Legacy (replaced by useTheme)
│   ├── components/
│   │   ├── ThemeSelector.vue    # Dropdown theme selector
│   │   └── ThemeToggle.vue      # Quick dark/light toggle
│   └── pages/
│       └── themes.vue           # Full themes management page
├── assets/
│   └── css/
│       └── main.css             # Theme CSS variables
└── THEMES.md                    # This file
```

### useTheme Composable

The `useTheme` composable provides:

```typescript
{
  currentTheme: ThemeVariant          // Current theme ID
  currentThemeData: Theme             // Current theme metadata
  isDark: boolean                     // Is current theme dark?
  availableThemes: Theme[]            // All available themes
  setTheme: (id: ThemeVariant) => void    // Set specific theme
  toggleDarkMode: () => void          // Toggle dark/light variant
}
```

### Theme Type Definition

```typescript
type ThemeVariant =
  | 'default-dark'
  | 'default-light'
  | 'neon-dark'
  | 'neon-light'
  | '8bit-dark'
  | '8bit-light'

interface Theme {
  id: ThemeVariant
  name: string
  description: string
  isDark: boolean
  fontFamily: string         // Primary font name
  fontDescription: string    // Font description
}
```

### CSS Variables

Each theme defines these CSS custom properties:

```css
/* Colors */
--bg-primary      /* Main background */
--bg-secondary    /* Secondary background */
--bg-tertiary     /* Card backgrounds */
--text-primary    /* Primary text color */
--text-secondary  /* Secondary text color */
--text-muted      /* Muted/disabled text */
--accent-purple   /* Purple accent color */
--accent-cyan     /* Cyan accent color */
--accent-green    /* Green accent color */
--accent-yellow   /* Yellow accent color */
--accent-orange   /* Orange accent color */
--border-color    /* Border color */
--code-bg         /* Code block background */
--card-bg         /* Card background */
--card-hover      /* Card hover state */
--status-success  /* Success status */
--status-warning  /* Warning status */
--status-error    /* Error status */

/* Typography */
--font-primary    /* Primary font family */
--font-mono       /* Monospace font family */
--font-size-scale /* Optional font size multiplier (8-bit themes) */
```

### Theme-Specific Fonts

Each theme uses custom fonts from Google Fonts to enhance its unique character:

- **Inter** (Default themes) - Clean, modern sans-serif with excellent readability
- **Orbitron** (Neon themes) - Futuristic, geometric display font perfect for cyberpunk aesthetics
- **Fira Code** (Nord themes) - Developer-focused monospaced font with programming ligatures for better code readability
- **Fira Sans** (Nord themes, fallback) - Companion sans-serif to Fira Code for UI elements
- **JetBrains Mono** (Dracula themes) - Professional coding font with excellent clarity and readability

### Adding New Themes

To add a new theme:

1. **Add Google Font (if using a new font)** to `nuxt.config.ts`:
```typescript
{
  rel: 'stylesheet',
  href: 'https://fonts.googleapis.com/css2?family=Your+Font:wght@400;500;600;700&display=swap'
}
```

2. **Define theme in useTheme.ts:**
```typescript
{
  id: 'my-new-theme-dark',
  name: 'My New Theme',
  description: 'A cool new theme',
  isDark: true,
  fontFamily: 'Your Font',
  fontDescription: 'Description of the font'
}
```

3. **Add CSS variables in main.css:**
```css
[data-theme="my-new-theme-dark"] {
  /* Colors */
  --bg-primary: #your-color;
  --bg-secondary: #your-color;
  /* ... define all color variables ... */

  /* Typography */
  --font-primary: 'Your Font', sans-serif;
  --font-mono: 'Your Mono Font', monospace;
  --font-size-scale: 1; /* Optional, defaults to 1 */
}
```

4. **Theme is now available** - Appears in selector and themes page automatically with font info!

## Components

### ThemeSelector Component

Dropdown selector with theme preview and quick switching.

**Props:**
- `showLabel?: boolean` - Show theme name next to icon (default: false)

**Usage:**
```vue
<ThemeSelector :show-label="true" />
```

### ThemeToggle Component

Simple sun/moon icon that toggles between dark/light variants.

**Usage:**
```vue
<ThemeToggle />
```

### Themes Page

Full-featured theme management page with:
- Current theme preview and info
- Grid of all available themes with previews
- Color palette display for current theme
- Click any theme card to activate

## Browser Compatibility

Themes use CSS custom properties (CSS variables) which are supported in:
- Chrome 49+
- Firefox 31+
- Safari 9.1+
- Edge 15+

## Future Enhancements

Potential future additions:
- Custom theme creator/editor
- Theme import/export
- Per-page theme overrides
- High contrast themes for accessibility
- Seasonal/special event themes
- User-uploaded theme packs
