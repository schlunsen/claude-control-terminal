<template>
  <div class="activity-history">
    <div class="header">
      <h2>Activity History</h2>
      <div v-if="connected" class="status connected">
        <span class="dot"></span>
        Live
      </div>
      <div v-else class="status disconnected">
        <span class="dot"></span>
        Disconnected
      </div>
    </div>

    <!-- Search -->
    <input
      v-model="searchTerm"
      type="text"
      placeholder="Search activity..."
      class="search-input"
    />

    <!-- Filter tabs -->
    <div class="filter-tabs">
      <button
        v-for="tab in ['all', 'shell', 'claude', 'prompt', 'notification']"
        :key="tab"
        :class="{ active: filter === tab }"
        @click="filter = tab"
        class="filter-tab"
      >
        {{ tab.charAt(0).toUpperCase() + tab.slice(1) }}
      </button>
    </div>

    <!-- Activity items -->
    <div class="activity-list">
      <div v-if="filteredHistory.length === 0" class="empty-state">
        No activity found
      </div>
      <div
        v-for="item in filteredHistory"
        :key="item.id"
        :class="['activity-item', getItemClass(item)]"
      >
        <div class="activity-header">
          <div class="activity-tool">{{ getToolLabel(item) }}</div>
          <div class="activity-time">{{ formatTime(item.timestamp) }}</div>
        </div>
        <div class="activity-content">{{ getDisplayText(item) }}</div>
        <div class="activity-meta">
          <span v-if="item.session_name" class="meta-item">
            <span class="meta-label">Session:</span> {{ item.session_name }}
          </span>
          <span v-if="item.git_branch" class="meta-item">
            <span class="meta-label">Branch:</span> {{ item.git_branch }}
          </span>
          <span v-if="item.type === 'shell' && item.exit_code !== undefined" :class="['meta-item', item.exit_code === 0 ? 'success' : 'error']">
            Exit: {{ item.exit_code }}
          </span>
          <span v-if="item.type === 'claude'" :class="['meta-item', item.success ? 'success' : 'error']">
            {{ item.success ? 'Success' : 'Error' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ActivityItem } from '~/types/analytics'

// State - ONLY updated via WebSocket or initial load
const history = ref<ActivityItem[]>([])
const filter = ref<'all' | 'shell' | 'claude' | 'prompt' | 'notification'>('all')
const searchTerm = ref('')

// WebSocket integration
const { connected, on } = useWebSocket()

// Handle WebSocket events - prepend new items
on('onNotification', (data: any) => {
  console.log('📝 Received notification:', data)
  history.value.unshift({
    id: data.id || Date.now(),
    type: 'notification',
    ...data,
    notification_type: data.notification_type,
    timestamp: new Date(data.notified_at || data.timestamp)
  })
})

on('onPrompt', (data: any) => {
  console.log('📝 Received prompt:', data)
  history.value.unshift({
    id: data.id || Date.now(),
    type: 'prompt',
    message: data.message,
    session_name: data.session_name,
    git_branch: data.git_branch,
    conversation_id: data.conversation_id,
    working_directory: data.working_directory,
    timestamp: new Date(data.submitted_at || data.timestamp)
  })
})

on('onCommand', (message: any) => {
  console.log('📝 Received command:', message)
  const cmd = message.data || message
  history.value.unshift({
    id: cmd.id || Date.now(),
    type: message.type || 'command',
    command: cmd.command,
    tool_name: cmd.tool_name,
    session_name: cmd.session_name,
    git_branch: cmd.git_branch,
    conversation_id: cmd.conversation_id,
    working_directory: cmd.working_directory,
    exit_code: cmd.exit_code,
    success: cmd.success,
    timestamp: new Date(cmd.executed_at || cmd.timestamp)
  })
})

on('onHistoryCleared', () => {
  history.value = []
})

// Computed filtered history
const filteredHistory = computed(() => {
  return history.value.filter(item => {
    // Filter by type
    if (filter.value !== 'all' && item.type !== filter.value) {
      return false
    }

    // Filter by search term
    if (searchTerm.value) {
      const searchLower = searchTerm.value.toLowerCase()
      const searchableText = [
        item.command,
        item.tool_name,
        item.message,
        item.git_branch,
        item.session_name
      ].filter(Boolean).join(' ').toLowerCase()

      return searchableText.includes(searchLower)
    }

    return true
  })
})

// Helper functions
function getItemClass(item: ActivityItem): string {
  if (item.type === 'notification') {
    return `notification-${item.notification_type}`
  }
  return item.type
}

function getToolLabel(item: ActivityItem): string {
  switch (item.type) {
    case 'shell':
      return 'Bash'
    case 'claude':
      return item.tool_name || 'Claude Tool'
    case 'prompt':
      return '💬 User Prompt'
    case 'notification':
      if (item.notification_type === 'permission_request') return '🔐 Permission'
      if (item.notification_type === 'idle_alert') return '⏱️ Idle'
      return '🔔 Notification'
    default:
      return 'Unknown'
  }
}

function getDisplayText(item: ActivityItem): string {
  if (item.type === 'shell') {
    return item.command || ''
  } else if (item.type === 'claude') {
    return item.tool_name || ''
  } else if (item.type === 'prompt') {
    const msg = item.message || ''
    return msg.length > 200 ? msg.substring(0, 200) + '...' : msg
  } else if (item.type === 'notification') {
    return item.command_details || item.message || ''
  }
  return ''
}

function formatTime(timestamp: Date | string): string {
  const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp
  return date.toLocaleTimeString()
}

// Initial load from API
onMounted(async () => {
  try {
    const { data } = await useFetch<any>('/api/history/all?limit=100')
    if (data.value?.history) {
      history.value = data.value.history.map((item: any) => ({
        ...item.content,
        id: item.id || item.content.id,
        type: item.type,
        timestamp: new Date(item.timestamp)
      }))
    }
  } catch (error) {
    console.error('Error loading activity history:', error)
  }
})
</script>

<style scoped>
.activity-history {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 24px;
  height: fit-content;
  transition: all 0.3s ease;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header h2 {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
}

.status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 600;
}

.status.connected {
  background: var(--status-success);
  color: var(--bg-primary);
}

.status.disconnected {
  background: var(--status-error);
  color: var(--bg-primary);
}

.status .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
}

.search-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 0.875rem;
  margin-bottom: 16px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  transition: all 0.2s ease;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-input:focus {
  outline: none;
  border-color: var(--accent-purple);
  background: var(--card-hover);
}

.filter-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 0;
}

.filter-tab {
  padding: 8px 16px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
}

.filter-tab:hover {
  color: var(--text-secondary);
}

.filter-tab.active {
  color: var(--accent-purple);
  border-bottom-color: var(--accent-purple);
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 70vh;
  overflow-y: auto;
  padding-right: 8px;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-muted);
  font-size: 0.875rem;
}

.activity-item {
  padding: 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  transition: all 0.2s ease;
}

.activity-item:hover {
  border-color: var(--accent-cyan);
  background: var(--card-hover);
}

.activity-item.notification-permission_request {
  background: var(--bg-secondary);
  border-left: 4px solid var(--status-error);
}

.activity-item.notification-idle_alert {
  background: var(--bg-secondary);
  border-left: 4px solid var(--accent-cyan);
}

.activity-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.activity-tool {
  font-family: 'SF Mono', Monaco, 'Consolas', 'Courier New', monospace;
  font-size: 0.875rem;
  color: var(--accent-purple);
  font-weight: 600;
}

.activity-time {
  font-size: 0.8125rem;
  color: var(--text-muted);
}

.activity-content {
  font-family: 'SF Mono', Monaco, 'Consolas', 'Courier New', monospace;
  font-size: 0.8125rem;
  color: var(--text-primary);
  padding: 8px 12px;
  background: var(--code-bg);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  margin-bottom: 8px;
  word-break: break-all;
  white-space: pre-wrap;
}

.activity-meta {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.meta-label {
  color: var(--text-muted);
}

.meta-item.success {
  color: var(--status-success);
  font-weight: 600;
}

.meta-item.error {
  color: var(--status-error);
  font-weight: 600;
}

@media (max-width: 768px) {
  .activity-history {
    padding: 24px;
  }

  .activity-list {
    max-height: 400px;
  }

  .filter-tabs {
    gap: 4px;
    overflow-x: auto;
  }

  .filter-tab {
    padding: 6px 12px;
    font-size: 0.8125rem;
    white-space: nowrap;
  }

  .activity-item {
    padding: 12px;
  }
}
</style>
