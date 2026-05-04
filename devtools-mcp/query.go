package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
)

// QueryEngine orchestrates semantic transformation of free-form prompts.
type QueryEngine struct {
	analyzer      PromptAnalyzer
	artifactStore *ArtifactStore
	schemaVersion string
}

// NewQueryEngine creates a new Query engine with a prompt analyzer and artifact store.
func NewQueryEngine(analyzer PromptAnalyzer, artifactStore *ArtifactStore) *QueryEngine {
	return &QueryEngine{
		analyzer:      analyzer,
		artifactStore: artifactStore,
		schemaVersion: "1.0",
	}
}

// Parse transforms a free-form prompt into a structured Query using LLM analysis.
func (qe *QueryEngine) Parse(prompt string) (*Query, error) {
	startTime := time.Now()

	// Step 1: Call analyzer to process prompt and generate Query structure
	analysisOutput, err := qe.analyzer.AnalyzePrompt(prompt, qe.schemaVersion)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Step 2: Unmarshal analysis output into Query struct
	var query Query
	err = json.Unmarshal([]byte(analysisOutput), &query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM output: %w", err)
	}

	// Step 3: Validate Query conforms to schema
	if err := qe.validateQuery(&query); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	// Step 4: Post-process: analyze missing info chains
	qe.analyzeMissingInfoChains(&query)

	// Step 5: Calculate confidence and metadata
	executionTime := time.Since(startTime).Milliseconds()
	qe.populateMetadata(&query, prompt, analysisOutput, executionTime)

	return &query, nil
}

// validateQuery checks that the Query structure is valid and complete.
func (qe *QueryEngine) validateQuery(q *Query) error {
	// Check required fields
	if q.Intent.Primary == "" {
		return fmt.Errorf("intent.primary is required")
	}

	if len(q.Tasks) == 0 {
		return fmt.Errorf("at least one task is required")
	}

	// Validate each task
	for i, task := range q.Tasks {
		if task.Title == "" {
			return fmt.Errorf("task[%d].title is required", i)
		}
		if task.Type == "" {
			return fmt.Errorf("task[%d].type is required", i)
		}

		// Validate task type
		validTypes := map[string]bool{
			"action":   true,
			"decision": true,
			"review":   true,
			"output":   true,
		}
		if !validTypes[task.Type] {
			return fmt.Errorf("task[%d].type must be one of: action, decision, review, output", i)
		}

		// Recursively validate subtasks
		if len(task.Subtasks) > 0 {
			for j, subtask := range task.Subtasks {
				if subtask.Title == "" {
					return fmt.Errorf("task[%d].subtask[%d].title is required", i, j)
				}
			}
		}
	}

	// Validate information requirements
	allReqs := append(q.Requirements.Required, q.Requirements.Optional...)
	allReqs = append(allReqs, q.Requirements.Derived...)

	for i, req := range allReqs {
		if req.Name == "" {
			return fmt.Errorf("information_requirement[%d].name is required", i)
		}
		if req.Status == "" {
			return fmt.Errorf("information_requirement[%d].status is required", i)
		}

		// Validate status
		validStatus := map[string]bool{
			"available": true,
			"missing":   true,
			"unknown":   true,
			"error":     true,
		}
		if !validStatus[req.Status] {
			return fmt.Errorf("information_requirement[%d].status must be one of: available, missing, unknown, error", i)
		}
	}

	return nil
}

// analyzeMissingInfoChains builds recursive dependency chains for missing information.
func (qe *QueryEngine) analyzeMissingInfoChains(q *Query) {
	allReqs := append(q.Requirements.Required, q.Requirements.Optional...)
	allReqs = append(allReqs, q.Requirements.Derived...)

	for i, req := range allReqs {
		if req.Status == "missing" {
			chain := qe.buildMissingChain(&req, allReqs, 0)
			q.Requirements.Required[i].MissingChain = chain
		}
	}
}

// buildMissingChain recursively identifies dependencies for missing information.
func (qe *QueryEngine) buildMissingChain(req *InformationRequirement, allReqs []InformationRequirement, depth int) *MissingInfoChain {
	if depth > 5 { // Prevent infinite recursion
		return nil
	}

	chain := &MissingInfoChain{
		MissingID:   req.ID,
		MissingName: req.Name,
		Depth:       depth,
	}

	// Identify what's needed to get this missing requirement
	var dependsOn []InformationRequirement
	for _, source := range req.Sources {
		// If source requires other info, mark those as dependencies
		switch source.Type {
		case "inference":
			// Inference might depend on other available data
			// For now, mark as resolvable
			chain.Resolution = fmt.Sprintf("Can be inferred from: %s", source.Location)
		case "user_input":
			chain.Resolution = fmt.Sprintf("Needs user input: %s", source.Location)
		case "file_system", "git_status", "environment":
			chain.Resolution = fmt.Sprintf("Can be retrieved from: %s", source.Location)
		case "external_api":
			chain.Resolution = fmt.Sprintf("Needs external API call: %s", source.Location)
		}
	}

	if len(dependsOn) > 0 {
		chain.DependsOn = dependsOn
	}

	return chain
}

// populateMetadata fills in execution metadata for the Query.
func (qe *QueryEngine) populateMetadata(q *Query, originalPrompt, llmOutput string, executionTime int64) {
	q.ID = generateQueryID()
	q.OriginalText = originalPrompt

	q.Metadata.QueryID = q.ID
	q.Metadata.CreatedAt = time.Now()
	q.Metadata.AnalyzerModel = qe.analyzer.GetModel()
	q.Metadata.ExecutionTime = executionTime
	q.Metadata.SchemaVersion = qe.schemaVersion
	q.Metadata.Status = "success"

	// Count tasks and requirements
	q.Metadata.TaskCount = countTasks(q.Tasks)
	q.Metadata.RequirementCount = len(q.Requirements.Required) + len(q.Requirements.Optional) + len(q.Requirements.Derived)

	// Calculate confidence based on schema completeness and clarity
	q.Metadata.Confidence = calculateConfidence(q)
}

// countTasks recursively counts total tasks including subtasks.
func countTasks(tasks []Task) int {
	count := len(tasks)
	for _, task := range tasks {
		count += countTasks(task.Subtasks)
	}
	return count
}

// calculateConfidence computes confidence score (0.0 to 1.0) based on query characteristics.
func calculateConfidence(q *Query) float64 {
	score := 0.8 // Start with base confidence

	// Reduce confidence if intent has ambiguities
	if len(q.Intent.Ambiguities) > 0 {
		score -= float64(len(q.Intent.Ambiguities)) * 0.05
	}

	// Reduce confidence if many requirements are missing
	total := len(q.Requirements.Required) + len(q.Requirements.Optional) + len(q.Requirements.Derived)
	if total > 0 {
		missing := 0
		for _, req := range q.Requirements.Required {
			if req.Status == "missing" {
				missing++
			}
		}
		missingRatio := float64(missing) / float64(total)
		score -= missingRatio * 0.3
	}

	// Reduce confidence if no description for tasks
	emptyDescCount := 0
	for _, task := range q.Tasks {
		if task.Description == "" {
			emptyDescCount++
		}
	}
	if len(q.Tasks) > 0 {
		score -= (float64(emptyDescCount) / float64(len(q.Tasks))) * 0.15
	}

	// Clamp between 0.0 and 1.0
	return math.Max(0.0, math.Min(1.0, score))
}

// generateQueryID creates a timestamp-based unique ID for the query.
func generateQueryID() string {
	now := time.Now()
	return now.Format("2006-01-02T15-04-05-000")
}

// FormatJSON returns the Query as formatted JSON.
func (q *Query) FormatJSON() (string, error) {
	data, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// LogQuery outputs Query to log (useful for debugging).
func (q *Query) LogQuery() {
	formatted, _ := q.FormatJSON()
	log.Printf("Query Analysis:\n%s\n", formatted)
}

// ValidateTaskDependencies checks that prerequisite task IDs exist.
func (qe *QueryEngine) ValidateTaskDependencies(q *Query) error {
	allTaskIDs := extractTaskIDs(q.Tasks)

	for _, task := range q.Tasks {
		for _, prereqID := range task.Prerequisites {
			if !stringInSlice(allTaskIDs, prereqID) {
				return fmt.Errorf("task %s has unknown prerequisite: %s", task.ID, prereqID)
			}
		}
	}

	return nil
}

// extractTaskIDs flattens all task IDs from hierarchical structure.
func extractTaskIDs(tasks []Task) []string {
	var ids []string
	for _, task := range tasks {
		ids = append(ids, task.ID)
		ids = append(ids, extractTaskIDs(task.Subtasks)...)
	}
	return ids
}

// stringInSlice checks if string is in string slice.
func stringInSlice(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// GetTasksFlat returns all tasks (and subtasks) as a flat list.
func (q *Query) GetTasksFlat() []Task {
	var flat []Task
	flattenTasks(q.Tasks, &flat)
	return flat
}

// flattenTasks recursively flattens hierarchical tasks.
func flattenTasks(tasks []Task, flat *[]Task) {
	for _, task := range tasks {
		*flat = append(*flat, task)
		flattenTasks(task.Subtasks, flat)
	}
}

// SummarizeQuery produces a human-readable text summary.
func (q *Query) SummarizeQuery() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("Query ID: %s\n", q.ID))
	summary.WriteString(fmt.Sprintf("Prompt: %s\n\n", q.OriginalText))

	summary.WriteString(fmt.Sprintf("Intent: %s\n", q.Intent.Primary))
	if len(q.Intent.Secondary) > 0 {
		summary.WriteString(fmt.Sprintf("  Secondary: %v\n", q.Intent.Secondary))
	}
	summary.WriteString(fmt.Sprintf("  Domain: %s | Urgency: %s\n\n", q.Intent.Domain, q.Intent.Urgency))

	summary.WriteString(fmt.Sprintf("Tasks (%d):\n", len(q.Tasks)))
	for i, task := range q.Tasks {
		summary.WriteString(fmt.Sprintf("  %d. %s (%s)\n", i+1, task.Title, task.Type))
		if len(task.Prerequisites) > 0 {
			summary.WriteString(fmt.Sprintf("     Requires: %v\n", task.Prerequisites))
		}
	}

	missingCount := 0
	for _, req := range q.Requirements.Required {
		if req.Status == "missing" {
			missingCount++
		}
	}
	summary.WriteString(fmt.Sprintf("\nInformation Requirements:\n"))
	summary.WriteString(fmt.Sprintf("  Available: %d\n", len(q.Requirements.Required)-missingCount))
	summary.WriteString(fmt.Sprintf("  Missing: %d\n", missingCount))
	summary.WriteString(fmt.Sprintf("  Optional: %d\n", len(q.Requirements.Optional)))

	summary.WriteString(fmt.Sprintf("\nMetadata:\n"))
	summary.WriteString(fmt.Sprintf("  Confidence: %.1f%%\n", q.Metadata.Confidence*100))
	summary.WriteString(fmt.Sprintf("  Analyzer Model: %s\n", q.Metadata.AnalyzerModel))
	summary.WriteString(fmt.Sprintf("  Execution Time: %dms\n", q.Metadata.ExecutionTime))

	return summary.String()
}
