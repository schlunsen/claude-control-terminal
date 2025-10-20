<template>
  <aside class="sessions-sidebar">
    <div class="sidebar-header">
      <h3>Sessions</h3>
      <div class="session-buttons">
        <button @click="$emit('create-new')" class="btn-new-session" :disabled="!connected || creating">
          <svg v-if="!creating" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          <div v-if="creating" class="btn-spinner-small"></div>
          <span v-if="!creating">New Session</span>
          <span v-else>Creating...</span>
        </button>
        <button @click="$emit('resume')" class="btn-resume-session" :disabled="!connected">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M3 15v4c0 1.1.9 2 2 2h14a2 2 0 0 0 2-2v-4M17 9l-5 5-5-5M12 12.8V2.5"/>
          </svg>
          Resume Session
        </button>
      </div>
    </div>

    <!-- Session Filter Tabs -->
    <SessionFilters
      :active-filter="activeFilter"
      :filters="filters"
      @update:active-filter="$emit('update:active-filter', $event)"
    />

    <div class="sessions-list">
      <div v-if="sessions.length === 0" class="no-sessions">
        No {{ activeFilter }} sessions
      </div>
      <SessionItem
        v-for="session in sessions"
        :key="session.id"
        :session="session"
        :is-active="activeSessionId === session.id"
        @select="$emit('select', $event)"
        @end="$emit('end', $event)"
        @delete="$emit('delete', $event)"
      />
    </div>
  </aside>
</template>

<script setup lang="ts">
import SessionFilters from './SessionFilters.vue'
import SessionItem from './SessionItem.vue'

interface Props {
  sessions: any[]
  activeSessionId: string | null
  activeFilter: string
  filters: any[]
  connected: boolean
  creating: boolean
}

defineProps<Props>()

defineEmits<{
  'create-new': []
  'resume': []
  'update:active-filter': [value: string]
  'select': [sessionId: string]
  'end': [sessionId: string]
  'delete': [sessionId: string]
}>()
</script>

<style scoped>
.sessions-sidebar {
  width: 300px;
  background: var(--card-bg);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.sidebar-header h3 {
  margin: 0 0 12px 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.session-buttons {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.btn-new-session,
.btn-resume-session {
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

.btn-resume-session {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.btn-new-session:hover:not(:disabled) {
  background: var(--accent-purple-hover);
}

.btn-resume-session:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--accent-purple);
}

.btn-new-session:disabled,
.btn-resume-session:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-spinner-small {
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

.sessions-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  min-height: 0;
}

.no-sessions {
  text-align: center;
  padding: 32px 16px;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

/* Responsive */
@media (max-width: 768px) {
  .sessions-sidebar {
    width: 240px;
  }
}

@media (max-width: 480px) {
  .sessions-sidebar {
    width: 100%;
    height: 200px;
    border-right: none;
    border-bottom: 1px solid var(--border-color);
  }
}
</style>
