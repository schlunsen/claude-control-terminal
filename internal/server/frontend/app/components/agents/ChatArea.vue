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
        <textarea
          ref="messageInput"
          :value="inputMessage"
          @input="$emit('update:input-message', ($event.target as HTMLTextAreaElement).value)"
          @keydown.enter.prevent="$emit('send')"
          placeholder="Type your message... (Enter to send)"
          class="message-input"
          :disabled="!connected"
          rows="3"
        ></textarea>
        <button
          @click="$emit('send')"
          class="btn-send"
          :disabled="!inputMessage.trim() || !connected"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="22" y1="2" x2="11" y2="13"></line>
            <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Props {
  hasActiveSession: boolean
  inputMessage: string
  connected: boolean
  isThinking: boolean
  isProcessing: boolean
}

defineProps<Props>()

defineEmits<{
  'update:input-message': [value: string]
  'send': []
}>()

const messagesContainer = ref<HTMLElement | null>(null)

defineExpose({
  messagesContainer
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
  gap: 12px;
  padding: 16px;
  border-top: 1px solid var(--border-color);
  background: var(--card-bg);
  flex-shrink: 0;
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
