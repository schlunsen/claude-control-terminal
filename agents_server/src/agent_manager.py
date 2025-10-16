"""Claude Agent SDK integration and management."""

import asyncio
import logging
import os
import uuid
from pathlib import Path
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
if os.environ.get('ANTHROPIC_API_KEY'):
    logger.info("ANTHROPIC_API_KEY found in environment")
else:
    logger.warning("ANTHROPIC_API_KEY not set in environment - Claude SDK may fail to initialize")


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

            # Build enhanced system prompt with history if provided
            system_prompt = options.system_prompt or "You are a helpful AI assistant."
            if options.conversation_history:
                system_prompt = f"""
{system_prompt}

You are resuming an existing coding session.
Working Directory: {options.working_directory or 'Not specified'}

Recent conversation context:
{options.conversation_history}

Continue helping the user from where they left off. You have access to the project files and can use tools as needed.
"""

            # Create agent options
            agent_options = ClaudeAgentOptions(
                system_prompt=system_prompt,
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
        logger.info(f"Sending prompt to session {session_id}: {prompt[:50]}...")

        # Get the client for this session
        client = self.active_agents.get(session_id)
        if not client:
            raise ValueError(f"No agent found for session {session_id}")

        try:
            # Send thinking indicator immediately
            logger.debug(f"Session {session_id}: Sending thinking indicator")
            yield {
                "type": "agent_thinking",
                "thinking": True
            }

            # Send the prompt to the client
            await client.query(prompt)
            logger.debug(f"Session {session_id}: Prompt sent, awaiting response")

            # Track message for streaming
            current_message_id = str(uuid.uuid4())  # Generate unique ID for this message
            message_buffer = []
            seen_content = set()
            thinking_sent = False
            message_count = 0

            # Receive and process responses with timeout
            try:
                async for message in client.receive_response():
                    message_count += 1
                    logger.debug(f"Session {session_id}: Received message #{message_count} type: {type(message).__name__}")
                    logger.debug(f"Session {session_id}: Message content: {message}")

                    if isinstance(message, AssistantMessage):
                        # Turn off thinking indicator when we get the first message
                        if not thinking_sent:
                            logger.debug(f"Session {session_id}: Turning off thinking indicator")
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
                                logger.debug(f"Session {session_id}: Yielding content chunk: {content[:50]}...")
                                yield {
                                    "type": "agent_message",
                                    "content": content,
                                    "complete": False,
                                    "message_id": current_message_id
                                }
                            else:
                                logger.debug(f"Session {session_id}: Skipping duplicate content")
                    elif isinstance(message, dict):
                        # Process dict messages but check for duplicates
                        processed = self._process_agent_message(message)
                        if processed.get("type") == "agent_message":
                            # Turn off thinking indicator when we get actual content
                            if not thinking_sent:
                                logger.debug(f"Session {session_id}: Turning off thinking indicator (dict)")
                                yield {
                                    "type": "agent_thinking",
                                    "thinking": False
                                }
                                thinking_sent = True
                            content = processed.get("content", "")
                            if content and content not in seen_content:
                                seen_content.add(content)
                                logger.debug(f"Session {session_id}: Yielding processed content: {content[:50]}...")
                                processed["message_id"] = current_message_id
                                yield processed
                            else:
                                logger.debug(f"Session {session_id}: Skipping duplicate processed content")
                        elif processed.get("type") != "agent_message":
                            # Non-message types (thinking, tool use, etc)
                            logger.debug(f"Session {session_id}: Yielding non-message type: {processed.get('type')}")
                            yield processed
                    elif hasattr(message, '__class__') and 'SystemMessage' in message.__class__.__name__:
                        # Skip system messages - they're internal to the SDK
                        logger.debug(f"Session {session_id}: Skipping system message: {type(message).__name__}")
                        continue
                    else:
                        # Skip other unknown message types - they often contain duplicates
                        logger.debug(f"Session {session_id}: Skipping unknown message type: {type(message).__name__}")

                # Send completion signal and ensure thinking is off
                if not thinking_sent:
                    logger.debug(f"Session {session_id}: Sending final thinking off signal")
                    yield {
                        "type": "agent_thinking",
                        "thinking": False
                    }

                if message_buffer:
                    # Send the accumulated message buffer as the final message
                    final_content = "".join(message_buffer)
                    logger.info(f"Session {session_id}: Sending completion message with {len(final_content)} chars (processed {message_count} messages)")
                    yield {
                        "type": "agent_message",
                        "content": final_content,
                        "complete": True,
                        "message_id": current_message_id
                    }
                else:
                    logger.warning(f"Session {session_id}: No content to send in completion (received {message_count} messages)")

            except Exception as e:
                logger.error(f"Session {session_id}: Error processing response: {e}")
                logger.error(f"Session {session_id}: Received {message_count} messages before error")
                yield {
                    "type": "agent_error",
                    "message": f"Error processing response: {str(e)}",
                    "session_id": str(session_id)
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

    async def kill_all_agents(self) -> int:
        """
        Kill all active agents immediately.

        Returns:
            Number of agents killed
        """
        killed_count = 0
        session_ids = list(self.active_agents.keys())

        logger.info(f"Killing all {len(session_ids)} active agents")

        for session_id in session_ids:
            try:
                await self.end_agent(session_id)
                killed_count += 1
            except Exception as e:
                logger.error(f"Error killing agent for session {session_id}: {e}")

        logger.info(f"Successfully killed {killed_count} agents")
        return killed_count

    async def cleanup_all(self):
        """Cleanup all active agents."""
        session_ids = list(self.active_agents.keys())
        for session_id in session_ids:
            await self.end_agent(session_id)

        logger.info("Cleaned up all agents")


# Global agent manager instance
agent_manager = AgentManager()