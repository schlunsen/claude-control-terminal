<template>
  <nav class="sidebar" :class="{ 'sidebar-collapsed': isCollapsed }">
    <div class="sidebar-content">
      <div class="nav-section">
        <NuxtLink 
          to="/" 
          class="nav-item"
          :class="{ 'nav-item-active': $route.path === '/' }"
        >
          <div class="nav-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
              <polyline points="9,22 9,12 15,12 15,22"/>
            </svg>
          </div>
          <span class="nav-label" v-show="!isCollapsed">Frontpage</span>
        </NuxtLink>

        <NuxtLink
          to="/stats"
          class="nav-item"
          :class="{ 'nav-item-active': $route.path === '/stats' }"
        >
          <div class="nav-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="20" x2="18" y2="10"/>
              <line x1="12" y1="20" x2="12" y2="4"/>
              <line x1="6" y1="20" x2="6" y2="14"/>
            </svg>
          </div>
          <span class="nav-label" v-show="!isCollapsed">Stats</span>
        </NuxtLink>

        <NuxtLink
          to="/agents"
          class="nav-item"
          :class="{ 'nav-item-active': $route.path === '/agents' }"
        >
          <div class="nav-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
            </svg>
          </div>
          <span class="nav-label" v-show="!isCollapsed">Live Agents</span>
        </NuxtLink>
      </div>
    </div>
  </nav>
</template>

<script setup>
// Use sidebar state
const { isCollapsed } = useSidebar()
</script>

<style scoped>
.sidebar {
  width: 250px;
  background: var(--card-bg);
  border-right: 1px solid var(--border-color);
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.sidebar-collapsed {
  width: 70px;
}

.sidebar-content {
  padding: 24px 16px;
  height: 100%;
}

.nav-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.2s ease;
  font-weight: 500;
  font-size: 0.9rem;
}

.sidebar-collapsed .nav-item {
  justify-content: center;
  padding: 12px 8px;
}

.nav-item:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.nav-item-active {
  background: var(--accent-purple);
  color: white;
}

.nav-item-active:hover {
  background: var(--accent-purple);
  color: white;
}

.nav-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
}

.nav-label {
  font-weight: 500;
}

@media (max-width: 768px) {
  .sidebar {
    width: 200px;
  }
  
  .sidebar-content {
    padding: 20px 12px;
  }
  
  .nav-item {
    padding: 10px 12px;
    font-size: 0.85rem;
  }
}

@media (max-width: 640px) {
  .sidebar {
    width: 60px;
  }
  
  .sidebar-collapsed {
    width: 60px;
  }
  
  .nav-label {
    display: none;
  }
  
  .nav-item {
    justify-content: center;
    padding: 12px 8px;
  }
}
</style>