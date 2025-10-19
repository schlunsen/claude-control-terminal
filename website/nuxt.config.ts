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
        { name: 'description', content: 'A powerful wrapper and control center for Claude Code - Manage components, configure AI providers, control permissions, run live agents with MCP integration, monitor analytics, and deploy with Docker.' },

        // OpenGraph
        { property: 'og:site_name', content: 'Claude Control Terminal' },
        { property: 'og:title', content: 'Claude Control Terminal (CCT)' },
        { property: 'og:description', content: 'A powerful wrapper and control center for Claude Code - Manage components, configure AI providers, control permissions, run live agents with MCP integration, monitor analytics, and deploy with Docker.' },
        { property: 'og:type', content: 'website' },
        { property: 'og:url', content: 'https://schlunsen.github.io/claude-control-terminal/' },
        { property: 'og:image', content: 'https://schlunsen.github.io/claude-control-terminal/images/cct-tui-main.png' },
        { property: 'og:image:width', content: '1200' },
        { property: 'og:image:height', content: '630' },
        { property: 'og:image:alt', content: 'Claude Control Terminal TUI Interface' },

        // Twitter Card
        { name: 'twitter:card', content: 'summary_large_image' },
        { name: 'twitter:title', content: 'Claude Control Terminal (CCT)' },
        { name: 'twitter:description', content: 'A powerful wrapper and control center for Claude Code - Manage components, configure AI providers, control permissions, run live agents with MCP integration, monitor analytics, and deploy with Docker.' },
        { name: 'twitter:image', content: 'https://schlunsen.github.io/claude-control-terminal/images/cct-tui-main.png' },
        { name: 'twitter:image:alt', content: 'Claude Control Terminal TUI Interface' }
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
