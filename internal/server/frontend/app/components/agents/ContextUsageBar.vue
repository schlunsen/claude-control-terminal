<template>
  <div class="context-usage-bar" v-if="usage">
    <div class="usage-header">
      <div class="header-row">
        <span class="usage-icon">📊</span>
        <span class="usage-title">Stats</span>
      </div>
      <button
        class="refresh-button"
        @click="$emit('refresh')"
        :disabled="loading"
        title="Refresh context usage"
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          :class="{ spinning: loading }"
        >
          <polyline points="23 4 23 10 17 10"></polyline>
          <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
        </svg>
        <span class="refresh-text">Refresh context info</span>
      </button>
    </div>

    <div class="usage-content">
      <!-- Overall Usage Summary -->
      <div class="overall-summary">
        <div class="summary-row">
          <span class="summary-label">Total Used:</span>
          <span class="summary-value" :class="contextUsageClass">
            {{ formatTokens(usage.total_tokens) }} / {{ formatTokens(usage.context_window) }}
          </span>
        </div>
        <div class="summary-row">
          <span class="summary-label">Usage:</span>
          <span class="summary-value" :class="contextUsageClass">
            {{ usage.percentage }}%
          </span>
        </div>
        <div class="summary-row model-row">
          <span class="summary-label">Model:</span>
          <span class="summary-value model-name">
            {{ formatModelName(usage.model) }}
          </span>
        </div>
      </div>

      <!-- Category Breakdown -->
      <div class="categories-section">
        <h4 class="section-title">Categories</h4>

        <div class="category-list">
          <div
            v-for="category in sortedCategories"
            :key="category.name"
            class="category-item"
            :class="{ 'free-space': category.name === 'Free space' }"
          >
            <div class="category-row">
              <span class="category-icon">{{ getCategoryIcon(category.name) }}</span>
              <span class="category-name">{{ category.name }}</span>
              <span class="category-value">{{ formatTokens(category.tokens) }}</span>
              <span class="category-percentage">{{ category.percentage.toFixed(1) }}%</span>
            </div>
            <div class="category-bar">
              <div
                class="category-fill"
                :style="{
                  width: category.percentage + '%',
                  background: getCategoryColor(category.name)
                }"
                :title="`${category.name}: ${formatTokens(category.tokens)} (${category.percentage.toFixed(1)}%)`"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Warning if approaching limit -->
      <div class="usage-warning" v-if="usage.percentage >= 80">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
          <line x1="12" y1="9" x2="12" y2="13"></line>
          <line x1="12" y1="17" x2="12.01" y2="17"></line>
        </svg>
        <span v-if="usage.percentage >= 95">
          Context window nearly full! Consider starting a new conversation.
        </span>
        <span v-else>
          Context window filling up ({{ usage.percentage }}%).
          Consider summarizing or starting a new conversation soon.
        </span>
      </div>
    </div>
  </div>

  <!-- No usage data available -->
  <div class="context-usage-bar no-data" v-else>
    <div class="usage-header">
      <div class="header-row">
        <span class="usage-icon">📊</span>
        <span class="usage-title">Stats</span>
      </div>
      <button
        class="refresh-button"
        @click="$emit('refresh')"
        :disabled="loading"
        title="Fetch context usage"
      >
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          :class="{ spinning: loading }"
        >
          <polyline points="23 4 23 10 17 10"></polyline>
          <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
        </svg>
      </button>
    </div>
    <div class="no-data-message">
      <p>No usage data available yet.</p>
      <p class="hint">Click refresh to fetch context usage.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ContextUsageData } from '~/composables/agents/useContextUsage'

const props = defineProps<{
  usage: ContextUsageData | null
  loading?: boolean
}>()

defineEmits<{
  (e: 'refresh'): void
}>()

// Sort categories: Free space last, others by percentage descending
const sortedCategories = computed(() => {
  if (!props.usage?.categories) return []

  const categories = [...props.usage.categories]
  return categories.sort((a, b) => {
    // Free space always goes last
    if (a.name === 'Free space') return 1
    if (b.name === 'Free space') return -1
    // Sort others by percentage descending
    return b.percentage - a.percentage
  })
})

// Context usage class (for color coding)
const contextUsageClass = computed(() => {
  if (!props.usage) return 'normal'
  const percentage = props.usage.percentage
  if (percentage >= 95) return 'critical'
  if (percentage >= 80) return 'warning'
  if (percentage >= 60) return 'high'
  return 'normal'
})

// Get category icon
const getCategoryIcon = (categoryName: string): string => {
  const icons: Record<string, string> = {
    'System prompt': '⚙️',
    'System tools': '🔧',
    'MCP tools': '🔌',
    'Memory files': '📝',
    'Messages': '💬',
    'Free space': '✨'
  }
  return icons[categoryName] || '📦'
}

// Get category color
const getCategoryColor = (categoryName: string): string => {
  const colors: Record<string, string> = {
    'System prompt': 'linear-gradient(90deg, #667eea 0%, #764ba2 100%)',
    'System tools': 'linear-gradient(90deg, #f093fb 0%, #f5576c 100%)',
    'MCP tools': 'linear-gradient(90deg, #4facfe 0%, #00f2fe 100%)',
    'Memory files': 'linear-gradient(90deg, #43e97b 0%, #38f9d7 100%)',
    'Messages': 'linear-gradient(90deg, #fa709a 0%, #fee140 100%)',
    'Free space': 'linear-gradient(90deg, #a8edea 0%, #fed6e3 100%)'
  }
  return colors[categoryName] || 'linear-gradient(90deg, #ccc 0%, #eee 100%)'
}

// Format tokens with k suffix
const formatTokens = (tokens: number): string => {
  if (tokens >= 1000) {
    return `${(tokens / 1000).toFixed(1)}k`
  }
  return tokens.toString()
}

// Format model name (shorten if needed)
const formatModelName = (model: string): string => {
  // Shorten claude-sonnet-4-5-20250929 to claude-sonnet-4.5
  const match = model.match(/claude-(\w+)-(\d+)-(\d+)/)
  if (match) {
    return `claude-${match[1]}-${match[2]}.${match[3]}`
  }
  return model
}
</script>

<style scoped>
.context-usage-bar {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  overflow: hidden;
  transition: all 0.3s ease;
}

.context-usage-bar:hover {
  border-color: var(--accent-purple);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.15);
}

.context-usage-bar.no-data {
  background: var(--bg-secondary);
}

.usage-header {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-bg);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.refresh-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-secondary);
}

.refresh-text {
  font-size: 0.85rem;
  font-weight: 500;
  white-space: nowrap;
}

.refresh-button:hover:not(:disabled) {
  background: var(--accent-purple);
  border-color: var(--accent-purple);
  color: white;
  transform: translateY(-1px);
}

.refresh-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.refresh-button svg.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.usage-icon {
  font-size: 1.2rem;
}

.usage-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.usage-content {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Overall Summary */
.overall-summary {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
}

.summary-row.model-row {
  margin-top: 4px;
}

.summary-label {
  color: var(--text-secondary);
  font-weight: 500;
}

.summary-value {
  font-weight: 700;
  font-family: 'Monaco', 'Menlo', monospace;
  color: var(--text-primary);
}

.summary-value.model-name {
  font-size: 0.75rem;
  opacity: 0.8;
  font-weight: 500;
}

.summary-value.normal {
  color: #43e97b;
}

.summary-value.high {
  color: #feca57;
}

.summary-value.warning {
  color: #ff9ff3;
}

.summary-value.critical {
  color: #ee5a6f;
}

/* Categories Section */
.categories-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-title {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0;
}

.category-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.category-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.category-item.free-space {
  opacity: 0.6;
}

.category-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.8rem;
}

.category-icon {
  font-size: 1rem;
  flex-shrink: 0;
}

.category-name {
  flex: 1;
  color: var(--text-primary);
  font-weight: 500;
}

.category-value {
  color: var(--text-secondary);
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.75rem;
}

.category-percentage {
  color: var(--text-secondary);
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.75rem;
  min-width: 45px;
  text-align: right;
}

.category-bar {
  height: 6px;
  background: var(--bg-secondary);
  border-radius: 3px;
  overflow: hidden;
  position: relative;
}

.category-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
  position: relative;
}

.usage-warning {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  background: rgba(238, 90, 111, 0.1);
  border: 1px solid rgba(238, 90, 111, 0.3);
  border-radius: 8px;
  color: #ee5a6f;
  font-size: 0.85rem;
  line-height: 1.5;
}

.usage-warning svg {
  flex-shrink: 0;
  margin-top: 2px;
}

.no-data-message {
  padding: 32px 16px;
  text-align: center;
  color: var(--text-secondary);
}

.no-data-message p {
  margin: 0 0 8px 0;
  font-size: 0.9rem;
}

.no-data-message .hint {
  font-size: 0.8rem;
  opacity: 0.7;
}

/* Responsive */
@media (max-width: 768px) {
  .usage-header {
    padding: 12px;
  }

  .usage-content {
    padding: 12px;
    gap: 16px;
  }

  .usage-title {
    font-size: 0.85rem;
  }

  .category-bar {
    height: 5px;
  }
}
</style>
