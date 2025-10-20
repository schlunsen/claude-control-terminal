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
    messagesContainer
  } = params

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

  // Permission request functionality
  const approvePermission = (request: any) => {
    sendPermissionResponse(request, true)
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

  return {
    sendMessage,
    approvePermission,
    denyPermission,
    sendPermissionResponse,
    deleteAllSessions,
    killAllAgents
  }
}
