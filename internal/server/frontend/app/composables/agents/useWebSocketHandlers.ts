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

  // Helper functions from useMessageHelpers
  parseTodoWrite: (content: string) => TodoItem[] | null
  parseToolUse: (tool: string) => ToolExecution | null
  extractToolName: (toolUses: any) => string | undefined
  isCompleteSignal: (content: any) => boolean
  extractCostData: (content: any) => any
  extractTextContent: (content: any) => string

  // Helper functions from useToolManagement
  addActiveTool: (sessionId: string, toolUse: any) => void
  completeActiveTool: (sessionId: string, toolUseId: string, isError: boolean) => void
  clearSessionToolExecution: (sessionId: string) => void
  updateSessionTodos: (sessionId: string, todos: TodoItem[]) => void
  updateSessionToolExecution: (sessionId: string, toolExecution: ToolExecution | null) => void

  // Other helper functions
  autoScrollIfNearBottom: (container: any) => void
  focusMessageInput: () => void
  messagesContainer: any
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
    parseTodoWrite,
    parseToolUse,
    extractToolName,
    isCompleteSignal,
    extractCostData,
    extractTextContent,
    addActiveTool,
    completeActiveTool,
    clearSessionToolExecution,
    updateSessionTodos,
    updateSessionToolExecution,
    autoScrollIfNearBottom,
    focusMessageInput,
    messagesContainer
  } = params

  // Setup all WebSocket event handlers
  const setupHandlers = () => {
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
        console.log('✅ Message complete (result received) - NOT ADDING TO UI')
        console.log(`📊 Current message count in UI: ${messages.value[data.session_id]?.length || 0}`)
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
        console.log('⏭️  Skipping empty/system message - NOT ADDING TO UI')
        console.log(`   textContent: "${textContent}"`)
        console.log(`   hasSystemMessage: ${textContent?.includes('SystemMessage')}`)
        console.log(`📊 Current message count in UI: ${messages.value[data.session_id]?.length || 0}`)
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
      console.log(`📊 Total messages in UI now: ${messages.value[data.session_id].length}`)

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

      // Debug: log ALL messages from DB before filtering
      console.log(`📊 Total messages from DB: ${data.messages.length}`)
      console.log('Message sequences from DB:', data.messages.map((m: any) => ({
        seq: m.sequence,
        role: m.role,
        content: m.content.substring(0, 50),
        filtered: m.role === 'system' ? '❌ FILTERED' : '✅ KEPT'
      })))

      // Calculate tool stats from loaded messages
      const toolStats: Record<string, number> = {}
      let toolCount = 0

      data.messages.forEach((dbMsg: any) => {
        // Parse tool_uses from assistant messages
        if (dbMsg.tool_uses) {
          try {
            const toolUses = typeof dbMsg.tool_uses === 'string'
              ? JSON.parse(dbMsg.tool_uses)
              : dbMsg.tool_uses

            if (Array.isArray(toolUses)) {
              toolUses.forEach((tool: any) => {
                const toolName = tool.name || tool.tool || 'Unknown'
                toolStats[toolName] = (toolStats[toolName] || 0) + 1
                toolCount++
              })
            }
          } catch (e) {
            console.warn('Failed to parse tool_uses:', e)
          }
        }
      })

      // Update session tool stats
      if (toolCount > 0) {
        sessionToolStats.value.set(data.session_id, toolStats)
        console.log(`🔧 Calculated tool stats for session ${data.session_id}:`, toolStats)
        console.log(`🔧 Total tool executions: ${toolCount}`)
      } else {
        console.log(`🔧 No tool uses found in loaded messages for session ${data.session_id}`)
      }

      // Convert DB messages to UI message format, filtering out system messages
      const beforeFilter = data.messages.length
      const uiMessages = data.messages
        .filter((dbMsg: any) => {
          const keep = dbMsg.role !== 'system'
          if (!keep) {
            console.log(`🚫 Filtering out system message (seq: ${dbMsg.sequence}): ${dbMsg.content.substring(0, 80)}`)
          }
          return keep
        })
        .map((dbMsg: any) => ({
          id: `msg-${dbMsg.session_id}-${dbMsg.sequence}`,
          role: dbMsg.role,
          content: dbMsg.content,
          timestamp: new Date(dbMsg.timestamp),
          sequence: dbMsg.sequence,
          isHistorical: true,
          toolUse: dbMsg.tool_uses ? extractToolName(dbMsg.tool_uses) : undefined,
          thinkingContent: dbMsg.thinking_content || undefined
        }))

      const afterFilter = uiMessages.length
      console.log(`🔍 Filtered ${beforeFilter - afterFilter} system messages, kept ${afterFilter} messages`)

      // Sort messages by sequence number first, then by timestamp for stable ordering
      // This handles cases where multiple messages have the same sequence number
      uiMessages.sort((a, b) => {
        if (a.sequence !== b.sequence) {
          return a.sequence - b.sequence
        }
        // If sequence numbers are equal, sort by timestamp
        return a.timestamp.getTime() - b.timestamp.getTime()
      })

      console.log('✅ Sorted message sequences:', uiMessages.map(m => ({
        seq: m.sequence,
        role: m.role,
        content: m.content.substring(0, 50)
      })))

      // Set or prepend messages for the session
      if (!messages.value[data.session_id]) {
        messages.value[data.session_id] = []
      }

      const existingCount = messages.value[data.session_id].length

      // Prepend historical messages (now sorted by sequence, oldest first)
      messages.value[data.session_id] = [...uiMessages, ...messages.value[data.session_id]]

      console.log(`📥 Loaded ${uiMessages.length} historical messages for session ${data.session_id}`)
      console.log(`📊 Total messages in UI now: ${messages.value[data.session_id].length} (${existingCount} existing + ${uiMessages.length} loaded)`)
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
  }

  return {
    setupHandlers
  }
}
