// Composable for managing diff display location setting
export const useDiffDisplaySetting = () => {
  const diffDisplayLocation = ref<'chat' | 'options' | 'panel'>('options')
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
        diffDisplayLocation.value = (setting.value || 'options') as 'chat' | 'options' | 'panel'
      }
    } catch (error) {
      console.warn('Failed to fetch diff display setting:', error)
      // Default to 'options' if fetch fails
      diffDisplayLocation.value = 'options'
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
