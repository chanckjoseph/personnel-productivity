#!/bin/bash
# Cross-platform release builder for devtools-mcp
# Builds binaries for Windows, Linux, and macOS

set -e

echo "========================================"
echo "Devtools MCP - Cross-Platform Build"
echo "========================================"
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
  echo "ERROR: Docker is not installed."
  exit 1
fi

echo "[1/5] Docker found: ✓"

# Create bin directory
mkdir -p bin
echo "[2/5] Created bin directory: ✓"
echo ""

# Build for Windows
echo "[3/5] Building for Windows (64-bit)..."
docker run --rm -v "$(pwd)":/src -w /src golang:1.22 sh -c \
  "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/devtools-mcp.exe ."
echo "  ✓ bin/devtools-mcp.exe"

# Build for Linux
echo "[4/5] Building for Linux (64-bit)..."
docker run --rm -v "$(pwd)":/src -w /src golang:1.22 sh -c \
  "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/devtools-mcp ."
chmod +x bin/devtools-mcp
echo "  ✓ bin/devtools-mcp"

# Build for macOS (Intel & ARM)
echo "[5/5] Building for macOS..."
docker run --rm -v "$(pwd)":/src -w /src golang:1.22 sh -c \
  "CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/devtools-mcp-darwin-amd64 . && \
   CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/devtools-mcp-darwin-arm64 ."
chmod +x bin/devtools-mcp-darwin-*
echo "  ✓ bin/devtools-mcp-darwin-amd64 (Intel)"
echo "  ✓ bin/devtools-mcp-darwin-arm64 (Apple Silicon)"

echo ""
echo "========================================"
echo "✓ All Binaries Built!"
echo "========================================"
echo ""
echo "Binaries ready:"
echo "  - bin/devtools-mcp.exe         (Windows 64-bit)"
echo "  - bin/devtools-mcp             (Linux 64-bit)"
echo "  - bin/devtools-mcp-darwin-amd64 (macOS Intel)"
echo "  - bin/devtools-mcp-darwin-arm64 (macOS Apple Silicon)"
echo ""
echo "Next steps:"
echo "  1. Create release: ./build-release.sh"
echo "  2. Windows installer: build-installer.bat"
echo "  3. Linux installer: ./install.sh"
echo ""

