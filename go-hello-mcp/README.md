# hello-go-mcp

A minimal MCP (Model Context Protocol) server written in Go. Demonstrates how to build a portable, simple AI tool server.

**What it does:** Provides a `hello_greeting` tool that returns personalized greeting messages.

## Quick Start (Windows)

1. **Make sure Docker is running** (or install it from https://www.docker.com/products/docker-desktop)

2. From this folder (`go-hello-mcp/`), run:
   ```batch
   setup.bat
   ```

3. Restart VS Code

4. The MCP auto-connects via `.vscode/mcp.json`

## Using the Tool

AI agents can call the `hello_greeting` tool with:

```json
{
  "name": "Alice"
}
```

Response:
```
Hello, Alice! Greeting from your Go MCP server.
```

If no name is provided, it defaults to greeting the "agent".

## How it Works

- **Language:** Go 1.22
- **Transport:** JSON-RPC over stdin/stdout
- **Dependencies:** None (stdlib only)
- **Cross-platform:** Compiles to Windows, Linux, macOS, ARM

## Build Details

The `setup.bat` script:
- Checks for Docker
- Compiles via Docker (no local Go install needed)
- Outputs: `bin/hello-mcp.exe`
- Auto-removes the build container

Manual build:
```bash
docker run --rm -v "%cd%":/src -w /src golang:1.22 sh -c "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/hello-mcp.exe ."
```

## MCP Implementation

**Methods supported:**
- `initialize` — handshake
- `tools/list` — advertise available tools
- `tools/call` — execute a tool
- `ping` — health check

**Tool schema:**
- `name: string (optional)` — Person to greet
