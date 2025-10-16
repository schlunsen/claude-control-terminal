"""Main FastAPI application with WebSocket endpoint for agent conversations."""

import asyncio
import json
import logging
from contextlib import asynccontextmanager
from typing import Dict

import uvicorn
from fastapi import FastAPI, WebSocket, WebSocketDisconnect, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from .agent_manager import agent_manager
from .auth import authenticate_websocket
from .config import settings
from .models import (
    AgentMessage,
    AgentsKilledMessage,
    CreateSessionMessage,
    EndSessionMessage,
    ErrorMessage,
    KillAllAgentsMessage,
    ListSessionsMessage,
    MessageType,
    PermissionResponseMessage,
    PingMessage,
    PongMessage,
    SendPromptMessage,
    SessionCreatedMessage,
    SessionEndedMessage,
    SessionsListMessage,
    SessionStatus,
)
from .session import session_manager

# Configure logging
logging.basicConfig(
    level=getattr(logging, settings.log_level),
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage application lifecycle."""
    # Startup
    logger.info("Starting agent server...")
    await session_manager.start()
    yield
    # Shutdown
    logger.info("Shutting down agent server...")
    await session_manager.stop()
    await agent_manager.cleanup_all()


app = FastAPI(
    title="Claude Agent WebSocket Server",
    description="WebSocket server for Claude agent conversations",
    version="1.0.0",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/")
async def root():
    """Health check endpoint."""
    return JSONResponse({
        "status": "ok",
        "service": "Claude Agent Server",
        "version": "1.0.0",
        "sessions": len(await session_manager.list_sessions())
    })


@app.get("/health")
async def health():
    """Health check endpoint."""
    return JSONResponse({"status": "healthy"})


class WebSocketConnection:
    """Manages a single WebSocket connection."""

    def __init__(self, websocket: WebSocket):
        self.websocket = websocket
        self.authenticated = False
        self.client_address = websocket.client.host

    async def send_json(self, data: dict):
        """Send JSON data to client."""
        await self.websocket.send_json(data)

    async def send_error(self, message: str, session_id=None, details=None):
        """Send error message to client."""
        error = ErrorMessage(
            message=message,
            session_id=session_id,
            details=details
        )
        await self.send_json(error.model_dump(mode='json'))

    async def handle_create_session(self, data: dict):
        """Handle create session request."""
        try:
            msg = CreateSessionMessage(**data)

            # Create session
            session = await session_manager.create_session(
                session_id=msg.session_id,
                options=msg.options
            )

            # Create agent
            await agent_manager.create_agent(session.id, msg.options)

            # Send response
            response = SessionCreatedMessage(
                session_id=session.id,
                session=session
            )
            await self.send_json(response.model_dump(mode='json'))

            logger.info(f"Created session {session.id} for {self.client_address}")

        except ValueError as e:
            await self.send_error(str(e))
        except Exception as e:
            logger.error(f"Error creating session: {e}")
            await self.send_error(f"Failed to create session: {e}")

    async def handle_send_prompt(self, data: dict):
        """Handle send prompt request."""
        try:
            msg = SendPromptMessage(**data)

            # Get session
            session = await session_manager.get_session(msg.session_id)
            if not session:
                await self.send_error(f"Session {msg.session_id} not found")
                return

            # Update session status
            await session_manager.update_session(
                msg.session_id,
                status=SessionStatus.PROCESSING
            )
            await session_manager.increment_message_count(msg.session_id)

            # Send prompt to agent and stream responses
            # Create a callback for permission requests and acknowledgments to send them directly
            async def send_permission_message(permission_data):
                """Send permission-related messages directly to WebSocket."""
                permission_data["session_id"] = str(msg.session_id)

                # Map message types
                if permission_data.get("type") == "permission_request":
                    permission_data["type"] = MessageType.PERMISSION_REQUEST.value
                elif permission_data.get("type") == "permission_acknowledged":
                    permission_data["type"] = MessageType.PERMISSION_ACKNOWLEDGED.value

                await self.send_json(permission_data)

            try:
                async for response in agent_manager.send_prompt(
                    msg.session_id,
                    msg.prompt,
                    send_message_callback=send_permission_message
                ):
                    # Add session_id to response
                    response["session_id"] = str(msg.session_id)

                    # Map response type
                    if response.get("type") == "agent_message":
                        response["type"] = MessageType.AGENT_MESSAGE.value
                    elif response.get("type") == "agent_thinking":
                        response["type"] = MessageType.AGENT_THINKING.value
                    elif response.get("type") == "agent_tool_use":
                        response["type"] = MessageType.AGENT_TOOL_USE.value
                    elif response.get("type") == "permission_request":
                        response["type"] = MessageType.PERMISSION_REQUEST.value
                    elif response.get("type") == "agent_error":
                        response["type"] = MessageType.AGENT_ERROR.value

                    await self.send_json(response)

                # Update session status back to idle
                await session_manager.update_session(
                    msg.session_id,
                    status=SessionStatus.IDLE
                )

            except Exception as e:
                logger.error(f"Error processing prompt: {e}")
                await session_manager.update_session(
                    msg.session_id,
                    status=SessionStatus.ERROR,
                    error_message=str(e)
                )
                await self.send_error(
                    f"Error processing prompt: {e}",
                    session_id=msg.session_id
                )

        except Exception as e:
            logger.error(f"Error handling prompt: {e}")
            await self.send_error(f"Failed to send prompt: {e}")

    async def handle_end_session(self, data: dict):
        """Handle end session request."""
        try:
            msg = EndSessionMessage(**data)

            # End agent
            await agent_manager.end_agent(msg.session_id)

            # End session
            ended = await session_manager.end_session(msg.session_id)

            if ended:
                response = SessionEndedMessage(session_id=msg.session_id)
                await self.send_json(response.model_dump(mode='json'))
                logger.info(f"Ended session {msg.session_id}")
            else:
                await self.send_error(f"Session {msg.session_id} not found")

        except Exception as e:
            logger.error(f"Error ending session: {e}")
            await self.send_error(f"Failed to end session: {e}")

    async def handle_list_sessions(self):
        """Handle list sessions request."""
        try:
            sessions = await session_manager.list_sessions()
            response = SessionsListMessage(sessions=sessions)
            await self.send_json(response.model_dump(mode='json'))

        except Exception as e:
            logger.error(f"Error listing sessions: {e}")
            await self.send_error(f"Failed to list sessions: {e}")

    async def handle_kill_all_agents(self, data: dict):
        """Handle kill all agents request."""
        try:
            msg = KillAllAgentsMessage(**data)

            # Kill all agents
            killed_count = await agent_manager.kill_all_agents()

            # Get list of ended sessions
            sessions_ended = list(agent_manager.active_agents.keys())

            # Send response
            response = AgentsKilledMessage(
                killed_count=killed_count,
                sessions_ended=sessions_ended
            )
            await self.send_json(response.model_dump(mode='json'))

            logger.info(f"Killed all {killed_count} agents for {self.client_address}")

        except Exception as e:
            logger.error(f"Error killing all agents: {e}")
            await self.send_error(f"Failed to kill all agents: {e}")

    async def handle_permission_response(self, data: dict):
        """Handle permission response request."""
        try:
            logger.debug(f"handle_permission_response called with data: {data}")
            msg = PermissionResponseMessage(**data)
            logger.info(f"Processing permission response for session {msg.session_id}, request {msg.request_id}, approved={msg.approved}")

            # Handle the permission response
            success = await agent_manager.handle_permission_response(
                msg.session_id,
                msg.request_id,
                msg.approved,
                msg.reason
            )

            if not success:
                await self.send_error(
                    f"Permission request {msg.request_id} not found",
                    session_id=msg.session_id
                )

            logger.info(f"Handled permission response for session {msg.session_id}: {msg.request_id} = {msg.approved}")

        except Exception as e:
            logger.error(f"Error handling permission response: {e}")
            await self.send_error(f"Failed to handle permission response: {e}")

    async def handle_ping(self, data: dict):
        """Handle ping request."""
        try:
            msg = PingMessage(**data)
            pong = PongMessage(timestamp=msg.timestamp)
            await self.send_json(pong.model_dump(mode='json'))
        except Exception:
            # Ignore ping errors
            pass

    async def handle_message(self, data: dict):
        """Route message to appropriate handler."""
        msg_type = data.get("type")

        if msg_type == MessageType.CREATE_SESSION.value:
            await self.handle_create_session(data)
        elif msg_type == MessageType.SEND_PROMPT.value:
            # Run prompt handling in background to not block permission responses
            asyncio.create_task(self.handle_send_prompt(data))
        elif msg_type == MessageType.END_SESSION.value:
            await self.handle_end_session(data)
        elif msg_type == MessageType.LIST_SESSIONS.value:
            await self.handle_list_sessions()
        elif msg_type == MessageType.KILL_ALL_AGENTS.value:
            await self.handle_kill_all_agents(data)
        elif msg_type == MessageType.PERMISSION_RESPONSE.value:
            await self.handle_permission_response(data)
        elif msg_type == MessageType.PING.value:
            await self.handle_ping(data)
        else:
            await self.send_error(f"Unknown message type: {msg_type}")


@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    """WebSocket endpoint for agent conversations."""
    await websocket.accept()
    connection = WebSocketConnection(websocket)

    # Authenticate connection
    if not await authenticate_websocket(websocket):
        await websocket.close(
            code=status.WS_1008_POLICY_VIOLATION,
            reason="Authentication required"
        )
        return

    connection.authenticated = True
    logger.info(f"WebSocket connection established from {connection.client_address}")

    try:
        while True:
            # Receive message
            try:
                data = await websocket.receive_text()
                message = json.loads(data)

                # Skip auth messages (already authenticated)
                if message.get("type") == MessageType.AUTH.value:
                    continue

                # Handle message
                await connection.handle_message(message)

            except json.JSONDecodeError as e:
                await connection.send_error(f"Invalid JSON: {e}")

    except WebSocketDisconnect:
        logger.info(f"WebSocket disconnected from {connection.client_address}")
    except Exception as e:
        logger.error(f"WebSocket error: {e}")
        await connection.send_error(f"Server error: {e}")
    finally:
        # Cleanup any sessions created by this connection
        # (In a production app, you'd track which sessions belong to which connection)
        pass


def main():
    """Run the server."""
    uvicorn.run(
        "src.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.reload,
        log_level=settings.log_level.lower()
    )


if __name__ == "__main__":
    main()