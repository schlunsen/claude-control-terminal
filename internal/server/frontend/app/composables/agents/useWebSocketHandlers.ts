import { type Ref } from 'vue'
import type { TodoItem } from '~/utils/agents/todoParser'
import type { ActiveTool } from '~/types/agents'

interface ToolExecution {
  toolName: string
  filePath?: string
  command?: string
  pattern?: string
  detail?: string
  timestamp?: Date
}

interface WebSocketHandlerParams {
  // WebSocket instance
  agentWs: any

  // State from useSessionState
  sessions: Ref<any[]>
  activeSessionId: Ref<string | null>
  messages: Ref<Record<string, any[]>>
  messagesLoaded: Ref<Set<string>>
  isProcessing: Ref<boolean>
  isThinking: Ref<boolean>
  sessionPermissions: Ref<Map<string, any[]>>
  awaitingToolResults: Ref<Set<string>>
  sessionTodos: Ref<Map<string, TodoItem[]>>
  sessionToolExecution: Ref<Map<string, ToolExecution | null>>
  todoHideTimers: Ref<Map<string, NodeJS.Timeout>>
  activeTools: Ref<Map<string, ActiveTool[]>>
  sessionToolStats: Ref<Map<string, Record<string, number>>>
  sessionPermissionStats: Ref<Map<string, { approved: number; denied: number; total: number }>>

  // Helper functions
  addActiveTool: (sessionId: string, toolUse: any) => void
  completeActiveTool: (sessionId: string, toolUseId: string, isError: boolean) => void
  clearSessionToolExecution: (sessionId: string) => void
  updateSessionTodos: (sessionId: string, todos: TodoItem[]) => void
  updateSessionToolExecution: (sessionId: string, toolExecution: ToolExecution) => void
  parseTodoWrite: (content: string) => TodoItem[] | null
  parseToolUse: (tool: string) => ToolExecution | undefined
  extractToolName: (toolUses: string) => string | undefined
  isCompleteSignal: (content: any) => boolean
  extractCostData: (content: any) => any
  extractTextContent: (content: any) => string
  autoScrollIfNearBottom: (container: any) => void
  focusMessageInput: () => void
}

export function useWebSocketHandlers(params: WebSocketHandlerParams) {
  const {
    agentWs,
    sessions,
    activeSessionId,
    messages,
    messagesLoaded,
    isProcessing,
    isThinking,
    sessionPermissions,
    awaitingToolResults,
    sessionTodos,
    sessionToolExecution,
    todoHideTimers,
    activeTools,
    sessionToolStats,
    sessionPermissionStats,
    addActiveTool,
    completeActiveTool,
    clearSessionToolExecution,
    updateSessionTodos,
    updateSessionToolExecution,
    parseTodoWrite,
    parseToolUse,
    extractToolName,
    isCompleteSignal,
    extractCostData,
    extractTextContent,
    autoScrollIfNearBottom,
    focusMessageInput
  } = params

  // Setup all WebSocket event handlers
  const setupHandlers = () => {
    // Session created
    agentWs.on('onSessionCreated', (data: any) => {
      sessions.value.push(data.session)
      activeSessionId.value = data.session_id
      messages.value[data.session_id] = []

      // Mark new session as loaded (it has no history to load)
      messagesLoaded.value.add(data.session_id)

      // Focus the input after session creation
      focusMessageInput()
    })

    // Agent message received
    agentWs.on('agent_message', (data: any) => {
      const sessionId = data.session_id

      console.log('Agent message:', data)

      if (data.content) {
        if (!messages.value[sessionId]) {
          messages.value[sessionId] = []
        }

        // Add agent message
        messages.value[sessionId].push({
          id: crypto.randomUUID(),
          role: 'assistant',
          content: data.content,
          timestamp: new Date(),
          type: data.type
        })
      }

      // Update session in list
      const session = sessions.value.find(s => s.id === sessionId)
      if (session) {
        session.message_count = messages.value[sessionId]?.length || 0
      }

      // Hide tool execution overlay when assistant message arrives
      sessionToolExecution.value.set(sessionId, null)

      // Auto-scroll if user is near bottom
      if (isUserNearBottom.value) {
        scrollToBottom(null, true)
      }
    })

    // Agent thinking
    agentWs.on('agent_thinking', (data: any) => {
      console.log('Agent thinking:', data)
      isThinking.value = data.thinking
    })

    // Agent tool use
    agentWs.on('agent_tool_use', (data: any) => {
      const sessionId = data.session_id
      console.log('Agent tool use:', data)

      // Track tool statistics
      if (!sessionToolStats.value.has(sessionId)) {
        sessionToolStats.value.set(sessionId, {})
      }
      const stats = sessionToolStats.value.get(sessionId)!
      stats[data.tool_name] = (stats[data.tool_name] || 0) + 1

      // Handle TodoWrite specially
      if (data.tool_name === 'TodoWrite' && data.input?.todos) {
        const todos = parseTodoWrite(data.input.todos)
        sessionTodos.value.set(sessionId, todos)

        // Clear any existing hide timer
        const existingTimer = todoHideTimers.value.get(sessionId)
        if (existingTimer) {
          clearTimeout(existingTimer)
        }

        // Check if all todos are completed
        const allCompleted = todos.every(t => t.status === 'completed')
        if (allCompleted && todos.length > 0) {
          // Hide after 3 seconds
          const timer = setTimeout(() => {
            sessionTodos.value.set(sessionId, [])
            todoHideTimers.value.delete(sessionId)
          }, 3000)
          todoHideTimers.value.set(sessionId, timer)
        }
      } else {
        // For other tools, show execution overlay
        const toolExecution: ToolExecution = {
          toolName: data.tool_name,
          timestamp: new Date()
        }

        // Extract relevant parameters based on tool
        if (data.input) {
          if (data.tool_name === 'Read' || data.tool_name === 'Write' || data.tool_name === 'Edit') {
            toolExecution.filePath = data.input.file_path
          } else if (data.tool_name === 'Bash') {
            toolExecution.command = data.input.command
          } else if (data.tool_name === 'Grep') {
            toolExecution.pattern = data.input.pattern
          }
        }

        sessionToolExecution.value.set(sessionId, toolExecution)
      }

      // Track active tools for overlays
      if (!activeTools.value.has(sessionId)) {
        activeTools.value.set(sessionId, [])
      }

      const tools = activeTools.value.get(sessionId)!
      const toolId = data.tool_use_id || crypto.randomUUID()

      tools.push({
        id: toolId,
        name: data.tool_name,
        input: data.input,
        timestamp: new Date()
      })

      // Keep only last 5 tools
      if (tools.length > 5) {
        tools.shift()
      }

      // Auto-scroll if user is near bottom
      if (isUserNearBottom.value) {
        scrollToBottom(null, true)
      }
    })

    // Permission request
    agentWs.on('permission_request', (data: any) => {
      const sessionId = data.session_id
      console.log('Permission request:', data)

      // Track permission statistics
      if (!sessionPermissionStats.value.has(sessionId)) {
        sessionPermissionStats.value.set(sessionId, {
          approved: 0,
          denied: 0,
          total: 0
        })
      }
      const permStats = sessionPermissionStats.value.get(sessionId)!
      permStats.total++

      if (!sessionPermissions.value.has(sessionId)) {
        sessionPermissions.value.set(sessionId, [])
      }

      const permissions = sessionPermissions.value.get(sessionId)!
      permissions.push({
        id: data.permission_id,
        tool_name: data.tool_name,
        tool_input: data.tool_input,
        timestamp: new Date(),
        status: 'pending'
      })

      awaitingToolResults.value.add(sessionId)

      // Auto-scroll if user is near bottom
      if (isUserNearBottom.value) {
        scrollToBottom(null, true)
      }
    })

    // Permission acknowledged
    agentWs.on('permission_acknowledged', (data: any) => {
      const sessionId = data.session_id
      console.log('Permission acknowledged:', data)

      const permissions = sessionPermissions.value.get(sessionId)
      if (permissions) {
        const permission = permissions.find(p => p.id === data.permission_id)
        if (permission) {
          permission.status = data.approved ? 'approved' : 'denied'

          // Update permission statistics
          const permStats = sessionPermissionStats.value.get(sessionId)
          if (permStats) {
            if (data.approved) {
              permStats.approved++
            } else {
              permStats.denied++
            }
          }
        }
      }

      awaitingToolResults.value.delete(sessionId)
    })

    // Error
    agentWs.on('error', (data: any) => {
      console.error('Agent error:', data)

      const sessionId = data.session_id || activeSessionId.value

      if (sessionId && messages.value[sessionId]) {
        messages.value[sessionId].push({
          id: crypto.randomUUID(),
          role: 'error',
          content: data.message || 'An error occurred',
          timestamp: new Date(),
          isError: true
        })
      }

      isProcessing.value = false
      isThinking.value = false

      // Auto-scroll if user is near bottom
      if (isUserNearBottom.value) {
        scrollToBottom(null, true)
      }
    })

    // Sessions list
    agentWs.on('sessions_list', (data: any) => {
      console.log('Sessions list:', data)

      if (data.sessions) {
        // Update sessions with latest data from server
        sessions.value = data.sessions.map((s: any) => ({
          ...s,
          // Preserve existing UI state if session already exists
          ...(sessions.value.find(existing => existing.id === s.id) || {})
        }))

        // Initialize messages for sessions that don't have any
        data.sessions.forEach((session: any) => {
          if (!messages.value[session.id]) {
            messages.value[session.id] = []
          }
        })
      }
    })

    // Session deleted
    agentWs.on('session_deleted', (data: any) => {
      console.log('Session deleted:', data)

      // Server has confirmed deletion, update local state
      sessions.value = sessions.value.filter(s => s.id !== data.session_id)
      delete messages.value[data.session_id]
      messagesLoaded.value.delete(data.session_id)

      if (activeSessionId.value === data.session_id) {
        activeSessionId.value = null
      }
    })

    // All sessions deleted
    agentWs.on('all_sessions_deleted', () => {
      console.log('All sessions deleted')
      sessions.value = []
      messages.value = {}
      messagesLoaded.value.clear()
      activeSessionId.value = null
    })

    // Messages loaded
    agentWs.on('messages_loaded', (data: any) => {
      console.log('Messages loaded:', data)

      if (data.messages && data.session_id) {
        const sessionId = data.session_id

        // Convert historical messages to chat format
        const historicalMessages = data.messages.map((msg: any) => ({
          id: crypto.randomUUID(),
          role: msg.role,
          content: msg.content,
          timestamp: new Date(msg.timestamp),
          isHistorical: true
        }))

        // Replace messages for this session
        messages.value[sessionId] = historicalMessages
        messagesLoaded.value.add(sessionId)

        // Auto-scroll to bottom after loading
        scrollToBottom(null, false)
      }
    })

    // Agents killed
    agentWs.on('agents_killed', (data: any) => {
      console.log('Agents killed:', data)

      // Update all sessions to ended status
      sessions.value.forEach(session => {
        if (session.status === 'active') {
          session.status = 'ended'
        }
      })

      isProcessing.value = false
      isThinking.value = false
    })
  }

  // Initialize handlers when WebSocket connects
  const initializeHandlers = () => {
    if (agentWs.connected) {
      setupHandlers()

      // Request list of sessions
      agentWs.send({ type: 'list_sessions' })
    }
  }

  return {
    setupHandlers,
    initializeHandlers
  }
}
