package main

import "time"

// SessionContext holds all debugging session state
type SessionContext struct {
	ID              string                 `json:"id"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         *time.Time             `json:"end_time,omitempty"`
	BugDescription  string                 `json:"bug_description,omitempty"`
	Environment     map[string]string      `json:"environment,omitempty"`
	Requests        map[string]*Request    `json:"requests,omitempty"`
	Results         map[string]*Result     `json:"results,omitempty"`
	Hypotheses      []Hypothesis           `json:"hypotheses,omitempty"`
	Experiments     []Experiment           `json:"experiments,omitempty"`
	CurrentStep     string                 `json:"current_step,omitempty"`     // Current step in workflow
	IterationCount  int                    `json:"iteration_count"`
	BugFixed        bool                   `json:"bug_fixed"`
	FixDescription  string                 `json:"fix_description,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
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
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	HypothesisText  string    `json:"hypothesis_text"`
	ExpectedOutcome string    `json:"expected_outcome,omitempty"`
	IsFalsifiable   bool      `json:"is_falsifiable"`
	ValidationNotes string    `json:"validation_notes,omitempty"`
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
