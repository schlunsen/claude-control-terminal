<template>
  <section id="features" class="features">
    <div class="container">
      <h2 class="section-title">> features</h2>

      <div class="provider-support">
        <h3 class="provider-title">🔄 Multi-Provider Support</h3>
        <p class="provider-label">Use CCT with any AI provider</p>
        <div class="provider-logos">
          <div class="provider-badge">
            <img src="https://img.shields.io/badge/Claude-1c1c24?style=for-the-badge&logo=anthropic&logoColor=white" alt="Claude">
          </div>
          <div class="provider-badge">
            <img src="https://img.shields.io/badge/OpenAI-1c1c24?style=for-the-badge&logo=openai&logoColor=white" alt="OpenAI">
          </div>
          <div class="provider-badge">
            <img src="https://img.shields.io/badge/DeepSeek-1c1c24?style=for-the-badge&logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDJMMiAxMkwxMiAyMkwyMiAxMkwxMiAyWiIgZmlsbD0id2hpdGUiLz4KPC9zdmc+&logoColor=white" alt="DeepSeek">
          </div>
          <div class="provider-badge">
            <img src="https://img.shields.io/badge/Kimi-1c1c24?style=for-the-badge&logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPGNpcmNsZSBjeD0iMTIiIGN5PSIxMiIgcj0iMTAiIGZpbGw9IndoaXRlIi8+Cjwvc3ZnPg==&logoColor=white" alt="Kimi">
          </div>
          <div class="provider-badge">
            <img src="https://img.shields.io/badge/Z.AI-1c1c24?style=for-the-badge&logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTUgNUgxOUw1IDE5SDE5IiBzdHJva2U9IndoaXRlIiBzdHJva2Utd2lkdGg9IjIiLz4KPC9zdmc+&logoColor=white" alt="Z.AI">
          </div>
          <div class="provider-badge">
            <img src="https://img.shields.io/badge/Custom-1c1c24?style=for-the-badge&logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDJMMTUgOUwyMiAxMkwxNSAxNUwxMiAyMkw5IDE1TDIgMTJMOSA5TDEyIDJaIiBmaWxsPSJ3aGl0ZSIvPgo8L3N2Zz4=&logoColor=white" alt="Custom">
          </div>
        </div>
        <p class="provider-note">✨ Configure any Anthropic-compatible API endpoint</p>
      </div>

      <div class="features-grid">
        <div v-for="feature in features" :key="feature.icon" class="feature-card">
          <div class="feature-icon">{{ feature.icon }}</div>
          <h3>{{ feature.title }}</h3>
          <p>{{ feature.description }}</p>
        </div>
      </div>

      <div class="screenshots">
        <h3 class="screenshots-title">> screenshots</h3>
        <div class="screenshots-grid">
          <div
            v-for="(screenshot, index) in screenshots"
            :key="screenshot.src"
            class="screenshot-item"
            :class="{ 'screenshot-hidden': index >= 4 && !showAllScreenshots }"
            @click="openLightbox(index)">
            <img :src="screenshot.src" :alt="screenshot.alt">
            <p class="screenshot-caption">{{ screenshot.caption }}</p>
          </div>
        </div>
        <div class="screenshots-toggle">
          <button @click="toggleScreenshots" class="btn btn-secondary">
            {{ showAllScreenshots ? 'Show Less' : 'Show More Screenshots' }}
          </button>
        </div>
      </div>
    </div>

    <Lightbox
      :images="screenshots"
      :initial-index="lightboxInitialIndex"
      :is-open="lightboxOpen"
      @close="closeLightbox"
    />
  </section>
</template>

<script setup>
const lightboxOpen = ref(false)
const lightboxInitialIndex = ref(0)
const showAllScreenshots = ref(false)

function openLightbox(index) {
  lightboxInitialIndex.value = index
  lightboxOpen.value = true
}

function closeLightbox() {
  lightboxOpen.value = false
}

function toggleScreenshots() {
  showAllScreenshots.value = !showAllScreenshots.value
}

const features = [
  {
    icon: '🎮',
    title: 'Control Center',
    description: 'Comprehensive wrapper for managing Claude Code environments'
  },
  {
    icon: '🤖',
    title: 'Live Agents Dashboard',
    description: 'Real-time Claude agent conversations with WebSocket streaming, session metrics, and tool tracking'
  },
  {
    icon: '🔌',
    title: 'MCP Server Integration',
    description: 'Extend Claude with Model Context Protocol servers for custom tools and resources'
  },
  {
    icon: '🎨',
    title: 'Multiple Themes',
    description: 'Choose from 5 theme families: Default, Neon, Nord, Dracula, and South Park'
  },
  {
    icon: '⌨️',
    title: 'Keyboard Shortcuts',
    description: 'Global keyboard shortcuts for enhanced productivity and navigation (press \'?\' for help)'
  },
  {
    icon: '💾',
    title: 'Session Persistence',
    description: 'SQLite-backed agent session storage with crash recovery and history loading'
  },
  {
    icon: '🚀',
    title: 'Interactive TUI',
    description: 'Modern terminal interface for browsing and installing components'
  },
  {
    icon: '📦',
    title: 'Component Management',
    description: 'Install agents, commands, and MCPs from 600+ templates'
  },
  {
    icon: '⚙️',
    title: 'Permissions Management',
    description: 'Granular control over Claude Code tool permissions'
  },
  {
    icon: '🐳',
    title: 'Docker Support',
    description: 'Containerize Claude environments with one command'
  },
  {
    icon: '📊',
    title: 'Analytics Dashboard',
    description: 'Real-time WebSocket-based monitoring with process correlation'
  },
  {
    icon: '⚡',
    title: 'High Performance',
    description: '50-100x faster than Node.js version, 5x lower memory'
  },
  {
    icon: '🔧',
    title: 'Zero Dependencies',
    description: 'Single 15MB self-contained binary'
  },
  {
    icon: '🌐',
    title: 'Cross-Platform',
    description: 'Linux, macOS, Windows (amd64/arm64)'
  },
  {
    icon: '✅',
    title: 'Well Tested',
    description: 'Comprehensive test coverage with automated testing'
  }
]

const screenshots = [
  {
    src: '/claude-control-terminal/images/cct-tui-main.png',
    alt: 'CCT TUI Main Screen',
    caption: 'Main TUI Interface'
  },
  {
    src: '/claude-control-terminal/images/cct-tui-agents.png',
    alt: 'Browse Agents',
    caption: 'Browse Agents'
  },
  {
    src: '/claude-control-terminal/images/cct-tui-mcps.png',
    alt: 'Browse MCPs',
    caption: 'Browse MCPs'
  },
  {
    src: '/claude-control-terminal/images/cct-analytics.png',
    alt: 'Analytics Dashboard',
    caption: 'Real-time Analytics Dashboard'
  },
  {
    src: '/claude-control-terminal/images/cct-live-agents.png',
    alt: 'Live Agents Dashboard',
    caption: 'Live Agent Conversations with WebSocket Streaming'
  },
  {
    src: '/claude-control-terminal/images/cct-stats.png',
    alt: 'Detailed Statistics',
    caption: 'Comprehensive Analytics and Performance Metrics'
  },
  {
    src: '/claude-control-terminal/images/cct-themes.png',
    alt: 'Theme Settings',
    caption: 'Multiple Theme Families with Dark and Light Modes'
  },
  {
    src: '/claude-control-terminal/images/cct-shortcuts.png',
    alt: 'Keyboard Shortcuts',
    caption: 'Global Keyboard Shortcuts for Enhanced Productivity'
  }
]
</script>

<style scoped>
.features {
  padding: 4rem 0;
  background: var(--bg-secondary);
}

.section-title {
  font-size: 2rem;
  margin-bottom: 3rem;
  color: var(--accent-cyan);
}

.provider-support {
  padding: 3rem 2rem;
  text-align: center;
  background: var(--bg-tertiary);
  border: 2px solid var(--border-color);
  border-radius: 12px;
  margin-bottom: 4rem;
}

.provider-title {
  font-size: 2.5rem;
  color: var(--text-primary);
  margin-bottom: 0.75rem;
  font-weight: 700;
}

.provider-label {
  font-size: 1.3rem;
  color: var(--text-secondary);
  margin-bottom: 2.5rem;
  font-weight: 500;
}

.provider-logos {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  gap: 2rem;
  margin-bottom: 2rem;
}

.provider-badge {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.provider-badge:hover {
  transform: translateY(-4px);
  opacity: 0.8;
}

.provider-badge img {
  height: 56px;
  width: auto;
}

.provider-note {
  font-size: 1.1rem;
  color: var(--text-muted);
  font-weight: 500;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
  margin-bottom: 4rem;
}

.feature-card {
  padding: 1.5rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  transition: transform 0.2s ease, border-color 0.2s ease;
}

.feature-card:hover {
  transform: translateY(-4px);
  border-color: var(--accent-purple);
}

.feature-icon {
  font-size: 2.5rem;
  margin-bottom: 1rem;
}

.feature-card h3 {
  font-size: 1.1rem;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.feature-card p {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.screenshots {
  margin-top: 4rem;
}

.screenshots-title {
  font-size: 1.5rem;
  margin-bottom: 2rem;
  color: var(--accent-purple);
}

.screenshots-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 2rem;
}

.screenshot-item {
  border: 2px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;
  opacity: 1;
  max-height: 1000px;
}

.screenshot-item.screenshot-hidden {
  opacity: 0;
  max-height: 0;
  margin: 0;
  padding: 0;
  border-width: 0;
  pointer-events: none;
  overflow: hidden;
}

.screenshot-item:hover {
  border-color: var(--accent-cyan);
  transform: translateY(-4px);
}

.screenshot-item img {
  width: 100%;
  height: auto;
  display: block;
}

.screenshot-caption {
  padding: 1rem;
  background: var(--bg-tertiary);
  text-align: center;
  font-size: 0.9rem;
  color: var(--text-muted);
}

.screenshots-toggle {
  text-align: center;
  margin-top: 2rem;
}

.screenshots-toggle .btn {
  min-width: 200px;
}

@media (max-width: 768px) {
  .section-title {
    font-size: 1.5rem;
    margin-bottom: 2rem;
  }

  .provider-support {
    padding: 2rem 1rem;
    margin-bottom: 2rem;
  }

  .provider-title {
    font-size: 1.75rem;
    margin-bottom: 0.5rem;
  }

  .provider-label {
    font-size: 1rem;
    margin-bottom: 1.5rem;
  }

  .provider-logos {
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .provider-badge img {
    height: 36px;
  }

  .provider-note {
    font-size: 0.9rem;
  }

  .features-grid {
    grid-template-columns: 1fr;
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .feature-card {
    padding: 1.25rem;
  }

  .feature-icon {
    font-size: 2rem;
  }

  .screenshots {
    margin-top: 2rem;
  }

  .screenshots-title {
    font-size: 1.25rem;
    margin-bottom: 1.5rem;
  }

  .screenshots-grid {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }

  .screenshot-item {
    border-width: 1px;
  }

  .screenshot-caption {
    padding: 0.75rem;
    font-size: 0.85rem;
  }

  .screenshots-toggle {
    margin-top: 1.5rem;
  }

  .screenshots-toggle .btn {
    min-width: 180px;
    font-size: 0.9rem;
  }
}
</style>
