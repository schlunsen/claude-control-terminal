<template>
  <div
    class="session-item"
    :class="{
      active: isActive,
      ended: session.status === 'ended'
    }"
    @click="$emit('select', session.id)"
  >
    <div class="session-status-dot" :class="session.status"></div>
    <img
      :src="avatar.avatar"
      :alt="avatar.name"
      class="session-avatar"
    />
    <div class="session-info">
      <div class="session-name">{{ avatar.name }}</div>
      <div class="session-meta">
        <span class="session-id">{{ session.id.slice(0, 8) }}</span>
        <span class="session-status" :class="session.status">{{ session.status }}</span>
        <span class="session-messages">{{ session.message_count }} messages</span>
        <span v-if="session.cost_usd && session.cost_usd > 0" class="session-cost">
          ${{ session.cost_usd.toFixed(4) }}
        </span>
      </div>
    </div>
    <div class="session-actions">
      <button
        v-if="session.status !== 'ended'"
        @click.stop="$emit('end', session.id)"
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
        @click.stop="$emit('delete', session.id)"
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
</template>

<script setup lang="ts">
import { useCharacterAvatar } from '~/composables/useCharacterAvatar'

interface Session {
  id: string
  status: string
  message_count: number
  cost_usd?: number
}

interface Props {
  session: Session
  isActive: boolean
}

const props = defineProps<Props>()
defineEmits<{
  (e: 'select', sessionId: string): void
  (e: 'end', sessionId: string): void
  (e: 'delete', sessionId: string): void
}>()

const avatar = useCharacterAvatar(props.session.id)
</script>

<style scoped>
.session-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.session-item:hover {
  background: var(--color-bg-secondary);
}

.session-item.active {
  background: var(--color-primary-alpha);
  border-left: 3px solid var(--color-primary);
}

.session-item.ended {
  opacity: 0.6;
}

.session-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.session-status-dot.active {
  background: var(--color-success);
  box-shadow: 0 0 8px var(--color-success);
}

.session-status-dot.ended {
  background: var(--color-text-tertiary);
}

.session-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.session-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.session-name {
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 0.875rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-meta {
  display: flex;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  flex-wrap: wrap;
}

.session-id {
  font-family: 'Monaco', 'Courier New', monospace;
}

.session-status {
  text-transform: capitalize;
}

.session-status.active {
  color: var(--color-success);
}

.session-cost {
  color: var(--color-warning);
  font-weight: 600;
}

.session-actions {
  display: flex;
  gap: 0.25rem;
  opacity: 0;
  transition: opacity 0.2s;
}

.session-item:hover .session-actions {
  opacity: 1;
}

.btn-end-session,
.btn-delete-session {
  padding: 0.25rem;
  border: none;
  background: transparent;
  color: var(--color-text-tertiary);
  cursor: pointer;
  border-radius: 0.25rem;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-end-session:hover {
  background: var(--color-error-alpha);
  color: var(--color-error);
}

.btn-delete-session:hover {
  background: var(--color-error-alpha);
  color: var(--color-error);
}
</style>
