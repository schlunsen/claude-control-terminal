<template>
  <div v-if="show" class="modal-overlay" @click="$emit('close')" @keydown.enter="handleEnterKey">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h2>Create New Session</h2>
        <button @click="$emit('close')" class="modal-close">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <div class="modal-body">
        <!-- Working Directory -->
        <div class="form-group">
          <label for="working-directory">Working Directory</label>
          <input
            id="working-directory"
            v-model="formData.workingDirectory"
            @change="$emit('workingDirectoryChange')"
            @blur="$emit('workingDirectoryChange')"
            type="text"
            placeholder="/home/user/projects"
            class="form-input"
          />
          <small class="form-help">The directory where the agent will work</small>
        </div>

        <!-- Permission Mode -->
        <div class="form-group">
          <label for="permission-mode">Permission Mode</label>
          <select id="permission-mode" v-model="formData.permissionMode" class="form-select">
            <option value="default">Default (ask for permissions)</option>
            <option value="acceptEdits">Allow All (full permissions)</option>
            <option value="plan">Read Only (no file modifications)</option>
          </select>
          <small class="form-help">Control what permissions the agent has</small>
        </div>

        <!-- Model Provider -->
        <div class="form-group">
          <label for="model-provider">Model Provider</label>
          <select
            id="model-provider"
            v-model="formData.modelProvider"
            class="form-select"
            :disabled="loadingProviders"
          >
            <option v-for="provider in providers" :key="provider.id" :value="provider.id">
              {{ provider.icon }} {{ provider.name }}
            </option>
          </select>
          <small class="form-help" v-if="currentProvider">
            Current: {{ currentProvider.provider_id }} {{ currentProvider.model_name ? `(${currentProvider.model_name})` : '' }}
          </small>
          <small class="form-help" v-else>Choose the AI model provider</small>
        </div>

        <!-- Model Selection -->
        <div class="form-group">
          <label for="model">Model</label>
          <select
            id="model"
            v-model="formData.model"
            class="form-select"
            :disabled="!formData.modelProvider || loadingProviders"
          >
            <option v-if="!formData.modelProvider" value="">Select a provider first</option>
            <template v-else>
              <option
                v-for="model in availableModels"
                :key="model"
                :value="model"
              >
                {{ model }}
              </option>
            </template>
          </select>
          <small class="form-help">Select the AI model to use</small>
        </div>

        <!-- System Prompt Mode -->
        <div class="form-group">
          <label for="prompt-mode">System Prompt</label>
          <div class="prompt-mode-toggle">
            <button
              type="button"
              class="mode-btn"
              :class="{ active: formData.promptMode === 'agent' }"
              @click="formData.promptMode = 'agent'"
            >
              📦 Project Agent
            </button>
            <button
              type="button"
              class="mode-btn"
              :class="{ active: formData.promptMode === 'custom' }"
              @click="formData.promptMode = 'custom'"
            >
              ✏️ Custom
            </button>
          </div>
          <small class="form-help">Choose a project agent or write a custom system prompt</small>
        </div>

        <!-- Agent Selection Mode -->
        <div v-if="formData.promptMode === 'agent'" class="form-group">
          <label for="agent-select">Select Agent</label>
          <div v-if="loadingAgents" class="agents-loading">
            <div class="loading-spinner-small"></div>
            <span>Loading agents...</span>
          </div>
          <div v-else-if="agents.length === 0" class="agents-empty">
            No agents found. Make sure you've set a valid working directory.
          </div>
          <div v-else class="agents-grid">
            <button
              v-for="agent in agents"
              :key="agent.name"
              type="button"
              class="agent-card"
              :class="{ selected: formData.selectedAgent === agent.name }"
              @click="handleAgentSelect(agent.name)"
            >
              <div class="agent-card-color" :style="{ backgroundColor: agent.color || '#8B5CF6' }"></div>
              <div class="agent-card-content">
                <div class="agent-card-name">{{ agent.name }}</div>
                <div v-if="agent.model" class="agent-card-model">{{ agent.model }}</div>
              </div>
              <div v-if="formData.selectedAgent === agent.name" class="agent-card-checkmark">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
              </div>
            </button>
          </div>
          <small class="form-help">Select from available agents</small>

          <!-- Agent Preview -->
          <div v-if="formData.selectedAgent && selectedAgentPreview" class="agent-preview">
            <div class="agent-preview-header">
              <strong>{{ selectedAgentPreview.name }}</strong>
              <span v-if="selectedAgentPreview.model" class="agent-model">{{ selectedAgentPreview.model }}</span>
            </div>
            <p v-if="selectedAgentPreview.description" class="agent-description">
              {{ selectedAgentPreview.description }}
            </p>
            <div class="agent-prompt-preview">
              <p class="preview-label">System Prompt Preview:</p>
              <div class="prompt-content">{{ selectedAgentPreview.system_prompt.substring(0, 300) }}...</div>
            </div>
          </div>
        </div>

        <!-- Custom Prompt Mode -->
        <div v-if="formData.promptMode === 'custom'" class="form-group">
          <label for="system-prompt">Custom System Prompt</label>
          <textarea
            id="system-prompt"
            v-model="formData.systemPrompt"
            placeholder="You are a helpful AI assistant."
            class="form-textarea"
            rows="4"
          ></textarea>
          <small class="form-help">Enter custom instructions for the agent</small>
        </div>

        <!-- Available Tools -->
        <div class="form-group">
          <label>Available Tools</label>
          <div class="tools-grid">
            <label class="tool-checkbox">
              <input type="checkbox" v-model="formData.tools" value="Read" />
              <span class="checkbox-custom"></span>
              <span class="checkbox-label">Read</span>
            </label>
            <label class="tool-checkbox">
              <input type="checkbox" v-model="formData.tools" value="Write" />
              <span class="checkbox-custom"></span>
              <span class="checkbox-label">Write</span>
            </label>
            <label class="tool-checkbox">
              <input type="checkbox" v-model="formData.tools" value="Edit" />
              <span class="checkbox-custom"></span>
              <span class="checkbox-label">Edit</span>
            </label>
            <label class="tool-checkbox">
              <input type="checkbox" v-model="formData.tools" value="Bash" />
              <span class="checkbox-custom"></span>
              <span class="checkbox-label">Bash</span>
            </label>
            <label class="tool-checkbox">
              <input type="checkbox" v-model="formData.tools" value="Search" />
              <span class="checkbox-custom"></span>
              <span class="checkbox-label">Search</span>
            </label>
            <label class="tool-checkbox">
              <input type="checkbox" v-model="formData.tools" value="Grep" />
              <span class="checkbox-custom"></span>
              <span class="checkbox-label">Grep</span>
            </label>
            <label class="tool-checkbox">
              <input type="checkbox" v-model="formData.tools" value="TodoWrite" />
              <span class="checkbox-custom"></span>
              <span class="checkbox-label">TodoWrite</span>
            </label>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button @click="$emit('close')" class="btn-cancel" :disabled="creating">
          Cancel
        </button>
        <button
          @click="$emit('create', formData)"
          class="btn-create"
          :disabled="!formData.workingDirectory || creating"
          :title="!creating ? 'Press Enter to submit' : ''"
        >
          <div v-if="creating" class="btn-spinner"></div>
          <span v-if="!creating">Create Session</span>
          <span v-else>Creating...</span>
          <kbd v-if="!creating" class="kbd-hint">↵</kbd>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch, nextTick } from 'vue'

interface Provider {
  id: string
  name: string
  icon: string
  models: string[]
  base_url?: string
}

interface Agent {
  name: string
  model?: string
  color?: string
  description?: string
  system_prompt?: string
}

interface SessionFormData {
  workingDirectory: string
  permissionMode: string
  modelProvider: string
  model: string
  systemPrompt: string
  promptMode: 'agent' | 'custom'
  selectedAgent: string
  tools: string[]
}

interface Props {
  show: boolean
  formData: SessionFormData
  providers: Provider[]
  currentProvider: any
  agents: Agent[]
  selectedAgentPreview: Agent | null
  loadingProviders: boolean
  loadingAgents: boolean
  creating: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'create', formData: SessionFormData): void
  (e: 'workingDirectoryChange'): void
  (e: 'agentSelect', agentName: string): void
}>()

const availableModels = computed(() => {
  const provider = props.providers.find(p => p.id === props.formData.modelProvider)
  return provider?.models || []
})

const handleAgentSelect = (agentName: string) => {
  props.formData.selectedAgent = agentName
  emit('agentSelect', agentName)
}

const handleEnterKey = (event: KeyboardEvent) => {
  // Only trigger if we're not in a textarea (allow Enter in custom prompt)
  const target = event.target as HTMLElement
  if (target.tagName === 'TEXTAREA') {
    return
  }

  // Only trigger if working directory is set and not already creating
  if (props.formData.workingDirectory && !props.creating) {
    event.preventDefault()
    emit('create', props.formData)
  }
}

// Auto-focus the working directory input when modal opens
watch(() => props.show, (show) => {
  if (show) {
    nextTick(() => {
      const workingDirInput = document.getElementById('working-directory') as HTMLInputElement
      if (workingDirInput) {
        workingDirInput.focus()
      }
    })
  }
})
</script>

<style scoped>
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
  max-width: 750px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
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

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  padding: 24px;
  border-top: 1px solid var(--border-color);
  background: var(--card-bg);
  flex-shrink: 0;
}

/* Form Styles */
.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.95rem;
}

.form-input,
.form-select,
.form-textarea {
  width: 100%;
  padding: 10px 12px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 0.95rem;
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--accent-purple);
  box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.1);
}

.form-textarea {
  resize: vertical;
  font-family: 'Monaco', 'Courier New', monospace;
  line-height: 1.5;
}

.form-help {
  display: block;
  margin-top: 6px;
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
  border-color: var(--accent-purple);
  background: var(--bg-tertiary);
}

.mode-btn.active {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
  color: var(--accent-purple);
}

/* Agents Grid */
.agents-loading,
.agents-empty {
  padding: 40px 20px;
  text-align: center;
  color: var(--text-secondary);
}

.agents-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.loading-spinner-small {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-purple);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
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
  justify-content: space-between;
  margin-bottom: 8px;
}

.agent-model {
  font-size: 0.85rem;
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: 4px 8px;
  border-radius: 4px;
}

.agent-description {
  color: var(--text-secondary);
  margin-bottom: 12px;
  line-height: 1.5;
}

.agent-prompt-preview {
  margin-top: 12px;
}

.preview-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.prompt-content {
  padding: 12px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.5;
  max-height: 150px;
  overflow-y: auto;
}

/* Tools Grid */
.tools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 12px;
}

.tool-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
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
  display: none;
}

.checkbox-custom {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border-color);
  border-radius: 4px;
  position: relative;
  transition: all 0.2s;
}

.tool-checkbox input[type="checkbox"]:checked + .checkbox-custom {
  background: var(--accent-purple);
  border-color: var(--accent-purple);
}

.tool-checkbox input[type="checkbox"]:checked + .checkbox-custom::after {
  content: '✓';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: white;
  font-size: 12px;
  font-weight: bold;
}

.checkbox-label {
  font-size: 0.9rem;
  color: var(--text-primary);
  font-weight: 500;
}

/* Action Buttons */
.btn-cancel,
.btn-create {
  padding: 12px 24px;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-cancel {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.btn-cancel:hover:not(:disabled) {
  background: var(--bg-tertiary);
}

.btn-create {
  background: var(--accent-purple);
  color: white;
  border: none;
}

.btn-create:hover:not(:disabled) {
  background: var(--accent-purple-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
}

.btn-create:disabled,
.btn-cancel:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Keyboard Hint */
.kbd-hint {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  height: 24px;
  padding: 0 6px;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 4px;
  font-size: 0.9rem;
  font-weight: 600;
  color: white;
  font-family: inherit;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  margin-left: 4px;
}

/* Responsive */
@media (max-width: 640px) {
  .modal-content {
    width: 95%;
    margin: 20px;
  }

  .tools-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }

  .modal-actions {
    flex-direction: column;
  }

  .btn-cancel,
  .btn-create {
    width: 100%;
    justify-content: center;
  }
}
</style>
