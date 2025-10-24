import { ref, computed } from 'vue'

export function useAgentProviders() {
  // Session creation form
  const sessionForm = ref({
    workingDirectory: '',
    permissionMode: 'default',
    modelProvider: 'anthropic',
    model: 'sonnet',
    systemPrompt: '',
    promptMode: 'agent', // 'agent' or 'custom'
    selectedAgent: '',
    tools: ['Read', 'Write', 'Edit', 'Bash', 'Search', 'TodoWrite']
  })

  // Resume session form
  const resumeForm = ref({
    workingDirectory: '',
    permissionMode: 'default',
    systemPrompt: '',
    tools: ['Read', 'Write', 'Edit', 'Bash', 'Search', 'TodoWrite']
  })

  // Agent selection state
  const availableAgents = ref<any[]>([])
  const selectedAgentPreview = ref<any | null>(null)
  const loadingAgents = ref(false)

  // Provider configuration
  const availableProviders = ref<any[]>([])
  const currentProvider = ref<any | null>(null)
  const loadingProviders = ref(false)

  // Get models for selected provider
  const getProviderModels = (providerId: string) => {
    const provider = availableProviders.value.find((p: any) => p.id === providerId)
    return provider?.models || []
  }

  // Get current provider models
  const currentProviderModels = computed(() => {
    return getProviderModels(sessionForm.value.modelProvider)
  })

  return {
    // Forms
    sessionForm,
    resumeForm,

    // Agent state
    availableAgents,
    selectedAgentPreview,
    loadingAgents,

    // Provider state
    availableProviders,
    currentProvider,
    loadingProviders,

    // Computed
    currentProviderModels,

    // Helpers
    getProviderModels
  }
}
