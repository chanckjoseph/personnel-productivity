package main

import (
	"time"
)

// Query represents a complete semantic transformation of a free-form prompt
// into structured, actionable instructions with dependencies and information requirements.
type Query struct {
	ID           string              `json:"id"`                      // Unique identifier (timestamp-based)
	OriginalText string              `json:"original_text"`           // The original free-form prompt
	Intent       Intent              `json:"intent"`                  // What the user wants to accomplish
	Tasks        []Task              `json:"tasks"`                   // Hierarchical breakdown of tasks
	Requirements InformationRequires `json:"information_requirements"` // Data needed to execute
	Constraints  []Constraint        `json:"constraints"`             // Scope, permissions, risks
	Metadata     Metadata            `json:"metadata"`                // Execution and quality info
}

// Intent describes the primary goal and high-level objectives.
type Intent struct {
	Primary     string   `json:"primary"`      // Main goal in 1-2 sentences
	Secondary   []string `json:"secondary"`    // Related/supporting goals
	Urgency     string   `json:"urgency"`      // "low", "medium", "high", "critical"
	Scope       string   `json:"scope"`        // "local", "project", "team", "organization", etc.
	Domain      string   `json:"domain"`       // "software", "writing", "research", "planning", etc.
	Ambiguities []string `json:"ambiguities"`  // Unclear aspects user should clarify
}

// Task represents a single unit of work or action.
type Task struct {
	ID              string                 `json:"id"`                   // e.g. "task_1", "task_1_1"
	Title           string                 `json:"title"`                // What this task does
	Description     string                 `json:"description"`          // Detailed explanation
	Type            string                 `json:"type"`                 // "action", "decision", "review", "output"
	Prerequisites   []string               `json:"prerequisites"`        // IDs of tasks that must complete first
	Parallelizable  bool                   `json:"parallelizable"`       // Can run simultaneously with siblings
	Requirements    []InformationRequirement `json:"requirements"`        // Info needed for THIS task
	Parameters      map[string]interface{} `json:"parameters"`           // Task-specific inputs/config
	ExpectedOutput  string                 `json:"expected_output"`      // What success looks like
	Subtasks        []Task                 `json:"subtasks,omitempty"`   // Nested breakdown
}

// InformationRequires groups all information needs for the query.
type InformationRequires struct {
	Required []InformationRequirement `json:"required"`   // Critical info needed
	Optional []InformationRequirement `json:"optional"`   // Nice-to-have info
	Derived  []InformationRequirement `json:"derived"`    // Can be inferred/computed from other data
}

// InformationRequirement describes a single piece of data the query needs.
type InformationRequirement struct {
	ID                  string              `json:"id"`                // Unique identifier (e.g. "req_commit_message")
	Name                string              `json:"name"`              // Human-readable name
	Description         string              `json:"description"`       // What this is and why it's needed
	Type                string              `json:"type"`              // "text", "number", "boolean", "choice", "file_path", "datetime", etc.
	Status              string              `json:"status"`            // "available", "missing", "unknown", "error"
	Sources             []string            `json:"sources"`           // Where to search (web_search, linkedin, twitter, etc.)
	Constraints         string              `json:"constraints,omitempty"`   // Min/max length, allowed values, format, etc.
	SearchHints         string              `json:"search_hints,omitempty"` // Specific guidance on how to search for this
	InferenceStrategy   string              `json:"inference_strategy,omitempty"` // What to assume if not found
	ConfidenceIfMissing string              `json:"confidence_if_missing,omitempty"` // Confidence level without this data
	DerivableFrom       []string            `json:"derivable_from,omitempty"` // Which other requirements this depends on
	MissingChain        *MissingInfoChain   `json:"missing_chain,omitempty"` // Recursive dependency if missing
	DefaultValue        interface{}         `json:"default_value,omitempty"` // If available, pre-fill here
}

// Source describes where information can be obtained.
type Source struct {
	Type        string `json:"type"`        // "user_input", "inference", "git_status", "file_system", "environment", "external_api", "hardcoded", etc.
	Location    string `json:"location"`    // How/where to get it (e.g., "git branch name", "current working directory", "user prompt")
	Confidence  string `json:"confidence"`  // "high", "medium", "low"
	Description string `json:"description"` // Additional context
}

// MissingInfoChain represents recursive dependencies when information is missing.
// It allows the LLM/user to understand what must be resolved first.
type MissingInfoChain struct {
	MissingID       string                  `json:"missing_id"`        // ID of what's missing
	MissingName     string                  `json:"missing_name"`      // Name of what's missing
	DependsOn       []InformationRequirement `json:"depends_on"`        // What's needed to get the missing info
	Resolution      string                  `json:"resolution"`        // How to resolve this chain
	Depth           int                     `json:"depth"`             // Recursion depth (for UI rendering)
}

// Constraint represents a limitation, risk, or requirement.
type Constraint struct {
	Type        string `json:"type"`        // "permission", "scope", "risk", "time", "resource", "dependency"
	Description string `json:"description"` // What the constraint is
	Impact      string `json:"impact"`      // "low", "medium", "high"
	Mitigation  string `json:"mitigation,omitempty"` // How to work around it
}

// ResolutionStrategy guides the agent on how to fill in missing information requirements
type ResolutionStrategy struct {
	Order     []string `json:"order"`     // Priority order for resolving requirements
	Approach  string   `json:"approach"`  // High-level strategy for filling gaps
	Fallback  string   `json:"fallback"`  // How to handle partial/speculative answers
}

// Metadata contains execution and quality information about the query analysis.
type Metadata struct {
	QueryID          string              `json:"query_id"`         // Unique ID (timestamp-based: 2026-05-04T12-30-45-123)
	CreatedAt        time.Time           `json:"created_at"`       // When this query was analyzed
	AnalyzerModel    string              `json:"analyzer_model"`   // Which analyzer model was used (e.g., "local-v1")
	ExecutionTime    int64               `json:"execution_time_ms"` // How long analysis took (milliseconds)
	Confidence       float64             `json:"confidence"`       // 0.0 to 1.0 - how confident in this breakdown
	TaskCount        int                 `json:"task_count"`       // Total number of tasks
	RequirementCount int                 `json:"requirement_count"` // Total information requirements
	Status           string              `json:"status"`           // "success", "partial", "error"
	ErrorMessage     string              `json:"error_message,omitempty"` // If status is "error"
	SchemaVersion    string              `json:"schema_version"`   // Version of Query schema used (e.g., "1.0")
	Notes            string              `json:"notes,omitempty"`  // Additional context or observations
	ResolutionStrategy *ResolutionStrategy `json:"resolution_strategy,omitempty"` // Guidance on filling information gaps
}
