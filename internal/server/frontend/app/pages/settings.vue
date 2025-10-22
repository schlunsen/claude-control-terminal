<template>
  <div class="settings-page">
    <div class="container">
      <!-- Header -->
      <header>
        <h1>Settings</h1>
        <p class="subtitle">Configure your Claude Control Terminal experience</p>
      </header>

      <!-- Agent Behavior Section -->
      <section class="section">
        <h2 class="section-title">Agent Behavior</h2>
        <div class="settings-group">
          <!-- Diff Display Location -->
          <div class="setting-item">
            <div class="setting-info">
              <h3 class="setting-title">Diff Display Location</h3>
              <p class="setting-description">
                Choose where to display file edit diffs when using the Edit tool.
                "In Chat" shows the full diff in the conversation, "In Options" shows diffs
                in a collapsible overlay.
              </p>
            </div>
            <div class="setting-control">
              <div class="radio-group">
                <label class="radio-option" :class="{ 'active': diffDisplayLocation === 'chat' }">
                  <input
                    type="radio"
                    name="diff-location"
                    value="chat"
                    v-model="diffDisplayLocation"
                    @change="saveDiffDisplayLocation"
                  />
                  <div class="radio-content">
                    <div class="radio-icon">
                      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                      </svg>
                    </div>
                    <div class="radio-text">
                      <strong>In Chat</strong>
                      <span>Show full diff in conversation</span>
                    </div>
                  </div>
                </label>
                <label class="radio-option" :class="{ 'active': diffDisplayLocation === 'options' }">
                  <input
                    type="radio"
                    name="diff-location"
                    value="options"
                    v-model="diffDisplayLocation"
                    @change="saveDiffDisplayLocation"
                  />
                  <div class="radio-content">
                    <div class="radio-icon">
                      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                        <line x1="9" y1="9" x2="15" y2="9"></line>
                        <line x1="9" y1="15" x2="15" y2="15"></line>
                      </svg>
                    </div>
                    <div class="radio-text">
                      <strong>In Options</strong>
                      <span>Show diff in collapsible overlay</span>
                    </div>
                  </div>
                </label>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Save Status -->
      <div class="save-status" :class="{ 'visible': showSaveStatus }">
        <div class="save-status-content">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
          <span>Settings saved successfully</span>
        </div>
        <button @click="showSaveStatus = false" class="save-status-close" title="Close">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

// Use authenticated fetch composable for API calls with auth
const { fetchWithAuth } = useAuthenticatedFetch()

// Settings state
const diffDisplayLocation = ref('chat')
const showSaveStatus = ref(false)

// Fetch current settings from API
const fetchSettings = async () => {
  try {
    const response = await fetchWithAuth('/api/settings/diff_display_location', {
      method: 'GET',
    })

    if (response.ok) {
      const setting = await response.json()
      diffDisplayLocation.value = setting.value || 'chat'
    }
  } catch (error) {
    console.error('Failed to fetch settings:', error)
    // Default to 'chat' if fetch fails
    diffDisplayLocation.value = 'chat'
  }
}

// Save diff display location setting
const saveDiffDisplayLocation = async () => {
  try {
    const response = await fetchWithAuth('/api/settings/diff_display_location', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        value: diffDisplayLocation.value,
        value_type: 'string',
        description: 'Where to display file diffs: "chat" or "options"',
      }),
    })

    if (response.ok) {
      // Show save status
      showSaveStatus.value = true
      setTimeout(() => {
        showSaveStatus.value = false
      }, 2000)
    } else {
      console.error('Failed to save setting:', await response.text())
    }
  } catch (error) {
    console.error('Error saving setting:', error)
  }
}

// Load settings on mount
onMounted(() => {
  fetchSettings()
})
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
  background: var(--bg-primary);
  padding: 40px 20px;
}

.container {
  max-width: 900px;
  margin: 0 auto;
}

header {
  margin-bottom: 40px;
}

header h1 {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 8px 0;
}

.subtitle {
  color: var(--text-secondary);
  font-size: 1.1rem;
  margin: 0;
}

.section {
  background: var(--card-bg);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  border: 1px solid var(--border-color);
}

.section-title {
  font-size: 1.3rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 20px 0;
}

.settings-group {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.setting-item {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.setting-info {
  flex: 1;
}

.setting-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 8px 0;
}

.setting-description {
  color: var(--text-secondary);
  font-size: 0.95rem;
  line-height: 1.6;
  margin: 0;
}

.setting-control {
  margin-top: 8px;
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.radio-option {
  display: block;
  position: relative;
  cursor: pointer;
}

.radio-option input[type="radio"] {
  position: absolute;
  opacity: 0;
  cursor: pointer;
}

.radio-content {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  transition: all 0.2s;
}

.radio-option:hover .radio-content {
  border-color: var(--accent-purple);
  background: var(--bg-hover);
}

.radio-option.active .radio-content {
  border-color: var(--accent-purple);
  background: var(--bg-active);
}

.radio-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: var(--bg-primary);
  border-radius: 8px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.radio-option.active .radio-icon {
  color: var(--accent-purple);
  background: var(--accent-purple-bg);
}

.radio-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.radio-text strong {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 1rem;
}

.radio-text span {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.save-status {
  position: fixed;
  bottom: 24px;
  right: 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px;
  background: var(--accent-green);
  color: white;
  border-radius: 8px;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  opacity: 0;
  transform: translateY(20px);
  transition: all 0.3s;
  pointer-events: none;
  min-width: 280px;
}

.save-status.visible {
  opacity: 1;
  transform: translateY(0);
  pointer-events: auto;
}

.save-status-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.save-status-content svg {
  flex-shrink: 0;
}

.save-status-close {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  background: transparent;
  border: none;
  color: white;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
  flex-shrink: 0;
}

.save-status-close:hover {
  background: rgba(255, 255, 255, 0.2);
}

.save-status-close svg {
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .settings-page {
    padding: 20px 16px;
  }

  header h1 {
    font-size: 1.5rem;
  }

  .subtitle {
    font-size: 1rem;
  }

  .section {
    padding: 20px 16px;
  }
}
</style>
