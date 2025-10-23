import { type Ref } from 'vue'
import type { TodoItem } from '~/utils/agents/todoParser'
import type { ActiveTool } from '~/types/agents'
import { useContextUsage } from '~/composables/agents/useContextUsage'

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
  sessionContextUsage: Ref<Map<string, any>>
  contextUsageLoading: Ref<boolean>
  contextUsageTimeoutId: Ref<number | null>

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
    sessionContextUsage,
    contextUsageLoading,
    contextUsageTimeoutId,
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

  // Import context usage parser
  const { parseContextResponse } = useContextUsage()

  // Setup all WebSocket event handlers
  const setupHandlers = () => {
    // WebSocket event handlers
    agentWs.on('onSessionCreated', (data) => {
      sessions.value.unshift(data.session)
      activeSessionId.value = data.session_id
      messages.value[data.session_id] = []

      // Mark new session as loaded (it has no history to load)
      messagesLoaded.value.add(data.session_id)

      // Automatically load context window for new sessions
      // This ensures Claude has access to project context from the start
      contextUsageLoading.value = true
      agentWs.send({
        type: 'send_prompt',
        session_id: data.session_id,
        prompt: '/context'
      })

      // Focus the input after session creation
      focusMessageInput()
    })

    agentWs.on('onSessionInterrupted', (data) => {
      console.log('Session interrupted:', data.session_id)

      // Update session status to idle
      const session = sessions.value.find(s => s.id === data.session_id)
      if (session) {
        session.status = 'idle'
      }

      // Update processing state
      isProcessing.value = false
      isThinking.value = false

      // Clear any pending tool execution state for this session
      clearSessionToolExecution(data.session_id)

      // Add system message to indicate interruption (confirmed by backend)
      if (!messages.value[data.session_id]) {
        messages.value[data.session_id] = []
      }

      messages.value[data.session_id].push({
        id: crypto.randomUUID(),
        role: 'system',
        content: '⚠️ Session interrupted by user',
        timestamp: new Date(),
        isInterruption: true
      })

      // Auto-scroll to show interruption message
      autoScrollIfNearBottom(messagesContainer.value)

      // Focus the input after interruption so user can send another message
      focusMessageInput()
    })

    agentWs.on('onAgentMessage', (data) => {
      if (!messages.value[data.session_id]) {
        messages.value[data.session_id] = []
      }

      // Check if this is a completion signal (result message)
      const isComplete = isCompleteSignal(data.content)

      // Extract cost data from result messages
      const costData = extractCostData(data.content)

      // Extract text content from nested object
      const textContent = extractTextContent(data.content)

      // Check if this is a /context response and parse it
      // Also check if it's the new format with "Context Usage"
      const isContextResponse = textContent && (
        textContent.toLowerCase().includes('context window') ||
        textContent.includes('## Context Usage')
      )

      if (isContextResponse) {
        const usage = parseContextResponse(textContent)
        if (usage) {
          sessionContextUsage.value.set(data.session_id, usage)

          // Clear the timeout and reset loading state
          if (contextUsageTimeoutId.value !== null) {
            clearTimeout(contextUsageTimeoutId.value)
            contextUsageTimeoutId.value = null
          }
          contextUsageLoading.value = false

        }
        // Don't display /context responses in the chat
        return
      }

      // Generate message ID for associating tools with messages
      const messageId = `msg-${data.session_id}-${Date.now()}`

      // Process tool uses (when Claude starts using a tool)
      if (data.content && data.content.tools && Array.isArray(data.content.tools)) {
        data.content.tools.forEach((toolUse: any) => {
          addActiveTool(data.session_id, toolUse, messageId)
        })
      }

      // Process tool results (when tool execution completes)
      if (data.content && data.content.tool_results && Array.isArray(data.content.tool_results)) {
        data.content.tool_results.forEach((toolResult: any) => {
          completeActiveTool(data.session_id, toolResult.tool_use_id, toolResult.is_error || false)
        })
      }

      // Update session status and metadata
      const session = sessions.value.find(s => s.id === data.session_id)
      if (session) {
        // Update git branch from metadata
        if (data.metadata && data.metadata.git_branch) {
          session.git_branch = data.metadata.git_branch
        }

        // Update costs from result message
        if (costData) {
          session.cost_usd = (session.cost_usd || 0) + costData.costUSD
          session.num_turns = costData.numTurns
          session.duration_ms = costData.durationMs
          session.usage = costData.usage
        }

        // Set status: idle when complete, processing when receiving content
        if (isComplete) {
          session.status = 'idle'
          session.message_count = (session.message_count || 0) + 1

          // Reload context window when session goes idle (only once per session)
          // This ensures Claude has fresh context after completing a task
          if (!session._contextReloaded) {
            session._contextReloaded = true

            // Send /context command to reload context window
            setTimeout(() => {
              contextUsageLoading.value = true
              agentWs.send({
                type: 'send_prompt',
                session_id: data.session_id,
                prompt: '/context'
              })
            }, 500) // Small delay to avoid race conditions
          }
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
        }

        return
      }

      // Skip empty content and system messages (they don't need UI display)
      if (!textContent || textContent.includes('SystemMessage')) {
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
        id: messageId,  // Use the same messageId we generated earlier
        role: 'assistant',
        content: textContent,
        timestamp: new Date(),
        streaming: false,
        isToolResult: isToolResult
      }

      messages.value[data.session_id].push(newMessage)

      // Don't reset processing state when receiving content - only reset on completion
      // This ensures the ESC interrupt hint remains visible throughout the conversation
      // isProcessing will be reset when result message arrives (line 251)

      // Auto-scroll to bottom if user is near bottom
      autoScrollIfNearBottom(messagesContainer.value)
    })

    agentWs.on('onAgentThinking', (data) => {
      if (data.session_id === activeSessionId.value) {
        isThinking.value = data.thinking
        // Don't reset isProcessing when thinking stops - let the result message handle that
        // This keeps the ESC interrupt hint visible throughout the entire conversation
      }

      // Update session status based on thinking state
      const session = sessions.value.find(s => s.id === data.session_id)
      if (session) {
        session.status = data.thinking ? 'processing' : session.status
      }
    })

    agentWs.on('onAgentToolUse', (data) => {
      // Update session status to processing when tool is being used
      const session = sessions.value.find(s => s.id === data.session_id)
      if (session) {
        session.status = 'processing'
      }

      // Track tool usage for metrics
      const currentStats = sessionToolStats.value.get(data.session_id) || {}
      const toolStats = { ...currentStats, [data.tool]: (currentStats[data.tool] || 0) + 1 }
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

      // Handle Edit tool specifically - add to activeTools for overlay display AND attach to message
      if (data.tool === 'Edit' && data.parameters) {
        const params = typeof data.parameters === 'string' ? JSON.parse(data.parameters) : data.parameters

        // Find the last assistant message ID to associate the tool with
        const sessionMessages = messages.value[data.session_id] || []
        const lastAssistantMessage = sessionMessages.findLast(m => m.role === 'assistant')
        const associatedMessageId = lastAssistantMessage?.id

        console.log('Edit tool detected!')
        console.log('Last assistant message:', lastAssistantMessage)
        console.log('Associated message ID:', associatedMessageId)
        console.log('Edit params:', params)

        // Attach Edit tool data directly to the message for persistent display
        if (lastAssistantMessage) {
          lastAssistantMessage.editToolData = {
            filePath: params.file_path,
            oldString: params.old_string,
            newString: params.new_string,
            replaceAll: params.replace_all || false,
            status: 'running'
          }
          console.log('Attached editToolData to message:', lastAssistantMessage.editToolData)
        } else {
          console.warn('No last assistant message found to attach Edit data!')
        }

        // Create a tool use object for the active tools overlay
        const toolUse = {
          id: `edit-${Date.now()}`, // Generate a unique ID
          name: 'Edit',
          input: params,
          status: 'running'
        }
        addActiveTool(data.session_id, toolUse, associatedMessageId)
      }

      // Handle TodoWrite specifically
      if (data.tool && data.tool.includes('TodoWrite')) {
        // Try to extract todos from the data.input property (for new format)
        let todos: TodoItem[] | null = null

        // If data has input with todos (new enhanced format), use that
        if (data.input && typeof data.input === 'object' && data.input.todos) {
          todos = data.input.todos
        } else {
          // Try legacy parsing from tool string representation
          const toolStr = String(data.tool || '')
          todos = parseTodoWrite(toolStr)
        }

        if (todos && Array.isArray(todos)) {
          updateSessionTodos(data.session_id, todos)

          // Set up auto-hide timer if all todos are completed
          const allCompleted = todos.every(todo => todo.status === 'completed')
          if (allCompleted) {
            setTimeout(() => {
              // Clear todos after delay, only if all are still completed
              const currentTodos = sessionTodos.value.get(data.session_id)
              if (currentTodos && currentTodos.every(todo => todo.status === 'completed')) {
                sessionTodos.value.delete(data.session_id)
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
      // Backend already returns sessions ordered by updated_at DESC (newest first)
      sessions.value = data.sessions
    })

    agentWs.on('onSessionDeleted', (data) => {
      // Session already removed from local state in deleteSession (optimistic update)
    })

    agentWs.on('onAllSessionsDeleted', (data) => {

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
      if (!data.session_id || !data.messages) return

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
      }

      // Convert DB messages to UI message format, filtering out system messages and /context commands
      const uiMessages = data.messages
        .filter((dbMsg: any) => {
          // Filter out system messages
          if (dbMsg.role === 'system') return false

          // Filter out /context user messages
          if (dbMsg.role === 'user' && dbMsg.content) {
            // Handle both string and array content formats
            if (typeof dbMsg.content === 'string') {
              return dbMsg.content.trim() !== '/context'
            }
            if (Array.isArray(dbMsg.content)) {
              // Check if content is a single text block with /context
              const hasOnlyContextCommand = dbMsg.content.length === 1 &&
                dbMsg.content[0].type === 'text' &&
                dbMsg.content[0].text?.trim() === '/context'
              return !hasOnlyContextCommand
            }
          }

          return true
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

      // Sort messages by sequence number first, then by timestamp for stable ordering
      // This handles cases where multiple messages have the same sequence number
      uiMessages.sort((a, b) => {
        if (a.sequence !== b.sequence) {
          return a.sequence - b.sequence
        }
        // If sequence numbers are equal, sort by timestamp
        return a.timestamp.getTime() - b.timestamp.getTime()
      })

      // Set or prepend messages for the session
      if (!messages.value[data.session_id]) {
        messages.value[data.session_id] = []
      }

      // Prepend historical messages (now sorted by sequence, oldest first)
      messages.value[data.session_id] = [...uiMessages, ...messages.value[data.session_id]]
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

    agentWs.on('onError', (data) => {
      console.error('WebSocket error:', data)

      // Check if this is a connection loss error
      if (data.message && data.message.includes('WebSocket connection lost')) {
        console.warn('WebSocket disconnected, clearing all pending permissions')

        // Clear all pending permissions from all sessions
        sessionPermissions.value.clear()

        // Optionally show a toast notification to the user
        // This helps them understand why permissions might disappear
        if (typeof window !== 'undefined' && activeSessionId.value) {
          console.warn('⚠️ Connection lost - pending permission requests have been cleared')
        }
      }
    })
  }

  return {
    setupHandlers
  }
}
