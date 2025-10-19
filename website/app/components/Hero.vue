<template>
  <section class="hero">
    <div class="container">
      <div class="hero-content">
        

        <h1 class="hero-title">
          <span class="text-purple">Claude </span>
          <span class="text-cyan">Control </span>
          <span class="text-green">Terminal</span>
        </h1>

        <p class="hero-subtitle">
          A powerful wrapper and control center for Claude Code
        </p>

        <p class="hero-description">
          Manage components, configure AI providers, control permissions, launch Claude, run live agents, monitor analytics, and deploy with Docker. Now with unified Go server, MCP integration, and real-time agent conversations.
        </p>

        <div class="hero-badges">
          <a href="https://github.com/schlunsen/claude-control-terminal/stargazers" target="_blank">
            <img src="https://img.shields.io/github/stars/schlunsen/claude-control-terminal?style=flat&logo=github&color=yellow" alt="GitHub Stars">
          </a>
          <img src="https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat&logo=go" alt="Go Version">
          <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License">
          <img src="https://img.shields.io/github/v/release/schlunsen/claude-control-terminal" alt="Release">
        </div>

        <div class="quick-install">
          <p class="install-label">Quick Install:</p>
          <pre><code>brew install schlunsen/cct/cct</code></pre>
        </div>

        <div class="hero-video">
          <div class="video-wrapper">
            <iframe
              width="100%"
              height="400"
              src="https://www.youtube.com/embed/iT4Oo3diySg"
              title="CCT Demo Video"
              frameborder="0"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowfullscreen>
            </iframe>
          </div>
          <p class="video-caption">▶️ Watch the full demo on YouTube</p>
        </div>

        <div class="hero-cta">
          <a href="#installation" class="btn btn-primary">Get Started</a>
          <a href="https://github.com/schlunsen/claude-control-terminal" class="btn btn-secondary" target="_blank">View on GitHub</a>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.hero {
  padding: 4rem 0;
  min-height: 100vh;
  display: flex;
  align-items: center;
}

.hero-content {
  text-align: center;
}

.ascii-art pre {
  color: var(--accent-purple);
  font-size: 0.7rem;
  margin-bottom: 1rem;
  line-height: 1.2;
}

.hero-title {
  font-size: 3.5rem;
  margin-bottom: 1rem;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.hero-subtitle {
  font-size: 1.5rem;
  color: var(--text-primary);
  margin-bottom: 1rem;
}

.hero-description {
  font-size: 1.1rem;
  max-width: 700px;
  margin: 0 auto 2.5rem;
  color: var(--text-secondary);
  line-height: 1.7;
}

.hero-badges {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
  margin-bottom: 2rem;
  flex-wrap: wrap;
  align-items: center;
}

.hero-badges img {
  height: auto;
  display: block;
}

.hero-video {
  max-width: 800px;
  margin: 3rem auto;
}

.video-wrapper {
  position: relative;
  padding-bottom: 56.25%;
  height: 0;
  overflow: hidden;
  border: 2px solid var(--border-color);
  border-radius: 8px;
}

.video-wrapper iframe {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.video-caption {
  margin-top: 1rem;
  font-size: 0.9rem;
  color: var(--text-muted);
}

.hero-cta {
  display: flex;
  gap: 1rem;
  justify-content: center;
  margin: 2rem 0;
  flex-wrap: wrap;
}

.quick-install {
  max-width: 600px;
  margin: 2rem auto 0;
  padding: 1.5rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  min-width: 0;
}

.quick-install pre {
  overflow-x: auto;
}

.install-label {
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
  color: var(--text-muted);
}

@media (max-width: 768px) {
  .hero {
    padding: 2rem 0;
    min-height: auto;
  }

  .ascii-art pre {
    font-size: 0.5rem;
  }

  .hero-title {
    font-size: 1.75rem;
    line-height: 1.4;
    padding: 0 0.5rem;
    word-spacing: 0.1em;
  }

  .hero-subtitle {
    font-size: 1.1rem;
  }

  .hero-description {
    font-size: 0.95rem;
    padding: 0 0.5rem;
  }

  .hero-badges {
    gap: 0.75rem;
    margin-bottom: 1.5rem;
  }

  .hero-badges img {
    max-width: 100%;
    height: auto;
  }

  .quick-install {
    padding: 1rem;
    margin: 1.5rem auto 0;
  }

  .quick-install pre {
    font-size: 0.85rem;
    padding: 1rem;
    overflow-x: auto;
  }

  .hero-video {
    margin: 2rem auto;
  }

  .video-caption {
    font-size: 0.85rem;
    padding: 0 0.5rem;
  }

  .hero-cta {
    flex-direction: column;
    gap: 0.75rem;
    margin: 1.5rem 0;
    align-items: center;
  }

  .btn {
    width: 100%;
    max-width: 300px;
  }
}

@media (max-width: 480px) {
  .hero {
    padding: 1.5rem 0;
  }

  .hero-title {
    font-size: 1.5rem;
    line-height: 1.3;
    margin-bottom: 0.75rem;
  }

  .hero-subtitle {
    font-size: 1rem;
    margin-bottom: 0.75rem;
  }

  .hero-description {
    font-size: 0.9rem;
  }

  .quick-install {
    padding: 0.875rem;
    margin: 1.25rem auto 0;
  }

  .install-label {
    font-size: 0.85rem;
  }

  .quick-install pre {
    font-size: 0.75rem;
    padding: 0.75rem;
  }

  .hero-video {
    margin: 1.5rem auto;
  }

  .hero-cta {
    margin: 1.25rem 0;
    align-items: center;
  }
}
</style>
