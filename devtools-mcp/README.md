# devtools-mcp

**MCP server** for automating developer workflows with intelligent query parsing, Git operations, and project exploration.

## What It Does

Provides AI agents with tools to:
- ✅ **Parse Queries** — Transform free-form prompts into structured task plans with dependencies
- ✅ **Automate Git** — Commit and push without credential hassles
- ✅ **Project Exploration** — Understand project structure and repository status

The server handles credentials automatically (reads `.pat` and `.username`), so agents never expose tokens.

## 🚀 For Agents: Setup Instructions

### 1. Configure MCP Connection

In your project's `.vscode/mcp.json`:

```json
{
  "mcpServers": {
    "devtools-mcp": {
      "type": "stdio",
      "command": "${workspaceFolder}/devtools-mcp/bin/devtools-mcp.exe"
    }
  }
}
```

Replace path based on your OS:
- Windows: `bin/devtools-mcp.exe`
- Linux/macOS: `bin/devtools-mcp`

### 2. Create Credential Files

In your **project root** (same level as `.vscode/`):

**`.pat`** — GitHub Personal Access Token
```
github_pat_11A...
```

**`.username`** — GitHub username
```
your_username
```

**On Linux/macOS, secure them:**
```bash
chmod 600 .pat .username
```

### 3. Verify Setup

You're ready. The server will auto-connect. Available tools:
- `parse_query` — Query parsing and task breakdown
- `project_structure` — Project exploration  
- `git_status` — Repository status
- `git_commit`, `git_push` — Git automation
- `self_build` — Self-rebuild and restart

---

## 📊 Architecture & Project Structure

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    devtools-mcp Server                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              MCP Protocol Handler                    │   │
│  │  (main.go: initialize, tools/list, tools/call)     │   │
│  └──────────────────────────────────────────────────────┘   │
│                           │                                  │
│           ┌───────────────┼───────────────┐                 │
│           │               │               │                 │
│           ▼               ▼               ▼                 │
│  ┌─────────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │  Query Engine   │  │ Git Handler │  │   Project   │     │
│  │ (query.go)      │  │ (git.go)    │  │   Handler   │     │
│  │                 │  │             │  │ (project.go)│     │
│  │ • Parse         │  │ • Status    │  │             │     │
│  │ • Validate      │  │ • Commit    │  │ • Structure │     │
│  │ • Enrich        │  │ • Push      │  │             │     │
│  └─────────────────┘  └─────────────┘  └─────────────┘     │
│           │                                                  │
│           ▼                                                  │
│  ┌─────────────────────────────────────┐                    │
│  │  Prompt Analyzer & Query Types      │                    │
│  │  (prompt_analyzer.go, query_types.go)                   │
│  │                                     │                    │
│  │  • Intent extraction                │                    │
│  │  • Task breakdown                   │                    │
│  │  • Requirement analysis             │                    │
│  │  • Constraint identification        │                    │
│  └─────────────────────────────────────┘                    │
│           │                                                  │
│           ▼                                                  │
│  ┌──────────────────────────────────────┐                   │
│  │     Artifact Store                   │                   │
│  │  (artifact_store.go)                 │                   │
│  │                                      │                   │
│  │  Caches & indexes parsed queries     │                   │
│  └──────────────────────────────────────┘                   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Code Organization

| File | Purpose |
|------|---------|
| `main.go` | MCP protocol handler, tool routing |
| `query.go` | QueryEngine orchestration, validation, metadata |
| `query_types.go` | Data structures (Query, Task, Intent, Requirements) |
| `prompt_analyzer.go` | LocalAnalyzer for parsing prompts locally |
| `artifact_store.go` | Query caching and manifest management |
| `git.go` | Git command handlers (status, commit, push) |
| `project.go` | Project structure exploration & query parsing handler |
| `types.go` | MCP protocol types (rpcRequest, toolDefinition) |
| `utils.go` | Helper functions (readFrame, sendResult, sendError) |
| `manager.go` | Manager stubs/documentation |

---

## 🔧 For Developers: Building the Binary

### Prerequisites
- Docker (required for cross-platform builds)

### Build for Current Platform

**Windows:**
```batch
setup.bat
```

**Linux/macOS:**
```bash
chmod +x setup.sh
./setup.sh
```

Output: `bin/devtools-mcp` (or `.exe` on Windows)

### Build All Platforms

To create binaries for Windows, Linux, and macOS:

```bash
chmod +x build-all.sh
./build-all.sh
```

Generates:
- `bin/devtools-mcp.exe` (Windows)
- `bin/devtools-mcp` (Linux)
- `bin/devtools-mcp-darwin-amd64` (macOS Intel)
- `bin/devtools-mcp-darwin-arm64` (macOS Apple Silicon)

---

## 📚 Available Tools & API Reference

### parse_query

**Parse a free-form prompt into a structured Query with intent, tasks, requirements, and constraints.**

Transforms user requests into actionable plans with dependencies and information requirements.

**Input:**
```json
{
  "prompt": "Create a new feature for user authentication"
}
```

**Response Structure:**
```json
{
  "query": {
    "id": "2026-05-04T12-30-45-123",
    "original_text": "Create a new feature for user authentication",
    "intent": {
      "primary": "Create a new feature for user authentication",
      "secondary": ["Secure user data", "Improve login flow"],
      "urgency": "medium",
      "scope": "project",
      "domain": "software",
      "ambiguities": ["Authentication method?", "Database system?"]
    },
    "tasks": [
      {
        "id": "task_1",
        "title": "Design authentication system",
        "description": "Plan user auth architecture",
        "type": "action",
        "prerequisites": [],
        "parallelizable": false,
        "expected_output": "Architecture document"
      }
    ],
    "information_requirements": {
      "required": [
        {
          "id": "req_auth_method",
          "name": "Authentication Method",
          "description": "How users will authenticate",
          "type": "choice",
          "status": "missing",
          "sources": []
        }
      ],
      "optional": [],
      "derived": []
    },
    "constraints": [
      {
        "type": "scope",
        "description": "Project scope includes backend only",
        "impact": "medium"
      }
    ],
    "metadata": {
      "query_id": "2026-05-04T12-30-45-123",
      "created_at": "2026-05-04T12:30:45Z",
      "analyzer_model": "local-v1",
      "execution_time_ms": 45,
      "confidence": 0.87,
      "task_count": 3,
      "requirement_count": 5,
      "status": "success",
      "schema_version": "1.0"
    }
  }
}
```

---

### project_structure

Explore the directory structure and organization of the project. Useful for understanding layout, finding where to make changes, and identifying project type from the file organization.

**Input:**
```json
{
  "max_depth": 5
}
```

`max_depth` is optional (default: 5). Controls how deep to traverse the directory tree.

**Response:** Tree view showing folders and files, ignoring common noise like `.git`, `node_modules`, `bin`, `.vscode`, etc.

**Example:**
```
Project Structure (depth: 5)

├── devtools-mcp/
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   ├── query.go
│   ├── query_types.go
│   ├── prompt_analyzer.go
│   ├── artifact_store.go
│   ├── git.go
│   ├── project.go
│   └── bin/
│       └── devtools-mcp.exe
├── go-hello-mcp/
│   ├── main.go
│   ├── go.mod
│   └── bin/
│       └── hello-mcp.exe
└── README.md
```

**What to infer:**
- See `go.mod` → Go project
- See `main.go` → Entry point
- See `query.go` → Query parsing capability
- See `Dockerfile` → Containerized

---

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

---

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

---

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

---

### self_build

Self-rebuild: Recompile devtools-mcp from source, kill the running server, and restart VS Code.

**Input:**
```json
{}
```

**What it does:**
1. Recompiles the binary from current source
2. Kills the running MCP server process
3. Restarts VS Code to load the new binary

**Use when:**
- You've modified devtools-mcp source code
- You want to test changes immediately without manual rebuild

---

## 🔐 Credentials & Security

### Create .pat file

1. Go to https://github.com/settings/tokens
2. Generate a new Personal Access Token with `repo` scope
3. Save the token to `.pat` in the workspace root:
   ```
   github_pat_11A...
   ```

### Create .username file

Save your GitHub username to `.username`:
```
chanckjoseph
```

### Security Notes

- `.pat` and `.username` are in `.gitignore` (never committed)
- On Linux/macOS, secure them: `chmod 600 .pat .username`
- The server reads these files locally; credentials never pass through the network
- Tokens are used only for Git operations, never exposed in API responses

---

## 🔧 Implementation Details

- **Language:** Go 1.22
- **Transport:** JSON-RPC 2.0 over stdin/stdout
- **Dependencies:** Standard library only (no external packages)
- **Error Handling:** Returns helpful error messages when credentials are missing

### File-by-File Responsibilities

| File | Exports |
|------|---------|
| `main.go` | MCP server loop, protocol routing |
| `query.go` | `NewQueryEngine`, `Parse()` |
| `query_types.go` | `Query`, `Intent`, `Task`, `InformationRequirement` |
| `prompt_analyzer.go` | `NewLocalAnalyzer`, `AnalyzePrompt()` |
| `artifact_store.go` | `NewArtifactStore`, query caching |
| `git.go` | `handleGitStatus`, `handleGitCommit`, `handleGitPush` |
| `project.go` | `handleGetProjectStructure`, `handleParseQuery` |
| `types.go` | `rpcRequest`, `toolDefinition`, `toolsCallParams` |
| `utils.go` | `readFrame()`, `sendResult()`, `sendError()` |

---

## 🚀 Recompiling After Changes

If you modify the source code or pull new changes, recompile:

**Windows:**
```batch
setup.bat
```

**Linux/macOS:**
```bash
bash setup.sh
```

Then restart VS Code to load the new binary.

**When to recompile:**
- Source code modifications
- Pull updates from GitHub
- Add new tools to the MCP
- Binary seems out of sync with source code
