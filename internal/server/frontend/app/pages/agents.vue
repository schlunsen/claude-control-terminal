<template>
  <div class="agents-page">
    <!-- Header -->
    <header class="header">
      <h1>Live Agents</h1>
      <p class="subtitle">Interactive Claude agent conversations</p>
    </header>

    <!-- Connection Status -->
    <div class="connection-status" :class="{ connected: agentWs.connected }">
      <div class="status-indicator"></div>
      <span>{{ agentWs.connected ? 'Connected' : 'Disconnected' }}</span>
    </div>

    <div class="agents-container">
      <!-- Sessions Sidebar -->
      <aside class="sessions-sidebar">
        <div class="sidebar-header">
          <h3>Sessions</h3>
          <button @click="createNewSession" class="btn-new-session" :disabled="!agentWs.connected">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"></line>
              <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
            New Session
          </button>
        </div>

        <div class="sessions-list">
          <div v-if="sessions.length === 0" class="no-sessions">
            No active sessions
          </div>
          <div
            v-for="session in sessions"
            :key="session.id"
            class="session-item"
            :class="{ active: activeSessionId === session.id }"
            @click="selectSession(session.id)"
          >
            <div class="session-info">
              <div class="session-name">Session {{ session.id.slice(0, 8) }}</div>
              <div class="session-meta">
                <span class="session-status" :class="session.status">{{ session.status }}</span>
                <span class="session-messages">{{ session.message_count }} messages</span>
              </div>
            </div>
            <button @click.stop="endSession(session.id)" class="btn-end-session" title="End session">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>
        </div>
      </aside>

      <!-- Chat Area -->
      <main class="chat-area">
        <div v-if="!activeSessionId" class="no-session-selected">
          <div class="empty-state">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.5">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
            </svg>
            <p>Select a session or create a new one to start</p>
          </div>
        </div>

        <div v-else class="chat-content">
          <!-- Messages -->
          <div class="messages-container" ref="messagesContainer">
            <div v-for="message in activeMessages" :key="message.id" class="message" :class="message.role">
              <div class="message-header">
                <span class="message-role">{{ message.role === 'user' ? 'You' : 'Claude' }}</span>
                <span class="message-time">{{ formatTime(message.timestamp) }}</span>
              </div>
              <div class="message-content" v-html="formatMessage(message.content)"></div>

              <!-- Tool use indicator -->
              <div v-if="message.toolUse" class="tool-use">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                </svg>
                Using {{ message.toolUse }}
              </div>
            </div>

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

          <!-- Input Area -->
          <div class="input-area">
            <textarea
              ref="messageInput"
              v-model="inputMessage"
              @keydown.enter.prevent="sendMessage"
              placeholder="Type your message... (Enter to send)"
              class="message-input"
              :disabled="!agentWs.connected"
              rows="3"
            ></textarea>
            <button
              @click="sendMessage"
              class="btn-send"
              :disabled="!inputMessage.trim() || !agentWs.connected"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="22" y1="2" x2="11" y2="13"></line>
                <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
              </svg>
            </button>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { useAgentWebSocket } from '~/composables/useAgentWebSocket'

// WebSocket connection
const agentWs = useAgentWebSocket()

// Refs
const messageInput = ref(null)

// State
const sessions = ref([])
const activeSessionId = ref(null)
const messages = ref({}) // { sessionId: [...messages] }
const inputMessage = ref('')
const isProcessing = ref(false)
const isThinking = ref(false)

// Computed
const activeMessages = computed(() => {
  if (!activeSessionId.value) return []
  return messages.value[activeSessionId.value] || []
})

// Message formatting
const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  })
}

const formatMessage = (content) => {
  // Skip system messages and JSON-like content
  if (typeof content === 'string' && content.includes('SystemMessage(') || content.startsWith('{') && content.includes('"type"')) {
    return '<em class="system-message">Processing...</em>'
  }

  // Clean up the content
  let cleanContent = content

  // If it's a string representation of an object, try to extract meaningful text
  if (typeof cleanContent === 'string' && cleanContent.includes('assistant:')) {
    const match = cleanContent.match(/assistant:\s*(.+?)(?:\n|$)/i)
    if (match) {
      cleanContent = match[1]
    }
  }

  // Convert markdown to HTML (basic)
  return cleanContent
    .replace(/```(.*?)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/\n/g, '<br>')
}

// Session management
const createNewSession = async () => {
  if (!agentWs.connected) return

  const sessionId = crypto.randomUUID()

  agentWs.send({
    type: 'create_session',
    session_id: sessionId,
    options: {
      tools: ['Read', 'Write', 'Edit', 'Bash', 'Search'],
      system_prompt: 'You are a helpful AI assistant.'
    }
  })
}

const selectSession = (sessionId) => {
  activeSessionId.value = sessionId

  // Focus the input when switching to a session
  focusMessageInput()
}

const endSession = async (sessionId) => {
  if (!agentWs.connected) return

  agentWs.send({
    type: 'end_session',
    session_id: sessionId
  })

  // Remove from local state
  sessions.value = sessions.value.filter(s => s.id !== sessionId)
  delete messages.value[sessionId]

  if (activeSessionId.value === sessionId) {
    activeSessionId.value = null
  }
}

// Messaging
const sendMessage = async () => {
  if (!inputMessage.value.trim() || !activeSessionId.value) return

  const message = inputMessage.value
  inputMessage.value = ''

  // Add user message to chat
  if (!messages.value[activeSessionId.value]) {
    messages.value[activeSessionId.value] = []
  }

  messages.value[activeSessionId.value].push({
    id: crypto.randomUUID(),
    role: 'user',
    content: message,
    timestamp: new Date()
  })

  isProcessing.value = true

  // Send to agent
  agentWs.send({
    type: 'send_prompt',
    session_id: activeSessionId.value,
    prompt: message
  })
}

// Helper function to focus message input
const focusMessageInput = () => {
  nextTick(() => {
    if (messageInput.value && !messageInput.value.disabled) {
      messageInput.value.focus()
    }
  })
}

// WebSocket event handlers
agentWs.on('onSessionCreated', (data) => {
  console.log('Session created handler called:', data)
  sessions.value.push(data.session)
  activeSessionId.value = data.session_id
  messages.value[data.session_id] = []

  // Focus the input after session creation
  focusMessageInput()
})

agentWs.on('onAgentMessage', (data) => {
  if (!messages.value[data.session_id]) {
    messages.value[data.session_id] = []
  }

  // Skip system messages and empty content (unless it's a completion signal)
  if (!data.complete && (!data.content ||
      data.content.includes('SystemMessage') ||
      data.content.includes('"type"') && data.content.includes('"subtype"'))) {
    return
  }

  // For streaming, use message_id or create one for this stream
  const messageId = data.message_id || 'stream-' + data.session_id

  const existingMessage = messages.value[data.session_id].find(
    m => (m.id === messageId || m.streaming) && m.role === 'assistant'
  )

  if (existingMessage && !data.complete) {
    // Append to existing message (streaming)
    existingMessage.content += data.content
    // Reset processing when we receive content
    if (data.content) {
      isProcessing.value = false
      isThinking.value = false
    }
  } else if (!existingMessage && data.content) {
    // New message - only add if there's actual content
    messages.value[data.session_id].push({
      id: messageId,
      role: 'assistant',
      content: data.content,
      timestamp: new Date(),
      streaming: !data.complete
    })
    // Reset processing when we receive content
    isProcessing.value = false
    isThinking.value = false
  }

  if (data.complete) {
    // Mark streaming as complete
    const msg = messages.value[data.session_id].find(m => m.streaming)
    if (msg) {
      msg.streaming = false
    }
    // Ensure processing is reset on completion
    isProcessing.value = false
    isThinking.value = false

    // Focus the input after Claude completes the response
    if (data.session_id === activeSessionId.value) {
      focusMessageInput()
    }
  }

  // Auto-scroll to bottom
  nextTick(() => {
    const container = document.querySelector('.messages-container')
    if (container) {
      container.scrollTop = container.scrollHeight
    }
  })
})

agentWs.on('onAgentThinking', (data) => {
  if (data.session_id === activeSessionId.value) {
    isThinking.value = data.thinking
    // When thinking stops, ensure processing is also reset
    if (!data.thinking) {
      isProcessing.value = false
    }
  }
})

agentWs.on('onAgentToolUse', (data) => {
  if (!messages.value[data.session_id]) return

  const lastMessage = messages.value[data.session_id][messages.value[data.session_id].length - 1]
  if (lastMessage && lastMessage.role === 'assistant') {
    lastMessage.toolUse = data.tool
  }
})

agentWs.on('onError', (data) => {
  console.error('Agent error:', data.message)
  // Always reset on error
  isProcessing.value = false
  isThinking.value = false

  // Show error message to user
  if (data.session_id && messages.value[data.session_id]) {
    messages.value[data.session_id].push({
      id: crypto.randomUUID(),
      role: 'assistant',
      content: `⚠️ Error: ${data.message}`,
      timestamp: new Date(),
      isError: true
    })
  }

  // Focus input after error so user can retry
  if (data.session_id === activeSessionId.value) {
    focusMessageInput()
  }
})

agentWs.on('onSessionsList', (data) => {
  console.log('Sessions list received:', data)
  sessions.value = data.sessions
})

// Load existing sessions on mount
onMounted(() => {
  if (agentWs.connected) {
    agentWs.send({ type: 'list_sessions' })
  }
})

// Watch for connection changes
watch(() => agentWs.connected, (connected) => {
  if (connected) {
    agentWs.send({ type: 'list_sessions' })
  }
})
</script>

<style scoped>
.agents-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
}

.header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-bg);
}

.header h1 {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
}

.subtitle {
  margin: 4px 0 0 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.connection-status {
  position: absolute;
  top: 24px;
  right: 24px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: var(--bg-secondary);
  border-radius: 20px;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #dc3545;
}

.connection-status.connected .status-indicator {
  background: #28a745;
}

.agents-container {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* Sessions Sidebar */
.sessions-sidebar {
  width: 300px;
  background: var(--card-bg);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
}

.sidebar-header h3 {
  margin: 0 0 12px 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.btn-new-session {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px;
  background: var(--accent-purple);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-new-session:hover:not(:disabled) {
  background: var(--accent-purple-hover);
}

.btn-new-session:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.sessions-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.no-sessions {
  text-align: center;
  padding: 32px 16px;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  margin-bottom: 8px;
  background: var(--bg-secondary);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.session-item:hover {
  background: var(--bg-tertiary);
}

.session-item.active {
  background: var(--accent-purple);
  color: white;
}

.session-info {
  flex: 1;
  overflow: hidden;
}

.session-name {
  font-size: 0.9rem;
  font-weight: 500;
  margin-bottom: 4px;
}

.session-meta {
  display: flex;
  gap: 12px;
  font-size: 0.8rem;
  opacity: 0.8;
}

.session-status {
  text-transform: capitalize;
}

.session-status.processing {
  color: var(--accent-blue);
}

.session-status.error {
  color: #dc3545;
}

.btn-end-session {
  padding: 4px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.btn-end-session:hover {
  opacity: 1;
}

.session-item.active .btn-end-session {
  color: white;
}

/* Chat Area */
.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  overflow: hidden; /* Prevent overflow of the entire chat area */
  min-height: 0; /* Important for flex children */
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
  overflow: hidden; /* Contain children */
  min-height: 0; /* Important for flex children */
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  min-height: 0; /* Important for proper scrolling */
}

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

.message-content code {
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.9em;
}

.message.user .message-content code {
  background: rgba(255, 255, 255, 0.2);
}

.message-content pre {
  background: var(--bg-secondary);
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}

.message-content .system-message {
  color: var(--text-secondary);
  font-style: italic;
  opacity: 0.7;
}

.message.isError .message-content {
  background: rgba(220, 53, 69, 0.1);
  border-color: rgba(220, 53, 69, 0.3);
  color: #dc3545;
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

.thinking-indicator {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  margin: 0 24px 24px 24px;
  background: linear-gradient(135deg, var(--accent-purple), var(--accent-purple-hover));
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.2);
  color: white;
  font-size: 0.95rem;
  font-weight: 500;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.thinking-dots {
  display: flex;
  gap: 6px;
}

.thinking-dots span {
  width: 10px;
  height: 10px;
  background: white;
  border-radius: 50%;
  animation: pulse 1.5s infinite ease-in-out;
  box-shadow: 0 0 8px rgba(255, 255, 255, 0.4);
}

.thinking-dots span:nth-child(1) {
  animation-delay: 0s;
}

.thinking-dots span:nth-child(2) {
  animation-delay: 0.15s;
}

.thinking-dots span:nth-child(3) {
  animation-delay: 0.3s;
}

@keyframes pulse {
  0%, 80%, 100% {
    opacity: 0.4;
    transform: scale(0.8) translateY(0);
  }
  40% {
    opacity: 1;
    transform: scale(1.2) translateY(-3px);
  }
}

.processing-indicator {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--card-bg);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-size: 0.9rem;
  opacity: 0.7;
}

.processing-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-purple);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Input Area */
.input-area {
  display: flex;
  gap: 12px;
  padding: 16px 24px;
  background: var(--card-bg);
  border-top: 1px solid var(--border-color);
  flex-shrink: 0; /* Never shrink the input area */
}

.message-input {
  flex: 1;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  font-size: 0.95rem;
  color: var(--text-primary);
  resize: none;
  font-family: inherit;
  transition: all 0.2s;
}

.message-input:focus {
  outline: none;
  border-color: var(--accent-purple);
}

.message-input:disabled {
  opacity: 0.5;
}

.btn-send {
  padding: 12px 20px;
  background: var(--accent-purple);
  color: white;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-send:hover:not(:disabled) {
  background: var(--accent-purple-hover);
}

.btn-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Dark mode adjustments */
:root[data-theme="dark"] {
  --accent-purple-hover: #7c3aed;
  --bg-tertiary: #2a2a2a;
}

/* Responsive */
@media (max-width: 768px) {
  .sessions-sidebar {
    width: 240px;
  }
}

@media (max-width: 640px) {
  .agents-container {
    flex-direction: column;
  }

  .sessions-sidebar {
    width: 100%;
    height: 200px;
    border-right: none;
    border-bottom: 1px solid var(--border-color);
  }
}
</style>