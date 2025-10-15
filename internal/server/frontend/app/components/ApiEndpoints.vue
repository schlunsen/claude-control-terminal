<template>
  <div class="api-endpoints">
    <h2 class="section-title">API Endpoints</h2>

    <div class="endpoints">
      <div class="endpoint">
        GET <code>/api/data</code> — Complete analytics data
      </div>
      <div class="endpoint">
        GET <code>/api/conversations</code> — All conversations
      </div>
      <div class="endpoint">
        GET <code>/api/processes</code> — Running Claude processes
      </div>
      <div class="endpoint">
        GET <code>/api/shells</code> — Background bash shells
      </div>
      <div class="endpoint">
        GET <code>/api/stats</code> — Statistics summary
      </div>
      <div class="endpoint">
        GET <code>/api/history/shell</code> — Shell command history
      </div>
      <div class="endpoint">
        GET <code>/api/history/claude</code> — Claude command history
      </div>
      <div class="endpoint">
        GET <code>/api/history/stats</code> — Command statistics
      </div>
      <div class="endpoint">
        GET <code>/api/prompts</code> — User prompts history
      </div>
      <div class="endpoint">
        GET <code>/api/prompts/stats</code> — Prompt statistics
      </div>
      <div class="endpoint">
        GET <code>/api/notifications</code> — Notifications history
      </div>
      <div class="endpoint">
        GET <code>/api/notifications/stats</code> — Notification statistics
      </div>
      <div class="endpoint">
        GET <code>/ws</code> — WebSocket real-time updates
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// No script needed - this is a static display component
</script>

<style scoped>
.api-endpoints {
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

.endpoints {
  display: grid;
  gap: 8px;
}

.endpoint {
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-family: 'SF Mono', Monaco, 'Consolas', 'Courier New', monospace;
  font-size: 0.8125rem;
  color: var(--text-secondary);
  transition: all 0.2s ease;
}

.endpoint:hover {
  border-color: var(--accent-purple);
  background: var(--card-hover);
}

.endpoint code {
  color: var(--accent-cyan);
  font-weight: 600;
  background: none;
  padding: 0;
}

@media (max-width: 768px) {
  .api-endpoints {
    padding: 24px;
  }

  .endpoint {
    padding: 10px 12px;
    font-size: 0.75rem;
  }
}
</style>
