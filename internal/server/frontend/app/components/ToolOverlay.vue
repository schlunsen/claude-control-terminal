<template>
  <transition name="slide-fade">
    <div v-if="visible" class="tool-overlay" :class="{ 'tool-completed': tool.status === 'completed' }">
      <div class="tool-header">
        <div class="tool-icon" :class="`tool-${tool.status}`">
          <span v-if="tool.status === 'running'">⚙️</span>
          <span v-else-if="tool.status === 'completed'">✅</span>
          <span v-else>❌</span>
        </div>
        <div class="tool-info">
          <div class="tool-name">{{ tool.name }}</div>
          <div class="tool-status">{{ statusText }}</div>
        </div>
      </div>

      <div v-if="inputSummary" class="tool-input">
        {{ inputSummary }}
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import type { ActiveTool } from '~/types/agents'

const props = defineProps<{
  tool: ActiveTool
  autoDismissDelay?: number
}>()

const emit = defineEmits<{
  dismiss: [toolId: string]
}>()

const visible = ref(true)
const autoDismissDelay = props.autoDismissDelay || 5000

const statusText = computed(() => {
  switch (props.tool.status) {
    case 'running':
      return 'Running...'
    case 'completed':
      return 'Completed'
    case 'error':
      return 'Error'
    default:
      return ''
  }
})

const inputSummary = computed(() => {
  if (!props.tool.input) return ''

  // Special handling for different tools
  switch (props.tool.name) {
    case 'Read':
      return props.tool.input.file_path || ''

    case 'Write':
      return props.tool.input.file_path || ''

    case 'Edit':
      return props.tool.input.file_path || ''

    case 'Bash':
      return props.tool.input.command?.substring(0, 50) + (props.tool.input.command?.length > 50 ? '...' : '') || ''

    default:
      return Object.keys(props.tool.input).join(', ')
  }

  return ''
})

// Watch for completion and auto-dismiss
watch(() => props.tool.status, (newStatus) => {
  if (newStatus === 'completed' || newStatus === 'error') {
    setTimeout(() => {
      visible.value = false
      setTimeout(() => {
        emit('dismiss', props.tool.id)
      }, 300) // Wait for animation
    }, autoDismissDelay)
  }
})
</script>

<style scoped>
.tool-overlay {
  position: relative;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 8px;
  min-width: 350px;
  max-width: 450px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.tool-overlay:hover {
  border-color: var(--accent-purple);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.tool-completed {
  opacity: 0.85;
}

.tool-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.tool-icon {
  font-size: 20px;
  line-height: 1;
}

.tool-icon.tool-running {
  animation: spin 2s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.tool-info {
  flex: 1;
  min-width: 0;
}

.tool-name {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-primary);
  margin-bottom: 2px;
}

.tool-status {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.tool-input {
  font-size: 0.8rem;
  color: var(--text-muted);
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 6px 8px;
  background: var(--bg-secondary);
  border-radius: 4px;
  margin-top: 8px;
}

/* Transition animations */
.slide-fade-enter-active {
  transition: all 0.3s ease-out;
}

.slide-fade-leave-active {
  transition: all 0.3s cubic-bezier(1, 0.5, 0.8, 1);
}

.slide-fade-enter-from {
  transform: translateX(20px);
  opacity: 0;
}

.slide-fade-leave-to {
  transform: translateX(20px);
  opacity: 0;
}
</style>
