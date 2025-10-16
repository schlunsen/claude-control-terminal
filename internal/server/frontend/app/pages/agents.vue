<template>
  <div class="agents-page">
    <!-- Header -->
    <header class="header">
      <div class="header-content">
        <div class="header-text">
          <h1>Live Agents</h1>
          <p class="subtitle">Interactive Claude agent conversations</p>
        </div>
        <div class="header-actions">
          <button
            @click="killAllAgents"
            class="btn-kill-all"
            :disabled="!agentWs.connected || sessions.length === 0"
            title="Kill all active agents"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="15" y1="9" x2="9" y2="15"></line>
              <line x1="9" y1="9" x2="15" y2="15"></line>
            </svg>
            Kill All Agents
          </button>
        </div>
      </div>
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
          <div class="session-buttons">
            <button @click="createNewSession" class="btn-new-session" :disabled="!agentWs.connected || creatingSession">
              <svg v-if="!creatingSession" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
              <div v-if="creatingSession" class="btn-spinner-small"></div>
              <span v-if="!creatingSession">New Session</span>
              <span v-else>Creating...</span>
            </button>
            <button @click="openResumeModal" class="btn-resume-session" :disabled="!agentWs.connected">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M3 15v4c0 1.1.9 2 2 2h14a2 2 0 0 0 2-2v-4M17 9l-5 5-5-5M12 12.8V2.5"/>
              </svg>
              Resume Session
            </button>
          </div>
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
            <div v-for="message in activeMessages" :key="message.id" class="message" :class="{
              [message.role]: true,
              isToolResult: message.isToolResult,
              isExecutionStatus: message.isExecutionStatus,
              isPermissionDecision: message.isPermissionDecision,
              isHistorical: message.isHistorical,
              isError: message.isError
            }">
              <div class="message-header">
                <span class="message-role">{{ message.role === 'user' ? 'You' : message.role === 'system' ? 'System' : 'Claude' }}</span>
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

          <!-- Permission Requests -->
          <div v-if="pendingPermissions.length > 0" class="permission-requests">
            <div
              v-for="permission in pendingPermissions"
              :key="permission.request_id"
              class="permission-request"
            >
              <div class="permission-header">
                <div class="permission-icon">🔐</div>
                <div class="permission-title">Permission Request</div>
                <div class="permission-time">{{ formatTime(permission.timestamp) }}</div>
              </div>
              <div class="permission-description">
                {{ permission.description }}
              </div>
              <div class="permission-actions">
                <button
                  @click="denyPermission(permission)"
                  class="btn-deny"
                  title="Deny this request"
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                  Deny
                </button>
                <button
                  @click="approvePermission(permission)"
                  class="btn-approve"
                  title="Approve this request"
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20,6 9,17 4,12"></polyline>
                  </svg>
                  Approve
                </button>
              </div>
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

    <!-- Create Session Modal -->
    <div v-if="showCreateSessionModal" class="modal-overlay" @click="showCreateSessionModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h2>Create New Session</h2>
          <button @click="showCreateSessionModal = false" class="modal-close">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label for="working-directory">Working Directory</label>
            <input
              id="working-directory"
              v-model="sessionForm.workingDirectory"
              type="text"
              placeholder="/Users/schlunsen/projects"
              class="form-input"
            />
            <small class="form-help">The directory where the agent will work</small>
          </div>

          <div class="form-group">
            <label for="permission-mode">Permission Mode</label>
            <select id="permission-mode" v-model="sessionForm.permissionMode" class="form-select">
              <option value="default">Default (ask for permissions)</option>
              <option value="acceptEdits">Allow All (full permissions)</option>
              <option value="plan">Read Only (no file modifications)</option>
            </select>
            <small class="form-help">Control what permissions the agent has</small>
          </div>

          <div class="form-group">
            <label for="system-prompt">System Prompt (optional)</label>
            <textarea
              id="system-prompt"
              v-model="sessionForm.systemPrompt"
              placeholder="You are a helpful AI assistant."
              class="form-textarea"
              rows="3"
            ></textarea>
            <small class="form-help">Custom instructions for the agent</small>
          </div>

          <div class="form-group">
            <label>Available Tools</label>
            <div class="tools-grid">
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Read" />
                <span>Read</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Write" />
                <span>Write</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Edit" />
                <span>Edit</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Bash" />
                <span>Bash</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Search" />
                <span>Search</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Grep" />
                <span>Grep</span>
              </label>
            </div>
          </div>

          <div class="modal-actions">
            <button @click="showCreateSessionModal = false" class="btn-cancel" :disabled="creatingSession">Cancel</button>
            <button @click="createSessionWithOptions" class="btn-create" :disabled="!sessionForm.workingDirectory || creatingSession">
              <div v-if="creatingSession" class="btn-spinner"></div>
              <span v-if="!creatingSession">Create Session</span>
              <span v-else>Creating...</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Resume Session Modal -->
    <div v-if="showResumeModal" class="modal-overlay" @click="showResumeModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h2>Resume Session</h2>
          <button @click="showResumeModal = false" class="modal-close">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div v-if="loadingSessions" class="loading-sessions">
            <div class="loading-spinner"></div>
            Loading available sessions...
          </div>
          <div v-else-if="availableSessions.length === 0" class="no-sessions-available">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.5">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
            </svg>
            <p>No previous sessions found</p>
          </div>
          <div v-else-if="!selectedResumeSession" class="sessions-list-modal">
            <div
              v-for="session in availableSessions"
              :key="session.conversation_id"
              class="session-item-modal"
              @click="selectSessionForResume(session)"
            >
              <div class="session-info-modal">
                <div class="session-name-modal">{{ session.session_name || 'Unnamed Session' }}</div>
                <div class="session-details">
                  <span class="session-directory">{{ session.working_directory || 'No directory' }}</span>
                  <span class="session-meta">
                    {{ session.total_messages }} messages •
                    {{ formatRelativeTime(session.last_activity) }}
                  </span>
                </div>
              </div>
              <div class="session-resume-indicator">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M5 12h14M12 5l7 7-7 7"/>
                </svg>
              </div>
            </div>
          </div>

          <!-- Resume Session Options -->
          <div v-else class="resume-session-options">
            <div class="selected-session-info">
              <h3>{{ selectedResumeSession.session_name || 'Selected Session' }}</h3>
              <p>Original working directory: <code>{{ selectedResumeSession.working_directory }}</code></p>
            </div>

            <div class="form-group">
              <label for="resume-working-directory">Working Directory</label>
              <input
                id="resume-working-directory"
                v-model="resumeForm.workingDirectory"
                type="text"
                :placeholder="selectedResumeSession.working_directory"
                class="form-input"
              />
              <small class="form-help">Directory where the agent will work (defaults to original)</small>
            </div>

            <div class="form-group">
              <label for="resume-permission-mode">Permission Mode</label>
              <select id="resume-permission-mode" v-model="resumeForm.permissionMode" class="form-select">
                <option value="default">Default (ask for permissions)</option>
                <option value="acceptEdits">Allow All (full permissions)</option>
                <option value="plan">Read Only (no file modifications)</option>
              </select>
              <small class="form-help">Control what permissions the agent has</small>
            </div>

            <div class="form-group">
              <label for="resume-system-prompt">System Prompt (optional)</label>
              <textarea
                id="resume-system-prompt"
                v-model="resumeForm.systemPrompt"
                placeholder="You are a helpful AI assistant."
                class="form-textarea"
                rows="3"
              ></textarea>
              <small class="form-help">Custom instructions for the agent</small>
            </div>

            <div class="form-group">
              <label>Available Tools</label>
              <div class="tools-grid">
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Read" />
                  <span>Read</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Write" />
                  <span>Write</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Edit" />
                  <span>Edit</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Bash" />
                  <span>Bash</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Search" />
                  <span>Search</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Grep" />
                  <span>Grep</span>
                </label>
              </div>
            </div>

            <div class="modal-actions">
              <button @click="selectedResumeSession = null" class="btn-cancel" :disabled="resumingSession">Back</button>
              <button @click="resumeSessionWithOptions" class="btn-create" :disabled="resumingSession">
                <div v-if="resumingSession" class="btn-spinner"></div>
                <span v-if="!resumingSession">Resume Session</span>
                <span v-else>Resuming...</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useAgentWebSocket } from '~/composables/useAgentWebSocket'
import { ref, computed, watch, nextTick, onMounted } from 'vue'

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
const showResumeModal = ref(false)
const availableSessions = ref([])
const loadingSessions = ref(false)
const showCreateSessionModal = ref(false)
const selectedResumeSession = ref(null)
const creatingSession = ref(false)
const resumingSession = ref(false)
const pendingPermissions = ref([]) // Array of pending permission requests
const awaitingToolResults = ref(new Set()) // Track sessions awaiting tool execution results

// Session creation form
const sessionForm = ref({
  workingDirectory: '/Users/schlunsen/projects',
  permissionMode: 'default',
  systemPrompt: '',
  tools: ['Read', 'Write', 'Edit', 'Bash', 'Search']
})

// Resume session form
const resumeForm = ref({
  workingDirectory: '/Users/schlunsen/projects',
  permissionMode: 'default',
  systemPrompt: '',
  tools: ['Read', 'Write', 'Edit', 'Bash', 'Search']
})

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
const createNewSession = () => {
  if (!agentWs.connected) return

  // Reset form to defaults
  sessionForm.value = {
    workingDirectory: '/Users/schlunsen/projects',
    permissionMode: 'default',
    systemPrompt: '',
    tools: ['Read', 'Write', 'Edit', 'Bash', 'Search']
  }

  showCreateSessionModal.value = true
}

const createSessionWithOptions = async () => {
  if (!agentWs.connected || !sessionForm.value.workingDirectory) return

  console.log('Starting session creation...')
  creatingSession.value = true

  try {
    const sessionId = crypto.randomUUID()
    console.log('Creating session with ID:', sessionId)

    agentWs.send({
      type: 'create_session',
      session_id: sessionId,
      options: {
        tools: sessionForm.value.tools,
        system_prompt: sessionForm.value.systemPrompt || 'You are a helpful AI assistant.',
        working_directory: sessionForm.value.workingDirectory,
        permission_mode: sessionForm.value.permissionMode
      }
    })
    console.log('Session creation message sent')

    showCreateSessionModal.value = false
    console.log('Create modal closed')
  } catch (error) {
    console.error('Failed to create session:', error)
    alert('Failed to create session. Please try again.')
  } finally {
    creatingSession.value = false
    console.log('creatingSession set to false')
  }
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
  awaitingToolResults.value.delete(sessionId)  // Clean up flag

  if (activeSessionId.value === sessionId) {
    activeSessionId.value = null
  }
}

// Resume session functionality
const loadAvailableSessions = async () => {
  loadingSessions.value = true
  try {
    const response = await $fetch('/api/prompts/sessions')
    availableSessions.value = response.sessions || []
  } catch (error) {
    console.error('Failed to load sessions:', error)
    availableSessions.value = []
  } finally {
    loadingSessions.value = false
  }
}

const openResumeModal = () => {
  console.log('Opening resume modal...')
  showResumeModal.value = true
}

const selectSessionForResume = async (session) => {
  console.log('Selected session for resume:', session)
  selectedResumeSession.value = session

  // Prefill the form with the session's data
  resumeForm.value = {
    workingDirectory: session.working_directory || '/Users/schlunsen/projects',
    permissionMode: 'default',
    systemPrompt: '',
    tools: ['Read', 'Write', 'Edit', 'Bash', 'Search']
  }
}

const resumeSessionWithOptions = async () => {
  try {
    if (!selectedResumeSession.value) return

    console.log('Starting session resume...')
    resumingSession.value = true

    // Fetch resume data from the backend
    console.log('Fetching resume data...')
    const resumeData = await $fetch(`/api/sessions/${selectedResumeSession.value.conversation_id}/resume-data`)
    console.log('Resume data fetched:', resumeData)

    // Create new agent session with history context and options
    const sessionId = crypto.randomUUID()
    console.log('Creating new session with ID:', sessionId)

    agentWs.send({
      type: 'create_session',
      session_id: sessionId,
      options: {
        tools: resumeForm.value.tools,
        system_prompt: resumeForm.value.systemPrompt || 'You are a helpful AI assistant.',
        working_directory: resumeForm.value.workingDirectory || resumeData.working_directory,
        permission_mode: resumeForm.value.permissionMode,
        conversation_history: resumeData.context,
        original_conversation_id: resumeData.conversation_id
      }
    })
    console.log('Session creation message sent')

    // Close the modal and reset selection
    showResumeModal.value = false
    selectedResumeSession.value = null

    // Add historical messages to the chat
    if (resumeData.messages && resumeData.messages.length > 0) {
      messages.value[sessionId] = []
      resumeData.messages.forEach(msg => {
        messages.value[sessionId].push({
          id: crypto.randomUUID(),
          role: 'user',
          content: msg.message,
          timestamp: new Date(msg.submitted_at),
          isHistorical: true
        })
      })
    }

    console.log('Session resume completed successfully')

  } catch (error) {
    console.error('Failed to resume session:', error)
    alert('Failed to resume session. Please try again.')
  } finally {
    resumingSession.value = false
    console.log('resumingSession set to false')
  }
}

const formatRelativeTime = (timestamp) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

// Watch for modal opening to load sessions
watch(showResumeModal, (show) => {
  if (show) {
    loadAvailableSessions()
  }
})

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

// Permission request functionality
const approvePermission = (request) => {
  sendPermissionResponse(request, true)
}

const denyPermission = (request, reason = '') => {
  sendPermissionResponse(request, false, reason)
}

const sendPermissionResponse = (request, approved, reason = '') => {
  try {
    agentWs.send({
      type: 'permission_response',
      session_id: request.session_id,
      request_id: request.request_id,
      approved: approved,
      reason: reason
    })

    // Remove from pending permissions
    pendingPermissions.value = pendingPermissions.value.filter(p => p.request_id !== request.request_id)

    // Add a system message to show the decision
    if (!messages.value[activeSessionId.value]) {
      messages.value[activeSessionId.value] = []
    }

    const decisionText = approved ? '✅ Approved' : '❌ Denied'
    const decisionMessage = reason ? `${decisionText} (Reason: ${reason})` : decisionText

    messages.value[activeSessionId.value].push({
      id: crypto.randomUUID(),
      role: 'system',
      content: `Permission request for "${request.description}" ${decisionMessage}`,
      timestamp: new Date(),
      isPermissionDecision: true
    })

    // Auto-scroll to bottom
    nextTick(() => {
      const container = document.querySelector('.messages-container')
      if (container) {
        container.scrollTop = container.scrollHeight
      }
    })

  } catch (error) {
    console.error('Failed to send permission response:', error)
    alert('Failed to send permission response. Please try again.')
  }
}

// Kill switch functionality
const killAllAgents = async () => {
  if (!agentWs.connected || sessions.value.length === 0) return

  if (!confirm('Are you sure you want to kill all active agents? This will end all sessions immediately.')) {
    return
  }

  try {
    agentWs.send({
      type: 'kill_all_agents'
    })
  } catch (error) {
    console.error('Failed to kill all agents:', error)
    alert('Failed to kill all agents. Please try again.')
  }
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
  console.log('Received agent message:', data)
  if (!messages.value[data.session_id]) {
    messages.value[data.session_id] = []
  }

  // Skip only empty content (unless it's a completion signal) and system messages
  if (!data.complete && (!data.content ||
      data.content.includes('SystemMessage'))) {
    return
  }

  // Check if we're awaiting tool results (after permission approval)
  const isToolResult = awaitingToolResults.value.has(data.session_id)

  // For streaming, use message_id or create one for this stream
  const messageId = data.message_id || 'stream-' + data.session_id

  const existingMessage = messages.value[data.session_id].find(
    m => (m.id === messageId || (m.streaming && !isToolResult)) && m.role === 'assistant'
  )

  // Force new message creation if this is a tool result
  if (isToolResult && data.content) {
    console.log('Creating new message for tool results')
    // Clear the flag since we're now creating the new message
    awaitingToolResults.value.delete(data.session_id)

    // Create a new message for tool results
    messages.value[data.session_id].push({
      id: messageId + '-result',  // Different ID to avoid conflicts
      role: 'assistant',
      content: data.content,
      timestamp: new Date(),
      streaming: !data.complete,
      isToolResult: true
    })

    // Reset processing when we receive content
    isProcessing.value = false
    isThinking.value = false
  } else if (existingMessage && !data.complete && !isToolResult && existingMessage.streaming) {
    // Append to existing streaming message only if it's still marked as streaming
    existingMessage.content += data.content
    // Reset processing when we receive content
    if (data.content) {
      isProcessing.value = false
      isThinking.value = false
    }
  } else if (!existingMessage && data.content && !isToolResult) {
    // New message - only add if there's actual content and not a tool result
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
    // Handle completion message
    if (data.content && !existingMessage && !isToolResult) {
      // This is a completion message with content but no existing message
      // Create a new message with the final content
      messages.value[data.session_id].push({
        id: messageId,
        role: 'assistant',
        content: data.content,
        timestamp: new Date(),
        streaming: false
      })
    } else if (existingMessage && !isToolResult) {
      // Mark existing message streaming as complete
      existingMessage.streaming = false
    } else {
      // If this was a tool result completion, find and mark it complete
      const toolResultMessage = messages.value[data.session_id].find(
        m => m.id === messageId + '-result' && m.role === 'assistant'
      )
      if (toolResultMessage) {
        toolResultMessage.streaming = false
      }
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

agentWs.on('onPermissionRequest', (data) => {
  console.log('Permission request received:', data)
  if (data.session_id === activeSessionId.value) {
    // Add to pending permissions for this session
    pendingPermissions.value.push({
      ...data,
      timestamp: new Date()
    })
  }
})

agentWs.on('onPermissionAcknowledged', (data) => {
  console.log('Permission acknowledged:', data)
  if (data.session_id === activeSessionId.value) {
    // Add a status message showing execution started
    if (!messages.value[data.session_id]) {
      messages.value[data.session_id] = []
    }

    const statusText = data.approved ?
      `⚡ Executing ${data.tool} command...` :
      `🚫 ${data.tool} command denied`

    messages.value[data.session_id].push({
      id: crypto.randomUUID(),
      role: 'system',
      content: statusText,
      timestamp: new Date(),
      isExecutionStatus: true
    })

    // If approved, mark that we're awaiting tool results (should appear as new message)
    if (data.approved) {
      awaitingToolResults.value.add(data.session_id)
      console.log('Marked session as awaiting tool results:', data.session_id)

      // Mark the last assistant message as complete (not streaming) so new messages
      // after tool execution don't get appended to it
      const lastMessage = messages.value[data.session_id].findLast(m => m.role === 'assistant')
      if (lastMessage && lastMessage.streaming) {
        lastMessage.streaming = false
        console.log('Finalized last assistant message to prevent appending')
      }
    }

    // Auto-scroll to bottom
    nextTick(() => {
      const container = document.querySelector('.messages-container')
      if (container) {
        container.scrollTop = container.scrollHeight
      }
    })
  }
})

agentWs.on('onError', (data) => {
  console.error('Agent error:', data.message)
  // Always reset on error
  isProcessing.value = false
  isThinking.value = false

  // Clear awaiting tool results flag on error
  if (data.session_id) {
    awaitingToolResults.value.delete(data.session_id)
  }

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

agentWs.on('onAgentsKilled', (data) => {
  console.log('Agents killed response:', data)

  // Clear all sessions and messages
  sessions.value = []
  messages.value = {}
  activeSessionId.value = null
  awaitingToolResults.value.clear()  // Clear all flags

  // Show success message
  alert(`Successfully killed ${data.killed_count} agents`)
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

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-text {
  flex: 1;
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

.header-actions {
  display: flex;
  gap: 12px;
}

.btn-kill-all {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-kill-all:hover:not(:disabled) {
  background: #c82333;
}

.btn-kill-all:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

/* Resume Session Styles */
.session-buttons {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.btn-resume-session {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-resume-session:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
}

.btn-resume-session:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--card-bg);
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
}

.modal-close {
  padding: 8px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.modal-close:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.modal-body {
  padding: 24px;
  max-height: 400px;
  overflow-y: auto;
}

.loading-sessions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-purple);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.no-sessions-available {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.no-sessions-available svg {
  margin-bottom: 16px;
}

.no-sessions-available p {
  margin: 0;
  font-size: 0.95rem;
}

.sessions-list-modal {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.session-item-modal {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: var(--bg-secondary);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.session-item-modal:hover {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
}

.session-info-modal {
  flex: 1;
  overflow: hidden;
}

.session-name-modal {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.session-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.session-directory {
  font-size: 0.85rem;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-meta {
  font-size: 0.8rem;
  color: var(--text-secondary);
  opacity: 0.8;
}

.session-resume-indicator {
  padding: 8px;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.session-item-modal:hover .session-resume-indicator {
  color: var(--accent-purple);
}

/* Historical message styling */
.message.isHistorical .message-content {
  background: var(--bg-secondary);
  border-color: var(--border-color);
  opacity: 0.7;
}

.message.isHistorical .message-role {
  opacity: 0.7;
}

/* Form Styles */
.form-group {
  margin-bottom: 24px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: var(--text-primary);
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.95rem;
  color: var(--text-primary);
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--accent-purple);
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
}

.form-help {
  display: block;
  margin-top: 4px;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.tools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 12px;
}

.tool-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.tool-checkbox:hover {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
}

.tool-checkbox input[type="checkbox"] {
  margin: 0;
}

.tool-checkbox span {
  font-size: 0.9rem;
  color: var(--text-primary);
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid var(--border-color);
}

.btn-cancel {
  padding: 12px 24px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: var(--bg-tertiary);
}

.btn-create {
  padding: 12px 24px;
  background: var(--accent-purple);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-create:hover:not(:disabled) {
  background: var(--accent-purple-hover);
}

.btn-create:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  position: relative;
}

.btn-create:disabled .btn-spinner {
  opacity: 1;
}

.btn-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-right: 8px;
}

.btn-spinner-small {
  width: 12px;
  height: 12px;
  border: 1.5px solid var(--accent-purple);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-right: 0;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.selected-session-info {
  padding: 20px;
  background: var(--bg-secondary);
  border-radius: 8px;
  margin-bottom: 24px;
}

.selected-session-info h3 {
  margin: 0 0 8px 0;
  color: var(--text-primary);
}

.selected-session-info p {
  margin: 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.selected-session-info code {
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.85em;
}

.resume-session-options {
  padding: 24px 0;
}

/* Permission Requests */
.permission-requests {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  max-height: 200px;
  overflow-y: auto;
}

.permission-request {
  background: linear-gradient(135deg, #fff3cd, #fef5e7);
  border: 1px solid #ffc107;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(255, 193, 7, 0.2);
  animation: slideIn 0.3s ease-out;
}

.permission-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.permission-icon {
  font-size: 1.2rem;
}

.permission-title {
  font-weight: 600;
  color: #856404;
  flex: 1;
}

.permission-time {
  font-size: 0.8rem;
  color: #856404;
  opacity: 0.7;
}

.permission-description {
  color: #856404;
  font-size: 0.9rem;
  line-height: 1.4;
  margin-bottom: 12px;
  padding: 8px 0;
}

.permission-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.btn-approve,
.btn-deny {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-approve {
  background: #28a745;
  color: white;
}

.btn-approve:hover {
  background: #218838;
}

.btn-deny {
  background: #dc3545;
  color: white;
}

.btn-deny:hover {
  background: #c82333;
}

/* Permission decision messages */
.message.isPermissionDecision .message-content {
  background: var(--bg-secondary);
  border-color: var(--border-color);
  opacity: 0.8;
  font-style: italic;
}

/* Execution status messages */
.message.isExecutionStatus .message-content {
  background: linear-gradient(135deg, #e8f5e9, #c8e6c9);
  border-color: #4caf50;
  color: #2e7d32;
  font-weight: 500;
  font-style: normal;
  animation: slideIn 0.3s ease-out;
}

/* Tool result messages */
.message.isToolResult .message-content {
  background: var(--bg-secondary);
  border: 2px solid var(--accent-purple);
  border-left: 4px solid var(--accent-purple);
  animation: slideIn 0.3s ease-out;
}

/* System messages styling */
.message.system .message-role {
  color: var(--text-secondary);
}

.message.system .message-content {
  background: transparent;
  border: none;
  padding: 8px 12px;
  font-size: 0.9rem;
  text-align: center;
  color: var(--text-secondary);
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

  .modal-content {
    width: 95%;
    margin: 20px;
  }

  .tools-grid {
    grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
  }

  .modal-actions {
    flex-direction: column;
  }

  .btn-cancel,
  .btn-create {
    width: 100%;
  }
}
</style>