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
            <div class="toolbar-right">
              <transition name="fade">
                <span v-if="isProcessing" class="interrupt-hint">
                  <kbd>ESC</kbd> to interrupt
                </span>
              </transition>
              <span class="char-counter" :class="{ 'warning': charCount > 4000 }">
                {{ charCount }} / 5000 characters
              </span>
            </div>
          </div>

          <!-- Textarea & Buttons Row -->
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
              placeholder="Type your message or paste/drop an image... (Enter to send, Shift+Enter for new line, ⇧⌥⌘R to record)"
              class="message-input"
              :disabled="!connected"
              :maxlength="5000"
              rows="3"
            ></textarea>
            <div class="button-group">
              <button
                @click="startVoiceRecording"
                class="btn-record"
                :disabled="!connected"
                title="Record voice message (⇧⌥⌘R)"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"></path>
                  <path d="M19 10v2a7 7 0 0 1-14 0v-2"></path>
                  <line x1="12" y1="19" x2="12" y2="23"></line>
                  <line x1="8" y1="23" x2="16" y2="23"></line>
                </svg>
              </button>
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

    <!-- Recording Modal -->
    <transition name="modal-fade">
      <div v-if="showRecordingModal" class="recording-modal-overlay" @click="cancelVoiceRecording">
        <div class="recording-modal" @click.stop>
          <div class="modal-header">
            <h3>Voice Recording</h3>
            <button @click="cancelVoiceRecording" class="close-btn" title="Cancel">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>

          <div class="modal-body">
            <!-- Recording Visualization -->
            <div class="recording-visualization" v-if="voiceRecording.isRecording.value && !whisperTranscription.isTranscribing.value">
              <div class="pulse-ring"></div>
              <div class="pulse-ring-2"></div>
              <div class="microphone-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1">
                  <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"></path>
                  <path d="M19 10v2a7 7 0 0 1-14 0v-2"></path>
                  <line x1="12" y1="19" x2="12" y2="23"></line>
                  <line x1="8" y1="23" x2="16" y2="23"></line>
                </svg>
              </div>
            </div>

            <!-- Transcribing State -->
            <div v-if="whisperTranscription.isTranscribing.value" class="transcribing-state">
              <div class="spinner"></div>
              <p class="status-text">Transcribing...</p>
            </div>

            <!-- Model Loading State -->
            <div v-if="whisperTranscription.isModelLoading.value" class="loading-state">
              <div class="loading-icon">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                  <polyline points="7 10 12 15 17 10"></polyline>
                  <line x1="12" y1="15" x2="12" y2="3"></line>
                </svg>
              </div>
              <h3 class="loading-title">Preparing AI Model</h3>
              <p class="status-text">Downloading Whisper speech recognition model...</p>
              <p class="status-subtext">This only happens once, then it's cached for instant use</p>
              <div class="progress-bar-container">
                <div class="progress-bar">
                  <div class="progress-fill" :style="{ width: whisperTranscription.transcriptionProgress.value + '%' }"></div>
                </div>
                <p class="progress-text">{{ whisperTranscription.transcriptionProgress.value }}%</p>
              </div>
            </div>

            <!-- Recording Status -->
            <div v-if="!whisperTranscription.isTranscribing.value && !whisperTranscription.isModelLoading.value" class="recording-status">
              <p class="status-text">
                {{ voiceRecording.isRecording.value ? 'Recording...' : 'Ready to record' }}
              </p>
              <p class="duration">{{ voiceRecording.formatDuration(voiceRecording.duration.value) }}</p>
              <p v-if="voiceRecording.isRecording.value" class="status-hint">Press Space to stop recording</p>
            </div>

            <!-- Error Display -->
            <div v-if="voiceRecording.error.value || whisperTranscription.error.value" class="error-message">
              {{ voiceRecording.error.value || whisperTranscription.error.value }}
            </div>
          </div>

          <div class="modal-footer">
            <button
              v-if="voiceRecording.isRecording.value"
              @click="finishRecording"
              class="btn-modal btn-stop"
              :disabled="whisperTranscription.isTranscribing.value"
            >
              Stop & Transcribe (Space)
            </button>
            <button
              @click="cancelVoiceRecording"
              class="btn-modal btn-cancel"
              :disabled="whisperTranscription.isModelLoading.value"
            >
              Cancel (Esc)
            </button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { useVoiceRecording } from '~/composables/useVoiceRecording'
import { useWhisperTranscription } from '~/composables/useWhisperTranscription'

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
  hasModalOpen?: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:input-message': [value: string]
  'send': []
  'interrupt': []
  'images-attached': [images: AttachedImage[]]
}>()

const messagesContainer = ref<HTMLElement | null>(null)
const messageInput = ref<HTMLTextAreaElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const attachedImages = ref<AttachedImage[]>([])
const isDragging = ref(false)
const isFocused = ref(false)

// Voice recording composables
const voiceRecording = useVoiceRecording()
const whisperTranscription = useWhisperTranscription()

// Recording state
const showRecordingModal = ref(false)

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

// Voice recording functions
async function startVoiceRecording() {
  try {
    // Show modal first
    showRecordingModal.value = true

    // Preload the Whisper model if not already loaded
    await whisperTranscription.initializeModel()

    // Auto-start recording immediately after model is loaded
    await voiceRecording.startRecording()
  } catch (error) {
    console.error('Failed to initialize voice recording:', error)
    showRecordingModal.value = false
  }
}

function stopVoiceRecording() {
  voiceRecording.stopRecording()
}

// Keyboard handler for shortcuts
function handleKeydown(event: KeyboardEvent) {
  // Shift + Option/Alt + Command/Ctrl + R to start recording
  if (event.code === 'KeyR' && event.shiftKey && event.altKey && (event.metaKey || event.ctrlKey) && !showRecordingModal.value && props.connected) {
    event.preventDefault()
    startVoiceRecording()
    return
  }

  // Space key to stop recording (only when modal is open)
  if (event.code === 'Space' && showRecordingModal.value && voiceRecording.isRecording.value) {
    event.preventDefault()
    finishRecording()
    return
  }

  // Escape to cancel recording (when modal is open)
  if (event.code === 'Escape' && showRecordingModal.value) {
    event.preventDefault()
    cancelVoiceRecording()
    return
  }

  // Escape to interrupt processing (when not in any modal)
  if (event.code === 'Escape' && !showRecordingModal.value && !props.hasModalOpen && props.isProcessing && props.connected) {
    event.preventDefault()
    emit('interrupt')
    return
  }
}

function cancelVoiceRecording() {
  voiceRecording.cancelRecording()
  showRecordingModal.value = false
}

async function finishRecording() {
  try {
    // Stop recording
    stopVoiceRecording()

    // Wait a bit for the blob to be ready
    await new Promise(resolve => setTimeout(resolve, 100))

    const audioBlob = voiceRecording.audioBlob.value
    if (!audioBlob) {
      console.error('No audio recorded')
      voiceRecording.error.value = 'No audio recorded'
      return
    }

    console.log('Audio blob ready:', audioBlob.size, 'bytes')

    // Transcribe audio
    const text = await whisperTranscription.transcribe(audioBlob)

    console.log('Transcription result:', text)

    // Add transcribed text to input
    if (text) {
      const currentText = props.inputMessage
      const newText = currentText ? `${currentText} ${text}` : text
      console.log('Emitting update:input-message with:', newText)
      emit('update:input-message', newText)

      // Focus the input field after transcription
      await nextTick()
      if (messageInput.value) {
        messageInput.value.focus()
        // Move cursor to end
        const length = newText.length
        messageInput.value.setSelectionRange(length, length)
      }
    } else {
      console.warn('Transcription returned empty text')
      whisperTranscription.error.value = 'No speech detected in audio'
    }

    // Reset and close modal
    voiceRecording.reset()
    showRecordingModal.value = false
  } catch (error) {
    console.error('Failed to transcribe audio:', error)
    whisperTranscription.error.value = error instanceof Error ? error.message : 'Transcription failed'
  }
}

// Add keyboard event listener for Space and Escape keys
onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

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

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
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

.interrupt-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.interrupt-hint kbd {
  display: inline-block;
  padding: 3px 8px;
  background: rgba(220, 53, 69, 0.1);
  border: 1px solid rgba(220, 53, 69, 0.3);
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.8rem;
  font-weight: 600;
  color: #dc3545;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
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

.button-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: stretch;
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
  width: 100%;
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

.btn-record {
  padding: 12px 20px;
  background: linear-gradient(135deg, #dc3545, #c82333);
  color: white;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(220, 53, 69, 0.3);
  height: 48px;
  width: 100%;
}

.btn-record:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(220, 53, 69, 0.4);
}

.btn-record:active:not(:disabled) {
  transform: translateY(0);
}

.btn-record:disabled {
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

/* Recording Modal */
.recording-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.recording-modal {
  background: var(--card-bg);
  border-radius: 16px;
  padding: 24px;
  min-width: 400px;
  max-width: 500px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  border: 1px solid var(--border-color);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.modal-body {
  min-height: 200px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 32px 16px;
}

.recording-visualization {
  position: relative;
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pulse-ring,
.pulse-ring-2 {
  position: absolute;
  width: 100%;
  height: 100%;
  border: 3px solid #dc3545;
  border-radius: 50%;
  animation: pulse 2s ease-out infinite;
  opacity: 0;
}

.pulse-ring-2 {
  animation-delay: 1s;
}

@keyframes pulse {
  0% {
    transform: scale(0.5);
    opacity: 0.8;
  }
  50% {
    opacity: 0.4;
  }
  100% {
    transform: scale(1.2);
    opacity: 0;
  }
}

.microphone-icon {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, #dc3545, #c82333);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  z-index: 1;
  box-shadow: 0 4px 20px rgba(220, 53, 69, 0.4);
}

.transcribing-state,
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 20px;
  text-align: center;
}

.loading-icon {
  color: var(--accent-purple);
  animation: bounce 2s ease-in-out infinite;
}

@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.loading-title {
  font-size: 1.3rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.status-subtext {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin: -8px 0 0 0;
  opacity: 0.8;
}

.progress-bar-container {
  width: 100%;
  max-width: 300px;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid var(--border-color);
  border-top-color: var(--accent-purple);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.recording-status {
  text-align: center;
}

.status-text {
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 0 0 8px 0;
  font-weight: 500;
}

.duration {
  font-size: 2rem;
  font-weight: 600;
  color: var(--accent-purple);
  margin: 0;
  font-variant-numeric: tabular-nums;
}

.keyboard-hint {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 12px 0 0 0;
  opacity: 0.9;
}

.keyboard-hint kbd {
  display: inline-block;
  padding: 3px 8px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--accent-purple);
  box-shadow: 0 2px 0 var(--border-color);
}

.status-hint {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin: 8px 0 0 0;
  opacity: 0.8;
  font-style: italic;
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: var(--bg-secondary);
  border-radius: 4px;
  overflow: hidden;
  margin-top: 8px;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--accent-purple), var(--accent-purple-hover));
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 4px 0 0 0;
}

.error-message {
  color: #dc3545;
  background: rgba(220, 53, 69, 0.1);
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.9rem;
  text-align: center;
}

.modal-footer {
  display: flex !important;
  flex-direction: column !important;
  gap: 12px;
  margin-top: 24px;
  align-items: stretch;
}

.btn-modal {
  width: 100% !important;
  padding: 14px 20px;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: block;
  flex-shrink: 0;
}

.btn-start {
  background: linear-gradient(135deg, #dc3545, #c82333);
  color: white;
}

.btn-start:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(220, 53, 69, 0.3);
}

.btn-stop {
  background: linear-gradient(135deg, var(--accent-purple), var(--accent-purple-hover));
  color: white;
}

.btn-stop:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
}

.btn-cancel {
  background: var(--bg-secondary);
  color: var(--text-secondary);
}

.btn-cancel:hover {
  background: var(--border-color);
  color: var(--text-primary);
}

.btn-modal:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Modal Transitions */
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.3s ease;
}

.modal-fade-enter-active .recording-modal,
.modal-fade-leave-active .recording-modal {
  transition: transform 0.3s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from .recording-modal {
  transform: scale(0.9) translateY(-20px);
}

.modal-fade-leave-to .recording-modal {
  transform: scale(0.9) translateY(-20px);
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

  .recording-modal {
    min-width: 90%;
    max-width: 90%;
  }
}
</style>
