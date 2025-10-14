// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  // Static site generation for GitHub Pages
  ssr: false,

  app: {
    baseURL: '/claude-control-terminal/',
    head: {
      title: 'Claude Control Terminal (CCT)',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'A powerful wrapper and control center for Claude Code - Manage components, configure AI providers, control permissions, and deploy with Docker.' },
        { property: 'og:title', content: 'Claude Control Terminal (CCT)' },
        { property: 'og:description', content: 'A powerful wrapper and control center for Claude Code' },
        { property: 'og:type', content: 'website' }
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/claude-control-terminal/favicon.ico' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&display=swap' }
      ]
    }
  },
})
