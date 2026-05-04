package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// PromptAnalyzer defines the interface for local prompt analysis.
type PromptAnalyzer interface {
	AnalyzePrompt(prompt, schemaVersion string) (string, error)
	GetModel() string
}

// LocalAnalyzer implements PromptAnalyzer using local parsing logic.
type LocalAnalyzer struct {
	model string
}

// NewLocalAnalyzer creates a new local analyzer (replaces external API calls).
func NewLocalAnalyzer(model string) *LocalAnalyzer {
	return &LocalAnalyzer{
		model: model,
	}
}

// AnalyzePrompt parses the prompt locally and returns a Query JSON string.
func (a *LocalAnalyzer) AnalyzePrompt(prompt, schemaVersion string) (string, error) {
	query := a.parsePrompt(prompt)
	
	output, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query: %w", err)
	}
	
	return string(output), nil
}

// parsePrompt analyzes the prompt and extracts intent, tasks, requirements, and constraints.
func (a *LocalAnalyzer) parsePrompt(prompt string) Query {
	lowerPrompt := strings.ToLower(prompt)
	
	intent := a.extractIntent(prompt, lowerPrompt)
	tasks := a.extractTasks(prompt, lowerPrompt)
	requirements := a.extractRequirements(prompt, lowerPrompt)
	constraints := a.extractConstraints(prompt, lowerPrompt)
	
	return Query{
		Intent:       intent,
		Tasks:        tasks,
		Requirements: requirements,
		Constraints:  constraints,
	}
}

// extractIntent identifies the primary goal, urgency, scope, and domain.
func (a *LocalAnalyzer) extractIntent(prompt, lowerPrompt string) Intent {
	// Detect urgency from keywords
	urgency := "medium"
	if strings.Contains(lowerPrompt, "urgent") || strings.Contains(lowerPrompt, "asap") || strings.Contains(lowerPrompt, "critical") {
		urgency = "high"
	} else if strings.Contains(lowerPrompt, "when you have time") || strings.Contains(lowerPrompt, "eventually") {
		urgency = "low"
	}
	
	// Detect domain from keywords
	domain := "software"
	if strings.Contains(lowerPrompt, "write") || strings.Contains(lowerPrompt, "document") || strings.Contains(lowerPrompt, "email") {
		domain = "writing"
	} else if strings.Contains(lowerPrompt, "research") || strings.Contains(lowerPrompt, "investigate") {
		domain = "research"
	} else if strings.Contains(lowerPrompt, "plan") || strings.Contains(lowerPrompt, "organize") {
		domain = "planning"
	}
	
	// Extract primary goal from the first sentence
	sentences := strings.Split(prompt, ".")
	primary := strings.TrimSpace(sentences[0])
	if len(primary) > 200 {
		primary = primary[:200]
	}
	
	return Intent{
		Primary: primary,
		Urgency: urgency,
		Scope:   "local",
		Domain:  domain,
	}
}

// extractTasks identifies action items and their dependencies.
func (a *LocalAnalyzer) extractTasks(prompt, lowerPrompt string) []Task {
	var tasks []Task
	taskID := 1
	
	// Look for action verbs
	actionVerbs := []string{"create", "build", "fix", "debug", "refactor", "analyze", "test", "review", "implement", "design", "document", "deploy", "setup", "configure", "optimize"}
	
	for _, verb := range actionVerbs {
		re := regexp.MustCompile(`(?i)(` + verb + `[^.!?]*)[.!?]`)
		matches := re.FindStringSubmatch(prompt)
		if len(matches) > 1 {
			taskTitle := strings.TrimSpace(matches[1])
			if len(taskTitle) > 100 {
				taskTitle = taskTitle[:100]
			}
			
			tasks = append(tasks, Task{
				ID:          fmt.Sprintf("task_%d", taskID),
				Title:       taskTitle,
				Description: taskTitle,
				Type:        "action",
				Parameters:  map[string]interface{}{},
			})
			taskID++
		}
	}
	
	// If no tasks found, create a generic one
	if len(tasks) == 0 {
		tasks = append(tasks, Task{
			ID:          "task_1",
			Title:       "Process request",
			Description: prompt,
			Type:        "action",
			Parameters:  map[string]interface{}{},
		})
	}
	
	return tasks
}

// extractRequirements identifies needed information.
func (a *LocalAnalyzer) extractRequirements(prompt, lowerPrompt string) InformationRequires {
	required := []InformationRequirement{}
	optional := []InformationRequirement{}
	
	// Look for file references
	if strings.Contains(lowerPrompt, "file") || strings.Contains(lowerPrompt, "path") {
		required = append(required, InformationRequirement{
			ID:       "req_1",
			Name:     "file_path",
			Type:     "file_path",
			Status:   "missing",
			Sources:  []Source{{Type: "user_input", Location: "user prompt", Confidence: "high"}},
		})
	}
	
	// Look for configuration references
	if strings.Contains(lowerPrompt, "config") || strings.Contains(lowerPrompt, "environment") {
		optional = append(optional, InformationRequirement{
			ID:       "req_2",
			Name:     "configuration",
			Type:     "text",
			Status:   "missing",
			Sources:  []Source{{Type: "environment", Location: "environment variables", Confidence: "high"}, {Type: "file_system", Location: "local files", Confidence: "high"}},
		})
	}
	
	return InformationRequires{
		Required: required,
		Optional: optional,
		Derived:  []InformationRequirement{},
	}
}

// extractConstraints identifies limitations or risks.
func (a *LocalAnalyzer) extractConstraints(prompt, lowerPrompt string) []Constraint {
	var constraints []Constraint
	
	// Check for permission-related constraints
	if strings.Contains(lowerPrompt, "permission") || strings.Contains(lowerPrompt, "access") {
		constraints = append(constraints, Constraint{
			Type:        "permission",
			Description: "May require specific permissions or credentials",
			Impact:      "medium",
		})
	}
	
	// Check for scope constraints
	if strings.Contains(lowerPrompt, "local") {
		constraints = append(constraints, Constraint{
			Type:        "scope",
			Description: "Limited to local workspace",
			Impact:      "low",
		})
	}
	
	return constraints
}

// GetModel returns the analyzer model name.
func (a *LocalAnalyzer) GetModel() string {
	return a.model
}


