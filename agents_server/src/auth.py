"""Authentication middleware for WebSocket connections."""

import logging
from typing import Optional

from fastapi import WebSocket, WebSocketDisconnect, status
from pydantic import BaseModel

from .config import get_api_key, settings

logger = logging.getLogger(__name__)


class AuthMessage(BaseModel):
    """Authentication message model."""

    type: str
    token: str


async def authenticate_websocket(websocket: WebSocket) -> bool:
    """
    Authenticate a WebSocket connection.

    Can authenticate via:
    1. Query parameter: ws://host/ws?token=<api_key>
    2. First message: {"type": "auth", "token": "<api_key>"}

    Args:
        websocket: The WebSocket connection to authenticate

    Returns:
        True if authentication successful, False otherwise
    """
    if not settings.auth_enabled:
        logger.info("Authentication disabled, allowing connection")
        return True

    # Get the expected API key
    expected_key = get_api_key(settings)
    if not expected_key:
        logger.error("No API key configured, rejecting connection")
        return False

    # Check query parameter first
    token = websocket.query_params.get("token")
    if token and token == expected_key:
        logger.info(f"WebSocket authenticated via query parameter from {websocket.client.host}")
        return True

    # If no query param, wait for auth message
    try:
        # Set a timeout for receiving the auth message
        data = await websocket.receive_json()

        # Validate it's an auth message
        if data.get("type") != "auth":
            logger.warning(f"Expected auth message, got type: {data.get('type')}")
            await websocket.send_json({
                "type": "error",
                "message": "First message must be authentication"
            })
            return False

        # Check the token
        token = data.get("token", "").replace("Bearer ", "").strip()
        if token == expected_key:
            logger.info(f"WebSocket authenticated via message from {websocket.client.host}")
            await websocket.send_json({"type": "auth_success"})
            return True
        else:
            logger.warning(f"Invalid token from {websocket.client.host}")
            await websocket.send_json({
                "type": "error",
                "message": "Invalid authentication token"
            })
            return False

    except WebSocketDisconnect:
        logger.info("Client disconnected during authentication")
        return False
    except Exception as e:
        logger.error(f"Authentication error: {e}")
        await websocket.send_json({
            "type": "error",
            "message": "Authentication failed"
        })
        return False


async def require_auth(websocket: WebSocket) -> Optional[WebSocket]:
    """
    Decorator/helper to require authentication for a WebSocket endpoint.

    Usage:
        @app.websocket("/ws")
        async def websocket_endpoint(websocket: WebSocket):
            await websocket.accept()
            if not await authenticate_websocket(websocket):
                await websocket.close(code=status.WS_1008_POLICY_VIOLATION)
                return
            # Proceed with authenticated connection

    Args:
        websocket: The WebSocket connection

    Returns:
        The WebSocket if authenticated, None otherwise
    """
    await websocket.accept()

    if not await authenticate_websocket(websocket):
        await websocket.close(
            code=status.WS_1008_POLICY_VIOLATION,
            reason="Authentication required"
        )
        return None

    return websocket