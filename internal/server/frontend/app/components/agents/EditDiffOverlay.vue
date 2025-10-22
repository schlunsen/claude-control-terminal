<template>
  <transition name="slide-fade">
    <div v-if="visible" class="edit-diff-overlay" :class="{ 'tool-completed': tool.status === 'completed' }">
      <div class="tool-header">
        <div class="tool-icon" :class="`tool-${tool.status}`">
          <span v-if="tool.status === 'running'">✏️</span>
          <span v-else-if="tool.status === 'completed'">✅</span>
          <span v-else>❌</span>
        </div>
        <div class="tool-info">
          <div class="tool-name">Edit File</div>
          <div class="tool-status">{{ statusText }}</div>
        </div>
        <button class="close-btn" @click="handleManualClose" title="Dismiss">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <div class="file-path">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
          <polyline points="13 2 13 9 20 9"></polyline>
        </svg>
        {{ filePath }}
      </div>

      <div v-if="diffLines.length > 0" class="diff-container">
        <div class="diff-header">
          <span class="diff-stats">
            <span class="additions">+{{ additionCount }}</span>
            <span class="deletions">-{{ deletionCount }}</span>
          </span>
          <button v-if="diffLines.length > 5" @click="expanded = !expanded" class="expand-btn">
            {{ expanded ? 'Show Less' : 'Show Full Diff' }}
          </button>
        </div>

        <div class="diff-content" :class="{ collapsed: !expanded && diffLines.length > 5 }">
          <div
            v-for="(line, idx) in displayLines"
            :key="idx"
            class="diff-line"
            :class="line.type"
          >
            <span class="diff-marker">{{ line.marker }}</span>
            <span class="diff-text">{{ line.text }}</span>
          </div>
          <div v-if="!expanded && diffLines.length > 5" class="diff-truncated">
            ... {{ diffLines.length - 3 }} more lines
          </div>
        </div>
      </div>

      <div v-if="replaceAll" class="replace-all-badge">
        Replace All Occurrences
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
const expanded = ref(false)
const autoDismissDelay = props.autoDismissDelay || 5000

const statusText = computed(() => {
  switch (props.tool.status) {
    case 'running':
      return 'Editing...'
    case 'completed':
      return 'Completed'
    case 'error':
      return 'Error'
    default:
      return ''
  }
})

const filePath = computed(() => {
  return props.tool.input?.file_path || 'Unknown file'
})

const oldString = computed(() => {
  return props.tool.input?.old_string || ''
})

const newString = computed(() => {
  return props.tool.input?.new_string || ''
})

const replaceAll = computed(() => {
  return props.tool.input?.replace_all === true
})

interface DiffLine {
  type: 'addition' | 'deletion' | 'context'
  marker: string
  text: string
}

const diffLines = computed<DiffLine[]>(() => {
  const old = oldString.value
  const newStr = newString.value

  if (!old && !newStr) return []

  const lines: DiffLine[] = []

  // Split by lines for better diff display
  const oldLines = old.split('\n')
  const newLines = newStr.split('\n')

  // Simple diff: show deletions then additions
  // For a more sophisticated diff, we could use a diff library

  if (oldLines.length === 1 && newLines.length === 1) {
    // Single line change - show inline
    if (old !== newStr) {
      lines.push({ type: 'deletion', marker: '-', text: old })
      lines.push({ type: 'addition', marker: '+', text: newStr })
    }
  } else {
    // Multi-line change
    oldLines.forEach(line => {
      lines.push({ type: 'deletion', marker: '-', text: line })
    })
    newLines.forEach(line => {
      lines.push({ type: 'addition', marker: '+', text: line })
    })
  }

  return lines
})

const additionCount = computed(() => {
  return diffLines.value.filter(l => l.type === 'addition').length
})

const deletionCount = computed(() => {
  return diffLines.value.filter(l => l.type === 'deletion').length
})

const displayLines = computed(() => {
  if (expanded.value || diffLines.value.length <= 5) {
    return diffLines.value
  }
  // Show first 3 lines when collapsed
  return diffLines.value.slice(0, 3)
})

// Manual close handler
const handleManualClose = () => {
  visible.value = false
  setTimeout(() => {
    emit('dismiss', props.tool.id)
  }, 300) // Wait for animation
}

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
.edit-diff-overlay {
  position: relative;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 8px;
  min-width: 400px;
  max-width: 600px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.edit-diff-overlay:hover {
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
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
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

.close-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

.close-btn:active {
  transform: scale(0.95);
}

.file-path {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
  padding: 6px 10px;
  background: var(--bg-secondary);
  border-radius: 6px;
  margin-bottom: 12px;
}

.file-path svg {
  flex-shrink: 0;
}

.diff-container {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow: hidden;
  background: var(--bg-secondary);
}

.diff-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
}

.diff-stats {
  display: flex;
  gap: 12px;
  font-size: 0.8rem;
  font-weight: 600;
  font-family: monospace;
}

.additions {
  color: #22c55e;
}

.deletions {
  color: #ef4444;
}

.expand-btn {
  background: none;
  border: none;
  color: var(--accent-purple);
  font-size: 0.75rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background 0.2s;
}

.expand-btn:hover {
  background: rgba(139, 92, 246, 0.1);
}

.diff-content {
  max-height: 300px;
  overflow-y: auto;
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
  font-size: 0.8rem;
  line-height: 1.5;
}

.diff-content.collapsed {
  max-height: 120px;
}

.diff-line {
  display: flex;
  padding: 2px 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

.diff-line.addition {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

.diff-line.deletion {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.diff-line.context {
  color: var(--text-secondary);
}

.diff-marker {
  flex-shrink: 0;
  width: 20px;
  font-weight: bold;
  user-select: none;
}

.diff-text {
  flex: 1;
  min-width: 0;
}

.diff-truncated {
  padding: 8px 12px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 0.75rem;
  font-style: italic;
  background: var(--bg-primary);
  border-top: 1px solid var(--border-color);
}

.replace-all-badge {
  margin-top: 8px;
  padding: 4px 8px;
  background: rgba(139, 92, 246, 0.1);
  border: 1px solid rgba(139, 92, 246, 0.3);
  border-radius: 4px;
  font-size: 0.75rem;
  color: var(--accent-purple);
  text-align: center;
  font-weight: 500;
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

/* Scrollbar styling */
.diff-content::-webkit-scrollbar {
  width: 6px;
}

.diff-content::-webkit-scrollbar-track {
  background: var(--bg-primary);
}

.diff-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
}

.diff-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-secondary);
}
</style>
