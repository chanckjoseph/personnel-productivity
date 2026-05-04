package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DirEntry represents a directory or file entry in the project structure
type DirEntry struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "dir" or "file"
	Children []DirEntry `json:"children,omitempty"`
}

// handleGetProjectStructure returns the directory tree and file organization of the project
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

// handleParseQuery parses a free-form prompt into a structured Query
func handleParseQuery(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	// Extract prompt from arguments
	promptRaw, ok := args["prompt"]
	if !ok {
		sendError(writer, id, -32602, "prompt parameter is required")
		return
	}

	prompt, ok := promptRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "prompt must be a string")
		return
	}

	if prompt == "" {
		sendError(writer, id, -32602, "prompt cannot be empty")
		return
	}

	// Get working directory
	workDir := os.Getenv("GIT_WORK_DIR")
	if workDir == "" {
		workDir = "."
	}

	// Initialize artifact store for caching queries
	artifactStore, err := NewArtifactStore(workDir)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Failed to initialize artifact store: %v", err))
		return
	}

	// Create query engine with local analyzer
	analyzer := NewLocalAnalyzer("local-v1")
	queryEngine := NewQueryEngine(analyzer, artifactStore)

	// Parse the prompt into a structured Query
	query, err := queryEngine.Parse(prompt)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Failed to parse query: %v", err))
		return
	}

	// Return the structured query
	sendResult(writer, id, map[string]interface{}{
		"query": query,
	})
}
