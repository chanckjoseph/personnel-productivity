package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

// handleDebugWorkflow orchestrates the full 6-step scientific debugging workflow
func handleDebugWorkflow(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	// Validate session_id (optional if step is "start")
	var sessionID string
	if sessionIDRaw, ok := args["session_id"]; ok {
		if sid, ok := sessionIDRaw.(string); ok {
			sessionID = strings.TrimSpace(sid)
		}
	}

	// Parse step parameter
	stepRaw, ok := args["step"]
	if !ok {
		sendError(writer, id, -32602, "missing required argument: step")
		return
	}

	step, ok := stepRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "step must be a string")
		return
	}

	step = strings.TrimSpace(strings.ToLower(step))

	switch step {
	case "start":
		handleWorkflowStart(writer, id, args)
	case "learn":
		handleWorkflowLearn(writer, id, sessionID, args)
	case "hypothesis":
		handleWorkflowHypothesis(writer, id, sessionID, args)
	case "experiment":
		handleWorkflowExperiment(writer, id, sessionID, args)
	case "analyze":
		handleWorkflowAnalyze(writer, id, sessionID, args)
	case "fix":
		handleWorkflowFix(writer, id, sessionID, args)
	case "iterate":
		handleWorkflowIterate(writer, id, sessionID, args)
	default:
		sendError(writer, id, -32602, "step must be one of: start, learn, hypothesis, experiment, analyze, fix, iterate")
	}
}

// Step handlers for the workflow

func handleWorkflowStart(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	// Extract bug description
	bugDescription := ""
	if descRaw, ok := args["bug_description"]; ok {
		if desc, ok := descRaw.(string); ok {
			bugDescription = strings.TrimSpace(desc)
		}
	}

	if bugDescription == "" {
		sendError(writer, id, -32602, "bug_description required for start step")
		return
	}

	// Create session
	session, err := sessionMgr.CreateSession(bugDescription)
	if err != nil {
		sendError(writer, id, -32603, fmt.Sprintf("Failed to create session: %v", err))
		return
	}

	session.CurrentStep = "start"
	_ = sessionMgr.UpdateSession(session)

	var guidance strings.Builder
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n")
	guidance.WriteString("  DEBUGGING SESSION STARTED\n")
	guidance.WriteString("  Scientific Method: Hypothesis-Driven Root Cause Analysis\n")
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n\n")
	guidance.WriteString("Session ID: " + session.ID + "\n\n")
	guidance.WriteString("Bug Description (Your Observation):\n")
	guidance.WriteString(fmt.Sprintf("  %s\n\n", bugDescription))
	guidance.WriteString("Process Overview:\n")
	guidance.WriteString("  The scientific method will guide us through these steps:\n")
	guidance.WriteString("    1. Observation      → Define what's broken (DONE)\n")
	guidance.WriteString("    🔄 2. Context         → Gather codebase context and investigation\n")
	guidance.WriteString("    3. Hypothesis       → Propose a testable root cause\n")
	guidance.WriteString("    4. Experiment       → Design controlled tests\n")
	guidance.WriteString("    5. Analysis         → Evaluate results\n")
	guidance.WriteString("    6. Fix/Iterate      → Apply fix or test new hypothesis\n\n")
	guidance.WriteString("NEXT STEP: Gather Context\n")
	guidance.WriteString("  Investigate the codebase and gather context about the bug.\n")
	guidance.WriteString("  Call 'learn' step to start the investigation phase.\n")

	sendResult(writer, id, map[string]interface{}{
		"session_id": session.ID,
		"step":       "start",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": guidance.String(),
			},
		},
	})
}

func handleWorkflowLearn(writer *bufio.Writer, id interface{}, sessionID string, args map[string]interface{}) {
	if sessionID == "" {
		sendError(writer, id, -32602, "session_id required")
		return
	}

	session, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Session not found: %v", err))
		return
	}

	// Phase enforcement: learn is only allowed in discovery phase
	if session.Phase != "discovery" {
		sendError(writer, id, -32602, fmt.Sprintf("learn step only allowed in discovery phase (current: %s)", session.Phase))
		return
	}

	// Extract optional investigation focus
	investigationFocus := ""
	if focusRaw, ok := args["investigation_focus"]; ok {
		if focus, ok := focusRaw.(string); ok {
			investigationFocus = strings.TrimSpace(focus)
		}
	}

	// Extract optional distilled knowledge (from agent's exploration)
	distilledKnowledge := ""
	if distilledRaw, ok := args["distilled_knowledge"]; ok {
		if distilled, ok := distilledRaw.(string); ok {
			distilledKnowledge = strings.TrimSpace(distilled)
		}
	}

	// Build learning context including what we already know
	var learningContext strings.Builder

	// 1. Surface previous hypotheses tested in this session
	if len(session.CompletedHypotheses) > 0 {
		learningContext.WriteString("PREVIOUS HYPOTHESES TESTED IN THIS SESSION:\n")
		for i, outcome := range session.CompletedHypotheses {
			learningContext.WriteString(fmt.Sprintf("  %d. %s\n", i+1, outcome.Hypothesis.HypothesisText))
			learningContext.WriteString(fmt.Sprintf("     Result: %s\n", outcome.Conclusion))
			if outcome.Findings != "" {
				learningContext.WriteString(fmt.Sprintf("     Findings: %s\n", outcome.Findings))
			}
			learningContext.WriteString("\n")
		}
	}

	// 2. Store distilled knowledge for hypothesis step
	if distilledKnowledge != "" {
		learningContext.WriteString("CURRENT SESSION INVESTIGATION FINDINGS:\n")
		learningContext.WriteString(distilledKnowledge)
		learningContext.WriteString("\n\n")
		session.DistilledKnowledge = distilledKnowledge
	}

	session.CurrentStep = "learn"
	_ = sessionMgr.UpdateSession(session)

	var guidance strings.Builder
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n")
	guidance.WriteString("  STEP 2: CONTEXT GATHERING (DISCOVERY PHASE)\n")
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n\n")
	guidance.WriteString("Before formulating a hypothesis, gather context about the codebase.\n")
	guidance.WriteString("Investigate DIRECTLY using available tools. Do NOT repeat exploration.\n\n")

	// Show what we already know
	if learningContext.Len() > 0 {
		guidance.WriteString("WHAT WE ALREADY KNOW:\n")
		guidance.WriteString(learningContext.String())
		guidance.WriteString("\n")
	}

	guidance.WriteString("INVESTIGATION CHECKLIST:\n\n")
	guidance.WriteString("  📁 Project Structure\n")
	guidance.WriteString("     - Use 'project_structure' tool to explore the codebase\n")
	guidance.WriteString("     - Understand the architecture and module organization\n\n")

	guidance.WriteString("  🔍 Bug Context\n")
	if investigationFocus != "" {
		guidance.WriteString(fmt.Sprintf("     - Focus: %s\n\n", investigationFocus))
	} else {
		guidance.WriteString("     - Where does the bug occur? (component, module, file)\n")
		guidance.WriteString("     - What's the execution path?\n\n")
	}

	guidance.WriteString("  📊 Evidence to Gather\n")
	guidance.WriteString("     - Error messages or logs\n")
	guidance.WriteString("     - Stack traces\n")
	guidance.WriteString("     - Reproduction steps\n")
	guidance.WriteString("     - Related code files\n")
	guidance.WriteString("     - Configuration or environment details\n\n")

	guidance.WriteString("  🧩 Components to Identify\n")
	guidance.WriteString("     - Which modules/functions are involved?\n")
	guidance.WriteString("     - What's the data flow?\n")
	guidance.WriteString("     - Where's the most likely failure point?\n\n")

	guidance.WriteString("DISTILLED KNOWLEDGE FORMAT:\n")
	guidance.WriteString("When calling 'hypothesis' step, provide distilled_knowledge parameter:\n")
	guidance.WriteString("  • Suspected files/modules: [list specific file paths]\n")
	guidance.WriteString("  • Key patterns found: [describe error patterns or behavior]\n")
	guidance.WriteString("  • Error indicators: [key symptoms or error messages]\n")
	guidance.WriteString("  • Relevant context: [essential environment/config details]\n\n")

	guidance.WriteString("THEN provide the 5 required hypothesis components:\n")
	guidance.WriteString("  • bug_observation: What was observed\n")
	guidance.WriteString("  • suspected_component: Where you think the bug is\n")
	guidance.WriteString("  • root_cause_theory: Why that component is broken\n")
	guidance.WriteString("  • evidence_chain: How cause leads to symptom\n")
	guidance.WriteString("  • falsification_test: What would prove you wrong\n")

	sendResult(writer, id, map[string]interface{}{
		"session_id": session.ID,
		"step":       "learn",
		"context": map[string]interface{}{
			"what_we_know":     learningContext.String(),
			"what_to_explore":  investigationFocus,
			"session_history":  len(session.CompletedHypotheses),
		},
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": guidance.String(),
			},
		},
	})
}

func handleWorkflowHypothesis(writer *bufio.Writer, id interface{}, sessionID string, args map[string]interface{}) {
	if sessionID == "" {
		sendError(writer, id, -32602, "session_id required")
		return
	}

	session, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Session not found: %v", err))
		return
	}

	// Phase enforcement: hypothesis is only allowed in discovery phase
	if session.Phase != "discovery" {
		sendError(writer, id, -32602, fmt.Sprintf("hypothesis step only allowed in discovery phase (current: %s)", session.Phase))
		return
	}

	// Extract structured hypothesis components
	extractString := func(key string) string {
		if raw, ok := args[key]; ok {
			if s, ok := raw.(string); ok {
				return strings.TrimSpace(s)
			}
		}
		return ""
	}

	// Get distilled knowledge from learn phase if available
	distilledKnowledge := extractString("distilled_knowledge")

	bugObservation := extractString("bug_observation")
	suspectedComponent := extractString("suspected_component")
	rootCauseTheory := extractString("root_cause_theory")
	evidenceChain := extractString("evidence_chain")
	falsificationTest := extractString("falsification_test")

	// Validate all required components are provided
	missingFields := []string{}
	if bugObservation == "" {
		missingFields = append(missingFields, "bug_observation")
	}
	if suspectedComponent == "" {
		missingFields = append(missingFields, "suspected_component")
	}
	if rootCauseTheory == "" {
		missingFields = append(missingFields, "root_cause_theory")
	}
	if evidenceChain == "" {
		missingFields = append(missingFields, "evidence_chain")
	}
	if falsificationTest == "" {
		missingFields = append(missingFields, "falsification_test")
	}

	if len(missingFields) > 0 {
		var guidance strings.Builder
		guidance.WriteString("╔════════════════════════════════════════════════════════════╗\n")
		guidance.WriteString("║  STEP 3: HYPOTHESIS FORMULATION (INCOMPLETE)              ║\n")
		guidance.WriteString("╚════════════════════════════════════════════════════════════╝\n\n")
		guidance.WriteString("A good hypothesis requires FIVE structured components:\n\n")
		
		components := []map[string]string{
			{"name": "bug_observation", "display": "BUG OBSERVATION", "desc": "What exactly was observed from the bug?"},
			{"name": "suspected_component", "display": "SUSPECTED COMPONENT", "desc": "Which code/component do you suspect is involved?"},
			{"name": "root_cause_theory", "display": "ROOT CAUSE THEORY", "desc": "Why do you think that component is broken?"},
			{"name": "evidence_chain", "display": "EVIDENCE CHAIN", "desc": "How does the root cause produce the observed symptom?"},
			{"name": "falsification_test", "display": "FALSIFICATION TEST", "desc": "What evidence would prove you wrong?"},
		}

		for _, comp := range components {
			status := "✓"
			for _, missing := range missingFields {
				if missing == comp["name"] {
					status = "✗"
					break
				}
			}
			guidance.WriteString(fmt.Sprintf("%s %s\n", status, comp["display"]))
			guidance.WriteString(fmt.Sprintf("   %s\n\n", comp["desc"]))
		}

		guidance.WriteString("EXAMPLE:\n")
		guidance.WriteString("  bug_observation: \"API returns 504 timeout every 5 minutes\"\n")
		guidance.WriteString("  suspected_component: \"DbManager.GetConnection()\"\n")
		guidance.WriteString("  root_cause_theory: \"Connection pool exhaustion - defaults to 10 but concurrent requests spike to 50+\"\n")
		guidance.WriteString("  evidence_chain: \"Pool maxed → connections block → 30s timeout → 504 error\"\n")
		guidance.WriteString("  falsification_test: \"Increase pool size to 100; if timeout stops, hypothesis is correct\"\n\n")

		guidance.WriteString("RETRY with all five components filled in.\n")

		sendError(writer, id, -32602, fmt.Sprintf("Missing components: %s", strings.Join(missingFields, ", ")))
		return
	}

	// Synthesize hypothesis text from components
	hypothesisText := fmt.Sprintf(
		"The bug is caused by %s: %s. This produces the symptom (%s) through this mechanism: %s",
		suspectedComponent,
		rootCauseTheory,
		bugObservation,
		evidenceChain,
	)

	// Create hypothesis with all structured components
	hyp := Hypothesis{
		ID:                fmt.Sprintf("hyp_%d", time.Now().UnixNano()),
		CreatedAt:         time.Now(),
		BugObservation:    bugObservation,
		SuspectedComponent: suspectedComponent,
		RootCauseTheory:   rootCauseTheory,
		EvidenceChain:     evidenceChain,
		FalsificationTest: falsificationTest,
		HypothesisText:    hypothesisText,
		IsFalsifiable:     true,
	}

	session.Hypotheses = append(session.Hypotheses, hyp)
	session.CurrentStep = "hypothesis"
	_ = sessionMgr.UpdateSession(session)

	var guidance strings.Builder
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n")
	guidance.WriteString("  STEP 3: HYPOTHESIS RECORDED ✓\n")
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	guidance.WriteString("STRUCTURED HYPOTHESIS COMPONENTS:\n\n")
	guidance.WriteString(fmt.Sprintf("📍 Bug Observation:\n  %s\n\n", bugObservation))
	guidance.WriteString(fmt.Sprintf("🔍 Suspected Component:\n  %s\n\n", suspectedComponent))
	guidance.WriteString(fmt.Sprintf("💡 Root Cause Theory:\n  %s\n\n", rootCauseTheory))
	guidance.WriteString(fmt.Sprintf("🔗 Evidence Chain (Cause → Symptom):\n  %s\n\n", evidenceChain))
	guidance.WriteString(fmt.Sprintf("❌ Falsification Test (What Would Prove You Wrong):\n  %s\n\n", falsificationTest))

	guidance.WriteString("SYNTHESIZED HYPOTHESIS:\n")
	guidance.WriteString(fmt.Sprintf("  \"%s\"\n\n", hypothesisText))

	guidance.WriteString("EVALUATION:\n")
	guidance.WriteString("  ✓ This hypothesis IS testable - the falsification test shows how to disprove it\n")
	guidance.WriteString("  ✓ It explains the observed symptoms through a causal mechanism\n")
	guidance.WriteString("  ✓ It identifies a specific component, not vague speculation\n\n")

	guidance.WriteString("Next Step:\n")
	guidance.WriteString("  Call 'experiment' step to design a test based on the falsification_test.\n")

	sendResult(writer, id, map[string]interface{}{
		"step":              "hypothesis",
		"hypothesis_id":     hyp.ID,
		"hypothesis_pillar": map[string]interface{}{
			"hypothesis_id":      hyp.ID,
			"bug_observation":    bugObservation,
			"suspected_component": suspectedComponent,
			"root_cause_theory":  rootCauseTheory,
			"evidence_chain":     evidenceChain,
			"falsification_test": falsificationTest,
			"hypothesis_text":    hypothesisText,
		},
		"distilled_knowledge": distilledKnowledge,
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": guidance.String(),
			},
		},
	})
}

func handleWorkflowExperiment(writer *bufio.Writer, id interface{}, sessionID string, args map[string]interface{}) {
	if sessionID == "" {
		sendError(writer, id, -32602, "session_id required")
		return
	}

	session, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Session not found: %v", err))
		return
	}

	// Phase enforcement: experiment is only allowed in discovery phase
	if session.Phase != "discovery" {
		sendError(writer, id, -32602, fmt.Sprintf("experiment step only allowed in discovery phase (current: %s)", session.Phase))
		return
	}

	if len(session.Hypotheses) == 0 {
		sendError(writer, id, -32602, "no hypothesis found")
		return
	}

	// Extract experiment details
	var steps []string
	if stepsRaw, ok := args["steps"]; ok {
		steps = parseStringArray(stepsRaw)
	}

	if len(steps) == 0 {
		sendError(writer, id, -32602, "steps required for experiment step")
		return
	}

	// Parse variables
	independentVars := parseStringArray(args["independent_vars"])
	dependentVars := parseStringArray(args["dependent_vars"])
	controlVars := parseStringArray(args["control_vars"])

	// Check if user is requesting confirmation vs execution
	confirmedRaw, confirmed := args["confirmed"]
	isConfirmed := false
	if confirmed {
		if confBool, ok := confirmedRaw.(bool); ok {
			isConfirmed = confBool
		}
	}

	// If not yet confirmed, show design and ask for approval
	if !isConfirmed {
		var guidance strings.Builder
		guidance.WriteString("╔════════════════════════════════════════════════════════════╗\n")
		guidance.WriteString("║  STEP 4: EXPERIMENT DESIGN (AWAITING CONFIRMATION)        ║\n")
		guidance.WriteString("╚════════════════════════════════════════════════════════════╝\n\n")

		guidance.WriteString("⚠️  SAFETY REQUIREMENTS:\n")
		guidance.WriteString("  □ Experiment MUST use sandboxed environment (console)\n")
		guidance.WriteString("  □ Experiment MUST use MOCK DATA ONLY\n")
		guidance.WriteString("  □ NO access to production database\n")
		guidance.WriteString("  □ NO destructive operations on real systems\n")
		guidance.WriteString("  □ Results reproducible in isolated environment\n\n")

		guidance.WriteString("EXPERIMENT PLAN:\n")
		guidance.WriteString(fmt.Sprintf("  Steps to execute: %d\n\n", len(steps)))
		guidance.WriteString("  Execution Steps:\n")
		for i, step := range steps {
			guidance.WriteString(fmt.Sprintf("    %d. %s\n", i+1, step))
		}

		if len(controlVars) > 0 {
			guidance.WriteString(fmt.Sprintf("\n  Control Variables (held constant):\n"))
			for _, cv := range controlVars {
				guidance.WriteString(fmt.Sprintf("    • %s\n", cv))
			}
		}
		if len(independentVars) > 0 {
			guidance.WriteString(fmt.Sprintf("\n  Independent Variables (what we change):\n"))
			for _, iv := range independentVars {
				guidance.WriteString(fmt.Sprintf("    • %s\n", iv))
			}
		}
		if len(dependentVars) > 0 {
			guidance.WriteString(fmt.Sprintf("\n  Dependent Variables (what we measure):\n"))
			for _, dv := range dependentVars {
				guidance.WriteString(fmt.Sprintf("    • %s\n", dv))
			}
		}

		guidance.WriteString("\n⏸️  USER CONFIRMATION REQUIRED:\n")
		guidance.WriteString("  Review the experiment plan above carefully.\n")
		guidance.WriteString("  Verify it meets ALL safety requirements.\n\n")
		guidance.WriteString("  THEN call this tool again with confirmed=true parameter:\n")
		guidance.WriteString("  Example: experiment confirmed=true\n")

		sendResult(writer, id, map[string]interface{}{
			"step":           "experiment",
			"design_status":  "awaiting_confirmation",
			"experiment_id":  fmt.Sprintf("exp_%d", time.Now().UnixNano()),
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": guidance.String(),
				},
			},
		})
		return
	}

	// User confirmed - proceed with experiment creation
	exp := Experiment{
		ID:              fmt.Sprintf("exp_%d", time.Now().UnixNano()),
		CreatedAt:       time.Now(),
		HypothesisID:    session.Hypotheses[len(session.Hypotheses)-1].ID,
		Steps:           steps,
		IndependentVars: independentVars,
		DependentVars:   dependentVars,
		ControlVars:     controlVars,
	}

	session.Experiments = append(session.Experiments, exp)
	session.CurrentStep = "experiment"
	_ = sessionMgr.UpdateSession(session)

	var guidance strings.Builder
	guidance.WriteString("╔════════════════════════════════════════════════════════════╗\n")
	guidance.WriteString("║  STEP 4: EXPERIMENT CONFIRMED & READY TO EXECUTE          ║\n")
	guidance.WriteString("╚════════════════════════════════════════════════════════════╝\n\n")

	guidance.WriteString("✓ Safety Requirements ACKNOWLEDGED:\n")
	guidance.WriteString("  ✓ Using sandboxed/console environment\n")
	guidance.WriteString("  ✓ Using mock data only\n")
	guidance.WriteString("  ✓ No production database access\n")
	guidance.WriteString("  ✓ No destructive operations\n\n")

	guidance.WriteString("APPROVED EXPERIMENT PLAN:\n")
	guidance.WriteString(fmt.Sprintf("  Experiment ID: %s\n", exp.ID))
	guidance.WriteString(fmt.Sprintf("  Total steps: %d\n\n", len(steps)))

	guidance.WriteString("  Execute these steps:\n")
	for i, step := range steps {
		guidance.WriteString(fmt.Sprintf("    %d. %s\n", i+1, step))
	}

	guidance.WriteString("\n📊 VARIABLE TRACKING:\n")
	if len(controlVars) > 0 {
		guidance.WriteString("  Control (constant):\n")
		for _, cv := range controlVars {
			guidance.WriteString(fmt.Sprintf("    • %s\n", cv))
		}
	}
	if len(independentVars) > 0 {
		guidance.WriteString("  Independent (changed):\n")
		for _, iv := range independentVars {
			guidance.WriteString(fmt.Sprintf("    • %s\n", iv))
		}
	}
	if len(dependentVars) > 0 {
		guidance.WriteString("  Dependent (measured):\n")
		for _, dv := range dependentVars {
			guidance.WriteString(fmt.Sprintf("    • %s\n", dv))
		}
	}

	guidance.WriteString("\n➡️  NEXT STEPS:\n")
	guidance.WriteString("  1. Execute the experiment steps in a sandboxed environment\n")
	guidance.WriteString("  2. Capture ALL observations and measurements\n")
	guidance.WriteString("  3. Call 'analyze' step with detailed observations\n")
	guidance.WriteString("  4. Use 'debug_session_history' to view session progress\n")

	sendResult(writer, id, map[string]interface{}{
		"step":           "experiment",
		"experiment_id":  exp.ID,
		"design_status":  "confirmed",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": guidance.String(),
			},
		},
	})
}

func handleWorkflowAnalyze(writer *bufio.Writer, id interface{}, sessionID string, args map[string]interface{}) {
	if sessionID == "" {
		sendError(writer, id, -32602, "session_id required")
		return
	}

	session, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Session not found: %v", err))
		return
	}

	// Phase enforcement: analyze is only allowed in discovery phase
	if session.Phase != "discovery" {
		sendError(writer, id, -32602, fmt.Sprintf("analyze step only allowed in discovery phase (current: %s)", session.Phase))
		return
	}

	if len(session.Experiments) == 0 {
		sendError(writer, id, -32602, "no experiment found - use experiment step first")
		return
	}

	// Extract observations and conclusion
	observationsRaw, ok := args["observations"]
	if !ok {
		sendError(writer, id, -32602, "observations required")
		return
	}

	observations, ok := observationsRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "observations must be a string")
		return
	}

	observations = strings.TrimSpace(observations)
	if observations == "" {
		sendError(writer, id, -32602, "observations cannot be empty")
		return
	}

	conclusionRaw, ok := args["conclusion"]
	if !ok {
		sendError(writer, id, -32602, "conclusion required (supported|refuted|inconclusive)")
		return
	}

	conclusion, ok := conclusionRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "conclusion must be a string")
		return
	}

	conclusion = strings.TrimSpace(strings.ToLower(conclusion))
	if conclusion != "supported" && conclusion != "refuted" && conclusion != "inconclusive" {
		sendError(writer, id, -32602, "conclusion must be: supported, refuted, or inconclusive")
		return
	}

	// Update last experiment
	lastExp := &session.Experiments[len(session.Experiments)-1]
	lastExp.ActualObservations = observations

	if conclusion == "supported" {
		supported := true
		lastExp.HypothesisSupported = &supported
	} else if conclusion == "refuted" {
		supported := false
		lastExp.HypothesisSupported = &supported
	}

	session.CurrentStep = "analyze"
	_ = sessionMgr.UpdateSession(session)

	var guidance strings.Builder
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n")
	guidance.WriteString("  STEP 5: DATA ANALYSIS & CONCLUSION\n")
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n\n")
	guidance.WriteString("Observations Recorded:\n")
	guidance.WriteString(fmt.Sprintf("  %s\n\n", observations))
	guidance.WriteString(fmt.Sprintf("Conclusion: %s\n\n", strings.ToUpper(conclusion)))

	if conclusion == "supported" {
		guidance.WriteString("Analysis:\n")
		guidance.WriteString("  ✓ The experimental results SUPPORT your hypothesis.\n")
		guidance.WriteString("  The evidence demonstrates that your proposed root cause is likely correct.\n\n")
		guidance.WriteString("PHASE TRANSITION - DISCOVERY → FIX:\n")
		guidance.WriteString("  You are now ready to move to the Fix Phase.\n")
		guidance.WriteString("  The next 'fix' step will transition your session to the Fix Phase.\n\n")
		guidance.WriteString("Next Step:\n")
		guidance.WriteString("  Call 'fix' step with your proposed fix description.\n")
	} else if conclusion == "refuted" {
		guidance.WriteString("Analysis:\n")
		guidance.WriteString("  ✗ The experimental results REFUTE your hypothesis.\n")
		guidance.WriteString("  Your proposed root cause is NOT supported by the evidence.\n\n")
		guidance.WriteString("Next Step:\n")
		guidance.WriteString("  Call 'iterate' step to formulate a new hypothesis.\n")
	} else {
		guidance.WriteString("Analysis:\n")
		guidance.WriteString("  ? The results were INCONCLUSIVE.\n")
		guidance.WriteString("  We cannot definitively prove or disprove the hypothesis from this data.\n\n")
		guidance.WriteString("Next Step:\n")
		guidance.WriteString("  Call 'iterate' step to improve the experiment design.\n")
	}

	sendResult(writer, id, map[string]interface{}{
		"step":       "analyze",
		"conclusion": conclusion,
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": guidance.String(),
			},
		},
	})
}

func handleWorkflowFix(writer *bufio.Writer, id interface{}, sessionID string, args map[string]interface{}) {
	if sessionID == "" {
		sendError(writer, id, -32602, "session_id required")
		return
	}

	session, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Session not found: %v", err))
		return
	}

	// Phase enforcement: fix is allowed from discovery (first transition) or fix (already in phase)
	if session.Phase != "discovery" && session.Phase != "fix" {
		sendError(writer, id, -32602, fmt.Sprintf("fix step only allowed after discovery phase analysis or in fix phase (current: %s)", session.Phase))
		return
	}

	// Extract fix description
	fixRaw, ok := args["fix_description"]
	if !ok {
		sendError(writer, id, -32602, "fix_description required")
		return
	}

	fixDescription, ok := fixRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "fix_description must be a string")
		return
	}

	fixDescription = strings.TrimSpace(fixDescription)
	if fixDescription == "" {
		sendError(writer, id, -32602, "fix_description cannot be empty")
		return
	}

	// Transition from discovery to fix phase on first fix call
	if session.Phase == "discovery" {
		session.Phase = "fix"
	}

	session.BugFixed = true
	session.FixDescription = fixDescription
	session.CurrentStep = "fix"
	endTime := time.Now()
	session.EndTime = &endTime
	_ = sessionMgr.UpdateSession(session)

	var guidance strings.Builder
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n")
	guidance.WriteString("  STEP 6: BUG FIX DOCUMENTED\n")
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n\n")
	guidance.WriteString("Fix Description:\n")
	guidance.WriteString(fmt.Sprintf("  %s\n\n", fixDescription))
	guidance.WriteString("Implementation Checklist:\n")
	guidance.WriteString("  ☐ Code changes have been applied\n")
	guidance.WriteString("  ☐ Tested with the original failure case\n")
	guidance.WriteString("  ☐ No regressions introduced\n")
	guidance.WriteString("  ☐ Code review completed\n\n")
	guidance.WriteString("READY TO COMMIT:\n")
	guidance.WriteString("  You can now:\n")
	guidance.WriteString("  • Commit the fix using git_commit tool\n")
	guidance.WriteString("  • Push changes using git_push tool\n")
	guidance.WriteString("  • View session summary using debug_session_history tool\n")

	sendResult(writer, id, map[string]interface{}{
		"step":         "fix",
		"bug_fixed":    true,
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": guidance.String(),
			},
		},
	})
}

func handleWorkflowIterate(writer *bufio.Writer, id interface{}, sessionID string, args map[string]interface{}) {
	if sessionID == "" {
		sendError(writer, id, -32602, "session_id required")
		return
	}

	session, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Session not found: %v", err))
		return
	}

	// Phase enforcement: iterate is only allowed in discovery phase
	if session.Phase != "discovery" {
		sendError(writer, id, -32602, fmt.Sprintf("iterate step only allowed in discovery phase (current: %s)", session.Phase))
		return
	}

	session.IterationCount++
	session.CurrentStep = "hypothesis"
	_ = sessionMgr.UpdateSession(session)

	var guidance strings.Builder
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n")
	guidance.WriteString(fmt.Sprintf("  ITERATION #%d\n", session.IterationCount))
	guidance.WriteString("═══════════════════════════════════════════════════════════════\n\n")
	guidance.WriteString("Previous Results Summary:\n")
	if len(session.Hypotheses) > 0 {
		lastHyp := session.Hypotheses[len(session.Hypotheses)-1]
		guidance.WriteString(fmt.Sprintf("  Hypothesis: %s\n", lastHyp.HypothesisText))
	}
	if len(session.Experiments) > 0 {
		lastExp := session.Experiments[len(session.Experiments)-1]
		if lastExp.HypothesisSupported != nil {
			if *lastExp.HypothesisSupported {
				guidance.WriteString("  Result: SUPPORTED\n")
			} else {
				guidance.WriteString("  Result: REFUTED\n")
			}
		}
	}
	guidance.WriteString("\nAnalysis:\n")
	guidance.WriteString("  The previous hypothesis did not lead to a solution.\n")
	guidance.WriteString("  Now we'll formulate an improved hypothesis based on what we learned.\n\n")
	guidance.WriteString("Guidance for Next Hypothesis:\n")
	guidance.WriteString("  • What contradicted the previous hypothesis?\n")
	guidance.WriteString("  • What additional observations can you make?\n")
	guidance.WriteString("  • What alternative root causes exist?\n\n")
	guidance.WriteString(fmt.Sprintf("Next: Formulate Hypothesis #%d\n", session.IterationCount+1))
	guidance.WriteString("  Use 'hypothesis' step with your new testable statement.\n")

	sendResult(writer, id, map[string]interface{}{
		"step":      "iterate",
		"iteration": session.IterationCount,
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": guidance.String(),
			},
		},
	})
}

// parseStringArray converts interface{} to []string
func parseStringArray(arr interface{}) []string {
	var result []string
	if arrSlice, ok := arr.([]interface{}); ok {
		for _, item := range arrSlice {
			if str, ok := item.(string); ok {
				result = append(result, strings.TrimSpace(str))
			}
		}
	}
	return result
}

// Stub handlers for other debug tools
func handleStartDebugSession(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	// Extract bug description (optional)
	bugDescription := ""
	if descRaw, ok := args["bug_description"]; ok {
		if desc, ok := descRaw.(string); ok {
			bugDescription = strings.TrimSpace(desc)
		}
	}

	// Create new session
	session, err := sessionMgr.CreateSession(bugDescription)
	if err != nil {
		sendError(writer, id, -32603, fmt.Sprintf("Failed to create session: %v", err))
		return
	}

	sendResult(writer, id, map[string]interface{}{
		"session_id": session.ID,
		"timestamp":  session.StartTime,
	})
}

func handleGetSessionState(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	sessionIDRaw, ok := args["session_id"]
	if !ok {
		sendError(writer, id, -32602, "missing required argument: session_id")
		return
	}

	sessionID, ok := sessionIDRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "session_id must be a string")
		return
	}

	session, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Session not found: %v", err))
		return
	}

	sendResult(writer, id, map[string]interface{}{
		"session": map[string]interface{}{
			"id":               session.ID,
			"start_time":       session.StartTime,
			"bug_description":  session.BugDescription,
			"hypothesis_count": len(session.Hypotheses),
			"experiment_count": len(session.Experiments),
			"bug_fixed":        session.BugFixed,
			"iteration_count":  session.IterationCount,
			"current_step":     session.CurrentStep,
		},
	})
}

func handleUpdateSessionContext(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	sessionIDRaw, ok := args["session_id"]
	if !ok {
		sendError(writer, id, -32602, "missing required argument: session_id")
		return
	}

	sessionID, ok := sessionIDRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "session_id must be a string")
		return
	}

	keyRaw, ok := args["key"]
	if !ok {
		sendError(writer, id, -32602, "missing required argument: key")
		return
	}

	key, ok := keyRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "key must be a string")
		return
	}

	value, ok := args["value"]
	if !ok {
		sendError(writer, id, -32602, "missing required argument: value")
		return
	}

	session, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Session not found: %v", err))
		return
	}

	if session.Metadata == nil {
		session.Metadata = make(map[string]interface{})
	}
	session.Metadata[key] = value

	if err := sessionMgr.UpdateSession(session); err != nil {
		sendError(writer, id, -32603, fmt.Sprintf("Failed to update session: %v", err))
		return
	}

	sendResult(writer, id, map[string]interface{}{
		"key":   key,
		"value": value,
	})
}

func handleFormulateHypothesis(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	sendError(writer, id, -32602, "Use debug_workflow tool instead")
}

func handleDesignExperiment(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	sendError(writer, id, -32602, "Use debug_workflow tool instead")
}

func handleAnalyzeResults(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	sendError(writer, id, -32602, "Use debug_workflow tool instead")
}

func handleTrackIteration(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	sessionIDRaw, ok := args["session_id"]
	if !ok {
		sendError(writer, id, -32602, "missing required argument: session_id")
		return
	}

	sessionID, ok := sessionIDRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "session_id must be a string")
		return
	}

	session, err := sessionMgr.GetSession(sessionID)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf("Session not found: %v", err))
		return
	}

	sendResult(writer, id, map[string]interface{}{
		"session_id":      session.ID,
		"hypotheses":      len(session.Hypotheses),
		"experiments":     len(session.Experiments),
		"iterations":      session.IterationCount,
		"bug_fixed":       session.BugFixed,
	})
}
