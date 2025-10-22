import { ref, computed, watch, onMounted } from 'vue'
import type { TodoItem } from '~/utils/agents/todoParser'
import type { ActiveTool } from '~/types/agents'

interface ToolExecution {
  toolName: string
  filePath?: string
  command?: string
  pattern?: string
  timestamp: Date
}

// LocalStorage keys
const TOOL_STATS_KEY = 'cct_session_tool_stats'
const PERMISSION_STATS_KEY = 'cct_session_permission_stats'

export function useSessionState() {
  // Core session state
  const sessions = ref<any[]>([])
  const activeSessionId = ref<string | null>(null)
  const messages = ref<Record<string, any[]>>({})
  const messagesLoaded = ref(new Set<string>())
  const inputMessage = ref('')
  const isProcessing = ref(false)
  const isThinking = ref(false)

  // Modal state
  const showResumeModal = ref(false)
  const showCreateSessionModal = ref(false)
  const selectedResumeSession = ref<any | null>(null)
  const availableSessions = ref<any[]>([])
  const loadingSessions = ref(false)
  const creatingSession = ref(false)
  const resumingSession = ref(false)

  // Session interactions
  const sessionPermissions = ref(new Map<string, any[]>())
  const awaitingToolResults = ref(new Set<string>())

  // Live agents state
  const sessionTodos = ref(new Map<string, TodoItem[]>())
  const sessionToolExecution = ref(new Map<string, ToolExecution | null>())
  const todoHideTimers = ref(new Map<string, NodeJS.Timeout>())

  // Tool overlays state
  const activeTools = ref(new Map<string, ActiveTool[]>())

  // Session filtering
  const activeFilter = ref('active')

  // Session metrics - initialize from localStorage
  const sessionToolStats = ref(new Map<string, Record<string, number>>())
  const sessionPermissionStats = ref(new Map<string, { approved: number; denied: number; total: number }>())

  // Load persisted stats from localStorage on mount
  const loadPersistedStats = () => {
    if (typeof window === 'undefined') return

    try {
      // Load tool stats
      const toolStatsData = localStorage.getItem(TOOL_STATS_KEY)
      if (toolStatsData) {
        const parsed = JSON.parse(toolStatsData)
        sessionToolStats.value = new Map(Object.entries(parsed))
      }

      // Load permission stats
      const permStatsData = localStorage.getItem(PERMISSION_STATS_KEY)
      if (permStatsData) {
        const parsed = JSON.parse(permStatsData)
        sessionPermissionStats.value = new Map(Object.entries(parsed))
      }
    } catch (error) {
      console.error('Failed to load stats from localStorage:', error)
    }
  }

  // Persist stats to localStorage whenever they change
  watch(sessionToolStats, (newStats) => {
    if (typeof window === 'undefined') return

    try {
      const obj = Object.fromEntries(newStats.entries())
      localStorage.setItem(TOOL_STATS_KEY, JSON.stringify(obj))
    } catch (error) {
      console.error('Failed to save tool stats to localStorage:', error)
    }
  }, { deep: true })

  watch(sessionPermissionStats, (newStats) => {
    if (typeof window === 'undefined') return

    try {
      const obj = Object.fromEntries(newStats.entries())
      localStorage.setItem(PERMISSION_STATS_KEY, JSON.stringify(obj))
    } catch (error) {
      console.error('Failed to save permission stats to localStorage:', error)
    }
  }, { deep: true })

  // Load on mount
  onMounted(() => {
    loadPersistedStats()
  })

  // Computed: Filtered sessions
  const filteredSessions = computed(() => {
    if (activeFilter.value === 'all') {
      return sessions.value
    } else if (activeFilter.value === 'active') {
      return sessions.value.filter((s: any) => s.status !== 'ended')
    } else if (activeFilter.value === 'ended') {
      return sessions.value.filter((s: any) => s.status === 'ended')
    }
    return sessions.value
  })

  // Get count for each filter
  const getFilterCount = (filter: string) => {
    if (filter === 'all') {
      return sessions.value.length
    } else if (filter === 'active') {
      return sessions.value.filter((s: any) => s.status !== 'ended').length
    } else if (filter === 'ended') {
      return sessions.value.filter((s: any) => s.status === 'ended').length
    }
    return 0
  }

  // Computed: Session filters with counts
  const sessionFiltersWithCounts = computed(() => [
    { label: 'Active', value: 'active', count: getFilterCount('active') },
    { label: 'All', value: 'all', count: getFilterCount('all') },
    { label: 'Ended', value: 'ended', count: getFilterCount('ended') }
  ])

  // Computed: Active session
  const activeSession = computed(() =>
    sessions.value.find((s: any) => s.id === activeSessionId.value)
  )

  // Computed: Active session messages
  const activeMessages = computed(() =>
    messages.value[activeSessionId.value as string] || []
  )

  // Computed: Active session permissions
  const activeSessionPermissions = computed(() =>
    sessionPermissions.value.get(activeSessionId.value as string) || []
  )

  // Computed: Active session todos
  const activeSessionTodos = computed(() =>
    sessionTodos.value.get(activeSessionId.value as string) || []
  )

  // Computed: Active session tool execution
  const activeSessionToolExecution = computed(() =>
    sessionToolExecution.value.get(activeSessionId.value as string) || null
  )

  // Computed: Active session tools
  const activeSessionTools = computed(() =>
    activeTools.value.get(activeSessionId.value as string) || []
  )

  // Computed: Should show todo box
  const shouldShowTodoBox = computed(() => {
    const todos = activeSessionTodos.value
    return todos.length > 0 && todos.some(t => t.status !== 'completed')
  })

  return {
    // State
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

    // Computed
    filteredSessions,
    sessionFiltersWithCounts,
    activeSession,
    activeMessages,
    activeSessionPermissions,
    activeSessionTodos,
    activeSessionToolExecution,
    activeSessionTools,
    shouldShowTodoBox,

    // Helpers
    getFilterCount
  }
}
