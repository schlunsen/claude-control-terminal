<template>
  <div class="notification-stats">
    <h2 class="section-title">Notification Insights</h2>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-value">{{ stats.permission_requests }}</div>
        <div class="stat-label">Permission Requests</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats.idle_alerts }}</div>
        <div class="stat-label">Idle Alerts</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats.total_notifications }}</div>
        <div class="stat-label">Total Notifications</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats.most_requested_tool || '—' }}</div>
        <div class="stat-label">Most Requested Tool</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { NotificationStats } from '~/types/analytics'

const stats = ref<NotificationStats>({
  total_notifications: 0,
  permission_requests: 0,
  idle_alerts: 0,
  most_requested_tool: '',
  most_requested_tool_count: 0
})

// WebSocket integration
const { on } = useWebSocket()

// Reload stats on WebSocket events
on('onNotification', loadStats)
on('onHistoryCleared', loadStats)

async function loadStats() {
  try {
    const { data } = await useFetch<NotificationStats>('/api/notifications/stats')
    if (data.value) {
      stats.value = data.value
    }
  } catch (error) {
    // Error loading notification stats
  }
}

// Load on mount
onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.notification-stats {
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
  color: var(--accent-cyan);
  margin-bottom: 4px;
  letter-spacing: -0.02em;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-weight: 400;
}

@media (max-width: 768px) {
  .notification-stats {
    padding: 24px;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  .stat-value {
    font-size: 2rem;
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
