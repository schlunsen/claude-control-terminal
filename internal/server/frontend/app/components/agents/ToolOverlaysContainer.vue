<template>
  <div v-if="tools.length > 0" class="tool-overlays-container">
    <slot></slot>
  </div>
</template>

<script setup lang="ts">
interface Props {
  tools: any[]
}

defineProps<Props>()
</script>

<style scoped>
.tool-overlays-container {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none; /* Allow clicks through container */
}

.tool-overlays-container :deep(> *) {
  pointer-events: auto; /* Re-enable clicks on overlay items */
}

@media (max-width: 768px) {
  .tool-overlays-container {
    left: 16px;
    right: 16px;
    width: calc(100% - 32px);
  }
}
</style>
