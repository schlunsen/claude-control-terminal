// Composable for making authenticated API requests with automatic API key handling
export const useAuthenticatedFetch = () => {
  const apiKey = ref<string | null>(null)
  const isLoading = ref(false)

  // Fetch and cache the API key
  const ensureAPIKey = async () => {
    if (apiKey.value) {
      return apiKey.value
    }

    try {
      const { data } = await useFetch<{ apiKey: string }>('/api/config/api-key')
      if (data.value?.apiKey) {
        apiKey.value = data.value.apiKey
        return apiKey.value
      }
    } catch (error) {
      console.warn('Failed to fetch API key:', error)
      // Continue without key - endpoint may not require auth or may be accessible without it
    }
    return null
  }

  // Make an authenticated fetch request
  const fetchWithAuth = async <T = any>(
    url: string,
    options?: RequestInit & { method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' }
  ): Promise<Response> => {
    isLoading.value = true
    try {
      const key = await ensureAPIKey()
      const headers = new Headers(options?.headers || {})

      // Add authorization header if we have an API key and this is a mutating request
      if (key && options?.method && ['POST', 'PUT', 'DELETE', 'PATCH'].includes(options.method)) {
        headers.set('Authorization', `Bearer ${key}`)
      }

      const response = await fetch(url, {
        ...options,
        headers,
      })

      return response
    } finally {
      isLoading.value = false
    }
  }

  return {
    fetchWithAuth,
    ensureAPIKey,
    isLoading,
    apiKey: readonly(apiKey),
  }
}
