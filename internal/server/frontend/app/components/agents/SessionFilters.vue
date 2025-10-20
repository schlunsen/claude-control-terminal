<template>
  <div class="session-filters">
    <button
      v-for="filter in filters"
      :key="filter.value"
      @click="$emit('update:activeFilter', filter.value)"
      class="filter-tab"
      :class="{ active: activeFilter === filter.value }"
    >
      {{ filter.label }}
      <span class="filter-count">{{ filter.count }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
interface Filter {
  label: string
  value: string
  count: number
}

interface Props {
  activeFilter: string
  filters: Filter[]
}

defineProps<Props>()
defineEmits<{
  (e: 'update:activeFilter', value: string): void
}>()
</script>

<style scoped>
.session-filters {
  display: flex;
  gap: 0.5rem;
  padding: 0 0.75rem;
  margin-bottom: 0.5rem;
}

.filter-tab {
  flex: 1;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
  background: var(--color-bg-primary);
  color: var(--color-text-secondary);
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.filter-tab:hover {
  border-color: var(--color-primary);
  background: var(--color-bg-secondary);
}

.filter-tab.active {
  border-color: var(--color-primary);
  background: var(--color-primary-alpha);
  color: var(--color-primary);
}

.filter-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.25rem;
  height: 1.25rem;
  padding: 0 0.375rem;
  border-radius: 0.625rem;
  background: var(--color-bg-tertiary);
  font-size: 0.75rem;
  font-weight: 600;
}

.filter-tab.active .filter-count {
  background: var(--color-primary);
  color: white;
}
</style>
