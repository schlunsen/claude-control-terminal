import { type Ref } from 'vue'
import type { TodoItem } from '~/utils/agents/todoParser'

interface MessagingParams {
  // WebSocket
  agentWs: any

  // State
  activeSessionId: Ref<string | null>
  inputMessage: Ref<string>
  isProcessing: Ref<boolean>
  messages: Ref<Record<string, any[]>>
  sessions: Ref<any[]>
  sessionTodos: Ref<Map<string, TodoItem[]>>
  todoHideTimers: Ref<Map<string, NodeJS.Timeout>>
  awaitingToolResults: Ref<Set<string>>
  sessionPermissions: Ref<Map<string, any[]>>
  sessionPermissionStats: Ref<Map<string, { approved: number; denied: number; total: number }>>

  // Helper functions
  autoScrollIfNearBottom: (container: any) => void
  messagesContainer: any
  refreshProjectPermissions?: () => Promise<void>
}

export function useMessaging(params: MessagingParams) {
  const {
    agentWs,
    activeSessionId,
    inputMessage,
    isProcessing,
    messages,
    sessions,
    sessionTodos,
    todoHideTimers,
    awaitingToolResults,
    sessionPermissions,
    sessionPermissionStats,
    autoScrollIfNearBottom,
    messagesContainer,
    refreshProjectPermissions
  } = params

  // Messaging
  const sendMessage = async (attachedImages: any[] = []) => {
    const hasMessage = inputMessage.value.trim()
    const hasImages = attachedImages.length > 0

    if ((!hasMessage && !hasImages) || !activeSessionId.value) return

    const message = inputMessage.value
    inputMessage.value = ''

    // Add user message to chat
    if (!messages.value[activeSessionId.value]) {
      messages.value[activeSessionId.value] = []
    }

    // Build content array for structured content (text + images)
    const content: any[] = []

    // Add text block if message exists
    if (hasMessage) {
      content.push({
        type: 'text',
        text: message
      })
    }

    // Add image blocks
    for (const img of attachedImages) {
      content.push({
        type: 'image',
        source: {
          type: 'base64',
          media_type: img.mediaType,
          data: img.base64Data
        }
      })
    }

    // Check if this is a /context command (should be hidden from chat)
    const isContextCommand = hasMessage && message.trim() === '/context'

    // Store message in local state with structured content
    // Don't add /context commands to the chat UI
    if (!isContextCommand) {
      messages.value[activeSessionId.value].push({
        id: crypto.randomUUID(),
        role: 'user',
        content: content,
        timestamp: new Date()
      })
    }

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
    // If we have structured content (text + images), send content array
    // Otherwise, send legacy prompt string for backward compatibility
    if (hasImages || content.length > 1) {
      agentWs.send({
        type: 'send_prompt',
        session_id: activeSessionId.value,
        content: content
      })
    } else {
      // Legacy format for text-only messages
      agentWs.send({
        type: 'send_prompt',
        session_id: activeSessionId.value,
        prompt: message
      })
    }
  }

  // Permission request functionality
  const approvePermission = (request: any) => {
    sendPermissionResponse(request, true)
  }

  const approvePermissionExact = async (request: any) => {
    // Optimistically remove the permission from UI before backend responds
    removePermissionFromUI(request)

    // Add an exact-match always-allow rule
    // The backend will handle approving the current permission request
    await addAlwaysAllowRule(request, 'exact')
  }

  const approvePermissionSimilar = async (request: any) => {
    // Optimistically remove the permission from UI before backend responds
    removePermissionFromUI(request)

    // Add a pattern-match always-allow rule
    // The backend will handle approving the current permission request
    await addAlwaysAllowRule(request, 'pattern')
  }

  const removePermissionFromUI = (request: any) => {
    // Remove from session-specific permissions immediately (optimistic update)
    const sessionPerms = sessionPermissions.value.get(request.session_id) || []
    sessionPermissions.value.set(
      request.session_id,
      sessionPerms.filter(p => p.request_id !== request.request_id)
    )
  }

  const addAlwaysAllowRule = async (request: any, matchMode: 'exact' | 'pattern') => {
    try {
      // Generate pattern if needed
      let pattern = null
      if (matchMode === 'pattern') {
        pattern = generatePattern(request.tool, request.details)
      }

      const rule = {
        tool: request.tool,
        match_mode: matchMode,
        parameters: matchMode === 'exact' ? request.details : undefined,
        pattern: matchMode === 'pattern' ? pattern : undefined,
        description: request.description
      }

      agentWs.send({
        type: 'add_always_allow_rule',
        session_id: request.session_id,
        rule: rule,
        permission_id: request.request_id  // Include the pending permission ID so backend can approve it
      })

      // Show confirmation message
      const modeText = matchMode === 'exact' ? 'exact match' : 'similar matches'
      if (!messages.value[request.session_id]) {
        messages.value[request.session_id] = []
      }

      messages.value[request.session_id].push({
        id: crypto.randomUUID(),
        role: 'system',
        content: `🔓 Always-allow rule added (${modeText}): ${request.description}`,
        timestamp: new Date(),
        isPermissionDecision: true
      })

      // Auto-scroll if viewing this session
      if (request.session_id === activeSessionId.value) {
        autoScrollIfNearBottom(messagesContainer)
      }

      // Refresh project permissions to show the new rule
      if (refreshProjectPermissions) {
        await refreshProjectPermissions()
      }

    } catch (error) {
      console.error('Failed to add always-allow rule:', error)
      alert('Failed to add always-allow rule. Please try again.')
    }
  }

  const generatePattern = (toolName: string, details: any) => {
    const pattern: any = {}

    // For "Allow Similar", we extract the command/path pattern
    switch (toolName) {
      case 'Bash':
        // Extract the command from the full bash command
        // e.g., "sed -i '' 's/foo/bar/g' file.txt" -> "sed"
        if (details?.command) {
          const command = details.command.trim()
          const commandName = command.split(/\s+/)[0] // Get first word
          // Backend will format this as "Bash(sed:*)"
          pattern.command_prefix = commandName
        } else {
          pattern.command_prefix = '*'
        }
        break

      case 'Read':
      case 'Write':
      case 'Edit':
        // Use wildcard pattern for all files (backend will format as ToolName(**))
        pattern.directory_path = '/**'
        break

      case 'Grep':
      case 'Glob':
        pattern.path_pattern = '*'
        break
    }

    return pattern
  }

  const denyPermission = (request: any, reason = '') => {
    sendPermissionResponse(request, false, reason)
  }

  const sendPermissionResponse = (request: any, approved: boolean, reason = '') => {
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
        autoScrollIfNearBottom(messagesContainer)
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

  // Interrupt current session
  const interruptSession = async () => {
    if (!agentWs.connected || !activeSessionId.value) return

    try {
      console.log('Interrupting session:', activeSessionId.value)
      agentWs.send({
        type: 'interrupt_session',
        session_id: activeSessionId.value
      })

      // Don't update processing state here - wait for backend confirmation
      // This ensures the ESC hint stays visible until interrupt is confirmed
      // The onSessionInterrupted handler will set isProcessing = false

      // Don't add system message here - backend will confirm the interruption
      // This prevents showing an interruption message if the backend fails to interrupt
    } catch (error) {
      console.error('Failed to interrupt session:', error)
      alert('Failed to interrupt session. Please try again.')
    }
  }

  return {
    sendMessage,
    approvePermission,
    approvePermissionExact,
    approvePermissionSimilar,
    denyPermission,
    sendPermissionResponse,
    deleteAllSessions,
    killAllAgents,
    interruptSession
  }
}
