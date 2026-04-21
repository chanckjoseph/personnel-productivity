package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type toolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func main() {
	fmt.Fprintln(os.Stderr, "devtools-mcp: server started")
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for {
		msg, err := readFrame(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Fprintf(os.Stderr, "mcp: read error: %v\n", err)
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			fmt.Fprintf(os.Stderr, "mcp: invalid JSON: %v\n", err)
			continue
		}

		if len(req.ID) == 0 {
			if req.Method == "notifications/initialized" {
				continue
			}
			continue
		}

		id, err := parseID(req.ID)
		if err != nil {
			sendError(writer, nil, -32600, "invalid request id")
			continue
		}

		switch req.Method {
		case "initialize":
			result := map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{
						"listChanged": false,
					},
				},
				"serverInfo": map[string]interface{}{
					"name":    "devtools-mcp",
					"version": "1.0.0",
				},
			}
			sendResult(writer, id, result)
		case "tools/list":
			tools := []toolDefinition{
				{
					Name:        "get_project_structure",
					Description: "Returns the directory tree and file organization of the project. Shows folder hierarchy, file names, and structure patterns that help identify project type, architecture, and where code/tests/configuration are located.",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"max_depth": map[string]interface{}{
								"type":        "integer",
								"description": "Maximum depth to traverse (default: 5)",
							},
						},
						"additionalProperties": false,
					},
				},
				{
					Name:        "git_status",
					Description: "Get repository status: URL, current branch, and modified files",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{},
						"additionalProperties": false,
					},
				},
				{
					Name:        "git_commit",
					Description: "Commit all staged changes with a message. Automatically runs 'git add -A' before committing.",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"message": map[string]interface{}{
								"type":        "string",
								"description": "Commit message",
							},
						},
						"required":               []string{"message"},
						"additionalProperties":   false,
					},
				},
				{
					Name:        "git_push",
					Description: "Push commits to GitHub using .pat and .username credentials",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"branch": map[string]interface{}{
								"type":        "string",
								"description": "Branch name to push (default: 'main')",
							},
						},
						"additionalProperties": false,
					},
				},
			}
			sendResult(writer, id, map[string]interface{}{"tools": tools})
		case "tools/call":
			var params toolsCallParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				sendError(writer, id, -32602, "invalid params")
				continue
			}

			switch params.Name {
			case "get_project_structure":
				handleGetProjectStructure(writer, id, params.Arguments)
			case "git_status":
				handleGitStatus(writer, id, params.Arguments)
			case "git_commit":
				handleGitCommit(writer, id, params.Arguments)
			case "git_push":
				handleGitPush(writer, id, params.Arguments)
			default:
				sendError(writer, id, -32602, fmt.Sprintf("unknown tool: %s", params.Name))
			}
		case "ping":
			sendResult(writer, id, map[string]interface{}{})
		default:
			sendError(writer, id, -32601, "method not found")
		}
	}
}

func handleGetProjectStructure(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	workDir := os.Getenv("GIT_WORK_DIR")
	if workDir == "" {
		workDir = "."
	}

	// Get max_depth parameter (default 5)
	maxDepth := 5
	if depthRaw, ok := args["max_depth"]; ok {
		if depthNum, ok := depthRaw.(float64); ok {
			maxDepth = int(depthNum)
		}
	}

	// Directories to ignore (common noise)
	ignoredDirs := map[string]bool{
		".git":            true,
		"node_modules":    true,
		".venv":           true,
		"venv":            true,
		"__pycache__":     true,
		".pytest_cache":   true,
		"bin":             true,
		".vscode":         true,
		".idea":           true,
		"build":           true,
		"dist":            true,
		".DS_Store":       true,
		".playwright-mcp": true,
	}

	// Walk the directory and build tree
	type DirEntry struct {
		Name     string       `json:"name"`
		Type     string       `json:"type"` // "file" or "dir"
		Children []DirEntry  `json:"children,omitempty"`
	}

	var walkDir func(path string, depth int) ([]DirEntry, error)
	walkDir = func(path string, depth int) ([]DirEntry, error) {
		if depth > maxDepth {
			return nil, nil
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		var result []DirEntry
		for _, entry := range entries {
			name := entry.Name()
			
			// Skip ignored directories
			if entry.IsDir() && ignoredDirs[name] {
				continue
			}

			if entry.IsDir() {
				children, _ := walkDir(filepath.Join(path, name), depth+1)
				result = append(result, DirEntry{
					Name:     name + "/",
					Type:     "dir",
					Children: children,
				})
			} else {
				result = append(result, DirEntry{
					Name: name,
					Type: "file",
				})
			}
		}

		return result, nil
	}

	structure, err := walkDir(workDir, 0)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Failed to read directory structure: %v", err))
		return
	}

	// Build text representation for display
	var textOutput strings.Builder
	textOutput.WriteString(fmt.Sprintf("Project Structure (depth: %d)\n\n", maxDepth))

	var buildText func(entries []DirEntry, indent string)
	buildText = func(entries []DirEntry, indent string) {
		for i, entry := range entries {
			isLast := i == len(entries)-1
			prefix := "├── "
			if isLast {
				prefix = "└── "
			}
			textOutput.WriteString(indent + prefix + entry.Name + "\n")

			if entry.Children != nil && len(entry.Children) > 0 {
				nextIndent := indent + "│   "
				if isLast {
					nextIndent = indent + "    "
				}
				buildText(entry.Children, nextIndent)
			}
		}
	}

	buildText(structure, "")

	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": textOutput.String(),
			},
		},
		"structure": structure,
		"max_depth": maxDepth,
	})
}

func handleGitStatus(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	workDir := os.Getenv("GIT_WORK_DIR")
	if workDir == "" {
		workDir = "."
	}

	var output strings.Builder

	// Get repo URL
	getRemoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	getRemoteCmd.Dir = workDir
	remoteOut, err := getRemoteCmd.Output()
	if err != nil {
		sendError(writer, id, -32602, "Failed to get remote URL. Is this a git repository?")
		return
	}
	repoURL := strings.TrimSpace(string(remoteOut))

	// Get current branch
	getBranchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	getBranchCmd.Dir = workDir
	branchOut, err := getBranchCmd.Output()
	if err != nil {
		sendError(writer, id, -32602, "Failed to get current branch")
		return
	}
	currentBranch := strings.TrimSpace(string(branchOut))

	// Get git status in porcelain format (easy to parse)
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = workDir
	statusOut, err := statusCmd.Output()
	if err != nil {
		sendError(writer, id, -32602, "Failed to get git status")
		return
	}

	// Parse status output
	var files []map[string]interface{}
	statusLines := strings.Split(strings.TrimSpace(string(statusOut)), "\n")
	
	modifiedCount := 0
	untrackedCount := 0
	
	for _, line := range statusLines {
		if line == "" {
			continue
		}
		
		// Format: "XY filename"
		if len(line) < 3 {
			continue
		}

		status := line[:2]
		filename := strings.TrimSpace(line[3:])

		// Determine file status
		var fileStatus string
		switch status {
		case "M ":
			fileStatus = "modified"
			modifiedCount++
		case " M":
			fileStatus = "modified (staged)"
			modifiedCount++
		case "A ":
			fileStatus = "added"
			modifiedCount++
		case "D ":
			fileStatus = "deleted"
			modifiedCount++
		case "??":
			fileStatus = "untracked"
			untrackedCount++
		case "MM", "AM", "DM":
			fileStatus = "conflict"
			modifiedCount++
		default:
			fileStatus = "unknown"
		}

		files = append(files, map[string]interface{}{
			"filename": filename,
			"status":   fileStatus,
		})
	}

	// Build summary
	output.WriteString(fmt.Sprintf("Repository: %s\n", repoURL))
	output.WriteString(fmt.Sprintf("Current branch: %s\n", currentBranch))
	output.WriteString(fmt.Sprintf("Modified files: %d, Untracked files: %d\n", modifiedCount, untrackedCount))

	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output.String(),
			},
		},
		"repo_url":       repoURL,
		"current_branch": currentBranch,
		"files":          files,
		"summary": map[string]interface{}{
			"modified":   modifiedCount,
			"untracked":  untrackedCount,
			"total":      len(files),
		},
	})
}

func handleGitCommit(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	// Validate required message
	messageRaw, ok := args["message"]
	if !ok {
		sendError(writer, id, -32602, "missing required argument: message")
		return
	}

	// Parse message
	message, ok := messageRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "message must be a string")
		return
	}

	message = strings.TrimSpace(message)
	if message == "" {
		sendError(writer, id, -32602, "message cannot be empty")
		return
	}

	// Run git commands
	var output strings.Builder

	// git add -A
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = os.Getenv("GIT_WORK_DIR")
	if addCmd.Dir == "" {
		addCmd.Dir = "."
	}

	addOut, err := addCmd.CombinedOutput()
	if err != nil {
		output.WriteString(fmt.Sprintf("git add -A failed: %s\n%s", err, string(addOut)))
	} else {
		output.WriteString("Staged all changes with: git add -A\n")
		if len(addOut) > 0 {
			output.WriteString(string(addOut))
		}
	}

	// git commit -m <message>
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = os.Getenv("GIT_WORK_DIR")
	if commitCmd.Dir == "" {
		commitCmd.Dir = "."
	}

	commitOut, err := commitCmd.CombinedOutput()
	if err != nil {
		output.WriteString(fmt.Sprintf("git commit failed: %s\n%s", err, string(commitOut)))
		sendResult(writer, id, map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		})
		return
	}

	output.WriteString(fmt.Sprintf("Commit successful: %s\n", message))
	if len(commitOut) > 0 {
		output.WriteString(string(commitOut))
	}

	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output.String(),
			},
		},
	})
}

func handleGitPush(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	// Get working directory (where .pat and .username are)
	workDir := os.Getenv("GIT_WORK_DIR")
	if workDir == "" {
		workDir = "."
	}

	// Read .username
	usernamePath := filepath.Join(workDir, ".username")
	usernameBytes, err := os.ReadFile(usernamePath)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf(
			"Cannot read .username file. Create it at: %s\nContent should be your GitHub username (e.g., chanckjoseph)",
			usernamePath))
		return
	}
	username := strings.TrimSpace(string(usernameBytes))
	if username == "" {
		sendError(writer, id, -32602, fmt.Sprintf(
			".username file is empty. Add your GitHub username to: %s",
			usernamePath))
		return
	}

	// Read .pat
	patPath := filepath.Join(workDir, ".pat")
	patBytes, err := os.ReadFile(patPath)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf(
			"Cannot read .pat file. Create it at: %s\nContent should be your GitHub Personal Access Token (https://github.com/settings/tokens)",
			patPath))
		return
	}
	token := strings.TrimSpace(string(patBytes))
	if token == "" {
		sendError(writer, id, -32602, fmt.Sprintf(
			".pat file is empty. Add your GitHub Personal Access Token to: %s",
			patPath))
		return
	}

	// Get branch (default: main)
	branch := "main"
	if branchRaw, ok := args["branch"]; ok {
		if branchStr, ok := branchRaw.(string); ok {
			branch = strings.TrimSpace(branchStr)
		}
	}

	// Get remote origin URL from git config
	getRemoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	getRemoteCmd.Dir = workDir
	remoteOut, err := getRemoteCmd.Output()
	if err != nil {
		sendError(writer, id, -32602, "Cannot read git remote. Is this a git repository?")
		return
	}

	remoteURL := strings.TrimSpace(string(remoteOut))
	// Extract owner/repo from URL
	// Handle both https://github.com/owner/repo.git and git@github.com:owner/repo.git
	if !strings.Contains(remoteURL, "github.com") {
		sendError(writer, id, -32602, "Remote origin is not a GitHub repository")
		return
	}

	// Build authenticated URL
	var authenticatedURL string
	if strings.HasPrefix(remoteURL, "git@") {
		// SSH format: git@github.com:owner/repo.git -> https://github.com/owner/repo.git
		parts := strings.Split(remoteURL, ":")
		if len(parts) == 2 {
			authenticatedURL = fmt.Sprintf("https://%s:%s@github.com/%s", username, token, parts[1])
		} else {
			sendError(writer, id, -32602, "Invalid SSH remote format")
			return
		}
	} else {
		// HTTPS format: https://github.com/owner/repo.git
		// Replace with authenticated URL
		authenticatedURL = fmt.Sprintf("https://%s:%s@github.com/%s",
			username, token, strings.TrimPrefix(remoteURL, "https://github.com/"))
	}

	// Run git push
	pushCmd := exec.Command("git", "push", authenticatedURL, branch, "--set-upstream")
	pushCmd.Dir = workDir

	var output strings.Builder
	pushOut, err := pushCmd.CombinedOutput()
	if err != nil {
		output.WriteString(fmt.Sprintf("git push failed: %s\n%s", err, string(pushOut)))
		sendResult(writer, id, map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		})
		return
	}

	output.WriteString(fmt.Sprintf("Push successful to branch: %s\n", branch))
	if len(pushOut) > 0 {
		output.WriteString(string(pushOut))
	}

	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output.String(),
			},
		},
	})
}

func parseID(raw json.RawMessage) (interface{}, error) {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString, nil
	}

	var asNumber float64
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return asNumber, nil
	}

	return nil, fmt.Errorf("unsupported id type")
}

func sendResult(writer *bufio.Writer, id interface{}, result interface{}) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	send(writer, resp)
}

func sendError(writer *bufio.Writer, id interface{}, code int, message string) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
		},
	}
	send(writer, resp)
}

func send(writer *bufio.Writer, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcp: marshal error: %v\n", err)
		return
	}
	body = append(body, '\n')
	if _, err := writer.Write(body); err != nil {
		fmt.Fprintf(os.Stderr, "mcp: write error: %v\n", err)
		return
	}
	_ = writer.Flush()
}

func readFrame(reader *bufio.Reader) ([]byte, error) {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(line) > 0 {
				return []byte(line), nil
			}
			return nil, err
		}

		line = strings.TrimSuffix(line, "\n")
		if line != "" {
			return []byte(line), nil
		}
	}
}
