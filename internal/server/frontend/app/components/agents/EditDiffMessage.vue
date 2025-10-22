<template>
  <div class="edit-diff-message" :class="`status-${status}`">
    <div class="diff-header" @click="expanded = !expanded" :style="{ cursor: 'pointer' }">
      <div class="diff-icon" :class="`status-${status}`">
        <span v-if="status === 'running'">✏️</span>
        <span v-else-if="status === 'completed'">✅</span>
        <span v-else>❌</span>
      </div>
      <div class="diff-title">
        <strong>Edit File</strong>
        <span class="diff-status">{{ statusText }}</span>
      </div>
      <div class="expand-toggle">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :style="{ transform: expanded ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </div>
    </div>

    <div class="file-path">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"></path>
        <polyline points="13 2 13 9 20 9"></polyline>
      </svg>
      {{ filePath }}
    </div>

    <transition name="expand">
      <div v-if="expanded && diffLines.length > 0" class="diff-container">
        <div class="diff-stats">
          <span class="additions">+{{ additionCount }}</span>
          <span class="deletions">-{{ deletionCount }}</span>
        </div>

        <div class="diff-content">
          <div
            v-for="(line, idx) in diffLines"
            :key="idx"
            class="diff-line"
            :class="line.type"
          >
            <span class="diff-marker">{{ line.marker }}</span>
            <span class="diff-text">{{ line.text }}</span>
          </div>
        </div>
      </div>
    </transition>

    <div v-if="replaceAll" class="replace-all-badge">
      Replace All Occurrences
    </div>
  </div>
</template>

<script setup lang="ts">
import { useDiff } from '~/composables/useDiff'

const props = defineProps<{
  filePath: string
  oldString: string
  newString: string
  replaceAll?: boolean
  status: 'running' | 'completed' | 'error'
}>()

const { computeContextualDiff, getDiffStats } = useDiff()

// Compute diff with context
const diffLines = computed(() => {
  return computeContextualDiff(props.oldString, props.newString, 3)
})

// Auto-expand small diffs, keep large ones collapsed
const isSmallDiff = computed(() => {
  return diffLines.value.length <= 20
})

const expanded = ref(isSmallDiff.value)

const statusText = computed(() => {
  switch (props.status) {
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

// Get diff statistics
const diffStats = computed(() => getDiffStats(diffLines.value))

const additionCount = computed(() => diffStats.value.additions)
const deletionCount = computed(() => diffStats.value.deletions)
</script>

<style scoped>
.edit-diff-message {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px;
  margin: 8px 0;
  max-width: 100%;
}

.status-running {
  border-left: 3px solid #3b82f6;
}

.status-completed {
  border-left: 3px solid #22c55e;
  opacity: 0.9;
}

.status-error {
  border-left: 3px solid #ef4444;
}

.diff-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
  user-select: none;
}

.diff-header:hover {
  background: rgba(139, 92, 246, 0.05);
  border-radius: 6px;
  margin: -4px;
  padding: 4px;
}

.expand-toggle {
  margin-left: auto;
  display: flex;
  align-items: center;
  color: var(--text-secondary);
}

.diff-icon {
  font-size: 18px;
  line-height: 1;
}

.diff-icon.status-running {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.diff-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.diff-title strong {
  font-size: 0.9rem;
  color: var(--text-primary);
}

.diff-status {
  font-size: 0.75rem;
  color: var(--text-secondary);
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
  margin-bottom: 10px;
  word-break: break-all;
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

.diff-stats {
  display: flex;
  gap: 12px;
  padding: 6px 12px;
  font-size: 0.75rem;
  font-weight: 600;
  font-family: monospace;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
}

.additions {
  color: #22c55e;
}

.deletions {
  color: #ef4444;
}

.diff-content {
  max-height: 400px;
  overflow-y: auto;
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
  font-size: 0.8rem;
  line-height: 1.5;
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

.replace-all-badge {
  margin-top: 10px;
  padding: 4px 8px;
  background: rgba(139, 92, 246, 0.1);
  border: 1px solid rgba(139, 92, 246, 0.3);
  border-radius: 4px;
  font-size: 0.75rem;
  color: var(--accent-purple);
  text-align: center;
  font-weight: 500;
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

/* Expand/Collapse transition */
.expand-enter-active,
.expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
  margin-top: 0;
}

.expand-enter-to,
.expand-leave-from {
  max-height: 500px;
  opacity: 1;
  margin-top: 10px;
}
</style>
