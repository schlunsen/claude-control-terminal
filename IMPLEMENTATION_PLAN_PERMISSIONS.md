# Implementation Plan: Project Permissions Display

## Overview
Replace the "Auto-Approved Actions" section in the MetricsSidebar with a "Project Permissions" section that displays permissions from the current working directory's `.claude/settings.local.json` file.

## Current State
- The MetricsSidebar currently shows `AlwaysAllowRules` component which displays runtime session-based auto-approved actions
- We want to replace this with project-level configured permissions from settings.local.json

## Goals
1. Read `.claude/settings.local.json` from the current working directory
2. Parse and categorize the `permissions.allow` array
3. Display permissions in the MetricsSidebar grouped by category
4. Match the existing UI/UX patterns from AlwaysAllowRules

---

## Implementation Tasks

### 1. Backend: Create API Endpoint

**File**: `internal/server/server.go`

#### Add Handler Function
```go
// Handler: Get project permissions from .claude/settings.local.json
func (s *Server) handleGetProjectPermissions(c *fiber.Ctx) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get current working directory",
		})
	}

	// Build path to settings.local.json
	settingsPath := filepath.Join(cwd, ".claude", "settings.local.json")

	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return c.JSON(fiber.Map{
			"total": 0,
			"categories": fiber.Map{},
			"file_path": settingsPath,
			"error": "settings.local.json not found",
		})
	}

	// Read file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to read settings file: %v", err),
		})
	}

	// Parse JSON
	var settings struct {
		Permissions struct {
			Allow []string `json:"allow"`
		} `json:"permissions"`
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to parse settings file: %v", err),
		})
	}

	// Categorize permissions
	categories := categorizePermissions(settings.Permissions.Allow)

	return c.JSON(fiber.Map{
		"total": len(settings.Permissions.Allow),
		"categories": categories,
		"file_path": settingsPath,
	})
}

// Helper function to categorize permissions
func categorizePermissions(permissions []string) map[string]interface{} {
	categories := map[string][]string{
		"bash":     []string{},
		"read":     []string{},
		"write":    []string{},
		"edit":     []string{},
		"webfetch": []string{},
		"mcp":      []string{},
		"other":    []string{},
	}

	for _, perm := range permissions {
		switch {
		case strings.HasPrefix(perm, "Bash("):
			categories["bash"] = append(categories["bash"], perm)
		case strings.HasPrefix(perm, "Read("):
			categories["read"] = append(categories["read"], perm)
		case strings.HasPrefix(perm, "Write("):
			categories["write"] = append(categories["write"], perm)
		case strings.HasPrefix(perm, "Edit("):
			categories["edit"] = append(categories["edit"], perm)
		case strings.HasPrefix(perm, "WebFetch("):
			categories["webfetch"] = append(categories["webfetch"], perm)
		case strings.HasPrefix(perm, "mcp__"):
			categories["mcp"] = append(categories["mcp"], perm)
		default:
			categories["other"] = append(categories["other"], perm)
		}
	}

	// Build response with counts
	result := make(map[string]interface{})
	for category, perms := range categories {
		if len(perms) > 0 {
			result[category] = fiber.Map{
				"count":       len(perms),
				"permissions": perms,
			}
		}
	}

	return result
}
```

#### Register Route
In `setupRoutes()` function, add:
```go
// Config endpoints (for frontend to get API key securely)
api.Get("/config/api-key", s.handleGetAPIKey)
api.Get("/config/cwd", s.handleGetCWD)
api.Get("/config/permissions", s.handleGetProjectPermissions)  // ADD THIS LINE
```

**Location**: Around line 391-392 in server.go

---

### 2. Frontend: Create ProjectPermissions Component

**File**: `internal/server/frontend/app/components/agents/ProjectPermissions.vue`

```vue
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
  color: var(--color-text-primary);
  margin: 0;
}

.permissions-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.5rem;
  height: 1.5rem;
  padding: 0 0.5rem;
  background: var(--color-info);
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
  color: var(--color-text-secondary);
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
  background: var(--color-bg-secondary);
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
  color: var(--color-text-primary);
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
  color: var(--color-text-secondary);
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
  color: var(--color-text-primary);
  word-break: break-all;
  display: block;
}
</style>
```

---

### 3. Update MetricsSidebar Component

**File**: `internal/server/frontend/app/components/agents/MetricsSidebar.vue`

#### Changes:

1. **Remove AlwaysAllowRules import** (line 86):
```vue
// REMOVE THIS LINE:
import AlwaysAllowRules from '~/components/agents/AlwaysAllowRules.vue'
```

2. **Add ProjectPermissions import** (line 86):
```vue
// ADD THIS LINE:
import ProjectPermissions from '~/components/agents/ProjectPermissions.vue'
```

3. **Update Props interface** (lines 88-97):
```typescript
interface Props {
  show: boolean
  session: any
  projectPermissions?: any  // CHANGE: Replace alwaysAllowRules with projectPermissions
  messageCount: number
  toolExecutions: any
  permissionStats: any
  contextUsage?: any
  contextLoading?: boolean
}
```

4. **Remove emit definitions** (lines 101-105):
```typescript
// REMOVE THESE:
const emit = defineEmits<{
  (e: 'refresh-context'): void
  (e: 'remove-rule', ruleId: string): void    // REMOVE
  (e: 'clear-all-rules'): void                // REMOVE
}>()

// REPLACE WITH:
const emit = defineEmits<{
  (e: 'refresh-context'): void
}>()
```

5. **Replace AlwaysAllowRules section** (lines 70-77):
```vue
<!-- REMOVE THIS SECTION:
<div v-if="alwaysAllowRules" class="always-allow-section">
  <AlwaysAllowRules
    :rules="alwaysAllowRules"
    @remove="(ruleId) => $emit('remove-rule', ruleId)"
    @clear-all="$emit('clear-all-rules')"
  />
</div>
-->

<!-- ADD THIS SECTION: -->
<div v-if="projectPermissions" class="project-permissions-section">
  <ProjectPermissions :permissions="projectPermissions" />
</div>
```

6. **Update CSS** (line 440-444):
```css
/* RENAME CLASS: */
.project-permissions-section {  /* was: .always-allow-section */
  padding: 0.75rem;
  border-top: 1px solid var(--border-color);
  background: var(--color-bg-primary);
}
```

---

### 4. Update Agents Page

**File**: `internal/server/frontend/app/pages/agents.vue`

#### Add to script section:

1. **Add state variable** (near other refs):
```typescript
const projectPermissions = ref(null)
```

2. **Add fetch function**:
```typescript
const fetchProjectPermissions = async () => {
  try {
    const response = await fetchWithAuth('/api/config/permissions')
    if (response.ok) {
      projectPermissions.value = await response.json()
    } else {
      console.error('Failed to fetch project permissions')
      projectPermissions.value = null
    }
  } catch (error) {
    console.error('Error fetching project permissions:', error)
    projectPermissions.value = null
  }
}
```

3. **Call on mount** (in onMounted):
```typescript
onMounted(() => {
  // ... existing code ...
  fetchProjectPermissions()  // ADD THIS
})
```

4. **Update MetricsSidebar props**:
```vue
<MetricsSidebar
  :show="showMetricsSidebar"
  :session="currentSession"
  :project-permissions="projectPermissions"  <!-- ADD THIS -->
  :message-count="messages.length"
  :tool-executions="toolExecutions"
  :permission-stats="permissionStats"
  :context-usage="contextUsage"
  :context-loading="contextLoading"
  @refresh-context="refreshContextUsage"
  <!-- REMOVE: @remove-rule and @clear-all-rules -->
/>
```

---

## Expected Data Flow

1. **User opens Agents page** → agents.vue mounts
2. **agents.vue calls** `fetchProjectPermissions()`
3. **API call** to `GET /api/config/permissions`
4. **Backend reads** `.claude/settings.local.json` from cwd
5. **Backend parses** and categorizes permissions
6. **Response sent** to frontend with structure:
   ```json
   {
     "total": 58,
     "categories": {
       "bash": { "count": 35, "permissions": [...] },
       "read": { "count": 15, "permissions": [...] },
       ...
     },
     "file_path": "..."
   }
   ```
7. **Data passed** to MetricsSidebar as `projectPermissions` prop
8. **MetricsSidebar passes** to ProjectPermissions component
9. **ProjectPermissions displays** categorized permissions with collapsible sections

---

## Testing Checklist

- [ ] Backend endpoint returns correct data structure
- [ ] Backend handles missing file gracefully
- [ ] Backend handles invalid JSON gracefully
- [ ] Frontend displays all permission categories
- [ ] Categories are collapsible/expandable
- [ ] Count badges show correct numbers
- [ ] Empty state displays when no permissions
- [ ] Error state displays when file not found
- [ ] Styling matches existing MetricsSidebar design
- [ ] Component is responsive

---

## Files to Modify

1. ✅ `internal/server/server.go` - Add endpoint and handler
2. ✅ `internal/server/frontend/app/components/agents/ProjectPermissions.vue` - Create new component
3. ✅ `internal/server/frontend/app/components/agents/MetricsSidebar.vue` - Replace AlwaysAllowRules
4. ✅ `internal/server/frontend/app/pages/agents.vue` - Fetch and pass permissions

---

## Notes

- The AlwaysAllowRules component can be kept in the codebase for future use or removed
- This implementation assumes the current working directory contains `.claude/settings.local.json`
- Permissions are read-only (no edit/remove functionality needed)
- Consider adding a refresh button in the ProjectPermissions header if needed
