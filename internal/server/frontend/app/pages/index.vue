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

      <!-- Main Dashboard Grid -->
      <div class="dashboard-grid">
        <!-- Left Column -->
        <div class="dashboard-column">
          <ActivityHistory />
          <NotificationStats />
        </div>
        
        <!-- Right Column -->
        <div class="dashboard-column">
          <!-- Claude Processes and Background Shells -->
          <div class="processes-grid">
            <ClaudeProcesses />
            <BackgroundShells />
          </div>
        </div>
      </div>

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
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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
}

/* Dashboard Layout */
.dashboard-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  margin-bottom: 24px;
}

.dashboard-column {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.processes-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 24px;
}

.footer {
  text-align: center;
  margin-top: 60px;
  padding-top: 32px;
  border-top: 1px solid var(--border-color);
  color: var(--text-muted);
  font-size: 0.8125rem;
}

@media (max-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
  
  .processes-grid {
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }
}

@media (max-width: 768px) {
  .dashboard {
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

  .dashboard-grid {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  
  .processes-grid {
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
