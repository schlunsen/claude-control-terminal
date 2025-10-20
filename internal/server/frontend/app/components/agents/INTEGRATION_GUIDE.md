# Integration Guide: Using Refactored Components

This guide shows how to integrate the newly extracted components back into `agents.vue`.

## 1. Import Statements

Add these imports to the `<script setup>` section of `agents.vue`:

```typescript
// Existing imports
import { useAgentWebSocket } from '~/composables/useAgentWebSocket'
import SessionMetrics from '~/components/SessionMetrics.vue'
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import type { ActiveTool } from '~/types/agents'

// NEW: Import refactored components
import SessionItem from '~/components/agents/SessionItem.vue'
import SessionFilters from '~/components/agents/SessionFilters.vue'
import PermissionRequest from '~/components/agents/PermissionRequest.vue'
import ToolExecutionBar from '~/components/agents/ToolExecutionBar.vue'
import CreateSessionModal from '~/components/agents/CreateSessionModal.vue'
import ResumeSessionModal from '~/components/agents/ResumeSessionModal.vue'

// NEW: Import utilities
import { formatTime, formatMessage, formatRelativeTime, truncatePath } from '~/utils/agents/messageFormatters'
import { parseTodoWrite, formatTodosForTool } from '~/utils/agents/todoParser'
import { parseToolUse, getToolIcon } from '~/utils/agents/toolParser'
import { useMessageScroll } from '~/composables/agents/useMessageScroll'

// NEW: Import existing overlays (these were already separate)
import TodoWriteOverlay from '~/components/TodoWriteOverlay.vue'
import ToolOverlay from '~/components/ToolOverlay.vue'
```

## 2. Replace Session Filters

**BEFORE (lines 73-84):**
```vue
<div class="session-filters">
  <button
    v-for="filter in sessionFilters"
    :key="filter.value"
    @click="activeFilter = filter.value"
    class="filter-tab"
    :class="{ active: activeFilter === filter.value }"
  >
    {{ filter.label }}
    <span class="filter-count">{{ getFilterCount(filter.value) }}</span>
  </button>
</div>
```

**AFTER:**
```vue
<SessionFilters
  :active-filter="activeFilter"
  :filters="sessionFiltersWithCounts"
  @update:active-filter="activeFilter = $event"
/>
```

**Add computed property:**
```typescript
const sessionFiltersWithCounts = computed(() => [
  { label: 'Active', value: 'active', count: getFilterCount('active') },
  { label: 'All', value: 'all', count: getFilterCount('all') },
  { label: 'Ended', value: 'ended', count: getFilterCount('ended') }
])
```

## 3. Replace Session Items

**BEFORE (lines 90-141):**
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
  <!-- ... lots of markup ... -->
</div>
```

**AFTER:**
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

## 4. Replace Permission Requests

**BEFORE (lines 245-283):**
```vue
<div v-if="activeSessionPermissions.length > 0" class="permission-requests">
  <div
    v-for="permission in activeSessionPermissions"
    :key="permission.request_id"
    class="permission-request"
  >
    <!-- ... lots of markup ... -->
  </div>
</div>
```

**AFTER:**
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

## 5. Replace Tool Execution Bar

**BEFORE (lines 286-312):**
```vue
<div v-if="shouldShowToolBar" class="tool-execution-bar">
  <div class="tool-execution-content">
    <!-- ... lots of markup ... -->
  </div>
</div>
```

**AFTER:**
```vue
<ToolExecutionBar :tool-execution="activeSessionToolExecution" />
```

Note: `shouldShowToolBar` computed is no longer needed - the component handles null checking.

## 6. Replace Create Session Modal

**BEFORE (lines 350-551 - ~200 lines of modal markup):**

**AFTER:**
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

**Update `createSessionWithOptions` to accept formData:**
```typescript
const createSessionWithOptions = async (formData?: typeof sessionForm.value) => {
  if (!agentWs.connected) return

  const form = formData || sessionForm.value
  if (!form.workingDirectory) return

  creatingSession.value = true

  try {
    const sessionId = crypto.randomUUID()
    const selectedProvider = availableProviders.value.find(p => p.id === form.modelProvider)

    const options: any = {
      tools: form.tools,
      working_directory: form.workingDirectory,
      permission_mode: form.permissionMode,
      provider: form.modelProvider,
      model: form.model
    }

    if (selectedProvider?.base_url) {
      options.base_url = selectedProvider.base_url
    }

    if (form.promptMode === 'agent' && form.selectedAgent) {
      options.agent_name = form.selectedAgent
    } else {
      options.system_prompt = form.systemPrompt || 'You are a helpful AI assistant.'
    }

    agentWs.send({
      type: 'create_session',
      session_id: sessionId,
      options
    })

    showCreateSessionModal.value = false
  } catch (error) {
    console.error('Failed to create session:', error)
    alert('Failed to create session. Please try again.')
  } finally {
    creatingSession.value = false
  }
}
```

## 7. Replace Resume Session Modal

**BEFORE (lines 553-693 - ~140 lines of modal markup):**

**AFTER:**
```vue
<ResumeSessionModal
  :show="showResumeModal"
  :sessions="availableSessions"
  :selected-session="selectedResumeSession"
  :form-data="resumeForm"
  :loading="loadingSessions"
  :resuming="resumingSession"
  @close="showResumeModal = false; selectedResumeSession = null"
  @select-session="selectSessionForResume"
  @back="selectedResumeSession = null"
  @resume="resumeSessionWithOptions"
/>
```

**Update `resumeSessionWithOptions` to accept formData:**
```typescript
const resumeSessionWithOptions = async (formData?: typeof resumeForm.value) => {
  try {
    if (!selectedResumeSession.value) return

    resumingSession.value = true
    const form = formData || resumeForm.value

    const resumeData = await $fetch(`/api/sessions/${selectedResumeSession.value.conversation_id}/resume-data`)
    const sessionId = crypto.randomUUID()

    agentWs.send({
      type: 'create_session',
      session_id: sessionId,
      options: {
        tools: form.tools,
        system_prompt: form.systemPrompt || 'You are a helpful AI assistant.',
        working_directory: form.workingDirectory || resumeData.working_directory,
        permission_mode: form.permissionMode,
        conversation_history: resumeData.context,
        original_conversation_id: resumeData.conversation_id
      }
    })

    showResumeModal.value = false
    selectedResumeSession.value = null

    if (resumeData.messages && resumeData.messages.length > 0) {
      messages.value[sessionId] = []
      resumeData.messages.forEach(msg => {
        messages.value[sessionId].push({
          id: crypto.randomUUID(),
          role: 'user',
          content: msg.message,
          timestamp: new Date(msg.submitted_at),
          isHistorical: true
        })
      })
    }
  } catch (error) {
    console.error('Failed to resume session:', error)
    alert('Failed to resume session. Please try again.')
  } finally {
    resumingSession.value = false
  }
}
```

## 8. Replace Utility Functions

**Remove these functions from agents.vue** (they're now in utilities):

```typescript
// DELETE - now in utils/agents/messageFormatters.ts
const formatTime = (timestamp) => { ... }
const formatMessage = (content) => { ... }
const formatRelativeTime = (timestamp) => { ... }
const truncatePath = (text, maxLength) => { ... }

// DELETE - now in utils/agents/todoParser.ts
const parseTodoWrite = (content: string) => { ... }
const formatTodosForTool = (todos: TodoItem[]) => { ... }

// DELETE - now in utils/agents/toolParser.ts
const parseToolUse = (content: string) => { ... }
```

## 9. Replace Message Scroll Logic

**BEFORE:**
```typescript
const isUserNearBottom = ref(true)

const handleScroll = () => {
  if (!messagesContainer.value) return
  const { scrollTop, scrollHeight, clientHeight } = messagesContainer.value
  const threshold = 100
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
```

**AFTER:**
```typescript
const { isUserNearBottom, handleScroll, scrollToBottom, autoScrollIfNearBottom } = useMessageScroll()

// Update handleScroll calls to pass container
const handleScrollEvent = () => handleScroll(messagesContainer.value)

// Update scrollToBottom calls
scrollToBottom(messagesContainer.value, false)
autoScrollIfNearBottom(messagesContainer.value, true)
```

## 10. Remove Duplicate CSS

After integrating components, remove the following CSS from `agents.vue`:

```css
/* DELETE - now in SessionItem.vue */
.session-item { ... }
.session-status-dot { ... }
.session-avatar { ... }
/* ... all session item styles ... */

/* DELETE - now in SessionFilters.vue */
.session-filters { ... }
.filter-tab { ... }
.filter-count { ... }

/* DELETE - now in PermissionRequest.vue */
.permission-request { ... }
.permission-header { ... }
/* ... all permission styles ... */

/* DELETE - now in ToolExecutionBar.vue */
.tool-execution-bar { ... }
.tool-execution-content { ... }
/* ... all tool execution styles ... */

/* DELETE - now in CreateSessionModal.vue */
.modal-overlay { ... }
.modal-content { ... }
.modal-header { ... }
.modal-body { ... }
.modal-actions { ... }
.agents-grid { ... }
.agent-card { ... }
.prompt-mode-toggle { ... }
/* ... all modal styles ... */

/* DELETE - now in ResumeSessionModal.vue */
.sessions-list-modal { ... }
.session-card-modal { ... }
.resume-session-options { ... }
/* ... all resume modal styles ... */
```

## 11. Testing Checklist

After integration, test these scenarios:

- [ ] Session list displays correctly with filters
- [ ] Session items show avatars and status
- [ ] Clicking session items switches active session
- [ ] End/delete session buttons work
- [ ] Create session modal opens and closes
- [ ] Agent selection in create modal works
- [ ] Custom prompt mode in create modal works
- [ ] Tool selection checkboxes work
- [ ] Resume session modal opens and closes
- [ ] Session selection in resume modal works
- [ ] Resume options form works
- [ ] Permission requests appear and can be approved/denied
- [ ] Tool execution bar shows during tool use
- [ ] Message scrolling works correctly
- [ ] All styling matches previous appearance
- [ ] Dark/light theme support works

## 12. Incremental Integration Strategy

You don't have to integrate everything at once. Here's a safe order:

### Phase 1: Utilities (Low Risk)
1. Import utility functions
2. Remove duplicate function definitions
3. Update function calls to use imports
4. Test - should be invisible to users

### Phase 2: Simple Components (Medium Risk)
1. Integrate `SessionFilters`
2. Integrate `PermissionRequest`
3. Integrate `ToolExecutionBar`
4. Test each one individually

### Phase 3: Session Items (Medium Risk)
1. Integrate `SessionItem`
2. Test session list interaction
3. Remove old markup and styles

### Phase 4: Modals (Higher Risk - Test Thoroughly)
1. Integrate `CreateSessionModal`
2. Test session creation thoroughly
3. Integrate `ResumeSessionModal`
4. Test session resumption thoroughly
5. Remove old modal markup and styles

## 13. Rollback Plan

If issues occur, you can easily rollback:

1. **Keep the original agents.vue** backed up
2. Each component is **self-contained** - you can remove one without affecting others
3. The old code remains in agents.vue until you delete it

## Benefits After Integration

- **agents.vue**: 4,124 lines → ~1,900 lines (53% reduction)
- **Modular**: Each piece can be edited independently
- **Testable**: Components can be tested in isolation
- **Reusable**: Components can be used in other pages
- **Maintainable**: Clear separation of concerns

## Need Help?

Check the component source files for:
- Props interface definitions
- Emit event signatures
- Usage examples in comments
- TypeScript types
