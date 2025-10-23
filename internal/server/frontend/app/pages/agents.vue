<template>
  <div class="agents-page">
    <!-- Header -->
    <header class="header">
      <div class="header-content">
        <div class="header-text">
          <h1>Live Agents</h1>
          <p class="subtitle">Interactive Claude agent conversations</p>
        </div>
        <div class="header-actions">
          <button
            @click="deleteAllSessions"
            class="btn-delete-all"
            :disabled="!agentWs.connected || sessions.length === 0"
            title="Delete all sessions from database"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              <line x1="10" y1="11" x2="10" y2="17"></line>
              <line x1="14" y1="11" x2="14" y2="17"></line>
            </svg>
            Delete All Sessions
          </button>
          <button
            @click="killAllAgents"
            class="btn-kill-all"
            :disabled="!agentWs.connected || sessions.length === 0"
            title="Kill all active agents"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="15" y1="9" x2="9" y2="15"></line>
              <line x1="9" y1="9" x2="15" y2="15"></line>
            </svg>
            Kill All Agents
          </button>
        </div>
      </div>
    </header>

    <!-- Connection Status -->
    <div class="connection-status" :class="{ connected: agentWs.connected }">
      <div class="status-indicator"></div>
      <span>{{ agentWs.connected ? 'Connected' : 'Disconnected' }}</span>
    </div>

    <div class="agents-container">
      <!-- Sessions Sidebar -->
      <SessionsSidebar
        :sessions="filteredSessions"
        :active-session-id="activeSessionId"
        :active-filter="activeFilter"
        :filters="sessionFiltersWithCounts"
        :connected="agentWs.connected"
        :creating="creatingSession"
        @create-new="createNewSession"
        @resume="openResumeModal"
        @update:active-filter="activeFilter = $event"
        @select="selectSession"
        @end="endSession"
        @delete="deleteSession"
      />

      <!-- Chat Area with Metrics -->
      <main class="chat-area-with-metrics">
        <ChatArea
          ref="chatAreaRef"
          :has-active-session="!!activeSessionId"
          v-model:input-message="inputMessage"
          :connected="agentWs.connected"
          :is-thinking="isThinking"
          :is-processing="isProcessing"
          @send="handleSendMessage"
        >
          <!-- Tool Overlays Slot -->
          <template #tool-overlays>
            <ToolOverlaysContainer :tools="activeSessionTools">
              <template v-for="tool in activeSessionTools" :key="tool.id">
                <TodoWriteOverlay
                  v-if="tool.name === 'TodoWrite'"
                  :tool="tool"
                  @dismiss="removeActiveTool(tool.sessionId, $event)"
                />
                <EditDiffOverlay
                  v-else-if="tool.name === 'Edit' && diffDisplayLocation === 'options'"
                  :tool="tool"
                  @dismiss="removeActiveTool(tool.sessionId, $event)"
                />
                <ToolOverlay
                  v-else-if="tool.name !== 'Edit'"
                  :tool="tool"
                  @dismiss="removeActiveTool(tool.sessionId, $event)"
                />
              </template>
            </ToolOverlaysContainer>
          </template>

          <!-- TodoWrite Box Slot -->
          <template #todo-box>
            <TodoWriteBox
              :show="shouldShowTodoBox"
              :todos="activeSessionTodos"
            />
          </template>

          <!-- Messages Slot -->
          <template #messages>
            <MessageBubble
              v-for="message in activeMessages"
              :key="message.id"
              :message="message"
              :format-time="formatTime"
              :format-message="formatMessage"
              @open-lightbox="openLightbox"
            >
              <!-- Edit Diff Slot (only when diffDisplayLocation is 'chat') -->
              <template #edit-diff>
                <EditDiffMessage
                  v-if="diffDisplayLocation === 'chat' && message.editToolData"
                  :file-path="message.editToolData.filePath"
                  :old-string="message.editToolData.oldString"
                  :new-string="message.editToolData.newString"
                  :replace-all="message.editToolData.replaceAll"
                  :status="message.editToolData.status"
                />
              </template>
            </MessageBubble>
          </template>

          <!-- Permissions Slot -->
          <template #permissions>
            <div v-if="activeSessionPermissions.length > 0" class="permission-requests">
              <PermissionRequest
                v-for="permission in activeSessionPermissions"
                :key="permission.request_id"
                :permission="permission"
                :connected="agentWs.connected"
                @approve="approvePermission"
                @deny="denyPermission"
              />
            </div>
          </template>

          <!-- Tool Execution Slot -->
          <template #tool-execution>
            <ToolExecutionBar :tool-execution="activeSessionToolExecution" />
          </template>
        </ChatArea>

        <!-- Session Metrics Sidebar -->
        <MetricsSidebar
          :show="!!activeSessionId"
          :session="activeSession"
          :message-count="activeMessages.length"
          :tool-executions="activeSessionToolExecutions"
          :permission-stats="activeSessionPermissionMetrics"
          :context-usage="activeSessionContextUsage"
          :context-loading="contextUsageLoading"
          @refresh-context="handleRefreshContext"
        />
      </main>
    </div>

    <!-- Create Session Modal -->
    <CreateSessionModal
      :show="showCreateSessionModal"
      :form-data="sessionForm"
      :providers="availableProviders"
      :current-provider="currentProvider"
      :agents="availableAgents"
      :selected-agent-preview="selectedAgentPreview"
      :loading-providers="loadingProviders"
      :loading-agents="loadingAgents"
      :creating="creatingSession"
      @close="showCreateSessionModal = false"
      @create="createSessionWithOptions"
      @working-directory-change="handleWorkingDirectoryChange"
      @agent-select="loadSelectedAgent"
    />

    <!-- Resume Session Modal -->
    <ResumeSessionModal
      :show="showResumeModal"
      :sessions="availableSessions"
      :selected-session="selectedResumeSession"
      :form-data="resumeForm"
      :loading="loadingSessions"
      :resuming="resumingSession"
      @close="showResumeModal = false; selectedResumeSession = null"
      @select-session="selectSessionForResume"
      @back="selectedResumeSession = null"
      @resume="resumeSessionWithOptions"
    />

    <!-- Image Lightbox -->
    <ImageLightbox
      :images="lightboxImages"
      :start-index="lightboxStartIndex"
      :is-open="showLightbox"
      @close="closeLightbox"
    />

  </div>
</template>

<script setup lang="ts">
import { useAgentWebSocket } from '~/composables/useAgentWebSocket'
import SessionMetrics from '~/components/SessionMetrics.vue'
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import type { ActiveTool } from '~/types/agents'

// Refactored Components
import SessionItem from '~/components/agents/SessionItem.vue'
import SessionFilters from '~/components/agents/SessionFilters.vue'
import PermissionRequest from '~/components/agents/PermissionRequest.vue'
import ToolExecutionBar from '~/components/agents/ToolExecutionBar.vue'
import CreateSessionModal from '~/components/agents/CreateSessionModal.vue'
import ResumeSessionModal from '~/components/agents/ResumeSessionModal.vue'
import SessionsSidebar from '~/components/agents/SessionsSidebar.vue'
import ChatArea from '~/components/agents/ChatArea.vue'
import MetricsSidebar from '~/components/agents/MetricsSidebar.vue'
import MessageBubble from '~/components/agents/MessageBubble.vue'
import TodoWriteBox from '~/components/agents/TodoWriteBox.vue'
import ToolOverlaysContainer from '~/components/agents/ToolOverlaysContainer.vue'
import ImageLightbox from '~/components/agents/ImageLightbox.vue'
import EditDiffMessage from '~/components/agents/EditDiffMessage.vue'

// Utilities
import { formatTime, formatMessage } from '~/utils/agents/messageFormatters'
import { type TodoItem } from '~/utils/agents/todoParser'
import { getToolIcon } from '~/utils/agents/toolParser'
import { useContextUsage } from '~/composables/agents/useContextUsage'

// Composables
import { useMessageScroll } from '~/composables/agents/useMessageScroll'
import { useSessionState } from '~/composables/agents/useSessionState'
import { useAgentProviders } from '~/composables/agents/useAgentProviders'
import { useDiffDisplaySetting } from '~/composables/useDiffDisplaySetting'
import { useSessionActions } from '~/composables/agents/useSessionActions'
import { useMessageHelpers } from '~/composables/agents/useMessageHelpers'
import { useToolManagement } from '~/composables/agents/useToolManagement'
import { useMessaging } from '~/composables/agents/useMessaging'
import { useWebSocketHandlers } from '~/composables/agents/useWebSocketHandlers'

// Existing overlays
import TodoWriteOverlay from '~/components/TodoWriteOverlay.vue'
import ToolOverlay from '~/components/ToolOverlay.vue'
import EditDiffOverlay from '~/components/agents/EditDiffOverlay.vue'

// WebSocket connection
const agentWs = useAgentWebSocket()

// Refs
const chatAreaRef = ref(null)
// Access messagesContainer and messageInput through chatAreaRef
const messagesContainer = computed(() => chatAreaRef.value?.messagesContainer || null)
const messageInput = computed(() => chatAreaRef.value?.messageInput || null)

// Composables - Session State Management
const {
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
  sessionContextUsage,
  filteredSessions,
  sessionFiltersWithCounts,
  activeSession,
  activeMessages,
  activeSessionPermissions,
  activeSessionTodos,
  activeSessionToolExecution,
  activeSessionTools,
  shouldShowTodoBox,
  activeSessionContextUsage
} = useSessionState()

// Context usage composable
const { parseContextResponse } = useContextUsage()
const contextUsageLoading = ref(false)
const contextUsageTimeoutId = ref<number | null>(null)

// Diff display setting
const { diffDisplayLocation } = useDiffDisplaySetting()

// Helper to find Edit tool for a message
const getEditToolForMessage = (messageId: string) => {
  const editTool = activeSessionTools.value.find(
    tool => tool.name === 'Edit' && tool.messageId === messageId
  )

  return editTool
}

// Composables - Provider & Agent Selection
const {
  sessionForm,
  resumeForm,
  availableAgents,
  selectedAgentPreview,
  loadingAgents,
  availableProviders,
  currentProvider,
  loadingProviders,
  getProviderModels
} = useAgentProviders()

// Auto-scroll composable
const { isUserNearBottom, handleScroll, scrollToBottom, autoScrollIfNearBottom } = useMessageScroll()

// Helper functions needed by session actions
const cleanupSessionData = (sessionId: string) => {
  sessionTodos.value.delete(sessionId)
  sessionToolExecution.value.delete(sessionId)
  activeTools.value.delete(sessionId)
}

const focusMessageInput = () => {
  nextTick(() => {
    if (messageInput.value && !messageInput.value.disabled) {
      messageInput.value.focus()
    }
  })
}

// Session actions composable
const {
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
} = useSessionActions({
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
})

// Message helpers composable
const {
  formatRelativeTime,
  parseTodoWrite,
  parseToolUse,
  formatTodosForTool,
  truncatePath,
  extractTextContent,
  isCompleteSignal,
  extractCostData,
  extractToolName
} = useMessageHelpers()

// Tool management composable
const {
  updateSessionTodos,
  updateSessionToolExecution,
  clearSessionToolExecution,
  addActiveTool,
  completeActiveTool,
  removeActiveTool
} = useToolManagement({
  sessionTodos,
  sessionToolExecution,
  activeTools
})

// Messaging composable
const {
  sendMessage,
  approvePermission,
  denyPermission,
  sendPermissionResponse,
  deleteAllSessions,
  killAllAgents
} = useMessaging({
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
})

// Handle send message with image attachments
const handleSendMessage = async () => {
  // Get attached images from ChatArea component
  const attachedImages = chatAreaRef.value?.attachedImages || []

  // Send message with images
  await sendMessage(attachedImages)

  // Clear image attachments in ChatArea after sending
  if (chatAreaRef.value?.clearAttachments) {
    chatAreaRef.value.clearAttachments()
  }
}

// Image lightbox state
const showLightbox = ref(false)
const lightboxImages = ref<any[]>([])
const lightboxStartIndex = ref(0)

// Open lightbox with images
const openLightbox = ({ images, startIndex }: { images: any[], startIndex: number }) => {
  lightboxImages.value = images
  lightboxStartIndex.value = startIndex
  showLightbox.value = true
}

// Close lightbox
const closeLightbox = () => {
  showLightbox.value = false
  // Clear images after animation completes
  setTimeout(() => {
    lightboxImages.value = []
    lightboxStartIndex.value = 0
  }, 300)
}

// Handle context usage refresh
const handleRefreshContext = async () => {
  if (!activeSessionId.value || contextUsageLoading.value) {
    return
  }

  // Clear any existing timeout
  if (contextUsageTimeoutId.value !== null) {
    clearTimeout(contextUsageTimeoutId.value)
  }

  contextUsageLoading.value = true

  // Set a 15-second timeout to reset loading state if no response
  contextUsageTimeoutId.value = setTimeout(() => {
    if (contextUsageLoading.value) {
      console.warn('Context usage request timed out after 15 seconds')
      contextUsageLoading.value = false
      contextUsageTimeoutId.value = null
    }
  }, 15000) as unknown as number

  try {
    // Set the input message to /context and send it
    inputMessage.value = '/context'
    await sendMessage([]) // Send with empty attachments array

    // The response will be handled by the WebSocket handlers
    // and will update sessionContextUsage through parseContextResponse
    // The timeout will be cleared in the WebSocket handler when response arrives
  } catch (error) {
    console.error('Failed to fetch context usage:', error)
    if (contextUsageTimeoutId.value !== null) {
      clearTimeout(contextUsageTimeoutId.value)
      contextUsageTimeoutId.value = null
    }
    contextUsageLoading.value = false
  }
}

// WebSocket handlers composable
const { setupHandlers } = useWebSocketHandlers({
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
})

// Initialize WebSocket handlers
setupHandlers()

// Computed properties for metrics to ensure reactivity
const activeSessionToolExecutions = computed(() => {
  if (!activeSessionId.value) return {}
  return sessionToolStats.value.get(activeSessionId.value) || {}
})

const activeSessionPermissionMetrics = computed(() => {
  if (!activeSessionId.value) return undefined
  return sessionPermissionStats.value.get(activeSessionId.value)
})

// Watch for modal opening to load sessions
watch(showResumeModal, (show) => {
  if (show) {
    loadAvailableSessions()
  }
})

// Watch for all todos completed and auto-hide after 5 seconds
watch(activeSessionTodos, (todos) => {
  if (!activeSessionId.value) return

  // Clear any existing timer for this session
  const existingTimer = todoHideTimers.value.get(activeSessionId.value)
  if (existingTimer) {
    clearTimeout(existingTimer)
    todoHideTimers.value.delete(activeSessionId.value)
  }

  // If all todos are completed, set a new timer
  if (todos.length > 0 && todos.every(todo => todo.status === 'completed')) {
    const timer = setTimeout(() => {
      const currentTodos = sessionTodos.value.get(activeSessionId.value)
      if (currentTodos && currentTodos.every(todo => todo.status === 'completed')) {
        sessionTodos.value.delete(activeSessionId.value)
        todoHideTimers.value.delete(activeSessionId.value)
      }
    }, 5000)
    todoHideTimers.value.set(activeSessionId.value, timer)
  }
}, { deep: true })

// Load existing sessions on mount
onMounted(() => {
  if (agentWs.connected) {
    agentWs.send({ type: 'list_sessions' })
  }
  // Load available providers
  loadProviders()

  // Register global action for keyboard shortcut
  const { setGlobalAction } = useKeyboardShortcuts()
  setGlobalAction('create-new-session', () => {
    createNewSession()
  })
})

// Cleanup on unmount
onUnmounted(() => {
  const { removeGlobalAction } = useKeyboardShortcuts()
  removeGlobalAction('create-new-session')
})

// Watch for connection changes
watch(() => agentWs.connected, (connected) => {
  if (connected) {
    agentWs.send({ type: 'list_sessions' })
  }
})

// Watch for new messages and auto-scroll if user is near bottom
watch(activeMessages, () => {
  autoScrollIfNearBottom(messagesContainer.value)
}, { deep: true, flush: 'post' })

// Auto-refresh context when a session is selected (both new and old sessions)
watch(activeSessionId, (newSessionId, oldSessionId) => {
  if (newSessionId && newSessionId !== oldSessionId) {
    // Wait a bit for messages to load, then fetch context
    setTimeout(() => {
      handleRefreshContext()
    }, 500)
  }
})
</script>

<style scoped>
/* Page-level layout styles only - component-specific styles are in their respective .vue files */

.agents-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  overflow: hidden;
}

.header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-bg);
  flex-shrink: 0;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-text {
  flex: 1;
}

.header h1 {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
}

.subtitle {
  margin: 4px 0 0 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.btn-delete-all,
.btn-kill-all {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-delete-all {
  background: #6c757d;
}

.btn-delete-all:hover:not(:disabled) {
  background: #5a6268;
}

.btn-kill-all:hover:not(:disabled) {
  background: #c82333;
}

.btn-delete-all:disabled,
.btn-kill-all:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.connection-status {
  position: absolute;
  top: 24px;
  right: 24px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: var(--bg-secondary);
  border-radius: 20px;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #dc3545;
}

.connection-status.connected .status-indicator {
  background: #28a745;
}

.agents-container {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 0;
}

.chat-area-with-metrics {
  flex: 1;
  display: flex;
  background: var(--bg-primary);
  overflow: hidden;
  min-height: 0;
  gap: 12px;
  padding: 12px;
}

.permission-requests {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  max-height: 200px;
  overflow-y: auto;
  flex-shrink: 0;
}
</style>
