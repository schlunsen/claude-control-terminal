# Universal Message Modal Viewer - Implementation Plan

## Overview

Currently, only Edit tool messages are clickable and show a modal with diff viewing. This plan extends clickability to **all messages** in the chat - both user and assistant messages - so users can view complete message content, metadata, and associated tool uses in a modal.

## Current State Analysis

### What We Have ✅

1. Edit tool click handler in `MessageBubble.vue` (lines 202-206)
2. Modal infrastructure in `agents.vue` (lines 211-236) - "Tool Diff Overlay Modal"
3. ESC key handling for modal closing (lines 685-703)
4. Tool uses display with details (lines 31-48 in MessageBubble.vue)
5. Message metadata (role, timestamp, content, toolUses, etc.)

### What We Need ❌

1. Click handler for entire message bubble (not just Edit tools)
2. Universal message modal component (replaces tool-specific modal)
3. Message detail viewer showing full content + metadata
4. Tool uses section in modal (showing all tools with their inputs/outputs)
5. Handling of different message types (user, assistant, system, historical, error)

---

## Implementation Plan

### Phase 1: Create Universal Message Modal Component

**File:** `internal/server/frontend/components/agents/MessageDetailModal.vue` (NEW)

#### Features

- **Full message content display** (formatted markdown/code)
- **Message metadata section:**
  - Role (You/Claude/System)
  - Timestamp (formatted)
  - Message ID
  - Sequence number (if historical)
  - Is Historical badge
  - Is Error badge
- **Tool Uses section** (expandable list):
  - Tool name + icon
  - Tool inputs (formatted JSON or key details)
  - Edit tools: Show diff viewer inline
  - Bash tools: Show command + output
  - File tools: Show file path + relevant details
- **Image attachments** (if any)
- **Thinking content** (if available)
- **Copy message button** (copy raw content to clipboard)
- **Close button** + ESC key support

#### Props

```typescript
interface Props {
  show: boolean
  message: Message | null
  formatTime: (date: Date) => string
  formatMessage: (content: string | ContentBlock[]) => string
}
```

#### Emits

```typescript
emits: ['close', 'open-lightbox']
```

---

### Phase 2: Update MessageBubble.vue

#### Changes

1. Make entire message bubble clickable (add click handler to root `.message` div)
2. Add hover styling to indicate clickability:
   - Subtle background color change
   - Cursor pointer
   - Optional "click to view details" icon/hint
3. Emit new event `@message-click` when message is clicked
4. Keep existing `@tool-click` for backward compatibility (but prioritize message click)

#### Updated Template

```vue
<div class="message"
     :class="{ /* existing classes */ }"
     @click="handleMessageClick"
     role="button"
     tabindex="0"
     @keydown.enter="handleMessageClick">
  <!-- existing content -->
</div>
```

#### New Styles

```css
.message {
  cursor: pointer;
  transition: background-color 0.2s, transform 0.1s;
}

.message:hover {
  background-color: rgba(139, 92, 246, 0.05);
}

.message:active {
  transform: scale(0.995);
}
```

#### New Emit

```typescript
emit('message-click', { message: props.message })
```

---

### Phase 3: Update agents.vue

#### Changes

1. Replace "Tool Diff Overlay Modal" with universal "Message Detail Modal"
2. Update state management:
   - `showMessageDetailModal` (replaces `showToolDiffOverlay`)
   - `selectedMessage` (replaces `selectedToolData`)
3. Add handler `handleMessageClick` (replaces `handleToolClick`)
4. Update ESC key handler to close message modal
5. Update `has-modal-open` prop to include message modal

#### Template Changes

```vue
<!-- Replace Tool Diff Overlay Modal (lines 210-236) with: -->
<Teleport to="body">
  <transition name="fade">
    <MessageDetailModal
      v-if="showMessageDetailModal && selectedMessage"
      :show="showMessageDetailModal"
      :message="selectedMessage"
      :format-time="formatTime"
      :format-message="formatMessage"
      @close="closeMessageDetailModal"
      @open-lightbox="openLightbox"
    />
  </transition>
</Teleport>
```

#### Script Changes

```typescript
// State
const showMessageDetailModal = ref(false)
const selectedMessage = ref<any>(null)

// Handler
const handleMessageClick = ({ message }: { message: any }) => {
  selectedMessage.value = message
  showMessageDetailModal.value = true
}

// Close handler
const closeMessageDetailModal = () => {
  showMessageDetailModal.value = false
  setTimeout(() => {
    selectedMessage.value = null
  }, 300) // Wait for fade transition
}

// Update ESC handler (line 688)
if (showMessageDetailModal.value) {
  closeMessageDetailModal()
  event.preventDefault()
  event.stopPropagation()
  return
}
```

#### Update MessageBubble Usage

```vue
<MessageBubble
  v-for="message in activeMessages"
  :key="message.id"
  :message="message"
  :format-time="formatTime"
  :format-message="formatMessage"
  @open-lightbox="openLightbox"
  @message-click="handleMessageClick"  <!-- NEW -->
  @tool-click="handleToolClick"  <!-- Keep for backward compat, but make optional -->
>
```

---

### Phase 4: Enhance Tool Display in Modal

#### Features

1. Show all tool uses from `message.toolUses` array
2. For each tool, display:
   - Tool icon (use existing `getToolIcon` utility)
   - Tool name
   - Tool inputs (expandable/collapsible JSON)
   - For Edit: Show EditDiffMessage component inline
   - For Bash: Show command + output (if available from tool result)
   - For Read/Write: Show file path + content preview
3. Make tool sections expandable (collapsed by default, expand to see full details)

#### Component Structure

```vue
<div class="tool-uses-section" v-if="message.toolUses && message.toolUses.length > 0">
  <h3>Tool Uses ({{ message.toolUses.length }})</h3>
  <div v-for="(tool, idx) in message.toolUses" :key="idx" class="tool-use-detail">
    <div class="tool-header" @click="toggleToolExpand(idx)">
      <svg><!-- tool icon --></svg>
      <span class="tool-name">{{ tool.name }}</span>
      <button class="expand-btn">{{ expandedTools[idx] ? '▼' : '▶' }}</button>
    </div>

    <div v-if="expandedTools[idx]" class="tool-body">
      <!-- For Edit tools: Show diff -->
      <EditDiffMessage v-if="tool.name === 'Edit'" /* ... */ />

      <!-- For other tools: Show input details -->
      <div v-else class="tool-input">
        <pre>{{ formatToolInput(tool.input) }}</pre>
      </div>
    </div>
  </div>
</div>
```

---

### Phase 5: Polish & Accessibility

#### Enhancements

1. **Keyboard navigation:**
   - Tab to focus message
   - Enter/Space to open modal
   - ESC to close modal
2. **ARIA attributes:**
   - `role="button"` on message bubble
   - `aria-label="View message details"`
   - `aria-expanded` for expandable sections
3. **Loading state** for historical messages (if they're being loaded)
4. **Copy-to-clipboard** button for message content
5. **Visual indicator** (icon) on message hover showing it's clickable
6. **Smooth animations** for modal open/close
7. **Mobile-responsive** modal (full-screen on mobile)

---

## File Changes Summary

| File | Type | Changes |
|------|------|---------|
| `components/agents/MessageDetailModal.vue` | NEW | Universal message detail viewer component |
| `components/agents/MessageBubble.vue` | MODIFY | Add click handler, hover styles, message-click emit |
| `app/pages/agents.vue` | MODIFY | Replace tool modal with message modal, update handlers |
| `composables/agents/useMessageHelpers.ts` | MODIFY | Add `formatToolInput()` helper for displaying tool parameters |

---

## Benefits

1. **Consistency**: Same interaction pattern for all messages (not just Edit tools)
2. **Discoverability**: Users can click any message to see full details
3. **Debugging**: Easy access to message metadata, tool uses, and raw content
4. **Accessibility**: Better keyboard navigation and screen reader support
5. **Extensibility**: Easy to add more message details in the future (e.g., token usage, model info)
6. **Mobile-friendly**: Full-screen modal works well on small screens

---

## Migration Path

1. ✅ Create `MessageDetailModal.vue` component
2. ✅ Update `MessageBubble.vue` with click handler
3. ✅ Update `agents.vue` to use new modal
4. ✅ Test with different message types (user, assistant, error, historical)
5. ✅ Add keyboard navigation and accessibility
6. ✅ Polish animations and styling
7. ✅ Optional: Keep old `@tool-click` behavior as fallback for Edit-only clicks

---

## Optional Enhancements (Future)

- [ ] Pagination for messages with many tool uses
- [ ] Search/filter within message content
- [ ] Export message to JSON/markdown
- [ ] Link to related messages (e.g., "this message references file X")
- [ ] Diff view between message versions (if edited)
- [ ] Syntax highlighting for code blocks in content

---

## Technical Notes

### Message Data Structure

```typescript
interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string | ContentBlock[]
  timestamp: Date
  toolUse?: string  // Legacy: single tool name (deprecated)
  toolUses?: ToolUse[]  // New: array of tool uses with details
  isToolResult?: boolean
  isExecutionStatus?: boolean
  isPermissionDecision?: boolean
  isHistorical?: boolean
  isError?: boolean
  editToolData?: EditToolData  // For Edit tools specifically
  thinkingContent?: string  // Claude's internal thinking
  sequence?: number  // For historical messages
}
```

### Tool Use Data Structure

```typescript
interface ToolUse {
  name: string
  input?: {
    file_path?: string
    command?: string
    pattern?: string
    old_string?: string
    new_string?: string
    replace_all?: boolean
    [key: string]: any
  }
}
```

---

## Implementation Order

1. **Phase 1** - Create modal component (core functionality)
2. **Phase 2** - Make messages clickable (interaction layer)
3. **Phase 3** - Wire up to parent component (integration)
4. **Phase 4** - Enhance tool display (rich content)
5. **Phase 5** - Polish and accessibility (refinement)

This phased approach allows for incremental development and testing at each stage.
