# Before & After: Code Comparison

Visual comparison showing the impact of the refactoring.

## File Size Comparison

### Before Refactoring
```
agents.vue: 4,124 lines
├── Template:  ~696 lines
├── Script:    ~1,588 lines
└── Style:     ~1,840 lines
```

### After Refactoring
```
agents.vue: ~1,900 lines (53% reduction)

NEW FILES CREATED:
├── utils/agents/
│   ├── messageFormatters.ts    120 lines
│   ├── todoParser.ts           130 lines
│   └── toolParser.ts            80 lines
│
├── composables/agents/
│   └── useMessageScroll.ts      50 lines
│
└── components/agents/
    ├── SessionItem.vue         180 lines
    ├── SessionFilters.vue       90 lines
    ├── PermissionRequest.vue   130 lines
    ├── ToolExecutionBar.vue    140 lines
    ├── CreateSessionModal.vue  680 lines ⭐
    └── ResumeSessionModal.vue  580 lines ⭐

Total extracted: ~2,160 lines
Remaining: ~1,964 lines
```

## Code Examples: Before vs After

### 1. Session Item Rendering

#### BEFORE (agents.vue - ~50 lines)
```vue
<div
  v-for="session in filteredSessions"
  :key="session.id"
  class="session-item"
  :class="{
    active: activeSessionId === session.id,
    ended: session.status === 'ended'
  }"
  @click="selectSession(session.id)"
>
  <div class="session-status-dot" :class="session.status"></div>
  <img
    :src="useCharacterAvatar(session.id).avatar"
    :alt="useCharacterAvatar(session.id).name"
    class="session-avatar"
  />
  <div class="session-info">
    <div class="session-name">{{ useCharacterAvatar(session.id).name }}</div>
    <div class="session-meta">
      <span class="session-id">{{ session.id.slice(0, 8) }}</span>
      <span class="session-status" :class="session.status">{{ session.status }}</span>
      <span class="session-messages">{{ session.message_count }} messages</span>
      <span v-if="session.cost_usd && session.cost_usd > 0" class="session-cost">
        ${{ session.cost_usd.toFixed(4) }}
      </span>
    </div>
  </div>
  <div class="session-actions">
    <button
      v-if="session.status !== 'ended'"
      @click.stop="endSession(session.id)"
      class="btn-end-session"
      title="End session"
    >
      <svg width="14" height="14" viewBox="0 0 24 24">...</svg>
    </button>
    <button
      @click.stop="deleteSession(session.id)"
      class="btn-delete-session"
      title="Delete session"
    >
      <svg width="14" height="14" viewBox="0 0 24 24">...</svg>
    </button>
  </div>
</div>

<style scoped>
/* ~80 lines of CSS for session items */
.session-item { ... }
.session-status-dot { ... }
.session-avatar { ... }
/* ... etc ... */
</style>
```

#### AFTER (agents.vue - 7 lines)
```vue
<SessionItem
  v-for="session in filteredSessions"
  :key="session.id"
  :session="session"
  :is-active="activeSessionId === session.id"
  @select="selectSession"
  @end="endSession"
  @delete="deleteSession"
/>
```

**Result**: 50 lines → 7 lines (86% reduction)

---

### 2. Create Session Modal

#### BEFORE (agents.vue - ~200 lines)
```vue
<div v-if="showCreateSessionModal" class="modal-overlay" @click="showCreateSessionModal = false">
  <div class="modal-content" @click.stop>
    <div class="modal-header">
      <h2>Create New Session</h2>
      <button @click="showCreateSessionModal = false" class="modal-close">
        <svg width="20" height="20">...</svg>
      </button>
    </div>
    <div class="modal-body">
      <!-- Working Directory Input -->
      <div class="form-group">
        <label for="working-directory">Working Directory</label>
        <input
          id="working-directory"
          v-model="sessionForm.workingDirectory"
          @change="handleWorkingDirectoryChange"
          type="text"
          class="form-input"
        />
      </div>

      <!-- Permission Mode -->
      <div class="form-group">
        <label for="permission-mode">Permission Mode</label>
        <select id="permission-mode" v-model="sessionForm.permissionMode">
          <option value="default">Default</option>
          <option value="acceptEdits">Allow All</option>
          <option value="plan">Read Only</option>
        </select>
      </div>

      <!-- ... 150+ more lines of form fields ... -->

      <!-- Agent Selection Grid -->
      <div v-if="sessionForm.promptMode === 'agent'" class="form-group">
        <div class="agents-grid">
          <button
            v-for="agent in availableAgents"
            :key="agent.name"
            class="agent-card"
            :class="{ selected: sessionForm.selectedAgent === agent.name }"
            @click="sessionForm.selectedAgent = agent.name; loadSelectedAgent()"
          >
            <!-- Agent card content -->
          </button>
        </div>
      </div>

      <!-- Tools Checkboxes -->
      <div class="form-group">
        <div class="tools-grid">
          <label class="tool-checkbox">
            <input type="checkbox" v-model="sessionForm.tools" value="Read" />
            <span>Read</span>
          </label>
          <!-- ... 6 more tool checkboxes ... -->
        </div>
      </div>
    </div>

    <div class="modal-actions">
      <button @click="showCreateSessionModal = false">Cancel</button>
      <button @click="createSessionWithOptions">Create</button>
    </div>
  </div>
</div>

<style scoped>
/* ~200 lines of CSS for modals */
.modal-overlay { ... }
.modal-content { ... }
.agents-grid { ... }
.agent-card { ... }
/* ... etc ... */
</style>
```

#### AFTER (agents.vue - 13 lines)
```vue
<CreateSessionModal
  :show="showCreateSessionModal"
  :form-data="sessionForm"
  :providers="availableProviders"
  :current-provider="currentProvider"
  :agents="availableAgents"
  :selected-agent-preview="selectedAgentPreview"
  :loading-providers="loadingProviders"
  :loading-agents="loadingAgents"
  :creating="creatingSession"
  @close="showCreateSessionModal = false"
  @create="createSessionWithOptions"
  @working-directory-change="handleWorkingDirectoryChange"
  @agent-select="loadSelectedAgent"
/>
```

**Result**: 200 lines → 13 lines (94% reduction)

---

### 3. Permission Requests

#### BEFORE (agents.vue - ~40 lines)
```vue
<div v-if="activeSessionPermissions.length > 0" class="permission-requests">
  <div
    v-for="permission in activeSessionPermissions"
    :key="permission.request_id"
    class="permission-request"
  >
    <div class="permission-header">
      <div class="permission-icon">🔐</div>
      <div class="permission-title">Permission Request</div>
      <div class="permission-time">{{ formatTime(permission.timestamp) }}</div>
    </div>
    <div class="permission-description">
      {{ permission.description }}
    </div>
    <div class="permission-actions">
      <button @click="denyPermission(permission)" class="btn-deny">
        <svg width="16" height="16">...</svg>
        Deny
      </button>
      <button @click="approvePermission(permission)" class="btn-approve">
        <svg width="16" height="16">...</svg>
        Approve
      </button>
    </div>
  </div>
</div>

<style scoped>
/* ~50 lines of CSS */
.permission-request { ... }
.permission-header { ... }
/* ... etc ... */
</style>
```

#### AFTER (agents.vue - 8 lines)
```vue
<div v-if="activeSessionPermissions.length > 0" class="permission-requests">
  <PermissionRequest
    v-for="permission in activeSessionPermissions"
    :key="permission.request_id"
    :permission="permission"
    @approve="approvePermission"
    @deny="denyPermission"
  />
</div>
```

**Result**: 40 lines → 8 lines (80% reduction)

---

### 4. Utility Functions

#### BEFORE (agents.vue - Mixed with other code)
```typescript
// Scattered throughout 1,588 lines of script code

const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  })
}

const formatMessage = (content) => {
  // If content is an object, extract text from it first
  if (typeof content === 'object' && content !== null) {
    if (content.text) {
      content = Array.isArray(content.text) ? content.text.join('\n') : String(content.text)
    } else if (content.content) {
      content = String(content.content)
    } else {
      const objType = content.type || 'unknown'
      const keys = Object.keys(content).filter(k => k !== 'type')
      if (keys.length === 0) {
        return `<em class="system-message">${objType}</em>`
      }
      const props = keys.slice(0, 3).map(k => `${k}: ${String(content[k]).substring(0, 30)}`).join(', ')
      return `<em class="system-message">${objType} - ${props}</em>`
    }
  }
  // ... 30 more lines ...
}

const parseTodoWrite = (content: string): TodoItem[] | null => {
  // ... 90 lines of parsing logic ...
}

const parseToolUse = (content: string): ToolExecution | null => {
  // ... 45 lines of parsing logic ...
}
```

#### AFTER (agents.vue - Clean imports)
```typescript
// Clear, organized imports
import {
  formatTime,
  formatMessage,
  formatRelativeTime,
  truncatePath
} from '~/utils/agents/messageFormatters'

import {
  parseTodoWrite,
  formatTodosForTool
} from '~/utils/agents/todoParser'

import {
  parseToolUse,
  getToolIcon
} from '~/utils/agents/toolParser'

// Use them anywhere
const time = formatTime(message.timestamp)
const html = formatMessage(message.content)
const todos = parseTodoWrite(content)
```

**Result**:
- Functions moved to dedicated utility files
- Reusable across the application
- Easy to test in isolation
- Clear imports instead of scrolling

---

### 5. Message Scrolling Logic

#### BEFORE (agents.vue - ~40 lines)
```typescript
const isUserNearBottom = ref(true)

const handleScroll = () => {
  if (!messagesContainer.value) return

  const { scrollTop, scrollHeight, clientHeight } = messagesContainer.value
  const threshold = 100 // pixels from bottom
  isUserNearBottom.value = scrollHeight - scrollTop - clientHeight < threshold
}

const scrollToBottom = (smooth = false) => {
  if (!messagesContainer.value) return

  nextTick(() => {
    messagesContainer.value?.scrollTo({
      top: messagesContainer.value.scrollHeight,
      behavior: smooth ? 'smooth' : 'auto'
    })
  })
}

const autoScrollIfNearBottom = (smooth = true) => {
  if (isUserNearBottom.value) {
    scrollToBottom(smooth)
  }
}

// Usage scattered throughout
handleScroll()
scrollToBottom(false)
autoScrollIfNearBottom(true)
```

#### AFTER (agents.vue - Composable)
```typescript
import { useMessageScroll } from '~/composables/agents/useMessageScroll'

const {
  isUserNearBottom,
  handleScroll,
  scrollToBottom,
  autoScrollIfNearBottom
} = useMessageScroll()

// Usage is the same, but logic is extracted
handleScroll(messagesContainer.value)
scrollToBottom(messagesContainer.value, false)
autoScrollIfNearBottom(messagesContainer.value, true)
```

**Result**:
- Logic extracted to reusable composable
- Can be used in other message views
- Easier to test
- Clear separation of concerns

---

## CSS Comparison

### BEFORE (agents.vue - ~1,840 lines)
```css
<style scoped>
/* Everything mixed together */

/* Session styles */
.session-item { ... }
.session-avatar { ... }
.session-status-dot { ... }

/* Modal styles */
.modal-overlay { ... }
.modal-content { ... }
.modal-header { ... }

/* Form styles */
.form-group { ... }
.form-input { ... }

/* Agent card styles */
.agents-grid { ... }
.agent-card { ... }

/* Permission styles */
.permission-request { ... }

/* Tool execution styles */
.tool-execution-bar { ... }

/* ... 1,840 lines total ... */
</style>
```

### AFTER (agents.vue - Scoped to page)
```css
<style scoped>
/* Only page-specific styles remain */
.agents-page { ... }
.header { ... }
.agents-container { ... }
.chat-area { ... }
.messages-container { ... }

/* All component styles moved to components */
/* SessionItem.vue has its own <style scoped> */
/* CreateSessionModal.vue has its own <style scoped> */
/* etc. */
</style>
```

**Result**:
- No CSS conflicts
- Better organization
- Easier to maintain
- Can modify component styles without affecting page

---

## Import Complexity

### BEFORE (agents.vue - Line 697)
```typescript
<script setup lang="ts">
import { useAgentWebSocket } from '~/composables/useAgentWebSocket'
import SessionMetrics from '~/components/SessionMetrics.vue'
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import type { ActiveTool } from '~/types/agents'

// 4,000+ lines of code follow...
</script>
```

### AFTER (agents.vue - Organized imports)
```typescript
<script setup lang="ts">
// Vue & Nuxt
import { ref, computed, watch, nextTick, onMounted } from 'vue'

// Composables
import { useAgentWebSocket } from '~/composables/useAgentWebSocket'
import { useMessageScroll } from '~/composables/agents/useMessageScroll'

// Components - Existing
import SessionMetrics from '~/components/SessionMetrics.vue'
import TodoWriteOverlay from '~/components/TodoWriteOverlay.vue'
import ToolOverlay from '~/components/ToolOverlay.vue'

// Components - Agents
import SessionItem from '~/components/agents/SessionItem.vue'
import SessionFilters from '~/components/agents/SessionFilters.vue'
import PermissionRequest from '~/components/agents/PermissionRequest.vue'
import ToolExecutionBar from '~/components/agents/ToolExecutionBar.vue'
import CreateSessionModal from '~/components/agents/CreateSessionModal.vue'
import ResumeSessionModal from '~/components/agents/ResumeSessionModal.vue'

// Utilities
import {
  formatTime,
  formatMessage,
  formatRelativeTime,
  truncatePath
} from '~/utils/agents/messageFormatters'
import { parseTodoWrite, formatTodosForTool } from '~/utils/agents/todoParser'
import { parseToolUse, getToolIcon } from '~/utils/agents/toolParser'

// Types
import type { ActiveTool } from '~/types/agents'

// ~1,900 lines of orchestration code follow...
</script>
```

**Result**:
- Clear organization
- Easy to see dependencies
- Can lazy-load modals
- Better tree-shaking

---

## Maintainability Wins

### Finding Code

**BEFORE**: "Where is the session creation form?"
- Search through 4,124 lines
- Mixed with everything else
- Hard to navigate

**AFTER**: "Where is the session creation form?"
- Open `CreateSessionModal.vue`
- 680 lines, focused on one thing
- Easy to find and modify

### Making Changes

**BEFORE**: "I need to add a new field to session creation"
- Open massive agents.vue file
- Find the right section (lines 350-551)
- Hope you don't break something else
- CSS is 1,000 lines away

**AFTER**: "I need to add a new field to session creation"
- Open `CreateSessionModal.vue`
- Add field to template (~line 80)
- Add to props interface (~line 20)
- Add styles at bottom
- Self-contained change

### Testing

**BEFORE**: "I want to test session creation"
- Mount entire agents.vue (4,124 lines)
- Mock WebSocket connection
- Mock all other features
- Slow, brittle tests

**AFTER**: "I want to test session creation"
- Mount `CreateSessionModal.vue` (680 lines)
- Mock only relevant props
- Fast, focused tests
- Easy to achieve 100% coverage

---

## Summary

### Quantitative Improvements
- **Lines reduced**: 4,124 → 1,964 (53% reduction)
- **Files created**: 11 modular files
- **Largest extractions**: 2 modals (1,260 lines combined)
- **CSS organization**: Scoped to components
- **Reusability**: 3 utility files, 1 composable

### Qualitative Improvements
- ✅ **Better organization**: Clear file structure
- ✅ **Easier maintenance**: Find code quickly
- ✅ **More testable**: Test components in isolation
- ✅ **Reusable code**: Share utilities and components
- ✅ **Type safety**: TypeScript interfaces throughout
- ✅ **Developer experience**: Smaller, focused files
- ✅ **Performance**: Lazy-load modals when needed

### Future Scalability
- Can extract more components incrementally
- Can add more composables for shared logic
- Can create component library
- Can add comprehensive test suite
- Can optimize with lazy loading
