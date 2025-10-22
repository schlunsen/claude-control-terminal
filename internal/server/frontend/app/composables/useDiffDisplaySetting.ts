// Composable for managing diff display location setting
export const useDiffDisplaySetting = () => {
  const diffDisplayLocation = ref<'chat' | 'options'>('chat')
  const isLoading = ref(false)
  const { fetchWithAuth } = useAuthenticatedFetch()

  // Fetch the current diff display setting
  const fetchDiffDisplaySetting = async () => {
    isLoading.value = true
    try {
      const response = await fetchWithAuth('/api/settings/diff_display_location', {
        method: 'GET',
      })

      if (response.ok) {
        const setting = await response.json()
        diffDisplayLocation.value = (setting.value || 'chat') as 'chat' | 'options'
      }
    } catch (error) {
      console.warn('Failed to fetch diff display setting:', error)
      // Default to 'chat' if fetch fails
      diffDisplayLocation.value = 'chat'
    } finally {
      isLoading.value = false
    }
  }

  // Initialize on first use
  onMounted(() => {
    fetchDiffDisplaySetting()
  })

  return {
    diffDisplayLocation: readonly(diffDisplayLocation),
    isLoading: readonly(isLoading),
    fetchDiffDisplaySetting,
  }
}
