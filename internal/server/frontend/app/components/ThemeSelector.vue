<template>
  <div class="theme-selector" ref="selectorRef">
    <button @click="toggleDropdown" class="theme-selector-button" :title="'Theme: ' + currentThemeData.name">
      <div class="theme-icon">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="3"/>
          <path d="M12 1v6m0 6v6m5-11l-4 4m-4 4l-4 4m11-5h-6m-6 0H1m11-5l4-4m4 4l4-4"/>
        </svg>
      </div>
      <span class="theme-name" v-show="showLabel">{{ currentThemeData.name }}</span>
      <svg class="dropdown-arrow" width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="2">
        <polyline points="2,4 6,8 10,4"/>
      </svg>
    </button>

    <Transition name="dropdown">
      <div v-if="isOpen" class="theme-dropdown">
        <div class="dropdown-header">
          <span>Select Theme</span>
        </div>
        <div class="dropdown-content">
          <button
            v-for="theme in availableThemes"
            :key="theme.id"
            @click="selectTheme(theme.id)"
            class="theme-option"
            :class="{ 'theme-option-active': currentTheme === theme.id }"
          >
            <div class="theme-option-content">
              <div class="theme-option-info">
                <span class="theme-option-name">{{ theme.name }}</span>
                <span class="theme-option-desc">{{ theme.description }}</span>
              </div>
              <div class="theme-option-badge" :class="{ 'badge-dark': theme.isDark }">
                {{ theme.isDark ? 'Dark' : 'Light' }}
              </div>
            </div>
            <div class="check-icon" v-if="currentTheme === theme.id">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                <polyline points="20 6 9 17 4 12"></polyline>
              </svg>
            </div>
          </button>
        </div>
        <div class="dropdown-footer">
          <NuxtLink to="/themes" class="manage-link" @click="closeDropdown">
            Manage Themes
          </NuxtLink>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import type { ThemeVariant } from '~/composables/useTheme'

interface Props {
  showLabel?: boolean
}

withDefaults(defineProps<Props>(), {
  showLabel: false
})

const { currentTheme, currentThemeData, availableThemes, setTheme } = useTheme()

const isOpen = ref(false)
const selectorRef = ref<HTMLElement | null>(null)

function toggleDropdown() {
  isOpen.value = !isOpen.value
}

function closeDropdown() {
  isOpen.value = false
}

function selectTheme(themeId: ThemeVariant) {
  setTheme(themeId)
  closeDropdown()
}

// Close dropdown when clicking outside
onMounted(() => {
  const handleClickOutside = (event: MouseEvent) => {
    if (selectorRef.value && !selectorRef.value.contains(event.target as Node)) {
      closeDropdown()
    }
  }

  document.addEventListener('click', handleClickOutside)

  onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside)
  })
})
</script>

<style scoped>
.theme-selector {
  position: relative;
}

.theme-selector-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
}

.theme-selector-button:hover {
  background: var(--card-hover);
  border-color: var(--accent-purple);
  color: var(--text-primary);
}

.theme-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
}

.theme-name {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dropdown-arrow {
  transition: transform 0.2s ease;
}

.theme-selector-button:hover .dropdown-arrow {
  transform: translateY(2px);
}

/* Dropdown */
.theme-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 320px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  z-index: 1000;
  overflow: hidden;
}

.dropdown-header {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.dropdown-content {
  max-height: 400px;
  overflow-y: auto;
  padding: 8px;
}

.theme-option {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-bottom: 4px;
  text-align: left;
}

.theme-option:hover {
  background: var(--bg-secondary);
  border-color: var(--border-color);
}

.theme-option-active {
  background: var(--bg-secondary);
  border-color: var(--accent-purple);
}

.theme-option-content {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.theme-option-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.theme-option-name {
  color: var(--text-primary);
  font-weight: 500;
  font-size: 0.875rem;
}

.theme-option-desc {
  color: var(--text-muted);
  font-size: 0.75rem;
}

.theme-option-badge {
  padding: 4px 10px;
  background: var(--code-bg);
  border-radius: 12px;
  font-size: 0.7rem;
  font-weight: 500;
  color: var(--accent-cyan);
  white-space: nowrap;
}

.badge-dark {
  color: var(--accent-purple);
}

.check-icon {
  color: var(--accent-purple);
  display: flex;
  align-items: center;
  margin-left: 8px;
}

.dropdown-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: center;
}

.manage-link {
  color: var(--accent-purple);
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
  transition: color 0.2s ease;
}

.manage-link:hover {
  color: var(--accent-cyan);
}

/* Dropdown animation */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* Scrollbar for dropdown */
.dropdown-content::-webkit-scrollbar {
  width: 6px;
}

.dropdown-content::-webkit-scrollbar-track {
  background: var(--bg-secondary);
  border-radius: 3px;
}

.dropdown-content::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
}

.dropdown-content::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

/* Responsive */
@media (max-width: 640px) {
  .theme-dropdown {
    right: -8px;
    min-width: 280px;
  }
}
</style>
