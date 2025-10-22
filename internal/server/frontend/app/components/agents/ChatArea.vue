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
        <transition name="preview-slide">
          <div v-if="attachedImages.length > 0" class="image-previews">
            <div class="preview-header">
              <span class="preview-count">{{ attachedImages.length }} image{{ attachedImages.length > 1 ? 's' : '' }} attached</span>
              <button @click="clearAllImages" class="clear-all-btn" type="button">
                Clear All
              </button>
            </div>
            <div class="preview-grid">
              <transition-group name="preview-item">
                <div
                  v-for="(img, idx) in attachedImages"
                  :key="`img-${idx}`"
                  class="preview-item"
                >
                  <img :src="img.dataUrl" :alt="`Preview ${idx + 1}`" class="preview-image" />
                  <button @click="removeImage(idx)" class="remove-btn" type="button" title="Remove image">
                    ×
                  </button>
                  <span class="image-info">{{ truncateFileName(img.fileName) }} ({{ formatSize(img.size) }})</span>
                </div>
              </transition-group>
            </div>
          </div>
        </transition>

        <!-- Input Container -->
        <div class="input-container" :class="{ 'drag-over': isDragging, 'focused': isFocused }">
          <!-- Character Counter & Upload Button -->
          <div class="input-toolbar">
            <button
              @click="triggerFileUpload"
              class="btn-upload"
              type="button"
              :disabled="!connected"
              title="Attach images (PNG, JPEG, GIF, WebP)"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                <circle cx="8.5" cy="8.5" r="1.5"></circle>
                <polyline points="21 15 16 10 5 21"></polyline>
              </svg>
              <span class="upload-text">Attach Image</span>
            </button>
            <input
              ref="fileInput"
              type="file"
              accept="image/png,image/jpeg,image/gif,image/webp"
              multiple
              @change="handleFileSelect"
              style="display: none"
            />
            <span class="char-counter" :class="{ 'warning': charCount > 4000 }">
              {{ charCount }} / 5000 characters
            </span>
          </div>

          <!-- Textarea & Send Button Row -->
          <div class="input-row">
            <textarea
              ref="messageInput"
              :value="inputMessage"
              @input="handleInput"
              @keydown.enter="handleEnter"
              @paste="handlePaste"
              @drop.prevent="handleDrop"
              @dragover.prevent="isDragging = true"
              @dragleave="isDragging = false"
              @focus="isFocused = true"
              @blur="isFocused = false"
              placeholder="Type your message or paste/drop an image... (Enter to send, Shift+Enter for new line)"
              class="message-input"
              :disabled="!connected"
              :maxlength="5000"
              rows="3"
            ></textarea>
            <button
              @click="$emit('send')"
              class="btn-send"
              :disabled="(!inputMessage.trim() && attachedImages.length === 0) || !connected"
              title="Send message (Enter)"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="22" y1="2" x2="11" y2="13"></line>
                <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
              </svg>
            </button>
          </div>

          <!-- Drag & Drop Overlay -->
          <transition name="fade">
            <div v-if="isDragging" class="drag-overlay">
              <div class="drag-content">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                  <circle cx="8.5" cy="8.5" r="1.5"></circle>
                  <polyline points="21 15 16 10 5 21"></polyline>
                </svg>
                <p>Drop images here</p>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

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

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:input-message': [value: string]
  'send': []
  'images-attached': [images: AttachedImage[]]
}>()

const messagesContainer = ref<HTMLElement | null>(null)
const messageInput = ref<HTMLTextAreaElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const attachedImages = ref<AttachedImage[]>([])
const isDragging = ref(false)
const isFocused = ref(false)

// Allowed image formats
const ALLOWED_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp']
const MAX_SIZE = 3.75 * 1024 * 1024 // 3.75 MB

// Character count computed property
const charCount = computed(() => props.inputMessage.length)

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

// Handle input
function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:input-message', target.value)
}

// Handle Enter key
function handleEnter(event: KeyboardEvent) {
  if (event.shiftKey) {
    // Allow Shift+Enter to create new line (don't prevent default)
    return
  }
  // Enter without Shift sends the message
  event.preventDefault()
  emit('send')
}

// Trigger file upload dialog
function triggerFileUpload() {
  fileInput.value?.click()
}

// Handle file selection from input
async function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (!files) return

  for (const file of Array.from(files)) {
    if (file.type.startsWith('image/') && ALLOWED_TYPES.includes(file.type)) {
      await addImageFile(file)
    }
  }

  // Reset input to allow selecting the same file again
  target.value = ''
}

// Truncate filename for display
function truncateFileName(fileName: string): string {
  if (fileName.length <= 20) return fileName
  const extension = fileName.split('.').pop()
  const nameWithoutExt = fileName.substring(0, fileName.lastIndexOf('.'))
  return `${nameWithoutExt.substring(0, 12)}...${extension}`
}

// Clear all images
function clearAllImages() {
  attachedImages.value = []
  emit('images-attached', attachedImages.value)
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
  gap: 0;
  border-top: 1px solid var(--border-color);
  background: var(--card-bg);
  flex-shrink: 0;
}

/* Image Previews */
.image-previews {
  padding: 16px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.preview-count {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.clear-all-btn {
  padding: 4px 12px;
  background: rgba(220, 53, 69, 0.1);
  color: #dc3545;
  border: 1px solid rgba(220, 53, 69, 0.3);
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.clear-all-btn:hover {
  background: #dc3545;
  color: white;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 12px;
}

.preview-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: 8px;
  border: 2px solid var(--border-color);
  overflow: hidden;
  background: var(--bg-primary);
  transition: all 0.2s;
}

.preview-item:hover {
  border-color: var(--accent-purple);
  transform: scale(1.02);
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.remove-btn {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 28px;
  height: 28px;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  font-size: 20px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  opacity: 0;
}

.preview-item:hover .remove-btn {
  opacity: 1;
}

.remove-btn:hover {
  background: #dc3545;
  transform: scale(1.1);
}

.image-info {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 6px 8px;
  background: rgba(0, 0, 0, 0.85);
  color: white;
  font-size: 0.7rem;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Input Container */
.input-container {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 16px;
  transition: all 0.2s;
}

.input-container.focused {
  background: var(--bg-primary);
}

.input-container.drag-over {
  background: rgba(139, 92, 246, 0.05);
}

/* Input Toolbar */
.input-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.btn-upload {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-upload:hover:not(:disabled) {
  background: var(--accent-purple);
  color: white;
  border-color: var(--accent-purple);
  transform: translateY(-1px);
}

.btn-upload:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.upload-text {
  font-weight: 500;
}

.char-counter {
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 500;
  font-variant-numeric: tabular-nums;
}

.char-counter.warning {
  color: #ffc107;
  font-weight: 600;
}

/* Input Row */
.input-row {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.message-input {
  flex: 1;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 0.95rem;
  font-family: inherit;
  resize: vertical;
  min-height: 80px;
  max-height: 200px;
  transition: all 0.2s;
  line-height: 1.5;
}

.message-input:focus {
  outline: none;
  border-color: var(--accent-purple);
  background: var(--bg-primary);
  box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.1);
}

.message-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-send {
  padding: 12px 20px;
  background: linear-gradient(135deg, var(--accent-purple), var(--accent-purple-hover));
  color: white;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.3);
  height: 48px;
  min-width: 48px;
}

.btn-send:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(139, 92, 246, 0.4);
}

.btn-send:active:not(:disabled) {
  transform: translateY(0);
}

.btn-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

/* Drag & Drop Overlay */
.drag-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(139, 92, 246, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  z-index: 10;
}

.drag-content {
  text-align: center;
  color: white;
}

.drag-content svg {
  margin-bottom: 12px;
  opacity: 0.9;
}

.drag-content p {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
}

/* Transitions */
.preview-slide-enter-active,
.preview-slide-leave-active {
  transition: all 0.3s ease;
}

.preview-slide-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.preview-slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

.preview-item-enter-active,
.preview-item-leave-active {
  transition: all 0.3s ease;
}

.preview-item-enter-from {
  opacity: 0;
  transform: scale(0.8);
}

.preview-item-leave-to {
  opacity: 0;
  transform: scale(0.8);
}

.preview-item-move {
  transition: transform 0.3s ease;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Responsive */
@media (max-width: 768px) {
  .input-toolbar {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }

  .upload-text {
    display: none;
  }

  .preview-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  }
}
</style>
