<template>
  <Transition name="lightbox-fade">
    <div v-if="isOpen" class="lightbox-overlay" @click.self="close">
      <button class="lightbox-close" @click="close" aria-label="Close lightbox">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>

      <button class="lightbox-nav lightbox-nav-prev" @click="prev" aria-label="Previous image">
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="15 18 9 12 15 6"></polyline>
        </svg>
      </button>

      <button class="lightbox-nav lightbox-nav-next" @click="next" aria-label="Next image">
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </button>

      <div class="lightbox-content">
        <Transition name="image-fade" mode="out-in">
          <img
            :key="currentIndex"
            :src="images[currentIndex].src"
            :alt="images[currentIndex].alt"
            class="lightbox-image"
          />
        </Transition>

        <div class="lightbox-info">
          <p class="lightbox-caption">{{ images[currentIndex].caption }}</p>
          <p class="lightbox-counter">{{ currentIndex + 1 }} / {{ images.length }}</p>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
interface Image {
  src: string
  alt: string
  caption: string
}

interface Props {
  images: Image[]
  initialIndex?: number
  isOpen: boolean
}

const props = withDefaults(defineProps<Props>(), {
  initialIndex: 0
})

const emit = defineEmits<{
  close: []
}>()

const currentIndex = ref(props.initialIndex)

watch(() => props.initialIndex, (newIndex) => {
  currentIndex.value = newIndex
})

watch(() => props.isOpen, (isOpen) => {
  if (isOpen) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

function next() {
  currentIndex.value = (currentIndex.value + 1) % props.images.length
}

function prev() {
  currentIndex.value = (currentIndex.value - 1 + props.images.length) % props.images.length
}

function close() {
  emit('close')
}

function handleKeydown(event: KeyboardEvent) {
  if (!props.isOpen) return

  switch (event.key) {
    case 'ArrowRight':
      next()
      break
    case 'ArrowLeft':
      prev()
      break
    case 'Escape':
      close()
      break
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = ''
})
</script>

<style scoped>
.lightbox-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.95);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.lightbox-close {
  position: absolute;
  top: 1.5rem;
  right: 1.5rem;
  background: rgba(255, 255, 255, 0.1);
  border: none;
  color: white;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease;
  z-index: 10001;
}

.lightbox-close:hover {
  background: rgba(255, 255, 255, 0.2);
}

.lightbox-nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(255, 255, 255, 0.1);
  border: none;
  color: white;
  width: 60px;
  height: 60px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease;
  z-index: 10001;
}

.lightbox-nav:hover {
  background: rgba(255, 255, 255, 0.2);
}

.lightbox-nav-prev {
  left: 2rem;
}

.lightbox-nav-next {
  right: 2rem;
}

.lightbox-content {
  max-width: 90vw;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
}

.lightbox-image {
  max-width: 100%;
  max-height: 75vh;
  width: auto;
  height: auto;
  object-fit: contain;
  border-radius: 8px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
}

.lightbox-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  color: white;
}

.lightbox-caption {
  font-size: 1.1rem;
  font-weight: 500;
  text-align: center;
}

.lightbox-counter {
  font-size: 0.9rem;
  color: rgba(255, 255, 255, 0.7);
}

/* Lightbox overlay fade transition */
.lightbox-fade-enter-active,
.lightbox-fade-leave-active {
  transition: opacity 0.3s ease;
}

.lightbox-fade-enter-from,
.lightbox-fade-leave-to {
  opacity: 0;
}

/* Image fade transition */
.image-fade-enter-active,
.image-fade-leave-active {
  transition: opacity 0.25s ease;
}

.image-fade-enter-from,
.image-fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .lightbox-overlay {
    padding: 1rem;
  }

  .lightbox-close {
    top: 1rem;
    right: 1rem;
    width: 40px;
    height: 40px;
  }

  .lightbox-nav {
    width: 48px;
    height: 48px;
  }

  .lightbox-nav-prev {
    left: 1rem;
  }

  .lightbox-nav-next {
    right: 1rem;
  }

  .lightbox-image {
    max-height: 65vh;
  }

  .lightbox-caption {
    font-size: 0.95rem;
  }

  .lightbox-counter {
    font-size: 0.85rem;
  }
}
</style>
