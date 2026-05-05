package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// MCP Server for devtools

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

		// Skip if no ID (notification)
		if req.ID == nil {
			if req.Method == "notifications/initialized" {
				continue
			}
			continue
		}

		// ID is already parsed, just use it directly
		id := req.ID

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
					Name:        "structure_prompt",
					Description: "Parse a free-form prompt into a structured Query template for agent enrichment.",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"prompt": map[string]interface{}{
								"type":        "string",
								"description": "The free-form prompt to parse",
							},
						},
						"required":               []string{"prompt"},
						"additionalProperties":   false,
					},
				},
				{
					Name:        "project_structure",
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
				{
					Name:        "self_build",
					Description: "Provides instructions to manually rebuild devtools-mcp from source and restart VS Code. Use after modifying devtools-mcp code to test changes.",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{},
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
			case "structure_prompt":
			handleParseQuery(writer, id, params.Arguments)
			case "project_structure":
				handleGetProjectStructure(writer, id, params.Arguments)
			case "git_status":
				handleGitStatus(writer, id, params.Arguments)
			case "git_commit":
				handleGitCommit(writer, id, params.Arguments)
			case "git_push":
				handleGitPush(writer, id, params.Arguments)
			case "self_build":
				handleSelfBuild(writer, id, params.Arguments)
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
