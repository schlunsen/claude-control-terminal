<template>
  <div class="project-permissions">
    <div class="permissions-header">
      <h3>Project Permissions</h3>
      <span v-if="totalCount > 0" class="permissions-count">{{ totalCount }}</span>
    </div>

    <div v-if="error" class="no-permissions">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"></path>
        <path d="M19 10v2a7 7 0 0 1-14 0v-2"></path>
        <line x1="12" y1="19" x2="12" y2="23"></line>
        <line x1="8" y1="23" x2="16" y2="23"></line>
      </svg>
      <p>No permissions file found</p>
      <p class="hint">Create .claude/settings.local.json to configure permissions</p>
    </div>

    <div v-else-if="totalCount === 0" class="no-permissions">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
        <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
      </svg>
      <p>No permissions configured</p>
      <p class="hint">Add permissions to settings.local.json</p>
    </div>

    <div v-else class="permissions-list">
      <!-- Bash Commands -->
      <div v-if="categories.bash" class="category-section">
        <button @click="toggleCategory('bash')" class="category-header">
          <span class="category-icon">🐚</span>
          <span class="category-title">Bash Commands</span>
          <span class="category-count">{{ categories.bash.count }}</span>
          <svg
            class="chevron"
            :class="{ 'expanded': expandedCategories.bash }"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
        <div v-show="expandedCategories.bash" class="category-items">
          <div v-for="(perm, index) in categories.bash.permissions" :key="index" class="permission-item">
            <code>{{ perm }}</code>
          </div>
        </div>
      </div>

      <!-- Read Operations -->
      <div v-if="categories.read" class="category-section">
        <button @click="toggleCategory('read')" class="category-header">
          <span class="category-icon">📖</span>
          <span class="category-title">Read Operations</span>
          <span class="category-count">{{ categories.read.count }}</span>
          <svg
            class="chevron"
            :class="{ 'expanded': expandedCategories.read }"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
        <div v-show="expandedCategories.read" class="category-items">
          <div v-for="(perm, index) in categories.read.permissions" :key="index" class="permission-item">
            <code>{{ perm }}</code>
          </div>
        </div>
      </div>

      <!-- Write Operations -->
      <div v-if="categories.write" class="category-section">
        <button @click="toggleCategory('write')" class="category-header">
          <span class="category-icon">✍️</span>
          <span class="category-title">Write Operations</span>
          <span class="category-count">{{ categories.write.count }}</span>
          <svg
            class="chevron"
            :class="{ 'expanded': expandedCategories.write }"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
        <div v-show="expandedCategories.write" class="category-items">
          <div v-for="(perm, index) in categories.write.permissions" :key="index" class="permission-item">
            <code>{{ perm }}</code>
          </div>
        </div>
      </div>

      <!-- Edit Operations -->
      <div v-if="categories.edit" class="category-section">
        <button @click="toggleCategory('edit')" class="category-header">
          <span class="category-icon">✏️</span>
          <span class="category-title">Edit Operations</span>
          <span class="category-count">{{ categories.edit.count }}</span>
          <svg
            class="chevron"
            :class="{ 'expanded': expandedCategories.edit }"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
        <div v-show="expandedCategories.edit" class="category-items">
          <div v-for="(perm, index) in categories.edit.permissions" :key="index" class="permission-item">
            <code>{{ perm }}</code>
          </div>
        </div>
      </div>

      <!-- WebFetch -->
      <div v-if="categories.webfetch" class="category-section">
        <button @click="toggleCategory('webfetch')" class="category-header">
          <span class="category-icon">🌐</span>
          <span class="category-title">Web Fetch</span>
          <span class="category-count">{{ categories.webfetch.count }}</span>
          <svg
            class="chevron"
            :class="{ 'expanded': expandedCategories.webfetch }"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
        <div v-show="expandedCategories.webfetch" class="category-items">
          <div v-for="(perm, index) in categories.webfetch.permissions" :key="index" class="permission-item">
            <code>{{ perm }}</code>
          </div>
        </div>
      </div>

      <!-- MCP Servers -->
      <div v-if="categories.mcp" class="category-section">
        <button @click="toggleCategory('mcp')" class="category-header">
          <span class="category-icon">🔌</span>
          <span class="category-title">MCP Servers</span>
          <span class="category-count">{{ categories.mcp.count }}</span>
          <svg
            class="chevron"
            :class="{ 'expanded': expandedCategories.mcp }"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
        <div v-show="expandedCategories.mcp" class="category-items">
          <div v-for="(perm, index) in categories.mcp.permissions" :key="index" class="permission-item">
            <code>{{ perm }}</code>
          </div>
        </div>
      </div>

      <!-- Other -->
      <div v-if="categories.other" class="category-section">
        <button @click="toggleCategory('other')" class="category-header">
          <span class="category-icon">⚙️</span>
          <span class="category-title">Other</span>
          <span class="category-count">{{ categories.other.count }}</span>
          <svg
            class="chevron"
            :class="{ 'expanded': expandedCategories.other }"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
        <div v-show="expandedCategories.other" class="category-items">
          <div v-for="(perm, index) in categories.other.permissions" :key="index" class="permission-item">
            <code>{{ perm }}</code>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface PermissionCategory {
  count: number
  permissions: string[]
}

interface Props {
  permissions: {
    total: number
    categories: Record<string, PermissionCategory>
    error?: string
  } | null
}

const props = defineProps<Props>()

const totalCount = computed(() => props.permissions?.total || 0)
const categories = computed(() => props.permissions?.categories || {})
const error = computed(() => props.permissions?.error)

// Expanded state for each category
const expandedCategories = ref<Record<string, boolean>>({
  bash: false,
  read: false,
  write: false,
  edit: false,
  webfetch: false,
  mcp: false,
  other: false,
})

const toggleCategory = (category: string) => {
  expandedCategories.value[category] = !expandedCategories.value[category]
}
</script>

<style scoped>
.project-permissions {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.permissions-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.permissions-header h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.permissions-count {
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

.no-permissions {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem;
  text-align: center;
  color: var(--text-secondary);
}

.no-permissions svg {
  opacity: 0.3;
  margin-bottom: 1rem;
}

.no-permissions p {
  margin: 0.25rem 0;
}

.no-permissions .hint {
  font-size: 0.875rem;
  opacity: 0.7;
}

.permissions-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.category-section {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  overflow: hidden;
}

.category-header {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.75rem;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.category-header:hover {
  background: rgba(139, 92, 246, 0.05);
}

.category-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.category-title {
  flex: 1;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
}

.category-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.5rem;
  height: 1.5rem;
  padding: 0 0.5rem;
  background: var(--accent-purple);
  color: white;
  border-radius: 0.75rem;
  font-size: 0.7rem;
  font-weight: 600;
}

.chevron {
  flex-shrink: 0;
  transition: transform 0.2s;
  color: var(--text-secondary);
}

.chevron.expanded {
  transform: rotate(180deg);
}

.category-items {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  padding: 0 0.75rem 0.75rem 0.75rem;
}

.permission-item {
  padding: 0.5rem;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 0.375rem;
  border-left: 2px solid var(--accent-purple);
}

.permission-item code {
  font-size: 0.75rem;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  color: var(--text-primary);
  word-break: break-all;
  display: block;
}
</style>
