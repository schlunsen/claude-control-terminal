<template>
  <div id="app">
    <nav class="navbar">
      <div class="navbar-container">
        <div class="nav-left">
          <button 
            @click="toggleSidebar" 
            class="sidebar-toggle"
            :title="isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="3" y1="6" x2="21" y2="6"/>
              <line x1="3" y1="12" x2="21" y2="12"/>
              <line x1="3" y1="18" x2="21" y2="18"/>
            </svg>
          </button>
          <span class="nav-logo">CCT</span>
        </div>
        <div class="nav-right">
          <button
            @click="openShortcutsDialog"
            class="shortcuts-button"
            title="Keyboard shortcuts (⇧⌥⌘H)"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="4" width="20" height="16" rx="2"/>
              <path d="M6 8h.01M10 8h.01M14 8h.01M18 8h.01M8 12h.01M12 12h.01M16 12h.01M7 16h10"/>
            </svg>
          </button>
          <div class="nav-brand">
            <span v-if="versionInfo.version" class="version-badge">v{{ versionInfo.version }}</span>
            <span class="nav-separator">|</span>
            <span class="nav-title">Analytics Dashboard</span>
          </div>
          <ThemeSelector :show-label="false" />
          <ThemeToggle />

          <!-- User menu (only show if authenticated) -->
          <div v-if="isAuthenticated" class="user-menu">
            <button
              @click="toggleUserMenu"
              class="user-button"
              :title="`Logged in as ${user?.username}`"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                <circle cx="12" cy="7" r="4"/>
              </svg>
              <span class="user-name">{{ user?.username }}</span>
            </button>

            <!-- Dropdown menu -->
            <div v-if="showUserMenu" class="user-dropdown">
              <div class="user-dropdown-header">
                <div class="user-info">
                  <span class="user-info-name">{{ user?.username }}</span>
                  <span v-if="user?.isAdmin" class="user-info-badge">Admin</span>
                </div>
              </div>
              <div class="user-dropdown-divider"></div>
              <button @click="handleLogout" class="user-dropdown-item">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
                  <polyline points="16 17 21 12 16 7"/>
                  <line x1="21" y1="12" x2="9" y2="12"/>
                </svg>
                Logout
              </button>
            </div>
          </div>
        </div>
      </div>
    </nav>

    <div class="app-layout">
      <Sidebar />
      <main class="main-content" :class="{ 'main-content-expanded': isCollapsed }">
        <slot />
      </main>
    </div>

    <!-- Shortcuts Dialog -->
    <ShortcutsDialog />
  </div>
</template>

<script setup>
import '../assets/css/main.css'

// Initialize theme system
const { isDark } = useTheme()

// Initialize authentication
const { isAuthenticated, user, logout, checkAuthStatus } = useAuth()
const showUserMenu = ref(false)
const router = useRouter()

// Toggle user menu
const toggleUserMenu = () => {
  showUserMenu.value = !showUserMenu.value
}

// Close menu when clicking outside
const closeUserMenu = () => {
  showUserMenu.value = false
}

// Handle logout
const handleLogout = async () => {
  try {
    await logout()
    router.push('/login')
  } catch (error) {
    console.error('Logout failed:', error)
  } finally {
    showUserMenu.value = false
  }
}

// Initialize sidebar state
const { isCollapsed, toggleSidebar } = useSidebar()

// Initialize keyboard shortcuts
const {
  openDialog: openShortcutsDialog,
  initializeShortcuts,
  cleanupShortcuts,
  registerDefaultShortcuts
} = useKeyboardShortcuts()

// Version info
const versionInfo = ref({
  version: '',
  name: ''
})

// Load version info
async function loadVersion() {
  try {
    const { data } = await useFetch('/api/version')
    if (data.value) {
      versionInfo.value = data.value
    }
  } catch (error) {
    // Error loading version
  }
}

// Open analytics in new tab
function openAnalytics() {
  window.open('http://localhost:3333', '_blank')
}

// Load version on mount
onMounted(async () => {
  loadVersion()

  // Check authentication status
  await checkAuthStatus()

  // Initialize keyboard shortcuts
  registerDefaultShortcuts()

  // Close user menu on click outside
  document.addEventListener('click', (e) => {
    const userMenu = document.querySelector('.user-menu')
    if (userMenu && !userMenu.contains(e.target)) {
      showUserMenu.value = false
    }
  })

  // Register new session shortcut (Shift+Option+Cmd+N)
  const { registerShortcut } = useKeyboardShortcuts()
  const router = useRouter()
  registerShortcut('n', 'Create New Session', 'Agents', () => {
    // Navigate to agents page and trigger create session
    if (router.currentRoute.value.path !== '/agents') {
      router.push('/agents').then(() => {
        // Use a small delay to ensure the agents page component has fully mounted
        setTimeout(() => {
          const { triggerGlobalAction } = useKeyboardShortcuts()
          triggerGlobalAction('create-new-session')
        }, 100)
      })
    } else {
      // Already on agents page, trigger immediately
      nextTick(() => {
        const { triggerGlobalAction } = useKeyboardShortcuts()
        triggerGlobalAction('create-new-session')
      })
    }
  })

  initializeShortcuts()
})

// Cleanup on unmount
onUnmounted(() => {
  cleanupShortcuts()
})
</script>

<style scoped>
#app {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  overflow: hidden;
}

.navbar {
  flex-shrink: 0;
  z-index: 1000;
  background: var(--bg-secondary);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--border-color);
  padding: 1rem 0;
  transition: background-color 0.3s ease;
}

.navbar-container {
  width: 100%;
  padding: 0 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.nav-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.sidebar-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: none;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid var(--border-color);
}

.sidebar-toggle:hover {
  background: var(--accent-purple);
  color: white;
  border-color: var(--accent-purple);
}

.shortcuts-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: none;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid var(--border-color);
}

.shortcuts-button:hover {
  background: var(--accent-purple);
  color: white;
  border-color: var(--accent-purple);
}


.nav-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 600;
}

.nav-logo {
  color: var(--accent-purple);
  font-size: 1.2rem;
  font-weight: 700;
}

.nav-separator {
  color: var(--text-muted);
}

.nav-title {
  color: var(--text-primary);
  font-size: 1rem;
}

.version-badge {
  background: var(--accent-purple);
  color: white;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.02em;
}

/* User Menu */
.user-menu {
  position: relative;
}

.user-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.user-button:hover {
  background: var(--accent-purple);
  color: white;
  border-color: var(--accent-purple);
}

.user-name {
  font-size: 0.875rem;
  font-weight: 500;
}

.user-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
  min-width: 200px;
  z-index: 1000;
  animation: slideDown 0.2s ease;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.user-dropdown-header {
  padding: 12px 16px;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-info-name {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
}

.user-info-badge {
  display: inline-block;
  background: var(--accent-purple);
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.02em;
  width: fit-content;
}

.user-dropdown-divider {
  height: 1px;
  background: var(--border-color);
  margin: 0;
}

.user-dropdown-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
}

.user-dropdown-item:hover {
  background: var(--bg-secondary);
}

.user-dropdown-item:last-child {
  border-bottom-left-radius: 8px;
  border-bottom-right-radius: 8px;
}

.app-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.main-content {
  flex: 1;
  overflow: hidden;
  transition: all 0.3s ease;
  min-height: 0;
}

.main-content-expanded {
  margin-left: 0;
}

@media (max-width: 768px) {
  .navbar {
    padding: 0.75rem 0;
  }

  .navbar-container {
    padding: 0 15px;
  }

  .nav-left {
    gap: 0.75rem;
  }

  .sidebar-toggle {
    width: 36px;
    height: 36px;
  }

  .shortcuts-button {
    width: 36px;
    height: 36px;
  }

  .nav-brand {
    gap: 0.5rem;
  }

  .nav-logo {
    font-size: 1.1rem;
  }

  .nav-title {
    display: none;
  }

  .user-name {
    display: none;
  }

  .user-button {
    padding: 8px;
  }
}

@media (max-width: 480px) {
  .navbar-container {
    padding: 0 15px;
  }
}
</style>