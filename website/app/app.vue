<script setup>
import '../assets/css/main.css'
</script>

<template>
  <div>
    <nav class="navbar">
      <div class="container">
        <div class="nav-brand">
          <span class="nav-logo">CCT</span>
          <span class="nav-separator">|</span>
          <span class="nav-title">Claude Control Terminal</span>
        </div>
        <div class="nav-links">
          <a href="#features">Features</a>
          <a href="#installation">Install</a>
          <a href="#documentation">Docs</a>
          <a href="https://github.com/schlunsen/claude-control-terminal" target="_blank">GitHub</a>
        </div>
      </div>
    </nav>

    <main>
      <Hero />
      <Features />
      <Installation />
      <Documentation />
    </main>

    <footer class="footer">
      <div class="container">
        <div class="footer-content">
          <p>
            <span class="text-purple">Claude Control Terminal</span>
            <span class="text-muted"> - MIT License</span>
          </p>
          <p class="footer-links">
            <a href="https://github.com/schlunsen/claude-control-terminal" target="_blank">GitHub</a>
            <span class="text-muted">•</span>
            <a href="https://github.com/schlunsen/claude-control-terminal/issues" target="_blank">Issues</a>
            <span class="text-muted">•</span>
            <a href="https://github.com/schlunsen/claude-control-terminal/blob/main/CONTRIBUTING.md" target="_blank">Contributing</a>
          </p>
          <p class="footer-credit">
            Based on <a href="https://github.com/davila7/claude-code-templates" target="_blank">claude-code-templates</a> by davila7
          </p>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.navbar {
  position: sticky;
  top: 0;
  z-index: 1000;
  background: rgba(10, 10, 15, 0.95);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--border-color);
  padding: 1rem 0;
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 600;
}

.nav-logo {
  color: var(--accent-purple);
  font-size: 1.2rem;
}

.nav-separator {
  color: var(--text-muted);
}

.nav-title {
  color: var(--text-primary);
  font-size: 1rem;
}

.navbar .container {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.nav-links {
  display: flex;
  gap: 2rem;
}

.nav-links a {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text-secondary);
  transition: color 0.2s ease;
}

.nav-links a:hover {
  color: var(--accent-cyan);
}

.footer {
  padding: 3rem 0 2rem;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-color);
  margin-top: 4rem;
}

.footer-content {
  text-align: center;
}

.footer-content p {
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
}

.footer-links {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  align-items: center;
}

.footer-links a {
  font-size: 0.9rem;
}

.footer-credit {
  font-size: 0.85rem;
  color: var(--text-muted);
}

@media (max-width: 768px) {
  .navbar {
    padding: 0.75rem 0;
  }

  .navbar .container {
    flex-direction: column;
    gap: 0.75rem;
  }

  .nav-brand {
    gap: 0.5rem;
  }

  .nav-logo {
    font-size: 1.1rem;
  }

  .nav-links {
    gap: 1rem;
    font-size: 0.85rem;
  }

  .nav-links a {
    font-size: 0.85rem;
  }

  .nav-title {
    display: none;
  }

  .footer {
    padding: 2rem 0 1.5rem;
    margin-top: 2rem;
  }

  .footer-content p {
    font-size: 0.85rem;
    margin-bottom: 0.75rem;
  }

  .footer-links {
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .footer-links a {
    font-size: 0.85rem;
  }

  .footer-credit {
    font-size: 0.8rem;
  }
}

@media (max-width: 480px) {
  .navbar .container {
    gap: 0.5rem;
  }

  .nav-links {
    gap: 0.75rem;
    width: 100%;
    justify-content: center;
  }

  .nav-links a {
    font-size: 0.8rem;
  }

  .footer-content p {
    font-size: 0.8rem;
  }

  .footer-links a {
    font-size: 0.8rem;
  }
}
</style>
