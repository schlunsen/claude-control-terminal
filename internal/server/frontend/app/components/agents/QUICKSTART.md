# Quick Start: Integrating Refactored Components

Fast-track guide to integrating the new components into `agents.vue`.

## ⚡ 5-Minute Quick Integration

### 1. Add Imports (30 seconds)

At the top of `agents.vue` `<script setup>` section, add:

```typescript
// NEW: Components
import SessionItem from '~/components/agents/SessionItem.vue'
import SessionFilters from '~/components/agents/SessionFilters.vue'
import PermissionRequest from '~/components/agents/PermissionRequest.vue'
import ToolExecutionBar from '~/components/agents/ToolExecutionBar.vue'
import CreateSessionModal from '~/components/agents/CreateSessionModal.vue'
import ResumeSessionModal from '~/components/agents/ResumeSessionModal.vue'

// NEW: Utilities
import { formatTime, formatMessage, formatRelativeTime, truncatePath } from '~/utils/agents/messageFormatters'
import { parseTodoWrite, formatTodosForTool } from '~/utils/agents/todoParser'
import { parseToolUse, getToolIcon } from '~/utils/agents/toolParser'
import { useMessageScroll } from '~/composables/agents/useMessageScroll'
```

### 2. Replace Session Filters (1 minute)

**Find (around line 73):**
```vue
<div class="session-filters">
  <button v-for="filter in sessionFilters" ...>
```

**Replace with:**
```vue
<SessionFilters
  :active-filter="activeFilter"
  :filters="sessionFiltersWithCounts"
  @update:active-filter="activeFilter = $event"
/>
```

**Add computed property (in script):**
```typescript
const sessionFiltersWithCounts = computed(() => [
  { label: 'Active', value: 'active', count: getFilterCount('active') },
  { label: 'All', value: 'all', count: getFilterCount('all') },
  { label: 'Ended', value: 'ended', count: getFilterCount('ended') }
])
```

### 3. Replace Session Items (1 minute)

**Find (around line 90):**
```vue
<div v-for="session in filteredSessions" class="session-item" ...>
  <!-- 50 lines of markup -->
</div>
```

**Replace with:**
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

### 4. Replace Permission Requests (30 seconds)

**Find (around line 245):**
```vue
<div v-for="permission in activeSessionPermissions" class="permission-request" ...>
  <!-- 38 lines -->
</div>
```

**Replace with:**
```vue
<PermissionRequest
  v-for="permission in activeSessionPermissions"
  :key="permission.request_id"
  :permission="permission"
  @approve="approvePermission"
  @deny="denyPermission"
/>
```

### 5. Replace Tool Execution Bar (30 seconds)

**Find (around line 286):**
```vue
<div v-if="shouldShowToolBar" class="tool-execution-bar">
  <!-- 26 lines -->
</div>
```

**Replace with:**
```vue
<ToolExecutionBar :tool-execution="activeSessionToolExecution" />
```

### 6. Replace Create Session Modal (1 minute)

**Find (around line 350):**
```vue
<div v-if="showCreateSessionModal" class="modal-overlay" ...>
  <!-- 200 lines -->
</div>
```

**Replace with:**
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

### 7. Replace Resume Session Modal (1 minute)

**Find (around line 553):**
```vue
<div v-if="showResumeModal" class="modal-overlay" ...>
  <!-- 140 lines -->
</div>
```

**Replace with:**
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

### 8. Remove Duplicate Functions (Optional - 5 minutes)

Search and delete these functions (they're now in utilities):

```typescript
// DELETE (now in utils/agents/messageFormatters.ts)
const formatTime = ...
const formatMessage = ...
const formatRelativeTime = ...

// DELETE (now in utils/agents/todoParser.ts)
const parseTodoWrite = ...
const formatTodosForTool = ...

// DELETE (now in utils/agents/toolParser.ts)
const parseToolUse = ...
```

### 9. Test (2 minutes)

```bash
# Start dev server
cd internal/server/frontend
npm run dev

# Open in browser
open http://localhost:3001
```

**Quick checks:**
- ✅ Sessions list loads
- ✅ Can filter sessions (Active/All/Ended)
- ✅ Can click session to view
- ✅ Can open Create Session modal
- ✅ Can open Resume Session modal
- ✅ Permissions appear and work
- ✅ Styling looks correct

---

## 🚀 Advanced Integration (Optional)

### Remove Duplicate CSS

After confirming everything works, remove these CSS sections from `agents.vue`:

```css
/* DELETE - now in SessionItem.vue */
.session-item { }
.session-status-dot { }
.session-avatar { }

/* DELETE - now in SessionFilters.vue */
.session-filters { }
.filter-tab { }

/* DELETE - now in PermissionRequest.vue */
.permission-request { }
.permission-header { }

/* DELETE - now in ToolExecutionBar.vue */
.tool-execution-bar { }

/* DELETE - now in CreateSessionModal.vue */
.modal-overlay { }
.modal-content { }
.agents-grid { }
.agent-card { }

/* DELETE - now in ResumeSessionModal.vue */
.sessions-list-modal { }
.session-card-modal { }
```

**How to safely remove:**
1. Comment out a section
2. Test the page
3. If it looks good, delete
4. Repeat for each section

---

## 📋 Verification Checklist

Use this checklist after integration:

### Visual Checks
- [ ] Sessions list appears correctly
- [ ] Filter tabs work (Active/All/Ended)
- [ ] Session avatars display
- [ ] Active session is highlighted
- [ ] End/Delete buttons appear on hover
- [ ] Create Session modal opens
- [ ] Resume Session modal opens
- [ ] Permission requests display
- [ ] Tool execution bar shows during tool use
- [ ] Dark/light theme works

### Functional Checks
- [ ] Can select a session
- [ ] Can end a session
- [ ] Can delete a session
- [ ] Can delete all sessions
- [ ] Can kill all agents
- [ ] Can create new session (agent mode)
- [ ] Can create new session (custom prompt mode)
- [ ] Can select tools in modals
- [ ] Can resume a session
- [ ] Can approve/deny permissions
- [ ] Messages display correctly
- [ ] Auto-scroll works

### Performance Checks
- [ ] Page loads quickly
- [ ] No console errors
- [ ] No console warnings
- [ ] Smooth animations
- [ ] Responsive on mobile

---

## 🐛 Troubleshooting

### Problem: Components not found

**Error:**
```
Cannot find module '~/components/agents/SessionItem.vue'
```

**Solution:**
Ensure files are in correct location:
```bash
ls -la internal/server/frontend/app/components/agents/
```

Should show:
- SessionItem.vue
- SessionFilters.vue
- PermissionRequest.vue
- ToolExecutionBar.vue
- CreateSessionModal.vue
- ResumeSessionModal.vue

---

### Problem: Utilities not found

**Error:**
```
Cannot find module '~/utils/agents/messageFormatters'
```

**Solution:**
Check utils directory:
```bash
ls -la internal/server/frontend/app/utils/agents/
```

Should show:
- messageFormatters.ts
- todoParser.ts
- toolParser.ts

---

### Problem: Styling looks different

**Cause:** CSS variables might not be defined

**Solution:**
Ensure these CSS variables exist in your theme:
```css
:root {
  --card-bg
  --bg-primary
  --bg-secondary
  --bg-tertiary
  --border-color
  --text-primary
  --text-secondary
  --text-tertiary
  --accent-purple
  --color-success
  --color-warning
  --color-error
}
```

---

### Problem: TypeScript errors

**Error:**
```
Type 'X' is not assignable to type 'Y'
```

**Solution:**
Check prop types match. Example:
```typescript
// Component expects
interface Session {
  id: string
  status: string
  message_count: number
}

// You're passing
const session = {
  id: 123,  // ❌ Should be string
  status: 'active',
  message_count: 5
}
```

---

### Problem: Events not firing

**Cause:** Event names might not match

**Solution:**
Check emit names:
```vue
<!-- Component emits -->
@update:active-filter

<!-- You're listening for -->
@update:activeFilter  <!-- ❌ Wrong case -->
@update:active-filter <!-- ✅ Correct -->
```

---

### Problem: Props are undefined

**Cause:** Not passing required props

**Solution:**
Check component requirements:
```vue
<!-- SessionItem requires these props -->
<SessionItem
  :session="session"     <!-- ✅ Required -->
  :is-active="isActive"  <!-- ✅ Required -->
/>
```

---

## 📚 Next Steps

After successful integration:

1. **Add Tests**
   - Unit tests for utilities
   - Component tests for each component
   - E2E tests for workflows

2. **Optimize Performance**
   - Lazy-load modals
   - Virtual scrolling for large lists
   - Debounce search/filter

3. **Enhance Features**
   - Add animations
   - Improve accessibility
   - Add keyboard shortcuts

4. **Extract More**
   - SessionSidebar component
   - MessagesList component
   - ChatArea component

---

## 🎯 Success Metrics

You've successfully integrated when:

✅ File size reduced by ~50%
✅ No visual changes (looks identical)
✅ All functionality works
✅ No console errors
✅ Tests pass (if you have tests)
✅ Performance is same or better

---

## 💡 Tips

- **Incremental**: Integrate one component at a time
- **Test often**: Test after each component
- **Keep backup**: Keep original agents.vue backed up
- **Use git**: Commit after each successful integration
- **Ask for help**: Check documentation if stuck

---

## 📖 Documentation

- **INTEGRATION_GUIDE.md** - Detailed integration steps
- **README.md** - Component usage examples
- **BEFORE_AFTER.md** - Code comparisons
- **COMPONENT_HIERARCHY.md** - Architecture overview
- **REFACTORING_SUMMARY.md** - Full refactoring summary

---

## ✨ Celebrate!

After integration, you've:
- Reduced file size by >2,000 lines
- Created 11 modular files
- Improved code maintainability
- Enabled component reuse
- Made testing easier

Great work! 🎉
