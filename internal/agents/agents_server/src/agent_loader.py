"""Agent loader for parsing and caching agents from .claude/agents/ directory."""

import logging
import os
import re
from pathlib import Path
from typing import Dict, Optional
import yaml

logger = logging.getLogger(__name__)


class AgentMetadata:
    """Metadata about a Claude agent."""

    def __init__(self, name: str, description: str, system_prompt: str,
                 tools: Optional[str] = None, model: Optional[str] = None,
                 color: Optional[str] = None):
        self.name = name
        self.description = description
        self.system_prompt = system_prompt
        self.tools = tools
        self.model = model
        self.color = color

    def to_dict(self, include_system_prompt: bool = False):
        """Convert to dictionary for JSON serialization.

        Args:
            include_system_prompt: Whether to include the full system_prompt content
        """
        result = {
            "name": self.name,
            "description": self.description,
            "tools": self.tools,
            "model": self.model,
            "color": self.color,
        }
        if include_system_prompt:
            result["system_prompt"] = self.system_prompt
            result["system_prompt_length"] = len(self.system_prompt)
        return result


class AgentLoader:
    """Loads and caches Claude agents from markdown files."""

    def __init__(self, agents_dir: Optional[str] = None):
        """Initialize the agent loader.

        Args:
            agents_dir: Path to the directory containing agent markdown files.
                       If None, defaults to ~/.claude/agents/
                       Can also be a project directory, in which case it will look for .claude/agents
        """
        if agents_dir is None:
            agents_dir = os.path.expanduser("~/.claude/agents")
        else:
            # If agents_dir is a project directory, check for .claude/agents subdirectory
            path = Path(agents_dir)
            agents_subdir = path / ".claude" / "agents"
            if agents_subdir.exists():
                agents_dir = str(agents_subdir)
                logger.debug(f"Found .claude/agents in project directory: {agents_dir}")

        self.agents_dir = Path(agents_dir)
        self.agents: Dict[str, AgentMetadata] = {}
        self._loaded = False

        logger.info(f"AgentLoader initialized with directory: {self.agents_dir}")

    def load_agents(self) -> Dict[str, AgentMetadata]:
        """Load all agents from the agents directory.

        Returns:
            Dictionary of agent name -> AgentMetadata
        """
        if self._loaded:
            logger.debug("Agents already loaded, returning cached agents")
            return self.agents

        if not self.agents_dir.exists():
            logger.warning(f"Agents directory not found: {self.agents_dir}")
            self._loaded = True
            return self.agents

        logger.info(f"Loading agents from {self.agents_dir}")

        agent_files = sorted(self.agents_dir.glob("*.md"))
        logger.debug(f"Found {len(agent_files)} agent files")

        for agent_file in agent_files:
            try:
                agent = self._parse_agent_file(agent_file)
                if agent:
                    self.agents[agent.name] = agent
                    logger.info(f"Loaded agent: {agent.name}")
            except Exception as e:
                logger.error(f"Error loading agent from {agent_file.name}: {e}")

        logger.info(f"Successfully loaded {len(self.agents)} agents")
        self._loaded = True
        return self.agents

    def get_agent(self, name: str) -> Optional[AgentMetadata]:
        """Get an agent by name.

        Args:
            name: The name of the agent

        Returns:
            AgentMetadata if found, None otherwise
        """
        if not self._loaded:
            self.load_agents()

        agent = self.agents.get(name)
        if agent:
            logger.debug(f"Retrieved agent: {name}")
        else:
            logger.warning(f"Agent not found: {name}")

        return agent

    def list_agents(self) -> Dict[str, Dict]:
        """List all available agents with their metadata.

        Returns:
            Dictionary of agent name -> metadata dict
        """
        if not self._loaded:
            self.load_agents()

        return {name: agent.to_dict() for name, agent in self.agents.items()}

    def _parse_agent_file(self, file_path: Path) -> Optional[AgentMetadata]:
        """Parse a single agent markdown file.

        Args:
            file_path: Path to the markdown file

        Returns:
            AgentMetadata if successfully parsed, None otherwise
        """
        try:
            content = file_path.read_text(encoding='utf-8')

            # Extract YAML frontmatter
            frontmatter_match = re.match(r'^---\n(.*?)\n---\n(.*)', content, re.DOTALL)

            if not frontmatter_match:
                logger.warning(f"No frontmatter found in {file_path.name}")
                return None

            yaml_content = frontmatter_match.group(1)
            system_prompt = frontmatter_match.group(2).strip()

            # Parse YAML
            metadata = yaml.safe_load(yaml_content) or {}

            # Extract required fields
            name = metadata.get('name')
            description = metadata.get('description', '')

            if not name:
                logger.warning(f"Agent file {file_path.name} missing 'name' field")
                return None

            if not system_prompt:
                logger.warning(f"Agent {name} has no system prompt content")
                return None

            # Extract optional fields
            tools = metadata.get('tools')
            model = metadata.get('model')
            color = metadata.get('color')

            agent = AgentMetadata(
                name=name,
                description=description,
                system_prompt=system_prompt,
                tools=tools,
                model=model,
                color=color
            )

            logger.debug(f"Parsed agent {name}: {len(system_prompt)} chars")
            return agent

        except Exception as e:
            logger.error(f"Exception parsing {file_path.name}: {e}")
            return None


# Cache of agent loaders per working directory
_agent_loaders: Dict[str, AgentLoader] = {}
_default_agent_loader: Optional[AgentLoader] = None


def get_agent_loader(working_directory: Optional[str] = None) -> AgentLoader:
    """Get an agent loader instance.

    Args:
        working_directory: Optional working directory for the session.
                         If provided, agents will be loaded from that project's .claude/agents directory.
                         If None, defaults to ~/.claude/agents/

    Returns:
        AgentLoader instance (cached per working directory)
    """
    global _agent_loaders, _default_agent_loader

    if working_directory is None:
        # Use default global loader
        if _default_agent_loader is None:
            _default_agent_loader = AgentLoader()
            _default_agent_loader.load_agents()
            logger.debug(f"Created default agent loader with {len(_default_agent_loader.agents)} agents")
        return _default_agent_loader

    # Use cached loader for this working directory
    if working_directory not in _agent_loaders:
        logger.debug(f"Creating new agent loader for working directory: {working_directory}")
        loader = AgentLoader(working_directory)
        loader.load_agents()
        _agent_loaders[working_directory] = loader
        logger.info(f"Created agent loader for {working_directory} with {len(loader.agents)} agents")

    return _agent_loaders[working_directory]


def initialize_agent_loader(agents_dir: Optional[str] = None):
    """Initialize the default agent loader.

    Args:
        agents_dir: Optional path to agents directory
    """
    global _default_agent_loader
    _default_agent_loader = AgentLoader(agents_dir)
    _default_agent_loader.load_agents()
    logger.info(f"Initialized default agent loader with {len(_default_agent_loader.agents)} agents")


def clear_agent_loader_cache():
    """Clear all cached agent loaders (useful for testing)."""
    global _agent_loaders
    _agent_loaders.clear()
    logger.debug("Cleared agent loader cache")
