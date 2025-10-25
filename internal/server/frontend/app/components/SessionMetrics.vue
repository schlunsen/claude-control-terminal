<template>
  <div class="session-metrics" v-if="session">
    <!-- Header with Session ID and Status -->
    <div v-if="!hideHeader" class="metrics-header">
      <div class="header-row">
        <div class="session-badge">
          <span class="badge-label">Session</span>
          <span class="badge-value">{{ session.id.slice(0, 8) }}</span>
        </div>
        <div class="duration" v-if="sessionDuration">
          ⏱️ {{ sessionDuration }}
        </div>
      </div>
      <div class="header-row">
        <div class="status-badge" :class="session.status">
          <span class="status-dot"></span>
          {{ session.status }}
        </div>
      </div>
    </div>

    <!-- Working Directory & Git Branch Section -->
    <div class="environment-section">
      <div class="environment-header">
        <span class="environment-icon">📂</span>
        <span class="environment-title">Environment</span>
      </div>
      <div class="environment-details">
        <div class="environment-row" v-if="session.options?.working_directory">
          <span class="env-label">Working Directory</span>
          <div class="env-value-wrapper">
            <code class="env-value" :title="session.options.working_directory">{{ session.options.working_directory }}</code>
          </div>
        </div>
        <div class="environment-row" v-else>
          <span class="env-label">Working Directory</span>
          <span class="env-not-available">No working directory set</span>
        </div>
        <div class="environment-row" v-if="session.git_branch">
          <span class="env-label">Git Branch</span>
          <div class="git-branch-badge">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="6" y1="3" x2="6" y2="15"></line>
              <circle cx="18" cy="6" r="3"></circle>
              <circle cx="6" cy="18" r="3"></circle>
              <path d="M18 9a9 9 0 0 1-9 9"></path>
            </svg>
            <span>{{ session.git_branch }}</span>
          </div>
        </div>
        <div class="environment-row" v-else>
          <span class="env-label">Git Branch</span>
          <span class="env-not-available">{{ session.options?.working_directory ? 'Not a git repository' : 'No working directory set' }}</span>
        </div>
      </div>
    </div>

    <!-- Main Metrics Grid -->
    <div class="metrics-grid">
      <!-- Message Count Card -->
      <div class="metric-card message-metric">
        <div class="metric-icon">💬</div>
        <div class="metric-content">
          <div class="metric-label">Messages</div>
          <div class="metric-value">{{ messageCount }}</div>
          <div class="metric-bar">
            <div class="metric-fill" :style="{ width: messagePercentage + '%' }"></div>
          </div>
        </div>
      </div>

      <!-- Tools Used Card -->
      <div class="metric-card tools-metric">
        <div class="metric-icon">🛠️</div>
        <div class="metric-content">
          <div class="metric-label">Tools Used</div>
          <div class="metric-value">{{ toolStats.count }}</div>
          <div class="tools-list">
            <span
              v-for="(count, tool) in toolStats.byName"
              :key="tool"
              class="tool-badge-wrapper"
            >
              <span class="tool-badge">
                {{ getToolIcon(tool) }} {{ tool }}
              </span>
              <span class="tool-tooltip">{{ count }} use{{ count !== 1 ? 's' : '' }}</span>
            </span>
          </div>
        </div>
      </div>

      <!-- Permissions Card -->
      <div class="metric-card permissions-metric">
        <div class="metric-icon">🔐</div>
        <div class="metric-content">
          <div class="metric-label">Permissions</div>
          <div class="metric-values">
            <span class="approved">✅ {{ permissionStats.approved }}</span>
            <span class="denied">❌ {{ permissionStats.denied }}</span>
          </div>
          <div class="permission-bar">
            <div class="approved-bar" :style="{ width: approvalPercentage + '%' }" v-if="permissionStats.total > 0"></div>
            <div v-else class="empty-bar">No permissions yet</div>
          </div>
        </div>
      </div>

      <!-- Status Details Card -->
      <div class="metric-card status-metric">
        <div class="metric-icon">📊</div>
        <div class="metric-content">
          <div class="metric-label">Details</div>
          <div class="status-details">
            <div class="detail-row">
              <span class="detail-label">Mode:</span>
              <span class="detail-value permission-mode">{{ session.options?.permission_mode }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Tools:</span>
              <span class="detail-value">{{ (session.options?.tools || []).length }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tool Breakdown Section -->
    <div class="tools-breakdown" v-if="Object.keys(toolStats.byName).length > 0">
      <div class="breakdown-title">Tool Breakdown</div>
      <div class="tool-list">
        <div v-for="(count, tool) in toolStats.byName" :key="tool" class="tool-item">
          <div class="tool-header">
            <span class="tool-name">{{ getToolIcon(tool) }} {{ tool }}</span>
            <span class="tool-count">{{ count }} use{{ count !== 1 ? 's' : '' }}</span>
          </div>
          <div class="tool-bar">
            <div class="tool-fill" :style="{ width: getToolPercentage(count) + '%' }"></div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'

interface SessionMetricsData {
  id: string
  status: string
  message_count: number
  error_message?: string
  git_branch?: string
  model_name?: string
  options?: {
    working_directory?: string
    permission_mode?: string
    tools?: string[]
    provider?: string
    model?: string
  }
  created_at?: string
  updated_at?: string
  git_branch?: string
}

const props = defineProps<{
  session: SessionMetricsData | null
  messageCount?: number
  toolExecutions?: Record<string, number>
  permissionStats?: {
    approved: number
    denied: number
    total: number
  }
  hideHeader?: boolean
}>()

// Reactive data
const toolStats = ref({ count: 0, byName: {} as Record<string, number> })
const permissionStats = ref({ approved: 0, denied: 0, total: 0 })
const sessionStartTime = ref<Date | null>(null)
const sessionDuration = ref('')

// Computed values
const messagePercentage = computed(() => {
  const count = props.messageCount ?? props.session?.message_count ?? 0
  const max = Math.max(count, 20)
  return (count / max) * 100
})

const approvalPercentage = computed(() => {
  if (permissionStats.value.total === 0) return 0
  return (permissionStats.value.approved / permissionStats.value.total) * 100
})

// Methods
const getToolIcon = (tool: string): string => {
  const iconMap: Record<string, string> = {
    'Read': '📖',
    'Write': '✏️',
    'Edit': '🔧',
    'Bash': '⚡',
    'Glob': '🔍',
    'Grep': '🔎',
    'Task': '📋',
    'TodoWrite': '✅',
    'WebSearch': '🌐',
    'WebFetch': '📡',
  }
  return iconMap[tool] || '🛠️'
}

const getToolPercentage = (count: number): number => {
  const max = Math.max(...Object.values(toolStats.value.byName || {}), 1)
  return (count / max) * 100
}

const truncatePath = (path?: string): string => {
  if (!path) return 'Not set'
  if (path.length <= 30) return path
  const start = path.substring(0, 15)
  const end = path.substring(path.length - 12)
  return `${start}...${end}`
}

const getProviderDisplay = (provider?: string): string => {
  // Default to Anthropic if no provider is set (for backward compatibility)
  const actualProvider = provider || 'anthropic'

  const providerMap: Record<string, string> = {
    'anthropic': '🟣 Anthropic',
    'glm': '🤖 GLM',
    'deepseek': '🔍 DeepSeek',
    'openai': '🟢 OpenAI',
    'google': '🔵 Google',
    'azure': '☁️ Azure',
    'cohere': '🟠 Cohere',
    'custom': '⚙️ Custom'
  }

  return providerMap[actualProvider.toLowerCase()] || `🔧 ${actualProvider}`
}

const formatDuration = (startTime: Date): string => {
  const now = new Date()
  const diff = now.getTime() - startTime.getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`
  } else {
    return `${seconds}s`
  }
}

// Watch for prop changes
watch(
  () => props.toolExecutions,
  (newVal) => {
    if (newVal && typeof newVal === 'object') {
      // Count unique tools (number of keys in the object)
      const uniqueToolCount = Object.keys(newVal).length

      toolStats.value = {
        count: uniqueToolCount,
        byName: newVal
      }
    } else {
      toolStats.value = {
        count: 0,
        byName: {}
      }
    }
  },
  { immediate: true, deep: true }
)

watch(
  () => props.permissionStats,
  (newVal) => {
    if (newVal) {
      permissionStats.value = newVal
    }
  },
  { immediate: true }
)

watch(
  () => props.session?.created_at,
  (newVal) => {
    if (newVal) {
      sessionStartTime.value = new Date(newVal)
    }
  },
  { immediate: true }
)

// Update duration every second
onMounted(() => {
  const interval = setInterval(() => {
    if (sessionStartTime.value && props.session?.status !== 'ended') {
      sessionDuration.value = formatDuration(sessionStartTime.value)
    }
  }, 1000)

  return () => clearInterval(interval)
})

// Initial duration calculation
watch(sessionStartTime, (newVal) => {
  if (newVal && props.session?.status !== 'ended') {
    sessionDuration.value = formatDuration(newVal)
  }
})
</script>

<style scoped>
.session-metrics {
  background: linear-gradient(135deg, var(--card-bg), var(--bg-secondary));
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.session-metrics:hover {
  border-color: var(--accent-purple);
  box-shadow: 0 8px 24px rgba(139, 92, 246, 0.15);
}

/* Header */
.metrics-header {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.session-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  flex: 1;
}

.badge-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.badge-value {
  font-size: 0.9rem;
  color: var(--accent-purple);
  font-weight: 700;
  font-family: 'Monaco', 'Menlo', monospace;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  text-transform: capitalize;
  border: 1px solid var(--border-color);
  flex: 1;
}

.status-badge.idle {
  background: rgba(40, 167, 69, 0.1);
  color: #28a745;
  border-color: rgba(40, 167, 69, 0.3);
}

.status-badge.processing {
  background: rgba(23, 162, 184, 0.1);
  color: #17a2b8;
  border-color: rgba(23, 162, 184, 0.3);
  animation: pulse 2s infinite;
}

.status-badge.error {
  background: rgba(220, 53, 69, 0.1);
  color: #dc3545;
  border-color: rgba(220, 53, 69, 0.3);
}

.status-badge.ended {
  background: rgba(108, 117, 125, 0.1);
  color: #6c757d;
  border-color: rgba(108, 117, 125, 0.3);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  animation: statusPulse 2s infinite;
}

.duration {
  padding: 8px 14px;
  background: var(--bg-secondary);
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
  white-space: nowrap;
}

/* Metrics Grid */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.metric-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  transition: all 0.2s ease;
}

.metric-card:hover {
  border-color: var(--accent-purple);
  background: var(--card-bg);
}

.metric-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.metric-content {
  flex: 1;
  min-width: 0;
}

.metric-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.5px;
  margin-bottom: 6px;
  display: block;
}

.metric-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--accent-purple);
  margin-bottom: 8px;
}

.metric-values {
  display: flex;
  gap: 12px;
  font-size: 0.9rem;
  font-weight: 600;
  margin-bottom: 8px;
}

.metric-values .approved {
  color: #28a745;
}

.metric-values .denied {
  color: #dc3545;
}

/* Progress Bars */
.metric-bar {
  height: 6px;
  background: var(--bg-secondary);
  border-radius: 3px;
  overflow: hidden;
}

.metric-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--accent-purple), var(--accent-purple-hover));
  border-radius: 3px;
  transition: width 0.3s ease;
}

.permission-bar {
  height: 6px;
  background: var(--bg-secondary);
  border-radius: 3px;
  overflow: hidden;
  display: flex;
}

.approved-bar {
  background: linear-gradient(90deg, #28a745, #20c997);
  transition: width 0.3s ease;
}

.empty-bar {
  width: 100%;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 0.7rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 500;
}

/* Tools List */
.tools-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.tool-badge-wrapper {
  position: relative;
  display: inline-block;
}

.tool-badge {
  display: inline-block;
  padding: 4px 8px;
  background: var(--bg-secondary);
  border-radius: 4px;
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 500;
  white-space: nowrap;
  cursor: help;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.tool-badge:hover {
  background: var(--accent-purple);
  color: white;
  border-color: var(--accent-purple-hover);
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(139, 92, 246, 0.3);
}

.tool-tooltip {
  position: absolute;
  bottom: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%) scale(0.9);
  padding: 6px 12px;
  background: linear-gradient(135deg, #2d2d3a 0%, #1a1a24 100%);
  border: 1px solid var(--accent-purple);
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  color: white;
  white-space: nowrap;
  pointer-events: none;
  opacity: 0;
  transition: all 0.2s cubic-bezier(0.68, -0.55, 0.265, 1.55);
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4), 0 0 20px rgba(139, 92, 246, 0.3);
}

.tool-tooltip::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 5px solid transparent;
  border-top-color: #2d2d3a;
  filter: drop-shadow(0 1px 1px rgba(0, 0, 0, 0.3));
}

.tool-badge-wrapper:hover .tool-tooltip {
  opacity: 1;
  transform: translateX(-50%) scale(1);
}

/* Status Details */
.status-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-row {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 0.85rem;
}

.detail-label {
  color: var(--text-secondary);
  font-weight: 500;
  min-width: 70px;
}

.detail-value {
  color: var(--text-primary);
  font-weight: 600;
  font-family: 'Monaco', 'Menlo', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.permission-mode {
  display: inline-block;
  padding: 2px 6px;
  background: var(--bg-secondary);
  border-radius: 3px;
  font-size: 0.8rem;
}

.git-branch {
  color: var(--accent-purple);
  font-weight: 700;
}

/* Tools Breakdown */
.tools-breakdown {
  padding: 16px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  margin-bottom: 16px;
}

.breakdown-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.tool-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tool-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.tool-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
}

.tool-count {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.tool-bar {
  height: 6px;
  background: var(--bg-secondary);
  border-radius: 3px;
  overflow: hidden;
}

.tool-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--accent-purple), var(--accent-purple-hover));
  border-radius: 3px;
  transition: width 0.3s ease;
}

/* Animations */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

@keyframes statusPulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* Provider Badge - Kept for reuse in stats header */
.provider-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  color: white;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
  width: fit-content;
}

/* Environment Section */
.environment-section {
  padding: 16px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  margin-bottom: 16px;
}

.environment-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.environment-icon {
  font-size: 1.2rem;
}

.environment-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.environment-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.environment-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.env-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.env-value-wrapper {
  overflow-x: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--border-color) transparent;
}

.env-value {
  display: block;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.85rem;
  color: var(--text-primary);
  white-space: nowrap;
  overflow-x: auto;
}

.env-value::-webkit-scrollbar {
  height: 6px;
}

.env-value::-webkit-scrollbar-track {
  background: var(--bg-tertiary);
  border-radius: 3px;
}

.env-value::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
}

.env-value::-webkit-scrollbar-thumb:hover {
  background: var(--accent-purple);
}

.git-branch-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: linear-gradient(135deg, var(--accent-purple), var(--accent-purple-hover));
  border-radius: 8px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.9rem;
  font-weight: 600;
  color: white;
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.3);
}

.git-branch-badge svg {
  flex-shrink: 0;
}

.env-not-available {
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-style: italic;
}

/* Responsive */
@media (max-width: 768px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }

  .header-row {
    flex-direction: column;
    align-items: stretch;
  }

  .session-badge,
  .status-badge,
  .duration {
    width: 100%;
  }

  .metric-card {
    flex-direction: column;
  }

  .metric-icon {
    font-size: 1.2rem;
  }

  .metric-values {
    flex-direction: column;
    gap: 4px;
  }
}
</style>
