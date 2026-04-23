#!/bin/bash
# Devtools MCP Setup + Installer Build (Linux/Mac)

set -e

echo "========================================"
echo "Devtools MCP Setup + Installer Build"
echo "========================================"
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
  echo "ERROR: Docker is not installed or not in PATH."
  echo "Please install Docker from https://www.docker.com/products/docker-desktop"
  exit 1
fi

echo "[1/4] Docker found: ✓"
echo ""

# Create bin directory if it doesn't exist
mkdir -p bin
echo "[2/4] Created bin directory (if needed): ✓"
echo ""

# Build the Linux binary
echo "[3/4] Building devtools-mcp (Linux)..."
docker run --rm -v "$(pwd)":/src -w /src golang:1.22 sh -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/devtools-mcp ."

if [ $? -ne 0 ]; then
  echo ""
  echo "ERROR: Build failed. Check Docker is running and you have internet access."
  exit 1
fi

chmod +x bin/devtools-mcp
echo "devtools-mcp (Linux) built successfully: ✓"
echo ""

# Check for fpm (for .deb package)
echo "[4/4] Checking for optional tools..."
if command -v fpm &> /dev/null; then
  echo "FPM found - can build .deb package"
  read -p "Build .deb package? (y/n) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    mkdir -p debian_build/devtools-mcp/usr/local/bin
    cp bin/devtools-mcp debian_build/devtools-mcp/usr/local/bin/
    chmod +x debian_build/devtools-mcp/usr/local/bin/devtools-mcp
    
    fpm -s dir -t deb -n devtools-mcp -v 1.0.0 \
      -C debian_build/devtools-mcp \
      -p devtools-mcp_VERSION_amd64.deb \
      --description "MCP server for automating Git workflows" \
      --maintainer "DevTools" \
      --after-install debian_post.sh 2>/dev/null || echo "Note: .deb build skipped (fpm issues)"
    
    rm -rf debian_build
  fi
else
  echo "FPM not found - skipping .deb package build"
  echo "To create .deb packages, install: sudo apt-get install ruby-dev && sudo gem install fpm"
fi

echo ""
echo "========================================"
echo "✓ Build Complete!"
echo "========================================"
echo ""
echo "Binaries built to:"
echo "  - bin/devtools-mcp (Linux 64-bit)"
echo ""
echo "Installation options:"
echo "  1. Run: ./install.sh"
echo "  2. Copy: cp bin/devtools-mcp /usr/local/bin/"
echo "  3. Or use in project: configure .vscode/mcp.json"
echo ""

