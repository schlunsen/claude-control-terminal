/**
 * Auto-scroll management for message containers
 */

import { ref, type Ref } from 'vue'
import { nextTick } from 'vue'

export const useMessageScroll = () => {
  const isUserNearBottom = ref(true)

  /**
   * Handle scroll event to track if user is near bottom
   */
  const handleScroll = (container: HTMLElement | null) => {
    if (!container) return

    const { scrollTop, scrollHeight, clientHeight } = container
    const threshold = 100 // pixels from bottom
    isUserNearBottom.value = scrollHeight - scrollTop - clientHeight < threshold
  }

  /**
   * Scroll to bottom of container
   */
  const scrollToBottom = (container: HTMLElement | null, smooth = false) => {
    if (!container) return

    nextTick(() => {
      container.scrollTo({
        top: container.scrollHeight,
        behavior: smooth ? 'smooth' : 'auto'
      })
    })
  }

  /**
   * Auto-scroll to bottom only if user is near bottom
   */
  const autoScrollIfNearBottom = (container: HTMLElement | null, smooth = true) => {
    if (isUserNearBottom.value) {
      scrollToBottom(container, smooth)
    }
  }

  return {
    isUserNearBottom,
    handleScroll,
    scrollToBottom,
    autoScrollIfNearBottom
  }
}
