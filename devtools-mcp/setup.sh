#!/bin/bash
set -e

echo "========================================"
echo "Devtools MCP Setup"
echo "========================================"
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
  echo "ERROR: Docker is not installed or not in PATH."
  echo "Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
  exit 1
fi

echo "[1/3] Docker found: ✓"
echo ""

# Create bin directory if it doesn't exist
mkdir -p bin
echo "[2/3] Created bin directory (if needed): ✓"
echo ""

# Build the executable
echo "[3/3] Building devtools-mcp (Linux)..."
docker run --rm -v "$(pwd)":/src -w /src golang:1.22 sh -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/devtools-mcp ."

if [ $? -ne 0 ]; then
  echo ""
  echo "ERROR: Build failed. Check Docker is running and you have internet access."
  exit 1
fi

chmod +x bin/devtools-mcp

echo ""
echo "========================================"
echo "✓ Setup Complete!"
echo "========================================"
echo ""
echo "Binary built to: bin/devtools-mcp"
echo ""
echo "Next steps:"
echo "  1. Update .vscode/mcp.json to include the devtools-mcp server"
echo "  2. Restart VS Code"
echo "  3. The MCP will auto-connect"
echo "  4. Try calling it via an AI agent"
echo ""
