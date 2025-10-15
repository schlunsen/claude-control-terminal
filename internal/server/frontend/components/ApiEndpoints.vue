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

.endpoints {
  display: grid;
  gap: 8px;
}

.endpoint {
  padding: 12px 16px;
  background: #fafafa;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.8125rem;
  color: #1a1a1a;
}

.endpoint code {
  color: #0066cc;
  font-weight: 500;
}
</style>
