# Component Hierarchy & Architecture

Visual representation of how the refactored agent components fit together.

## Page Structure

```
pages/agents.vue (PARENT)
├── Header
│   ├── Delete All Sessions Button
│   └── Kill All Agents Button
│
├── Connection Status Badge
│
└── Main Container
    ├── SessionSidebar (LEFT)
    │   ├── Sidebar Header
    │   │   ├── New Session Button
    │   │   └── Resume Session Button
    │   │
    │   ├── SessionFilters ✅ EXTRACTED
    │   │   └── Filter Tabs (Active/All/Ended)
    │   │
    │   └── Sessions List
    │       └── SessionItem ✅ EXTRACTED (multiple)
    │           ├── Avatar
    │           ├── Session Info
    │           └── Actions (End/Delete)
    │
    ├── Chat Area (CENTER)
    │   ├── Empty State (if no session)
    │   │
    │   └── Chat Content (if session active)
    │       ├── Tool Overlays Container
    │       │   ├── TodoWriteOverlay (existing)
    │       │   └── ToolOverlay (existing)
    │       │
    │       ├── TodoWrite Box
    │       │
    │       ├── Messages Container
    │       │   ├── Message Items (multiple)
    │       │   ├── Thinking Indicator
    │       │   └── Processing Indicator
    │       │
    │       ├── Permission Requests
    │       │   └── PermissionRequest ✅ EXTRACTED (multiple)
    │       │       ├── Description
    │       │       └── Actions (Approve/Deny)
    │       │
    │       ├── ToolExecutionBar ✅ EXTRACTED
    │       │   ├── Tool Icon
    │       │   ├── Tool Details
    │       │   └── Pulse Animation
    │       │
    │       └── Input Area
    │           ├── Textarea
    │           └── Send Button
    │
    └── SessionMetrics (RIGHT SIDEBAR)
        └── Metrics Component (existing)

Modals (OVERLAY)
├── CreateSessionModal ✅ EXTRACTED
│   ├── Working Directory Input
│   ├── Permission Mode Select
│   ├── Provider & Model Select
│   ├── Prompt Mode Toggle
│   │   ├── Agent Selection Grid
│   │   │   └── Agent Preview
│   │   └── Custom Prompt Textarea
│   └── Tools Checkboxes
│
└── ResumeSessionModal ✅ EXTRACTED
    ├── Session List View
    │   └── Session Cards (multiple)
    │
    └── Resume Options View
        ├── Selected Session Info
        ├── Working Directory Input
        ├── Permission Mode Select
        ├── System Prompt Textarea
        └── Tools Checkboxes
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                        agents.vue                            │
│                     (Orchestration Layer)                    │
│                                                              │
│  State Management:                                          │
│  • sessions (ref)                                           │
│  • activeSessionId (ref)                                    │
│  • messages (ref)                                           │
│  • sessionPermissions (ref)                                 │
│  • sessionTodos (ref)                                       │
│  • sessionToolExecution (ref)                               │
│                                                              │
│  WebSocket Connection:                                      │
│  • agentWs (useAgentWebSocket composable)                   │
└───────────────┬─────────────────────────────────────────────┘
                │
                │ Props ↓ & Emits ↑
                │
    ┌───────────┴─────────────┬──────────────┬────────────────┐
    │                         │              │                │
    ▼                         ▼              ▼                ▼
┌─────────┐          ┌──────────────┐  ┌─────────┐   ┌──────────┐
│Session  │          │ Permission   │  │  Tool   │   │  Modal   │
│Components│          │ Components   │  │Components│  │Components│
└─────────┘          └──────────────┘  └─────────┘   └──────────┘
    │                       │               │               │
    ├─ SessionFilters       └─ Permission   ├─ ToolExecution│
    │  • Props: filters         Request      │  Bar         ├─ Create
    │  • Emits: update      • Props: perm    │  • Props:    │  Session
    │                       • Emits: approve  │    execution │  Modal
    └─ SessionItem            deny           │              │
       • Props: session                      └─ TodoWrite   └─ Resume
       • Emits: select,                         Overlay        Session
         end, delete                            (existing)     Modal
```

## Utility Layer

```
┌─────────────────────────────────────────────────────────────┐
│                    Utilities (Pure Functions)                │
├─────────────────────────────────────────────────────────────┤
│  messageFormatters.ts                                       │
│  • formatTime(timestamp)           → "2:30 PM"              │
│  • formatRelativeTime(timestamp)   → "5m ago"               │
│  • formatMessage(content)          → HTML string            │
│  • truncatePath(path, maxLen)      → "...file.ts"           │
├─────────────────────────────────────────────────────────────┤
│  todoParser.ts                                              │
│  • parseTodoWrite(content)         → TodoItem[]             │
│  • formatTodosForTool(todos)       → API format             │
├─────────────────────────────────────────────────────────────┤
│  toolParser.ts                                              │
│  • parseToolUse(content)           → ToolExecution          │
│  • getToolIcon(toolName)           → emoji string           │
└─────────────────────────────────────────────────────────────┘
```

## Composables Layer

```
┌─────────────────────────────────────────────────────────────┐
│              Composables (Stateful Logic)                    │
├─────────────────────────────────────────────────────────────┤
│  useMessageScroll.ts                                        │
│  Returns:                                                    │
│  • isUserNearBottom (ref)                                   │
│  • handleScroll(container)                                  │
│  • scrollToBottom(container, smooth)                        │
│  • autoScrollIfNearBottom(container, smooth)                │
└─────────────────────────────────────────────────────────────┘
```

## Component Communication Patterns

### Parent → Child (Props)

```typescript
// agents.vue passes data down
<SessionItem
  :session="session"           // Data
  :is-active="isActive"        // State
/>

<CreateSessionModal
  :show="showModal"            // UI State
  :form-data="sessionForm"     // Form State
  :providers="providers"       // Options Data
  :loading="loading"           // Loading State
/>
```

### Child → Parent (Events)

```typescript
// Components emit events up
<SessionItem
  @select="handleSelect"       // User Action
  @end="handleEnd"             // User Action
  @delete="handleDelete"       // User Action
/>

<PermissionRequest
  @approve="handleApprove"     // Decision Event
  @deny="handleDeny"           // Decision Event
/>
```

### Sibling Communication (via Parent)

```typescript
// SessionFilters changes activeFilter
<SessionFilters
  @update:active-filter="activeFilter = $event"
/>

// SessionItem uses activeFilter from parent
<SessionItem
  :is-active="session.id === activeFilter"
/>
```

## State Management Strategy

### Local Component State
- UI state (hover, focus, etc.)
- Form validation
- Animation states

### Parent Component State (agents.vue)
- Sessions list
- Active session ID
- Messages per session
- Permissions per session
- Tool execution per session

### WebSocket State (useAgentWebSocket)
- Connection status
- Message queue
- Event callbacks

## Component Lifecycle

### SessionItem
```
1. Created with props (session, isActive)
2. Computes avatar using useCharacterAvatar
3. Renders with reactive styling
4. Emits events on user interaction
5. Updates when props change
```

### CreateSessionModal
```
1. Opens (show = true)
2. Loads providers/agents (parent calls API)
3. User fills form (reactive v-model)
4. User clicks Create
5. Emits create event with formData
6. Parent handles session creation
7. Closes on success
```

### PermissionRequest
```
1. Created when permission arrives via WebSocket
2. Displays permission details
3. User clicks Approve/Deny
4. Emits event to parent
5. Parent sends response via WebSocket
6. Removed from DOM when acknowledged
```

## Performance Considerations

### Lazy Loading
```vue
<!-- Modals loaded only when needed -->
<CreateSessionModal v-if="showModal" ... />
<ResumeSessionModal v-if="showResumeModal" ... />
```

### List Virtualization (Future)
```vue
<!-- For large session lists (100+) -->
<virtual-scroller :items="sessions">
  <template #default="{ item }">
    <SessionItem :session="item" />
  </template>
</virtual-scroller>
```

### Computed Properties
```typescript
// Efficient filtering
const filteredSessions = computed(() => {
  return sessions.value.filter(s => s.status !== 'ended')
})
```

## Error Boundaries

```
agents.vue (Top Level)
  ├─ Try/Catch for WebSocket errors
  ├─ Try/Catch for API calls
  └─ Components handle their own errors
      ├─ SessionItem: Graceful avatar fallback
      ├─ Modals: Validation & user feedback
      └─ PermissionRequest: Disable on error
```

## Testing Strategy

### Unit Tests (Utilities)
```typescript
// Test pure functions
describe('formatTime', () => {
  it('formats timestamp correctly', () => {
    expect(formatTime(date)).toBe('2:30 PM')
  })
})
```

### Component Tests (Components)
```typescript
// Test props, emits, rendering
describe('SessionItem', () => {
  it('emits select on click', async () => {
    const wrapper = mount(SessionItem, { props })
    await wrapper.trigger('click')
    expect(wrapper.emitted('select')).toBeTruthy()
  })
})
```

### Integration Tests (Parent)
```typescript
// Test component interactions
describe('agents.vue', () => {
  it('selects session when SessionItem clicked', async () => {
    const wrapper = mount(AgentsPage)
    await wrapper.findComponent(SessionItem).trigger('click')
    expect(wrapper.vm.activeSessionId).toBe('123')
  })
})
```

### E2E Tests (Full Flow)
```typescript
// Test complete workflows
test('create new session', async ({ page }) => {
  await page.click('[data-test="new-session"]')
  await page.fill('[data-test="working-dir"]', '/home/user')
  await page.click('[data-test="create-button"]')
  await expect(page.locator('.session-item')).toBeVisible()
})
```

## Scalability Paths

### Future Enhancements

1. **Extract More Components**
   - `MessagesList.vue` - Message container
   - `ChatArea.vue` - Full chat interface
   - `SessionSidebar.vue` - Complete sidebar

2. **Add More Composables**
   - `useSessionManagement()` - CRUD operations
   - `useSessionMessages()` - Message handling
   - `useSessionPermissions()` - Permission logic
   - `useAgentProviders()` - Provider selection

3. **State Management Library** (if needed)
   - Pinia store for complex state
   - Shared state across multiple pages
   - Better dev tools integration

4. **Component Library**
   - Publish as npm package
   - Use in other projects
   - Storybook documentation

## Migration Path

### Phase 1: ✅ COMPLETE
- Extract utilities
- Extract simple components
- Extract complex modals

### Phase 2: Future
- Integrate into agents.vue
- Remove duplicate code
- Test thoroughly

### Phase 3: Future
- Extract remaining components
- Add more composables
- Optimize performance

### Phase 4: Future
- Add comprehensive tests
- Document API completely
- Consider state management library

## Questions?

- **Integration**: See `INTEGRATION_GUIDE.md`
- **Usage**: See `README.md`
- **Refactoring Summary**: See `REFACTORING_SUMMARY.md`
