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
    console.error('Error loading notification stats:', error)
  }
}

// Load on mount
onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.notification-stats {
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

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
