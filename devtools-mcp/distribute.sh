#!/bin/bash
# Distribution build for devtools-mcp
# Ensures LATEST version is built before distribution

set -e

echo "========================================"
echo "Devtools MCP - Distribution Build"
echo "========================================"
echo ""
echo "This script ensures the LATEST version is"
echo "built before distribution."
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
  echo "ERROR: Docker is not installed."
  exit 1
fi

echo "[1/2] Docker found: ✓"
echo ""

# Create bin directory
mkdir -p bin

# Build Linux
echo "[2/2] Building LATEST devtools-mcp from source..."
docker run --rm -v "$(pwd)":/src -w /src golang:1.22 sh -c \
  "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/devtools-mcp ."
chmod +x bin/devtools-mcp

echo ""
echo "========================================"
echo "✓ Distribution Package Ready!"
echo "========================================"
echo ""
echo "Binary: bin/devtools-mcp"
echo "Built:  $(date)"
echo ""
echo "Ready to distribute:"
echo "   - install.sh (for Unix users)"
echo "   - devtools-mcp (for direct distribution)"
echo ""
echo "Next steps:"
echo "   1. Run: ./build-linux.sh (to create install.sh)"
echo "   2. Distribute install.sh to end users"
echo ""
