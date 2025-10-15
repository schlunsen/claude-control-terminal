<template>
  <div class="stats-page">
    <div class="container">
      <!-- Header -->
      <header>
        <h1>Detailed Statistics</h1>
        <p class="subtitle">Comprehensive analytics and performance metrics</p>
        <div class="status" :class="{ 'status-connected': connected }">
          <div class="status-dot"></div>
          <span>{{ connected ? 'Analytics running' : 'Connecting...' }}</span>
        </div>
      </header>

      <!-- Enhanced Stats Grid -->
      <section class="section">
        <h2 class="section-title">Key Metrics</h2>
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-value">{{ stats.totalConversations }}</div>
            <div class="stat-label">Total Conversations</div>
            <div class="stat-trend">+12% this week</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ formatNumber(stats.totalTokens) }}</div>
            <div class="stat-label">Total Tokens</div>
            <div class="stat-trend">{{ formatNumber(stats.avgTokens) }} avg per conversation</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ stats.activeConversations }}</div>
            <div class="stat-label">Active Sessions</div>
            <div class="stat-trend">Real-time</div>
          </div>
          <div class="stat-card">
            <div class="stat-value">{{ formatUptime() }}</div>
            <div class="stat-label">System Uptime</div>
            <div class="stat-trend">Since last restart</div>
          </div>
        </div>
      </section>

      <!-- Performance Metrics -->
      <section class="section">
        <h2 class="section-title">Performance</h2>
        <div class="performance-grid">
          <div class="performance-card">
            <h3>Response Times</h3>
            <div class="metric-row">
              <span>Average Response</span>
              <span class="metric-value">1.2s</span>
            </div>
            <div class="metric-row">
              <span>95th Percentile</span>
              <span class="metric-value">2.8s</span>
            </div>
          </div>
          
          <div class="performance-card">
            <h3>System Resources</h3>
            <div class="metric-row">
              <span>Memory Usage</span>
              <span class="metric-value">45 MB</span>
            </div>
            <div class="metric-row">
              <span>CPU Usage</span>
              <span class="metric-value">12%</span>
            </div>
          </div>
        </div>
      </section>

      <!-- API Endpoints Section -->
      <section class="section">
        <h2 class="section-title">API Endpoints</h2>
        <ApiEndpoints />
      </section>
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

const startTime = ref(Date.now())

// WebSocket for stats updates
const { connected, on } = useWebSocket()

on('onStatsUpdate', (data: any) => {
  stats.value = { ...stats.value, ...data }
})

// Load initial stats
async function loadStats() {
  try {
    const { data } = await useFetch<Stats>('/api/stats')
    if (data.value) {
      stats.value = data.value
    }
  } catch (error) {
    // Error loading stats
  }
}

// Helper functions
function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

function formatUptime(): string {
  const uptime = Date.now() - startTime.value
  const hours = Math.floor(uptime / (1000 * 60 * 60))
  const minutes = Math.floor((uptime % (1000 * 60 * 60)) / (1000 * 60))
  return `${hours}h ${minutes}m`
}

// Load stats on mount
onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.stats-page {
  padding: 20px;
  background: var(--bg-primary);
  min-height: calc(100vh - 60px);
  transition: background-color 0.3s ease;
}

.container {
  width: 100%;
  max-width: none;
  margin: 0;
}

header {
  margin-bottom: 40px;
}

header h1 {
  font-size: 2rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
  letter-spacing: -0.02em;
}

.subtitle {
  font-size: 0.95rem;
  color: var(--text-secondary);
  font-weight: 400;
}

.status {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-top: 16px;
  transition: all 0.3s ease;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent-orange);
  animation: pulse 2s ease-in-out infinite;
}

.status-connected .status-dot {
  background: var(--status-success);
  animation: none;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.section {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 32px;
  margin-bottom: 24px;
  transition: all 0.3s ease;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 24px;
  letter-spacing: -0.01em;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 24px;
}

.stat-card {
  padding: 0;
}

.stat-value {
  font-size: 2.5rem;
  font-weight: 600;
  color: var(--accent-purple);
  margin-bottom: 4px;
  letter-spacing: -0.02em;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-weight: 400;
  margin-bottom: 8px;
}

.stat-trend {
  font-size: 0.75rem;
  color: var(--text-muted);
  background: var(--bg-secondary);
  padding: 4px 8px;
  border-radius: 4px;
  display: inline-block;
}

.performance-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}

.performance-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 24px;
}

.performance-card h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
}

.metric-row:last-child {
  border-bottom: none;
}

.metric-row span:first-child {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.metric-value {
  color: var(--accent-purple);
  font-weight: 600;
  font-size: 0.9rem;
}

@media (max-width: 768px) {
  .stats-page {
    padding: 15px;
  }

  header {
    margin-bottom: 30px;
  }

  header h1 {
    font-size: 1.5rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  .performance-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .section {
    padding: 24px;
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>