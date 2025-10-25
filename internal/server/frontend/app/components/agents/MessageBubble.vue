<template>
  <div
    class="message"
    :class="{
      [message.role]: true,
      isToolResult: message.isToolResult,
      isExecutionStatus: message.isExecutionStatus,
      isPermissionDecision: message.isPermissionDecision,
      isHistorical: message.isHistorical,
      isError: message.isError
    }"
    @click="handleMessageClick"
    role="button"
    tabindex="0"
    @keydown.enter="handleMessageClick"
    @keydown.space.prevent="handleMessageClick"
    aria-label="View message details"
  >
    <div class="message-header">
      <span class="message-role">{{ roleName }}</span>
      <span class="message-time">{{ formattedTime }}</span>
      <svg class="click-hint-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <line x1="12" y1="16" x2="12" y2="12"></line>
        <line x1="12" y1="8" x2="12.01" y2="8"></line>
      </svg>
    </div>

    <!-- Text content -->
    <div v-if="textContent" class="message-content" v-html="formattedContent"></div>

    <!-- Images -->
    <div v-if="imageBlocks.length > 0" class="message-images">
      <img
        v-for="(img, idx) in imageBlocks"
        :key="idx"
        :src="img.dataUrl"
        :alt="`Image ${idx + 1}`"
        class="message-image"
        @click="$emit('open-lightbox', { images: imageBlocks, startIndex: idx })"
      />
    </div>

    <!-- Tool use indicators -->
    <div v-if="displayToolUses.length > 0" class="tool-uses">
      <div
        v-for="(tool, idx) in displayToolUses"
        :key="idx"
        class="tool-use"
        :class="{ clickable: tool.isClickable }"
        @click="handleToolClick(tool)"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
        </svg>
        <span>Using {{ tool.name }}</span>
        <span v-if="tool.detail" class="tool-detail">{{ tool.detail }}</span>
        <svg v-if="tool.isClickable" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="click-icon">
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </div>
    </div>

    <!-- Expandable Edit Diff (when diffDisplayLocation is 'chat') -->
    <slot name="edit-diff"></slot>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface ContentBlock {
  type: string
  text?: string
  source?: {
    type: string
    media_type: string
    data: string
  }
}

interface ToolUse {
  name: string
  input?: any
}

interface Message {
  id: string
  role: string
  content: string | ContentBlock[]
  timestamp: Date
  toolUse?: string  // Legacy: single tool name (deprecated)
  toolUses?: ToolUse[]  // New: array of tool uses with details
  isToolResult?: boolean
  isExecutionStatus?: boolean
  isPermissionDecision?: boolean
  isHistorical?: boolean
  isError?: boolean
}

interface Props {
  message: Message
  formatTime: (date: Date) => string
  formatMessage: (content: string | ContentBlock[]) => string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'open-lightbox': [{ images: any[], startIndex: number }]
  'tool-click': [{ tool: any }]
  'message-click': [{ message: Message }]
}>()

const roleName = computed(() => {
  if (props.message.role === 'user') return 'You'
  if (props.message.role === 'system') return 'System'
  return 'Claude'
})

const formattedTime = computed(() => props.formatTime(props.message.timestamp))

// Extract text content from message
const textContent = computed(() => {
  const content = props.message.content

  // Handle string content (legacy)
  if (typeof content === 'string') {
    return content
  }

  // Handle array content (structured)
  if (Array.isArray(content)) {
    const textBlocks = content
      .filter((block: ContentBlock) => block.type === 'text')
      .map((block: ContentBlock) => block.text)
      .filter(Boolean)

    return textBlocks.join('\n\n')
  }

  return ''
})

// Extract image blocks from message
const imageBlocks = computed(() => {
  const content = props.message.content

  if (!Array.isArray(content)) return []

  return content
    .filter((block: ContentBlock) => block.type === 'image' && block.source)
    .map((block: ContentBlock) => ({
      dataUrl: `data:${block.source!.media_type};base64,${block.source!.data}`,
      mediaType: block.source!.media_type
    }))
})

// Format text content
const formattedContent = computed(() => {
  if (!textContent.value) return ''
  return props.formatMessage(textContent.value)
})

// Extract tool uses for display
const displayToolUses = computed(() => {
  // Prefer new toolUses array format
  if (props.message.toolUses && Array.isArray(props.message.toolUses)) {
    return props.message.toolUses.map(tool => {
      let detail = ''
      let isClickable = false

      // Extract relevant details based on tool type
      if (tool.input) {
        if (tool.name === 'Edit' && tool.input.file_path) {
          // Extract filename from path
          const filename = tool.input.file_path.split('/').pop() || tool.input.file_path
          detail = filename
          isClickable = true // Edit operations are clickable to show diff
        } else if (tool.name === 'Read' && tool.input.file_path) {
          const filename = tool.input.file_path.split('/').pop() || tool.input.file_path
          detail = filename
        } else if (tool.name === 'Write' && tool.input.file_path) {
          const filename = tool.input.file_path.split('/').pop() || tool.input.file_path
          detail = filename
        } else if (tool.name === 'Bash' && tool.input.command) {
          // Show full command (will wrap if needed)
          detail = tool.input.command
        } else if (tool.name === 'Grep' && tool.input.pattern) {
          detail = tool.input.pattern
        }
      }

      return {
        name: tool.name,
        detail,
        isClickable,
        fullData: tool // Keep full tool data for click handler
      }
    })
  }

  // Fall back to legacy single toolUse string
  if (props.message.toolUse) {
    return [{
      name: props.message.toolUse,
      detail: '',
      isClickable: false,
      fullData: null
    }]
  }

  return []
})

// Handle message click
const handleMessageClick = (event: Event) => {
  emit('message-click', { message: props.message })
}

// Handle tool click (for backward compatibility)
const handleToolClick = (tool: any) => {
  if (tool.isClickable) {
    emit('tool-click', { tool: tool.fullData })
  }
}
</script>

<style scoped>
.message {
  margin-bottom: 24px;
  cursor: pointer;
  transition: background-color 0.2s, transform 0.1s;
  padding: 8px;
  margin-left: -8px;
  margin-right: -8px;
  border-radius: 12px;
}

.message:hover {
  background-color: rgba(139, 92, 246, 0.05);
}

.message:active {
  transform: scale(0.995);
}

.message:focus {
  outline: 2px solid var(--accent-purple);
  outline-offset: 2px;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.message-role {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-primary);
}

.message.assistant .message-role {
  color: var(--accent-purple);
}

.message-time {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.click-hint-icon {
  color: var(--text-secondary);
  opacity: 0;
  transition: opacity 0.2s;
  margin-left: auto;
}

.message:hover .click-hint-icon {
  opacity: 0.5;
}

.message-content {
  background: var(--card-bg);
  padding: 12px 16px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  font-size: 0.95rem;
  line-height: 1.6;
  color: var(--text-primary);
}

.message.user .message-content {
  background: var(--accent-purple);
  color: white;
  border-color: var(--accent-purple);
  margin-left: 48px;
}

.message.assistant .message-content {
  margin-right: 48px;
}

.message-content :deep(code) {
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.9em;
}

.message.user .message-content :deep(code) {
  background: rgba(255, 255, 255, 0.2);
}

.message-content :deep(pre) {
  background: var(--bg-secondary);
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}

.message-content :deep(.system-message) {
  color: var(--text-secondary);
  font-style: italic;
  opacity: 0.7;
}

.message-content :deep(.message-link) {
  color: var(--accent-purple);
  text-decoration: underline;
  transition: all 0.2s;
  word-break: break-all;
}

.message-content :deep(.message-link:hover) {
  color: var(--accent-purple-hover);
  text-decoration: none;
  opacity: 0.8;
}

.message.user .message-content :deep(.message-link) {
  color: rgba(255, 255, 255, 0.9);
  text-decoration: underline;
}

.message.user .message-content :deep(.message-link:hover) {
  color: white;
  text-decoration: none;
}

.message.isError .message-content {
  background: rgba(220, 53, 69, 0.1);
  border-color: rgba(220, 53, 69, 0.3);
  color: #dc3545;
}

.message.isHistorical {
  opacity: 0.8;
}

.message.isToolResult .message-content {
  background: var(--bg-secondary);
  border-left: 3px solid var(--accent-purple);
  font-family: monospace;
  font-size: 0.85rem;
}

.message.isExecutionStatus .message-content {
  background: rgba(139, 92, 246, 0.1);
  border-left: 3px solid var(--accent-purple);
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.message.isPermissionDecision .message-content {
  background: rgba(34, 197, 94, 0.1);
  border-left: 3px solid #22c55e;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.tool-uses {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.tool-use {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  background: var(--bg-secondary);
  border-radius: 12px;
  font-size: 0.85rem;
  color: var(--text-secondary);
  transition: all 0.2s;
  max-width: 100%;
  word-break: break-word;
}

.tool-use.clickable {
  cursor: pointer;
  border: 1px solid transparent;
}

.tool-use.clickable:hover {
  background: var(--accent-purple);
  color: white;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.3);
}

.tool-use.clickable:hover .tool-detail {
  color: rgba(255, 255, 255, 0.9);
}

.tool-detail {
  color: var(--accent-purple);
  font-weight: 500;
  margin-left: 4px;
}

.click-icon {
  margin-left: 4px;
  opacity: 0.6;
}

/* Message Images */
.message-images {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 12px;
}

.message.user .message-images {
  margin-left: 48px;
}

.message.assistant .message-images {
  margin-right: 48px;
}

.message-image {
  max-width: 300px;
  max-height: 300px;
  object-fit: contain;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: all 0.2s;
}

.message-image:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.message.user .message-image {
  border-color: rgba(255, 255, 255, 0.3);
}

@media (max-width: 768px) {
  .message.user .message-content {
    margin-left: 24px;
  }

  .message.assistant .message-content {
    margin-right: 24px;
  }

  .message.user .message-images {
    margin-left: 24px;
  }

  .message.assistant .message-images {
    margin-right: 24px;
  }

  .message-image {
    max-width: 200px;
    max-height: 200px;
  }
}
</style>
