<template>
  <section id="installation" class="installation">
    <div class="container">
      <h2 class="section-title">> installation</h2>

      <div class="install-grid">
        <div v-for="method in installMethods" :key="method.title" class="install-card">
          <div class="install-header">
            <span class="install-icon">{{ method.icon }}</span>
            <h3>{{ method.title }}</h3>
          </div>
          <p class="install-description">{{ method.description }}</p>
          <pre><code>{{ method.command }}</code></pre>
        </div>
      </div>

      <div class="quick-start">
        <h3 class="subsection-title">> quick start</h3>
        <div class="quick-start-content">
          <div class="quick-start-option">
            <h4>🎮 Interactive TUI (Recommended)</h4>
            <p>Launch the interactive Terminal User Interface:</p>
            <pre><code># Run without arguments to launch TUI
cct

# Or with custom directory
cct -d ~/my-project</code></pre>
          </div>

          <div class="quick-start-features">
            <h4>What you can do with CCT:</h4>
            <ul class="feature-list">
              <li>
                <span class="feature-icon">📦</span>
                <div class="feature-content">
                  <strong>Browse & Install Components</strong>
                  <p>Toggle MCPs, agents, and commands for your project with a single keypress. Search through 600+ agents and 200+ commands.</p>
                </div>
              </li>
              <li>
                <span class="feature-icon">🔄</span>
                <div class="feature-content">
                  <strong>Switch AI Providers</strong>
                  <p>Change between Claude, OpenAI, DeepSeek, Kimi, Z.AI, or configure custom Anthropic-compatible endpoints.</p>
                </div>
              </li>
              <li>
                <span class="feature-icon">⚙️</span>
                <div class="feature-content">
                  <strong>Manage Permissions</strong>
                  <p>Easily set multiple permissions across Global, Project, and Local levels. Control bash commands, file operations, and web access.</p>
                </div>
              </li>
              <li>
                <span class="feature-icon">🚀</span>
                <div class="feature-content">
                  <strong>Launch & Monitor</strong>
                  <p>Start Claude CLI directly from CCT and monitor conversations in real-time with the built-in analytics dashboard.</p>
                </div>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
const installMethods = [
  {
    icon: '🍺',
    title: 'Homebrew (macOS/Linux)',
    description: 'Recommended for macOS and Linux users',
    command: 'brew install schlunsen/cct/cct'
  },
  {
    icon: '📦',
    title: 'Download Binary',
    description: 'Pre-built binaries for all platforms',
    command: `curl -L https://github.com/schlunsen/claude-control-terminal/releases/latest/download/cct-<platform>-<arch> -o cct
chmod +x cct
sudo mv cct /usr/local/bin/`
  },
  {
    icon: '🔧',
    title: 'Install with Go',
    description: 'For Go developers',
    command: 'go install github.com/schlunsen/claude-control-terminal/cmd/cct@latest'
  },
  {
    icon: '⚙️',
    title: 'Build from Source',
    description: 'Build it yourself',
    command: `git clone https://github.com/schlunsen/claude-control-terminal
cd claude-control-terminal
make build`
  }
]
</script>

<style scoped>
.installation {
  padding: 4rem 0;
  background: var(--bg-primary);
}

.section-title {
  font-size: 2rem;
  margin-bottom: 3rem;
  color: var(--accent-green);
}

.subsection-title {
  font-size: 1.5rem;
  margin-bottom: 2rem;
  color: var(--accent-yellow);
}

.install-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
  margin-bottom: 4rem;
}

.install-card {
  padding: 1.5rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.install-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.install-icon {
  font-size: 1.5rem;
}

.install-card h3 {
  font-size: 1.1rem;
  color: var(--text-primary);
  margin: 0;
}

.install-description {
  font-size: 0.9rem;
  color: var(--text-muted);
  margin-bottom: 1rem;
}

.quick-start {
  margin-top: 4rem;
}

.quick-start-content {
  display: grid;
  gap: 2rem;
}

.quick-start-option {
  padding: 1.5rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.quick-start-option h4 {
  font-size: 1.1rem;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.quick-start-option p {
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.quick-start-features {
  padding: 2rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  margin-top: 2rem;
}

.quick-start-features h4 {
  font-size: 1.3rem;
  margin-bottom: 1.5rem;
  color: var(--text-primary);
}

.feature-list {
  list-style: none;
  display: grid;
  gap: 1.5rem;
}

.feature-list li {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.feature-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
  margin-top: 0.2rem;
}

.feature-content {
  flex: 1;
}

.feature-content strong {
  display: block;
  font-size: 1rem;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}

.feature-content p {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.5;
}

@media (max-width: 768px) {
  .quick-start-features {
    padding: 1.5rem;
  }

  .feature-list li {
    flex-direction: column;
    gap: 0.5rem;
  }

  .feature-icon {
    margin-top: 0;
  }
}
</style>
