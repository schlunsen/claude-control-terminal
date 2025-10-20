import { type Ref } from 'vue'

interface SessionActionParams {
  // WebSocket
  agentWs: any

  // State from useSessionState
  sessions: Ref<any[]>
  activeSessionId: Ref<string | null>
  messages: Ref<Record<string, any[]>>
  messagesLoaded: Ref<Set<string>>
  showCreateSessionModal: Ref<boolean>
  showResumeModal: Ref<boolean>
  selectedResumeSession: Ref<any | null>
  availableSessions: Ref<any[]>
  loadingSessions: Ref<boolean>
  creatingSession: Ref<boolean>
  resumingSession: Ref<boolean>
  sessionPermissions: Ref<Map<string, any[]>>
  awaitingToolResults: Ref<Set<string>>
  todoHideTimers: Ref<Map<string, NodeJS.Timeout>>
  sessionToolStats: Ref<Map<string, Record<string, number>>>
  sessionPermissionStats: Ref<Map<string, { approved: number; denied: number; total: number }>>
  isUserNearBottom: Ref<boolean>

  // State from useAgentProviders
  sessionForm: Ref<any>
  resumeForm: Ref<any>
  availableAgents: Ref<any[]>
  selectedAgentPreview: Ref<any | null>
  loadingAgents: Ref<boolean>
  availableProviders: Ref<any[]>
  currentProvider: Ref<any | null>
  loadingProviders: Ref<boolean>

  // Helper functions
  scrollToBottom: (container: any, smooth?: boolean) => void
  focusMessageInput: () => void
  cleanupSessionData: (sessionId: string) => void
}

export function useSessionActions(params: SessionActionParams) {
  const {
    agentWs,
    sessions,
    activeSessionId,
    messages,
    messagesLoaded,
    showCreateSessionModal,
    showResumeModal,
    selectedResumeSession,
    availableSessions,
    loadingSessions,
    creatingSession,
    resumingSession,
    sessionPermissions,
    awaitingToolResults,
    todoHideTimers,
    sessionToolStats,
    sessionPermissionStats,
    isUserNearBottom,
    sessionForm,
    resumeForm,
    availableAgents,
    selectedAgentPreview,
    loadingAgents,
    availableProviders,
    currentProvider,
    loadingProviders,
    scrollToBottom,
    focusMessageInput,
    cleanupSessionData
  } = params

  // Create new session
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

  // Create session with options
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

  // Load available agents
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

  // Load available providers
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

  // Handle working directory change
  const handleWorkingDirectoryChange = async () => {
    if (sessionForm.value.workingDirectory) {
      await loadAvailableAgents()
      // Clear selected agent when working directory changes
      sessionForm.value.selectedAgent = ''
      selectedAgentPreview.value = null
    }
  }

  // Load selected agent preview
  const loadSelectedAgent = async () => {
    if (!sessionForm.value.selectedAgent) {
      selectedAgentPreview.value = null
      return
    }

    try {
      const response = await fetch(`/api/agents/${sessionForm.value.selectedAgent}`)
      if (response.ok) {
        selectedAgentPreview.value = await response.json()
      }
    } catch (error) {
      console.error('Error loading agent preview:', error)
    }
  }

  // Select session
  const selectSession = (sessionId: string) => {
    activeSessionId.value = sessionId

    // Load historical messages if not already loaded
    if (!messagesLoaded.value.has(sessionId)) {
      console.log(`Loading messages for session ${sessionId}`)
      agentWs.send({
        type: 'load_messages',
        session_id: sessionId,
        limit: 100,
        offset: 0
      })
      messagesLoaded.value.add(sessionId)
    }

    // Reset scroll state and scroll to bottom when switching sessions
    isUserNearBottom.value = true
    scrollToBottom(null, false)

    // Focus the input when switching to a session
    focusMessageInput()
  }

  // End session
  const endSession = async (sessionId: string) => {
    if (!agentWs.connected) return

    agentWs.send({
      type: 'end_session',
      session_id: sessionId
    })

    // Remove from local state
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

  // Delete session
  const deleteSession = async (sessionId: string) => {
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

  // Load available sessions for resume
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

  // Open resume modal
  const openResumeModal = () => {
    showResumeModal.value = true
  }

  // Select session for resume
  const selectSessionForResume = async (session: any) => {
    selectedResumeSession.value = session

    // Prefill the form with the session's data
    resumeForm.value = {
      workingDirectory: session.working_directory || '',
      permissionMode: 'default',
      systemPrompt: '',
      tools: ['Read', 'Write', 'Edit', 'Bash', 'Search', 'TodoWrite']
    }
  }

  // Resume session with options
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
        resumeData.messages.forEach((msg: any) => {
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

  return {
    createNewSession,
    createSessionWithOptions,
    loadAvailableAgents,
    loadProviders,
    handleWorkingDirectoryChange,
    loadSelectedAgent,
    selectSession,
    endSession,
    deleteSession,
    loadAvailableSessions,
    openResumeModal,
    selectSessionForResume,
    resumeSessionWithOptions
  }
}
