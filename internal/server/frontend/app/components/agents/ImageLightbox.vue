<template>
  <Teleport to="body">
    <Transition name="lightbox">
      <div v-if="isOpen" class="lightbox-overlay" @click="close">
        <div class="lightbox-container" @click.stop>
          <!-- Close button -->
          <button @click="close" class="close-btn" aria-label="Close">
            ×
          </button>

          <!-- Image display -->
          <div class="image-wrapper">
            <img
              :src="currentImage.dataUrl"
              :alt="`Image ${currentIndex + 1}`"
              class="lightbox-image"
              :style="imageStyle"
            />
          </div>

          <!-- Controls -->
          <div class="lightbox-controls">
            <!-- Zoom controls -->
            <div class="zoom-controls">
              <button @click="zoomOut" :disabled="zoom <= 0.5" aria-label="Zoom out">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="11" cy="11" r="8"></circle>
                  <line x1="8" y1="11" x2="14" y2="11"></line>
                  <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                </svg>
              </button>
              <span class="zoom-level">{{ Math.round(zoom * 100) }}%</span>
              <button @click="zoomIn" :disabled="zoom >= 3" aria-label="Zoom in">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="11" cy="11" r="8"></circle>
                  <line x1="11" y1="8" x2="11" y2="14"></line>
                  <line x1="8" y1="11" x2="14" y2="11"></line>
                  <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                </svg>
              </button>
              <button @click="resetZoom" aria-label="Reset zoom">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M1 4v6h6"></path>
                  <path d="M23 20v-6h-6"></path>
                  <path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15"></path>
                </svg>
              </button>
            </div>

            <!-- Navigation (if multiple images) -->
            <div v-if="images.length > 1" class="nav-controls">
              <button @click="prevImage" :disabled="currentIndex === 0" aria-label="Previous image">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="15 18 9 12 15 6"></polyline>
                </svg>
              </button>
              <span class="image-counter">{{ currentIndex + 1 }} / {{ images.length }}</span>
              <button @click="nextImage" :disabled="currentIndex === images.length - 1" aria-label="Next image">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="9 18 15 12 9 6"></polyline>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

interface ImageData {
  dataUrl: string
  mediaType: string
}

interface Props {
  images: ImageData[]
  startIndex?: number
  isOpen: boolean
}

const props = withDefaults(defineProps<Props>(), {
  startIndex: 0
})

const emit = defineEmits<{
  close: []
}>()

const currentIndex = ref(props.startIndex)
const zoom = ref(1)

const currentImage = computed(() => props.images[currentIndex.value] || { dataUrl: '', mediaType: '' })
const imageStyle = computed(() => ({
  transform: `scale(${zoom.value})`,
  cursor: zoom.value > 1 ? 'move' : 'default'
}))

// Reset index when opening with new images
watch(() => props.isOpen, (newValue) => {
  if (newValue) {
    currentIndex.value = props.startIndex
    zoom.value = 1
  }
})

function close() {
  zoom.value = 1
  emit('close')
}

function zoomIn() {
  zoom.value = Math.min(3, zoom.value + 0.25)
}

function zoomOut() {
  zoom.value = Math.max(0.5, zoom.value - 0.25)
}

function resetZoom() {
  zoom.value = 1
}

function nextImage() {
  if (currentIndex.value < props.images.length - 1) {
    currentIndex.value++
    zoom.value = 1
  }
}

function prevImage() {
  if (currentIndex.value > 0) {
    currentIndex.value--
    zoom.value = 1
  }
}

// Keyboard shortcuts
function handleKeydown(e: KeyboardEvent) {
  if (!props.isOpen) return

  switch (e.key) {
    case 'Escape':
      close()
      break
    case 'ArrowLeft':
      prevImage()
      break
    case 'ArrowRight':
      nextImage()
      break
    case '+':
    case '=':
      zoomIn()
      break
    case '-':
    case '_':
      zoomOut()
      break
    case '0':
      resetZoom()
      break
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.lightbox-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.95);
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.lightbox-container {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px 100px;
}

.close-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 48px;
  height: 48px;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border: none;
  border-radius: 50%;
  font-size: 32px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  z-index: 10001;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: scale(1.1);
}

.image-wrapper {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  width: 100%;
}

.lightbox-image {
  max-width: 90vw;
  max-height: 80vh;
  object-fit: contain;
  transition: transform 0.2s ease-out;
}

.lightbox-controls {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 24px;
  align-items: center;
  background: rgba(0, 0, 0, 0.8);
  padding: 16px 24px;
  border-radius: 12px;
  backdrop-filter: blur(10px);
}

.zoom-controls,
.nav-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.zoom-controls button,
.nav-controls button {
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.zoom-controls button:hover:not(:disabled),
.nav-controls button:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.2);
  transform: scale(1.05);
}

.zoom-controls button:disabled,
.nav-controls button:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.zoom-level,
.image-counter {
  color: white;
  font-size: 0.9rem;
  min-width: 60px;
  text-align: center;
}

/* Transitions */
.lightbox-enter-active,
.lightbox-leave-active {
  transition: opacity 0.3s ease;
}

.lightbox-enter-from,
.lightbox-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .lightbox-container {
    padding: 40px 10px 120px;
  }

  .lightbox-controls {
    flex-direction: column;
    gap: 16px;
    bottom: 10px;
    padding: 12px 16px;
  }

  .close-btn {
    width: 40px;
    height: 40px;
    font-size: 28px;
  }

  .lightbox-image {
    max-width: 95vw;
    max-height: 70vh;
  }
}
</style>
