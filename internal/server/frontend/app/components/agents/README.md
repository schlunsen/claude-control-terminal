# Agent Components Library

Modular, reusable Vue components for the Live Agents feature.

## Quick Reference

### 🧩 Components

| Component | Size | Purpose | Props | Emits |
|-----------|------|---------|-------|-------|
| **SessionItem** | 180 lines | Session card display | `session`, `isActive` | `select`, `end`, `delete` |
| **SessionFilters** | 90 lines | Filter tabs | `activeFilter`, `filters` | `update:activeFilter` |
| **PermissionRequest** | 130 lines | Permission card | `permission` | `approve`, `deny` |
| **ToolExecutionBar** | 140 lines | Tool status | `toolExecution` | - |
| **CreateSessionModal** | 680 lines | Session creation | `show`, `formData`, `providers`, `agents`, etc. | `close`, `create`, `workingDirectoryChange`, `agentSelect` |
| **ResumeSessionModal** | 580 lines | Session resume | `show`, `sessions`, `selectedSession`, `formData`, etc. | `close`, `selectSession`, `back`, `resume` |

### 🔧 Utilities

| File | Functions | Purpose |
|------|-----------|---------|
| **messageFormatters.ts** | `formatTime`, `formatRelativeTime`, `formatMessage`, `truncatePath` | Message & time formatting |
| **todoParser.ts** | `parseTodoWrite`, `formatTodosForTool` | TodoWrite parsing |
| **toolParser.ts** | `parseToolUse`, `getToolIcon` | Tool execution parsing |

### 🎯 Composables

| Composable | Functions | Purpose |
|------------|-----------|---------|
| **useMessageScroll** | `handleScroll`, `scrollToBottom`, `autoScrollIfNearBottom` | Auto-scroll management |

## Usage Examples

### SessionItem

```vue
<template>
  <SessionItem
    v-for="session in sessions"
    :key="session.id"
    :session="session"
    :is-active="activeSessionId === session.id"
    @select="handleSelect"
    @end="handleEnd"
    @delete="handleDelete"
  />
</template>

<script setup>
import SessionItem from '~/components/agents/SessionItem.vue'

const handleSelect = (sessionId) => {
  console.log('Selected:', sessionId)
}
</script>
```

### SessionFilters

```vue
<template>
  <SessionFilters
    :active-filter="activeFilter"
    :filters="filtersWithCounts"
    @update:active-filter="activeFilter = $event"
  />
</template>

<script setup>
import SessionFilters from '~/components/agents/SessionFilters.vue'

const activeFilter = ref('active')
const filtersWithCounts = computed(() => [
  { label: 'Active', value: 'active', count: 5 },
  { label: 'All', value: 'all', count: 10 },
  { label: 'Ended', value: 'ended', count: 5 }
])
</script>
```

### PermissionRequest

```vue
<template>
  <PermissionRequest
    :permission="permission"
    @approve="handleApprove"
    @deny="handleDeny"
  />
</template>

<script setup>
import PermissionRequest from '~/components/agents/PermissionRequest.vue'

const permission = {
  request_id: '123',
  description: 'Write to config.json',
  timestamp: new Date()
}
</script>
```

### ToolExecutionBar

```vue
<template>
  <ToolExecutionBar :tool-execution="currentTool" />
</template>

<script setup>
import ToolExecutionBar from '~/components/agents/ToolExecutionBar.vue'

const currentTool = {
  toolName: 'Read',
  filePath: '/path/to/file.ts',
  timestamp: new Date()
}
</script>
```

### CreateSessionModal

```vue
<template>
  <CreateSessionModal
    :show="showModal"
    :form-data="sessionForm"
    :providers="providers"
    :current-provider="currentProvider"
    :agents="agents"
    :selected-agent-preview="preview"
    :loading-providers="false"
    :loading-agents="false"
    :creating="false"
    @close="showModal = false"
    @create="handleCreate"
    @working-directory-change="loadAgents"
    @agent-select="loadPreview"
  />
</template>

<script setup>
import CreateSessionModal from '~/components/agents/CreateSessionModal.vue'

const sessionForm = ref({
  workingDirectory: '',
  permissionMode: 'default',
  modelProvider: 'anthropic',
  model: 'claude-sonnet-4.5-20250514',
  systemPrompt: '',
  promptMode: 'agent',
  selectedAgent: '',
  tools: ['Read', 'Write', 'Edit', 'Bash', 'Search', 'TodoWrite']
})

const handleCreate = (formData) => {
  console.log('Creating session with:', formData)
}
</script>
```

### ResumeSessionModal

```vue
<template>
  <ResumeSessionModal
    :show="showModal"
    :sessions="availableSessions"
    :selected-session="selectedSession"
    :form-data="resumeForm"
    :loading="false"
    :resuming="false"
    @close="showModal = false"
    @select-session="handleSelect"
    @back="selectedSession = null"
    @resume="handleResume"
  />
</template>

<script setup>
import ResumeSessionModal from '~/components/agents/ResumeSessionModal.vue'

const resumeForm = ref({
  workingDirectory: '',
  permissionMode: 'default',
  systemPrompt: '',
  tools: ['Read', 'Write', 'Edit', 'Bash', 'Search', 'TodoWrite']
})

const handleResume = (formData) => {
  console.log('Resuming session with:', formData)
}
</script>
```

### Utilities

```typescript
// Message Formatters
import {
  formatTime,
  formatRelativeTime,
  formatMessage,
  truncatePath
} from '~/utils/agents/messageFormatters'

const time = formatTime(new Date()) // "2:30 PM"
const relative = formatRelativeTime(Date.now() - 300000) // "5m ago"
const html = formatMessage('**bold** and *italic*') // "<strong>bold</strong> and <em>italic</em>"
const short = truncatePath('/very/long/path/to/file.ts', 20) // ".../file.ts"

// Todo Parser
import { parseTodoWrite, formatTodosForTool } from '~/utils/agents/todoParser'

const todos = parseTodoWrite('1. First task\n2. Second task')
// Returns: [{ content: 'First task', status: 'in_progress' }, ...]

// Tool Parser
import { parseToolUse, getToolIcon } from '~/utils/agents/toolParser'

const tool = parseToolUse('Using Read file: config.json')
// Returns: { toolName: 'Read', filePath: 'config.json', timestamp: Date }

const icon = getToolIcon('Bash') // Returns: '⚡'
```

### Composables

```typescript
import { useMessageScroll } from '~/composables/agents/useMessageScroll'

const messagesContainer = ref<HTMLElement | null>(null)
const { isUserNearBottom, handleScroll, scrollToBottom, autoScrollIfNearBottom } = useMessageScroll()

// In template
<div
  ref="messagesContainer"
  @scroll="handleScroll(messagesContainer)"
>
  <!-- messages -->
</div>

// In script
scrollToBottom(messagesContainer.value, true) // smooth scroll
autoScrollIfNearBottom(messagesContainer.value) // only if near bottom
```

## TypeScript Interfaces

### SessionItem Props

```typescript
interface Session {
  id: string
  status: string
  message_count: number
  cost_usd?: number
}

interface Props {
  session: Session
  isActive: boolean
}
```

### SessionFilters Props

```typescript
interface Filter {
  label: string
  value: string
  count: number
}

interface Props {
  activeFilter: string
  filters: Filter[]
}
```

### PermissionRequest Props

```typescript
interface Permission {
  request_id: string
  description: string
  timestamp: string | Date
}

interface Props {
  permission: Permission
}
```

### ToolExecutionBar Props

```typescript
interface ToolExecution {
  toolName: string
  filePath?: string
  command?: string
  pattern?: string
  detail?: string
  timestamp: Date
}

interface Props {
  toolExecution: ToolExecution | null
}
```

### CreateSessionModal Props

```typescript
interface Provider {
  id: string
  name: string
  icon: string
  models: string[]
  base_url?: string
}

interface Agent {
  name: string
  model?: string
  color?: string
  description?: string
  system_prompt?: string
}

interface SessionFormData {
  workingDirectory: string
  permissionMode: string
  modelProvider: string
  model: string
  systemPrompt: string
  promptMode: 'agent' | 'custom'
  selectedAgent: string
  tools: string[]
}

interface Props {
  show: boolean
  formData: SessionFormData
  providers: Provider[]
  currentProvider: any
  agents: Agent[]
  selectedAgentPreview: Agent | null
  loadingProviders: boolean
  loadingAgents: boolean
  creating: boolean
}
```

### ResumeSessionModal Props

```typescript
interface Session {
  conversation_id: string
  session_name?: string
  working_directory?: string
  total_messages: number
  last_activity: string | Date
}

interface ResumeFormData {
  workingDirectory: string
  permissionMode: string
  systemPrompt: string
  tools: string[]
}

interface Props {
  show: boolean
  sessions: Session[]
  selectedSession: Session | null
  formData: ResumeFormData
  loading: boolean
  resuming: boolean
}
```

## Styling

All components use CSS variables for theming:

```css
--card-bg          /* Component backgrounds */
--bg-primary       /* Input backgrounds */
--bg-secondary     /* Hover states */
--bg-tertiary      /* Active states */
--border-color     /* Borders */
--text-primary     /* Main text */
--text-secondary   /* Secondary text */
--text-tertiary    /* Disabled text */
--accent-purple    /* Primary accent */
--color-success    /* Success states */
--color-warning    /* Warning states */
--color-error      /* Error states */
```

## Responsive Design

All components are mobile-responsive:
- **Desktop**: Optimal layout (≥1024px)
- **Tablet**: Adapted layout (640px-1024px)
- **Mobile**: Single-column layout (≤640px)

## Accessibility

Components include:
- Semantic HTML
- ARIA labels where appropriate
- Keyboard navigation support
- Focus states
- Screen reader friendly

## Testing

Components are ready for:
- **Unit tests**: Test props, emits, computed properties
- **Component tests**: Test user interactions
- **E2E tests**: Test complete workflows

Example test:
```typescript
import { mount } from '@vue/test-utils'
import SessionItem from './SessionItem.vue'

describe('SessionItem', () => {
  it('emits select event when clicked', async () => {
    const wrapper = mount(SessionItem, {
      props: {
        session: { id: '123', status: 'active', message_count: 5 },
        isActive: false
      }
    })

    await wrapper.trigger('click')
    expect(wrapper.emitted('select')).toBeTruthy()
    expect(wrapper.emitted('select')[0]).toEqual(['123'])
  })
})
```

## Performance

- All components use `<style scoped>` to avoid CSS pollution
- Computed properties for efficient reactivity
- Event delegation where appropriate
- Lazy loading recommended for modals

## Browser Support

Supports all modern browsers:
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Contributing

When adding new components:
1. Follow the existing patterns
2. Include TypeScript interfaces
3. Add JSDoc comments
4. Include usage examples
5. Write tests
6. Update this README

## Questions?

See `INTEGRATION_GUIDE.md` for detailed integration instructions.
