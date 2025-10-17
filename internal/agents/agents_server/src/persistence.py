"""Persistence layer for saving/loading agent sessions and messages to/from Go API."""

import asyncio
import json
import logging
import ssl
from typing import Optional
from urllib.parse import urljoin
from uuid import UUID

import aiohttp

from .config import get_api_key, settings

logger = logging.getLogger(__name__)


class PersistenceClient:
    """Client for persisting sessions and messages to the Go API."""

    def __init__(self):
        self.base_url = self._get_base_url()
        self.api_key = get_api_key(settings)
        self.session: Optional[aiohttp.ClientSession] = None

        # Log API key status
        if self.api_key:
            logger.info(f"🔑 API key loaded: {self.api_key[:10]}...{self.api_key[-10:]}")
        else:
            logger.warning("⚠️  No API key found - requests will not be authenticated")

    def _get_base_url(self) -> str:
        """Get the base URL for the Go API."""
        protocol = "https" if settings.go_api_tls else "http"
        return f"{protocol}://{settings.go_api_host}:{settings.go_api_port}"

    async def init(self):
        """Initialize the aiohttp session."""
        try:
            logger.info(f"🔌 Initializing persistence client for {self.base_url}")

            if settings.go_api_tls:
                # For self-signed certificates, create an SSL context that doesn't verify
                ssl_context = ssl.create_default_context()
                ssl_context.check_hostname = False
                ssl_context.verify_mode = ssl.CERT_NONE
                connector = aiohttp.TCPConnector(ssl=ssl_context)
                logger.info("🔐 Using self-signed certificate (SSL verification disabled)")
            else:
                connector = None
                logger.info("🔌 Using HTTP (no SSL)")

            self.session = aiohttp.ClientSession(connector=connector)
            logger.info(f"✅ Persistence client initialized: {self.base_url}")
        except Exception as e:
            logger.error(f"❌ Failed to initialize persistence client: {type(e).__name__}: {e}")
            import traceback
            logger.error(traceback.format_exc())
            raise

    async def close(self):
        """Close the aiohttp session."""
        if self.session:
            await self.session.close()
            logger.info("Persistence client closed")

    def _get_headers(self) -> dict:
        """Get headers for API requests."""
        headers = {"Content-Type": "application/json"}
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
            logger.debug("📤 Adding Authorization header to request")
        else:
            logger.warning("⚠️  No Authorization header - request is unauthenticated")
        return headers

    async def save_session(
        self,
        session_id: UUID,
        session_name: str,
        avatar_name: str = "",
        working_directory: str = "",
        agent_name: str = "",
        system_prompt: str = "",
        permission_mode: str = "default",
        tools: list = None
    ) -> bool:
        """Save a session to the database."""
        if not self.session:
            logger.warning("Persistence client not initialized")
            return False

        try:
            url = urljoin(self.base_url, "/api/agent-sessions")
            payload = {
                "session_id": str(session_id),
                "session_name": session_name,
                "avatar_name": avatar_name,
                "working_directory": working_directory,
                "agent_name": agent_name,
                "system_prompt": system_prompt,
                "permission_mode": permission_mode,
                "tools": tools or []
            }

            async with self.session.post(
                url,
                json=payload,
                headers=self._get_headers(),
                timeout=aiohttp.ClientTimeout(total=5)
            ) as response:
                if response.status == 200:
                    logger.debug(f"Saved session {session_id} to database")
                    return True
                else:
                    logger.warning(
                        f"Failed to save session {session_id}: "
                        f"{response.status} - {await response.text()}"
                    )
                    return False

        except asyncio.TimeoutError:
            logger.warning(f"Timeout saving session {session_id}")
            return False
        except Exception as e:
            logger.warning(f"Error saving session {session_id}: {e}")
            return False

    async def save_message(
        self,
        session_id: UUID,
        message_id: str,
        role: str,
        content: str,
        tool_name: str = "",
        tool_result: str = "",
        token_count: int = 0
    ) -> bool:
        """Save a message to the database."""
        if not self.session:
            logger.error("❌ Persistence client not initialized for save_message")
            return False

        try:
            url = urljoin(self.base_url, f"/api/agent-sessions/{session_id}/messages")
            payload = {
                "message_id": message_id,
                "role": role,
                "content": content,
                "tool_name": tool_name,
                "tool_result": tool_result,
                "token_count": token_count
            }

            logger.info(f"🔄 Saving message {message_id} ({role}) to {url}")
            async with self.session.post(
                url,
                json=payload,
                headers=self._get_headers(),
                timeout=aiohttp.ClientTimeout(total=5)
            ) as response:
                if response.status == 200:
                    logger.info(f"✅ Saved message {message_id} for session {session_id}")
                    return True
                else:
                    resp_text = await response.text()
                    logger.error(
                        f"❌ Failed to save message {message_id} for session {session_id}: "
                        f"Status {response.status} - {resp_text}"
                    )
                    return False

        except asyncio.TimeoutError:
            logger.error(f"❌ Timeout saving message {message_id} for session {session_id}")
            return False
        except Exception as e:
            logger.error(f"❌ Exception saving message {message_id} for session {session_id}: {type(e).__name__}: {e}")
            import traceback
            logger.error(traceback.format_exc())
            return False

    async def load_active_sessions(self) -> list:
        """Load all active sessions from the database."""
        if not self.session:
            logger.warning("Persistence client not initialized")
            return []

        try:
            url = urljoin(self.base_url, "/api/agent-sessions")

            async with self.session.get(
                url,
                headers=self._get_headers(),
                timeout=aiohttp.ClientTimeout(total=5)
            ) as response:
                if response.status == 200:
                    data = await response.json()
                    sessions = data.get("sessions", [])
                    logger.info(f"Loaded {len(sessions)} active sessions from database")
                    return sessions
                else:
                    logger.warning(
                        f"Failed to load sessions: {response.status} - {await response.text()}"
                    )
                    return []

        except asyncio.TimeoutError:
            logger.warning("Timeout loading sessions from database")
            return []
        except Exception as e:
            logger.warning(f"Error loading sessions from database: {e}")
            return []

    async def load_session_messages(self, session_id: UUID, limit: int = 0) -> list:
        """Load all messages for a session from the database."""
        if not self.session:
            logger.warning("Persistence client not initialized")
            return []

        try:
            url = urljoin(self.base_url, f"/api/agent-sessions/{session_id}")

            async with self.session.get(
                url,
                headers=self._get_headers(),
                timeout=aiohttp.ClientTimeout(total=5)
            ) as response:
                if response.status == 200:
                    data = await response.json()
                    messages = data.get("messages", [])
                    logger.debug(f"Loaded {len(messages)} messages for session {session_id}")
                    return messages
                else:
                    logger.warning(
                        f"Failed to load messages for session {session_id}: "
                        f"{response.status} - {await response.text()}"
                    )
                    return []

        except asyncio.TimeoutError:
            logger.warning(f"Timeout loading messages for session {session_id}")
            return []
        except Exception as e:
            logger.warning(f"Error loading messages for session {session_id}: {e}")
            return []


# Global persistence client instance
persistence_client = PersistenceClient()
