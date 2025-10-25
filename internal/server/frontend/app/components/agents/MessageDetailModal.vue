<template>
  <Teleport to="body">
    <transition name="fade">
      <div v-if="show && message" class="message-detail-modal-backdrop" @click="$emit('close')">
        <div class="message-detail-modal" @click.stop>
          <!-- Modal Header -->
          <div class="modal-header">
            <div class="header-left">
              <h3>Message Details</h3>
              <div class="badges">
                <span v-if="message.isHistorical" class="badge historical">Historical</span>
                <span v-if="message.isError" class="badge error">Error</span>
                <span v-if="message.isToolResult" class="badge tool-result">Tool Result</span>
                <span v-if="message.isExecutionStatus" class="badge execution">Execution Status</span>
                <span v-if="message.isPermissionDecision" class="badge permission">Permission</span>
              </div>
            </div>
            <button class="close-btn" @click="$emit('close')" aria-label="Close modal">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>

          <!-- Modal Body -->
          <div class="modal-body">
            <!-- Message Metadata -->
            <div class="metadata-section">
              <div class="metadata-grid">
                <div class="metadata-item">
                  <span class="metadata-label">Role</span>
                  <span class="metadata-value" :class="`role-${message.role}`">{{ roleName }}</span>
                </div>
                <div class="metadata-item">
                  <span class="metadata-label">Timestamp</span>
                  <span class="metadata-value">{{ formattedTimestamp }}</span>
                </div>
                <div class="metadata-item">
                  <span class="metadata-label">Message ID</span>
                  <span class="metadata-value monospace">{{ message.id }}</span>
                </div>
                <div v-if="message.sequence !== undefined" class="metadata-item">
                  <span class="metadata-label">Sequence</span>
                  <span class="metadata-value">#{{ message.sequence }}</span>
                </div>
              </div>
            </div>

            <!-- Thinking Content (if available) -->
            <div v-if="thinkingContent" class="thinking-section">
              <div class="section-header" @click="expandedThinking = !expandedThinking">
                <h4>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"></circle>
                    <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
                    <line x1="12" y1="17" x2="12.01" y2="17"></line>
                  </svg>
                  Thinking
                </h4>
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  :style="{ transform: expandedThinking ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }"
                >
                  <polyline points="6 9 12 15 18 9"></polyline>
                </svg>
              </div>
              <transition name="expand">
                <div v-if="expandedThinking" class="thinking-content" v-html="formattedThinking"></div>
              </transition>
            </div>

            <!-- Message Content -->
            <div class="content-section">
              <div class="section-header">
                <h4>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                    <polyline points="14 2 14 8 20 8"></polyline>
                  </svg>
                  Content
                </h4>
                <button class="copy-btn" @click="copyContent" :title="copyButtonText">
                  <svg v-if="!copied" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                  </svg>
                  <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"></polyline>
                  </svg>
                  {{ copyButtonText }}
                </button>
              </div>
              <div class="message-content" v-html="formattedContent"></div>
            </div>

            <!-- Images (if any) -->
            <div v-if="imageBlocks.length > 0" class="images-section">
              <div class="section-header">
                <h4>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                    <circle cx="8.5" cy="8.5" r="1.5"></circle>
                    <polyline points="21 15 16 10 5 21"></polyline>
                  </svg>
                  Images ({{ imageBlocks.length }})
                </h4>
              </div>
              <div class="images-grid">
                <img
                  v-for="(img, idx) in imageBlocks"
                  :key="idx"
                  :src="img.dataUrl"
                  :alt="`Image ${idx + 1}`"
                  class="modal-image"
                  @click="$emit('open-lightbox', { images: imageBlocks, startIndex: idx })"
                />
              </div>
            </div>

            <!-- Tool Uses Section -->
            <div v-if="displayToolUses.length > 0" class="tools-section">
              <div class="section-header">
                <h4>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                  </svg>
                  Tool Uses ({{ displayToolUses.length }})
                </h4>
              </div>

              <div class="tools-list">
                <div
                  v-for="(tool, idx) in displayToolUses"
                  :key="idx"
                  class="tool-item"
                  :class="{ expanded: expandedTools[idx] }"
                >
                  <div class="tool-header" @click="toggleToolExpand(idx)">
                    <div class="tool-name-section">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                      </svg>
                      <span class="tool-name">{{ tool.name }}</span>
                      <span v-if="tool.detail" class="tool-detail-badge">{{ tool.detail }}</span>
                    </div>
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      class="expand-icon"
                      :style="{ transform: expandedTools[idx] ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }"
                    >
                      <polyline points="6 9 12 15 18 9"></polyline>
                    </svg>
                  </div>

                  <transition name="expand">
                    <div v-if="expandedTools[idx]" class="tool-body">
                      <!-- Special handling for Edit tools: Show diff -->
                      <EditDiffMessage
                        v-if="tool.name === 'Edit' && tool.input"
                        :file-path="tool.input.file_path || 'Unknown file'"
                        :old-string="tool.input.old_string || ''"
                        :new-string="tool.input.new_string || ''"
                        :replace-all="tool.input.replace_all || false"
                        status="completed"
                      />

                      <!-- For other tools: Show formatted input -->
                      <div v-else-if="tool.input" class="tool-input">
                        <h5>Tool Input</h5>
                        <pre class="tool-input-json">{{ formatToolInput(tool.input) }}</pre>
                      </div>

                      <!-- No input data -->
                      <div v-else class="tool-no-input">
                        <em>No input data available</em>
                      </div>
                    </div>
                  </transition>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import EditDiffMessage from './EditDiffMessage.vue'

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
  toolUse?: string
  toolUses?: ToolUse[]
  isToolResult?: boolean
  isExecutionStatus?: boolean
  isPermissionDecision?: boolean
  isHistorical?: boolean
  isError?: boolean
  thinkingContent?: string
  sequence?: number
}

interface Props {
  show: boolean
  message: Message | null
  formatTime: (date: Date) => string
  formatMessage: (content: string | ContentBlock[]) => string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'close': []
  'open-lightbox': [{ images: any[], startIndex: number }]
}>()

// State
const expandedThinking = ref(false)
const expandedTools = ref<Record<number, boolean>>({})
const copied = ref(false)

// Computed properties
const roleName = computed(() => {
  if (!props.message) return ''
  if (props.message.role === 'user') return 'You'
  if (props.message.role === 'system') return 'System'
  return 'Claude'
})

const formattedTimestamp = computed(() => {
  if (!props.message) return ''
  return props.formatTime(props.message.timestamp)
})

const thinkingContent = computed(() => {
  return props.message?.thinkingContent || ''
})

const formattedThinking = computed(() => {
  if (!thinkingContent.value) return ''
  return props.formatMessage(thinkingContent.value)
})

// Extract text content
const textContent = computed(() => {
  if (!props.message) return ''
  const content = props.message.content

  if (typeof content === 'string') {
    return content
  }

  if (Array.isArray(content)) {
    const textBlocks = content
      .filter((block: ContentBlock) => block.type === 'text')
      .map((block: ContentBlock) => block.text)
      .filter(Boolean)

    return textBlocks.join('\n\n')
  }

  return ''
})

// Extract image blocks
const imageBlocks = computed(() => {
  if (!props.message) return []
  const content = props.message.content

  if (!Array.isArray(content)) return []

  return content
    .filter((block: ContentBlock) => block.type === 'image' && block.source)
    .map((block: ContentBlock) => ({
      dataUrl: `data:${block.source!.media_type};base64,${block.source!.data}`,
      mediaType: block.source!.media_type
    }))
})

const formattedContent = computed(() => {
  if (!textContent.value) return '<em style="color: var(--text-secondary);">No text content</em>'
  return props.formatMessage(textContent.value)
})

// Extract tool uses
const displayToolUses = computed(() => {
  if (!props.message) return []

  // Prefer new toolUses array format
  if (props.message.toolUses && Array.isArray(props.message.toolUses)) {
    return props.message.toolUses.map(tool => {
      let detail = ''

      if (tool.input) {
        if (tool.name === 'Edit' && tool.input.file_path) {
          const filename = tool.input.file_path.split('/').pop() || tool.input.file_path
          detail = filename
        } else if (tool.name === 'Read' && tool.input.file_path) {
          const filename = tool.input.file_path.split('/').pop() || tool.input.file_path
          detail = filename
        } else if (tool.name === 'Write' && tool.input.file_path) {
          const filename = tool.input.file_path.split('/').pop() || tool.input.file_path
          detail = filename
        } else if (tool.name === 'Bash' && tool.input.command) {
          detail = tool.input.command.length > 50
            ? tool.input.command.substring(0, 50) + '...'
            : tool.input.command
        } else if (tool.name === 'Grep' && tool.input.pattern) {
          detail = tool.input.pattern
        }
      }

      return {
        name: tool.name,
        detail,
        input: tool.input
      }
    })
  }

  // Fall back to legacy single toolUse string
  if (props.message.toolUse) {
    return [{
      name: props.message.toolUse,
      detail: '',
      input: null
    }]
  }

  return []
})

const copyButtonText = computed(() => {
  return copied.value ? 'Copied!' : 'Copy'
})

// Methods
const toggleToolExpand = (idx: number) => {
  expandedTools.value[idx] = !expandedTools.value[idx]
}

const formatToolInput = (input: any): string => {
  if (!input) return 'No input data'

  try {
    return JSON.stringify(input, null, 2)
  } catch (e) {
    return String(input)
  }
}

const copyContent = async () => {
  if (!textContent.value) return

  try {
    await navigator.clipboard.writeText(textContent.value)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}
</script>

<style scoped>
.message-detail-modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  padding: 20px;
}

.message-detail-modal {
  background: var(--bg-primary);
  border-radius: 16px;
  border: 1px solid var(--border-color);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-width: 900px;
  width: 100%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
}

.badges {
  display: flex;
  gap: 6px;
}

.badge {
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
}

.badge.historical {
  background: rgba(139, 92, 246, 0.1);
  color: var(--accent-purple);
}

.badge.error {
  background: rgba(220, 53, 69, 0.1);
  color: #dc3545;
}

.badge.tool-result {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

.badge.execution {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.badge.permission {
  background: rgba(251, 191, 36, 0.1);
  color: #fbbf24;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-secondary);
  padding: 8px;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.modal-body {
  overflow-y: auto;
  padding: 24px;
  flex: 1;
}

/* Metadata Section */
.metadata-section {
  margin-bottom: 24px;
}

.metadata-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  background: var(--bg-secondary);
  padding: 16px;
  border-radius: 12px;
}

.metadata-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metadata-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--text-secondary);
  letter-spacing: 0.5px;
}

.metadata-value {
  font-size: 0.95rem;
  color: var(--text-primary);
}

.metadata-value.monospace {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.85rem;
}

.metadata-value.role-user {
  color: var(--accent-purple);
  font-weight: 600;
}

.metadata-value.role-assistant {
  color: var(--accent-purple);
  font-weight: 600;
}

/* Section Headers */
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  cursor: pointer;
  user-select: none;
}

.section-header h4 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-header svg {
  color: var(--accent-purple);
}

/* Thinking Section */
.thinking-section {
  margin-bottom: 24px;
}

.thinking-content {
  background: var(--bg-secondary);
  padding: 16px;
  border-radius: 12px;
  border-left: 3px solid var(--accent-purple);
  font-size: 0.9rem;
  line-height: 1.6;
  color: var(--text-secondary);
  font-style: italic;
}

/* Content Section */
.content-section {
  margin-bottom: 24px;
}

.copy-btn {
  background: none;
  border: 1px solid var(--border-color);
  padding: 6px 12px;
  border-radius: 8px;
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 0.85rem;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}

.copy-btn:hover {
  background: var(--accent-purple);
  color: white;
  border-color: var(--accent-purple);
}

.message-content {
  background: var(--card-bg);
  padding: 16px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  font-size: 0.95rem;
  line-height: 1.6;
  color: var(--text-primary);
}

.message-content :deep(code) {
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.9em;
}

.message-content :deep(pre) {
  background: var(--bg-secondary);
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}

/* Images Section */
.images-section {
  margin-bottom: 24px;
}

.images-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.modal-image {
  width: 100%;
  height: 200px;
  object-fit: cover;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: all 0.2s;
}

.modal-image:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* Tools Section */
.tools-section {
  margin-bottom: 24px;
}

.tools-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tool-item {
  background: var(--bg-secondary);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  overflow: hidden;
  transition: all 0.2s;
}

.tool-item.expanded {
  border-color: var(--accent-purple);
}

.tool-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s;
}

.tool-header:hover {
  background: var(--card-bg);
}

.tool-name-section {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.tool-name-section svg {
  color: var(--accent-purple);
}

.tool-name {
  font-weight: 600;
  color: var(--text-primary);
}

.tool-detail-badge {
  font-size: 0.85rem;
  color: var(--text-secondary);
  background: var(--bg-primary);
  padding: 2px 8px;
  border-radius: 6px;
}

.expand-icon {
  color: var(--text-secondary);
}

.tool-body {
  padding: 0 16px 16px;
}

.tool-input h5 {
  margin: 0 0 8px 0;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.tool-input-json {
  background: var(--bg-primary);
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.85rem;
  color: var(--text-primary);
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}

.tool-no-input {
  color: var(--text-secondary);
  font-style: italic;
}

/* Transitions */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

.expand-enter-active, .expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.expand-enter-from, .expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.expand-enter-to, .expand-leave-from {
  opacity: 1;
  max-height: 1000px;
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .message-detail-modal-backdrop {
    padding: 0;
  }

  .message-detail-modal {
    max-width: 100%;
    max-height: 100vh;
    border-radius: 0;
  }

  .metadata-grid {
    grid-template-columns: 1fr;
  }

  .images-grid {
    grid-template-columns: 1fr;
  }
}
</style>
