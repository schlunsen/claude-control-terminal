<template>
  <div class="permission-request">
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
        @click="$emit('deny', permission)"
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
        @click="$emit('approve', permission)"
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
</template>

<script setup lang="ts">
import { formatTime } from '~/utils/agents/messageFormatters'

interface Permission {
  request_id: string
  description: string
  timestamp: string | Date
}

interface Props {
  permission: Permission
}

defineProps<Props>()
defineEmits<{
  (e: 'approve', permission: Permission): void
  (e: 'deny', permission: Permission): void
}>()
</script>

<style scoped>
.permission-request {
  background: var(--color-warning-alpha);
  border: 2px solid var(--color-warning);
  border-radius: 0.5rem;
  padding: 1rem;
  margin-bottom: 1rem;
}

.permission-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.permission-icon {
  font-size: 1.5rem;
}

.permission-title {
  flex: 1;
  font-weight: 600;
  color: var(--color-text-primary);
}

.permission-time {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.permission-description {
  color: var(--color-text-primary);
  margin-bottom: 1rem;
  line-height: 1.5;
}

.permission-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

.btn-deny,
.btn-approve {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-deny {
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
}

.btn-deny:hover {
  background: var(--color-error);
  color: white;
}

.btn-approve {
  background: var(--color-success);
  color: white;
}

.btn-approve:hover {
  background: var(--color-success-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}
</style>
