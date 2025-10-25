<template>
  <div v-if="toolExecution" class="tool-execution-bar">
    <div class="tool-execution-content">
      <div class="tool-execution-icon">
        <span>{{ getToolIcon(toolExecution.toolName) }}</span>
      </div>
      <div class="tool-execution-details">
        <div class="tool-execution-name">
          {{ toolExecution.toolName }}
          <span v-if="toolExecution.detail" class="tool-execution-detail-badge">
            {{ truncatePath(toolExecution.detail, 40) }}
          </span>
        </div>
        <div class="tool-execution-info">
          <span v-if="toolExecution.command">{{ truncatePath(toolExecution.command, 60) }}</span>
          <span v-else-if="toolExecution.filePath">{{ truncatePath(toolExecution.filePath, 60) }}</span>
          <span v-else-if="toolExecution.pattern">{{ truncatePath(toolExecution.pattern, 60) }}</span>
          <span v-else>Executing...</span>
        </div>
      </div>
      <div class="tool-execution-pulse"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getToolIcon } from '~/utils/agents/toolParser'
import { truncatePath } from '~/utils/agents/messageFormatters'

interface ToolExecution {
  toolName: string
  filePath?: string
  command?: string
  pattern?: string
  detail?: string
  timestamp: Date
}

interface Props {
  toolExecution: ToolExecution | null
}

defineProps<Props>()
</script>

<style scoped>
.tool-execution-bar {
  padding: 0.75rem;
  background: rgba(139, 92, 246, 0.1);
  border-left: 3px solid var(--accent-purple);
  margin-bottom: 1rem;
  border-radius: 0.375rem;
}

.tool-execution-content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.tool-execution-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.tool-execution-details {
  flex: 1;
  min-width: 0;
}

.tool-execution-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.tool-execution-detail-badge {
  padding: 0.125rem 0.5rem;
  background: var(--bg-secondary);
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-secondary);
  font-family: 'Monaco', 'Courier New', monospace;
}

.tool-execution-info {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-top: 0.25rem;
  font-family: 'Monaco', 'Courier New', monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tool-execution-pulse {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--accent-purple);
  flex-shrink: 0;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.5;
    transform: scale(1.2);
  }
}
</style>
