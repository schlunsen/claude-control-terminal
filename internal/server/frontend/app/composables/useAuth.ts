export const useAuth = () => {
  const isAuthenticated = useState('isAuthenticated', () => false)
  const user = useState('user', () => null as { username: string; isAdmin: boolean } | null)
  const authEnabled = useState('authEnabled', () => false)
  const requireLogin = useState('requireLogin', () => false)
  const showLoginModal = useState('showLoginModal', () => false)

  // Check authentication status
  const checkAuthStatus = async () => {
    try {
      const response = await $fetch('/api/auth/status') as {
        enabled: boolean
        authenticated: boolean
        requireLogin: boolean
        username?: string
        isAdmin?: boolean
      }

      authEnabled.value = response.enabled
      requireLogin.value = response.requireLogin
      isAuthenticated.value = response.authenticated

      if (response.authenticated && response.username) {
        user.value = {
          username: response.username,
          isAdmin: response.isAdmin || false
        }
      } else {
        user.value = null
      }

      // Show login modal if auth is required and user is not authenticated
      if (authEnabled.value && requireLogin.value && !isAuthenticated.value) {
        showLoginModal.value = true
      }

      return response
    } catch (error) {
      console.error('Failed to check auth status:', error)
      return {
        enabled: false,
        authenticated: false,
        requireLogin: false
      }
    }
  }

  // Login
  const login = async (username: string, password: string) => {
    const response = await $fetch('/api/auth/login', {
      method: 'POST',
      body: {
        username,
        password
      }
    }) as {
      token: string
      username: string
      expiresAt: string
    }

    if (response && response.username) {
      await checkAuthStatus()
      showLoginModal.value = false
      return true
    }

    return false
  }

  // Logout
  const logout = async () => {
    try {
      await $fetch('/api/auth/logout', {
        method: 'POST'
      })
    } catch (error) {
      console.error('Logout failed:', error)
    } finally {
      isAuthenticated.value = false
      user.value = null
      showLoginModal.value = true
    }
  }

  // Change password
  const changePassword = async (oldPassword: string, newPassword: string) => {
    await $fetch('/api/auth/change-password', {
      method: 'POST',
      body: {
        old_password: oldPassword,
        new_password: newPassword
      }
    })
  }

  // Create user (admin only)
  const createUser = async (username: string, password: string, isAdmin: boolean) => {
    await $fetch('/api/auth/users', {
      method: 'POST',
      body: {
        username,
        password,
        is_admin: isAdmin
      }
    })
  }

  // List users (admin only)
  const listUsers = async () => {
    const response = await $fetch('/api/auth/users') as {
      users: string[]
      count: number
    }
    return response.users
  }

  // Delete user (admin only)
  const deleteUser = async (username: string) => {
    await $fetch(`/api/auth/users/${username}`, {
      method: 'DELETE'
    })
  }

  return {
    isAuthenticated,
    user,
    authEnabled,
    requireLogin,
    showLoginModal,
    checkAuthStatus,
    login,
    logout,
    changePassword,
    createUser,
    listUsers,
    deleteUser
  }
}
