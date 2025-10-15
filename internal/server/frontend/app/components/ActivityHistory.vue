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
  history.value.unshift({
    id: data.id || Date.now(),
    type: 'notification',
    ...data,
    notification_type: data.notification_type,
    timestamp: new Date(data.notified_at || data.timestamp)
  })
})

on('onPrompt', (data: any) => {
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
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 32px;
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
}

.status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
}

.status.connected {
  background: #e8f5e9;
  color: #2e7d32;
}

.status.disconnected {
  background: #ffebee;
  color: #c62828;
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
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  font-size: 0.875rem;
  margin-bottom: 16px;
}

.search-input:focus {
  outline: none;
  border-color: #666;
}

.filter-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
  border-bottom: 1px solid #e0e0e0;
  padding-bottom: 0;
}

.filter-tab {
  padding: 8px 16px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  font-size: 0.875rem;
  font-weight: 500;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-tab:hover {
  color: #1a1a1a;
}

.filter-tab.active {
  color: #1a1a1a;
  border-bottom-color: #1a1a1a;
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 600px;
  overflow-y: auto;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #999;
  font-size: 0.875rem;
}

.activity-item {
  padding: 16px;
  background: #fafafa;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  transition: border-color 0.2s;
}

.activity-item:hover {
  border-color: #d0d0d0;
}

.activity-item.notification-permission_request {
  background: #fef2f2;
  border-left: 4px solid #ef4444;
}

.activity-item.notification-idle_alert {
  background: #f0f9ff;
  border-left: 4px solid #3b82f6;
}

.activity-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.activity-tool {
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 0.875rem;
  color: #1a1a1a;
  font-weight: 500;
}

.activity-time {
  font-size: 0.8125rem;
  color: #999;
}

.activity-content {
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 0.8125rem;
  color: #1a1a1a;
  padding: 8px 12px;
  background: #fff;
  border: 1px solid #e8e8e8;
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
  color: #666;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.meta-label {
  color: #999;
}

.meta-item.success {
  color: #2e7d32;
}

.meta-item.error {
  color: #c53030;
}
</style>
