<template>
  <div class="dashboard">
    <div class="container">
      <!-- Header -->
      <header>
        <h1>Claude Code Analytics</h1>
        <p class="subtitle">Real-time monitoring and process management</p>
        <div class="status" :class="{ 'status-connected': connected }">
          <div class="status-dot"></div>
          <span>{{ connected ? 'Analytics running' : 'Connecting...' }}</span>
        </div>
      </header>

      <!-- Stats Overview -->
      <section class="section">
        <h2 class="section-title">Overview</h2>
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-value">{{ stats.totalConversations }}</div>
            <div class="stat-label">Conversations</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ formatNumber(stats.totalTokens) }}</div>
            <div class="stat-label">Total Tokens</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ stats.activeConversations }}</div>
            <div class="stat-label">Active Sessions</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ formatNumber(stats.avgTokens) }}</div>
            <div class="stat-label">Avg Tokens</div>
          </div>
        </div>

        <!-- Reset Controls -->
        <ResetControls />
      </section>

      <!-- Claude Processes and Background Shells (side-by-side) -->
      <div class="grid-2">
        <ClaudeProcesses />
        <BackgroundShells />
      </div>

      <!-- Activity History -->
      <ActivityHistory />

      <!-- Notification Stats -->
      <NotificationStats />

      <!-- API Endpoints -->
      <ApiEndpoints />

      <!-- Footer -->
      <footer class="footer">
        Claude Control Terminal
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Stats } from '~/types/analytics'

// State
const stats = ref<Stats>({
  totalConversations: 0,
  totalTokens: 0,
  activeConversations: 0,
  avgTokens: 0,
  timestamp: ''
})

// WebSocket for stats updates
const { connected, on } = useWebSocket()

on('onStatsUpdate', (data: any) => {
  stats.value = { ...stats.value, ...data }
})

on('onReset', () => {
  // Reload stats after reset
  loadStats()
})

// Load initial stats
async function loadStats() {
  try {
    const { data } = await useFetch<Stats>('/api/stats')
    if (data.value) {
      stats.value = data.value
    }
  } catch (error) {
    console.error('Error loading stats:', error)
  }
}

// Helper function
function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

// Load stats on mount
onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.dashboard {
  padding: 40px 20px;
  background: #f5f5f5;
  min-height: 100vh;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
}

header {
  margin-bottom: 60px;
}

header h1 {
  font-size: 2rem;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8px;
  letter-spacing: -0.02em;
}

.subtitle {
  font-size: 0.95rem;
  color: #666;
  font-weight: 400;
}

.status {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  font-size: 0.875rem;
  color: #666;
  margin-top: 16px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ff9800;
  animation: pulse 2s ease-in-out infinite;
}

.status-connected .status-dot {
  background: #4caf50;
  animation: none;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.section {
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 32px;
  margin-bottom: 24px;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24px;
  letter-spacing: -0.01em;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 24px;
}

.stat-card {
  padding: 0;
}

.stat-value {
  font-size: 2.5rem;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 4px;
  letter-spacing: -0.02em;
}

.stat-label {
  font-size: 0.875rem;
  color: #666;
  font-weight: 400;
}

.grid-2 {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
  gap: 24px;
  margin-bottom: 24px;
}

.footer {
  text-align: center;
  margin-top: 60px;
  padding-top: 32px;
  border-top: 1px solid #e0e0e0;
  color: #999;
  font-size: 0.8125rem;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .grid-2 {
    grid-template-columns: 1fr;
  }
}
</style>
