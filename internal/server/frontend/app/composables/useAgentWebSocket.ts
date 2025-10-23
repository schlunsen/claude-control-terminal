import type { Ref } from 'vue'

interface AgentWebSocketCallbacks {
  onSessionCreated: ((data: any) => void) | null
  onSessionInterrupted: ((data: any) => void) | null
  onAgentMessage: ((data: any) => void) | null
  onAgentThinking: ((data: any) => void) | null
  onAgentToolUse: ((data: any) => void) | null
  onPermissionRequest: ((data: any) => void) | null
  onPermissionAcknowledged: ((data: any) => void) | null
  onSessionsList: ((data: any) => void) | null
  onMessagesLoaded: ((data: any) => void) | null
  onSessionDeleted: ((data: any) => void) | null
  onAllSessionsDeleted: ((data: any) => void) | null
  onAgentsKilled: ((data: any) => void) | null
  onError: ((data: any) => void) | null
}

export const useAgentWebSocket = () => {
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  const authenticated = ref(false)
  const reconnectTimer = ref<ReturnType<typeof setTimeout> | null>(null)

  // Event callback registry
  const callbacks = reactive<AgentWebSocketCallbacks>({
    onSessionCreated: null,
    onSessionInterrupted: null,
    onAgentMessage: null,
    onAgentThinking: null,
    onAgentToolUse: null,
    onPermissionRequest: null,
    onPermissionAcknowledged: null,
    onSessionsList: null,
    onMessagesLoaded: null,
    onSessionDeleted: null,
    onAllSessionsDeleted: null,
    onAgentsKilled: null,
    onError: null,
  })

  const connect = async () => {
    // Get API key from analytics secret file
    try {
      const response = await fetch('/api/config/api-key')
      const data = await response.json()
      const apiKey = data.apiKey

      if (!apiKey) {
        console.error('No API key found for agent server')
        return
      }

      // Determine protocol based on current page protocol
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'

      // Connect to unified server's agent WebSocket endpoint at /agent/ws
      // The analytics server directly handles agent functionality (no proxy)
      const host = window.location.host
      const path = '/agent/ws'
      const wsUrl = `${protocol}//${host}${path}?token=${apiKey}`

      ws.value = new WebSocket(wsUrl)

      ws.value.onopen = () => {
        connected.value = true

        // Clear reconnect timer
        if (reconnectTimer.value) {
          clearTimeout(reconnectTimer.value)
          reconnectTimer.value = null
        }

        // Automatically request session list when connected
        ws.value?.send(JSON.stringify({ type: 'list_sessions' }))
      }

      ws.value.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)

          // Handle authentication success
          if (message.type === 'auth_success') {
            authenticated.value = true
            return
          }

          // Route to appropriate handler
          switch (message.type) {
            case 'session_created':
              callbacks.onSessionCreated?.(message)
              break

            case 'session_interrupted':
              callbacks.onSessionInterrupted?.(message)
              break

            case 'agent_message':
              callbacks.onAgentMessage?.(message)
              break

            case 'agent_thinking':
              callbacks.onAgentThinking?.(message)
              break

            case 'agent_tool_use':
              callbacks.onAgentToolUse?.(message)
              break

            case 'permission_request':
              callbacks.onPermissionRequest?.(message)
              break

            case 'permission_acknowledged':
              callbacks.onPermissionAcknowledged?.(message)
              break

            case 'sessions_list':
              callbacks.onSessionsList?.(message)
              break

            case 'messages_loaded':
              callbacks.onMessagesLoaded?.(message)
              break

            case 'session_deleted':
              callbacks.onSessionDeleted?.(message)
              break

            case 'all_sessions_deleted':
              callbacks.onAllSessionsDeleted?.(message)
              break

            case 'agents_killed':
              callbacks.onAgentsKilled?.(message)
              break

            case 'error':
              callbacks.onError?.(message)
              break

            default:
              console.warn('Unknown agent message type:', message.type)
          }
        } catch (error) {
          console.error('Error parsing agent WebSocket message:', error)
        }
      }

      ws.value.onerror = (error) => {
        console.error('Agent WebSocket error:', error)
        connected.value = false
        authenticated.value = false
      }

      ws.value.onclose = () => {
        connected.value = false
        authenticated.value = false

        // Notify about connection loss via error callback
        // This allows components to clean up stale state like pending permissions
        if (callbacks.onError) {
          callbacks.onError({
            type: 'error',
            message: 'WebSocket connection lost - pending permissions have been cleared'
          })
        }

        // Reconnect after 5 seconds
        if (!reconnectTimer.value) {
          reconnectTimer.value = setTimeout(() => {
            reconnectTimer.value = null
            if (!ws.value || ws.value.readyState === WebSocket.CLOSED) {
              connect()
            }
          }, 5000)
        }
      }

    } catch (error) {
      console.error('Failed to connect to agent server:', error)
    }
  }

  const disconnect = () => {
    if (reconnectTimer.value) {
      clearTimeout(reconnectTimer.value)
      reconnectTimer.value = null
    }

    if (ws.value) {
      ws.value.close()
      ws.value = null
      connected.value = false
      authenticated.value = false
    }
  }

  const send = (data: any) => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      console.warn('Cannot send message, WebSocket not connected')
      return false
    }

    try {
      ws.value.send(JSON.stringify(data))
      return true
    } catch (error) {
      console.error('Error sending message:', error)
      return false
    }
  }

  // Register event handlers
  const on = (event: keyof AgentWebSocketCallbacks, handler: any) => {
    callbacks[event] = handler
  }

  // Auto-connect on mount, disconnect on unmount
  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected: readonly(connected),
    authenticated: readonly(authenticated),
    on,
    send,
    disconnect,
    reconnect: connect
  }
}