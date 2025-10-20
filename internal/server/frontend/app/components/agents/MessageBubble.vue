<template>
  <div class="message" :class="{
    [message.role]: true,
    isToolResult: message.isToolResult,
    isExecutionStatus: message.isExecutionStatus,
    isPermissionDecision: message.isPermissionDecision,
    isHistorical: message.isHistorical,
    isError: message.isError
  }">
    <div class="message-header">
      <span class="message-role">{{ roleName }}</span>
      <span class="message-time">{{ formattedTime }}</span>
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

    <!-- Tool use indicator -->
    <div v-if="message.toolUse" class="tool-use">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
      </svg>
      Using {{ message.toolUse }}
    </div>
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

interface Message {
  id: string
  role: string
  content: string | ContentBlock[]
  timestamp: Date
  toolUse?: string
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

defineEmits<{
  'open-lightbox': [{ images: any[], startIndex: number }]
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
</script>

<style scoped>
.message {
  margin-bottom: 24px;
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

.tool-use {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 4px 12px;
  background: var(--bg-secondary);
  border-radius: 12px;
  font-size: 0.85rem;
  color: var(--text-secondary);
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
