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
      <aside class="sessions-sidebar">
        <div class="sidebar-header">
          <h3>Sessions</h3>
          <div class="session-buttons">
            <button @click="createNewSession" class="btn-new-session" :disabled="!agentWs.connected || creatingSession">
              <svg v-if="!creatingSession" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
              <div v-if="creatingSession" class="btn-spinner-small"></div>
              <span v-if="!creatingSession">New Session</span>
              <span v-else>Creating...</span>
            </button>
            <button @click="openResumeModal" class="btn-resume-session" :disabled="!agentWs.connected">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M3 15v4c0 1.1.9 2 2 2h14a2 2 0 0 0 2-2v-4M17 9l-5 5-5-5M12 12.8V2.5"/>
              </svg>
              Resume Session
            </button>
          </div>
        </div>

        <!-- Session Filter Tabs -->
        <div class="session-filters">
          <button
            v-for="filter in sessionFilters"
            :key="filter.value"
            @click="activeFilter = filter.value"
            class="filter-tab"
            :class="{ active: activeFilter === filter.value }"
          >
            {{ filter.label }}
            <span class="filter-count">{{ getFilterCount(filter.value) }}</span>
          </button>
        </div>

        <div class="sessions-list">
          <div v-if="filteredSessions.length === 0" class="no-sessions">
            No {{ activeFilter }} sessions
          </div>
          <div
            v-for="session in filteredSessions"
            :key="session.id"
            class="session-item"
            :class="{
              active: activeSessionId === session.id,
              ended: session.status === 'ended'
            }"
            @click="selectSession(session.id)"
          >
            <div class="session-status-dot" :class="session.status"></div>
            <div class="session-info">
              <div class="session-name">Session {{ session.id.slice(0, 8) }}</div>
              <div class="session-meta">
                <span class="session-status" :class="session.status">{{ session.status }}</span>
                <span class="session-messages">{{ session.message_count }} messages</span>
                <span v-if="session.cost_usd && session.cost_usd > 0" class="session-cost">${{ session.cost_usd.toFixed(4) }}</span>
              </div>
            </div>
            <div class="session-actions">
              <button
                v-if="session.status !== 'ended'"
                @click.stop="endSession(session.id)"
                class="btn-end-session"
                title="End session"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"></circle>
                  <line x1="15" y1="9" x2="9" y2="15"></line>
                  <line x1="9" y1="9" x2="15" y2="15"></line>
                </svg>
              </button>
              <button
                @click.stop="deleteSession(session.id)"
                class="btn-delete-session"
                title="Delete session"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="3 6 5 6 21 6"></polyline>
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                  <line x1="10" y1="11" x2="10" y2="17"></line>
                  <line x1="14" y1="11" x2="14" y2="17"></line>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </aside>

      <!-- Chat Area with Metrics -->
      <main class="chat-area-with-metrics">
        <div class="chat-main-area">
        <div v-if="!activeSessionId" class="no-session-selected">
          <div class="empty-state">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.5">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
            </svg>
            <p>Select a session or create a new one to start</p>
          </div>
        </div>

        <div v-else class="chat-content">
          <!-- Tool Overlays Container -->
          <div v-if="activeSessionTools.length > 0" class="tool-overlays-container">
            <template v-for="tool in activeSessionTools" :key="tool.id">
              <TodoWriteOverlay
                v-if="tool.name === 'TodoWrite'"
                :tool="tool"
                @dismiss="removeActiveTool(tool.sessionId, $event)"
              />
              <ToolOverlay
                v-else
                :tool="tool"
                @dismiss="removeActiveTool(tool.sessionId, $event)"
              />
            </template>
          </div>

          <!-- TodoWrite Box -->
          <div v-if="shouldShowTodoBox" class="todo-write-box">
            <div class="todo-box-header">
              <div class="todo-box-icon">📋</div>
              <div class="todo-box-title">Tasks</div>
            </div>
            <div class="todo-list">
              <div
                v-for="(todo, index) in activeSessionTodos"
                :key="index"
                class="todo-item"
                :class="todo.status"
              >
                <div class="todo-status-icon">
                  <span v-if="todo.status === 'completed'">✅</span>
                  <span v-else-if="todo.status === 'in_progress'" class="in-progress-icon">🔄</span>
                  <span v-else>📝</span>
                </div>
                <div class="todo-content">
                  <div class="todo-text">{{ todo.content }}</div>
                  <div v-if="todo.activeForm && todo.status === 'in_progress'" class="todo-active-form">
                    {{ todo.activeForm }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Messages -->
          <div class="messages-container" ref="messagesContainer" @scroll="handleScroll">
            <div v-for="message in activeMessages" :key="message.id" class="message" :class="{
              [message.role]: true,
              isToolResult: message.isToolResult,
              isExecutionStatus: message.isExecutionStatus,
              isPermissionDecision: message.isPermissionDecision,
              isHistorical: message.isHistorical,
              isError: message.isError
            }">
              <div class="message-header">
                <span class="message-role">{{ message.role === 'user' ? 'You' : message.role === 'system' ? 'System' : 'Claude' }}</span>
                <span class="message-time">{{ formatTime(message.timestamp) }}</span>
              </div>
              <div class="message-content" v-html="formatMessage(message.content)"></div>

              <!-- Tool use indicator -->
              <div v-if="message.toolUse" class="tool-use">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                </svg>
                Using {{ message.toolUse }}
              </div>
            </div>

            <!-- Thinking indicator -->
            <div v-if="isThinking" class="thinking-indicator">
              <div class="thinking-dots">
                <span></span>
                <span></span>
                <span></span>
              </div>
              Claude is thinking...
            </div>

            <!-- Processing indicator -->
            <div v-if="isProcessing && !isThinking" class="processing-indicator">
              <div class="processing-spinner"></div>
              Processing your message...
            </div>
          </div>

          <!-- Permission Requests -->
          <div v-if="activeSessionPermissions.length > 0" class="permission-requests">
            <div
              v-for="permission in activeSessionPermissions"
              :key="permission.request_id"
              class="permission-request"
            >
              <div class="permission-header">
                <div class="permission-icon">🔐</div>
                <div class="permission-title">Permission Request</div>
                <div class="permission-time">{{ formatTime(permission.timestamp) }}</div>
              </div>
              <div class="permission-description">
                {{ permission.description }}
              </div>
              <div class="permission-actions">
                <button
                  @click="denyPermission(permission)"
                  class="btn-deny"
                  title="Deny this request"
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                  Deny
                </button>
                <button
                  @click="approvePermission(permission)"
                  class="btn-approve"
                  title="Approve this request"
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20,6 9,17 4,12"></polyline>
                  </svg>
                  Approve
                </button>
              </div>
            </div>
          </div>

          <!-- Tool Execution Bar -->
          <div v-if="shouldShowToolBar" class="tool-execution-bar">
            <div class="tool-execution-content">
              <div class="tool-execution-icon">
                <span v-if="activeSessionToolExecution?.toolName === 'Bash'">⚡</span>
                <span v-else-if="activeSessionToolExecution?.toolName === 'Read'">📖</span>
                <span v-else-if="activeSessionToolExecution?.toolName === 'Write'">✏️</span>
                <span v-else-if="activeSessionToolExecution?.toolName === 'Edit'">🔧</span>
                <span v-else-if="activeSessionToolExecution?.toolName === 'Search' || activeSessionToolExecution?.toolName === 'Grep'">🔍</span>
                <span v-else>🛠️</span>
              </div>
              <div class="tool-execution-details">
                <div class="tool-execution-name">
                  {{ activeSessionToolExecution?.toolName }}
                  <span v-if="activeSessionToolExecution?.detail" class="tool-execution-detail-badge">
                    {{ truncatePath(activeSessionToolExecution.detail, 40) }}
                  </span>
                </div>
                <div class="tool-execution-info">
                  <span v-if="activeSessionToolExecution?.command">{{ truncatePath(activeSessionToolExecution.command, 60) }}</span>
                  <span v-else-if="activeSessionToolExecution?.filePath">{{ truncatePath(activeSessionToolExecution.filePath, 60) }}</span>
                  <span v-else-if="activeSessionToolExecution?.pattern">{{ truncatePath(activeSessionToolExecution.pattern, 60) }}</span>
                  <span v-else>Executing...</span>
                </div>
              </div>
              <div class="tool-execution-pulse"></div>
            </div>
          </div>

          <!-- Input Area -->
          <div class="input-area">
            <textarea
              ref="messageInput"
              v-model="inputMessage"
              @keydown.enter.prevent="sendMessage"
              placeholder="Type your message... (Enter to send)"
              class="message-input"
              :disabled="!agentWs.connected"
              rows="3"
            ></textarea>
            <button
              @click="sendMessage"
              class="btn-send"
              :disabled="!inputMessage.trim() || !agentWs.connected"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="22" y1="2" x2="11" y2="13"></line>
                <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
              </svg>
            </button>
          </div>
        </div>
        </div>

        <!-- Session Metrics Sidebar -->
        <aside class="metrics-sidebar" v-if="activeSessionId">
          <SessionMetrics
            :session="activeSession"
            :tool-executions="sessionToolStats.get(activeSessionId)"
            :permission-stats="sessionPermissionStats.get(activeSessionId)"
          />
        </aside>
      </main>
    </div>

    <!-- Create Session Modal -->
    <div v-if="showCreateSessionModal" class="modal-overlay" @click="showCreateSessionModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h2>Create New Session</h2>
          <button @click="showCreateSessionModal = false" class="modal-close">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label for="working-directory">Working Directory</label>
            <input
              id="working-directory"
              v-model="sessionForm.workingDirectory"
              @change="handleWorkingDirectoryChange"
              @blur="handleWorkingDirectoryChange"
              type="text"
              placeholder="/home/user/projects"
              class="form-input"
            />
            <small class="form-help">The directory where the agent will work</small>
          </div>

          <div class="form-group">
            <label for="permission-mode">Permission Mode</label>
            <select id="permission-mode" v-model="sessionForm.permissionMode" class="form-select">
              <option value="default">Default (ask for permissions)</option>
              <option value="acceptEdits">Allow All (full permissions)</option>
              <option value="plan">Read Only (no file modifications)</option>
            </select>
            <small class="form-help">Control what permissions the agent has</small>
          </div>

          <div class="form-group">
            <label for="prompt-mode">System Prompt</label>
            <div class="prompt-mode-toggle">
              <button
                type="button"
                class="mode-btn"
                :class="{ active: sessionForm.promptMode === 'agent' }"
                @click="sessionForm.promptMode = 'agent'"
              >
                📦 Project Agent
              </button>
              <button
                type="button"
                class="mode-btn"
                :class="{ active: sessionForm.promptMode === 'custom' }"
                @click="sessionForm.promptMode = 'custom'"
              >
                ✏️ Custom
              </button>
            </div>
            <small class="form-help">Choose a project agent or write a custom system prompt</small>
          </div>

          <!-- Agent Selection Mode -->
          <div v-if="sessionForm.promptMode === 'agent'" class="form-group">
            <label for="agent-select">Select Agent</label>
            <div v-if="loadingAgents" class="agents-loading">
              <div class="loading-spinner-small"></div>
              <span>Loading agents...</span>
            </div>
            <div v-else-if="availableAgents.length === 0" class="agents-empty">
              No agents found. Make sure you've set a valid working directory.
            </div>
            <div v-else class="agents-grid">
              <button
                v-for="agent in availableAgents"
                :key="agent.name"
                type="button"
                class="agent-card"
                :class="{ selected: sessionForm.selectedAgent === agent.name }"
                @click="sessionForm.selectedAgent = agent.name; loadSelectedAgent()"
              >
                <div class="agent-card-color" :style="{ backgroundColor: agent.color || '#8B5CF6' }"></div>
                <div class="agent-card-content">
                  <div class="agent-card-name">{{ agent.name }}</div>
                  <div v-if="agent.model" class="agent-card-model">{{ agent.model }}</div>
                </div>
                <div v-if="sessionForm.selectedAgent === agent.name" class="agent-card-checkmark">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                    <polyline points="20 6 9 17 4 12"></polyline>
                  </svg>
                </div>
              </button>
            </div>
            <small class="form-help">Select from available agents</small>

            <!-- Agent Preview -->
            <div v-if="sessionForm.selectedAgent && selectedAgentPreview" class="agent-preview">
              <div class="agent-preview-header">
                <strong>{{ selectedAgentPreview.name }}</strong>
                <span v-if="selectedAgentPreview.model" class="agent-model">{{ selectedAgentPreview.model }}</span>
              </div>
              <p v-if="selectedAgentPreview.description" class="agent-description">{{ selectedAgentPreview.description }}</p>
              <div class="agent-prompt-preview">
                <p class="preview-label">System Prompt Preview:</p>
                <div class="prompt-content">{{ selectedAgentPreview.system_prompt.substring(0, 300) }}...</div>
              </div>
            </div>
          </div>

          <!-- Custom Prompt Mode -->
          <div v-if="sessionForm.promptMode === 'custom'" class="form-group">
            <label for="system-prompt">Custom System Prompt</label>
            <textarea
              id="system-prompt"
              v-model="sessionForm.systemPrompt"
              placeholder="You are a helpful AI assistant."
              class="form-textarea"
              rows="4"
            ></textarea>
            <small class="form-help">Enter custom instructions for the agent</small>
          </div>

          <div class="form-group">
            <label>Available Tools</label>
            <div class="tools-grid">
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Read" />
                <span>Read</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Write" />
                <span>Write</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Edit" />
                <span>Edit</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Bash" />
                <span>Bash</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Search" />
                <span>Search</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="Grep" />
                <span>Grep</span>
              </label>
              <label class="tool-checkbox">
                <input type="checkbox" v-model="sessionForm.tools" value="TodoWrite" />
                <span>TodoWrite</span>
              </label>
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button @click="showCreateSessionModal = false" class="btn-cancel" :disabled="creatingSession">Cancel</button>
          <button @click="createSessionWithOptions" class="btn-create" :disabled="!sessionForm.workingDirectory || creatingSession">
            <div v-if="creatingSession" class="btn-spinner"></div>
            <span v-if="!creatingSession">Create Session</span>
            <span v-else>Creating...</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Resume Session Modal -->
    <div v-if="showResumeModal" class="modal-overlay" @click="showResumeModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h2>Resume Session</h2>
          <button @click="showResumeModal = false" class="modal-close">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div v-if="loadingSessions" class="loading-sessions">
            <div class="loading-spinner"></div>
            Loading available sessions...
          </div>
          <div v-else-if="availableSessions.length === 0" class="no-sessions-available">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.5">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
            </svg>
            <p>No previous sessions found</p>
          </div>
          <div v-else-if="!selectedResumeSession" class="sessions-list-modal">
            <div
              v-for="session in availableSessions"
              :key="session.conversation_id"
              class="session-card-modal"
              @click="selectSessionForResume(session)"
            >
              <div class="session-card-avatar">
                <img
                  :src="`/avatars/${session.session_name || 'default'}.png`"
                  :alt="session.session_name"
                  @error="$event.target.src = '/avatars/default.png'"
                  class="avatar-image"
                />
              </div>
              <div class="session-card-info">
                <div class="session-card-name">{{ session.session_name || 'Unnamed Session' }}</div>
                <div class="session-card-directory">📁 {{ session.working_directory || 'No directory' }}</div>
                <div class="session-card-meta">
                  <span class="session-card-messages">💬 {{ session.total_messages }} messages</span>
                  <span class="session-card-time">⏱️ {{ formatRelativeTime(session.last_activity) }}</span>
                </div>
              </div>
              <div class="session-card-arrow">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M5 12h14M12 5l7 7-7 7"/>
                </svg>
              </div>
            </div>
          </div>

          <!-- Resume Session Options -->
          <div v-else class="resume-session-options">
            <div class="selected-session-info">
              <h3>{{ selectedResumeSession.session_name || 'Selected Session' }}</h3>
              <p>Original working directory: <code>{{ selectedResumeSession.working_directory }}</code></p>
            </div>

            <div class="form-group">
              <label for="resume-working-directory">Working Directory</label>
              <input
                id="resume-working-directory"
                v-model="resumeForm.workingDirectory"
                type="text"
                :placeholder="selectedResumeSession.working_directory"
                class="form-input"
              />
              <small class="form-help">Directory where the agent will work (defaults to original)</small>
            </div>

            <div class="form-group">
              <label for="resume-permission-mode">Permission Mode</label>
              <select id="resume-permission-mode" v-model="resumeForm.permissionMode" class="form-select">
                <option value="default">Default (ask for permissions)</option>
                <option value="acceptEdits">Allow All (full permissions)</option>
                <option value="plan">Read Only (no file modifications)</option>
              </select>
              <small class="form-help">Control what permissions the agent has</small>
            </div>

            <div class="form-group">
              <label for="resume-system-prompt">System Prompt (optional)</label>
              <textarea
                id="resume-system-prompt"
                v-model="resumeForm.systemPrompt"
                placeholder="You are a helpful AI assistant."
                class="form-textarea"
                rows="3"
              ></textarea>
              <small class="form-help">Custom instructions for the agent</small>
            </div>

            <div class="form-group">
              <label>Available Tools</label>
              <div class="tools-grid">
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Read" />
                  <span>Read</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Write" />
                  <span>Write</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Edit" />
                  <span>Edit</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Bash" />
                  <span>Bash</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Search" />
                  <span>Search</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="Grep" />
                  <span>Grep</span>
                </label>
                <label class="tool-checkbox">
                  <input type="checkbox" v-model="resumeForm.tools" value="TodoWrite" />
                  <span>TodoWrite</span>
                </label>
              </div>
            </div>

            <div class="modal-actions">
              <button @click="selectedResumeSession = null" class="btn-cancel" :disabled="resumingSession">Back</button>
              <button @click="resumeSessionWithOptions" class="btn-create" :disabled="resumingSession">
                <div v-if="resumingSession" class="btn-spinner"></div>
                <span v-if="!resumingSession">Resume Session</span>
                <span v-else>Resuming...</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAgentWebSocket } from '~/composables/useAgentWebSocket'
import SessionMetrics from '~/components/SessionMetrics.vue'
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import type { ActiveTool } from '~/types/agents'

interface TodoItem {
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  activeForm?: string
}

interface ToolExecution {
  toolName: string
  filePath?: string
  command?: string
  pattern?: string
  timestamp: Date
}

// WebSocket connection
const agentWs = useAgentWebSocket()

// Refs
const messageInput = ref(null)
const messagesContainer = ref<HTMLElement | null>(null)

// State
const sessions = ref([])
const activeSessionId = ref(null)
const messages = ref({}) // { sessionId: [...messages] }
const messagesLoaded = ref(new Set()) // Track which sessions have loaded messages from DB
const inputMessage = ref('')
const isProcessing = ref(false)
const isThinking = ref(false)
const showResumeModal = ref(false)
const availableSessions = ref([])
const loadingSessions = ref(false)
const showCreateSessionModal = ref(false)
const selectedResumeSession = ref(null)
const creatingSession = ref(false)
const resumingSession = ref(false)
const sessionPermissions = ref(new Map<string, any[]>()) // { sessionId: [...permissions] }
const awaitingToolResults = ref(new Set()) // Track sessions awaiting tool execution results

// Live agents state
const sessionTodos = ref(new Map<string, TodoItem[]>()) // { sessionId: [...todos] }
const sessionToolExecution = ref(new Map<string, ToolExecution | null>()) // { sessionId: toolExecution }
const todoHideTimers = ref(new Map<string, NodeJS.Timeout>()) // { sessionId: timeoutId }

// Tool overlays state
const activeTools = ref(new Map<string, ActiveTool[]>()) // { sessionId: [...activeTools] }

// Auto-scroll state
const isUserNearBottom = ref(true) // Track if user is near bottom of messages

// Session filtering state
const activeFilter = ref('active') // 'all', 'active', 'ended'
const sessionFilters = [
  { label: 'Active', value: 'active' },
  { label: 'All', value: 'all' },
  { label: 'Ended', value: 'ended' }
]

// Computed: Filtered sessions based on active filter
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

// Session metrics state
const sessionToolStats = ref(new Map<string, Record<string, number>>()) // { sessionId: { toolName: count } }
const sessionPermissionStats = ref(new Map<string, { approved: number; denied: number; total: number }>()) // { sessionId: stats }

// Session creation form
const sessionForm = ref({
  workingDirectory: '',
  permissionMode: 'default',
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
const availableAgents = ref([])
const selectedAgentPreview = ref(null)
const loadingAgents = ref(false)

// Computed
const activeMessages = computed(() => {
  if (!activeSessionId.value) return []
  return messages.value[activeSessionId.value] || []
})

const activeSessionTodos = computed(() => {
  if (!activeSessionId.value) return []
  return sessionTodos.value.get(activeSessionId.value) || []
})

const activeSessionTools = computed(() => {
  if (!activeSessionId.value) return []
  return activeTools.value.get(activeSessionId.value) || []
})

const activeSessionPermissions = computed(() => {
  if (!activeSessionId.value) return []
  return sessionPermissions.value.get(activeSessionId.value) || []
})

const activeSessionToolExecution = computed(() => {
  if (!activeSessionId.value) return null
  return sessionToolExecution.value.get(activeSessionId.value) || null
})

const shouldShowTodoBox = computed(() => {
  return activeSessionTodos.value.length > 0
})

const shouldShowToolBar = computed(() => {
  return activeSessionToolExecution.value !== null
})

const activeSession = computed(() => {
  if (!activeSessionId.value) return null
  return sessions.value.find(s => s.id === activeSessionId.value) || null
})

// Message formatting
const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  })
}

const formatMessage = (content) => {
  // If content is an object, extract text from it first
  if (typeof content === 'object' && content !== null) {
    // Try to extract meaningful text from the object
    if (content.text) {
      content = Array.isArray(content.text) ? content.text.join('\n') : String(content.text)
    } else if (content.content) {
      content = String(content.content)
    } else {
      // For other objects, create a concise representation
      const objType = content.type || 'unknown'
      const keys = Object.keys(content).filter(k => k !== 'type')
      if (keys.length === 0) {
        return `<em class="system-message">${objType}</em>`
      }
      // Show key properties in a readable format
      const props = keys.slice(0, 3).map(k => `${k}: ${String(content[k]).substring(0, 30)}`).join(', ')
      return `<em class="system-message">${objType} - ${props}</em>`
    }
  }

  // Ensure content is a string at this point
  content = String(content)

  // Skip system messages and JSON-like content
  if (content.includes('SystemMessage(') || (content.startsWith('{') && content.includes('"type"'))) {
    return '<em class="system-message">Processing...</em>'
  }

  // Clean up the content
  let cleanContent = content

  // If it's a string representation of an object, try to extract meaningful text
  if (cleanContent.includes('assistant:')) {
    const match = cleanContent.match(/assistant:\s*(.+?)(?:\n|$)/i)
    if (match) {
      cleanContent = match[1]
    }
  }

  // Convert markdown to HTML (basic)
  return cleanContent
    .replace(/```(.*?)\n([\s\S]*?)```/g, '<pre><code class="language-$1">$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/\n/g, '<br>')
}

// Session management
const createNewSession = async () => {
  if (!agentWs.connected) return

  // Reset form to defaults
  sessionForm.value = {
    workingDirectory: '',
    permissionMode: 'default',
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

const createSessionWithOptions = async () => {
  if (!agentWs.connected || !sessionForm.value.workingDirectory) return

  creatingSession.value = true

  try {
    const sessionId = crypto.randomUUID()

    const options: any = {
      tools: sessionForm.value.tools,
      working_directory: sessionForm.value.workingDirectory,
      permission_mode: sessionForm.value.permissionMode
    }

    // Use agent_name if agent mode is selected, otherwise use system_prompt
    if (sessionForm.value.promptMode === 'agent' && sessionForm.value.selectedAgent) {
      options.agent_name = sessionForm.value.selectedAgent
    } else {
      options.system_prompt = sessionForm.value.systemPrompt || 'You are a helpful AI assistant.'
    }

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

// Reload agents (agents auto-load from cwd)
const handleWorkingDirectoryChange = async () => {
  if (sessionForm.value.workingDirectory) {
    await loadAvailableAgents()
    // Clear selected agent when working directory changes
    sessionForm.value.selectedAgent = ''
    selectedAgentPreview.value = null
  }
}

// Auto-scroll helpers
const handleScroll = () => {
  if (!messagesContainer.value) return

  const { scrollTop, scrollHeight, clientHeight } = messagesContainer.value
  const threshold = 100 // pixels from bottom
  isUserNearBottom.value = scrollHeight - scrollTop - clientHeight < threshold
}

const scrollToBottom = (smooth = false) => {
  if (!messagesContainer.value) return

  nextTick(() => {
    messagesContainer.value?.scrollTo({
      top: messagesContainer.value.scrollHeight,
      behavior: smooth ? 'smooth' : 'auto'
    })
  })
}

const autoScrollIfNearBottom = (smooth = true) => {
  if (isUserNearBottom.value) {
    scrollToBottom(smooth)
  }
}

const loadSelectedAgent = async () => {
  if (!sessionForm.value.selectedAgent) {
    selectedAgentPreview.value = null
    return
  }

  try {
    const response = await fetch(`/api/agents/${sessionForm.value.selectedAgent}`)
    if (!response.ok) throw new Error('Failed to fetch agent')
    const data = await response.json()
    selectedAgentPreview.value = data.agent
    console.log(`Loaded agent: ${sessionForm.value.selectedAgent}`, data.agent)
  } catch (error) {
    console.error('Error loading agent:', error)
    selectedAgentPreview.value = null
  }
}

const selectSession = (sessionId) => {
  activeSessionId.value = sessionId

  // Load historical messages if not already loaded
  // Only load if we haven't loaded this session before (prevents loading for new sessions)
  if (!messagesLoaded.value.has(sessionId)) {
    console.log(`Loading messages for session ${sessionId}`)
    agentWs.send({
      type: 'load_messages',
      session_id: sessionId,
      limit: 100,
      offset: 0
    })
    // Mark as loaded (even if empty) to prevent re-loading
    messagesLoaded.value.add(sessionId)
  }

  // Reset scroll state and scroll to bottom when switching sessions
  isUserNearBottom.value = true
  scrollToBottom(false)

  // Focus the input when switching to a session
  focusMessageInput()
}

const endSession = async (sessionId) => {
  if (!agentWs.connected) return

  agentWs.send({
    type: 'end_session',
    session_id: sessionId
  })

  // Remove from local state
  sessions.value = sessions.value.filter(s => s.id !== sessionId)
  delete messages.value[sessionId]
  messagesLoaded.value.delete(sessionId)  // Clean up loaded tracking
  awaitingToolResults.value.delete(sessionId)  // Clean up flag

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

const deleteSession = async (sessionId) => {
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

// Resume session functionality
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

const openResumeModal = () => {
  showResumeModal.value = true
}

const selectSessionForResume = async (session) => {
  selectedResumeSession.value = session

  // Prefill the form with the session's data
  resumeForm.value = {
    workingDirectory: session.working_directory || '',
    permissionMode: 'default',
    systemPrompt: '',
    tools: ['Read', 'Write', 'Edit', 'Bash', 'Search', 'TodoWrite']
  }
}

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
      resumeData.messages.forEach(msg => {
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

const formatRelativeTime = (timestamp) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

// Helper methods for parsing TodoWrite and tool execution data
const parseTodoWrite = (content: string): TodoItem[] | null => {
  if (!content || typeof content !== 'string') return null

  try {
    console.log('Parsing TodoWrite from content:', content)

    // Pattern 1: Look for numbered lists (1. task, 2. task, etc.)
    const numberedListMatch = content.match(/(?:\d+\.\s+)([^\n]+)/g)
    if (numberedListMatch) {
      console.log('Found numbered list matches:', numberedListMatch)
      const todos: TodoItem[] = []
      for (let i = 0; i < numberedListMatch.length; i++) {
        const taskContent = numberedListMatch[i].replace(/^\d+\.\s+/, '').trim()
        if (taskContent) {
          todos.push({
            content: taskContent,
            status: i === 0 ? 'in_progress' : 'pending'
          })
        }
      }
      if (todos.length > 0) {
        console.log('Successfully parsed todos from numbered list:', todos)
        return todos
      }
    }

    // Pattern 2: Look for checkmark-style lists (- task, * task, etc.)
    const bulletListMatch = content.match(/[-*]\s+([^\n]+)/g)
    if (bulletListMatch) {
      console.log('Found bullet list matches:', bulletListMatch)
      const todos: TodoItem[] = []
      for (const match of bulletListMatch) {
        const taskContent = match.replace(/^[-*]\s+/, '').trim()
        if (taskContent) {
          todos.push({
            content: taskContent,
            status: 'pending'
          })
        }
      }
      if (todos.length > 0) {
        console.log('Successfully parsed todos from bullet list:', todos)
        return todos
      }
    }

    // Pattern 3: Look for explicit todo markers ("Todo:", "Task:", etc.)
    const todoMarkerMatch = content.match(/(?:todo|task|items?):\s*([\s\S]*?)(?=\n\n|\n\w+:|$)/i)
    if (todoMarkerMatch) {
      const todoText = todoMarkerMatch[1].trim()
      const items = todoText.split(/\n\s*\n/).filter(item => item.trim())
      if (items.length > 0) {
        const todos = items.map(item => ({
          content: item.trim(),
          status: 'pending'
        }))
        console.log('Successfully parsed todos from todo marker:', todos)
        return todos
      }
    }

    // Pattern 4: Look for task-related patterns (common task descriptions)
    const taskPatterns = [
      /(?:i'll create|let me create|creating|here are)\s+a\s+(?:todo|task|list):\s*([\s\S]*?)(?=\n\n|\n|$)/i,
      /(?:tasks?:\s*\n)((?:\d+\.\s+[^\n]+\n)+)/i,
      /(?:items?:\s*\n)((?:[-*]\s+[^\n]+\n)+)/i
    ]

    for (const pattern of taskPatterns) {
      const match = content.match(pattern)
      if (match) {
        const taskContent = match[1] || match[0]
        const lines = taskContent.split('\n').filter(line => line.trim())
        const todos = lines.map(line => ({
          content: line.trim().replace(/^\d+\.\s+/, '').replace(/^[-*]\s+/, ''),
          status: 'pending'
        })).filter(todo => todo.content)

        if (todos.length > 0) {
          console.log('Successfully parsed todos from task pattern:', todos)
          return todos
        }
      }
    }

    console.log('No TodoWrite data found in content')
    return null
  } catch (e) {
    console.warn('Failed to parse TodoWrite content:', e)
    return null
  }
}

const parseToolUse = (content: string): ToolExecution | null => {
  if (!content || typeof content !== 'string') return null

  try {
    // Look for tool use patterns
    const patterns = [
      /Using (\w+)/g,
      /(\w+)\s*\(/g, // Function calls
      /Running (\w+)/g,
      /Executing (\w+)/g
    ]

    for (const pattern of patterns) {
      const matches = [...content.matchAll(pattern)]
      if (matches.length > 0) {
        const toolName = matches[0][1]

        // Extract additional details based on tool type
        let filePath, command, patternStr

        if (toolName === 'Read' || toolName === 'Write' || toolName === 'Edit') {
          const fileMatch = content.match(/(?:file|path):\s*([^\s\n]+)/i)
          if (fileMatch) filePath = fileMatch[1]
        } else if (toolName === 'Bash') {
          const commandMatch = content.match(/(?:command|cmd):\s*([^\n]+)/i)
          if (commandMatch) command = commandMatch[1].trim()
        } else if (toolName === 'Search' || toolName === 'Grep') {
          const patternMatch = content.match(/(?:pattern|search):\s*([^\n]+)/i)
          if (patternMatch) patternStr = patternMatch[1].trim()
        }

        return {
          toolName,
          filePath,
          command,
          pattern: patternStr,
          timestamp: new Date()
        }
      }
    }

    return null
  } catch (e) {
    console.warn('Failed to parse tool use:', e)
    return null
  }
}

// Update session data methods
const updateSessionTodos = (sessionId: string, todos: TodoItem[]) => {
  sessionTodos.value.set(sessionId, todos)
}

// Helper to format todos for TodoWrite tool (includes activeForm only when present)
const formatTodosForTool = (todos: TodoItem[]): any[] => {
  return todos.map(todo => ({
    content: todo.content,
    status: todo.status,
    ...(todo.activeForm && { activeForm: todo.activeForm })
  }))
}

const updateSessionToolExecution = (sessionId: string, toolExecution: ToolExecution | null) => {
  sessionToolExecution.value.set(sessionId, toolExecution)
}

const clearSessionToolExecution = (sessionId: string) => {
  sessionToolExecution.value.delete(sessionId)
}

// Tool overlay management
const addActiveTool = (sessionId: string, toolUse: any) => {
  const tools = activeTools.value.get(sessionId) || []
  const activeTool: ActiveTool = {
    id: toolUse.id,
    name: toolUse.name,
    input: toolUse.input,
    status: 'running',
    startTime: Date.now(),
    sessionId
  }
  tools.push(activeTool)
  activeTools.value.set(sessionId, tools)
}

const completeActiveTool = (sessionId: string, toolUseId: string, isError: boolean = false) => {
  const tools = activeTools.value.get(sessionId) || []
  const tool = tools.find(t => t.id === toolUseId)
  if (tool) {
    tool.status = isError ? 'error' : 'completed'
    tool.endTime = Date.now()
    activeTools.value.set(sessionId, [...tools])
  }
}

const removeActiveTool = (sessionId: string, toolId: string) => {
  const tools = activeTools.value.get(sessionId) || []
  const filtered = tools.filter(t => t.id !== toolId)
  activeTools.value.set(sessionId, filtered)
}

const cleanupSessionData = (sessionId: string) => {
  sessionTodos.value.delete(sessionId)
  sessionToolExecution.value.delete(sessionId)
  activeTools.value.delete(sessionId)
}

const truncatePath = (path: string): string => {
  if (!path) return ''
  if (path.length <= 50) return path

  // Truncate from the middle, keeping the beginning and end
  const start = path.substring(0, 25)
  const end = path.substring(path.length - 20)
  return `${start}...${end}`
}

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
    console.log('All todos completed, setting 5 second auto-hide timer')
    const timer = setTimeout(() => {
      const currentTodos = sessionTodos.value.get(activeSessionId.value)
      if (currentTodos && currentTodos.every(todo => todo.status === 'completed')) {
        sessionTodos.value.delete(activeSessionId.value)
        todoHideTimers.value.delete(activeSessionId.value)
        console.log('Auto-hid todos after 5 seconds')
      }
    }, 5000)
    todoHideTimers.value.set(activeSessionId.value, timer)
  }
}, { deep: true })

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

// Helper function to focus message input
const focusMessageInput = () => {
  nextTick(() => {
    if (messageInput.value && !messageInput.value.disabled) {
      messageInput.value.focus()
    }
  })
}

// Permission request functionality
const approvePermission = (request) => {
  sendPermissionResponse(request, true)
}

const denyPermission = (request, reason = '') => {
  sendPermissionResponse(request, false, reason)
}

const sendPermissionResponse = (request, approved, reason = '') => {
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
      autoScrollIfNearBottom()
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

// Helper function to extract text content from nested content object
const extractTextContent = (content: any): string => {
  if (!content) return ''

  // If content is already a string, return it
  if (typeof content === 'string') return content

  // If content is an object with nested structure
  if (typeof content === 'object') {
    // Handle assistant messages with text array
    if (content.type === 'assistant') {
      // Check if text array exists and is not null/empty
      if (Array.isArray(content.text) && content.text.length > 0) {
        return content.text.join('\n')
      }
      // Empty or null text array - no content to display
      return ''
    }

    // Handle user messages
    if (content.type === 'user' && content.content) {
      return String(content.content)
    }

    // Handle result messages (completion signal - no visible content)
    if (content.type === 'result') {
      return ''
    }

    // Handle system messages
    if (content.type === 'system') {
      return `SystemMessage: ${content.subtype || 'unknown'}`
    }

    // Fallback: stringify the object
    return JSON.stringify(content)
  }

  return String(content)
}

// Helper to check if content is complete signal
const isCompleteSignal = (content: any): boolean => {
  return typeof content === 'object' && content.type === 'result'
}

// Helper to extract cost/usage data from result messages
const extractCostData = (content: any) => {
  if (typeof content === 'object' && content.type === 'result') {
    return {
      costUSD: content.cost_usd || 0,
      numTurns: content.num_turns || 0,
      durationMs: content.duration_ms || 0,
      usage: content.usage || null
    }
  }
  return null
}

// Helper to extract tool name from tool_uses JSON
const extractToolName = (toolUses: any): string | undefined => {
  if (!toolUses) return undefined

  try {
    // Parse if it's a JSON string
    const parsed = typeof toolUses === 'string' ? JSON.parse(toolUses) : toolUses

    // If it's an array, get the first tool
    if (Array.isArray(parsed) && parsed.length > 0) {
      return parsed[0].name || parsed[0].type
    }

    // If it's a single object
    if (parsed.name || parsed.type) {
      return parsed.name || parsed.type
    }
  } catch (e) {
    console.warn('Failed to parse tool_uses:', e)
  }

  return undefined
}

// WebSocket event handlers
agentWs.on('onSessionCreated', (data) => {
  sessions.value.push(data.session)
  activeSessionId.value = data.session_id
  messages.value[data.session_id] = []

  // Mark new session as loaded (it has no history to load)
  messagesLoaded.value.add(data.session_id)

  // Focus the input after session creation
  focusMessageInput()
})

agentWs.on('onAgentMessage', (data) => {
  console.log('📨 Received agent message:', data)

  if (!messages.value[data.session_id]) {
    messages.value[data.session_id] = []
  }

  // Check if this is a completion signal (result message)
  const isComplete = isCompleteSignal(data.content)

  // Extract cost data from result messages
  const costData = extractCostData(data.content)

  // Extract text content from nested object
  const textContent = extractTextContent(data.content)

  console.log('💬 Extracted:', { isComplete, costData, textContent: textContent.substring(0, 50) })

  // Process tool uses (when Claude starts using a tool)
  if (data.content && data.content.tools && Array.isArray(data.content.tools)) {
    data.content.tools.forEach((toolUse: any) => {
      console.log('🔧 Tool use detected:', toolUse.name)
      addActiveTool(data.session_id, toolUse)
    })
  }

  // Process tool results (when tool execution completes)
  if (data.content && data.content.tool_results && Array.isArray(data.content.tool_results)) {
    data.content.tool_results.forEach((toolResult: any) => {
      console.log('✅ Tool result received:', toolResult.tool_use_id)
      completeActiveTool(data.session_id, toolResult.tool_use_id, toolResult.is_error || false)
    })
  }

  // Update session status and metadata
  const session = sessions.value.find(s => s.id === data.session_id)
  if (session) {
    // Update git branch from metadata
    if (data.metadata && data.metadata.git_branch) {
      session.git_branch = data.metadata.git_branch
      console.log('🌿 Updated git branch:', session.git_branch)
    }

    // Update costs from result message
    if (costData) {
      session.cost_usd = (session.cost_usd || 0) + costData.costUSD
      session.num_turns = costData.numTurns
      session.duration_ms = costData.durationMs
      session.usage = costData.usage
      console.log('💰 Updated session cost:', session.cost_usd)
    }

    // Set status: idle when complete, processing when receiving content
    if (isComplete) {
      session.status = 'idle'
      session.message_count = (session.message_count || 0) + 1
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
    console.log('✅ Message complete (result received)')
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
      console.log('🔧 Created tool result message:', toolMessage.content)
    }

    return
  }

  // Skip empty content and system messages (they don't need UI display)
  if (!textContent || textContent.includes('SystemMessage')) {
    console.log('⏭️  Skipping empty/system message')
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
    id: `msg-${data.session_id}-${Date.now()}`,
    role: 'assistant',
    content: textContent,
    timestamp: new Date(),
    streaming: false,
    isToolResult: isToolResult
  }

  messages.value[data.session_id].push(newMessage)
  console.log('✨ Created new message:', newMessage.id)

  // Reset processing state when we receive content
  isProcessing.value = false
  isThinking.value = false

  // Auto-scroll to bottom if user is near bottom
  autoScrollIfNearBottom()
})

agentWs.on('onAgentThinking', (data) => {
  if (data.session_id === activeSessionId.value) {
    isThinking.value = data.thinking
    // When thinking stops, ensure processing is also reset
    if (!data.thinking) {
      isProcessing.value = false
    }
  }

  // Update session status based on thinking state
  const session = sessions.value.find(s => s.id === data.session_id)
  if (session) {
    session.status = data.thinking ? 'processing' : 'idle'
  }
})

agentWs.on('onAgentToolUse', (data) => {
  // Update session status to processing when tool is being used
  const session = sessions.value.find(s => s.id === data.session_id)
  if (session) {
    session.status = 'processing'
  }

  // Track tool usage for metrics
  const toolStats = sessionToolStats.value.get(data.session_id) || {}
  toolStats[data.tool] = (toolStats[data.tool] || 0) + 1
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

  // Handle TodoWrite specifically
  if (data.tool && data.tool.includes('TodoWrite')) {
    console.log('TodoWrite tool used with data:', data)

    // Try to extract todos from the data.input property (for new format)
    let todos: TodoItem[] | null = null

    // If data has input with todos (new enhanced format), use that
    if (data.input && typeof data.input === 'object' && data.input.todos) {
      todos = data.input.todos
      console.log('Extracted todos from data.input:', todos)
    } else {
      // Try legacy parsing from tool string representation
      const toolStr = String(data.tool || '')
      todos = parseTodoWrite(toolStr)
      console.log('Parsed todos from legacy tool string:', todos)
    }

    if (todos && Array.isArray(todos)) {
      console.log('Updating session', data.session_id, 'with todos:', todos)
      updateSessionTodos(data.session_id, todos)

      // Set up auto-hide timer if all todos are completed
      const allCompleted = todos.every(todo => todo.status === 'completed')
      if (allCompleted) {
        console.log('All todos completed, setting auto-hide timer for 5 seconds')
        setTimeout(() => {
          // Clear todos after delay, only if all are still completed
          const currentTodos = sessionTodos.value.get(data.session_id)
          if (currentTodos && currentTodos.every(todo => todo.status === 'completed')) {
            sessionTodos.value.delete(data.session_id)
            console.log('Auto-hiding todos for session', data.session_id)
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
    autoScrollIfNearBottom()
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
  sessions.value = data.sessions
})

agentWs.on('onSessionDeleted', (data) => {
  console.log('🗑️ Session deleted:', data.session_id)
  // Session already removed from local state in deleteSession (optimistic update)
  // Just log confirmation
})

agentWs.on('onAllSessionsDeleted', (data) => {
  console.log('🗑️ All sessions deleted, count:', data.count)

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
  console.log('📥 Messages loaded:', data)

  if (!data.session_id || !data.messages) return

  // Debug: log sequence numbers
  console.log('Message sequences from DB:', data.messages.map((m: any) => ({ seq: m.sequence, role: m.role, content: m.content.substring(0, 50) })))

  // Convert DB messages to UI message format
  const uiMessages = data.messages.map((dbMsg: any) => ({
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

  console.log('Sorted message sequences:', uiMessages.map(m => ({ seq: m.sequence, role: m.role, content: m.content.substring(0, 50) })))

  // Set or prepend messages for the session
  if (!messages.value[data.session_id]) {
    messages.value[data.session_id] = []
  }

  // Prepend historical messages (now sorted by sequence, oldest first)
  messages.value[data.session_id] = [...uiMessages, ...messages.value[data.session_id]]

  console.log(`📥 Loaded ${uiMessages.length} historical messages for session ${data.session_id}`)
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

// Load existing sessions on mount
onMounted(() => {
  if (agentWs.connected) {
    agentWs.send({ type: 'list_sessions' })
  }
})

// Watch for connection changes
watch(() => agentWs.connected, (connected) => {
  if (connected) {
    agentWs.send({ type: 'list_sessions' })
  }
})

// Watch for new messages and auto-scroll if user is near bottom
watch(activeMessages, () => {
  autoScrollIfNearBottom()
}, { deep: true, flush: 'post' })
</script>

<style scoped>
.agents-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
}

.header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  background: var(--card-bg);
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
}

/* Sessions Sidebar */
.sessions-sidebar {
  width: 300px;
  background: var(--card-bg);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
}

.sidebar-header h3 {
  margin: 0 0 12px 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

/* Session Filters */
.session-filters {
  display: flex;
  gap: 4px;
  padding: 12px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-primary);
}

.filter-tab {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 12px;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.filter-tab:hover {
  background: var(--bg-secondary);
  border-color: var(--accent-purple);
}

.filter-tab.active {
  background: var(--accent-purple);
  border-color: var(--accent-purple);
  color: white;
}

.filter-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 10px;
  font-size: 0.75rem;
  font-weight: 600;
}

.filter-tab.active .filter-count {
  background: rgba(255, 255, 255, 0.2);
}

.btn-new-session {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px;
  background: var(--accent-purple);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-new-session:hover:not(:disabled) {
  background: var(--accent-purple-hover);
}

.btn-new-session:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.sessions-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.no-sessions {
  text-align: center;
  padding: 32px 16px;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  margin-bottom: 8px;
  background: var(--bg-secondary);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.session-item:hover {
  background: var(--bg-tertiary);
}

.session-item.active {
  background: var(--accent-purple);
  color: white;
}

.session-item.ended {
  opacity: 0.7;
}

.session-item.ended:hover {
  opacity: 0.85;
}

/* Session Status Dot */
.session-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-right: 8px;
}

.session-status-dot.active,
.session-status-dot.processing,
.session-status-dot.idle {
  background: var(--status-success);
  box-shadow: 0 0 8px rgba(52, 211, 153, 0.5);
}

.session-status-dot.ended {
  background: var(--text-muted);
}

.session-status-dot.error {
  background: var(--status-error);
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
}

.session-info {
  flex: 1;
  overflow: hidden;
}

.session-name {
  font-size: 0.9rem;
  font-weight: 500;
  margin-bottom: 4px;
}

.session-meta {
  display: flex;
  gap: 12px;
  font-size: 0.8rem;
  opacity: 0.8;
  flex-wrap: wrap;
}

.session-messages {
  color: var(--text-secondary);
}

.session-cost {
  color: var(--accent-green);
  font-weight: 600;
  font-family: 'Monaco', 'Consolas', monospace;
}

.session-item.active .session-messages,
.session-item.active .session-cost {
  color: rgba(255, 255, 255, 0.95);
}

.session-status {
  text-transform: capitalize;
}

.session-status.processing {
  color: var(--accent-blue);
}

.session-status.error {
  color: #dc3545;
}

.session-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.btn-end-session,
.btn-delete-session {
  padding: 4px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  opacity: 0.6;
  transition: all 0.2s;
  border-radius: 4px;
}

.btn-end-session:hover,
.btn-delete-session:hover {
  opacity: 1;
}

.btn-delete-session:hover {
  background: rgba(220, 53, 69, 0.1);
  color: #dc3545;
}

.session-item.active .btn-end-session,
.session-item.active .btn-delete-session {
  color: white;
}

.session-item.active .btn-delete-session:hover {
  background: rgba(255, 255, 255, 0.2);
  color: #ff6b6b;
}

/* Chat Area with Metrics */
.chat-area-with-metrics {
  flex: 1;
  display: flex;
  background: var(--bg-primary);
  overflow: hidden;
  min-height: 0;
  gap: 12px;
  padding: 12px;
}

.chat-main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  overflow: hidden;
  min-height: 0;
}

/* Chat Area */
.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  overflow: hidden; /* Prevent overflow of the entire chat area */
  min-height: 0; /* Important for flex children */
}

/* Metrics Sidebar */
.metrics-sidebar {
  width: 320px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow-y: auto;
  min-height: 0;
  flex-shrink: 0;
}

.no-session-selected {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-state {
  text-align: center;
  color: var(--text-secondary);
}

.empty-state svg {
  margin-bottom: 16px;
}

.empty-state p {
  font-size: 0.95rem;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden; /* Contain children */
  min-height: 0; /* Important for flex children */
  position: relative; /* For absolutely positioned TodoWrite box and tool overlays */
}

/* Tool Overlays Container - positioned in top right */
.tool-overlays-container {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none; /* Allow clicks through container */
}

.tool-overlays-container > * {
  pointer-events: auto; /* Re-enable clicks on overlay items */
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  min-height: 0; /* Important for proper scrolling */
}

.message {
  margin-bottom: 24px;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.message-role {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-primary);
}

.message.assistant .message-role {
  color: var(--accent-purple);
}

.message-time {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.message-content {
  background: var(--card-bg);
  padding: 12px 16px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  font-size: 0.95rem;
  line-height: 1.6;
  color: var(--text-primary);
}

.message.user .message-content {
  background: var(--accent-purple);
  color: white;
  border-color: var(--accent-purple);
  margin-left: 48px;
}

.message.assistant .message-content {
  margin-right: 48px;
}

.message-content code {
  background: var(--bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.9em;
}

.message.user .message-content code {
  background: rgba(255, 255, 255, 0.2);
}

.message-content pre {
  background: var(--bg-secondary);
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}

.message-content .system-message {
  color: var(--text-secondary);
  font-style: italic;
  opacity: 0.7;
}

.message.isError .message-content {
  background: rgba(220, 53, 69, 0.1);
  border-color: rgba(220, 53, 69, 0.3);
  color: #dc3545;
}

.tool-use {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 4px 12px;
  background: var(--bg-secondary);
  border-radius: 12px;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.thinking-indicator {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  margin: 0 24px 24px 24px;
  background: linear-gradient(135deg, var(--accent-purple), var(--accent-purple-hover));
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.2);
  color: white;
  font-size: 0.95rem;
  font-weight: 500;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.thinking-dots {
  display: flex;
  gap: 6px;
}

.thinking-dots span {
  width: 10px;
  height: 10px;
  background: white;
  border-radius: 50%;
  animation: pulse 1.5s infinite ease-in-out;
  box-shadow: 0 0 8px rgba(255, 255, 255, 0.4);
}

.thinking-dots span:nth-child(1) {
  animation-delay: 0s;
}

.thinking-dots span:nth-child(2) {
  animation-delay: 0.15s;
}

.thinking-dots span:nth-child(3) {
  animation-delay: 0.3s;
}

@keyframes pulse {
  0%, 80%, 100% {
    opacity: 0.4;
    transform: scale(0.8) translateY(0);
  }
  40% {
    opacity: 1;
    transform: scale(1.2) translateY(-3px);
  }
}

.processing-indicator {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--card-bg);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-size: 0.9rem;
  opacity: 0.7;
}

.processing-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-purple);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Input Area */
.input-area {
  display: flex;
  gap: 12px;
  padding: 16px 24px;
  background: var(--card-bg);
  border-top: 1px solid var(--border-color);
  flex-shrink: 0; /* Never shrink the input area */
}

.message-input {
  flex: 1;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  font-size: 0.95rem;
  color: var(--text-primary);
  resize: none;
  font-family: inherit;
  transition: all 0.2s;
}

.message-input:focus {
  outline: none;
  border-color: var(--accent-purple);
}

.message-input:disabled {
  opacity: 0.5;
}

.btn-send {
  padding: 12px 20px;
  background: var(--accent-purple);
  color: white;
  border: none;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-send:hover:not(:disabled) {
  background: var(--accent-purple-hover);
}

.btn-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Dark mode adjustments */
:root[data-theme="dark"] {
  --accent-purple-hover: #7c3aed;
  --bg-tertiary: #2a2a2a;
}

/* Resume Session Styles */
.session-buttons {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.btn-resume-session {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-resume-session:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
}

.btn-resume-session:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--card-bg);
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  width: 90%;
  max-width: 600px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

@media (min-width: 1024px) {
  .modal-content {
    max-width: 750px;
    max-height: 90vh;
  }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
}

.modal-close {
  padding: 8px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.modal-close:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.modal-body {
  padding: 24px;
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.loading-sessions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-purple);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.no-sessions-available {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.no-sessions-available svg {
  margin-bottom: 16px;
}

.no-sessions-available p {
  margin: 0;
  font-size: 0.95rem;
}

.sessions-list-modal {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.session-card-modal {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.session-card-modal:hover {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
  transform: translateX(4px);
}

.session-card-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, var(--accent-purple), #A78BFA);
  border-radius: 8px;
  color: white;
  font-weight: 700;
  font-size: 1rem;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.3);
  overflow: hidden;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 6px;
}

.session-card-info {
  flex: 1;
  overflow: hidden;
}

.session-card-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-card-directory {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-card-meta {
  display: flex;
  gap: 12px;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.session-card-messages,
.session-card-time {
  display: inline-flex;
  gap: 4px;
  white-space: nowrap;
}

.session-card-arrow {
  padding: 8px;
  color: var(--text-secondary);
  transition: all 0.2s;
  flex-shrink: 0;
}

.session-card-modal:hover .session-card-arrow {
  color: var(--accent-purple);
  transform: translateX(4px);
}

/* Historical message styling */
.message.isHistorical .message-content {
  background: var(--bg-secondary);
  border-color: var(--border-color);
  opacity: 0.7;
}

.message.isHistorical .message-role {
  opacity: 0.7;
}

/* Form Styles */
.form-group {
  margin-bottom: 24px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: var(--text-primary);
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.95rem;
  color: var(--text-primary);
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--accent-purple);
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
}

.form-help {
  display: block;
  margin-top: 4px;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

/* Prompt Mode Toggle */
.prompt-mode-toggle {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.mode-btn {
  flex: 1;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn:hover {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
}

.mode-btn.active {
  background: var(--accent-purple);
  border-color: var(--accent-purple);
  color: white;
}

/* Agent Grid Selection */
.agents-loading {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-secondary);
  justify-content: center;
}

.loading-spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-purple);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.agents-empty {
  padding: 24px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-secondary);
  text-align: center;
  font-size: 0.9rem;
}

.agents-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.agent-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
  flex: 1;
  position: relative;
}

.agent-card:hover {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
  transform: translateY(-2px);
}

.agent-card.selected {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
}

.agent-card-color {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.agent-card-content {
  flex: 1;
  min-width: 0;
}

.agent-card-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.agent-card-model {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-top: 2px;
}

.agent-card-checkmark {
  color: var(--accent-purple);
  flex-shrink: 0;
}

/* Agent Preview */
.agent-preview {
  margin-top: 16px;
  padding: 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.agent-preview-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.agent-preview-header strong {
  color: var(--text-primary);
  font-size: 1rem;
}

.agent-model {
  display: inline-block;
  padding: 4px 8px;
  background: var(--accent-purple);
  color: white;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}

.agent-description {
  margin: 0 0 12px 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.4;
}

.agent-prompt-preview {
  margin-top: 12px;
}

.preview-label {
  margin: 0 0 8px 0;
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.prompt-content {
  padding: 12px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 0.85rem;
  line-height: 1.5;
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.tools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 12px;
}

.tool-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.tool-checkbox:hover {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
}

.tool-checkbox input[type="checkbox"] {
  margin: 0;
}

.tool-checkbox span {
  font-size: 0.9rem;
  color: var(--text-primary);
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  padding: 24px;
  border-top: 1px solid var(--border-color);
  background: var(--card-bg);
  flex-shrink: 0;
}

.btn-cancel {
  padding: 12px 24px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  background: var(--bg-tertiary);
}

.btn-create {
  padding: 12px 24px;
  background: var(--accent-purple);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-create:hover:not(:disabled) {
  background: var(--accent-purple-hover);
}

.btn-create:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  position: relative;
}

.btn-create:disabled .btn-spinner {
  opacity: 1;
}

.btn-spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-right: 8px;
}

.btn-spinner-small {
  width: 12px;
  height: 12px;
  border: 1.5px solid var(--accent-purple);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-right: 0;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.selected-session-info {
  padding: 20px;
  background: var(--bg-secondary);
  border-radius: 8px;
  margin-bottom: 24px;
}

.selected-session-info h3 {
  margin: 0 0 8px 0;
  color: var(--text-primary);
}

.selected-session-info p {
  margin: 0;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.selected-session-info code {
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.85em;
}

.resume-session-options {
  padding: 24px 0;
}

/* Permission Requests */
.permission-requests {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  max-height: 200px;
  overflow-y: auto;
}

.permission-request {
  background: linear-gradient(135deg, #fff3cd, #fef5e7);
  border: 1px solid #ffc107;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(255, 193, 7, 0.2);
  animation: slideIn 0.3s ease-out;
}

.permission-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.permission-icon {
  font-size: 1.2rem;
}

.permission-title {
  font-weight: 600;
  color: #856404;
  flex: 1;
}

.permission-time {
  font-size: 0.8rem;
  color: #856404;
  opacity: 0.7;
}

.permission-description {
  color: #856404;
  font-size: 0.9rem;
  line-height: 1.4;
  margin-bottom: 12px;
  padding: 8px 0;
}

.permission-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.btn-approve,
.btn-deny {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-approve {
  background: #28a745;
  color: white;
}

.btn-approve:hover {
  background: #218838;
}

.btn-deny {
  background: #dc3545;
  color: white;
}

.btn-deny:hover {
  background: #c82333;
}

/* Permission decision messages */
.message.isPermissionDecision .message-content {
  background: var(--bg-secondary);
  border-color: var(--border-color);
  opacity: 0.8;
  font-style: italic;
}

/* Execution status messages */
.message.isExecutionStatus .message-content {
  background: linear-gradient(135deg, #e8f5e9, #c8e6c9);
  border-color: #4caf50;
  color: #2e7d32;
  font-weight: 500;
  font-style: normal;
  animation: slideIn 0.3s ease-out;
}

/* Tool result messages */
.message.isToolResult .message-content {
  background: var(--bg-secondary);
  border: 2px solid var(--accent-purple);
  border-left: 4px solid var(--accent-purple);
  animation: slideIn 0.3s ease-out;
}

/* System messages styling */
.message.system .message-role {
  color: var(--text-secondary);
}

.message.system .message-content {
  background: transparent;
  border: none;
  padding: 8px 12px;
  font-size: 0.9rem;
  text-align: center;
  color: var(--text-secondary);
}

/* TodoWrite Box Styles */
.todo-write-box {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 320px;
  max-height: 400px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  z-index: 100;
  animation: fadeIn 0.3s ease-out;
  overflow: hidden;
}

.todo-box-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, var(--accent-purple), var(--accent-purple-hover));
  color: white;
  border-bottom: 1px solid var(--border-color);
}

.todo-box-icon {
  font-size: 1.2rem;
}

.todo-box-title {
  font-size: 0.9rem;
  font-weight: 600;
}

.todo-list {
  max-height: 320px;
  overflow-y: auto;
  padding: 8px;
}

.todo-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 8px 12px;
  margin-bottom: 6px;
  background: var(--bg-secondary);
  border-radius: 8px;
  transition: all 0.2s ease;
  animation: slideInRight 0.3s ease-out;
}

.todo-item:hover {
  background: var(--bg-tertiary);
  transform: translateX(-2px);
}

.todo-item.completed {
  opacity: 0.7;
}

.todo-item.completed .todo-text {
  text-decoration: line-through;
  color: var(--text-secondary);
}

.todo-status-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.in-progress-icon {
  animation: spin 2s linear infinite;
}

.todo-content {
  flex: 1;
  min-width: 0;
}

.todo-text {
  font-size: 0.9rem;
  color: var(--text-primary);
  line-height: 1.4;
  word-wrap: break-word;
}

.todo-active-form {
  font-size: 0.85rem;
  color: var(--accent-purple);
  font-style: italic;
  margin-top: 2px;
}

/* Tool Execution Bar Styles */
.tool-execution-bar {
  background: linear-gradient(135deg, #e3f2fd, #bbdefb);
  border: 1px solid #2196f3;
  border-radius: 12px;
  margin: 0 24px 12px 24px;
  padding: 12px 16px;
  animation: slideInTop 0.3s ease-out;
  box-shadow: 0 4px 16px rgba(33, 150, 243, 0.25);
  position: relative;
  z-index: 50;
}

.tool-execution-content {
  display: flex;
  align-items: center;
  gap: 12px;
  position: relative;
}

.tool-execution-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.tool-execution-details {
  flex: 1;
  min-width: 0;
}

.tool-execution-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: #1565c0;
  margin-bottom: 2px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.tool-execution-detail-badge {
  display: inline-block;
  padding: 2px 8px;
  background: var(--bg-secondary);
  border-radius: 4px;
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 500;
  font-family: 'Monaco', 'Menlo', monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 300px;
}

.tool-execution-info {
  font-size: 0.85rem;
  color: #1976d2;
  font-family: 'Monaco', 'Menlo', monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tool-execution-pulse {
  position: absolute;
  top: 50%;
  right: 0;
  transform: translateY(-50%);
  width: 8px;
  height: 8px;
  background: #2196f3;
  border-radius: 50%;
  animation: pulse 1.5s infinite ease-in-out;
}

/* Animations */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideInRight {
  from {
    opacity: 0;
    transform: translateX(20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes slideInTop {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes pulse {
  0%, 80%, 100% {
    opacity: 0.3;
    transform: translateY(-50%) scale(0.8);
  }
  40% {
    opacity: 1;
    transform: translateY(-50%) scale(1.2);
  }
}

/* Responsive adjustments for live agents components */
@media (max-width: 1024px) {
  .chat-area-with-metrics {
    padding: 8px;
    gap: 8px;
  }

  .metrics-sidebar {
    width: 280px;
  }
}

@media (max-width: 768px) {
  .sessions-sidebar {
    width: 240px;
  }

  .todo-write-box {
    width: 280px;
    right: 8px;
    top: 8px;
  }

  .chat-area-with-metrics {
    flex-direction: column-reverse;
    padding: 0;
    gap: 0;
  }

  .metrics-sidebar {
    width: 100%;
    max-height: 300px;
    border-radius: 0;
    border: none;
    border-top: 1px solid var(--border-color);
    flex-shrink: 1;
  }

  .chat-main-area {
    flex: 1;
    min-height: 300px;
  }
}

@media (max-width: 640px) {
  .agents-container {
    flex-direction: column;
  }

  .sessions-sidebar {
    width: 100%;
    height: 200px;
    border-right: none;
    border-bottom: 1px solid var(--border-color);
  }

  .modal-content {
    width: 95%;
    margin: 20px;
  }

  .tools-grid {
    grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
  }

  .modal-actions {
    flex-direction: column;
  }

  .btn-cancel,
  .btn-create {
    width: 100%;
  }

  .todo-write-box {
    position: relative;
    top: auto;
    right: auto;
    width: 100%;
    margin: 0 0 12px 0;
    border-radius: 8px;
  }

  .tool-execution-bar {
    margin: 0 0 12px 0;
  }
}
</style>