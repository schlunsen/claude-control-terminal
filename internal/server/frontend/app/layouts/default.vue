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
          <div class="nav-brand">
            <span v-if="versionInfo.version" class="version-badge">v{{ versionInfo.version }}</span>
            <span class="nav-separator">|</span>
            <span class="nav-title">Analytics Dashboard</span>
          </div>
          <ThemeToggle />
        </div>
      </div>
    </nav>

    <div class="app-layout">
      <Sidebar />
      <main class="main-content" :class="{ 'main-content-expanded': isCollapsed }">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup>
import '../assets/css/main.css'

// Initialize dark mode
const { isDark } = useDarkMode()

// Initialize sidebar state
const { isCollapsed, toggleSidebar } = useSidebar()

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

// Load version on mount
onMounted(() => {
  loadVersion()
})
</script>

<style scoped>
#app {
  min-height: 100vh;
  background: var(--bg-primary);
}

.navbar {
  position: sticky;
  top: 0;
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

.app-layout {
  display: flex;
  min-height: calc(100vh - 80px);
}

.main-content {
  flex: 1;
  overflow: auto;
  transition: all 0.3s ease;
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

  .nav-brand {
    gap: 0.5rem;
  }

  .nav-logo {
    font-size: 1.1rem;
  }

  .nav-title {
    display: none;
  }
}

@media (max-width: 480px) {
  .navbar-container {
    padding: 0 15px;
  }
}
</style>