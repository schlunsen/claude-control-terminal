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
      ],
      link: [
        // Google Fonts for theme-specific typography
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        // Default: Inter (modern, clean)
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap' },
        // Neon: Orbitron (futuristic, cyberpunk)
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Orbitron:wght@400;500;600;700;800;900&display=swap' },
        // Nord: Fira Code (developer-focused with ligatures)
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Fira+Code:wght@300;400;500;600;700&display=swap' },
        // Nord: Fira Sans (companion to Fira Code)
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Fira+Sans:wght@400;500;600;700&display=swap' },
        // Dracula: JetBrains Mono (professional coding font)
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&display=swap' }
      ]
    }
  }
})
