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

	discoveryInstructions := `ENRICH THIS QUERY TEMPLATE BY ANALYZING THE PROMPT:

IMPORTANT: Each information requirement MUST include these guidance fields:
- sources: Array of where to search (web_search, linkedin, twitter, github, knowledge_base, etc.)
- search_hints: Specific search queries or platform guidance (e.g., "Search 'term1' on LinkedIn" or "Check 'source2' documentation")
- inference_strategy: What to assume if info cannot be found (include fallback logic)
- confidence_if_missing: Confidence without this data - high/medium/low
- derivable_from: For derived requirements, list IDs of prerequisite requirements

See the 'required' array for an example structure (req_example) showing all fields populated correctly.

DISCOVERY QUESTIONS:

1. INTENT ANALYSIS:
   - What is the user fundamentally trying to accomplish? (fill: intent.primary)
   - What domain does this belong to? (software/personal/research/planning/writing/business/other)
   - How urgent is this? (low/medium/high/critical)
   - What is the scope? (local/project/team/organization)
   - What ambiguities or unclear aspects exist in the prompt?

2. TASK BREAKDOWN:
   - What are the concrete steps needed to address this?
   - What must happen first? What depends on what?
   - Can any tasks run in parallel?
   - What is the expected output of each task?

3. INFORMATION REQUIREMENTS (WITH SEARCH GUIDANCE):
   - What information is absolutely required to proceed? (required)
     * For each requirement: populate ALL guidance fields (sources, search_hints, inference_strategy, confidence_if_missing)
   - What information would be nice to have but isn't critical? (optional)
   - What information can be derived/inferred from other data? (derived)
     * For derived info: specify which requirements it depends on (derivable_from)

4. CONSTRAINTS & RISKS:
   - What limitations or restrictions apply?
   - What could prevent successful completion?
   - What permissions or access might be needed?
   - What are the known risks?

INSTRUCTIONS:
1. Use semantic understanding, not just keyword matching
2. For EVERY information requirement, populate search_hints with specific queries
3. For EVERY requirement, include inference_strategy - what to assume if not found
4. Prioritize requirements by criticality (required vs optional vs derived)
5. Update resolution_strategy.order with the priority order for filling gaps
6. Fill in ALL fields in the template based on your analysis
7. Save the enriched JSON to {workspace_root}/.cache/queries/{query_id}.json`

	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": responseText,
			},
		},
		"status":       "pending_enrichment",
		"query_id":     queryID,
		"prompt":       prompt,
		"template":     string(templateJSON),
		"cache_dir":    ".cache/queries",
		"instructions": discoveryInstructions,
	})
}
