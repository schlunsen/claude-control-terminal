<template>
  <section id="documentation" class="documentation">
    <div class="container">
      <h2 class="section-title">> documentation</h2>

      <div class="docs-grid">
        <div v-for="doc in docs" :key="doc.title" class="doc-card">
          <h3>{{ doc.icon }} {{ doc.title }}</h3>
          <p>{{ doc.description }}</p>
          <ul>
            <li v-for="item in doc.items" :key="item">{{ item }}</li>
          </ul>
          <a v-if="doc.link" :href="doc.link" class="doc-link" target="_blank">
            Learn more →
          </a>
        </div>
      </div>

      <div class="resources">
        <h3 class="subsection-title">> resources</h3>
        <div class="resource-links">
          <a href="https://github.com/schlunsen/claude-control-terminal" class="resource-link" target="_blank">
            📚 GitHub Repository
          </a>
          <a href="https://github.com/schlunsen/claude-control-terminal/blob/main/README.md" class="resource-link" target="_blank">
            📖 Full Documentation
          </a>
          <a href="https://github.com/schlunsen/claude-control-terminal/blob/main/CONTRIBUTING.md" class="resource-link" target="_blank">
            🤝 Contributing Guide
          </a>
          <a href="https://github.com/schlunsen/claude-control-terminal/releases" class="resource-link" target="_blank">
            🚀 Releases
          </a>
          <a href="https://github.com/schlunsen/claude-control-terminal/issues" class="resource-link" target="_blank">
            🐛 Issues
          </a>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
const docs = [
  {
    icon: '📦',
    title: 'Component Management',
    description: 'Install agents, commands, and MCPs from 600+ templates',
    items: [
      '600+ agents across 25+ categories',
      '200+ commands across 19+ categories',
      'MCP server integrations',
      'Smart category search'
    ],
    link: 'https://github.com/schlunsen/claude-control-terminal#component-installation'
  },
  {
    icon: '🔄',
    title: 'AI Provider Configuration',
    description: 'Configure and switch between different AI providers',
    items: [
      'Claude (Official Anthropic)',
      'DeepSeek AI',
      'GLM (Z.ai)',
      'Kimi (Moonshot AI)',
      'Custom providers'
    ],
    link: 'https://github.com/schlunsen/claude-control-terminal#ai-provider-configuration'
  },
  {
    icon: '⚙️',
    title: 'Permissions Management',
    description: 'Granular control over Claude Code tool permissions',
    items: [
      'Global, Project, and Local levels',
      'File operation permissions',
      'Bash command permissions',
      'Web access control'
    ],
    link: 'https://github.com/schlunsen/claude-control-terminal#permissions-management'
  },
  {
    icon: '🐳',
    title: 'Docker Support',
    description: 'Containerize Claude Code environments',
    items: [
      '4 Dockerfile types',
      'Docker Compose templates',
      'MCP integration in containers',
      'Production-ready setups'
    ],
    link: 'https://github.com/schlunsen/claude-control-terminal#docker-support'
  },
  {
    icon: '📊',
    title: 'Analytics Dashboard',
    description: 'Real-time monitoring with WebSocket updates',
    items: [
      'Conversation monitoring',
      'Process correlation',
      'System statistics',
      'REST API endpoints'
    ],
    link: 'https://github.com/schlunsen/claude-control-terminal#analytics-dashboard'
  },
  {
    icon: '🧪',
    title: 'Testing & Development',
    description: 'Comprehensive testing and development tools',
    items: [
      '35%+ test coverage',
      'Multi-platform testing',
      'GitHub Actions CI/CD',
      'Make and Just build tools'
    ],
    link: 'https://github.com/schlunsen/claude-control-terminal/blob/main/TESTING.md'
  }
]
</script>

<style scoped>
.documentation {
  padding: 4rem 0;
  background: var(--bg-secondary);
}

.section-title {
  font-size: 2rem;
  margin-bottom: 3rem;
  color: var(--accent-purple);
}

.subsection-title {
  font-size: 1.5rem;
  margin-bottom: 2rem;
  color: var(--accent-cyan);
}

.docs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 2rem;
  margin-bottom: 4rem;
}

.doc-card {
  padding: 1.5rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  transition: border-color 0.2s ease;
}

.doc-card:hover {
  border-color: var(--accent-purple);
}

.doc-card h3 {
  font-size: 1.2rem;
  margin-bottom: 0.75rem;
  color: var(--text-primary);
}

.doc-card p {
  font-size: 0.9rem;
  margin-bottom: 1rem;
  color: var(--text-secondary);
}

.doc-card ul {
  list-style: none;
  margin-bottom: 1rem;
}

.doc-card li {
  font-size: 0.85rem;
  color: var(--text-muted);
  padding: 0.25rem 0;
}

.doc-card li::before {
  content: '→ ';
  color: var(--accent-cyan);
  margin-right: 0.5rem;
}

.doc-link {
  display: inline-block;
  font-size: 0.9rem;
  color: var(--accent-purple);
  font-weight: 600;
  margin-top: 0.5rem;
}

.doc-link:hover {
  color: var(--accent-cyan);
}

.resources {
  margin-top: 4rem;
}

.resource-links {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.resource-link {
  padding: 1rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  text-align: center;
  font-size: 0.9rem;
  transition: border-color 0.2s ease, background 0.2s ease;
}

.resource-link:hover {
  border-color: var(--accent-cyan);
  background: var(--bg-primary);
}

@media (max-width: 768px) {
  .documentation {
    padding: 2rem 0;
  }

  .section-title {
    font-size: 1.5rem;
    margin-bottom: 2rem;
  }

  .subsection-title {
    font-size: 1.25rem;
    margin-bottom: 1.5rem;
  }

  .docs-grid {
    grid-template-columns: 1fr;
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .doc-card {
    padding: 1.25rem;
  }

  .doc-card h3 {
    font-size: 1.05rem;
  }

  .doc-card p {
    font-size: 0.85rem;
  }

  .doc-card li {
    font-size: 0.8rem;
  }

  .resources {
    margin-top: 2rem;
  }

  .resource-links {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }

  .resource-link {
    padding: 0.875rem;
    font-size: 0.85rem;
  }
}
</style>
