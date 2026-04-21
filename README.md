# Personnel Productivity

Personal collection of productivity tools and MCP servers for AI agents.

## MCP Servers

This workspace includes two Model Context Protocol (MCP) servers:

- **[hello-go-mcp](./go-hello-mcp)** - Simple greeting tool (example/reference)
- **[devtools-mcp](./devtools-mcp)** - Git workflow automation (commit, push, status)

### Getting Started with MCPs

1. See [.vscode/mcp.json](.vscode/mcp.json) for server configuration
2. Follow setup instructions in each MCP's README
3. Agents can call these tools after setup and VS Code restart

### Development

If you modify MCP source code:
1. Edit the source files in `devtools-mcp/main.go` or `go-hello-mcp/main.go`
2. Recompile: `cd devtools-mcp && ./setup.bat` (or `bash setup.sh` on Linux/Mac)
3. Restart VS Code to load the new binary

See individual MCP READMEs for detailed development information.

## Tools
- [md-to-docx](./md-to-docx): Service to convert Markdown to DOCX.

## Docker Management

Use the manage_docker.py script to control the application container.

*   **Start**: python manage_docker.py up (Builds if needed, starts on port 8989)
*   **Stop**: python manage_docker.py down
*   **Rebuild & Start**: python manage_docker.py restart
*   **Logs**: python manage_docker.py logs

Access the application at [http://localhost:8989](http://localhost:8989).

