export default defineNuxtRouteMiddleware(async (to, from) => {
  // Skip middleware on server-side
  if (process.server) {
    return
  }

  const { checkAuthStatus, authEnabled, requireLogin, isAuthenticated } = useAuth()

  // Check authentication status
  await checkAuthStatus()

  // If auth is not enabled, allow access
  if (!authEnabled.value) {
    return
  }

  // If on login page, allow access
  if (to.path === '/login') {
    return
  }

  // If login is required and user is not authenticated, redirect to login
  if (requireLogin.value && !isAuthenticated.value) {
    return navigateTo({
      path: '/login',
      query: {
        redirect: to.fullPath,
        required: 'true'
      }
    })
  }

  // Allow access
})
