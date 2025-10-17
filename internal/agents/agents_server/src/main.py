"""Main FastAPI application with WebSocket endpoint for agent conversations."""

import asyncio
import hashlib
import json
import logging
from contextlib import asynccontextmanager
from typing import Dict

import uvicorn
from fastapi import FastAPI, WebSocket, WebSocketDisconnect, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from .agent_manager import agent_manager
from .agent_loader import get_agent_loader, initialize_agent_loader
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
from .persistence import persistence_client

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

    # Initialize agent loader
    logger.info("Initializing agent loader...")
    try:
        initialize_agent_loader()
        logger.info("Agent loader initialized")
    except Exception as e:
        logger.error(f"❌ Failed to initialize agent loader: {type(e).__name__}: {e}")
        raise

    # Initialize persistence client
    logger.info("Initializing persistence client...")
    try:
        await persistence_client.init()
        logger.info("Persistence client initialized")
    except Exception as e:
        logger.error(f"❌ Failed to initialize persistence client: {type(e).__name__}: {e}")
        logger.error("⚠️  Agent data persistence is disabled - messages will not be saved to database")
        # Don't raise - allow server to continue without persistence

    await session_manager.start()
    yield

    # Shutdown
    logger.info("Shutting down agent server...")
    await session_manager.stop()
    await persistence_client.close()
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


@app.get("/agents")
async def list_agents(working_directory: str = None):
    """List all available agents.

    Query Parameters:
        working_directory: Optional path to project directory to load project-specific agents
    """
    try:
        agent_loader = get_agent_loader(working_directory)
        agents = agent_loader.list_agents()
        logger.info(f"Listed {len(agents)} available agents from {agent_loader.agents_dir}")
        return JSONResponse({
            "status": "ok",
            "agents": agents,
            "count": len(agents),
            "agents_dir": str(agent_loader.agents_dir)
        })
    except Exception as e:
        logger.error(f"Error listing agents: {e}")
        return JSONResponse(
            {"status": "error", "message": str(e)},
            status_code=500
        )


@app.get("/agents/{agent_name}")
async def get_agent(agent_name: str, working_directory: str = None):
    """Get details about a specific agent, including full system prompt.

    Query Parameters:
        working_directory: Optional path to project directory to load project-specific agents
    """
    try:
        agent_loader = get_agent_loader(working_directory)
        agent = agent_loader.get_agent(agent_name)
        if not agent:
            logger.warning(f"Agent '{agent_name}' not found in {agent_loader.agents_dir}")
            return JSONResponse(
                {"status": "error", "message": f"Agent '{agent_name}' not found in {agent_loader.agents_dir}"},
                status_code=404
            )
        logger.info(f"Retrieved agent: {agent_name} from {agent_loader.agents_dir}")
        return JSONResponse({
            "status": "ok",
            "agent": agent.to_dict(include_system_prompt=True),
            "agents_dir": str(agent_loader.agents_dir)
        })
    except Exception as e:
        logger.error(f"Error getting agent '{agent_name}': {e}")
        return JSONResponse(
            {"status": "error", "message": str(e)},
            status_code=500
        )


class WebSocketConnection:
    """Manages a single WebSocket connection."""

    def __init__(self, websocket: WebSocket):
        self.websocket = websocket
        self.authenticated = False
        self.client_address = websocket.client.host
        self.pending_tasks = []  # Track background tasks

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
            logger.debug(f"Client {self.client_address}: ======== CREATE SESSION REQUEST ========")
            logger.debug(f"Client {self.client_address}: Request data: {data}")

            msg = CreateSessionMessage(**data)
            logger.debug(f"Client {self.client_address}: Parsed message - session_id: {msg.session_id}")
            logger.debug(f"Client {self.client_address}: Session options - tools: {[t.value for t in msg.options.tools]}")
            logger.debug(f"Client {self.client_address}: Permission mode: {msg.options.permission_mode}")

            # Create session
            logger.debug(f"Client {self.client_address}: Creating session...")
            session = await session_manager.create_session(
                session_id=msg.session_id,
                options=msg.options
            )
            logger.info(f"Client {self.client_address}: Session created: {session.id}")

            # Save session to database
            logger.debug(f"Client {self.client_address}: Saving session to database...")
            tools_list = [t.value for t in msg.options.tools] if msg.options.tools else []
            await persistence_client.save_session(
                session_id=session.id,
                session_name=f"Session {str(session.id)[:8]}",
                avatar_name="",
                working_directory=msg.options.working_directory or "",
                agent_name=msg.options.agent_name or "",
                system_prompt=msg.options.system_prompt or "",
                permission_mode=msg.options.permission_mode or "default",
                tools=tools_list
            )

            # Create agent
            logger.debug(f"Client {self.client_address}: Creating agent for session...")
            await agent_manager.create_agent(session.id, msg.options)
            logger.info(f"Client {self.client_address}: Agent created for session {session.id}")

            # Send response
            response = SessionCreatedMessage(
                session_id=session.id,
                session=session
            )
            logger.debug(f"Client {self.client_address}: Sending session created response")
            await self.send_json(response.model_dump(mode='json'))

            logger.info(f"Client {self.client_address}: ======== SESSION CREATED: {session.id} ========")

        except ValueError as e:
            logger.error(f"Client {self.client_address}: ValueError in create_session: {e}")
            await self.send_error(str(e))
        except Exception as e:
            logger.error(f"Client {self.client_address}: Exception in create_session: {type(e).__name__}: {e}")
            logger.error(f"Client {self.client_address}: Traceback: {__import__('traceback').format_exc()}")
            await self.send_error(f"Failed to create session: {e}")

    async def handle_send_prompt(self, data: dict):
        """Handle send prompt request."""
        try:
            logger.warning(f"Client {self.client_address}: 🟠 BACKGROUND TASK STARTED: handle_send_prompt")
            logger.debug(f"Client {self.client_address}: ======== SEND PROMPT REQUEST ========")
            logger.debug(f"Client {self.client_address}: Request data keys: {list(data.keys())}")

            msg = SendPromptMessage(**data)
            logger.info(f"Client {self.client_address}: Send prompt request - session: {msg.session_id}")
            logger.debug(f"Client {self.client_address}: Prompt: {msg.prompt[:100]}...")

            # Get session
            session = await session_manager.get_session(msg.session_id)
            if not session:
                logger.warning(f"Client {self.client_address}: Session {msg.session_id} not found")
                await self.send_error(f"Session {msg.session_id} not found")
                return

            logger.debug(f"Client {self.client_address}: Session found - status: {session.status}, message_count: {session.message_count}")

            # Save user message first so it has earliest timestamp
            logger.info(f"Client {self.client_address}: 💾 Saving user prompt to database")
            await persistence_client.save_message(
                session_id=msg.session_id,
                message_id="user-prompt",
                role="user",
                content=msg.prompt
            )
            logger.info(f"Client {self.client_address}: ✓ Saved user prompt")

            # Update session status
            logger.debug(f"Client {self.client_address}: Updating session status to PROCESSING")
            await session_manager.update_session(
                msg.session_id,
                status=SessionStatus.PROCESSING
            )
            await session_manager.increment_message_count(msg.session_id)
            logger.debug(f"Client {self.client_address}: Session status updated and message count incremented")

            # Send prompt to agent and stream responses
            # Create a callback for permission requests and acknowledgments to send them directly
            response_count = 0

            async def send_permission_message(permission_data):
                """Send permission-related messages directly to WebSocket."""
                logger.debug(f"Client {self.client_address}: send_permission_message called")
                logger.debug(f"Client {self.client_address}: Permission message type: {permission_data.get('type')}")

                permission_data["session_id"] = str(msg.session_id)

                # Map message types
                if permission_data.get("type") == "permission_request":
                    permission_data["type"] = MessageType.PERMISSION_REQUEST.value
                    logger.info(f"Client {self.client_address}: Sending permission request for tool: {permission_data.get('tool')}")
                elif permission_data.get("type") == "permission_acknowledged":
                    permission_data["type"] = MessageType.PERMISSION_ACKNOWLEDGED.value
                    logger.info(f"Client {self.client_address}: Sending permission acknowledgment: {permission_data.get('request_id')} -> {permission_data.get('approved')}")

                logger.debug(f"Client {self.client_address}: Sending permission message via WebSocket")
                await self.send_json(permission_data)

            try:
                logger.debug(f"Client {self.client_address}: Starting to iterate over agent responses...")
                logger.info(f"Client {self.client_address}: 🔵 ENTERING RESPONSE LOOP")

                # Track seen message content hashes to prevent duplicates
                seen_message_hashes = set()

                async for response in agent_manager.send_prompt(
                    msg.session_id,
                    msg.prompt,
                    send_message_callback=send_permission_message
                ):
                    response_count += 1
                    logger.info(f"Client {self.client_address}: 🟢 GOT RESPONSE #{response_count}")
                    logger.debug(f"Client {self.client_address}: ======== RESPONSE #{response_count} ========")
                    resp_type = response.get('type')
                    logger.info(f"Client {self.client_address}: Response type: {resp_type} (keys: {list(response.keys())})")

                    # Add session_id to response
                    response["session_id"] = str(msg.session_id)

                    # Map response type and save messages
                    if resp_type == "agent_message":
                        response["type"] = MessageType.AGENT_MESSAGE.value
                        content = response.get("content", "")
                        content_preview = content[:80]
                        logger.debug(f"Client {self.client_address}: Agent message ({len(content)} chars): {content_preview}...")

                        # Check for duplicate messages (same role and content)
                        message_hash = hashlib.md5(f"assistant:{content}".encode()).hexdigest()
                        if message_hash in seen_message_hashes:
                            logger.warning(f"Client {self.client_address}: ⚠️  DUPLICATE agent_message detected - skipping save")
                            logger.debug(f"Client {self.client_address}: Hash: {message_hash}, Content length: {len(content)}")
                        else:
                            seen_message_hashes.add(message_hash)
                            # Save message to database
                            logger.info(f"Client {self.client_address}: 💾 Attempting to save assistant message msg-{response_count}")
                            await persistence_client.save_message(
                                session_id=msg.session_id,
                                message_id=f"msg-{response_count}",
                                role="assistant",
                                content=content,
                                token_count=response.get("token_count", 0)
                            )
                            logger.info(f"Client {self.client_address}: ✓ Saved assistant message msg-{response_count}")
                    elif resp_type == "agent_thinking":
                        response["type"] = MessageType.AGENT_THINKING.value
                        logger.debug(f"Client {self.client_address}: Thinking: {response.get('thinking')}")
                    elif resp_type == "agent_tool_use":
                        response["type"] = MessageType.AGENT_TOOL_USE.value
                        logger.info(f"Client {self.client_address}: Tool use: {response.get('tool')} with ID: {response.get('tool_use_id')}")

                        # Check for duplicate tool use messages
                        tool_use_id = response.get('tool_use_id', f'tool-{response_count}')
                        tool_content = f"Using tool: {response.get('tool')}"
                        tool_hash = hashlib.md5(f"assistant:{tool_content}:{tool_use_id}".encode()).hexdigest()

                        if tool_hash not in seen_message_hashes:
                            seen_message_hashes.add(tool_hash)
                            # Save tool use message to database
                            await persistence_client.save_message(
                                session_id=msg.session_id,
                                message_id=tool_use_id,
                                role="assistant",
                                content=tool_content,
                                tool_name=response.get('tool', '')
                            )
                        else:
                            logger.warning(f"Client {self.client_address}: ⚠️  DUPLICATE agent_tool_use detected - skipping save")
                    elif resp_type == "permission_request":
                        response["type"] = MessageType.PERMISSION_REQUEST.value
                        logger.info(f"Client {self.client_address}: Permission request: {response.get('request_id')}")
                    elif resp_type == "agent_error":
                        response["type"] = MessageType.AGENT_ERROR.value
                        logger.error(f"Client {self.client_address}: Agent error: {response.get('message')}")

                    logger.debug(f"Client {self.client_address}: Sending response via WebSocket")
                    await self.send_json(response)

                logger.info(f"Client {self.client_address}: Response stream complete ({response_count} responses)")

                # Update session status back to idle
                logger.debug(f"Client {self.client_address}: Updating session status back to IDLE")
                await session_manager.update_session(
                    msg.session_id,
                    status=SessionStatus.IDLE
                )
                logger.info(f"Client {self.client_address}: ======== PROMPT HANDLING COMPLETE ========")

            except Exception as e:
                logger.error(f"Client {self.client_address}: Exception processing prompt: {type(e).__name__}: {e}")
                logger.error(f"Client {self.client_address}: Responses sent before error: {response_count}")
                logger.error(f"Client {self.client_address}: Traceback: {__import__('traceback').format_exc()}")

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
            logger.error(f"Client {self.client_address}: Exception handling prompt: {type(e).__name__}: {e}")
            logger.error(f"Client {self.client_address}: Traceback: {__import__('traceback').format_exc()}")
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
        logger.debug(f"Client {self.client_address}: ======== MESSAGE ROUTING ========")
        logger.debug(f"Client {self.client_address}: Message type to route: {msg_type}")

        if msg_type == MessageType.CREATE_SESSION.value:
            logger.debug(f"Client {self.client_address}: Routing to handle_create_session")
            await self.handle_create_session(data)
        elif msg_type == MessageType.SEND_PROMPT.value:
            logger.warning(f"Client {self.client_address}: 📍 ABOUT TO CREATE BACKGROUND TASK for send_prompt")
            logger.debug(f"Client {self.client_address}: Routing to handle_send_prompt (background task)")
            # Run prompt handling in background to not block permission responses
            logger.warning(f"Client {self.client_address}: 🔨 CREATING TASK NOW...")
            task = asyncio.create_task(self.handle_send_prompt(data))
            self.pending_tasks.append(task)  # Track the task
            task.add_done_callback(lambda t: self.pending_tasks.remove(t) if t in self.pending_tasks else None)
            logger.warning(f"Client {self.client_address}: ✅ TASK CREATED AND TRACKED (total pending: {len(self.pending_tasks)})")
        elif msg_type == MessageType.END_SESSION.value:
            logger.debug(f"Client {self.client_address}: Routing to handle_end_session")
            await self.handle_end_session(data)
        elif msg_type == MessageType.LIST_SESSIONS.value:
            logger.debug(f"Client {self.client_address}: Routing to handle_list_sessions")
            await self.handle_list_sessions()
        elif msg_type == MessageType.KILL_ALL_AGENTS.value:
            logger.debug(f"Client {self.client_address}: Routing to handle_kill_all_agents")
            await self.handle_kill_all_agents(data)
        elif msg_type == MessageType.PERMISSION_RESPONSE.value:
            logger.debug(f"Client {self.client_address}: Routing to handle_permission_response")
            await self.handle_permission_response(data)
        elif msg_type == MessageType.PING.value:
            logger.debug(f"Client {self.client_address}: Routing to handle_ping (no-op)")
            await self.handle_ping(data)
        else:
            logger.warning(f"Client {self.client_address}: Unknown message type: {msg_type}")
            await self.send_error(f"Unknown message type: {msg_type}")


@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    """WebSocket endpoint for agent conversations."""
    logger.debug("======== NEW WEBSOCKET CONNECTION ========")
    await websocket.accept()
    connection = WebSocketConnection(websocket)
    logger.info(f"WebSocket accepted from {connection.client_address}")

    # Authenticate connection
    logger.debug(f"Client {connection.client_address}: Authenticating WebSocket...")
    if not await authenticate_websocket(websocket):
        logger.warning(f"Client {connection.client_address}: Authentication failed")
        await websocket.close(
            code=status.WS_1008_POLICY_VIOLATION,
            reason="Authentication required"
        )
        return

    connection.authenticated = True
    logger.info(f"Client {connection.client_address}: ======== AUTHENTICATED ========")

    message_count = 0
    try:
        while True:
            # Receive message
            try:
                message_count += 1
                logger.debug(f"Client {connection.client_address}: Waiting for message #{message_count}...")
                data = await websocket.receive_text()
                logger.debug(f"Client {connection.client_address}: Received raw text: {len(data)} bytes")

                message = json.loads(data)
                msg_type = message.get("type")
                logger.debug(f"Client {connection.client_address}: ======== MESSAGE #{message_count} ========")
                logger.debug(f"Client {connection.client_address}: Message type: {msg_type}")

                # Skip auth messages (already authenticated)
                if msg_type == MessageType.AUTH.value:
                    logger.debug(f"Client {connection.client_address}: Skipping AUTH message (already authenticated)")
                    continue

                # Handle message
                logger.debug(f"Client {connection.client_address}: Routing to handler for message type: {msg_type}")
                await connection.handle_message(message)

            except json.JSONDecodeError as e:
                logger.error(f"Client {connection.client_address}: JSON decode error: {e}")
                logger.error(f"Client {connection.client_address}: Raw data: {data[:200]}")
                await connection.send_error(f"Invalid JSON: {e}")

    except WebSocketDisconnect:
        logger.info(f"Client {connection.client_address}: ======== WEBSOCKET DISCONNECTED ========")
        logger.debug(f"Client {connection.client_address}: Total messages received: {message_count}")
        # Wait for pending tasks to complete before closing
        if connection.pending_tasks:
            logger.info(f"Client {connection.client_address}: Waiting for {len(connection.pending_tasks)} pending task(s) to complete...")
            try:
                await asyncio.gather(*connection.pending_tasks, return_exceptions=True)
                logger.info(f"Client {connection.client_address}: All pending tasks completed")
            except Exception as e:
                logger.error(f"Client {connection.client_address}: Error waiting for pending tasks: {e}")
    except Exception as e:
        logger.error(f"Client {connection.client_address}: ======== WEBSOCKET ERROR ========")
        logger.error(f"Client {connection.client_address}: Error type: {type(e).__name__}")
        logger.error(f"Client {connection.client_address}: Error message: {e}")
        logger.error(f"Client {connection.client_address}: Messages processed: {message_count}")
        logger.error(f"Client {connection.client_address}: Traceback: {__import__('traceback').format_exc()}")
        await connection.send_error(f"Server error: {e}")
    finally:
        # Cleanup any sessions created by this connection
        # (In a production app, you'd track which sessions belong to which connection)
        logger.debug(f"Client {connection.client_address}: WebSocket connection cleanup complete")
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