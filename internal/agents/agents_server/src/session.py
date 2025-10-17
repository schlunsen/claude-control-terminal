"""Session management for agent conversations."""

import asyncio
import json
import logging
from datetime import datetime, timedelta
from typing import Dict, Optional
from uuid import UUID

from .config import settings
from .models import AgentSession, SessionOptions, SessionStatus

logger = logging.getLogger(__name__)


class SessionManager:
    """Manages multiple agent sessions."""

    def __init__(self):
        self.sessions: Dict[UUID, AgentSession] = {}
        self.session_locks: Dict[UUID, asyncio.Lock] = {}
        self._cleanup_task: Optional[asyncio.Task] = None

    async def start(self):
        """Start the session manager and cleanup task."""
        # Import here to avoid circular imports
        from .persistence import persistence_client

        # Initialize persistence client
        await persistence_client.init()

        # Load active sessions from database
        try:
            sessions_data = await persistence_client.load_active_sessions()
            if sessions_data:
                logger.info(f"Loading {len(sessions_data)} active sessions from database...")
                for session_data in sessions_data:
                    try:
                        session_id = UUID(session_data["session_id"])

                        # Recreate session from database data
                        tools = []
                        if session_data.get("tools"):
                            try:
                                tools_str = session_data["tools"]
                                if isinstance(tools_str, str):
                                    tools = json.loads(tools_str)
                                else:
                                    tools = tools_str
                            except (json.JSONDecodeError, TypeError):
                                tools = []

                        options = SessionOptions(
                            system_prompt=session_data.get("system_prompt", ""),
                            agent_name=session_data.get("agent_name", ""),
                            tools=tools or [],
                            working_directory=session_data.get("working_directory", ""),
                            permission_mode=session_data.get("permission_mode", "default")
                        )

                        session = AgentSession(
                            id=session_id,
                            options=options,
                            status=SessionStatus.IDLE
                        )
                        session.message_count = session_data.get("message_count", 0)

                        self.sessions[session.id] = session
                        self.session_locks[session.id] = asyncio.Lock()
                        logger.info(f"Restored session {session_id} with {session.message_count} messages")

                    except Exception as e:
                        logger.warning(f"Failed to restore session from database: {e}")
                        continue

        except Exception as e:
            logger.warning(f"Error loading active sessions from database: {e}")

        self._cleanup_task = asyncio.create_task(self._cleanup_loop())
        logger.info("Session manager started")

    async def stop(self):
        """Stop the session manager and cleanup all sessions."""
        if self._cleanup_task:
            self._cleanup_task.cancel()
            try:
                await self._cleanup_task
            except asyncio.CancelledError:
                pass

        # End all active sessions
        for session_id in list(self.sessions.keys()):
            await self.end_session(session_id)

        # Close persistence client
        from .persistence import persistence_client
        await persistence_client.close()

        logger.info("Session manager stopped")

    async def create_session(
        self,
        session_id: Optional[UUID] = None,
        options: Optional[SessionOptions] = None
    ) -> AgentSession:
        """
        Create a new agent session.

        Args:
            session_id: Optional session ID, will generate if not provided
            options: Session configuration options

        Returns:
            The created AgentSession

        Raises:
            ValueError: If max concurrent sessions reached or session already exists
        """
        if len(self.sessions) >= settings.max_concurrent_sessions:
            raise ValueError(
                f"Maximum concurrent sessions ({settings.max_concurrent_sessions}) reached"
            )

        if session_id and session_id in self.sessions:
            raise ValueError(f"Session {session_id} already exists")

        session = AgentSession(
            id=session_id or UUID(),
            options=options or SessionOptions(),
            status=SessionStatus.IDLE
        )

        self.sessions[session.id] = session
        self.session_locks[session.id] = asyncio.Lock()

        logger.info(f"Created session {session.id}")
        return session

    async def get_session(self, session_id: UUID) -> Optional[AgentSession]:
        """Get a session by ID."""
        return self.sessions.get(session_id)

    async def update_session(
        self,
        session_id: UUID,
        status: Optional[SessionStatus] = None,
        error_message: Optional[str] = None
    ) -> Optional[AgentSession]:
        """
        Update a session's status.

        Args:
            session_id: The session to update
            status: New status
            error_message: Error message if status is ERROR

        Returns:
            The updated session or None if not found
        """
        session = self.sessions.get(session_id)
        if not session:
            return None

        async with self.session_locks[session_id]:
            if status:
                session.status = status
            if error_message:
                session.error_message = error_message
            session.updated_at = datetime.now()

        return session

    async def increment_message_count(self, session_id: UUID) -> Optional[AgentSession]:
        """Increment the message count for a session."""
        session = self.sessions.get(session_id)
        if not session:
            return None

        async with self.session_locks[session_id]:
            session.message_count += 1
            session.updated_at = datetime.now()

        return session

    async def end_session(self, session_id: UUID) -> bool:
        """
        End and remove a session.

        Args:
            session_id: The session to end

        Returns:
            True if session was ended, False if not found
        """
        if session_id not in self.sessions:
            return False

        # Update status before removing
        await self.update_session(session_id, status=SessionStatus.ENDED)

        # Remove session and its lock
        del self.sessions[session_id]
        if session_id in self.session_locks:
            del self.session_locks[session_id]

        logger.info(f"Ended session {session_id}")
        return True

    async def list_sessions(self) -> list[AgentSession]:
        """Get all active sessions."""
        return list(self.sessions.values())

    async def _cleanup_loop(self):
        """Periodically clean up idle sessions."""
        while True:
            try:
                await asyncio.sleep(60)  # Check every minute
                await self._cleanup_idle_sessions()
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in cleanup loop: {e}")

    async def _cleanup_idle_sessions(self):
        """Remove sessions that have been idle too long."""
        timeout_threshold = datetime.now() - timedelta(
            seconds=settings.session_timeout_seconds
        )

        sessions_to_remove = []
        for session_id, session in self.sessions.items():
            if (
                session.status == SessionStatus.IDLE
                and session.updated_at < timeout_threshold
            ):
                sessions_to_remove.append(session_id)

        for session_id in sessions_to_remove:
            logger.info(f"Removing idle session {session_id}")
            await self.end_session(session_id)


# Global session manager instance
session_manager = SessionManager()