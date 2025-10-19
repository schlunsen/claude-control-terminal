<template>
  <div class="themes-page">
    <div class="container">
      <!-- Header -->
      <header>
        <h1>Theme Settings</h1>
        <p class="subtitle">Customize the look and feel of your dashboard</p>
      </header>

      <!-- Current Theme Display -->
      <section class="section current-theme-section">
        <h2 class="section-title">Current Theme</h2>
        <div class="current-theme-display">
          <div class="theme-preview" :data-theme="currentTheme">
            <div class="preview-header">
              <div class="preview-dot dot-1"></div>
              <div class="preview-dot dot-2"></div>
              <div class="preview-dot dot-3"></div>
            </div>
            <div class="preview-body">
              <div class="preview-card"></div>
              <div class="preview-card small"></div>
            </div>
          </div>
          <div class="theme-info">
            <h3>{{ currentThemeData.name }}</h3>
            <p>{{ currentThemeData.description }}</p>
            <div class="theme-badges">
              <div class="theme-badge" :class="{ 'badge-dark': currentThemeData.isDark }">
                {{ currentThemeData.isDark ? 'Dark Mode' : 'Light Mode' }}
              </div>
              <div class="font-badge">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="4 7 4 4 20 4 20 7"/>
                  <line x1="9" y1="20" x2="15" y2="20"/>
                  <line x1="12" y1="4" x2="12" y2="20"/>
                </svg>
                {{ currentThemeData.fontFamily }}
              </div>
            </div>
            <p class="font-description">{{ currentThemeData.fontDescription }}</p>
          </div>
        </div>
      </section>

      <!-- Theme Selector -->
      <section class="section">
        <div class="carousel-header">
          <h2 class="section-title">Available Themes</h2>
          <div class="carousel-controls">
            <button
              @click="scrollLeft"
              class="carousel-arrow"
              :class="{ 'disabled': scrollPosition <= 0 }"
              :disabled="scrollPosition <= 0"
            >
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="15 18 9 12 15 6"/>
              </svg>
            </button>
            <button
              @click="scrollRight"
              class="carousel-arrow"
              :class="{ 'disabled': scrollPosition >= maxScroll }"
              :disabled="scrollPosition >= maxScroll"
            >
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="9 18 15 12 9 6"/>
              </svg>
            </button>
          </div>
        </div>
        <div class="themes-carousel-wrapper">
          <div
            class="themes-carousel"
            ref="carouselRef"
            @scroll="updateScrollPosition"
          >
            <div
              v-for="theme in availableThemes"
              :key="theme.id"
              class="theme-card"
              :class="{ 'theme-card-active': currentTheme === theme.id }"
              @click="selectTheme(theme.id)"
            >
              <div class="theme-card-preview" :data-theme="theme.id">
                <div class="preview-header">
                  <div class="preview-dot dot-1"></div>
                  <div class="preview-dot dot-2"></div>
                  <div class="preview-dot dot-3"></div>
                </div>
                <div class="preview-body">
                  <div class="preview-card"></div>
                  <div class="preview-card small"></div>
                </div>
              </div>
              <div class="theme-card-content">
                <div class="theme-card-header">
                  <h3>{{ theme.name }}</h3>
                  <div class="check-icon" v-if="currentTheme === theme.id">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                      <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                  </div>
                </div>
                <p>{{ theme.description }}</p>
                <div class="theme-card-footer">
                  <span class="theme-mode-badge" :class="{ 'badge-dark': theme.isDark }">
                    {{ theme.isDark ? 'Dark' : 'Light' }}
                  </span>
                  <span class="theme-font-badge" :title="theme.fontDescription">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="4 7 4 4 20 4 20 7"/>
                      <line x1="9" y1="20" x2="15" y2="20"/>
                      <line x1="12" y1="4" x2="12" y2="20"/>
                    </svg>
                    {{ theme.fontFamily }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Color Palette Preview -->
      <section class="section">
        <h2 class="section-title">Color Palette</h2>
        <div class="palette-grid">
          <div class="palette-item">
            <div class="palette-color" style="background: var(--accent-purple)"></div>
            <span>Purple</span>
          </div>
          <div class="palette-item">
            <div class="palette-color" style="background: var(--accent-cyan)"></div>
            <span>Cyan</span>
          </div>
          <div class="palette-item">
            <div class="palette-color" style="background: var(--accent-green)"></div>
            <span>Green</span>
          </div>
          <div class="palette-item">
            <div class="palette-color" style="background: var(--accent-yellow)"></div>
            <span>Yellow</span>
          </div>
          <div class="palette-item">
            <div class="palette-color" style="background: var(--accent-orange)"></div>
            <span>Orange</span>
          </div>
        </div>
      </section>

      <!-- Footer -->
      <footer class="footer">
        Claude Control Terminal - Theme preferences are saved locally
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ThemeVariant } from '~/composables/useTheme'

const { currentTheme, currentThemeData, availableThemes, setTheme } = useTheme()

// Carousel refs and state
const carouselRef = ref<HTMLElement | null>(null)
const scrollPosition = ref(0)
const maxScroll = ref(0)

function selectTheme(themeId: ThemeVariant) {
  setTheme(themeId)
}

function scrollLeft() {
  if (carouselRef.value) {
    const scrollAmount = 650 // Scroll ~2 cards at a time for faster navigation
    carouselRef.value.scrollBy({
      left: -scrollAmount,
      behavior: 'smooth'
    })
  }
}

function scrollRight() {
  if (carouselRef.value) {
    const scrollAmount = 650 // Scroll ~2 cards at a time for faster navigation
    carouselRef.value.scrollBy({
      left: scrollAmount,
      behavior: 'smooth'
    })
  }
}

function updateScrollPosition() {
  if (carouselRef.value) {
    scrollPosition.value = carouselRef.value.scrollLeft
    maxScroll.value = carouselRef.value.scrollWidth - carouselRef.value.clientWidth
  }
}

// Initialize scroll state on mount
onMounted(() => {
  updateScrollPosition()
  window.addEventListener('resize', updateScrollPosition)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateScrollPosition)
})
</script>

<style scoped>
.themes-page {
  padding: 20px;
  background: var(--bg-primary);
  min-height: calc(100vh - 60px);
  transition: background-color 0.3s ease;
}

.container {
  width: 100%;
  max-width: none;
  margin: 0;
}

header {
  margin-bottom: 40px;
}

header h1 {
  font-size: 2rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: -0.02em;
}

.subtitle {
  font-size: 0.95rem;
  color: var(--text-secondary);
  font-weight: 400;
  margin-top: 8px;
}

.section {
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

/* Current Theme Display */
.current-theme-display {
  display: flex;
  gap: 32px;
  align-items: center;
}

.theme-preview {
  width: 240px;
  height: 160px;
  border-radius: 12px;
  overflow: hidden;
  border: 2px solid var(--border-color);
  flex-shrink: 0;
  transition: transform 0.2s ease;
}

.theme-preview:hover {
  transform: scale(1.05);
}

.theme-info {
  flex: 1;
}

.theme-info h3 {
  font-size: 1.5rem;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.theme-info p {
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.theme-badges {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.theme-badge {
  display: inline-block;
  padding: 6px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--accent-cyan);
}

.badge-dark {
  color: var(--accent-purple);
}

.font-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  background: var(--code-bg);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--accent-green);
}

.font-description {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin-bottom: 0;
}

/* Carousel Header */
.carousel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.carousel-controls {
  display: flex;
  gap: 12px;
}

.carousel-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: var(--card-bg);
  border: 2px solid var(--border-color);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  color: var(--text-primary);
}

.carousel-arrow:hover:not(.disabled) {
  background: var(--accent-purple);
  border-color: var(--accent-purple);
  color: white;
  transform: scale(1.1);
}

.carousel-arrow:active:not(.disabled) {
  transform: scale(0.95);
}

.carousel-arrow.disabled {
  opacity: 0.3;
  cursor: not-allowed;
  border-color: var(--border-color);
}

/* Themes Carousel */
.themes-carousel-wrapper {
  position: relative;
  overflow: hidden;
}

.themes-carousel {
  display: flex;
  gap: 24px;
  overflow-x: auto;
  scroll-behavior: smooth;
  padding: 4px 4px 24px 4px;
  -webkit-overflow-scrolling: touch;
}

/* Hide scrollbar but keep functionality */
.themes-carousel::-webkit-scrollbar {
  height: 8px;
}

.themes-carousel::-webkit-scrollbar-track {
  background: var(--bg-secondary);
  border-radius: 4px;
}

.themes-carousel::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 4px;
}

.themes-carousel::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

.theme-card {
  background: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  flex-shrink: 0;
  width: 300px;
}

.theme-card:hover {
  transform: translateY(-4px);
  border-color: var(--accent-purple);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
}

.theme-card-active {
  border-color: var(--accent-purple);
  background: var(--card-bg);
  box-shadow: 0 0 0 2px var(--accent-purple);
}

.theme-card-preview {
  width: 100%;
  height: 140px;
  transition: all 0.3s ease;
}

.theme-card-content {
  padding: 20px;
}

.theme-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.theme-card-header h3 {
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 0;
}

.check-icon {
  color: var(--accent-purple);
  display: flex;
  align-items: center;
  animation: checkPop 0.3s ease;
}

@keyframes checkPop {
  0% {
    transform: scale(0);
  }
  50% {
    transform: scale(1.2);
  }
  100% {
    transform: scale(1);
  }
}

.theme-card-content p {
  color: var(--text-secondary);
  font-size: 0.875rem;
  margin-bottom: 12px;
}

.theme-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.theme-mode-badge {
  padding: 4px 12px;
  background: var(--code-bg);
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--accent-cyan);
}

.theme-font-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: var(--bg-secondary);
  border-radius: 12px;
  font-size: 0.7rem;
  font-weight: 500;
  color: var(--text-muted);
  cursor: help;
}

.theme-font-badge:hover {
  color: var(--accent-green);
  background: var(--code-bg);
}

/* Theme Preview Styling */
.preview-header {
  background: var(--bg-secondary);
  height: 40px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 8px;
  border-bottom: 1px solid var(--border-color);
}

.preview-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.dot-1 {
  background: var(--status-error);
}

.dot-2 {
  background: var(--status-warning);
}

.dot-3 {
  background: var(--status-success);
}

.preview-body {
  background: var(--bg-primary);
  height: calc(100% - 40px);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.preview-card {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  height: 40px;
}

.preview-card.small {
  height: 24px;
  width: 60%;
}

/* Color Palette */
.palette-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 20px;
}

.palette-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.palette-color {
  width: 80px;
  height: 80px;
  border-radius: 12px;
  border: 2px solid var(--border-color);
  transition: transform 0.2s ease;
  cursor: pointer;
}

.palette-color:hover {
  transform: scale(1.1);
}

.palette-item span {
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
}

.footer {
  text-align: center;
  margin-top: 60px;
  padding-top: 32px;
  border-top: 1px solid var(--border-color);
  color: var(--text-muted);
  font-size: 0.8125rem;
}

/* Responsive Design */
@media (max-width: 768px) {
  .themes-page {
    padding: 15px;
  }

  .current-theme-display {
    flex-direction: column;
    align-items: flex-start;
  }

  .theme-preview {
    width: 100%;
    max-width: 300px;
  }

  .themes-grid {
    grid-template-columns: 1fr;
  }

  .section {
    padding: 24px;
  }
}

@media (max-width: 480px) {
  header h1 {
    font-size: 1.5rem;
  }

  .palette-grid {
    grid-template-columns: repeat(3, 1fr);
  }

  .palette-color {
    width: 60px;
    height: 60px;
  }
}
</style>
