<template>
  <aside v-if="show" class="metrics-sidebar">
    <SessionMetrics
      :session="session"
      :tool-executions="toolExecutions"
      :permission-stats="permissionStats"
    />
  </aside>
</template>

<script setup lang="ts">
import SessionMetrics from '~/components/SessionMetrics.vue'

interface Props {
  show: boolean
  session: any
  toolExecutions: any
  permissionStats: any
}

defineProps<Props>()
</script>

<style scoped>
.metrics-sidebar {
  width: 320px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow-y: auto;
  min-height: 0;
  flex-shrink: 0;
}

@media (max-width: 1200px) {
  .metrics-sidebar {
    width: 280px;
  }
}

@media (max-width: 768px) {
  .metrics-sidebar {
    display: none;
  }
}
</style>
