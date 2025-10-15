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
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 32px;
  transition: all 0.3s ease;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
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
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  transition: all 0.2s ease;
}

.process-item:hover {
  border-color: var(--accent-purple);
  background: var(--card-hover);
}

.process-info {
  flex: 1;
}

.process-pid {
  font-family: 'SF Mono', Monaco, 'Consolas', 'Courier New', monospace;
  font-size: 0.875rem;
  color: var(--accent-cyan);
  font-weight: 600;
  margin-bottom: 4px;
}

.process-command {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-bottom: 4px;
  word-break: break-all;
}

.process-dir {
  font-size: 0.8125rem;
  color: var(--text-muted);
}

.process-status {
  padding: 4px 10px;
  background: var(--status-success);
  color: var(--bg-primary);
  border-radius: 4px;
  font-size: 0.8125rem;
  font-weight: 600;
  white-space: nowrap;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-muted);
  font-size: 0.875rem;
}

@media (max-width: 768px) {
  .claude-processes {
    padding: 24px;
  }

  .process-item {
    padding: 12px;
  }
}
</style>
