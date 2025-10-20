<template>
  <div class="chat-main-area">
    <div v-if="!hasActiveSession" class="no-session-selected">
      <div class="empty-state">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.5">
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
        </svg>
        <p>Select a session or create a new one to start</p>
      </div>
    </div>

    <div v-else class="chat-content">
      <!-- Tool Overlays -->
      <slot name="tool-overlays"></slot>

      <!-- TodoWrite Box -->
      <slot name="todo-box"></slot>

      <!-- Messages Container -->
      <div class="messages-container" ref="messagesContainer">
        <slot name="messages"></slot>

        <!-- Thinking indicator -->
        <div v-if="isThinking" class="thinking-indicator">
          <div class="thinking-dots">
            <span></span>
            <span></span>
            <span></span>
          </div>
          Claude is thinking...
        </div>

        <!-- Processing indicator -->
        <div v-if="isProcessing && !isThinking" class="processing-indicator">
          <div class="processing-spinner"></div>
          Processing your message...
        </div>
      </div>

      <!-- Permission Requests -->
      <slot name="permissions"></slot>

      <!-- Tool Execution Bar -->
      <slot name="tool-execution"></slot>

      <!-- Input Area -->
      <div class="input-area">
        <!-- Image Preview Area -->
        <div v-if="attachedImages.length > 0" class="image-previews">
          <div
            v-for="(img, idx) in attachedImages"
            :key="idx"
            class="preview-item"
          >
            <img :src="img.dataUrl" :alt="`Preview ${idx + 1}`" class="preview-image" />
            <button @click="removeImage(idx)" class="remove-btn" type="button">
              ×
            </button>
            <span class="image-info">{{ img.fileName }} ({{ formatSize(img.size) }})</span>
          </div>
        </div>

        <!-- Input Container -->
        <div class="input-container" :class="{ 'drag-over': isDragging }">
          <textarea
            ref="messageInput"
            :value="inputMessage"
            @input="$emit('update:input-message', ($event.target as HTMLTextAreaElement).value)"
            @keydown.enter.prevent="handleEnter"
            @paste="handlePaste"
            @drop.prevent="handleDrop"
            @dragover.prevent="isDragging = true"
            @dragleave="isDragging = false"
            placeholder="Type your message or paste/drop an image... (Enter to send)"
            class="message-input"
            :disabled="!connected"
            rows="3"
          ></textarea>
          <button
            @click="$emit('send')"
            class="btn-send"
            :disabled="(!inputMessage.trim() && attachedImages.length === 0) || !connected"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="22" y1="2" x2="11" y2="13"></line>
              <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface AttachedImage {
  fileName: string
  mediaType: string
  size: number
  dataUrl: string
  base64Data: string
}

interface Props {
  hasActiveSession: boolean
  inputMessage: string
  connected: boolean
  isThinking: boolean
  isProcessing: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  'update:input-message': [value: string]
  'send': []
  'images-attached': [images: AttachedImage[]]
}>()

const messagesContainer = ref<HTMLElement | null>(null)
const messageInput = ref<HTMLTextAreaElement | null>(null)
const attachedImages = ref<AttachedImage[]>([])
const isDragging = ref(false)

// Allowed image formats
const ALLOWED_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp']
const MAX_SIZE = 3.75 * 1024 * 1024 // 3.75 MB

// Handle paste event
async function handlePaste(event: ClipboardEvent) {
  const items = event.clipboardData?.items
  if (!items) return

  for (const item of Array.from(items)) {
    if (item.type.startsWith('image/')) {
      event.preventDefault()
      const file = item.getAsFile()
      if (file) {
        await addImageFile(file)
      }
    }
  }
}

// Handle drop event
async function handleDrop(event: DragEvent) {
  isDragging.value = false
  const files = event.dataTransfer?.files
  if (!files) return

  for (const file of Array.from(files)) {
    if (file.type.startsWith('image/') && ALLOWED_TYPES.includes(file.type)) {
      await addImageFile(file)
    }
  }
}

// Add image file to attachments
async function addImageFile(file: File) {
  // Validate format
  if (!ALLOWED_TYPES.includes(file.type)) {
    console.error(`Unsupported image format: ${file.type}. Supported: PNG, JPEG, GIF, WebP`)
    return
  }

  // Validate size
  if (file.size > MAX_SIZE) {
    console.error(`Image too large: ${(file.size / 1024 / 1024).toFixed(2)} MB. Maximum: 3.75 MB`)
    return
  }

  try {
    // Convert to base64
    const base64 = await fileToBase64(file)

    const image: AttachedImage = {
      fileName: file.name,
      mediaType: file.type,
      size: file.size,
      dataUrl: `data:${file.type};base64,${base64}`,
      base64Data: base64
    }

    attachedImages.value.push(image)
    emit('images-attached', attachedImages.value)
  } catch (error) {
    console.error('Failed to process image:', error)
  }
}

// Convert file to base64
function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result as string
      // Extract base64 data (remove data:image/...;base64, prefix)
      const base64 = result.split(',')[1]
      resolve(base64)
    }
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

// Remove image from attachments
function removeImage(idx: number) {
  attachedImages.value.splice(idx, 1)
  emit('images-attached', attachedImages.value)
}

// Format file size
function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

// Handle Enter key
function handleEnter(event: KeyboardEvent) {
  if (!event.shiftKey) {
    emit('send')
  }
}

// Clear attachments (called from parent)
function clearAttachments() {
  attachedImages.value = []
}

defineExpose({
  messagesContainer,
  messageInput,
  attachedImages,
  clearAttachments
})
</script>

<style scoped>
.chat-main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  overflow: hidden;
  min-height: 0;
}

.no-session-selected {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-state {
  text-align: center;
  color: var(--text-secondary);
}

.empty-state svg {
  margin-bottom: 16px;
}

.empty-state p {
  font-size: 0.95rem;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
  position: relative;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  min-height: 0;
}

.thinking-indicator,
.processing-indicator {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-bottom: 16px;
}

.thinking-dots {
  display: flex;
  gap: 6px;
}

.thinking-dots span {
  width: 8px;
  height: 8px;
  background: var(--accent-purple);
  border-radius: 50%;
  animation: thinking 1.4s infinite ease-in-out both;
}

.thinking-dots span:nth-child(1) {
  animation-delay: -0.32s;
}

.thinking-dots span:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes thinking {
  0%, 80%, 100% {
    transform: scale(0.6);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

.processing-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-purple);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Input Area */
.input-area {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  border-top: 1px solid var(--border-color);
  background: var(--card-bg);
  flex-shrink: 0;
}

/* Image Previews */
.image-previews {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding-bottom: 8px;
}

.preview-item {
  position: relative;
  width: 120px;
  height: 120px;
  border-radius: 8px;
  border: 2px solid var(--border-color);
  overflow: hidden;
  background: var(--bg-secondary);
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.remove-btn {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 24px;
  height: 24px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.remove-btn:hover {
  background: rgba(255, 0, 0, 0.8);
}

.image-info {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 4px;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  font-size: 0.7rem;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Input Container */
.input-container {
  display: flex;
  gap: 12px;
  transition: all 0.2s;
}

.input-container.drag-over {
  background: rgba(138, 107, 255, 0.1);
  border-radius: 8px;
  padding: 4px;
}

.message-input {
  flex: 1;
  padding: 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 0.95rem;
  font-family: inherit;
  resize: none;
  transition: border-color 0.2s;
}

.message-input:focus {
  outline: none;
  border-color: var(--accent-purple);
}

.message-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-send {
  padding: 12px 20px;
  background: var(--accent-purple);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  align-self: flex-end;
}

.btn-send:hover:not(:disabled) {
  background: var(--accent-purple-hover);
  transform: translateY(-1px);
}

.btn-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
