# devtools-mcp

MCP server for automating developer workflows: Git commit and push operations with credential management.

**What it does:** Provides `git_commit` and `git_push` tools that agents can call without worrying about credentials. The server reads `.pat` and `.username` from the workspace root and handles authentication automatically.

## Quick Start (Windows)

1. **Make sure Docker is running** (or install from https://www.docker.com/products/docker-desktop)

2. **Ensure repo has credentials:**
   - [`.pat`](../.pat) - Your GitHub Personal Access Token
   - [`.username`](../.username) - Your GitHub username

3. From this folder (`devtools-mcp/`), run:
   ```batch
   setup.bat
   ```

4. Update `.vscode/mcp.json` to include:
   ```json
   "devtools-mcp": {
     "type": "stdio",
     "command": "${workspaceFolder}/devtools-mcp/bin/devtools-mcp.exe"
   }
   ```

5. Restart VS Code

## Using the Tools

### get_project_structure

Explore the directory structure and organization of the project. Useful for understanding layout, finding where to make changes, and identifying project type from the file organization.

**Input:**
```json
{
  "max_depth": 3
}
```

`max_depth` is optional (default: 3). Controls how deep to traverse the directory tree.

**Response:** Tree view showing folders and files, ignoring common noise like `.git`, `node_modules`, `bin`, `.vscode`, etc.

**Example:**
```
Project Structure (depth: 3)

├── devtools-mcp/
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   ├── setup.bat
│   ├── setup.sh
│   ├── README.md
│   └── bin/
│       └── devtools-mcp.exe
├── go-hello-mcp/
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   └── bin/
│       └── hello-mcp.exe
├── md-to-docx/
├── .vscode/
│   └── mcp.json
└── README.md
```

**What to infer:**
- See `go.mod` → Go project
- See `main.go` → Entry point
- See `Dockerfile` → Containerized
- See `test/` or `tests/` → Has unit tests

### git_status

Check repository status: URL, current branch, and list of modified/untracked files.

**Input:**
```json
{}
```

**Response:**
```json
{
  "repo_url": "https://github.com/chanckjoseph/personnel-productivity.git",
  "current_branch": "main",
  "files": [
    {"filename": "src/main.go", "status": "modified"},
    {"filename": "README.md", "status": "untracked"}
  ],
  "summary": {
    "modified": 1,
    "untracked": 1,
    "total": 2
  }
}
```

### git_commit

Commit all staged changes with a message. Automatically runs `git add -A` to stage everything (respecting `.gitignore`), then commits with your message.

**Input:**
```json
{
  "message": "Add new feature"
}
```

**What it does:**
1. Runs `git add -A` to stage all modified/new files (except those in `.gitignore`)
2. Runs `git commit -m "your message"`

**Response:** Git output showing what was staged and the commit result

### git_push

Push commits to GitHub using stored credentials.

**Input:**
```json
{
  "branch": "main"
}
```

Or omit `branch` to use the default "main".

**Response:** Git push output and status

## How it Works

- **Language:** Go 1.22
- **Transport:** JSON-RPC over stdin/stdout
- **Dependencies:** None (stdlib + os/exec)
- **Credentials:** Reads from `.pat` and `.username` files (never exposed in agent commands)

## Prerequisites

### Create .pat file

1. Go to https://github.com/settings/tokens
2. Generate a new Personal Access Token with `repo` scope
3. Save the token to [`.pat`](../.pat) in the workspace root:
   ```
   github_pat_11A...
   ```

### Create .username file

Save your GitHub username to [`.username`](../.username):
```
chanckjoseph
```

## Error Handling

If `.pat` or `.username` are missing, tools return helpful error messages:

```
Cannot read .pat file. Create it at: /path/to/.pat
Content should be your GitHub Personal Access Token (https://github.com/settings/tokens)
```

This guides users (or agents reporting to users) on what to do.

## Build Details

The `setup.bat` script:
- Checks for Docker
- Compiles via Docker (no local Go install needed)
- Outputs: `bin/devtools-mcp.exe`
- Auto-removes the build container

Manual build:
```bash
docker run --rm -v "%cd%":/src -w /src golang:1.22 sh -c "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/devtools-mcp.exe ."
```

## Recompiling After Changes

If you modify the source code in `main.go` or pull new changes from GitHub, you need to recompile the binary.

**Windows:**
1. Stop the MCP server (or manually close it in VS Code)
2. From the `devtools-mcp/` folder, run:
   ```batch
   setup.bat
   ```
3. Restart VS Code to load the new binary

**Linux/macOS:**
1. Stop the MCP server
2. From the `devtools-mcp/` folder, run:
   ```bash
   bash setup.sh
   ```
3. Restart VS Code to load the new binary

**When to recompile:**
- You edited `main.go` or other source files
- You pulled changes from GitHub
- You want to add new tools to the MCP
- The binary seems out of sync with the code

**After recompiling:**
- Always restart VS Code for the changes to take effect
- The new binary will be loaded when VS Code restarts

## Tool Schemas

**get_project_structure:**
- `max_depth` (integer, optional) - How deep to traverse (default: 3)

**git_status:**
- No parameters required

**git_commit:**
- `message` (string, required) - Commit message

**git_push:**
- `branch` (string, optional) - Branch name (defaults to "main")

## Configuration

In [`.vscode/mcp.json`](../.vscode/mcp.json), both servers will look like:

```json
{
  "servers": {
    "hello-go-mcp": { ... },
    "devtools-mcp": {
      "type": "stdio",
      "command": "${workspaceFolder}/devtools-mcp/bin/devtools-mcp.exe"
    }
  }
}
```

Agents can now call either tool set independently.
