<template>
  <div class="permission-request" :class="{ 'disconnected': !connected }">
    <div v-if="!connected" class="connection-warning">
      ⚠️ Connection lost - this request may no longer be valid
    </div>
    <div class="permission-header">
      <div class="permission-icon">🔐</div>
      <div class="permission-title">Permission Request</div>
      <div class="permission-time">{{ formatTime(permission.timestamp) }}</div>
    </div>
    <div class="permission-description">
      {{ permission.description }}
    </div>

    <!-- Show pattern preview for "Allow Similar" -->
    <div v-if="similarPattern" class="similar-preview">
      <div class="preview-label">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="16" x2="12" y2="12"></line>
          <line x1="12" y1="8" x2="12.01" y2="8"></line>
        </svg>
        <strong>Allow Similar will approve:</strong>
      </div>
      <code>{{ similarPattern }}</code>
    </div>

    <div class="permission-actions">
      <button
        @click="$emit('deny', permission)"
        class="btn-deny"
        :disabled="!connected"
        title="Deny this request"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
        Deny
      </button>

      <button
        @click="$emit('approve-exact', permission)"
        class="btn-approve-exact"
        :disabled="!connected"
        title="Always allow this exact command/file"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2L2 7l10 5 10-5-10-5z"></path>
          <path d="M2 17l10 5 10-5M2 12l10 5 10-5"></path>
        </svg>
        Always Exact
      </button>

      <button
        @click="$emit('approve-similar', permission)"
        class="btn-approve-similar"
        :disabled="!connected || !similarPattern"
        :title="similarPattern ? `Allow similar: ${similarPattern}` : 'No pattern available'"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path>
          <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>
        </svg>
        Allow Similar
      </button>

      <button
        @click="$emit('approve', permission)"
        class="btn-approve"
        :disabled="!connected"
        title="Approve this request once"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="20,6 9,17 4,12"></polyline>
        </svg>
        Approve Once
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatTime } from '~/utils/agents/messageFormatters'

interface Permission {
  request_id: string
  description: string
  timestamp: string | Date
  tool: string
  details: any
}

interface Props {
  permission: Permission
  connected?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  connected: true
})

defineEmits<{
  (e: 'approve', permission: Permission): void
  (e: 'approve-exact', permission: Permission): void
  (e: 'approve-similar', permission: Permission): void
  (e: 'deny', permission: Permission): void
}>()

// Generate preview of what "Allow Similar" will match
const similarPattern = computed(() => {
  const { tool, details } = props.permission

  // "Allow Similar" creates a pattern based on the command/file
  switch (tool) {
    case 'Bash': {
      // Extract command name from the bash command
      if (details?.command) {
        const command = details.command.trim()
        const commandName = command.split(/\s+/)[0] // Get first word
        return `Bash(${commandName}:*)`
      }
      return `Bash(*)`
    }

    case 'Read':
      return `Read(/**)`

    case 'Write':
      return `Write(/**)`

    case 'Edit':
      return `Edit(/**)`

    case 'Grep':
      return `Grep(*)`

    case 'Glob':
      return `Glob(*)`

    default:
      return `${tool}(*)`
  }
})
</script>

<style scoped>
.permission-request {
  background: rgba(251, 191, 36, 0.1);
  border: 2px solid var(--status-warning);
  border-radius: 0.5rem;
  padding: 1rem;
  margin-bottom: 1rem;
}

.permission-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.permission-icon {
  font-size: 1.5rem;
}

.permission-title {
  flex: 1;
  font-weight: 600;
  color: var(--text-primary);
}

.permission-time {
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.permission-description {
  color: var(--text-primary);
  margin-bottom: 1rem;
  line-height: 1.5;
}

.similar-preview {
  background: #e8f4f8;
  border-left: 3px solid #3b82f6;
  padding: 0.75rem;
  margin: 0.75rem 0;
  border-radius: 0.375rem;
}

.preview-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: #1e40af;
}

.preview-label svg {
  flex-shrink: 0;
}

.similar-preview code {
  display: block;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.875rem;
  color: #1e40af;
  background: rgba(59, 130, 246, 0.1);
  padding: 0.5rem;
  border-radius: 0.25rem;
}

.permission-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.btn-deny,
.btn-approve,
.btn-approve-exact,
.btn-approve-similar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.875rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-deny {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.btn-deny:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--text-secondary);
  transform: translateY(-1px);
}

.btn-approve-exact {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.btn-approve-exact:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
  color: var(--accent-purple);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.2);
}

.btn-approve-similar {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.btn-approve-similar:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
  color: var(--accent-purple);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.2);
}

.btn-approve {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.btn-approve:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
  color: var(--accent-purple);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.2);
}

.btn-deny:disabled,
.btn-approve:disabled,
.btn-approve-exact:disabled,
.btn-approve-similar:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.permission-request.disconnected {
  opacity: 0.7;
  border-color: var(--status-error);
}

.connection-warning {
  background: rgba(248, 113, 113, 0.1);
  border: 1px solid var(--status-error);
  border-radius: 0.375rem;
  padding: 0.5rem;
  margin-bottom: 0.75rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--status-error);
  text-align: center;
}
</style>
