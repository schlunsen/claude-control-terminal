<script setup>
import '../assets/css/main.css'

// Initialize dark mode
const { isDark } = useDarkMode()
</script>

<template>
  <div id="app">
    <nav class="navbar">
      <div class="navbar-container">
        <div class="nav-brand">
          <span class="nav-logo">CCT</span>
          <span class="nav-separator">|</span>
          <span class="nav-title">Analytics Dashboard</span>
        </div>
        <ThemeToggle />
      </div>
    </nav>

    <NuxtRouteAnnouncer />
    <NuxtPage />
  </div>
</template>

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
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
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

@media (max-width: 768px) {
  .navbar {
    padding: 0.75rem 0;
  }

  .navbar-container {
    padding: 0 1rem;
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
    padding: 0 0.75rem;
  }
}
</style>
