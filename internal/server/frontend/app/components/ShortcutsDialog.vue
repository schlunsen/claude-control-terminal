<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="showDialog" class="modal-overlay" @click="closeDialog">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h2>Keyboard Shortcuts</h2>
            <button class="close-button" @click="closeDialog" title="Close (ESC)">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>

          <div class="modal-body">
            <div v-for="(shortcuts, category) in groupedShortcuts" :key="category" class="shortcuts-section">
              <h3 class="category-title">{{ category }}</h3>
              <div class="shortcuts-list">
                <div v-for="shortcut in shortcuts" :key="shortcut.key" class="shortcut-item">
                  <span class="shortcut-description">{{ shortcut.description }}</span>
                  <div class="shortcut-keys">
                    <kbd v-if="shortcut.modifiers.shift" class="key">⇧</kbd>
                    <kbd v-if="shortcut.modifiers.alt" class="key">⌥</kbd>
                    <kbd v-if="shortcut.modifiers.meta" class="key">⌘</kbd>
                    <kbd v-if="shortcut.modifiers.ctrl" class="key">Ctrl</kbd>
                    <kbd class="key key-primary">{{ shortcut.key.toUpperCase() }}</kbd>
                  </div>
                </div>
              </div>
            </div>

            <div class="shortcuts-section">
              <h3 class="category-title">Special</h3>
              <div class="shortcuts-list">
                <div class="shortcut-item">
                  <span class="shortcut-description">Close this dialog</span>
                  <div class="shortcut-keys">
                    <kbd class="key key-primary">ESC</kbd>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <p class="footer-note">
              All shortcuts use <kbd class="key-inline">⇧ Shift</kbd> + <kbd class="key-inline">⌥ Option</kbd> + <kbd class="key-inline">⌘ Command</kbd> on macOS
            </p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
const { showDialog, closeDialog, getAllShortcuts } = useKeyboardShortcuts()

const groupedShortcuts = computed(() => {
  return getAllShortcuts()
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
}

.modal-content {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  max-width: 600px;
  width: 100%;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 28px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.close-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-button:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.modal-body {
  padding: 28px;
  overflow-y: auto;
  flex: 1;
}

.shortcuts-section {
  margin-bottom: 28px;
}

.shortcuts-section:last-child {
  margin-bottom: 0;
}

.category-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--accent-purple);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 12px 0;
}

.shortcuts-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.shortcut-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
}

.shortcut-description {
  font-size: 0.95rem;
  color: var(--text-primary);
  font-weight: 400;
}

.shortcut-keys {
  display: flex;
  gap: 6px;
  align-items: center;
}

.key {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 8px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', monospace;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-secondary);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.key-primary {
  background: var(--accent-purple);
  border-color: var(--accent-purple);
  color: white;
}

.key-inline {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px 6px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 3px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', monospace;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.modal-footer {
  padding: 20px 28px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-secondary);
}

.footer-note {
  margin: 0;
  font-size: 0.8rem;
  color: var(--text-secondary);
  text-align: center;
}

/* Modal transition */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .modal-content,
.modal-leave-active .modal-content {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-content,
.modal-leave-to .modal-content {
  transform: scale(0.95);
  opacity: 0;
}

/* Responsive */
@media (max-width: 640px) {
  .modal-content {
    max-width: 100%;
    max-height: 90vh;
    border-radius: 8px;
  }

  .modal-header {
    padding: 20px;
  }

  .modal-header h2 {
    font-size: 1.25rem;
  }

  .modal-body {
    padding: 20px;
  }

  .shortcuts-section {
    margin-bottom: 24px;
  }

  .shortcut-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    padding: 12px 0;
  }

  .modal-footer {
    padding: 16px 20px;
  }
}
</style>
