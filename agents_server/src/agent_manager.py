"""Claude Agent SDK integration and management."""

import asyncio
import logging
import os
from typing import AsyncIterator, Dict, Optional
from uuid import UUID

from claude_agent_sdk import (
    ClaudeSDKClient,
    ClaudeAgentOptions,
    AssistantMessage,
    TextBlock,
)

from .config import settings
from .models import SessionOptions, Tool

logger = logging.getLogger(__name__)

# Ensure ANTHROPIC_API_KEY is set for Claude SDK
if not os.environ.get('ANTHROPIC_API_KEY'):
    # Try to get from config or use a placeholder
    api_key = getattr(settings, 'anthropic_api_key', None)
    if api_key:
        os.environ['ANTHROPIC_API_KEY'] = api_key
    else:
        logger.warning("ANTHROPIC_API_KEY not set - Claude SDK may fail to initialize")


class AgentManager:
    """Manages Claude Agent SDK interactions."""

    def __init__(self):
        self.active_agents: Dict[UUID, ClaudeSDKClient] = {}
        self.agent_options: Dict[UUID, ClaudeAgentOptions] = {}

    async def create_agent(
        self,
        session_id: UUID,
        options: SessionOptions
    ) -> ClaudeSDKClient:
        """
        Create a new Claude agent for a session.

        Args:
            session_id: The session ID
            options: Session configuration options

        Returns:
            The created ClaudeSDKClient

        Raises:
            RuntimeError: If agent creation fails
        """
        try:
            # Convert our Tool enum to string list for Claude SDK
            tools = [tool.value for tool in options.tools]

            # Create agent options
            agent_options = ClaudeAgentOptions(
                system_prompt=options.system_prompt,
                allowed_tools=tools,
                cwd=options.working_directory,
                permission_mode='default',  # Default permission mode
            )

            # Create and connect the client using async context manager
            client = ClaudeSDKClient(options=agent_options)

            # Connect the client (enter the async context)
            await client.__aenter__()

            # Store both client and options
            self.active_agents[session_id] = client
            self.agent_options[session_id] = agent_options

            logger.info(f"Created and connected agent for session {session_id}")

            return client

        except Exception as e:
            logger.error(f"Failed to create agent for session {session_id}: {e}")
            raise RuntimeError(f"Failed to create agent: {e}")

    async def send_prompt(
        self,
        session_id: UUID,
        prompt: str
    ) -> AsyncIterator[Dict]:
        """
        Send a prompt to an agent and yield responses.

        Args:
            session_id: The session ID
            prompt: The prompt to send

        Yields:
            Response chunks from the agent

        Raises:
            ValueError: If session not found
            RuntimeError: If prompt sending fails
        """
        # Get the client for this session
        client = self.active_agents.get(session_id)
        if not client:
            raise ValueError(f"No agent found for session {session_id}")

        try:
            # Send thinking indicator immediately
            yield {
                "type": "agent_thinking",
                "thinking": True
            }

            # Send the prompt to the client
            await client.query(prompt)

            # Track message for streaming
            current_message_id = None
            message_buffer = []
            seen_content = set()
            thinking_sent = False

            # Receive and process responses
            async for message in client.receive_response():
                logger.debug(f"Received message type: {type(message).__name__}")

                if isinstance(message, AssistantMessage):
                    # Turn off thinking indicator when we get the first message
                    if not thinking_sent:
                        yield {
                            "type": "agent_thinking",
                            "thinking": False
                        }
                        thinking_sent = True

                    # Extract text from assistant message
                    content_parts = []
                    for block in message.content:
                        if isinstance(block, TextBlock):
                            content_parts.append(block.text)

                    if content_parts:
                        content = "".join(content_parts)
                        # Avoid sending duplicate content
                        if content not in seen_content:
                            seen_content.add(content)
                            message_buffer.append(content)
                            yield {
                                "type": "agent_message",
                                "content": content,
                                "complete": False
                            }
                elif isinstance(message, dict):
                    # Process dict messages but check for duplicates
                    processed = self._process_agent_message(message)
                    if processed.get("type") == "agent_message":
                        # Turn off thinking indicator when we get actual content
                        if not thinking_sent:
                            yield {
                                "type": "agent_thinking",
                                "thinking": False
                            }
                            thinking_sent = True
                        content = processed.get("content", "")
                        if content and content not in seen_content:
                            seen_content.add(content)
                            yield processed
                    elif processed.get("type") != "agent_message":
                        # Non-message types (thinking, tool use, etc)
                        yield processed
                elif hasattr(message, '__class__') and 'SystemMessage' in message.__class__.__name__:
                    # Skip system messages - they're internal to the SDK
                    logger.debug(f"Skipping system message: {type(message).__name__}")
                    continue
                else:
                    # Skip other unknown message types - they often contain duplicates
                    logger.debug(f"Skipping unknown message type: {type(message).__name__}")

            # Send completion signal and ensure thinking is off
            if not thinking_sent:
                yield {
                    "type": "agent_thinking",
                    "thinking": False
                }

            if message_buffer:
                yield {
                    "type": "agent_message",
                    "content": "",
                    "complete": True
                }

        except Exception as e:
            logger.error(f"Error sending prompt to session {session_id}: {e}")
            raise RuntimeError(f"Failed to send prompt: {e}")

    def _process_agent_message(self, message: Dict) -> Dict:
        """
        Process and normalize agent messages.

        Args:
            message: Raw message from Claude SDK

        Returns:
            Normalized message dict
        """
        msg_type = message.get("type", "text")

        if msg_type == "text":
            return {
                "type": "agent_message",
                "content": message.get("content", ""),
                "complete": message.get("complete", False)
            }
        elif msg_type == "thinking":
            return {
                "type": "agent_thinking",
                "thinking": True
            }
        elif msg_type == "tool_use":
            return {
                "type": "agent_tool_use",
                "tool": message.get("tool", ""),
                "parameters": message.get("parameters", {}),
                "result": message.get("result")
            }
        elif msg_type == "error":
            return {
                "type": "agent_error",
                "message": message.get("error", "Unknown error")
            }
        else:
            # Unknown message type, pass through
            return {
                "type": "agent_message",
                "content": str(message),
                "complete": False
            }

    async def end_agent(self, session_id: UUID) -> bool:
        """
        End an agent session and cleanup resources.

        Args:
            session_id: The session ID

        Returns:
            True if agent was ended, False if not found
        """
        client = self.active_agents.get(session_id)
        if not client:
            return False

        try:
            # Close the client if it has a close method
            if hasattr(client, '__aexit__'):
                await client.__aexit__(None, None, None)
            elif hasattr(client, 'close'):
                await client.close()

            # Remove from active agents and options
            del self.active_agents[session_id]
            if session_id in self.agent_options:
                del self.agent_options[session_id]

            logger.info(f"Ended agent for session {session_id}")
            return True

        except Exception as e:
            logger.error(f"Error ending agent for session {session_id}: {e}")
            # Still remove from active agents
            if session_id in self.active_agents:
                del self.active_agents[session_id]
            if session_id in self.agent_options:
                del self.agent_options[session_id]
            return False

    async def cleanup_all(self):
        """Cleanup all active agents."""
        session_ids = list(self.active_agents.keys())
        for session_id in session_ids:
            await self.end_agent(session_id)

        logger.info("Cleaned up all agents")


# Global agent manager instance
agent_manager = AgentManager()