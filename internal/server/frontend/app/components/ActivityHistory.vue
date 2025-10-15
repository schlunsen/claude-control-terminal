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

    <!-- Session Selector -->
    <div class="session-selector" v-if="uniqueSessions.length > 0">
      <button
        class="session-trigger"
        @click="toggleSessionDropdown"
        :class="{ active: isSessionDropdownOpen }"
      >
        <div class="session-trigger-content">
          <img
            v-if="selectedSession"
            :src="useCharacterAvatar(selectedSession).avatar"
            :alt="useCharacterAvatar(selectedSession).name"
            class="trigger-avatar"
          />
          <svg v-else class="trigger-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="8" r="4"/>
            <path d="M6 21v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2"/>
          </svg>
          <span class="trigger-text">
            {{ selectedSession || 'All Sessions' }}
          </span>
        </div>
        <svg class="trigger-chevron" :class="{ open: isSessionDropdownOpen }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="6 9 12 15 18 9"/>
        </svg>
      </button>

      <div v-if="isSessionDropdownOpen" class="session-dropdown">
        <div class="session-option" @click="clearSessionFilter">
          <svg class="session-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="8" r="4"/>
            <path d="M6 21v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2"/>
          </svg>
          <span class="session-name">All Sessions</span>
          <span v-if="!selectedSession" class="session-check">✓</span>
        </div>
        <div class="session-divider"></div>
        <div
          v-for="session in uniqueSessions"
          :key="session.name"
          class="session-option"
          :class="{ selected: selectedSession === session.name }"
          @click="selectSession(session.name)"
        >
          <img
            :src="useCharacterAvatar(session.name).avatar"
            :alt="useCharacterAvatar(session.name).name"
            class="session-avatar"
          />
          <div class="session-info">
            <span class="session-name">{{ session.name }}</span>
            <span v-if="session.id" class="session-id">{{ session.id }}</span>
            <span v-if="session.startTime" class="session-time">Started {{ formatSessionTime(session.startTime) }}</span>
          </div>
          <span v-if="selectedSession === session.name" class="session-check">✓</span>
        </div>
      </div>
    </div>

    <!-- Filter tabs -->
    <div class="filter-tabs">
      <button
        v-for="tab in ['all', 'shell', 'claude', 'command', 'prompt', 'notification']"
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
          <span v-if="item.session_name" class="meta-item session-item">
            <img
              :src="useCharacterAvatar(item.session_name).avatar"
              :alt="useCharacterAvatar(item.session_name).name"
              :title="useCharacterAvatar(item.session_name).name"
              class="session-avatar"
            />
            <span class="meta-label">Session:</span> {{ item.session_name }}
          </span>
          <span v-if="item.git_branch" class="meta-item">
            <span class="meta-label">Branch:</span> {{ item.git_branch }}
          </span>
          <span v-if="item.type === 'shell' && item.exit_code !== undefined" :class="['meta-item', item.exit_code === 0 ? 'success' : 'error']">
            Exit: {{ item.exit_code }}
          </span>
          <span v-if="(item.type === 'claude' || item.type === 'command') && item.success !== undefined" :class="['meta-item', item.success ? 'success' : 'error']">
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
const filter = ref<'all' | 'shell' | 'claude' | 'prompt' | 'notification' | 'command'>('all')
const searchTerm = ref('')
const selectedSession = ref<string>('')
const isSessionDropdownOpen = ref(false)

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
  
  // Extract file_path from parameters if it's a file operation
  let filePath = cmd.file_path
  if (!filePath && cmd.parameters) {
    try {
      const params = typeof cmd.parameters === 'string' ? JSON.parse(cmd.parameters) : cmd.parameters
      filePath = params?.file_path || params?.path
    } catch (e) {
      // If parameters isn't valid JSON, ignore
    }
  }
  
  // Determine the correct type
  const itemType = message.type === 'claude' ? 'claude' : (message.type || 'command')
  
  history.value.unshift({
    id: cmd.id || Date.now(),
    type: itemType,
    command: cmd.command,
    tool_name: cmd.tool_name,
    session_name: cmd.session_name,
    git_branch: cmd.git_branch,
    conversation_id: cmd.conversation_id,
    working_directory: cmd.working_directory,
    exit_code: cmd.exit_code,
    success: cmd.success,
    description: cmd.description,
    parameters: typeof cmd.parameters === 'object' ? JSON.stringify(cmd.parameters) : cmd.parameters,
    duration_ms: cmd.duration_ms,
    executed_at: cmd.executed_at,
    file_path: filePath,
    timestamp: new Date(cmd.executed_at || cmd.timestamp),
    raw_data: cmd
  })
})

on('onHistoryCleared', () => {
  history.value = []
})

// Get unique sessions from history with conversation IDs and start time
const uniqueSessions = computed(() => {
  const sessions = new Map<string, { name: string, id: string, startTime: Date | null }>()

  history.value.forEach(item => {
    if (item.session_name) {
      const existing = sessions.get(item.session_name)
      const itemTimestamp = item.timestamp ? new Date(item.timestamp) : null

      if (!existing) {
        sessions.set(item.session_name, {
          name: item.session_name,
          id: item.conversation_id || '',
          startTime: item.type === 'prompt' ? itemTimestamp : null
        })
      } else {
        // Update if this is a prompt and it's earlier than current start time
        if (item.type === 'prompt' && itemTimestamp) {
          if (!existing.startTime || itemTimestamp < existing.startTime) {
            existing.startTime = itemTimestamp
          }
        }
        // Update conversation ID if missing
        if (!existing.id && item.conversation_id) {
          existing.id = item.conversation_id
        }
      }
    }
  })

  return Array.from(sessions.values()).sort((a, b) => a.name.localeCompare(b.name))
})

// Computed filtered history
const filteredHistory = computed(() => {
  return history.value.filter(item => {
    // Filter by type
    if (filter.value !== 'all' && item.type !== filter.value) {
      return false
    }

    // Filter by session
    if (selectedSession.value && item.session_name !== selectedSession.value) {
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

// Toggle session dropdown
function toggleSessionDropdown() {
  isSessionDropdownOpen.value = !isSessionDropdownOpen.value
}

// Select session
function selectSession(session: string) {
  selectedSession.value = session
  isSessionDropdownOpen.value = false
}

// Clear session filter
function clearSessionFilter() {
  selectedSession.value = ''
  isSessionDropdownOpen.value = false
}

// Close dropdown when clicking outside
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.session-selector')) {
    isSessionDropdownOpen.value = false
  }
}

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
    case 'command':
      return item.tool_name || 'Command'
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
    // Special handling for TodoWrite to show todo content
    if (item.tool_name === 'TodoWrite' && item.parameters) {
      try {
        const params = typeof item.parameters === 'string' ? JSON.parse(item.parameters) : item.parameters
        if (params?.todos && Array.isArray(params.todos)) {
          const todoList = params.todos.map((todo: any) => `${todo.status === 'completed' ? '✅' : todo.status === 'in_progress' ? '🔄' : '📝'} ${todo.content}`).join('\n')
          return `Updated ${params.todos.length} todo${params.todos.length !== 1 ? 's' : ''}:\n${todoList}`
        }
      } catch (e) {
        // Fall back to default behavior if parsing fails
      }
    }
    
    // Show tool name with file path if available, like "Read(/path/to/file)"
    if (item.file_path && item.tool_name) {
      return `${item.tool_name}(${item.file_path})`
    }
    return item.tool_name || ''
  } else if (item.type === 'command') {
    // Show tool name with file path if available, like "Read(/path/to/file)"
    if (item.file_path && item.tool_name) {
      return `${item.tool_name}(${item.file_path})`
    }
    return item.command || item.description || item.tool_name || ''
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

function formatSessionTime(timestamp: Date): string {
  const now = new Date()
  const date = new Date(timestamp)
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString()
}


// Initial load from API
onMounted(async () => {
  // Add click outside listener
  document.addEventListener('click', handleClickOutside)

  // Load initial data
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
    // Error loading activity history
  }
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
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

.meta-item.session-item {
  gap: 8px;
}

.session-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 2px solid var(--border-color);
  background: var(--bg-primary);
  object-fit: cover;
  flex-shrink: 0;
  transition: all 0.2s ease;
}

.session-avatar:hover {
  transform: scale(1.1);
  border-color: var(--accent-purple);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
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

/* Session Selector Styles */
.session-selector {
  position: relative;
  margin-bottom: 16px;
}

.session-trigger {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 0.875rem;
  color: var(--text-primary);
  font-family: inherit;
}

.session-trigger:hover {
  background: var(--card-hover);
  border-color: var(--accent-purple);
}

.session-trigger.active {
  border-color: var(--accent-purple);
  background: var(--card-hover);
}

.session-trigger-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.trigger-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--border-color);
  background: var(--bg-primary);
}

.trigger-icon {
  width: 20px;
  height: 20px;
  color: var(--text-muted);
}

.trigger-text {
  font-weight: 500;
  color: var(--text-primary);
}

.trigger-chevron {
  width: 20px;
  height: 20px;
  color: var(--text-muted);
  transition: transform 0.2s ease;
}

.trigger-chevron.open {
  transform: rotate(180deg);
}

.session-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  max-height: 300px;
  overflow-y: auto;
  z-index: 100;
  animation: slideDown 0.2s ease;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.session-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.15s ease;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.session-option:hover {
  background: var(--card-hover);
}

.session-option.selected {
  background: var(--bg-secondary);
}

.session-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--border-color);
  background: var(--bg-primary);
  flex-shrink: 0;
}

.session-icon {
  width: 20px;
  height: 20px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.session-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.session-name {
  font-weight: 500;
  color: var(--text-primary);
}

.session-id {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-family: 'SF Mono', Monaco, 'Consolas', 'Courier New', monospace;
}

.session-time {
  font-size: 0.7rem;
  color: var(--accent-cyan);
  font-weight: 500;
}

.session-check {
  color: var(--accent-purple);
  font-weight: 600;
  font-size: 1rem;
  flex-shrink: 0;
}

.session-divider {
  height: 1px;
  background: var(--border-color);
  margin: 4px 0;
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

  .session-avatar {
    width: 28px;
    height: 28px;
  }
}
</style>
