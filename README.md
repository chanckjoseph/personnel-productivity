# Personnel Productivity

Personal collection of productivity tools and MCP servers for AI agents.

## 📚 Documentation Hub

**Start here:**
- **[TOOLS_REFERENCE.md](TOOLS_REFERENCE.md)** — All 12 tools organized by category with clear naming
- **[DEBUGGING_DEMO.md](DEBUGGING_DEMO.md)** — Real-world debugging example (bank account race condition)

**Detailed guides:**
- [devtools-mcp/DEBUGGING.md](devtools-mcp/DEBUGGING.md) — Complete debugging tool documentation
- [devtools-mcp/README.md](devtools-mcp/README.md) — MCP server setup

## 🛠️ Available Tools (12 Total)

### Project Tools (1)
| Tool | Purpose |
|------|---------|
| `project_structure` | Inspect project layout and code organization |

### Git Tools (3)
| Tool | Purpose |
|------|---------|
| `git_status` | Get repository status (URL, branch, changes) |
| `git_commit` | Stage and commit all changes with message |
| `git_push` | Push commits to GitHub with authentication |

### Debug Tools (8)
| Tool | Purpose |
|------|---------|
| `debug_start_session` | Create new debugging session |
| `debug_session_state` | Query session state and metadata |
| `debug_update_context` | Store debugging context in session |
| `debug_formulate_hypothesis` | Record testable hypothesis |
| `debug_design_experiment` | Design controlled experiment |
| `debug_analyze_results` | Analyze results and validate hypothesis |
| `debug_session_history` | Show iteration history |
| `debug_workflow` | Interactive 6-step debugging orchestrator |

## MCP Servers

This workspace includes MCP servers:

- **[devtools-mcp](./devtools-mcp)** - Main server: Project inspection, Git automation, Scientific debugging
- **[go-hello-mcp](./go-hello-mcp)** - Example server (reference implementation)

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

