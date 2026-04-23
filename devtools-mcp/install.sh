#!/bin/bash
# Devtools MCP - Installation Script (Linux/Mac)

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Display header
clear
echo ""
echo "========================================"
echo "Devtools MCP - Installation"
echo "========================================"
echo ""
echo "This installer will set up Devtools MCP for use in your projects."
echo ""

# Determine installation directory
if [ -n "$1" ]; then
    INSTALL_PATH="$1"
else
    INSTALL_PATH="${HOME}/.local/devtools-mcp"
    echo "Default installation path: $INSTALL_PATH"
    read -p "Enter installation path (or press Enter for default): " USER_PATH
    if [ -n "$USER_PATH" ]; then
        INSTALL_PATH="$USER_PATH"
    fi
fi

INSTALL_PATH="${INSTALL_PATH%/}"  # Remove trailing slash

echo ""
echo "Creating directory: $INSTALL_PATH"
mkdir -p "$INSTALL_PATH"

# Check for binary
if [ ! -f "$SCRIPT_DIR/bin/devtools-mcp" ]; then
    echo "ERROR: Binary not found at $SCRIPT_DIR/bin/devtools-mcp"
    echo "Please run: ./build-linux.sh"
    exit 1
fi

echo "Copying binary..."
cp "$SCRIPT_DIR/bin/devtools-mcp" "$INSTALL_PATH/devtools-mcp"
chmod +x "$INSTALL_PATH/devtools-mcp"

# Create README
cat > "$INSTALL_PATH/README.txt" << EOF
# Devtools MCP Installed Successfully

Installation Location: $INSTALL_PATH

## Next Steps - Configure in Your Project:

1. In your project directory, create or edit \`.vscode/mcp.json\`:

\`\`\`json
{
  "mcpServers": {
    "devtools-mcp": {
      "type": "stdio",
      "command": "$INSTALL_PATH/devtools-mcp"
    }
  }
}
\`\`\`

2. Create credential files in your project root:
   - \`.pat\` - GitHub Personal Access Token
   - \`.username\` - GitHub username

3. Restart VS Code

## System-wide Installation (Optional)

To make devtools-mcp available system-wide:

\`\`\`bash
sudo cp $INSTALL_PATH/devtools-mcp /usr/local/bin/devtools-mcp
sudo chmod +x /usr/local/bin/devtools-mcp
\`\`\`

Then use in mcp.json:
\`\`\`json
"command": "devtools-mcp"
\`\`\`
EOF

echo "Creating deployment guide..."
cat > "$INSTALL_PATH/DEPLOYMENT.txt" << 'EOF'
# Devtools MCP - Linux Deployment Guide

## Installation Options

### Option 1: User-local Installation (Recommended)
Run the install script:
```bash
./install.sh
```

This installs to ~/.local/devtools-mcp

### Option 2: System-wide Installation
```bash
sudo cp bin/devtools-mcp /usr/local/bin/
sudo chmod +x /usr/local/bin/devtools-mcp
```

### Option 3: Direct Binary Use
Copy the binary anywhere and reference the full path in mcp.json.

## Configuration in Projects

Edit your project's `.vscode/mcp.json`:

```json
{
  "mcpServers": {
    "devtools-mcp": {
      "type": "stdio",
      "command": "~/.local/devtools-mcp/devtools-mcp"
    }
  }
}
```

Or if installed system-wide:
```json
{
  "mcpServers": {
    "devtools-mcp": {
      "type": "stdio",
      "command": "devtools-mcp"
    }
  }
}
```

## Credentials

Create `.pat` and `.username` files in your project root:

```bash
echo "your_github_token" > .pat
echo "your_username" > .username
chmod 600 .pat .username
```

## Troubleshooting

**Binary not executable:**
```bash
chmod +x ~/.local/devtools-mcp/devtools-mcp
```

**Not found in PATH:**
```bash
export PATH="$PATH:$HOME/.local/devtools-mcp"
# Add to ~/.bashrc or ~/.zshrc to make it persistent
```

**Permission denied:**
Make sure the binary is executable: `chmod +x /path/to/devtools-mcp`
EOF

echo ""
echo "========================================"
echo "✓ Installation Complete!"
echo "========================================"
echo ""
echo "Binary installed to: $INSTALL_PATH/devtools-mcp"
echo ""
echo "Command to use in mcp.json:"
echo "  \"command\": \"$INSTALL_PATH/devtools-mcp\""
echo ""
echo "For more details, see: $INSTALL_PATH/README.txt"
echo ""

