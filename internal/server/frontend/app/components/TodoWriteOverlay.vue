<template>
  <transition name="slide-fade">
    <div v-if="visible" class="todowrite-overlay" :class="{ 'tool-completed': tool.status === 'completed' }">
      <div class="tool-header">
        <div class="tool-icon" :class="`tool-${tool.status}`">
          <span v-if="tool.status === 'running'">📋</span>
          <span v-else-if="tool.status === 'completed'">✅</span>
          <span v-else>❌</span>
        </div>
        <div class="tool-info">
          <div class="tool-name">TodoWrite</div>
          <div class="tool-status">{{ statusText }}</div>
        </div>
      </div>

      <div v-if="todoItems.length > 0" class="todo-items">
        <div v-for="(item, index) in todoItems" :key="index" class="todo-item" :class="item.status">
          <div class="todo-item-icon">
            <span v-if="item.status === 'completed'">✅</span>
            <span v-else-if="item.status === 'in_progress'">🔄</span>
            <span v-else>📝</span>
          </div>
          <div class="todo-item-content">{{ item.content }}</div>
        </div>
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
      return 'Updating tasks...'
    case 'completed':
      return 'Tasks updated'
    case 'error':
      return 'Error'
    default:
      return ''
  }
})

const todoItems = computed(() => {
  if (props.tool.input?.todos && Array.isArray(props.tool.input.todos)) {
    return props.tool.input.todos
  }
  return []
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
.todowrite-overlay {
  position: relative;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 8px;
  min-width: 400px;
  max-width: 500px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.todowrite-overlay:hover {
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
  margin-bottom: 12px;
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

/* Todo items list */
.todo-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.todo-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 8px 10px;
  background: var(--bg-secondary);
  border-radius: 6px;
  border-left: 3px solid transparent;
  transition: all 0.2s ease;
}

.todo-item.pending {
  border-left-color: var(--text-muted);
}

.todo-item.in_progress {
  border-left-color: var(--accent-orange);
  background: rgba(255, 165, 0, 0.05);
}

.todo-item.completed {
  border-left-color: var(--status-success);
  background: rgba(0, 255, 0, 0.03);
  opacity: 0.7;
}

.todo-item-icon {
  font-size: 14px;
  line-height: 1;
  margin-top: 2px;
  flex-shrink: 0;
}

.todo-item.in_progress .todo-item-icon {
  animation: spin 2s linear infinite;
}

.todo-item-content {
  flex: 1;
  font-size: 0.85rem;
  line-height: 1.4;
  color: var(--text-primary);
  word-wrap: break-word;
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
