export interface KeyboardShortcut {
  key: string
  description: string
  category: string
  action: () => void
  modifiers: {
    shift: boolean
    alt: boolean
    meta: boolean
    ctrl: boolean
  }
}

// Global state for shortcuts dialog
const showDialog = ref(false)
const shortcuts = ref<Map<string, KeyboardShortcut>>(new Map())
// Global state for page-specific actions
const globalShortcutActions = ref<Map<string, () => void>>(new Map())

export const useKeyboardShortcuts = () => {
  const router = useRouter()

  /**
   * Register a keyboard shortcut
   */
  const registerShortcut = (
    key: string,
    description: string,
    category: string,
    action: () => void,
    modifiers = { shift: true, alt: true, meta: true, ctrl: false }
  ) => {
    const shortcutKey = `${modifiers.shift ? 'shift+' : ''}${modifiers.alt ? 'alt+' : ''}${modifiers.meta ? 'meta+' : ''}${modifiers.ctrl ? 'ctrl+' : ''}${key.toLowerCase()}`

    shortcuts.value.set(shortcutKey, {
      key,
      description,
      category,
      action,
      modifiers
    })
  }

  /**
   * Unregister a keyboard shortcut
   */
  const unregisterShortcut = (key: string, modifiers = { shift: true, alt: true, meta: true, ctrl: false }) => {
    const shortcutKey = `${modifiers.shift ? 'shift+' : ''}${modifiers.alt ? 'alt+' : ''}${modifiers.meta ? 'meta+' : ''}${modifiers.ctrl ? 'ctrl+' : ''}${key.toLowerCase()}`
    shortcuts.value.delete(shortcutKey)
  }

  /**
   * Get all registered shortcuts grouped by category
   */
  const getAllShortcuts = (): Record<string, KeyboardShortcut[]> => {
    const grouped: Record<string, KeyboardShortcut[]> = {}

    shortcuts.value.forEach((shortcut) => {
      if (!grouped[shortcut.category]) {
        grouped[shortcut.category] = []
      }
      grouped[shortcut.category].push(shortcut)
    })

    return grouped
  }

  /**
   * Toggle shortcuts dialog
   */
  const toggleDialog = () => {
    showDialog.value = !showDialog.value
  }

  /**
   * Close shortcuts dialog
   */
  const closeDialog = () => {
    showDialog.value = false
  }

  /**
   * Open shortcuts dialog
   */
  const openDialog = () => {
    showDialog.value = true
  }

  /**
   * Handle keyboard events
   */
  const handleKeyDown = (event: KeyboardEvent) => {
    // Close dialog on ESC
    if (event.key === 'Escape' && showDialog.value) {
      closeDialog()
      return
    }

    // Don't trigger shortcuts when typing in input fields
    const target = event.target as HTMLElement
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
      return
    }

    // Ignore modifier keys themselves (Shift, Control, Alt, Meta)
    const modifierKeys = ['Shift', 'Control', 'Alt', 'Meta', 'Command', 'Option']
    if (modifierKeys.includes(event.key)) {
      return
    }

    // Use event.code to get the physical key, not the character produced
    // This handles macOS special characters (e.g., Shift+Option+L = "ﬂ")
    let key = event.key.toLowerCase()

    // For special characters and symbols, prefer event.key
    // For letters, use event.code to avoid macOS special character mappings
    if (event.code && event.code.startsWith('Key')) {
      // Letter keys: use event.code to avoid special characters
      key = event.code.replace('Key', '').toLowerCase()
    } else if (event.code === 'Slash' || event.key === '/') {
      // Slash key - handle both code and key
      key = '/'
    } else {
      // For everything else (numbers, symbols, etc.), use event.key
      key = event.key.toLowerCase()
    }

    const modifiers = {
      shift: event.shiftKey,
      alt: event.altKey,
      meta: event.metaKey,
      ctrl: event.ctrlKey
    }

    const shortcutKey = `${modifiers.shift ? 'shift+' : ''}${modifiers.alt ? 'alt+' : ''}${modifiers.meta ? 'meta+' : ''}${modifiers.ctrl ? 'ctrl+' : ''}${key}`

    // Execute the shortcut if it exists
    const shortcut = shortcuts.value.get(shortcutKey)
    if (shortcut) {
      event.preventDefault()
      shortcut.action()
    }
  }

  /**
   * Initialize keyboard shortcuts
   */
  const initializeShortcuts = () => {
    // Add event listener
    if (typeof window !== 'undefined') {
      window.addEventListener('keydown', handleKeyDown)
    }
  }

  /**
   * Cleanup keyboard shortcuts
   */
  const cleanupShortcuts = () => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }

  /**
   * Register default shortcuts
   */
  const registerDefaultShortcuts = () => {
    // Navigation shortcuts
    registerShortcut('f', 'Navigate to Frontpage', 'Navigation', () => {
      router.push('/')
    })

    registerShortcut('s', 'Navigate to Stats', 'Navigation', () => {
      router.push('/stats')
    })

    registerShortcut('l', 'Navigate to Live Agents', 'Navigation', () => {
      // If already on agents page, trigger create session modal
      if (router.currentRoute.value.path === '/agents') {
        // Use small delay in case page is still mounting
        setTimeout(() => {
          triggerGlobalAction('create-new-session')
        }, 50)
      } else {
        router.push('/agents').then(() => {
          // Use a small delay to ensure the agents page component has fully mounted
          // and registered its global action handler
          setTimeout(() => {
            triggerGlobalAction('create-new-session')
          }, 100)
        })
      }
    })

    // UI Controls
    const { toggleSidebar } = useSidebar()
    registerShortcut('b', 'Toggle sidebar', 'UI Controls', () => {
      toggleSidebar()
    })

    const { toggleDarkMode } = useDarkMode()
    registerShortcut('t', 'Toggle theme', 'UI Controls', () => {
      toggleDarkMode()
    })

    // Show shortcuts dialog
    registerShortcut('h', 'Show shortcuts', 'Help', () => {
      openDialog()
    })
  }

  /**
   * Register a global shortcut action (for page-specific actions)
   */
  const setGlobalAction = (action: string, handler: () => void) => {
    globalShortcutActions.value.set(action, handler)
  }

  const removeGlobalAction = (action: string) => {
    globalShortcutActions.value.delete(action)
  }

  const triggerGlobalAction = (action: string) => {
    const handler = globalShortcutActions.value.get(action)
    if (handler) {
      handler()
    }
  }

  return {
    showDialog: readonly(showDialog),
    registerShortcut,
    unregisterShortcut,
    getAllShortcuts,
    toggleDialog,
    closeDialog,
    openDialog,
    initializeShortcuts,
    cleanupShortcuts,
    registerDefaultShortcuts,
    setGlobalAction,
    removeGlobalAction,
    triggerGlobalAction
  }
}
