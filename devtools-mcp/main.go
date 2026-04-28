package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Global session manager
var sessionMgr *SessionManager

func main() {
	fmt.Fprintln(os.Stderr, "devtools-mcp: server started")
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	// Initialize session manager
	workDir := os.Getenv("GIT_WORK_DIR")
	if workDir == "" {
		workDir = "."
	}
	var err error
	sessionMgr, err = NewSessionManager(workDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize session manager: %v\n", err)
		return
	}

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
					Name:        "debug_start_session",
					Description: "Create a new debugging session to track scientific debugging workflow",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"bug_description": map[string]interface{}{
								"type":        "string",
								"description": "Description of the bug being debugged (optional)",
							},
						},
						"additionalProperties": false,
					},
				},
				{
					Name:        "debug_session_state",
					Description: "Get current session state, metadata, and debugging progress",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"session_id": map[string]interface{}{
								"type":        "string",
								"description": "Session ID to retrieve",
							},
						},
						"required":               []string{"session_id"},
						"additionalProperties":   false,
					},
				},
				{
					Name:        "debug_update_context",
					Description: "Update arbitrary context/metadata in a debugging session",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"session_id": map[string]interface{}{
								"type":        "string",
								"description": "Session ID to update",
							},
							"key": map[string]interface{}{
								"type":        "string",
								"description": "Context key to set",
							},
							"value": map[string]interface{}{
								"description": "Value to store (any type)",
							},
						},
						"required":               []string{"session_id", "key", "value"},
						"additionalProperties":   false,
					},
				},
				{
					Name:        "debug_formulate_hypothesis",
					Description: "Formulate a testable, falsifiable hypothesis for the bug",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"session_id": map[string]interface{}{
								"type":        "string",
								"description": "Session ID",
							},
							"hypothesis_text": map[string]interface{}{
								"type":        "string",
								"description": "The hypothesis statement (must be falsifiable)",
							},
							"expected_outcome": map[string]interface{}{
								"type":        "string",
								"description": "Expected outcome if hypothesis is true (optional)",
							},
						},
						"required":               []string{"session_id", "hypothesis_text"},
						"additionalProperties":   false,
					},
				},
				{
					Name:        "debug_design_experiment",
					Description: "Design an experiment to test a hypothesis",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"session_id": map[string]interface{}{
								"type":        "string",
								"description": "Session ID",
							},
							"hypothesis_id": map[string]interface{}{
								"type":        "string",
								"description": "ID of the hypothesis to test",
							},
							"steps": map[string]interface{}{
								"type":        "array",
								"description": "Array of steps to execute",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
							"independent_vars": map[string]interface{}{
								"type":        "array",
								"description": "Variables to manipulate (optional)",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
							"dependent_vars": map[string]interface{}{
								"type":        "array",
								"description": "Variables to measure (optional)",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
							"control_vars": map[string]interface{}{
								"type":        "array",
								"description": "Variables to keep constant (optional)",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
						},
						"required":               []string{"session_id", "hypothesis_id", "steps"},
						"additionalProperties":   false,
					},
				},
				{
					Name:        "debug_analyze_results",
					Description: "Analyze experiment results to test hypothesis validity",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"session_id": map[string]interface{}{
								"type":        "string",
								"description": "Session ID",
							},
							"experiment_id": map[string]interface{}{
								"type":        "string",
								"description": "ID of the experiment",
							},
							"observations": map[string]interface{}{
								"type":        "string",
								"description": "What was actually observed",
							},
							"conclusion": map[string]interface{}{
								"type":        "string",
								"description": "One of: supported, refuted, inconclusive",
								"enum":        []string{"supported", "refuted", "inconclusive"},
							},
						},
						"required":               []string{"session_id", "experiment_id", "observations", "conclusion"},
						"additionalProperties":   false,
					},
				},
				{
					Name:        "debug_session_history",
					Description: "Display iteration history and debugging progress",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"session_id": map[string]interface{}{
								"type":        "string",
								"description": "Session ID",
							},
						},
						"required":               []string{"session_id"},
						"additionalProperties":   false,
					},
				},
				{
					Name:        "debug_workflow",
					Description: "Interactive 6-step scientific debugging workflow: start → learn → hypothesis → experiment → analyze → fix (or iterate)",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"step": map[string]interface{}{
								"type":        "string",
								"description": "Step: start, learn, hypothesis, experiment, analyze, fix, iterate",
								"enum":        []string{"start", "learn", "hypothesis", "experiment", "analyze", "fix", "iterate"},
							},
							"session_id": map[string]interface{}{
								"type":        "string",
								"description": "Session ID (required for all steps except 'start')",
							},
							"bug_description": map[string]interface{}{
								"type":        "string",
								"description": "Bug description (required for 'start')",
							},
						"bug_observation": map[string]interface{}{
							"type":        "string",
							"description": "What was observed from the bug (required for 'hypothesis')",
						},
						"suspected_component": map[string]interface{}{
							"type":        "string",
							"description": "Suspected code/component causing the bug (required for 'hypothesis')",
						},
						"root_cause_theory": map[string]interface{}{
							"type":        "string",
							"description": "Theory on why this component is broken (required for 'hypothesis')",
						},
						"evidence_chain": map[string]interface{}{
							"type":        "string",
							"description": "How the root cause produces the observed symptom (required for 'hypothesis')",
						},
						"falsification_test": map[string]interface{}{
							"type":        "string",
							"description": "What evidence would prove this hypothesis wrong (required for 'hypothesis')",
							},

							"steps": map[string]interface{}{
								"type":        "array",
								"description": "Experiment steps (required for 'experiment')",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
							"independent_vars": map[string]interface{}{
								"type":        "array",
								"description": "Independent variables (optional for 'experiment')",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
							"dependent_vars": map[string]interface{}{
								"type":        "array",
								"description": "Dependent variables (optional for 'experiment')",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
							"control_vars": map[string]interface{}{
								"type":        "array",
								"description": "Control variables (optional for 'experiment')",
								"items": map[string]interface{}{
									"type": "string",
								},
							},
							"observations": map[string]interface{}{
								"type":        "string",
								"description": "Actual observations (required for 'analyze')",
							},
							"conclusion": map[string]interface{}{
								"type":        "string",
								"description": "Conclusion (required for 'analyze')",
								"enum":        []string{"supported", "refuted", "inconclusive"},
							},
							"fix_description": map[string]interface{}{
								"type":        "string",
								"description": "Fix description (required for 'fix')",
							},
						},
						"required": []string{"step"},
					},
				},
				{
					Name:        "self_build",
					Description: "Self-rebuild: Recompile devtools-mcp from source, kill the running server, and restart VS Code. Use this after modifying devtools-mcp code to test changes immediately. No parameters required.",
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
			case "project_structure":
				handleGetProjectStructure(writer, id, params.Arguments)
			case "git_status":
				handleGitStatus(writer, id, params.Arguments)
			case "git_commit":
				handleGitCommit(writer, id, params.Arguments)
			case "git_push":
				handleGitPush(writer, id, params.Arguments)
			case "debug_start_session":
				handleStartDebugSession(writer, id, params.Arguments)
			case "debug_session_state":
				handleGetSessionState(writer, id, params.Arguments)
			case "debug_update_context":
				handleUpdateSessionContext(writer, id, params.Arguments)
			case "debug_formulate_hypothesis":
				handleFormulateHypothesis(writer, id, params.Arguments)
			case "debug_design_experiment":
				handleDesignExperiment(writer, id, params.Arguments)
			case "debug_analyze_results":
				handleAnalyzeResults(writer, id, params.Arguments)
			case "debug_session_history":
				handleTrackIteration(writer, id, params.Arguments)
			case "debug_workflow":
				handleDebugWorkflow(writer, id, params.Arguments)
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
