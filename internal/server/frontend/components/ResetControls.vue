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
    console.error('Error loading reset status:', error)
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
  border: none;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.btn:active {
  transform: translateY(0);
}

.btn-primary {
  background: #2d3748;
  color: #f7fafc;
  border: 1px solid #4a5568;
}

.btn-primary:hover {
  background: #1a202c;
  border-color: #2d3748;
}

.btn-secondary {
  background: #f7fafc;
  color: #4a5568;
  border: 1px solid #cbd5e0;
}

.btn-secondary:hover {
  background: #edf2f7;
  border-color: #a0aec0;
}

.reset-info {
  margin-top: 16px;
  padding: 12px 16px;
  background: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: 6px;
  font-size: 0.875rem;
  color: #92400e;
}

.reset-info strong {
  color: #78350f;
}
</style>
