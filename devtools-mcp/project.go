package main

import (
	"bufio"
	"encoding/json"
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

// handleParseQuery returns template for agent to enrich
func handleParseQuery(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
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

	// Create template for agent to fill in, with example guidance fields
	queryID := generateQueryID()
	
	// Create example information requirements with guidance fields to show agent the expected structure
	exampleRequirement := InformationRequirement{
		ID:                  "req_example",
		Name:                "Example Information Need",
		Description:         "This is an example of what an information requirement should look like",
		Type:                "text",
		Status:              "missing",
		Sources:             []string{"web_search", "documentation", "knowledge_base"},
		Constraints:         "Must be specific and verifiable",
		SearchHints:         "Specific web search queries like 'search term 1' or check 'source 2' for this info",
		InferenceStrategy:   "If not found via search, assume X based on Y context; fallback to Z",
		ConfidenceIfMissing: "medium",
		DerivableFrom:       []string{},
	}
	
	// Initialize template for agent to fill in, with guidance fields
	query := Query{
		ID:           queryID,
		OriginalText: prompt,
		Intent: Intent{
			Primary:     "",
			Secondary:   []string{},
			Urgency:     "",
			Scope:       "",
			Domain:      "",
			Ambiguities: []string{},
		},
		Tasks:        []Task{},
		Requirements: InformationRequires{
			Required: []InformationRequirement{exampleRequirement},
			Optional: []InformationRequirement{},
			Derived:  []InformationRequirement{},
		},
		Constraints: []Constraint{},
		Metadata: Metadata{
			QueryID:        queryID,
			CreatedAt:      getCurrentTime(),
			AnalyzerModel:  "agent-discovery-guided",
			Confidence:     0.0,
			Status:         "pending_enrichment",
			SchemaVersion:  "1.0",
			ResolutionStrategy: &ResolutionStrategy{
				Order:    []string{},
				Approach: "Agent should: 1) Analyze prompt to identify information gaps; 2) Create requirements following the example structure (with search_hints, inference_strategy, confidence_if_missing); 3) Prioritize by criticality; 4) Apply inference strategies for missing data; 5) Report confidence levels",
				Fallback: "If critical information cannot be found or inferred, provide speculative answer with clear confidence scores and uncertainty notes",
			},
		},
	}

	// Return template + prompt + instructions for agent to enrich
	templateJSON, err := json.Marshal(query)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Failed to marshal template: %v", err))
		return
	}

	responseText := fmt.Sprintf("Query ID: %s\n\nTemplate (fill in the fields below):\n%s", queryID, string(templateJSON))

	// Create enrichment todo list with embedded guidance - todos are self-governing
	enrichmentTodos := []map[string]interface{}{
		{
			"id":     1,
			"title":  "Analyze intent",
			"status": "not-started",
			"description": "Fill intent.primary, intent.domain (software/personal/research/planning/writing/business), intent.urgency (low/medium/high/critical), intent.scope (local/project/team/org), and intent.ambiguities (list unclear aspects)",
		},
		{
			"id":     2,
			"title":  "Populate required information requirements",
			"status": "not-started",
			"description": "For each critical piece of info needed: populate id, name, description, type, sources (array), search_hints (specific queries), inference_strategy (what to assume if not found), confidence_if_missing (high/medium/low). See req_example in template for structure.",
		},
		{
			"id":     3,
			"title":  "Identify optional information requirements",
			"status": "not-started",
			"description": "Add nice-to-have info to optional array using same field structure as required (sources, search_hints, inference_strategy, etc.)",
		},
		{
			"id":     4,
			"title":  "Identify derived information requirements",
			"status": "not-started",
			"description": "Add derived info to derived array. For each: specify derivable_from (list IDs of prerequisite requirements it depends on), and populate all guidance fields.",
		},
		{
			"id":     5,
			"title":  "Analyze constraints & risks",
			"status": "not-started",
			"description": "Populate constraints array with: limitations, restrictions, permissions needed, known risks that could prevent completion",
		},
		{
			"id":     6,
			"title":  "Update resolution strategy",
			"status": "not-started",
			"description": "Set resolution_strategy.order to array of requirement IDs in priority order (critical first). Later requirements depend on earlier ones.",
		},
		{
			"id":     7,
			"title":  "Save enriched query",
			"status": "not-started",
			"description": "Mark all todos as completed. Save enriched JSON (entire query object + todos array) to queries_cache/{query_id}.json",
		},
	}

	// Build enhanced instructions emphasizing 7-step workflow
	enhancedInstructions := `========================================
YOUR WORKFLOW: CREATE A 7-STEP TODO LIST
========================================

You have received a Query template to enrich. Use the manage_todo_list tool to create and track these 7 steps:

**STEP 1:** Analyze intent (primary goal, domain, urgency, scope, ambiguities)
**STEP 2:** Populate required information requirements (critical data needed)
**STEP 3:** Identify optional information requirements (nice-to-have data)
**STEP 4:** Identify derived information requirements (data derivable from requirements)
**STEP 5:** Analyze constraints & risks (limitations, permissions, risks)
**STEP 6:** Update resolution strategy (order requirements by priority)
**STEP 7:** Save enriched query (use create_file tool to save enriched template + todos to queries_cache/{query_id}.json)

========================================

WORKFLOW GUIDANCE:
1. Call manage_todo_list to create a list with all 7 steps
2. Mark each step IN-PROGRESS as you work
3. Mark each step COMPLETED immediately after finishing
4. Let your todo list track progress through the entire workflow
5. In step 7, use create_file tool to persist enriched template and todos

Reference structure is available in the template (req_example shows required info requirements format).`

	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": enhancedInstructions,
			},
			{
				"type": "text",
				"text": responseText,
			},
		},
		"status":          "pending_enrichment",
		"query_id":        queryID,
		"prompt":          prompt,
		"template":        string(templateJSON),
		"todos":           enrichmentTodos,
	})
}
