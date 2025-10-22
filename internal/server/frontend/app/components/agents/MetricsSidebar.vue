<template>
  <aside v-if="show" class="metrics-sidebar" :class="{ 'collapsed': isCollapsed }">
    <!-- Collapse/Expand Toggle Button -->
    <button
      @click="toggleCollapse"
      class="collapse-toggle"
      :title="isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
    >
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        :class="{ 'rotated': isCollapsed }"
      >
        <polyline points="9 18 15 12 9 6"></polyline>
      </svg>
    </button>

    <!-- Sidebar Content -->
    <div class="sidebar-content" v-show="!isCollapsed">
      <SessionMetrics
        :session="session"
        :message-count="messageCount"
        :tool-executions="toolExecutions"
        :permission-stats="permissionStats"
      />
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import SessionMetrics from '~/components/SessionMetrics.vue'

interface Props {
  show: boolean
  session: any
  messageCount: number
  toolExecutions: any
  permissionStats: any
}

defineProps<Props>()

// Sidebar collapse state
const isCollapsed = ref(false)

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}
</script>

<style scoped>
.metrics-sidebar {
  position: relative;
  width: 320px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
  min-height: 0;
  flex-shrink: 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.metrics-sidebar:hover {
  box-shadow: 0 4px 16px rgba(139, 92, 246, 0.1);
}

.metrics-sidebar.collapsed {
  width: 48px;
}

/* Collapse/Expand Toggle Button */
.collapse-toggle {
  position: absolute;
  top: 16px;
  left: 16px;
  z-index: 10;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-secondary);
}

.collapse-toggle:hover {
  background: var(--accent-purple);
  color: white;
  border-color: var(--accent-purple);
  transform: scale(1.05);
}

.collapse-toggle svg {
  transition: transform 0.3s ease;
}

.collapse-toggle svg.rotated {
  transform: rotate(180deg);
}

/* Sidebar Content */
.sidebar-content {
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-width: thin;
  scrollbar-color: var(--accent-purple) transparent;
}

.sidebar-content::-webkit-scrollbar {
  width: 6px;
}

.sidebar-content::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
  transition: background 0.2s;
}

.sidebar-content::-webkit-scrollbar-thumb:hover {
  background: var(--accent-purple);
}

/* Responsive */
@media (max-width: 1200px) {
  .metrics-sidebar {
    width: 280px;
  }

  .metrics-sidebar.collapsed {
    width: 48px;
  }
}

@media (max-width: 768px) {
  .metrics-sidebar {
    display: none;
  }
}

/* Animation for smooth transitions */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.metrics-sidebar {
  animation: slideIn 0.3s ease-out;
}
</style>
