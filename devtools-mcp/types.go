package main

import (
	"encoding/json"
	"time"
)

// MCP Protocol Types

// rpcRequest represents an incoming JSON-RPC 2.0 request
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// rpcResponse represents an outgoing JSON-RPC 2.0 response
type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// rpcError represents an error response
type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// toolDefinition defines an MCP tool
type toolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// toolsCallParams represents parameters for tools/call
type toolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// SessionContext holds all debugging session state
type SessionContext struct {
	ID                 string                 `json:"id"`
	StartTime          time.Time              `json:"start_time"`
	EndTime            *time.Time             `json:"end_time,omitempty"`
	BugDescription     string                 `json:"bug_description,omitempty"`
	Environment        map[string]string      `json:"environment,omitempty"`
	Requests           map[string]*Request    `json:"requests,omitempty"`
	Results            map[string]*Result     `json:"results,omitempty"`
	Hypotheses         []Hypothesis           `json:"hypotheses,omitempty"`
	CompletedHypotheses []HypothesisOutcome   `json:"completed_hypotheses,omitempty"` // Tested hypotheses in this session
	Experiments        []Experiment           `json:"experiments,omitempty"`
	CurrentStep        string                 `json:"current_step,omitempty"`     // Current step in workflow
	IterationCount     int                    `json:"iteration_count"`
	BugFixed           bool                   `json:"bug_fixed"`
	FixDescription     string                 `json:"fix_description,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	Phase              string                 `json:"phase,omitempty"` // discovery or fix
	DistilledKnowledge string                 `json:"distilled_knowledge,omitempty"` // Condensed findings from learn phase
}

// Request represents a single tool request in the session
type Request struct {
	ID           string            `json:"id"`
	ToolName     string            `json:"tool_name"`
	Timestamp    time.Time         `json:"timestamp"`
	Arguments    map[string]interface{} `json:"arguments"`
	Duration     int64             `json:"duration_ms"` // Milliseconds
	Success      bool              `json:"success"`
}

// Result represents the result of a tool execution
type Result struct {
	ID           string      `json:"id"`
	RequestID    string      `json:"request_id"`
	ToolName     string      `json:"tool_name"`
	Output       string      `json:"output"`
	StructuredData interface{} `json:"structured_data,omitempty"`
	Error        string      `json:"error,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
}

// Hypothesis represents a testable hypothesis in the scientific method
type Hypothesis struct {
	ID                  string    `json:"id"`
	CreatedAt           time.Time `json:"created_at"`
	BugObservation      string    `json:"bug_observation"`          // What was observed from the bug
	SuspectedComponent  string    `json:"suspected_component"`      // Where in the code/system
	RootCauseTheory     string    `json:"root_cause_theory"`        // Why you think it's broken
	EvidenceChain       string    `json:"evidence_chain"`           // How cause produces symptom
	FalsificationTest   string    `json:"falsification_test"`       // What would prove you wrong
	HypothesisText      string    `json:"hypothesis_text"`          // Synthesized statement
	ExpectedOutcome     string    `json:"expected_outcome,omitempty"`
	IsFalsifiable       bool      `json:"is_falsifiable"`
	ValidationNotes     string    `json:"validation_notes,omitempty"`
	Conclusion          string    `json:"conclusion,omitempty"`     // supported, refuted, inconclusive
	TestedAt            *time.Time `json:"tested_at,omitempty"`     // When hypothesis was tested
}

// HypothesisOutcome tracks what was learned from testing a hypothesis
type HypothesisOutcome struct {
	Hypothesis  Hypothesis `json:"hypothesis"`
	Conclusion  string     `json:"conclusion"`           // supported, refuted, inconclusive
	Findings    string     `json:"findings"`             // What we learned
	TestedAt    time.Time  `json:"tested_at"`
	IterationNum int       `json:"iteration_number"`
}

// SessionSummary provides a condensed view of a completed debugging session for reference
type SessionSummary struct {
	SessionID       string            `json:"session_id"`
	BugDescription  string            `json:"bug_description"`
	TestedHypotheses []HypothesisOutcome `json:"tested_hypotheses"`
	FinalConclusion string            `json:"final_conclusion"`      // supported, refuted, inconclusive, or fixed
	SimilarityScore float64           `json:"similarity_score,omitempty"` // 0-1, how similar to current bug
	CreatedAt       time.Time         `json:"created_at"`
	ResolvedAt      *time.Time        `json:"resolved_at,omitempty"`
}

// Experiment represents an experiment design in the scientific method
type Experiment struct {
	ID                string    `json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	HypothesisID      string    `json:"hypothesis_id"`
	Steps             []string  `json:"steps"`
	IndependentVars   []string  `json:"independent_vars"`
	DependentVars     []string  `json:"dependent_vars"`
	ControlVars       []string  `json:"control_vars"`
	Prediction        string    `json:"prediction,omitempty"`
	ActualObservations string   `json:"actual_observations,omitempty"`
	Conclusion        string    `json:"conclusion,omitempty"`
	HypothesisSupported *bool   `json:"hypothesis_supported,omitempty"` // nil = inconclusive
}

// ValidationError holds validation failure details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}
