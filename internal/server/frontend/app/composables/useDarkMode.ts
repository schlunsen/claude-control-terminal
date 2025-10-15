export const useDarkMode = () => {
  const isDark = useState<boolean>('darkMode', () => {
    // Check localStorage first, default to dark mode
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('theme')
      if (stored) {
        return stored === 'dark'
      }
    }
    return true // Default to dark mode
  })

  const toggleDarkMode = () => {
    isDark.value = !isDark.value
    updateTheme()
  }

  const setDarkMode = (value: boolean) => {
    isDark.value = value
    updateTheme()
  }

  const updateTheme = () => {
    if (typeof window !== 'undefined') {
      const html = document.documentElement
      if (isDark.value) {
        html.removeAttribute('data-theme')
        localStorage.setItem('theme', 'dark')
      } else {
        html.setAttribute('data-theme', 'light')
        localStorage.setItem('theme', 'light')
      }
    }
  }

  // Initialize theme on mount
  if (typeof window !== 'undefined') {
    onMounted(() => {
      updateTheme()
    })
  }

  return {
    isDark: readonly(isDark),
    toggleDarkMode,
    setDarkMode
  }
}
