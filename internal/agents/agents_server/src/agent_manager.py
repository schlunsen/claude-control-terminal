"""Claude Agent SDK integration and management."""

import asyncio
import logging
import os
import uuid
from pathlib import Path
from typing import Any, AsyncIterator, Dict, Optional
from uuid import UUID

from claude_agent_sdk import (
    ClaudeSDKClient,
    ClaudeAgentOptions,
    AssistantMessage,
    TextBlock,
    ToolUseBlock,
    HookMatcher,
)

from .config import settings
from .models import SessionOptions, Tool, MCPServerConfig

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
        self.pending_permissions: Dict[UUID, Dict[str, any]] = {}  # session_id -> {request_id: permission_data}
        self.permission_futures: Dict[str, asyncio.Future] = {}  # request_id -> Future[bool]
        self.permission_callbacks: Dict[UUID, Any] = {}  # session_id -> callback for sending permission requests
        self.mcp_server_configs: Dict[UUID, Dict[str, MCPServerConfig]] = {}  # session_id -> {server_name: config}

    def _create_pretool_hook(self, session_id: UUID):
        """
        Create a PreToolUse hook for permission handling.

        Args:
            session_id: The session ID

        Returns:
            Async hook function that handles permission requests
        """
        async def pretool_hook(input_data, tool_use_id, context):
            """Hook called before any tool execution."""
            tool_name = input_data.get("tool_name", "")
            tool_input = input_data.get("tool_input", {})

            logger.debug(f"Session {session_id}: PreToolUse hook called for {tool_name}")

            # Check if this tool requires permission
            requires_perm = self._requires_permission(session_id, tool_name, tool_input)
            logger.debug(f"Session {session_id}: Tool {tool_name} requires_permission={requires_perm}")

            if not requires_perm:
                # No permission needed, allow execution
                logger.debug(f"Session {session_id}: Allowing {tool_name} without permission")
                return {}

            logger.info(f"Session {session_id}: Tool {tool_name} requires permission (tool_use_id: {tool_use_id})")

            # Get the callback first (before we use it)
            callback = self.permission_callbacks.get(session_id)

            # Create permission request
            request_id = str(uuid.uuid4())

            # Store pending permission
            if session_id not in self.pending_permissions:
                self.pending_permissions[session_id] = {}

            self.pending_permissions[session_id][request_id] = {
                'tool': tool_name,
                'parameters': tool_input,
                'tool_use_id': tool_use_id,
                'timestamp': asyncio.get_event_loop().time(),
                'send_callback': callback  # Store callback for acknowledgment
            }

            # Create a future for the permission response
            future = asyncio.Future()
            self.permission_futures[request_id] = future

            # Create human-readable description
            description = self._create_permission_description(tool_name, tool_input)

            # Send permission request via callback
            if callback:
                try:
                    await callback({
                        "type": "permission_request",
                        "session_id": str(session_id),
                        "request_id": request_id,
                        "tool": tool_name,
                        "parameters": tool_input,
                        "description": description
                    })
                    logger.info(f"Session {session_id}: Sent permission request {request_id}, waiting for response...")
                except Exception as e:
                    logger.error(f"Session {session_id}: Failed to send permission request: {e}")
                    # Clean up and deny
                    if request_id in self.permission_futures:
                        del self.permission_futures[request_id]
                    return {
                        "hookSpecificOutput": {
                            "hookEventName": "PreToolUse",
                            "permissionDecision": "deny",
                            "permissionDecisionReason": f"Failed to send permission request: {e}",
                        }
                    }
            else:
                logger.warning(f"Session {session_id}: No callback registered for permission requests, denying by default")
                return {
                    "hookSpecificOutput": {
                        "hookEventName": "PreToolUse",
                        "permissionDecision": "deny",
                        "permissionDecisionReason": "No permission callback registered",
                    }
                }

            # Wait for user response (with timeout)
            try:
                approved = await asyncio.wait_for(future, timeout=30.0)  # 30 second timeout

                # Clean up
                if request_id in self.permission_futures:
                    del self.permission_futures[request_id]

                if approved:
                    logger.info(f"Session {session_id}: Permission {request_id} approved")
                    return {}  # Empty dict = allow execution
                else:
                    logger.info(f"Session {session_id}: Permission {request_id} denied")
                    return {
                        "hookSpecificOutput": {
                            "hookEventName": "PreToolUse",
                            "permissionDecision": "deny",
                            "permissionDecisionReason": "User denied permission",
                        }
                    }
            except asyncio.TimeoutError:
                logger.warning(f"Session {session_id}: Permission request {request_id} timed out")
                # Clean up
                if request_id in self.permission_futures:
                    del self.permission_futures[request_id]
                return {
                    "hookSpecificOutput": {
                        "hookEventName": "PreToolUse",
                        "permissionDecision": "deny",
                        "permissionDecisionReason": "Permission request timed out",
                    }
                }

        return pretool_hook

    def _build_mcp_servers(self, session_id: UUID, mcp_configs: Optional[list]) -> Dict[str, Any]:
        """
        Build MCP servers configuration for ClaudeAgentOptions.

        Args:
            session_id: The session ID
            mcp_configs: List of MCPServerConfig objects

        Returns:
            Dictionary of MCP servers for ClaudeAgentOptions
        """
        mcp_servers = {}

        if not mcp_configs:
            logger.debug(f"Session {session_id}: No MCP servers configured")
            return mcp_servers

        # Store configurations for permission checking later
        if session_id not in self.mcp_server_configs:
            self.mcp_server_configs[session_id] = {}

        for config in mcp_configs:
            try:
                server_config: MCPServerConfig = config if isinstance(config, MCPServerConfig) else MCPServerConfig(**config)
                logger.info(f"Session {session_id}: Configuring MCP server '{server_config.name}' of type '{server_config.type}'")

                # Store configuration for permission checking
                self.mcp_server_configs[session_id][server_config.name] = server_config

                # Build server config for Claude SDK
                if server_config.type == "stdio":
                    if not server_config.command:
                        logger.error(f"Session {session_id}: stdio MCP server '{server_config.name}' missing command")
                        continue

                    server_dict = {
                        "type": "stdio",
                        "command": server_config.command,
                    }

                    if server_config.args:
                        server_dict["args"] = server_config.args

                    if server_config.env:
                        server_dict["env"] = server_config.env

                    mcp_servers[server_config.name] = server_dict
                    logger.info(f"Session {session_id}: Registered stdio MCP server '{server_config.name}' with command: {server_config.command}")

                else:
                    logger.warning(f"Session {session_id}: Unknown MCP server type '{server_config.type}', skipping")

            except Exception as e:
                logger.error(f"Session {session_id}: Error configuring MCP server: {e}")
                continue

        logger.info(f"Session {session_id}: Configured {len(mcp_servers)} MCP servers")
        return mcp_servers

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
            logger.debug(f"Session {session_id}: ======== AGENT CREATION STARTED ========")

            # Convert our Tool enum to string list for Claude SDK
            tools = [tool.value for tool in options.tools]
            logger.debug(f"Session {session_id}: Allowed tools: {tools}")
            logger.debug(f"Session {session_id}: Tool count: {len(tools)}")

            # Log the permission mode for debugging
            logger.info(f"Session {session_id}: Creating agent with permission_mode: {options.permission_mode}")
            logger.debug(f"Session {session_id}: Working directory: {options.working_directory}")
            logger.debug(f"Session {session_id}: Max tokens: {options.max_tokens}")
            logger.debug(f"Session {session_id}: Temperature: {options.temperature}")
            logger.debug(f"Session {session_id}: Has conversation history: {bool(options.conversation_history)}")

            # Build enhanced system prompt with history if provided
            system_prompt = options.system_prompt or "You are a helpful AI assistant."
            if options.conversation_history:
                logger.debug(f"Session {session_id}: Building system prompt with conversation history")
                system_prompt = f"""
{system_prompt}

You are resuming an existing coding session.
Working Directory: {options.working_directory or 'Not specified'}

Recent conversation context:
{options.conversation_history}

Continue helping the user from where they left off. You have access to the project files and can use tools as needed.
"""
            else:
                logger.debug(f"Session {session_id}: No conversation history - fresh session")

            logger.debug(f"Session {session_id}: System prompt length: {len(system_prompt)} chars")

            # Build MCP servers configuration
            logger.debug(f"Session {session_id}: Building MCP server configuration...")
            mcp_servers = self._build_mcp_servers(session_id, options.mcp_servers)
            logger.debug(f"Session {session_id}: MCP servers configured: {list(mcp_servers.keys())}")

            # Create PreToolUse hooks for permission handling (only if in default mode)
            hooks = {}
            if options.permission_mode == 'default':
                logger.debug(f"Session {session_id}: Creating PreToolUse hook for permission handling")
                pretool_hook = self._create_pretool_hook(session_id)
                hooks = {
                    "PreToolUse": [
                        HookMatcher(matcher="*", hooks=[pretool_hook]),  # Match all tools
                    ]
                }
                logger.info(f"Session {session_id}: Registered PreToolUse hook for permission handling")
            else:
                logger.debug(f"Session {session_id}: Permission mode '{options.permission_mode}' - skipping hooks")

            # Create agent options
            logger.debug(f"Session {session_id}: Creating ClaudeAgentOptions...")
            agent_options_dict = {
                "system_prompt": system_prompt,
                "allowed_tools": tools,
                "cwd": options.working_directory,
                "permission_mode": options.permission_mode or 'default',
            }

            # Add hooks if we have them
            if hooks:
                agent_options_dict["hooks"] = hooks
                logger.debug(f"Session {session_id}: Added hooks to agent options")

            # Add MCP servers if configured
            if mcp_servers:
                agent_options_dict["mcp_servers"] = mcp_servers
                logger.info(f"Session {session_id}: Added {len(mcp_servers)} MCP servers to agent options")

            logger.debug(f"Session {session_id}: Agent options dict keys: {list(agent_options_dict.keys())}")
            agent_options = ClaudeAgentOptions(**agent_options_dict)
            logger.debug(f"Session {session_id}: ClaudeAgentOptions created successfully")

            # Create and connect the client using async context manager
            logger.debug(f"Session {session_id}: Creating ClaudeSDKClient...")
            client = ClaudeSDKClient(options=agent_options)
            logger.debug(f"Session {session_id}: ClaudeSDKClient instance created")

            # Connect the client (enter the async context)
            logger.debug(f"Session {session_id}: Connecting client (entering async context)...")
            await client.__aenter__()
            logger.debug(f"Session {session_id}: Client connected successfully")

            # Store both client and options
            self.active_agents[session_id] = client
            self.agent_options[session_id] = agent_options
            logger.debug(f"Session {session_id}: Agent stored in active_agents (total: {len(self.active_agents)})")

            logger.info(f"Session {session_id}: ======== AGENT CREATION COMPLETE ========")
            return client

        except Exception as e:
            logger.error(f"Session {session_id}: ======== AGENT CREATION FAILED ========")
            logger.error(f"Session {session_id}: Error type: {type(e).__name__}")
            logger.error(f"Session {session_id}: Error message: {str(e)}")
            logger.error(f"Session {session_id}: Failed to create agent for session {session_id}: {e}")
            raise RuntimeError(f"Failed to create agent: {e}")

    async def send_prompt(
        self,
        session_id: UUID,
        prompt: str,
        send_message_callback=None
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
        logger.debug(f"Session {session_id}: ======== PROMPT SENDING STARTED ========")
        logger.info(f"Session {session_id}: Sending prompt to session: {prompt[:100]}...")
        logger.debug(f"Session {session_id}: Prompt length: {len(prompt)} chars")
        logger.debug(f"Session {session_id}: Has send_message_callback: {send_message_callback is not None}")

        # Get the client for this session
        client = self.active_agents.get(session_id)
        if not client:
            logger.error(f"Session {session_id}: No agent found in active_agents (total agents: {len(self.active_agents)})")
            raise ValueError(f"No agent found for session {session_id}")

        logger.debug(f"Session {session_id}: Client found, type: {type(client).__name__}")

        # Create permission callback that sends directly if callback provided
        async def permission_callback(permission_request):
            """Callback to send permission requests to the frontend."""
            logger.debug(f"Session {session_id}: permission_callback invoked")
            logger.debug(f"Session {session_id}: Permission request data: {permission_request}")
            if send_message_callback:
                # Send directly to WebSocket
                logger.info(f"Session {session_id}: Sending permission request directly via callback")
                await send_message_callback(permission_request)
            else:
                logger.warning(f"Session {session_id}: No send_message_callback provided for permission request")

        # Register the callback
        self.permission_callbacks[session_id] = permission_callback
        logger.debug(f"Session {session_id}: Permission callback registered (total callbacks: {len(self.permission_callbacks)})")

        try:
            # Send thinking indicator immediately
            logger.debug(f"Session {session_id}: Sending thinking indicator to client")
            yield {
                "type": "agent_thinking",
                "thinking": True
            }

            # Send the prompt to the client
            logger.debug(f"Session {session_id}: Calling client.query() with prompt...")
            await client.query(prompt)
            logger.debug(f"Session {session_id}: client.query() completed, awaiting response stream...")

            # Track message for streaming
            current_message_id = str(uuid.uuid4())  # Generate unique ID for this message
            logger.debug(f"Session {session_id}: Created message stream with ID: {current_message_id}")

            message_buffer = []
            seen_content = set()
            thinking_sent = False
            message_count = 0
            last_message_had_tool_use = False  # Track if last message had tool use
            logger.debug(f"Session {session_id}: Initialized streaming variables")

            # Receive and process responses with timeout
            # Permission requests are now sent directly via callback, not through the generator
            try:
                logger.debug(f"Session {session_id}: Starting to iterate over client.receive_response()...")
                async for message in client.receive_response():
                    message_count += 1
                    logger.debug(f"Session {session_id}: ======== MESSAGE #{message_count} RECEIVED ========")
                    logger.debug(f"Session {session_id}: Message type: {type(message).__name__}")
                    logger.debug(f"Session {session_id}: Message content: {message}")
                    if hasattr(message, '__dict__'):
                        logger.debug(f"Session {session_id}: Message attributes: {message.__dict__}")

                    if isinstance(message, AssistantMessage):
                        logger.debug(f"Session {session_id}: Processing AssistantMessage")
                        logger.debug(f"Session {session_id}: Message has {len(message.content)} content blocks")

                        # Turn off thinking indicator when we get the first message
                        if not thinking_sent:
                            logger.debug(f"Session {session_id}: Turning off thinking indicator")
                            yield {
                                "type": "agent_thinking",
                                "thinking": False
                            }
                            thinking_sent = True

                        # Check if this message contains tool use - if so, we'll start a new message for the next response
                        has_tool_use = any(isinstance(block, ToolUseBlock) for block in message.content)
                        logger.debug(f"Session {session_id}: Message has tool use: {has_tool_use}")

                        # If last message had tool use and this one has text, start a new message ID
                        # This separates responses before and after tool execution
                        if last_message_had_tool_use and any(isinstance(block, TextBlock) for block in message.content):
                            logger.debug(f"Session {session_id}: Tool execution detected, creating new message for response after tool")
                            logger.debug(f"Session {session_id}: Message buffer before flush: {len(message_buffer)} parts")
                            # Flush previous message buffer if any
                            if message_buffer:
                                final_content = "".join(message_buffer)
                                logger.info(f"Session {session_id}: Flushing buffer message (ID: {current_message_id}, {len(final_content)} chars)")
                                yield {
                                    "type": "agent_message",
                                    "content": final_content,
                                    "complete": True,
                                    "message_id": current_message_id
                                }
                                message_buffer = []
                                seen_content = set()
                                logger.debug(f"Session {session_id}: Buffer cleared")
                            # Generate new message ID for post-tool-execution content
                            current_message_id = str(uuid.uuid4())
                            logger.debug(f"Session {session_id}: New message ID after tool execution: {current_message_id}")

                        last_message_had_tool_use = has_tool_use

                        # Extract text from assistant message and handle tool use blocks
                        content_parts = []
                        tool_use_count = 0
                        text_block_count = 0

                        for block_idx, block in enumerate(message.content):
                            if isinstance(block, TextBlock):
                                content_parts.append(block.text)
                                text_block_count += 1
                                logger.debug(f"Session {session_id}: Text block #{block_idx}: {len(block.text)} chars")
                            elif isinstance(block, ToolUseBlock):
                                tool_use_count += 1
                                logger.debug(f"Session {session_id}: Tool block #{block_idx}: tool='{block.name}', tool_use_id='{block.id}'")
                                logger.debug(f"Session {session_id}: Tool parameters: {block.input}")

                                # Yield tool use event immediately
                                tool_data = {
                                    "type": "agent_tool_use",
                                    "tool": block.name,
                                    "parameters": block.input,
                                    "tool_use_id": block.id
                                }

                                # Special handling for TodoWrite - capture the full input data
                                if block.name == "TodoWrite" and block.input:
                                    tool_data["input"] = block.input
                                    logger.debug(f"Session {session_id}: Enhanced TodoWrite tool with input data: {block.input}")

                                logger.debug(f"Session {session_id}: Yielding tool use event for {block.name}")
                                yield tool_data
                            else:
                                logger.debug(f"Session {session_id}: Unknown block type: {type(block).__name__}")

                        logger.debug(f"Session {session_id}: Block summary - Text: {text_block_count}, Tools: {tool_use_count}")

                        if content_parts:
                            content = "".join(content_parts)
                            logger.debug(f"Session {session_id}: Assembled content from {len(content_parts)} parts: {len(content)} chars")

                            # Avoid sending duplicate content
                            if content not in seen_content:
                                seen_content.add(content)
                                message_buffer.append(content)
                                logger.debug(f"Session {session_id}: Added to buffer (total buffer now: {len(message_buffer)} parts)")
                                logger.debug(f"Session {session_id}: Yielding content chunk (ID: {current_message_id}): {content[:80]}...")
                                yield {
                                    "type": "agent_message",
                                    "content": content,
                                    "complete": False,
                                    "message_id": current_message_id
                                }
                            else:
                                logger.debug(f"Session {session_id}: Skipping duplicate content (already in seen_content)")
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

                logger.debug(f"Session {session_id}: ======== RESPONSE STREAM COMPLETED ========")
                logger.info(f"Session {session_id}: Received total of {message_count} messages from agent")
                logger.debug(f"Session {session_id}: Current message ID: {current_message_id}")
                logger.debug(f"Session {session_id}: Current buffer size: {len(message_buffer)} parts")
                logger.debug(f"Session {session_id}: Thinking sent: {thinking_sent}")

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
                    logger.info(f"Session {session_id}: Sending final completion message (ID: {current_message_id}, {len(final_content)} chars)")
                    logger.debug(f"Session {session_id}: Final message buffer had {len(message_buffer)} parts, {len(seen_content)} unique contents")
                    yield {
                        "type": "agent_message",
                        "content": final_content,
                        "complete": True,
                        "message_id": current_message_id
                    }
                    logger.debug(f"Session {session_id}: Final message yielded")
                else:
                    logger.warning(f"Session {session_id}: No content to send in completion (received {message_count} messages, no text content assembled)")

            except Exception as e:
                logger.error(f"Session {session_id}: ======== RESPONSE STREAM ERROR ========")
                logger.error(f"Session {session_id}: Error type: {type(e).__name__}")
                logger.error(f"Session {session_id}: Error message: {str(e)}")
                logger.error(f"Session {session_id}: Received {message_count} messages before error")
                logger.error(f"Session {session_id}: Current buffer: {len(message_buffer)} parts")
                import traceback
                logger.error(f"Session {session_id}: Traceback: {traceback.format_exc()}")
                yield {
                    "type": "agent_error",
                    "message": f"Error processing response: {str(e)}",
                    "session_id": str(session_id)
                }

        except Exception as e:
            logger.error(f"Session {session_id}: ======== PROMPT SENDING FAILED ========")
            logger.error(f"Session {session_id}: Error type: {type(e).__name__}")
            logger.error(f"Session {session_id}: Error message: {str(e)}")
            import traceback
            logger.error(f"Session {session_id}: Traceback: {traceback.format_exc()}")
            raise RuntimeError(f"Failed to send prompt: {e}")
        finally:
            # Clean up permission callback
            logger.debug(f"Session {session_id}: ======== PROMPT SENDING CLEANUP ========")
            logger.debug(f"Session {session_id}: Total messages processed: {message_count}")
            logger.debug(f"Session {session_id}: Final buffer size: {len(message_buffer)} parts")
            logger.debug(f"Session {session_id}: Total permission callbacks registered: {len(self.permission_callbacks)}")
            logger.debug(f"Session {session_id}: Total pending permissions: {len(self.pending_permissions.get(session_id, {}))}")

            if session_id in self.permission_callbacks:
                del self.permission_callbacks[session_id]
                logger.debug(f"Session {session_id}: Cleaned up permission callback (remaining callbacks: {len(self.permission_callbacks)})")
            logger.debug(f"Session {session_id}: ======== PROMPT SENDING COMPLETE ========")

    async def handle_permission_response(
        self,
        session_id: UUID,
        request_id: str,
        approved: bool,
        reason: Optional[str] = None
    ) -> bool:
        """
        Handle user response to a permission request.

        Args:
            session_id: The session ID
            request_id: The permission request ID
            approved: Whether the user approved the request
            reason: Optional reason for denial

        Returns:
            True if handled successfully, False if request not found
        """
        logger.debug(f"Session {session_id}: ======== PERMISSION RESPONSE RECEIVED ========")
        logger.info(f"Session {session_id}: Permission response for {request_id}: approved={approved}, reason={reason}")
        logger.debug(f"Session {session_id}: Total pending permissions: {sum(len(p) for p in self.pending_permissions.values())}")

        # Look up the future for this permission request
        future = self.permission_futures.get(request_id)
        logger.debug(f"Session {session_id}: Future found: {future is not None}")
        logger.debug(f"Session {session_id}: Future done: {future.done() if future else 'N/A'}")

        if not future:
            logger.warning(f"Session {session_id}: Permission future {request_id} not found (request may have already been processed or timed out)")
            logger.debug(f"Session {session_id}: Known futures: {list(self.permission_futures.keys())}")
            return False

        # Check if we have the pending permission data
        session_permissions = self.pending_permissions.get(session_id)
        logger.debug(f"Session {session_id}: Session permissions exist: {session_permissions is not None}")
        logger.debug(f"Session {session_id}: Request in permissions: {request_id in session_permissions if session_permissions else False}")

        if session_permissions and request_id in session_permissions:
            permission_data = session_permissions[request_id]
            tool_name = permission_data.get('tool', 'unknown')
            tool_params = permission_data.get('parameters', {})
            send_callback = permission_data.get('send_callback')
            tool_use_id = permission_data.get('tool_use_id', 'unknown')

            logger.debug(f"Session {session_id}: Permission request details - tool: {tool_name}, tool_use_id: {tool_use_id}")
            logger.debug(f"Session {session_id}: Tool parameters: {tool_params}")

            if approved:
                logger.info(f"Session {session_id}: Permission {request_id} APPROVED for tool '{tool_name}'")
            else:
                logger.info(f"Session {session_id}: Permission {request_id} DENIED for tool '{tool_name}'. Reason: {reason}")

            # Send immediate acknowledgment to frontend
            if send_callback:
                try:
                    ack_message = {
                        "type": "permission_acknowledged",
                        "session_id": str(session_id),
                        "request_id": request_id,
                        "approved": approved,
                        "tool": tool_name,
                        "status": "executing" if approved else "denied"
                    }
                    logger.debug(f"Session {session_id}: Sending permission acknowledgment: {ack_message}")
                    await send_callback(ack_message)
                    logger.info(f"Session {session_id}: Sent permission acknowledgment for {request_id}")
                except Exception as e:
                    logger.error(f"Session {session_id}: Failed to send permission acknowledgment: {e}")
            else:
                logger.warning(f"Session {session_id}: No send_callback registered for permission acknowledgment")

            # Remove from pending permissions
            del session_permissions[request_id]
            logger.debug(f"Session {session_id}: Removed from pending permissions (remaining: {len(session_permissions)})")

            if not session_permissions:
                del self.pending_permissions[session_id]
                logger.debug(f"Session {session_id}: Cleared all pending permissions for this session")
        else:
            logger.warning(f"Session {session_id}: Permission data {request_id} not found in pending_permissions")
            logger.debug(f"Session {session_id}: Known pending permissions: {list(session_permissions.keys()) if session_permissions else 'none'}")

        # Resolve the future with the approval decision
        if not future.done():
            logger.debug(f"Session {session_id}: Setting future result to {approved}")
            future.set_result(approved)
            logger.info(f"Session {session_id}: Resolved permission future {request_id} with result={approved}")
        else:
            logger.warning(f"Session {session_id}: Future {request_id} was already done/resolved (current result: {future.result()})")

        logger.debug(f"Session {session_id}: ======== PERMISSION RESPONSE COMPLETE ========")
        return True

    def _requires_permission(self, session_id: UUID, tool_name: str, parameters: Dict[str, Any]) -> bool:
        """
        Check if a tool use requires permission based on session settings.

        Args:
            session_id: The session ID
            tool_name: The name of the tool being used
            parameters: The tool parameters

        Returns:
            True if permission is required, False otherwise
        """
        # Get the session options
        agent_options = self.agent_options.get(session_id)
        if not agent_options:
            logger.warning(f"Session {session_id}: No agent options found, denying permission")
            return False

        permission_mode = getattr(agent_options, 'permission_mode', None)
        logger.debug(f"Session {session_id}: permission_mode = {permission_mode}")

        # If permission mode is not 'default', no permission requests needed
        if permission_mode != 'default':
            logger.debug(f"Session {session_id}: Permission mode is '{permission_mode}', not requiring permission")
            return False

        # Define tools that require permission
        permission_required_tools = {'Write', 'Edit', 'Bash'}

        # Check if this is an MCP tool
        if tool_name.startswith('mcp__'):
            # Extract server name from mcp__<server_name>__<tool_name>
            parts = tool_name.split('__')
            if len(parts) >= 3:
                server_name = parts[1]
                mcp_configs = self.mcp_server_configs.get(session_id, {})
                if server_name in mcp_configs:
                    mcp_config = mcp_configs[server_name]
                    # Use require_permission flag from MCP server config
                    requires = mcp_config.require_permission
                    logger.debug(f"Session {session_id}: MCP tool '{tool_name}' from server '{server_name}' require_permission={requires}")
                    return requires
                else:
                    # Default to requiring permission for unknown MCP servers
                    logger.debug(f"Session {session_id}: MCP tool '{tool_name}' server config not found, requiring permission by default")
                    return True
            else:
                # Malformed MCP tool name, require permission
                logger.warning(f"Session {session_id}: Malformed MCP tool name: {tool_name}")
                return True

        requires = tool_name in permission_required_tools
        logger.debug(f"Session {session_id}: Tool '{tool_name}' in permission_required_tools = {requires}")

        return requires

    def _create_permission_request(self, session_id: UUID, tool_name: str, parameters: Dict[str, Any]) -> Dict:
        """
        Create a permission request for a tool use.

        Args:
            session_id: The session ID
            tool_name: The name of the tool being used
            parameters: The tool parameters

        Returns:
            Permission request message dict
        """
        request_id = str(uuid.uuid4())

        # Store the pending permission
        if session_id not in self.pending_permissions:
            self.pending_permissions[session_id] = {}

        self.pending_permissions[session_id][request_id] = {
            'tool': tool_name,
            'parameters': parameters,
            'timestamp': asyncio.get_event_loop().time()
        }

        # Create a human-readable description
        description = self._create_permission_description(tool_name, parameters)

        logger.info(f"Session {session_id}: Permission request created for {tool_name}")

        return {
            "type": "permission_request",
            "session_id": str(session_id),
            "request_id": request_id,
            "tool": tool_name,
            "parameters": parameters,
            "description": description
        }

    def _create_permission_description(self, tool_name: str, parameters: Dict[str, Any]) -> str:
        """Create a human-readable description of the permission request."""
        if tool_name == "Write":
            file_path = parameters.get("file_path", "unknown file")
            return f"Agent wants to create or modify the file: {file_path}"
        elif tool_name == "Edit":
            file_path = parameters.get("file_path", "unknown file")
            return f"Agent wants to edit the file: {file_path}"
        elif tool_name == "Bash":
            command = parameters.get("command", "unknown command")
            return f"Agent wants to run the command: {command}"
        elif tool_name.startswith('mcp__'):
            # Handle MCP tool names: mcp__<server_name>__<tool_name>
            parts = tool_name.split('__')
            if len(parts) >= 3:
                server_name = parts[1]
                actual_tool_name = '__'.join(parts[2:])  # In case tool name has __ in it
                return f"Agent wants to use the '{actual_tool_name}' tool from the MCP server '{server_name}' with parameters: {parameters}"
            else:
                return f"Agent wants to use the MCP tool '{tool_name}' with parameters: {parameters}"
        else:
            return f"Agent wants to use the {tool_name} tool with parameters: {parameters}"

    # NOTE: Permission detection is now handled by PreToolUse hooks, not text-based detection.
    # The _detect_permission_request method has been removed because it was causing false positives
    # by matching conversational phrases like "would you like me to" which are just the agent being polite.

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
            tool_name = message.get("tool", "")
            parameters = message.get("parameters", {})

            # For TodoWrite, we need to capture the full tool block data
            tool_data = {
                "type": "agent_tool_use",
                "tool": tool_name,
                "parameters": parameters,
                "result": message.get("result")
            }

            # Check if this tool requires permission and the session is in default mode
            session_id = message.get("session_id")
            if session_id and self._requires_permission(session_id, tool_name, parameters):
                return self._create_permission_request(session_id, tool_name, parameters)

            # Special handling for TodoWrite - capture the full input data
            if tool_name == "TodoWrite" and message.get("input"):
                tool_data["input"] = message["input"]
                logger.debug(f"Session {session_id}: Enhanced TodoWrite tool with input data")

            return tool_data
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
            if session_id in self.mcp_server_configs:
                del self.mcp_server_configs[session_id]

            logger.info(f"Ended agent for session {session_id}")
            return True

        except Exception as e:
            logger.error(f"Error ending agent for session {session_id}: {e}")
            # Still remove from active agents
            if session_id in self.active_agents:
                del self.active_agents[session_id]
            if session_id in self.agent_options:
                del self.agent_options[session_id]
            if session_id in self.mcp_server_configs:
                del self.mcp_server_configs[session_id]
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