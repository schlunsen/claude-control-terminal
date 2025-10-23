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
      <!-- Session Header at Top -->
      <div v-if="session" class="session-header-section">
        <div class="metrics-header">
          <div class="header-row">
            <div class="session-badge">
              <span class="badge-label">Session</span>
              <span class="badge-value">{{ session.id.slice(0, 8) }}</span>
            </div>
            <div class="duration" v-if="sessionDuration">
              ⏱️ {{ sessionDuration }}
            </div>
          </div>
          <div class="header-row">
            <div class="status-badge" :class="session.status">
              <span class="status-dot"></span>
              {{ session.status }}
            </div>
          </div>
        </div>
      </div>

      <!-- Provider Section -->
      <div v-if="session" class="provider-section">
        <span class="provider-label">PROVIDER</span>
        <div class="provider-badge">
          <span>{{ getProviderDisplay(session.options?.provider) }}</span>
        </div>
      </div>

      <!-- Context Usage Bar -->
      <div class="context-usage-section">
        <ContextUsageBar
          :usage="contextUsage"
          :loading="contextLoading"
          @refresh="$emit('refresh-context')"
        />
      </div>

      <SessionMetrics
        :session="session"
        :message-count="messageCount"
        :tool-executions="toolExecutions"
        :permission-stats="permissionStats"
        :hide-header="true"
      />
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import SessionMetrics from '~/components/SessionMetrics.vue'
import ContextUsageBar from '~/components/agents/ContextUsageBar.vue'

interface Props {
  show: boolean
  session: any
  messageCount: number
  toolExecutions: any
  permissionStats: any
  contextUsage?: any
  contextLoading?: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'refresh-context'): void
}>()

// Sidebar collapse state
const isCollapsed = ref(false)

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}

// Provider display function
const getProviderDisplay = (provider?: string): string => {
  const actualProvider = provider || 'anthropic'

  const providerMap: Record<string, string> = {
    'anthropic': '🟣 Anthropic',
    'glm': '🤖 GLM',
    'deepseek': '🔍 DeepSeek',
    'openai': '🟢 OpenAI',
    'google': '🔵 Google',
    'azure': '☁️ Azure',
    'cohere': '🟠 Cohere',
    'custom': '⚙️ Custom'
  }

  return providerMap[actualProvider.toLowerCase()] || `🔧 ${actualProvider}`
}

// Session duration tracking
const sessionStartTime = ref<Date | null>(null)
const sessionDuration = ref('')
let durationInterval: NodeJS.Timeout | null = null

const formatDuration = (startTime: Date): string => {
  const now = new Date()
  const diff = now.getTime() - startTime.getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`
  } else {
    return `${seconds}s`
  }
}

// Watch for session changes
watch(
  () => props.session?.created_at,
  (newVal) => {
    if (newVal) {
      sessionStartTime.value = new Date(newVal)
    }
  },
  { immediate: true }
)

// Update duration every second
onMounted(() => {
  durationInterval = setInterval(() => {
    if (sessionStartTime.value && props.session?.status !== 'ended') {
      sessionDuration.value = formatDuration(sessionStartTime.value)
    }
  }, 1000)
})

onUnmounted(() => {
  if (durationInterval) {
    clearInterval(durationInterval)
  }
})

// Initial duration calculation
watch(sessionStartTime, (newVal) => {
  if (newVal && props.session?.status !== 'ended') {
    sessionDuration.value = formatDuration(newVal)
  }
})

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
  top: 12px;
  right: 12px;
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

/* Session Header Section */
.session-header-section {
  padding: 16px 12px 12px 12px;
  border-bottom: 1px solid var(--border-color);
}

.metrics-header {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.session-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  flex: 1;
}

.badge-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.badge-value {
  font-size: 0.9rem;
  color: var(--accent-purple);
  font-weight: 700;
  font-family: 'Monaco', 'Menlo', monospace;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  text-transform: capitalize;
  border: 1px solid var(--border-color);
  flex: 1;
}

.status-badge.idle {
  background: rgba(40, 167, 69, 0.1);
  color: #28a745;
  border-color: rgba(40, 167, 69, 0.3);
}

.status-badge.processing {
  background: rgba(23, 162, 184, 0.1);
  color: #17a2b8;
  border-color: rgba(23, 162, 184, 0.3);
  animation: pulse 2s infinite;
}

.status-badge.error {
  background: rgba(220, 53, 69, 0.1);
  color: #dc3545;
  border-color: rgba(220, 53, 69, 0.3);
}

.status-badge.ended {
  background: rgba(108, 117, 125, 0.1);
  color: #6c757d;
  border-color: rgba(108, 117, 125, 0.3);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  animation: statusPulse 2s infinite;
}

.duration {
  padding: 8px 14px;
  background: var(--bg-secondary);
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
  white-space: nowrap;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

@keyframes statusPulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* Provider Section */
.provider-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  border-bottom: 1px solid var(--border-color);
}

.provider-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.5px;
  padding: 0 4px;
}

.provider-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  color: white;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
  width: 100%;
}

/* Context Usage Section */
.context-usage-section {
  padding: 12px;
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
