import type { WebSocketMessage, ActivityItem } from '~/types/analytics'

export const useWebSocket = () => {
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)

  // Event callback registry
  const callbacks = reactive<{
    onNotification: ((data: any) => void) | null
    onPrompt: ((data: any) => void) | null
    onCommand: ((data: any) => void) | null
    onStatsUpdate: ((data: any) => void) | null
    onReset: ((data: any) => void) | null
    onHistoryCleared: (() => void) | null
  }>({
    onNotification: null,
    onPrompt: null,
    onCommand: null,
    onStatsUpdate: null,
    onReset: null,
    onHistoryCleared: null,
  })

  const connect = () => {
    // Determine protocol based on current page protocol
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    
    // In development, connect directly to backend port 3333
    // In production, use the same host (proxy handles it)
    const isDev = process.dev || window.location.port === '3001' || window.location.port === '3002'
    const host = isDev ? 'localhost:3333' : window.location.host
    const wsUrl = `${protocol}//${host}/ws`

    console.log(`Connecting to WebSocket: ${wsUrl}`)
    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      connected.value = true
      console.log('✅ WebSocket connected')
    }

    ws.value.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data)

        // Handle different event types
        switch (message.event) {
          case 'notification_recorded':
            callbacks.onNotification?.(message.data)
            break

          case 'prompt_recorded':
            callbacks.onPrompt?.(message.data)
            break

          case 'command_recorded':
            callbacks.onCommand?.(message.data)
            break

          case 'claude':
            // Handle claude tool events with proper structure
            callbacks.onCommand?.({
              data: message.data,
              type: 'claude'
            })
            break

          case 'stats_updated':
            callbacks.onStatsUpdate?.(message.data)
            break

          case 'reset_archive':
          case 'reset_clear':
          case 'reset_soft':
          case 'reset_cleared':
            callbacks.onReset?.(message.data)
            break

          case 'history_cleared':
          case 'notifications_cleared':
            callbacks.onHistoryCleared?.()
            break

          default:
            console.log('Unknown WebSocket event:', message.event)
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error, event.data)
      }
    }

    ws.value.onerror = (error) => {
      console.error('❌ WebSocket error:', error)
      connected.value = false
    }

    ws.value.onclose = () => {
      connected.value = false
      console.log('🔌 WebSocket closed, reconnecting in 5s...')

      // Reconnect after 5 seconds
      setTimeout(() => {
        if (!ws.value || ws.value.readyState === WebSocket.CLOSED) {
          connect()
        }
      }, 5000)
    }
  }

  const disconnect = () => {
    if (ws.value) {
      ws.value.close()
      ws.value = null
      connected.value = false
    }
  }

  // Register event handlers
  const on = (event: keyof typeof callbacks, handler: any) => {
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
    on,
    disconnect,
    reconnect: connect
  }
}
