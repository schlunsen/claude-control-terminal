export type ThemeVariant = 'default-dark' | 'default-light' | 'neon-dark' | 'neon-light' | 'nord-dark' | 'nord-light' | 'dracula-dark' | 'dracula-light' | 'southpark-dark' | 'southpark-light'

export interface Theme {
  id: ThemeVariant
  name: string
  description: string
  isDark: boolean
  fontFamily: string
  fontDescription: string
}

export const availableThemes: Theme[] = [
  {
    id: 'default-dark',
    name: 'Default Dark',
    description: 'Classic dark theme with purple accents',
    isDark: true,
    fontFamily: 'Inter',
    fontDescription: 'Modern, clean sans-serif'
  },
  {
    id: 'default-light',
    name: 'Default Light',
    description: 'Clean light theme',
    isDark: false,
    fontFamily: 'Inter',
    fontDescription: 'Modern, clean sans-serif'
  },
  {
    id: 'neon-dark',
    name: 'Neon Dark',
    description: 'Cyberpunk-inspired neon dark theme',
    isDark: true,
    fontFamily: 'Orbitron',
    fontDescription: 'Futuristic, geometric display font'
  },
  {
    id: 'neon-light',
    name: 'Neon Light',
    description: 'Bright neon accents on light background',
    isDark: false,
    fontFamily: 'Orbitron',
    fontDescription: 'Futuristic, geometric display font'
  },
  {
    id: 'nord-dark',
    name: 'Nord Dark',
    description: 'Arctic-inspired cool blues and muted tones',
    isDark: true,
    fontFamily: 'Fira Code',
    fontDescription: 'Developer-focused monospaced font with ligatures'
  },
  {
    id: 'nord-light',
    name: 'Nord Light',
    description: 'Bright Nordic palette with subtle blues',
    isDark: false,
    fontFamily: 'Fira Code',
    fontDescription: 'Developer-focused monospaced font with ligatures'
  },
  {
    id: 'dracula-dark',
    name: 'Dracula Dark',
    description: 'Vibrant purple and pink tones on dark background',
    isDark: true,
    fontFamily: 'JetBrains Mono',
    fontDescription: 'Professional coding font with excellent clarity'
  },
  {
    id: 'dracula-light',
    name: 'Dracula Light',
    description: 'Soft pastels with Dracula accent colors',
    isDark: false,
    fontFamily: 'JetBrains Mono',
    fontDescription: 'Professional coding font with excellent clarity'
  },
  {
    id: 'southpark-dark',
    name: 'South Park Dark',
    description: 'Respect my authoritah! Bold colors and comic style',
    isDark: true,
    fontFamily: 'Comic Sans MS',
    fontDescription: 'Oh my God! They used Comic Sans!'
  },
  {
    id: 'southpark-light',
    name: 'South Park Light',
    description: 'Screw you guys, I\'m going light mode!',
    isDark: false,
    fontFamily: 'Comic Sans MS',
    fontDescription: 'Oh my God! They used Comic Sans!'
  }
]

export const useTheme = () => {
  const currentTheme = useState<ThemeVariant>('theme', () => {
    // Check localStorage first, default to default-dark
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('cct-theme') as ThemeVariant
      if (stored && availableThemes.some(t => t.id === stored)) {
        return stored
      }
    }
    return 'default-dark'
  })

  const currentThemeData = computed(() => {
    return availableThemes.find(t => t.id === currentTheme.value) || availableThemes[0]
  })

  const isDark = computed(() => currentThemeData.value.isDark)

  const setTheme = (themeId: ThemeVariant) => {
    currentTheme.value = themeId
    updateTheme()
  }

  const toggleDarkMode = () => {
    // Toggle between dark and light variant of current theme family
    const currentId = currentTheme.value

    if (currentId.includes('neon')) {
      currentTheme.value = isDark.value ? 'neon-light' : 'neon-dark'
    } else if (currentId.includes('nord')) {
      currentTheme.value = isDark.value ? 'nord-light' : 'nord-dark'
    } else if (currentId.includes('dracula')) {
      currentTheme.value = isDark.value ? 'dracula-light' : 'dracula-dark'
    } else if (currentId.includes('southpark')) {
      currentTheme.value = isDark.value ? 'southpark-light' : 'southpark-dark'
    } else {
      currentTheme.value = isDark.value ? 'default-light' : 'default-dark'
    }

    updateTheme()
  }

  const updateTheme = () => {
    if (typeof window !== 'undefined') {
      const html = document.documentElement
      html.setAttribute('data-theme', currentTheme.value)
      localStorage.setItem('cct-theme', currentTheme.value)
    }
  }

  // Initialize theme on mount
  if (typeof window !== 'undefined') {
    onMounted(() => {
      updateTheme()
    })
  }

  return {
    currentTheme: readonly(currentTheme),
    currentThemeData: readonly(currentThemeData),
    isDark: readonly(isDark),
    availableThemes,
    setTheme,
    toggleDarkMode
  }
}
