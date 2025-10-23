<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-card">
        <div class="login-header">
          <div class="logo">
            <span class="logo-text">CCT</span>
          </div>
          <h1>{{ title }}</h1>
          <p>{{ subtitle }}</p>
        </div>

        <div class="login-body">
          <form @submit.prevent="handleSubmit">
            <div class="form-group">
              <label for="username">Username</label>
              <input
                id="username"
                v-model="username"
                type="text"
                placeholder="Enter your username"
                required
                autocomplete="username"
                :disabled="loading"
                autofocus
              />
            </div>

            <div class="form-group">
              <label for="password">Password</label>
              <input
                id="password"
                v-model="password"
                type="password"
                placeholder="Enter your password"
                required
                autocomplete="current-password"
                :disabled="loading"
              />
            </div>

            <div v-if="error" class="error-message">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              {{ error }}
            </div>

            <button
              type="submit"
              class="btn btn-primary"
              :disabled="loading || !username || !password"
            >
              <span v-if="loading">Logging in...</span>
              <span v-else>Login</span>
            </button>
          </form>
        </div>

        <div class="login-footer">
          <p class="footer-text">Claude Control Terminal v{{ version }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: false // Use no layout for login page
})

const route = useRoute()
const router = useRouter()
const { login, checkAuthStatus } = useAuth()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const version = ref('')

const title = computed(() => {
  return route.query.required === 'true' ? 'Login Required' : 'Welcome Back'
})

const subtitle = computed(() => {
  return route.query.required === 'true'
    ? 'Please login to access the dashboard'
    : 'Sign in to your account'
})

// Load version info
async function loadVersion() {
  try {
    const { data } = await useFetch('/api/version')
    if (data.value) {
      version.value = data.value.version || ''
    }
  } catch (err) {
    // Ignore version errors
  }
}

const handleSubmit = async () => {
  if (!username.value || !password.value) {
    return
  }

  loading.value = true
  error.value = ''

  try {
    await login(username.value, password.value)

    // Check auth status to update global state
    await checkAuthStatus()

    // Redirect to the page they were trying to access, or home
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (err: any) {
    error.value = err.data?.error || 'Login failed. Please check your credentials.'
  } finally {
    loading.value = false
  }
}

// Check if already authenticated
onMounted(async () => {
  loadVersion()

  const status = await checkAuthStatus()
  if (status.authenticated) {
    // Already logged in, redirect to home or requested page
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  }
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  padding: 20px;
}

.login-container {
  width: 100%;
  max-width: 440px;
}

.login-card {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.login-header {
  text-align: center;
  padding: 48px 40px 32px;
  border-bottom: 1px solid var(--border-color);
}

.logo {
  margin-bottom: 24px;
}

.logo-text {
  display: inline-block;
  font-size: 2rem;
  font-weight: 700;
  color: var(--accent-purple);
  padding: 12px 24px;
  background: rgba(138, 108, 255, 0.1);
  border-radius: 8px;
}

.login-header h1 {
  margin: 0 0 8px;
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--text-primary);
}

.login-header p {
  margin: 0;
  font-size: 0.9375rem;
  color: var(--text-secondary);
}

.login-body {
  padding: 40px;
}

.form-group {
  margin-bottom: 24px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: var(--text-primary);
  font-size: 0.875rem;
}

.form-group input {
  width: 100%;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 0.9375rem;
  transition: all 0.2s ease;
}

.form-group input:focus {
  outline: none;
  border-color: var(--accent-purple);
  box-shadow: 0 0 0 3px rgba(138, 108, 255, 0.1);
}

.form-group input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: rgba(255, 100, 100, 0.1);
  border: 1px solid rgba(255, 100, 100, 0.3);
  border-radius: 6px;
  color: #ff6464;
  font-size: 0.875rem;
  margin-bottom: 24px;
}

.error-message svg {
  flex-shrink: 0;
}

.btn {
  width: 100%;
  padding: 12px 24px;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  font-size: 0.9375rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary {
  background: var(--accent-purple);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-purple-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(138, 108, 255, 0.3);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-footer {
  padding: 20px 40px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-secondary);
}

.footer-text {
  margin: 0;
  text-align: center;
  font-size: 0.8125rem;
  color: var(--text-muted);
}

@media (max-width: 480px) {
  .login-page {
    padding: 15px;
  }

  .login-header {
    padding: 36px 24px 24px;
  }

  .login-header h1 {
    font-size: 1.5rem;
  }

  .login-body {
    padding: 32px 24px;
  }

  .login-footer {
    padding: 16px 24px;
  }
}
</style>
