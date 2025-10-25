<template>
  <div class="session-metrics" v-if="session">
    <!-- Session Info Section (Collapsible) -->
    <div class="collapsible-section">
      <button class="section-header" @click="toggleSection('sessionInfo')">
        <span class="header-content">
          <span class="section-icon">ℹ️</span>
          <span class="section-title">Session Info</span>
        </span>
        <svg class="toggle-arrow" :class="{ 'collapsed': !expandedSections.sessionInfo }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>
      <div v-show="expandedSections.sessionInfo" class="section-content">
        <!-- Context Usage Bar -->
        <ContextUsageBar
          :usage="contextUsage"
          :loading="contextLoading"
          @refresh="$emit('refresh-context')"
        />

        <!-- Environment Card -->
        <div class="metric-card environment-metric">
          <div class="metric-content">
            <div class="metric-label">
              <span class="metric-label-icon">📂</span>
              <span>Environment</span>
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
        </div>
      </div>
    </div>

    <!-- Tools and Permissions Section (Collapsible) -->
    <div class="collapsible-section" v-if="Object.keys(toolStats.byName).length > 0">
      <button class="section-header" @click="toggleSection('toolsPermissions')">
        <span class="header-content">
          <span class="section-icon">🔧</span>
          <span class="section-title">Tools & Permissions</span>
        </span>
        <svg class="toggle-arrow" :class="{ 'collapsed': !expandedSections.toolsPermissions }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>
      <div v-show="expandedSections.toolsPermissions" class="section-content">
        <!-- Project Permissions Card (First) -->
        <div v-if="projectPermissions" class="metric-card project-permissions-metric">
          <div class="metric-content">
            <div class="metric-label">
              <span class="metric-label-icon">🔐</span>
              <span>Project Permissions</span>
            </div>
            <ProjectPermissions :permissions="projectPermissions" />
          </div>
        </div>

        <!-- Summary Cards Grid -->
        <div class="metrics-grid">
          <!-- Tools Used Card -->
          <div class="metric-card tools-metric">
            <div class="metric-content">
              <div class="metric-label">
                <span class="metric-label-icon">🛠️</span>
                <span>Tools Used</span>
              </div>
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
            <div class="metric-content">
              <div class="metric-label">
                <span class="metric-label-icon">🔐</span>
                <span>Permissions</span>
              </div>
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
            <div class="metric-content">
              <div class="metric-label">
                <span class="metric-label-icon">⚙️</span>
                <span>Details</span>
              </div>
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

        <!-- Tool Breakdown Card -->
        <div class="metric-card tool-breakdown-metric">
          <div class="metric-content">
            <div class="metric-label">
              <span class="metric-label-icon">📊</span>
              <span>Tool Breakdown</span>
            </div>
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
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import ProjectPermissions from '~/components/agents/ProjectPermissions.vue'
import ContextUsageBar from '~/components/agents/ContextUsageBar.vue'

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
  projectPermissions?: any
  contextUsage?: any
  contextLoading?: boolean
}>()

const emit = defineEmits<{
  (e: 'refresh-context'): void
}>()

// Reactive data
const toolStats = ref({ count: 0, byName: {} as Record<string, number> })
const permissionStats = ref({ approved: 0, denied: 0, total: 0 })

// Collapsible sections state
const expandedSections = ref({
  sessionInfo: true,
  toolsPermissions: true
})

const toggleSection = (section: keyof typeof expandedSections.value) => {
  expandedSections.value[section] = !expandedSections.value[section]
}

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
</script>

<style scoped>
.session-metrics {
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* Collapsible Sections */
.collapsible-section {
  margin-bottom: 16px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  overflow: hidden;
  background: var(--bg-primary);
}

.section-header {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: var(--bg-primary);
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.95rem;
}

.section-header:hover {
  background: var(--bg-secondary);
}

.header-content {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.section-icon {
  font-size: 1.2rem;
}

.section-title {
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.toggle-arrow {
  flex-shrink: 0;
  transition: transform 0.3s ease;
  color: var(--text-secondary);
}

.toggle-arrow.collapsed {
  transform: rotate(-90deg);
}

.section-content {
  padding: 16px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-primary);
  animation: slideDown 0.3s ease;
}

@keyframes slideDown {
  from {
    opacity: 0;
    max-height: 0;
  }
  to {
    opacity: 1;
    max-height: 500px;
  }
}

/* Metrics Grid */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
  margin-bottom: 0;
}

.metric-card {
  display: flex;
  flex-direction: column;
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

.metric-content {
  flex: 1;
  width: 100%;
}

.metric-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.5px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.metric-label-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
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
  margin-top: 8px;
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

/* Environment Card */
.environment-metric {
  margin-top: 16px;
}

/* Environment Details (inside metric card) */
.environment-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 8px;
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

/* Project Permissions Card */
.project-permissions-metric {
  margin-bottom: 16px;
}

.project-permissions-metric .metric-content {
  width: 100%;
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
