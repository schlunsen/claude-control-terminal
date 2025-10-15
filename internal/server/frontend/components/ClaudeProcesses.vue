<template>
  <div class="claude-processes">
    <h2 class="section-title">Claude Processes</h2>

    <div v-if="processes.length === 0" class="empty-state">
      No Claude processes detected
    </div>

    <div v-else class="process-list">
      <div
        v-for="proc in processes"
        :key="proc.PID"
        class="process-item"
      >
        <div class="process-info">
          <div class="process-pid">PID {{ proc.PID }}</div>
          <div class="process-command">{{ proc.Command }}</div>
          <div class="process-dir">{{ proc.WorkingDir }}</div>
        </div>
        <div class="process-status">{{ proc.Status }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Process } from '~/types/analytics'

const processes = ref<Process[]>([])

// WebSocket integration
const { on } = useWebSocket()

// Reload processes on WebSocket events
on('onStatsUpdate', loadProcesses)
on('onReset', loadProcesses)

async function loadProcesses() {
  try {
    const { data } = await useFetch<any>('/api/processes')
    if (data.value?.processes) {
      processes.value = data.value.processes
    }
  } catch (error) {
    console.error('Error loading Claude processes:', error)
  }
}

// Load on mount
onMounted(() => {
  loadProcesses()
})
</script>

<style scoped>
.claude-processes {
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 32px;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24px;
  letter-spacing: -0.01em;
}

.process-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.process-item {
  padding: 16px;
  background: #fafafa;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  transition: border-color 0.2s;
}

.process-item:hover {
  border-color: #d0d0d0;
}

.process-info {
  flex: 1;
}

.process-pid {
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.875rem;
  color: #1a1a1a;
  font-weight: 500;
  margin-bottom: 4px;
}

.process-command {
  font-size: 0.875rem;
  color: #666;
  margin-bottom: 4px;
  word-break: break-all;
}

.process-dir {
  font-size: 0.8125rem;
  color: #999;
}

.process-status {
  padding: 4px 10px;
  background: #e8f5e9;
  color: #2e7d32;
  border-radius: 4px;
  font-size: 0.8125rem;
  font-weight: 500;
  white-space: nowrap;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #999;
  font-size: 0.875rem;
}
</style>
