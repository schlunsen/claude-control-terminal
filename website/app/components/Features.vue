<template>
  <section id="features" class="features">
    <div class="container">
      <h2 class="section-title">> features</h2>

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
          <div v-for="screenshot in screenshots" :key="screenshot.src" class="screenshot-item">
            <img :src="screenshot.src" :alt="screenshot.alt">
            <p class="screenshot-caption">{{ screenshot.caption }}</p>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
const features = [
  {
    icon: '🎮',
    title: 'Control Center',
    description: 'Comprehensive wrapper for managing Claude Code environments'
  },
  {
    icon: '🤖',
    title: 'Claude CLI Installer',
    description: 'Automatic installation and setup of Claude CLI (no Node.js required)'
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
    icon: '🔄',
    title: 'AI Provider Configuration',
    description: 'Switch between Claude, DeepSeek, GLM, Kimi, and custom providers'
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
  transition: border-color 0.2s ease;
}

.screenshot-item:hover {
  border-color: var(--accent-cyan);
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

@media (max-width: 768px) {
  .features-grid {
    grid-template-columns: 1fr;
  }

  .screenshots-grid {
    grid-template-columns: 1fr;
  }
}
</style>
