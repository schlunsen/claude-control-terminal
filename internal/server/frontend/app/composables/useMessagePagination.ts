/**
 * Composable for managing message pagination and lazy loading
 * Handles loading messages in chunks and infinite scroll behavior
 */

import type { Ref } from 'vue'

export interface Message {
  id: string
  session_id: string
  sequence: number
  role: 'user' | 'assistant' | 'system'
  content: string
  thinking_content?: string
  tool_uses?: any
  timestamp: string
  tokens_used: number
}

export interface MessagePaginationState {
  messages: Message[]
  isLoading: boolean
  hasMore: boolean
  currentOffset: number
  limit: number
  error: string | null
}

export function useMessagePagination(sessionId: Ref<string | null>, limit = 50) {
  const state = reactive<MessagePaginationState>({
    messages: [],
    isLoading: false,
    hasMore: true,
    currentOffset: 0,
    limit,
    error: null
  })

  // Load initial messages
  const loadInitialMessages = async () => {
    if (!sessionId.value) return

    state.messages = []
    state.currentOffset = 0
    state.hasMore = true
    state.error = null

    await loadMoreMessages()
  }

  // Load more messages (for infinite scroll)
  const loadMoreMessages = async () => {
    if (!sessionId.value || state.isLoading || !state.hasMore) return

    state.isLoading = true
    state.error = null

    try {
      const response = await $fetch<{
        messages: Message[]
        has_more: boolean
        count: number
        offset: number
      }>(`/api/agent/sessions/${sessionId.value}/messages`, {
        params: {
          limit: state.limit,
          offset: state.currentOffset
        }
      })

      if (response.messages && response.messages.length > 0) {
        // Prepend messages (oldest first for scrolling up)
        state.messages = [...response.messages, ...state.messages]
        state.currentOffset += response.messages.length
        state.hasMore = response.has_more
      } else {
        state.hasMore = false
      }
    } catch (err: any) {
      state.error = err.message || 'Failed to load messages'
      console.error('Failed to load messages:', err)
    } finally {
      state.isLoading = false
    }
  }

  // Load messages via WebSocket (alternative method)
  const loadMessagesViaWS = (ws: any) => {
    if (!sessionId.value || state.isLoading || !state.hasMore) return

    state.isLoading = true

    const message = {
      type: 'load_messages',
      session_id: sessionId.value,
      limit: state.limit,
      offset: state.currentOffset
    }

    ws.send(JSON.stringify(message))
  }

  // Handle WebSocket response
  const handleMessagesLoaded = (data: any) => {
    if (data.messages && data.messages.length > 0) {
      // Prepend messages (oldest first)
      state.messages = [...data.messages, ...state.messages]
      state.currentOffset += data.messages.length
      state.hasMore = data.has_more
    } else {
      state.hasMore = false
    }
    state.isLoading = false
  }

  // Scroll event handler with debouncing
  let scrollTimeout: NodeJS.Timeout | null = null

  const handleScroll = (element: HTMLElement, threshold = 100) => {
    if (scrollTimeout) clearTimeout(scrollTimeout)

    scrollTimeout = setTimeout(() => {
      // Check if scrolled near top
      const scrolledToTop = element.scrollTop < threshold

      if (scrolledToTop && !state.isLoading && state.hasMore) {
        const previousHeight = element.scrollHeight
        const previousScrollTop = element.scrollTop

        loadMoreMessages().then(() => {
          // Maintain scroll position after prepending messages
          nextTick(() => {
            const newHeight = element.scrollHeight
            const heightDiff = newHeight - previousHeight
            element.scrollTop = previousScrollTop + heightDiff
          })
        })
      }
    }, 150) // Debounce 150ms
  }

  // Reset pagination state
  const reset = () => {
    state.messages = []
    state.currentOffset = 0
    state.hasMore = true
    state.isLoading = false
    state.error = null
  }

  // Add a new message to the list (for real-time updates)
  const addMessage = (message: Message) => {
    // Check if message already exists
    const exists = state.messages.some(m => m.id === message.id)
    if (!exists) {
      state.messages.push(message)
    }
  }

  return {
    // State
    messages: computed(() => state.messages),
    isLoading: computed(() => state.isLoading),
    hasMore: computed(() => state.hasMore),
    error: computed(() => state.error),

    // Methods
    loadInitialMessages,
    loadMoreMessages,
    loadMessagesViaWS,
    handleMessagesLoaded,
    handleScroll,
    reset,
    addMessage
  }
}
