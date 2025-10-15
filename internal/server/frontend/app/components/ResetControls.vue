<template>
  <div class="reset-controls">
    <div class="actions">
      <button class="btn btn-primary" @click="handleSoftReset">
        🔄 Soft Reset
      </button>
      <button
        v-if="resetActive"
        class="btn btn-secondary"
        @click="handleClearReset"
      >
        ↩️ Undo Reset
      </button>
    </div>

    <div v-if="resetActive" class="reset-info">
      <strong>Soft Reset Active:</strong>
      Reset since {{ formatDate(resetTimestamp) }} - {{ resetReason }}
    </div>
  </div>
</template>

<script setup lang="ts">
const resetActive = ref(false)
const resetTimestamp = ref('')
const resetReason = ref('')

// WebSocket integration
const { on } = useWebSocket()

// Reload reset status on WebSocket events
on('onReset', loadResetStatus)
on('onStatsUpdate', loadResetStatus)

async function loadResetStatus() {
  try {
    const { data } = await useFetch<any>('/api/reset/status')
    if (data.value) {
      resetActive.value = data.value.active || false
      resetTimestamp.value = data.value.timestamp || ''
      resetReason.value = data.value.reason || ''
    }
  } catch (error) {
    // Error loading reset status
  }
}

async function handleSoftReset() {
  if (!confirm('Apply soft reset? This will reset counts to zero without deleting any data. You can undo this later.')) {
    return
  }

  try {
    const response = await $fetch<any>('/api/reset/soft', { method: 'POST' })

    if (response.status === 'reset') {
      alert(`✅ ${response.message}\n\nPrevious: ${response.previousTokens.toLocaleString()} tokens, ${response.previousConversations} conversations`)
      await loadResetStatus()
    } else {
      alert('❌ Reset failed: ' + (response.error || 'Unknown error'))
    }
  } catch (error: any) {
    alert('❌ Error performing soft reset: ' + error.message)
  }
}

async function handleClearReset() {
  if (!confirm('Restore original counts? This will undo the soft reset.')) {
    return
  }

  try {
    const response = await $fetch<any>('/api/reset', { method: 'DELETE' })
    alert(`✅ ${response.message}`)
    await loadResetStatus()
  } catch (error: any) {
    alert('❌ Error clearing reset: ' + error.message)
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleString()
}

// Load on mount
onMounted(() => {
  loadResetStatus()
})
</script>

<style scoped>
.reset-controls {
  margin-top: 24px;
}

.actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.btn {
  padding: 10px 20px;
  border: 2px solid;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

.btn:active {
  transform: translateY(0);
}

.btn-primary {
  background: transparent;
  color: var(--accent-purple);
  border-color: var(--accent-purple);
}

.btn-primary:hover {
  background: var(--accent-purple);
  color: var(--bg-primary);
}

.btn-secondary {
  background: transparent;
  color: var(--accent-cyan);
  border-color: var(--accent-cyan);
}

.btn-secondary:hover {
  background: var(--accent-cyan);
  color: var(--bg-primary);
}

.reset-info {
  margin-top: 16px;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--accent-yellow);
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.reset-info strong {
  color: var(--accent-yellow);
  font-weight: 600;
}

@media (max-width: 768px) {
  .actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
  }
}
</style>
