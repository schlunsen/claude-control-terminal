<template>
  <div v-if="show" class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h2>Resume Session</h2>
        <button @click="$emit('close')" class="modal-close">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <div class="modal-body">
        <!-- Loading State -->
        <div v-if="loading" class="loading-sessions">
          <div class="loading-spinner"></div>
          Loading available sessions...
        </div>

        <!-- No Sessions State -->
        <div v-else-if="sessions.length === 0" class="no-sessions-available">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.5">
            <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
          </svg>
          <p>No previous sessions found</p>
        </div>

        <!-- Session Selection List -->
        <div v-else-if="!selectedSession" class="sessions-list-modal">
          <div
            v-for="session in sessions"
            :key="session.conversation_id"
            class="session-card-modal"
            @click="$emit('selectSession', session)"
          >
            <div class="session-card-avatar">
              <img
                :src="`/avatars/${session.session_name || 'default'}.png`"
                :alt="session.session_name"
                @error="handleImageError"
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
            <h3>{{ selectedSession.session_name || 'Selected Session' }}</h3>
            <p>Original working directory: <code>{{ selectedSession.working_directory }}</code></p>
          </div>

          <div class="form-group">
            <label for="resume-working-directory">Working Directory</label>
            <input
              id="resume-working-directory"
              v-model="formData.workingDirectory"
              type="text"
              :placeholder="selectedSession.working_directory"
              class="form-input"
            />
            <small class="form-help">Directory where the agent will work (defaults to original)</small>
          </div>

          <div class="form-group">
            <label for="resume-permission-mode">Permission Mode</label>
            <select id="resume-permission-mode" v-model="formData.permissionMode" class="form-select">
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
              v-model="formData.systemPrompt"
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

          <div class="modal-actions">
            <button @click="$emit('back')" class="btn-cancel" :disabled="resuming">
              Back
            </button>
            <button @click="$emit('resume', formData)" class="btn-create" :disabled="resuming">
              <div v-if="resuming" class="btn-spinner"></div>
              <span v-if="!resuming">Resume Session</span>
              <span v-else>Resuming...</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatRelativeTime } from '~/utils/agents/messageFormatters'

interface Session {
  conversation_id: string
  session_name?: string
  working_directory?: string
  total_messages: number
  last_activity: string | Date
}

interface ResumeFormData {
  workingDirectory: string
  permissionMode: string
  systemPrompt: string
  tools: string[]
}

interface Props {
  show: boolean
  sessions: Session[]
  selectedSession: Session | null
  formData: ResumeFormData
  loading: boolean
  resuming: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'selectSession', session: Session): void
  (e: 'back'): void
  (e: 'resume', formData: ResumeFormData): void
}>()

const handleImageError = (event: Event) => {
  const target = event.target as HTMLImageElement
  target.src = '/avatars/default.png'
}
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

/* Loading & Empty States */
.loading-sessions,
.no-sessions-available {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top-color: var(--accent-purple);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Session List */
.sessions-list-modal {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.session-card-modal {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.session-card-modal:hover {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.session-card-avatar {
  flex-shrink: 0;
}

.avatar-image {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--border-color);
}

.session-card-info {
  flex: 1;
  min-width: 0;
}

.session-card-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 1rem;
  margin-bottom: 4px;
}

.session-card-directory {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 6px;
  font-family: 'Monaco', 'Courier New', monospace;
}

.session-card-meta {
  display: flex;
  gap: 12px;
  font-size: 0.85rem;
  color: var(--text-tertiary);
}

.session-card-arrow {
  color: var(--text-tertiary);
  flex-shrink: 0;
  transition: transform 0.2s;
}

.session-card-modal:hover .session-card-arrow {
  color: var(--accent-purple);
  transform: translateX(4px);
}

/* Resume Options */
.resume-session-options {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.selected-session-info {
  padding: 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.selected-session-info h3 {
  margin: 0 0 8px 0;
  font-size: 1.1rem;
  color: var(--text-primary);
}

.selected-session-info p {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.selected-session-info code {
  padding: 2px 6px;
  background: var(--bg-tertiary);
  border-radius: 4px;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.85rem;
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
.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

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

  .session-card-modal {
    padding: 12px;
    gap: 12px;
  }

  .avatar-image {
    width: 40px;
    height: 40px;
  }
}
</style>
