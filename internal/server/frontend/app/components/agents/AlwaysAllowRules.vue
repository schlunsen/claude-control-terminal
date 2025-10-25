<template>
  <div class="always-allow-rules">
    <div class="rules-header">
      <h3>Auto-Approved Actions</h3>
      <span v-if="rules.length > 0" class="rules-count">{{ rules.length }}</span>
    </div>

    <div v-if="rules.length === 0" class="no-rules">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
        <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
      </svg>
      <p>No auto-approved actions yet</p>
      <p class="hint">Use "Always Exact" or "Allow Similar" on permission requests</p>
    </div>

    <div v-else class="rules-list">
      <div v-for="rule in rules" :key="rule.id" class="rule-item">
        <div class="rule-badge" :class="rule.match_mode">
          <svg v-if="rule.match_mode === 'exact'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2L2 7l10 5 10-5-10-5z"></path>
            <path d="M2 17l10 5 10-5M2 12l10 5 10-5"></path>
          </svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path>
            <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>
          </svg>
          {{ rule.match_mode === 'exact' ? 'Exact' : 'Similar' }}
        </div>

        <div class="rule-content">
          <div class="rule-description">{{ rule.description }}</div>

          <div v-if="rule.match_mode === 'pattern'" class="rule-pattern">
            {{ formatPattern(rule) }}
          </div>

          <div class="rule-meta">
            <span class="rule-tool">{{ rule.tool }}</span>
            <span class="rule-time">{{ formatRelativeTime(rule.created_at) }}</span>
          </div>
        </div>

        <button
          @click="$emit('remove', rule.id)"
          class="btn-remove"
          title="Remove this rule"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>
    </div>

    <button
      v-if="rules.length > 0"
      @click="$emit('clear-all')"
      class="btn-clear-all"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <polyline points="3 6 5 6 21 6"></polyline>
        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
      </svg>
      Clear All Rules
    </button>
  </div>
</template>

<script setup lang="ts">
interface AlwaysAllowRule {
  id: string
  tool: string
  match_mode: 'exact' | 'pattern'
  parameters?: any
  pattern?: {
    command_prefix?: string
    directory_path?: string
    file_path_pattern?: string
    path_pattern?: string
  }
  description: string
  created_at: string
}

interface Props {
  rules: AlwaysAllowRule[]
}

defineProps<Props>()

defineEmits<{
  (e: 'remove', ruleId: string): void
  (e: 'clear-all'): void
}>()

const formatPattern = (rule: AlwaysAllowRule) => {
  if (!rule.pattern) return 'Pattern match'

  // Check if this is a wildcard rule (allow all)
  const isWildcard =
    rule.pattern.command_prefix === '*' ||
    rule.pattern.directory_path === '*' ||
    rule.pattern.path_pattern === '*'

  if (isWildcard) {
    // Wildcard pattern - allow ALL of this tool type
    switch (rule.tool) {
      case 'Bash':
        return `All Bash commands`
      case 'Read':
        return `All Read operations (any file)`
      case 'Write':
        return `All Write operations (any file)`
      case 'Edit':
        return `All Edit operations (any file)`
      case 'Grep':
        return `All Grep operations`
      case 'Glob':
        return `All Glob operations`
      default:
        return `All ${rule.tool} operations`
    }
  }

  // Non-wildcard patterns (for future use)
  switch (rule.tool) {
    case 'Bash':
      if (rule.pattern.command_prefix) {
        return `All commands: ${rule.pattern.command_prefix}*`
      }
      break

    case 'Read':
      if (rule.pattern.directory_path) {
        return `All files in: ${rule.pattern.directory_path}/`
      }
      break

    case 'Write':
      if (rule.pattern.directory_path) {
        return `All writes to: ${rule.pattern.directory_path}/`
      }
      break

    case 'Edit':
      if (rule.pattern.directory_path) {
        return `All edits in: ${rule.pattern.directory_path}/`
      }
      break

    case 'Grep':
    case 'Glob':
      if (rule.pattern.path_pattern) {
        return `Pattern: ${rule.pattern.path_pattern}`
      }
      break
  }

  return 'Pattern match'
}

const formatRelativeTime = (timestamp: string) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}d ago`
  if (hours > 0) return `${hours}h ago`
  if (minutes > 0) return `${minutes}m ago`
  return 'Just now'
}
</script>

<style scoped>
.always-allow-rules {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.rules-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.rules-header h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.rules-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.5rem;
  height: 1.5rem;
  padding: 0 0.5rem;
  background: var(--accent-purple);
  color: white;
  border-radius: 0.75rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.no-rules {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem;
  text-align: center;
  color: var(--text-secondary);
}

.no-rules svg {
  opacity: 0.3;
  margin-bottom: 1rem;
}

.no-rules p {
  margin: 0.25rem 0;
}

.no-rules .hint {
  font-size: 0.875rem;
  opacity: 0.7;
}

.rules-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.rule-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.875rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  transition: all 0.2s;
}

.rule-item:hover {
  border-color: var(--accent-purple);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.rule-badge {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  border-radius: 0.375rem;
  font-size: 0.75rem;
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
}

.rule-badge.exact {
  background: #dbeafe;
  color: #1e40af;
}

.rule-badge.pattern {
  background: #fef3c7;
  color: #92400e;
}

.rule-content {
  flex: 1;
  min-width: 0;
}

.rule-description {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 0.375rem;
  word-break: break-word;
}

.rule-pattern {
  font-size: 0.8125rem;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  color: var(--text-secondary);
  background: rgba(0, 0, 0, 0.03);
  padding: 0.375rem 0.5rem;
  border-radius: 0.25rem;
  margin-bottom: 0.375rem;
  word-break: break-all;
}

.rule-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.rule-tool {
  padding: 0.125rem 0.5rem;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 0.25rem;
  font-weight: 500;
}

.rule-time {
  opacity: 0.7;
}

.btn-remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 0.375rem;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.btn-remove:hover {
  background: var(--status-error);
  color: white;
}

.btn-clear-all {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-clear-all:hover {
  background: var(--status-error);
  border-color: var(--status-error);
  color: white;
}
</style>
