<template>
  <div class="background-shells">
    <h2 class="section-title">Background Shells</h2>

    <div v-if="shells.length === 0" class="empty-state">
      No background shells running
    </div>

    <div v-else class="shell-list">
      <div
        v-for="shell in shells"
        :key="shell.shell_id || shell.pid"
        class="shell-item"
      >
        <div class="shell-info">
          <div class="shell-id">
            {{ shell.shell_id ? 'Shell #' + shell.shell_id : 'PID ' + shell.pid }}
          </div>
          <div class="shell-command">{{ shell.command }}</div>
          <div class="shell-dir">{{ shell.working_dir || 'Unknown directory' }}</div>
        </div>
        <div class="shell-status">{{ shell.status }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Shell } from '~/types/analytics'

const shells = ref<Shell[]>([])

// WebSocket integration
const { on } = useWebSocket()

// Reload shells on WebSocket events
on('onStatsUpdate', loadShells)
on('onReset', loadShells)

async function loadShells() {
  try {
    const { data } = await useFetch<any>('/api/shells')
    if (data.value?.shells) {
      shells.value = data.value.shells
    }
  } catch (error) {
    console.error('Error loading background shells:', error)
  }
}

// Load on mount
onMounted(() => {
  loadShells()
})
</script>

<style scoped>
.background-shells {
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

.shell-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.shell-item {
  padding: 16px;
  background: #f0f4ff;
  border: 1px solid #d0dcf0;
  border-radius: 6px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  transition: border-color 0.2s;
}

.shell-item:hover {
  border-color: #b0c4e0;
}

.shell-info {
  flex: 1;
}

.shell-id {
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.875rem;
  color: #1a1a1a;
  font-weight: 500;
  margin-bottom: 4px;
}

.shell-command {
  font-size: 0.875rem;
  color: #666;
  margin-bottom: 4px;
  word-break: break-all;
}

.shell-dir {
  font-size: 0.8125rem;
  color: #999;
}

.shell-status {
  padding: 4px 10px;
  background: #e3f2fd;
  color: #1565c0;
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
