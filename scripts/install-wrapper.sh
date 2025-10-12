#!/bin/bash
# Install CCT wrapper for Claude Code
# This script creates a wrapper that intercepts claude commands

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing Claude Control Terminal Wrapper${NC}"
echo ""

# Find cct binary
CCT_BIN=$(which cct 2>/dev/null || echo "")
if [ -z "$CCT_BIN" ]; then
    echo -e "${RED}Error: cct binary not found in PATH${NC}"
    echo "Please install cct first or add it to your PATH"
    exit 1
fi

# Find claude binary
CLAUDE_BIN=$(which claude 2>/dev/null || echo "")
if [ -z "$CLAUDE_BIN" ]; then
    echo -e "${RED}Error: claude binary not found in PATH${NC}"
    echo "Please install Claude Code first"
    exit 1
fi

# Resolve symlinks
CLAUDE_REAL=$(readlink -f "$CLAUDE_BIN" 2>/dev/null || realpath "$CLAUDE_BIN" 2>/dev/null || echo "$CLAUDE_BIN")

echo -e "Found claude at: ${YELLOW}$CLAUDE_REAL${NC}"
echo -e "Found cct at: ${YELLOW}$CCT_BIN${NC}"
echo ""

# Create wrapper directory
WRAPPER_DIR="$HOME/.cct/bin"
mkdir -p "$WRAPPER_DIR"

# Backup original claude
CLAUDE_BACKUP="$HOME/.cct/claude.original"
if [ ! -f "$CLAUDE_BACKUP" ]; then
    cp "$CLAUDE_REAL" "$CLAUDE_BACKUP"
    echo -e "${GREEN}✓${NC} Backed up original claude to: $CLAUDE_BACKUP"
fi

# Create wrapper script
WRAPPER_SCRIPT="$WRAPPER_DIR/claude"
cat > "$WRAPPER_SCRIPT" << 'WRAPPER_EOF'
#!/bin/bash
# CCT Wrapper for Claude Code

# Path to original claude
CLAUDE_ORIGINAL="$HOME/.cct/claude.original"

# Execute original claude with all arguments
exec "$CLAUDE_ORIGINAL" "$@"
WRAPPER_EOF

chmod +x "$WRAPPER_SCRIPT"
echo -e "${GREEN}✓${NC} Created wrapper script at: $WRAPPER_SCRIPT"
echo ""

# Update PATH instructions
echo -e "${YELLOW}Installation complete!${NC}"
echo ""
echo -e "To use the wrapper, add this line to your ${GREEN}~/.bashrc${NC} or ${GREEN}~/.zshrc${NC}:"
echo ""
echo -e "  ${GREEN}export PATH=\"$WRAPPER_DIR:\$PATH\"${NC}"
echo ""
echo -e "Then reload your shell:"
echo -e "  ${GREEN}source ~/.bashrc${NC}  # or source ~/.zshrc"
echo ""
echo -e "To verify:"
echo -e "  ${GREEN}which claude${NC} should show: $WRAPPER_SCRIPT"
echo ""

# Offer to update shell config automatically
read -p "Would you like to update your shell config automatically? [y/N] " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Detect shell
    if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
        SHELL_CONFIG="$HOME/.zshrc"
    else
        SHELL_CONFIG="$HOME/.bashrc"
    fi

    # Check if already added
    if grep -q "$WRAPPER_DIR" "$SHELL_CONFIG" 2>/dev/null; then
        echo -e "${YELLOW}PATH already updated in $SHELL_CONFIG${NC}"
    else
        echo "" >> "$SHELL_CONFIG"
        echo "# CCT Claude wrapper" >> "$SHELL_CONFIG"
        echo "export PATH=\"$WRAPPER_DIR:\$PATH\"" >> "$SHELL_CONFIG"
        echo -e "${GREEN}✓${NC} Updated $SHELL_CONFIG"
        echo -e "Please run: ${GREEN}source $SHELL_CONFIG${NC}"
    fi
fi

echo ""
echo -e "${GREEN}Installation complete!${NC}"
