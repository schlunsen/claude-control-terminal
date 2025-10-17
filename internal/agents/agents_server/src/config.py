"""Configuration management for the agent server."""

import os
from pathlib import Path
from typing import Optional

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings."""

    # Server settings
    host: str = "127.0.0.1"
    port: int = 8001
    reload: bool = False

    # Authentication
    auth_enabled: bool = True
    api_key_path: Path = Path.home() / ".claude" / "analytics" / ".secret"

    # Claude Agent SDK settings
    claude_code_command: str = "claude-code"
    max_concurrent_sessions: int = 10
    session_timeout_seconds: int = 3600  # 1 hour

    # WebSocket settings
    ws_ping_interval: int = 30  # seconds
    ws_ping_timeout: int = 10  # seconds

    # Logging
    log_level: str = "INFO"

    # Go API settings (for persistence)
    go_api_host: str = "127.0.0.1"
    go_api_port: int = 3333
    go_api_tls: bool = True

    class Config:
        env_file = ".env"
        env_prefix = "AGENT_SERVER_"


def get_api_key(settings: Settings) -> Optional[str]:
    """Read the API key from the file system."""
    if not settings.auth_enabled:
        return None

    secret_file = settings.api_key_path
    if secret_file.exists():
        return secret_file.read_text().strip()

    # Try alternative location
    alt_path = Path.home() / ".claude" / "analytics_data" / ".secret"
    if alt_path.exists():
        return alt_path.read_text().strip()

    return None


# Global settings instance
settings = Settings()