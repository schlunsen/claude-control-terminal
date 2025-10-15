// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },

  // SPA mode for real-time dashboard
  ssr: false,

  // Generate static SPA for Go server
  nitro: {
    preset: 'static',
    prerender: {
      crawlLinks: false,
      routes: ['/']
    }
  },

  // Development server configuration
  devServer: {
    port: 3001
  },

  // Vite configuration for API proxy
  vite: {
    server: {
      proxy: {
        '/api': {
          target: 'http://localhost:3333',
          changeOrigin: true
        },
        '/ws': {
          target: 'ws://localhost:3333',
          ws: true
        }
      }
    }
  },

  // App configuration
  app: {
    head: {
      title: 'Claude Control Terminal - Analytics',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Real-time analytics dashboard for Claude Code' }
      ]
    }
  }
})
