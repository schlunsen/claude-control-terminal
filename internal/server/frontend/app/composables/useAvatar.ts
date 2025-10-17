// Composable for avatar management and selection
export const useAvatar = () => {
  const availableAvatars = ref<string[]>([])
  const isLoading = ref(false)

  // Fetch available avatars from the API
  const loadAvatars = async () => {
    isLoading.value = true
    try {
      const { data } = await useFetch<{ avatars: string[] }>('/api/avatars')
      if (data.value?.avatars) {
        availableAvatars.value = data.value.avatars
      }
    } catch (error) {
      console.warn('Failed to load avatars:', error)
      // Fallback to default avatars if API fails
      availableAvatars.value = ['default']
    } finally {
      isLoading.value = false
    }
  }

  // Get a random avatar
  const getRandomAvatar = (): string => {
    if (availableAvatars.value.length === 0) {
      return 'default'
    }
    const randomIndex = Math.floor(Math.random() * availableAvatars.value.length)
    return availableAvatars.value[randomIndex]
  }

  return {
    availableAvatars: readonly(availableAvatars),
    isLoading: readonly(isLoading),
    loadAvatars,
    getRandomAvatar,
  }
}
