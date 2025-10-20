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
            @click="deleteAllSessions"
            class="btn-delete-all"
            :disabled="!agentWs.connected || sessions.length === 0"
            title="Delete all sessions from database"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              <line x1="10" y1="11" x2="10" y2="17"></line>
              <line x1="14" y1="11" x2="14" y2="17"></line>
            </svg>
            Delete All Sessions
          </button>
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
      <SessionsSidebar
        :sessions="filteredSessions"
        :active-session-id="activeSessionId"
        :active-filter="activeFilter"
        :filters="sessionFiltersWithCounts"
        :connected="agentWs.connected"
        :creating="creatingSession"
        @create-new="createNewSession"
        @resume="openResumeModal"
        @update:active-filter="activeFilter = $event"
        @select="selectSession"
        @end="endSession"
        @delete="deleteSession"
      />

      <!-- Chat Area with Metrics -->
      <main class="chat-area-with-metrics">
        <ChatArea
          ref="chatAreaRef"
          :has-active-session="!!activeSessionId"
          v-model:input-message="inputMessage"
          :connected="agentWs.connected"
          :is-thinking="isThinking"
          :is-processing="isProcessing"
          @send="sendMessage"
        >
          <!-- Tool Overlays Slot -->
          <template #tool-overlays>
            <ToolOverlaysContainer :tools="activeSessionTools">
              <template v-for="tool in activeSessionTools" :key="tool.id">
                <TodoWriteOverlay
                  v-if="tool.name === 'TodoWrite'"
                  :tool="tool"
                  @dismiss="removeActiveTool(tool.sessionId, $event)"
                />
                <ToolOverlay
                  v-else
                  :tool="tool"
                  @dismiss="removeActiveTool(tool.sessionId, $event)"
                />
              </template>
            </ToolOverlaysContainer>
          </template>

          <!-- TodoWrite Box Slot -->
          <template #todo-box>
            <TodoWriteBox
              :show="shouldShowTodoBox"
              :todos="activeSessionTodos"
            />
          </template>

          <!-- Messages Slot -->
          <template #messages>
            <MessageBubble
              v-for="message in activeMessages"
              :key="message.id"
              :message="message"
              :format-time="formatTime"
              :format-message="formatMessage"
            />
          </template>

          <!-- Permissions Slot -->
          <template #permissions>
            <div v-if="activeSessionPermissions.length > 0" class="permission-requests">
              <PermissionRequest
                v-for="permission in activeSessionPermissions"
                :key="permission.request_id"
                :permission="permission"
                @approve="approvePermission"
                @deny="denyPermission"
              />
            </div>
          </template>

          <!-- Tool Execution Slot -->
          <template #tool-execution>
            <ToolExecutionBar :tool-execution="activeSessionToolExecution" />
          </template>
        </ChatArea>

        <!-- Session Metrics Sidebar -->
        <MetricsSidebar
          :show="!!activeSessionId"
          :session="activeSession"
          :tool-executions="sessionToolStats.get(activeSessionId)"
          :permission-stats="sessionPermissionStats.get(activeSessionId)"
        />
      </main>
    </div>

    <!-- Create Session Modal -->
    <CreateSessionModal
      :show="showCreateSessionModal"
      :form-data="sessionForm"
      :providers="availableProviders"
      :current-provider="currentProvider"
      :agents="availableAgents"
      :selected-agent-preview="selectedAgentPreview"
      :loading-providers="loadingProviders"
      :loading-agents="loadingAgents"
      :creating="creatingSession"
      @close="showCreateSessionModal = false"
      @create="createSessionWithOptions"
      @working-directory-change="handleWorkingDirectoryChange"
      @agent-select="loadSelectedAgent"
    />

    <!-- Resume Session Modal -->
    <ResumeSessionModal
      :show="showResumeModal"
      :sessions="availableSessions"
      :selected-session="selectedResumeSession"
      :form-data="resumeForm"
      :loading="loadingSessions"
      :resuming="resumingSession"
      @close="showResumeModal = false; selectedResumeSession = null"
      @select-session="selectSessionForResume"
      @back="selectedResumeSession = null"
      @resume="resumeSessionWithOptions"
    />

  </div>
</template>

<script setup lang="ts">
import { useAgentWebSocket } from '~/composables/useAgentWebSocket'
import SessionMetrics from '~/components/SessionMetrics.vue'
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import type { ActiveTool } from '~/types/agents'

// Refactored Components
import SessionItem from '~/components/agents/SessionItem.vue'
import SessionFilters from '~/components/agents/SessionFilters.vue'
import PermissionRequest from '~/components/agents/PermissionRequest.vue'
import ToolExecutionBar from '~/components/agents/ToolExecutionBar.vue'
import CreateSessionModal from '~/components/agents/CreateSessionModal.vue'
import ResumeSessionModal from '~/components/agents/ResumeSessionModal.vue'
import SessionsSidebar from '~/components/agents/SessionsSidebar.vue'
import ChatArea from '~/components/agents/ChatArea.vue'
import MetricsSidebar from '~/components/agents/MetricsSidebar.vue'
import MessageBubble from '~/components/agents/MessageBubble.vue'
import TodoWriteBox from '~/components/agents/TodoWriteBox.vue'
import ToolOverlaysContainer from '~/components/agents/ToolOverlaysContainer.vue'

// Utilities
import { formatTime, formatMessage, formatRelativeTime, truncatePath } from '~/utils/agents/messageFormatters'
import { parseTodoWrite, formatTodosForTool, type TodoItem } from '~/utils/agents/todoParser'
import { parseToolUse, getToolIcon } from '~/utils/agents/toolParser'

// Composables
import { useMessageScroll } from '~/composables/agents/useMessageScroll'
import { useSessionState } from '~/composables/agents/useSessionState'
import { useAgentProviders } from '~/composables/agents/useAgentProviders'

// Existing overlays
import TodoWriteOverlay from '~/components/TodoWriteOverlay.vue'
import ToolOverlay from '~/components/ToolOverlay.vue'

// WebSocket connection
const agentWs = useAgentWebSocket()

// Refs
const messageInput = ref(null)
const chatAreaRef = ref(null)
// Access messagesContainer through chatAreaRef
const messagesContainer = computed(() => chatAreaRef.value?.messagesContainer || null)

// Composables - Session State Management
const {
  sessions,
  activeSessionId,
  messages,
  messagesLoaded,
  inputMessage,
  isProcessing,
  isThinking,
  showResumeModal,
  showCreateSessionModal,
  selectedResumeSession,
  availableSessions,
  loadingSessions,
  creatingSession,
  resumingSession,
  sessionPermissions,
  awaitingToolResults,
  sessionTodos,
  sessionToolExecution,
  todoHideTimers,
  activeTools,
  activeFilter,
  sessionToolStats,
  sessionPermissionStats,
  filteredSessions,
  sessionFiltersWithCounts,
  activeSession,
  activeMessages,
  activeSessionPermissions,
  activeSessionTodos,
  activeSessionToolExecution,
  activeSessionTools,
  shouldShowTodoBox
} = useSessionState()

// Composables - Provider & Agent Selection
const {
  sessionForm,
  resumeForm,
  availableAgents,
  selectedAgentPreview,
  loadingAgents,
  availableProviders,
  currentProvider,
  loadingProviders,
  getProviderModels
} = useAgentProviders()

// Auto-scroll composable
const { isUserNearBottom, handleScroll, scrollToBottom, autoScrollIfNearBottom } = useMessageScroll()

// Session management
const createNewSession = async () => {
  if (!agentWs.connected) return

  // Reset form to defaults
  sessionForm.value = {
    workingDirectory: '',
    permissionMode: 'default',
    modelProvider: 'anthropic',
    model: 'claude-sonnet-4.5-20250514',
    systemPrompt: '',
    promptMode: 'agent',
    selectedAgent: '',
    tools: ['Read', 'Write', 'Edit', 'Bash', 'Search', 'TodoWrite']
  }

  // Fetch current working directory
  try {
    const response = await fetch('/api/config/cwd')
    if (response.ok) {
      const data = await response.json()
      if (data.cwd) {
        sessionForm.value.workingDirectory = data.cwd
        console.log('Auto-populated working directory:', data.cwd)
        // Load agents from this directory
        await loadAvailableAgents()
      }
    }
  } catch (error) {
    console.error('Error fetching current working directory:', error)
  }

  showCreateSessionModal.value = true
}

const createSessionWithOptions = async () => {
  if (!agentWs.connected || !sessionForm.value.workingDirectory) return

  creatingSession.value = true

  try {
    const sessionId = crypto.randomUUID()

    // Find the selected provider to get base_url
    const selectedProvider = availableProviders.value.find(p => p.id === sessionForm.value.modelProvider)

    const options: any = {
      tools: sessionForm.value.tools,
      working_directory: sessionForm.value.workingDirectory,
      permission_mode: sessionForm.value.permissionMode,
      provider: sessionForm.value.modelProvider,
      model: sessionForm.value.model
    }

    // Add base_url if the provider has one (for non-default Anthropic providers)
    if (selectedProvider?.base_url) {
      options.base_url = selectedProvider.base_url
    }

    // Use agent_name if agent mode is selected, otherwise use system_prompt
    if (sessionForm.value.promptMode === 'agent' && sessionForm.value.selectedAgent) {
      options.agent_name = sessionForm.value.selectedAgent
    } else {
      options.system_prompt = sessionForm.value.systemPrompt || 'You are a helpful AI assistant.'
    }

    console.log('Creating session with options:', options)

    agentWs.send({
      type: 'create_session',
      session_id: sessionId,
      options
    })

    showCreateSessionModal.value = false
  } catch (error) {
    console.error('Failed to create session:', error)
    alert('Failed to create session. Please try again.')
  } finally {
    creatingSession.value = false
  }
}

const loadAvailableAgents = async () => {
  loadingAgents.value = true
  try {
    const response = await fetch('/api/agents')
    if (!response.ok) throw new Error(`Failed to fetch agents: ${response.status}`)
    const data = await response.json()
    availableAgents.value = data.agents ? Object.values(data.agents) : []
    console.log(`Loaded ${availableAgents.value.length} agents from project`, availableAgents.value)
  } catch (error) {
    console.error('Error loading agents:', error)
    availableAgents.value = []
  } finally {
    loadingAgents.value = false
  }
}

// Load available providers from backend
const loadProviders = async () => {
  loadingProviders.value = true
  try {
    const response = await fetch('/api/providers')
    if (!response.ok) throw new Error(`Failed to fetch providers: ${response.status}`)
    const data = await response.json()
    availableProviders.value = data.providers || []
    currentProvider.value = data.current

    // Update form defaults based on current provider
    if (currentProvider.value) {
      const provider = availableProviders.value.find(p => p.id === currentProvider.value.provider_id)
      if (provider) {
        sessionForm.value.modelProvider = provider.id
        if (currentProvider.value.model_name) {
          sessionForm.value.model = currentProvider.value.model_name
        } else if (provider.default_model) {
          sessionForm.value.model = provider.default_model
        }
      }
    }

    console.log(`Loaded ${availableProviders.value.length} providers`, availableProviders.value)
  } catch (error) {
    console.error('Error loading providers:', error)
    availableProviders.value = []
  } finally {
    loadingProviders.value = false
  }
}

// Reload agents (agents auto-load from cwd)
const handleWorkingDirectoryChange = async () => {
  if (sessionForm.value.workingDirectory) {
    await loadAvailableAgents()
    // Clear selected agent when working directory changes
    sessionForm.value.selectedAgent = ''
    selectedAgentPreview.value = null
  }
}

// Auto-scroll helpers
// Scroll functions are now provided by useMessageScroll() composable

const loadSelectedAgent = async () => {
  if (!sessionForm.value.selectedAgent) {
    selectedAgentPreview.value = null
    return
  }

  try {
    const response = await fetch(`/api/agents/${sessionForm.value.selectedAgent}`)
    if (!response.ok) throw new Error('Failed to fetch agent')
    const data = await response.json()
    selectedAgentPreview.value = data.agent
    console.log(`Loaded agent: ${sessionForm.value.selectedAgent}`, data.agent)
  } catch (error) {
    console.error('Error loading agent:', error)
    selectedAgentPreview.value = null
  }
}

const selectSession = (sessionId) => {
  activeSessionId.value = sessionId

  // Load historical messages if not already loaded
  // Only load if we haven't loaded this session before (prevents loading for new sessions)
  if (!messagesLoaded.value.has(sessionId)) {
    console.log(`Loading messages for session ${sessionId}`)
    agentWs.send({
      type: 'load_messages',
      session_id: sessionId,
      limit: 100,
      offset: 0
    })
    // Mark as loaded (even if empty) to prevent re-loading
    messagesLoaded.value.add(sessionId)
  }

  // Reset scroll state and scroll to bottom when switching sessions
  isUserNearBottom.value = true
  scrollToBottom(messagesContainer.value, false)

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
  messagesLoaded.value.delete(sessionId)  // Clean up loaded tracking
  awaitingToolResults.value.delete(sessionId)  // Clean up flag

  // Clean up any pending timers
  const existingTimer = todoHideTimers.value.get(sessionId)
  if (existingTimer) {
    clearTimeout(existingTimer)
    todoHideTimers.value.delete(sessionId)
  }

  // Clean up live agents session data
  cleanupSessionData(sessionId)

  // Clean up session permissions
  sessionPermissions.value.delete(sessionId)

  // Clean up session metrics
  sessionToolStats.value.delete(sessionId)
  sessionPermissionStats.value.delete(sessionId)

  if (activeSessionId.value === sessionId) {
    activeSessionId.value = null
  }
}

const deleteSession = async (sessionId) => {
  if (!agentWs.connected) return

  // Confirm deletion
  if (!confirm('Are you sure you want to delete this session? This action cannot be undone.')) {
    return
  }

  agentWs.send({
    type: 'delete_session',
    session_id: sessionId
  })

  // Remove from local state immediately (optimistic update)
  sessions.value = sessions.value.filter(s => s.id !== sessionId)
  delete messages.value[sessionId]
  messagesLoaded.value.delete(sessionId)
  awaitingToolResults.value.delete(sessionId)

  // Clean up any pending timers
  const existingTimer = todoHideTimers.value.get(sessionId)
  if (existingTimer) {
    clearTimeout(existingTimer)
    todoHideTimers.value.delete(sessionId)
  }

  // Clean up live agents session data
  cleanupSessionData(sessionId)

  // Clean up session permissions
  sessionPermissions.value.delete(sessionId)

  // Clean up session metrics
  sessionToolStats.value.delete(sessionId)
  sessionPermissionStats.value.delete(sessionId)

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
  showResumeModal.value = true
}

const selectSessionForResume = async (session) => {
  selectedResumeSession.value = session

  // Prefill the form with the session's data
  resumeForm.value = {
    workingDirectory: session.working_directory || '',
    permissionMode: 'default',
    systemPrompt: '',
    tools: ['Read', 'Write', 'Edit', 'Bash', 'Search', 'TodoWrite']
  }
}

const resumeSessionWithOptions = async () => {
  try {
    if (!selectedResumeSession.value) return

    resumingSession.value = true

    // Fetch resume data from the backend
    const resumeData = await $fetch(`/api/sessions/${selectedResumeSession.value.conversation_id}/resume-data`)

    // Create new agent session with history context and options
    const sessionId = crypto.randomUUID()

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

  } catch (error) {
    console.error('Failed to resume session:', error)
    alert('Failed to resume session. Please try again.')
  } finally {
    resumingSession.value = false
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

// Helper methods for parsing TodoWrite and tool execution data
const parseTodoWrite = (content: string): TodoItem[] | null => {
  if (!content || typeof content !== 'string') return null

  try {
    console.log('Parsing TodoWrite from content:', content)

    // Pattern 1: Look for numbered lists (1. task, 2. task, etc.)
    const numberedListMatch = content.match(/(?:\d+\.\s+)([^\n]+)/g)
    if (numberedListMatch) {
      console.log('Found numbered list matches:', numberedListMatch)
      const todos: TodoItem[] = []
      for (let i = 0; i < numberedListMatch.length; i++) {
        const taskContent = numberedListMatch[i].replace(/^\d+\.\s+/, '').trim()
        if (taskContent) {
          todos.push({
            content: taskContent,
            status: i === 0 ? 'in_progress' : 'pending'
          })
        }
      }
      if (todos.length > 0) {
        console.log('Successfully parsed todos from numbered list:', todos)
        return todos
      }
    }

    // Pattern 2: Look for checkmark-style lists (- task, * task, etc.)
    const bulletListMatch = content.match(/[-*]\s+([^\n]+)/g)
    if (bulletListMatch) {
      console.log('Found bullet list matches:', bulletListMatch)
      const todos: TodoItem[] = []
      for (const match of bulletListMatch) {
        const taskContent = match.replace(/^[-*]\s+/, '').trim()
        if (taskContent) {
          todos.push({
            content: taskContent,
            status: 'pending'
          })
        }
      }
      if (todos.length > 0) {
        console.log('Successfully parsed todos from bullet list:', todos)
        return todos
      }
    }

    // Pattern 3: Look for explicit todo markers ("Todo:", "Task:", etc.)
    const todoMarkerMatch = content.match(/(?:todo|task|items?):\s*([\s\S]*?)(?=\n\n|\n\w+:|$)/i)
    if (todoMarkerMatch) {
      const todoText = todoMarkerMatch[1].trim()
      const items = todoText.split(/\n\s*\n/).filter(item => item.trim())
      if (items.length > 0) {
        const todos = items.map(item => ({
          content: item.trim(),
          status: 'pending'
        }))
        console.log('Successfully parsed todos from todo marker:', todos)
        return todos
      }
    }

    // Pattern 4: Look for task-related patterns (common task descriptions)
    const taskPatterns = [
      /(?:i'll create|let me create|creating|here are)\s+a\s+(?:todo|task|list):\s*([\s\S]*?)(?=\n\n|\n|$)/i,
      /(?:tasks?:\s*\n)((?:\d+\.\s+[^\n]+\n)+)/i,
      /(?:items?:\s*\n)((?:[-*]\s+[^\n]+\n)+)/i
    ]

    for (const pattern of taskPatterns) {
      const match = content.match(pattern)
      if (match) {
        const taskContent = match[1] || match[0]
        const lines = taskContent.split('\n').filter(line => line.trim())
        const todos = lines.map(line => ({
          content: line.trim().replace(/^\d+\.\s+/, '').replace(/^[-*]\s+/, ''),
          status: 'pending'
        })).filter(todo => todo.content)

        if (todos.length > 0) {
          console.log('Successfully parsed todos from task pattern:', todos)
          return todos
        }
      }
    }

    console.log('No TodoWrite data found in content')
    return null
  } catch (e) {
    console.warn('Failed to parse TodoWrite content:', e)
    return null
  }
}

const parseToolUse = (content: string): ToolExecution | null => {
  if (!content || typeof content !== 'string') return null

  try {
    // Look for tool use patterns
    const patterns = [
      /Using (\w+)/g,
      /(\w+)\s*\(/g, // Function calls
      /Running (\w+)/g,
      /Executing (\w+)/g
    ]

    for (const pattern of patterns) {
      const matches = [...content.matchAll(pattern)]
      if (matches.length > 0) {
        const toolName = matches[0][1]

        // Extract additional details based on tool type
        let filePath, command, patternStr

        if (toolName === 'Read' || toolName === 'Write' || toolName === 'Edit') {
          const fileMatch = content.match(/(?:file|path):\s*([^\s\n]+)/i)
          if (fileMatch) filePath = fileMatch[1]
        } else if (toolName === 'Bash') {
          const commandMatch = content.match(/(?:command|cmd):\s*([^\n]+)/i)
          if (commandMatch) command = commandMatch[1].trim()
        } else if (toolName === 'Search' || toolName === 'Grep') {
          const patternMatch = content.match(/(?:pattern|search):\s*([^\n]+)/i)
          if (patternMatch) patternStr = patternMatch[1].trim()
        }

        return {
          toolName,
          filePath,
          command,
          pattern: patternStr,
          timestamp: new Date()
        }
      }
    }

    return null
  } catch (e) {
    console.warn('Failed to parse tool use:', e)
    return null
  }
}

// Update session data methods
const updateSessionTodos = (sessionId: string, todos: TodoItem[]) => {
  sessionTodos.value.set(sessionId, todos)
}

// Helper to format todos for TodoWrite tool (includes activeForm only when present)
const formatTodosForTool = (todos: TodoItem[]): any[] => {
  return todos.map(todo => ({
    content: todo.content,
    status: todo.status,
    ...(todo.activeForm && { activeForm: todo.activeForm })
  }))
}

const updateSessionToolExecution = (sessionId: string, toolExecution: ToolExecution | null) => {
  sessionToolExecution.value.set(sessionId, toolExecution)
}

const clearSessionToolExecution = (sessionId: string) => {
  sessionToolExecution.value.delete(sessionId)
}

// Tool overlay management
const addActiveTool = (sessionId: string, toolUse: any) => {
  const tools = activeTools.value.get(sessionId) || []
  const activeTool: ActiveTool = {
    id: toolUse.id,
    name: toolUse.name,
    input: toolUse.input,
    status: 'running',
    startTime: Date.now(),
    sessionId
  }
  tools.push(activeTool)
  activeTools.value.set(sessionId, tools)
}

const completeActiveTool = (sessionId: string, toolUseId: string, isError: boolean = false) => {
  const tools = activeTools.value.get(sessionId) || []
  const tool = tools.find(t => t.id === toolUseId)
  if (tool) {
    tool.status = isError ? 'error' : 'completed'
    tool.endTime = Date.now()
    activeTools.value.set(sessionId, [...tools])
  }
}

const removeActiveTool = (sessionId: string, toolId: string) => {
  const tools = activeTools.value.get(sessionId) || []
  const filtered = tools.filter(t => t.id !== toolId)
  activeTools.value.set(sessionId, filtered)
}

const cleanupSessionData = (sessionId: string) => {
  sessionTodos.value.delete(sessionId)
  sessionToolExecution.value.delete(sessionId)
  activeTools.value.delete(sessionId)
}

const truncatePath = (path: string): string => {
  if (!path) return ''
  if (path.length <= 50) return path

  // Truncate from the middle, keeping the beginning and end
  const start = path.substring(0, 25)
  const end = path.substring(path.length - 20)
  return `${start}...${end}`
}

// Watch for modal opening to load sessions
watch(showResumeModal, (show) => {
  if (show) {
    loadAvailableSessions()
  }
})

// Watch for all todos completed and auto-hide after 5 seconds
watch(activeSessionTodos, (todos) => {
  if (!activeSessionId.value) return

  // Clear any existing timer for this session
  const existingTimer = todoHideTimers.value.get(activeSessionId.value)
  if (existingTimer) {
    clearTimeout(existingTimer)
    todoHideTimers.value.delete(activeSessionId.value)
  }

  // If all todos are completed, set a new timer
  if (todos.length > 0 && todos.every(todo => todo.status === 'completed')) {
    console.log('All todos completed, setting 5 second auto-hide timer')
    const timer = setTimeout(() => {
      const currentTodos = sessionTodos.value.get(activeSessionId.value)
      if (currentTodos && currentTodos.every(todo => todo.status === 'completed')) {
        sessionTodos.value.delete(activeSessionId.value)
        todoHideTimers.value.delete(activeSessionId.value)
        console.log('Auto-hid todos after 5 seconds')
      }
    }, 5000)
    todoHideTimers.value.set(activeSessionId.value, timer)
  }
}, { deep: true })

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

  // Clear previous todos when sending a new message (moving to new task)
  const sessionId = activeSessionId.value
  const existingTimer = todoHideTimers.value.get(sessionId)
  if (existingTimer) {
    clearTimeout(existingTimer)
    todoHideTimers.value.delete(sessionId)
  }
  sessionTodos.value.delete(sessionId)

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

    // Update permission stats
    const permStats = sessionPermissionStats.value.get(request.session_id) || { approved: 0, denied: 0, total: 0 }
    if (approved) {
      permStats.approved++
    } else {
      permStats.denied++
    }
    sessionPermissionStats.value.set(request.session_id, permStats)

    // Remove from session-specific permissions
    const sessionPerms = sessionPermissions.value.get(request.session_id) || []
    sessionPermissions.value.set(
      request.session_id,
      sessionPerms.filter(p => p.request_id !== request.request_id)
    )

    // Add a system message to show the decision (to the correct session)
    if (!messages.value[request.session_id]) {
      messages.value[request.session_id] = []
    }

    const decisionText = approved ? '✅ Approved' : '❌ Denied'
    const decisionMessage = reason ? `${decisionText} (Reason: ${reason})` : decisionText

    messages.value[request.session_id].push({
      id: crypto.randomUUID(),
      role: 'system',
      content: `Permission request for "${request.description}" ${decisionMessage}`,
      timestamp: new Date(),
      isPermissionDecision: true
    })

    // Auto-scroll to bottom only if viewing this session
    if (request.session_id === activeSessionId.value) {
      autoScrollIfNearBottom(messagesContainer.value)
    }

  } catch (error) {
    console.error('Failed to send permission response:', error)
    alert('Failed to send permission response. Please try again.')
  }
}

// Delete all sessions functionality
const deleteAllSessions = async () => {
  if (!agentWs.connected || sessions.value.length === 0) return

  if (!confirm('Are you sure you want to delete ALL sessions? This will permanently delete all session data from the database. This action cannot be undone.')) {
    return
  }

  try {
    agentWs.send({
      type: 'delete_all_sessions'
    })
  } catch (error) {
    console.error('Failed to delete all sessions:', error)
    alert('Failed to delete all sessions. Please try again.')
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

// Helper function to extract text content from nested content object
const extractTextContent = (content: any): string => {
  if (!content) return ''

  // If content is already a string, return it
  if (typeof content === 'string') return content

  // If content is an object with nested structure
  if (typeof content === 'object') {
    // Handle assistant messages with text array
    if (content.type === 'assistant') {
      // Check if text array exists and is not null/empty
      if (Array.isArray(content.text) && content.text.length > 0) {
        return content.text.join('\n')
      }
      // Empty or null text array - no content to display
      return ''
    }

    // Handle user messages
    if (content.type === 'user' && content.content) {
      return String(content.content)
    }

    // Handle result messages (completion signal - no visible content)
    if (content.type === 'result') {
      return ''
    }

    // Handle system messages
    if (content.type === 'system') {
      return `SystemMessage: ${content.subtype || 'unknown'}`
    }

    // Fallback: stringify the object
    return JSON.stringify(content)
  }

  return String(content)
}

// Helper to check if content is complete signal
const isCompleteSignal = (content: any): boolean => {
  return typeof content === 'object' && content.type === 'result'
}

// Helper to extract cost/usage data from result messages
const extractCostData = (content: any) => {
  if (typeof content === 'object' && content.type === 'result') {
    return {
      costUSD: content.cost_usd || 0,
      numTurns: content.num_turns || 0,
      durationMs: content.duration_ms || 0,
      usage: content.usage || null
    }
  }
  return null
}

// Helper to extract tool name from tool_uses JSON
const extractToolName = (toolUses: any): string | undefined => {
  if (!toolUses) return undefined

  try {
    // Parse if it's a JSON string
    const parsed = typeof toolUses === 'string' ? JSON.parse(toolUses) : toolUses

    // If it's an array, get the first tool
    if (Array.isArray(parsed) && parsed.length > 0) {
      return parsed[0].name || parsed[0].type
    }

    // If it's a single object
    if (parsed.name || parsed.type) {
      return parsed.name || parsed.type
    }
  } catch (e) {
    console.warn('Failed to parse tool_uses:', e)
  }

  return undefined
}

// WebSocket event handlers
agentWs.on('onSessionCreated', (data) => {
  sessions.value.push(data.session)
  activeSessionId.value = data.session_id
  messages.value[data.session_id] = []

  // Mark new session as loaded (it has no history to load)
  messagesLoaded.value.add(data.session_id)

  // Focus the input after session creation
  focusMessageInput()
})

agentWs.on('onAgentMessage', (data) => {
  console.log('📨 Received agent message:', data)

  if (!messages.value[data.session_id]) {
    messages.value[data.session_id] = []
  }

  // Check if this is a completion signal (result message)
  const isComplete = isCompleteSignal(data.content)

  // Extract cost data from result messages
  const costData = extractCostData(data.content)

  // Extract text content from nested object
  const textContent = extractTextContent(data.content)

  console.log('💬 Extracted:', { isComplete, costData, textContent: textContent.substring(0, 50) })

  // Process tool uses (when Claude starts using a tool)
  if (data.content && data.content.tools && Array.isArray(data.content.tools)) {
    data.content.tools.forEach((toolUse: any) => {
      console.log('🔧 Tool use detected:', toolUse.name)
      addActiveTool(data.session_id, toolUse)
    })
  }

  // Process tool results (when tool execution completes)
  if (data.content && data.content.tool_results && Array.isArray(data.content.tool_results)) {
    data.content.tool_results.forEach((toolResult: any) => {
      console.log('✅ Tool result received:', toolResult.tool_use_id)
      completeActiveTool(data.session_id, toolResult.tool_use_id, toolResult.is_error || false)
    })
  }

  // Update session status and metadata
  const session = sessions.value.find(s => s.id === data.session_id)
  if (session) {
    // Update git branch from metadata
    if (data.metadata && data.metadata.git_branch) {
      session.git_branch = data.metadata.git_branch
      console.log('🌿 Updated git branch:', session.git_branch)
    }

    // Update costs from result message
    if (costData) {
      session.cost_usd = (session.cost_usd || 0) + costData.costUSD
      session.num_turns = costData.numTurns
      session.duration_ms = costData.durationMs
      session.usage = costData.usage
      console.log('💰 Updated session cost:', session.cost_usd)
    }

    // Set status: idle when complete, processing when receiving content
    if (isComplete) {
      session.status = 'idle'
      session.message_count = (session.message_count || 0) + 1
    } else if (textContent) {
      session.status = 'processing'
    }
  }

  // Clear tool execution when we receive a message
  clearSessionToolExecution(data.session_id)

  // Clear todos when message completes (agent moving to next task)
  if (isComplete) {
    const existingTimer = todoHideTimers.value.get(data.session_id)
    if (existingTimer) {
      clearTimeout(existingTimer)
      todoHideTimers.value.delete(data.session_id)
    }
    sessionTodos.value.delete(data.session_id)

    // Reset processing state
    isProcessing.value = false
    isThinking.value = false

    // Focus the input after Claude completes the response
    if (data.session_id === activeSessionId.value) {
      focusMessageInput()
    }

    // Don't create a UI message for result/completion
    console.log('✅ Message complete (result received)')
    return
  }

  // Handle user messages with tool results differently
  if (data.content && data.content.type === 'user' && data.content.tool_results && Array.isArray(data.content.tool_results)) {
    // Format tool results as readable messages
    const sessionTools = activeTools.value.get(data.session_id) || []
    const formattedTools: string[] = []

    data.content.tool_results.forEach((toolResult: any) => {
      // Find the original tool use by tool_use_id
      const tool = sessionTools.find(t => t.id === toolResult.tool_use_id)

      if (tool && tool.name !== 'TodoWrite') {
        // Format based on tool type
        let formatted = ''

        switch (tool.name) {
          case 'Read':
            formatted = `Read(${tool.input.file_path || ''})`
            break
          case 'Write':
            formatted = `Write(${tool.input.file_path || ''})`
            break
          case 'Edit':
            formatted = `Edit(${tool.input.file_path || ''})`
            break
          case 'Bash':
            const cmd = tool.input.command || ''
            formatted = `Bash(${cmd.length > 50 ? cmd.substring(0, 50) + '...' : cmd})`
            break
          case 'Glob':
            formatted = `Glob(${tool.input.pattern || ''})`
            break
          case 'Grep':
            formatted = `Grep(${tool.input.pattern || ''})`
            break
          default:
            formatted = `${tool.name}()`
        }

        formattedTools.push(formatted)
      }
    })

    // Only create a message if we have tools to display
    if (formattedTools.length > 0) {
      const toolMessage = {
        id: `msg-${data.session_id}-${Date.now()}`,
        role: 'assistant',
        content: formattedTools.join(', '),
        timestamp: new Date(),
        isToolResult: true
      }

      messages.value[data.session_id].push(toolMessage)
      console.log('🔧 Created tool result message:', toolMessage.content)
    }

    return
  }

  // Skip empty content and system messages (they don't need UI display)
  if (!textContent || textContent.includes('SystemMessage')) {
    console.log('⏭️  Skipping empty/system message')
    return
  }

  // Check if we're awaiting tool results (after permission approval)
  const isToolResult = awaitingToolResults.value.has(data.session_id)
  if (isToolResult) {
    awaitingToolResults.value.delete(data.session_id)
  }

  // Create or update assistant message
  // Since backend sends complete messages (not character-by-character streaming),
  // we just create a new message for each response
  const newMessage = {
    id: `msg-${data.session_id}-${Date.now()}`,
    role: 'assistant',
    content: textContent,
    timestamp: new Date(),
    streaming: false,
    isToolResult: isToolResult
  }

  messages.value[data.session_id].push(newMessage)
  console.log('✨ Created new message:', newMessage.id)

  // Reset processing state when we receive content
  isProcessing.value = false
  isThinking.value = false

  // Auto-scroll to bottom if user is near bottom
  autoScrollIfNearBottom(messagesContainer.value)
})

agentWs.on('onAgentThinking', (data) => {
  if (data.session_id === activeSessionId.value) {
    isThinking.value = data.thinking
    // When thinking stops, ensure processing is also reset
    if (!data.thinking) {
      isProcessing.value = false
    }
  }

  // Update session status based on thinking state
  const session = sessions.value.find(s => s.id === data.session_id)
  if (session) {
    session.status = data.thinking ? 'processing' : 'idle'
  }
})

agentWs.on('onAgentToolUse', (data) => {
  // Update session status to processing when tool is being used
  const session = sessions.value.find(s => s.id === data.session_id)
  if (session) {
    session.status = 'processing'
  }

  // Track tool usage for metrics
  const toolStats = sessionToolStats.value.get(data.session_id) || {}
  toolStats[data.tool] = (toolStats[data.tool] || 0) + 1
  sessionToolStats.value.set(data.session_id, toolStats)

  // Extract tool details from parameters for display
  let toolDetail = ''
  if (data.parameters) {
    const params = typeof data.parameters === 'string' ? JSON.parse(data.parameters) : data.parameters
    if (data.tool === 'Read' || data.tool === 'Write' || data.tool === 'Edit') {
      toolDetail = params.file_path
    } else if (data.tool === 'Bash') {
      toolDetail = params.command
    } else if (data.tool === 'Glob') {
      toolDetail = params.pattern
    } else if (data.tool === 'Grep') {
      toolDetail = params.pattern
    }
  }

  // Update tool execution display
  if (data.session_id === activeSessionId.value) {
    sessionToolExecution.value.set(data.session_id, {
      toolName: data.tool,
      filePath: data.tool === 'Read' || data.tool === 'Write' || data.tool === 'Edit' ? toolDetail : undefined,
      command: data.tool === 'Bash' ? toolDetail : undefined,
      pattern: data.tool === 'Glob' || data.tool === 'Grep' ? toolDetail : undefined,
      detail: toolDetail
    })
  }

  if (!messages.value[data.session_id]) return

  const lastMessage = messages.value[data.session_id][messages.value[data.session_id].length - 1]
  if (lastMessage && lastMessage.role === 'assistant') {
    lastMessage.toolUse = data.tool
  }

  // Handle TodoWrite specifically
  if (data.tool && data.tool.includes('TodoWrite')) {
    console.log('TodoWrite tool used with data:', data)

    // Try to extract todos from the data.input property (for new format)
    let todos: TodoItem[] | null = null

    // If data has input with todos (new enhanced format), use that
    if (data.input && typeof data.input === 'object' && data.input.todos) {
      todos = data.input.todos
      console.log('Extracted todos from data.input:', todos)
    } else {
      // Try legacy parsing from tool string representation
      const toolStr = String(data.tool || '')
      todos = parseTodoWrite(toolStr)
      console.log('Parsed todos from legacy tool string:', todos)
    }

    if (todos && Array.isArray(todos)) {
      console.log('Updating session', data.session_id, 'with todos:', todos)
      updateSessionTodos(data.session_id, todos)

      // Set up auto-hide timer if all todos are completed
      const allCompleted = todos.every(todo => todo.status === 'completed')
      if (allCompleted) {
        console.log('All todos completed, setting auto-hide timer for 5 seconds')
        setTimeout(() => {
          // Clear todos after delay, only if all are still completed
          const currentTodos = sessionTodos.value.get(data.session_id)
          if (currentTodos && currentTodos.every(todo => todo.status === 'completed')) {
            sessionTodos.value.delete(data.session_id)
            console.log('Auto-hiding todos for session', data.session_id)
          }
        }, 5000)
      }
    }
  } else {
    // Parse tool execution from the tool use data (for non-TodoWrite tools)
    const toolExecution = parseToolUse(data.tool || '')
    if (toolExecution) {
      updateSessionToolExecution(data.session_id, toolExecution)
    }
  }
})

agentWs.on('onPermissionRequest', (data) => {
  // Track permission request for metrics
  const permStats = sessionPermissionStats.value.get(data.session_id) || { approved: 0, denied: 0, total: 0 }
  permStats.total++
  sessionPermissionStats.value.set(data.session_id, permStats)

  // Add to session-specific permissions map
  const sessionPerms = sessionPermissions.value.get(data.session_id) || []
  sessionPerms.push({
    ...data,
    timestamp: new Date()
  })
  sessionPermissions.value.set(data.session_id, sessionPerms)
})

agentWs.on('onPermissionAcknowledged', (data) => {
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

      // Mark the last assistant message as complete (not streaming) so new messages
      // after tool execution don't get appended to it
      const lastMessage = messages.value[data.session_id].findLast(m => m.role === 'assistant')
      if (lastMessage && lastMessage.streaming) {
        lastMessage.streaming = false
      }
    }

    // Auto-scroll to bottom if user is near bottom
    autoScrollIfNearBottom(messagesContainer.value)
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
  sessions.value = data.sessions
})

agentWs.on('onSessionDeleted', (data) => {
  console.log('🗑️ Session deleted:', data.session_id)
  // Session already removed from local state in deleteSession (optimistic update)
  // Just log confirmation
})

agentWs.on('onAllSessionsDeleted', (data) => {
  console.log('🗑️ All sessions deleted, count:', data.count)

  // Clear all sessions and messages
  sessions.value = []
  messages.value = {}
  messagesLoaded.value.clear()
  activeSessionId.value = null
  awaitingToolResults.value.clear()

  // Clear all pending timers
  todoHideTimers.value.forEach((timer) => clearTimeout(timer))
  todoHideTimers.value.clear()

  // Clear all live agents session data
  sessionTodos.value.clear()
  sessionToolExecution.value.clear()
  sessionPermissions.value.clear()

  // Clear all session metrics
  sessionToolStats.value.clear()
  sessionPermissionStats.value.clear()

  // Show success message
  alert(`Successfully deleted ${data.count} sessions from the database`)
})

agentWs.on('onMessagesLoaded', (data) => {
  console.log('📥 Messages loaded:', data)

  if (!data.session_id || !data.messages) return

  // Debug: log sequence numbers
  console.log('Message sequences from DB:', data.messages.map((m: any) => ({ seq: m.sequence, role: m.role, content: m.content.substring(0, 50) })))

  // Convert DB messages to UI message format
  const uiMessages = data.messages.map((dbMsg: any) => ({
    id: `msg-${dbMsg.session_id}-${dbMsg.sequence}`,
    role: dbMsg.role,
    content: dbMsg.content,
    timestamp: new Date(dbMsg.timestamp),
    sequence: dbMsg.sequence,
    isHistorical: true,
    toolUse: dbMsg.tool_uses ? extractToolName(dbMsg.tool_uses) : undefined,
    thinkingContent: dbMsg.thinking_content || undefined
  }))

  // Sort messages by sequence number first, then by timestamp for stable ordering
  // This handles cases where multiple messages have the same sequence number
  uiMessages.sort((a, b) => {
    if (a.sequence !== b.sequence) {
      return a.sequence - b.sequence
    }
    // If sequence numbers are equal, sort by timestamp
    return a.timestamp.getTime() - b.timestamp.getTime()
  })

  console.log('Sorted message sequences:', uiMessages.map(m => ({ seq: m.sequence, role: m.role, content: m.content.substring(0, 50) })))

  // Set or prepend messages for the session
  if (!messages.value[data.session_id]) {
    messages.value[data.session_id] = []
  }

  // Prepend historical messages (now sorted by sequence, oldest first)
  messages.value[data.session_id] = [...uiMessages, ...messages.value[data.session_id]]

  console.log(`📥 Loaded ${uiMessages.length} historical messages for session ${data.session_id}`)
})

agentWs.on('onAgentsKilled', (data) => {

  // Clear all sessions and messages
  sessions.value = []
  messages.value = {}
  messagesLoaded.value.clear()  // Clear loaded messages tracking
  activeSessionId.value = null
  awaitingToolResults.value.clear()  // Clear all flags

  // Clear all pending timers
  todoHideTimers.value.forEach((timer) => clearTimeout(timer))
  todoHideTimers.value.clear()

  // Clear all live agents session data
  sessionTodos.value.clear()
  sessionToolExecution.value.clear()
  sessionPermissions.value.clear()  // Clear all session permissions

  // Clear all session metrics
  sessionToolStats.value.clear()
  sessionPermissionStats.value.clear()

  // Show success message
  alert(`Successfully killed ${data.killed_count} agents`)
})

// Load existing sessions on mount
onMounted(() => {
  if (agentWs.connected) {
    agentWs.send({ type: 'list_sessions' })
  }
  // Load available providers
  loadProviders()
})

// Watch for connection changes
watch(() => agentWs.connected, (connected) => {
  if (connected) {
    agentWs.send({ type: 'list_sessions' })
  }
})

// Watch for new messages and auto-scroll if user is near bottom
watch(activeMessages, () => {
  autoScrollIfNearBottom(messagesContainer.value)
}, { deep: true, flush: 'post' })
</script>

<style scoped>
/* Page-level layout styles only - component-specific styles are in their respective .vue files */

.agents-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  overflow: hidden;
}

.header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-bg);
  flex-shrink: 0;
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

.btn-delete-all,
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

.btn-delete-all {
  background: #6c757d;
}

.btn-delete-all:hover:not(:disabled) {
  background: #5a6268;
}

.btn-kill-all:hover:not(:disabled) {
  background: #c82333;
}

.btn-delete-all:disabled,
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
  min-height: 0;
}

.chat-area-with-metrics {
  flex: 1;
  display: flex;
  background: var(--bg-primary);
  overflow: hidden;
  min-height: 0;
  gap: 12px;
  padding: 12px;
}

.permission-requests {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  max-height: 200px;
  overflow-y: auto;
  flex-shrink: 0;
}
</style>
