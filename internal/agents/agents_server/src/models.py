"""Pydantic models for WebSocket messages and agent sessions."""

from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional
from uuid import UUID, uuid4

from pydantic import BaseModel, Field


class MessageType(str, Enum):
    """WebSocket message types."""

    # Authentication
    AUTH = "auth"
    AUTH_SUCCESS = "auth_success"

    # Session management
    CREATE_SESSION = "create_session"
    SESSION_CREATED = "session_created"
    END_SESSION = "end_session"
    SESSION_ENDED = "session_ended"
    LIST_SESSIONS = "list_sessions"
    SESSIONS_LIST = "sessions_list"

    # Agent interaction
    SEND_PROMPT = "send_prompt"
    AGENT_MESSAGE = "agent_message"
    AGENT_THINKING = "agent_thinking"
    AGENT_TOOL_USE = "agent_tool_use"
    AGENT_ERROR = "agent_error"

    # Permission requests
    PERMISSION_REQUEST = "permission_request"
    PERMISSION_RESPONSE = "permission_response"
    PERMISSION_ACKNOWLEDGED = "permission_acknowledged"

    # Kill switch
    KILL_ALL_AGENTS = "kill_all_agents"
    AGENTS_KILLED = "agents_killed"

    # System
    ERROR = "error"
    PING = "ping"
    PONG = "pong"


class Tool(str, Enum):
    """Available Claude Code tools."""

    READ = "Read"
    WRITE = "Write"
    EDIT = "Edit"
    BASH = "Bash"
    SEARCH = "Search"
    TASK = "Task"
    TODO_WRITE = "TodoWrite"
    WEB_SEARCH = "WebSearch"
    WEB_FETCH = "WebFetch"


class SessionOptions(BaseModel):
    """Options for creating an agent session."""

    system_prompt: Optional[str] = None
    tools: List[Tool] = Field(
        default_factory=lambda: [
            Tool.READ,
            Tool.WRITE,
            Tool.EDIT,
            Tool.BASH,
            Tool.SEARCH,
        ]
    )
    working_directory: Optional[str] = None
    max_tokens: Optional[int] = None
    temperature: Optional[float] = None
    permission_mode: Optional[str] = "default"  # Permission mode: default, allow-all, read-only
    conversation_history: Optional[str] = None  # Formatted conversation history for resume
    original_conversation_id: Optional[str] = None  # ID of the conversation being resumed


class SessionStatus(str, Enum):
    """Agent session status."""

    ACTIVE = "active"
    IDLE = "idle"
    PROCESSING = "processing"
    ERROR = "error"
    ENDED = "ended"


class AgentSession(BaseModel):
    """Represents an agent conversation session."""

    id: UUID = Field(default_factory=uuid4)
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)
    status: SessionStatus = SessionStatus.IDLE
    options: SessionOptions
    message_count: int = 0
    error_message: Optional[str] = None

    class Config:
        json_encoders = {
            UUID: str,
            datetime: lambda v: v.isoformat()
        }


# WebSocket message models
class BaseMessage(BaseModel):
    """Base WebSocket message."""

    type: MessageType

    class Config:
        json_encoders = {
            UUID: str,
            datetime: lambda v: v.isoformat()
        }


class AuthMessage(BaseMessage):
    """Authentication request."""

    type: MessageType = MessageType.AUTH
    token: str


class CreateSessionMessage(BaseMessage):
    """Create a new agent session."""

    type: MessageType = MessageType.CREATE_SESSION
    session_id: Optional[UUID] = Field(default_factory=uuid4)
    options: SessionOptions = Field(default_factory=SessionOptions)


class SessionCreatedMessage(BaseMessage):
    """Session creation response."""

    type: MessageType = MessageType.SESSION_CREATED
    session_id: UUID
    session: AgentSession


class SendPromptMessage(BaseMessage):
    """Send a prompt to an agent session."""

    type: MessageType = MessageType.SEND_PROMPT
    session_id: UUID
    prompt: str


class AgentMessage(BaseMessage):
    """Message from the agent."""

    type: MessageType = MessageType.AGENT_MESSAGE
    session_id: UUID
    content: str
    complete: bool = False
    message_id: Optional[str] = None


class AgentThinkingMessage(BaseMessage):
    """Agent is thinking/processing."""

    type: MessageType = MessageType.AGENT_THINKING
    session_id: UUID
    thinking: bool = True


class AgentToolUseMessage(BaseMessage):
    """Agent is using a tool."""

    type: MessageType = MessageType.AGENT_TOOL_USE
    session_id: UUID
    tool: str
    parameters: Dict[str, Any]
    result: Optional[Any] = None


class PermissionRequestMessage(BaseMessage):
    """Agent is requesting permission for an action."""

    type: MessageType = MessageType.PERMISSION_REQUEST
    session_id: UUID
    tool: str
    parameters: Dict[str, Any]
    description: str
    request_id: str = Field(default_factory=lambda: str(uuid4()))


class PermissionResponseMessage(BaseMessage):
    """User response to a permission request."""

    type: MessageType = MessageType.PERMISSION_RESPONSE
    session_id: UUID
    request_id: str
    approved: bool
    reason: Optional[str] = None


class PermissionAcknowledgedMessage(BaseMessage):
    """Acknowledgment that permission was received and tool is executing."""

    type: MessageType = MessageType.PERMISSION_ACKNOWLEDGED
    session_id: UUID
    request_id: str
    approved: bool
    tool: str
    status: str  # "executing" or "denied"


class EndSessionMessage(BaseMessage):
    """End an agent session."""

    type: MessageType = MessageType.END_SESSION
    session_id: UUID


class SessionEndedMessage(BaseMessage):
    """Session ended response."""

    type: MessageType = MessageType.SESSION_ENDED
    session_id: UUID


class ListSessionsMessage(BaseMessage):
    """Request list of active sessions."""

    type: MessageType = MessageType.LIST_SESSIONS


class SessionsListMessage(BaseMessage):
    """List of active sessions."""

    type: MessageType = MessageType.SESSIONS_LIST
    sessions: List[AgentSession]


class ErrorMessage(BaseMessage):
    """Error message."""

    type: MessageType = MessageType.ERROR
    message: str
    session_id: Optional[UUID] = None
    details: Optional[Dict[str, Any]] = None


class PingMessage(BaseMessage):
    """Ping message for keepalive."""

    type: MessageType = MessageType.PING
    timestamp: datetime = Field(default_factory=datetime.now)


class PongMessage(BaseMessage):
    """Pong response."""

    type: MessageType = MessageType.PONG
    timestamp: datetime = Field(default_factory=datetime.now)


class KillAllAgentsMessage(BaseMessage):
    """Kill all active agents."""

    type: MessageType = MessageType.KILL_ALL_AGENTS


class AgentsKilledMessage(BaseMessage):
    """Response when all agents are killed."""

    type: MessageType = MessageType.AGENTS_KILLED
    killed_count: int
    sessions_ended: List[UUID]