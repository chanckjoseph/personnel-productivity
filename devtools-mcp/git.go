package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// handleGitStatus returns repository status: URL, current branch, and modified files
func handleGitStatus(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	workDir := os.Getenv("GIT_WORK_DIR")
	if workDir == "" {
		workDir = "."
	}

	var output strings.Builder

	// Get repo URL
	getRemoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	getRemoteCmd.Dir = workDir
	remoteOut, err := getRemoteCmd.Output()
	if err != nil {
		sendError(writer, id, -32602, "Failed to get remote URL. Is this a git repository?")
		return
	}
	repoURL := strings.TrimSpace(string(remoteOut))

	// Get current branch
	getBranchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	getBranchCmd.Dir = workDir
	branchOut, err := getBranchCmd.Output()
	if err != nil {
		sendError(writer, id, -32602, "Failed to get current branch")
		return
	}
	currentBranch := strings.TrimSpace(string(branchOut))

	// Get git status in porcelain format (easy to parse)
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = workDir
	statusOut, err := statusCmd.Output()
	if err != nil {
		sendError(writer, id, -32602, "Failed to get git status")
		return
	}

	// Parse status output
	var files []map[string]interface{}
	statusLines := strings.Split(strings.TrimSpace(string(statusOut)), "\n")

	modifiedCount := 0
	untrackedCount := 0

	for _, line := range statusLines {
		if line == "" {
			continue
		}

		// Format: "XY filename"
		if len(line) < 3 {
			continue
		}

		status := line[:2]
		filename := strings.TrimSpace(line[3:])

		// Determine file status
		var fileStatus string
		switch status {
		case "M ":
			fileStatus = "modified"
			modifiedCount++
		case " M":
			fileStatus = "modified (staged)"
			modifiedCount++
		case "A ":
			fileStatus = "added"
			modifiedCount++
		case "D ":
			fileStatus = "deleted"
			modifiedCount++
		case "??":
			fileStatus = "untracked"
			untrackedCount++
		case "MM", "AM", "DM":
			fileStatus = "conflict"
			modifiedCount++
		default:
			fileStatus = "unknown"
		}

		files = append(files, map[string]interface{}{
			"filename": filename,
			"status":   fileStatus,
		})
	}

	// Build summary
	output.WriteString(fmt.Sprintf("Repository: %s\n", repoURL))
	output.WriteString(fmt.Sprintf("Current branch: %s\n", currentBranch))
	output.WriteString(fmt.Sprintf("Modified files: %d, Untracked files: %d\n", modifiedCount, untrackedCount))

	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output.String(),
			},
		},
		"repo_url":       repoURL,
		"current_branch": currentBranch,
		"files":          files,
		"summary": map[string]interface{}{
			"modified":  modifiedCount,
			"untracked": untrackedCount,
			"total":     len(files),
		},
	})
}

// handleGitCommit commits all staged changes with a message
func handleGitCommit(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	// Validate required message
	messageRaw, ok := args["message"]
	if !ok {
		sendError(writer, id, -32602, "missing required argument: message")
		return
	}

	// Parse message
	message, ok := messageRaw.(string)
	if !ok {
		sendError(writer, id, -32602, "message must be a string")
		return
	}

	message = strings.TrimSpace(message)
	if message == "" {
		sendError(writer, id, -32602, "message cannot be empty")
		return
	}

	// Run git commands
	var output strings.Builder

	// git add -A
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = os.Getenv("GIT_WORK_DIR")
	if addCmd.Dir == "" {
		addCmd.Dir = "."
	}

	addOut, err := addCmd.CombinedOutput()
	if err != nil {
		output.WriteString(fmt.Sprintf("git add -A failed: %s\n%s", err, string(addOut)))
	} else {
		output.WriteString("Staged all changes with: git add -A\n")
		if len(addOut) > 0 {
			output.WriteString(string(addOut))
		}
	}

	// git commit -m <message>
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = os.Getenv("GIT_WORK_DIR")
	if commitCmd.Dir == "" {
		commitCmd.Dir = "."
	}

	commitOut, err := commitCmd.CombinedOutput()
	if err != nil {
		output.WriteString(fmt.Sprintf("git commit failed: %s\n%s", err, string(commitOut)))
		sendResult(writer, id, map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		})
		return
	}

	output.WriteString(fmt.Sprintf("Commit successful: %s\n", message))
	if len(commitOut) > 0 {
		output.WriteString(string(commitOut))
	}

	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output.String(),
			},
		},
	})
}

// handleGitPush pushes commits to GitHub using .pat and .username credentials
func handleGitPush(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	// Get working directory (where .pat and .username are)
	workDir := os.Getenv("GIT_WORK_DIR")
	if workDir == "" {
		workDir = "."
	}

	// Read .username
	usernamePath := filepath.Join(workDir, ".username")
	usernameBytes, err := os.ReadFile(usernamePath)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf(
			"Cannot read .username file. Create it at: %s\nContent should be your GitHub username (e.g., chanckjoseph)",
			usernamePath))
		return
	}
	username := strings.TrimSpace(string(usernameBytes))
	if username == "" {
		sendError(writer, id, -32602, fmt.Sprintf(
			".username file is empty. Add your GitHub username to: %s",
			usernamePath))
		return
	}

	// Read .pat
	patPath := filepath.Join(workDir, ".pat")
	patBytes, err := os.ReadFile(patPath)
	if err != nil {
		sendError(writer, id, -32602, fmt.Sprintf(
			"Cannot read .pat file. Create it at: %s\nContent should be your GitHub Personal Access Token (https://github.com/settings/tokens)",
			patPath))
		return
	}
	token := strings.TrimSpace(string(patBytes))
	if token == "" {
		sendError(writer, id, -32602, fmt.Sprintf(
			".pat file is empty. Add your GitHub Personal Access Token to: %s",
			patPath))
		return
	}

	// Get branch (default: main)
	branch := "main"
	if branchRaw, ok := args["branch"]; ok {
		if branchStr, ok := branchRaw.(string); ok {
			branch = strings.TrimSpace(branchStr)
		}
	}

	// Get remote origin URL from git config
	getRemoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	getRemoteCmd.Dir = workDir
	remoteOut, err := getRemoteCmd.Output()
	if err != nil {
		sendError(writer, id, -32602, "Cannot read git remote. Is this a git repository?")
		return
	}

	remoteURL := strings.TrimSpace(string(remoteOut))
	// Extract owner/repo from URL
	// Handle both https://github.com/owner/repo.git and git@github.com:owner/repo.git
	if !strings.Contains(remoteURL, "github.com") {
		sendError(writer, id, -32602, "Remote origin is not a GitHub repository")
		return
	}

	// Build authenticated URL
	var authenticatedURL string
	if strings.HasPrefix(remoteURL, "git@") {
		// SSH format: git@github.com:owner/repo.git -> https://github.com/owner/repo.git
		parts := strings.Split(remoteURL, ":")
		if len(parts) == 2 {
			authenticatedURL = fmt.Sprintf("https://%s:%s@github.com/%s", username, token, parts[1])
		} else {
			sendError(writer, id, -32602, "Invalid SSH remote format")
			return
		}
	} else {
		// HTTPS format: https://github.com/owner/repo.git
		// Replace with authenticated URL
		authenticatedURL = fmt.Sprintf("https://%s:%s@github.com/%s",
			username, token, strings.TrimPrefix(remoteURL, "https://github.com/"))
	}

	// Run git push
	pushCmd := exec.Command("git", "push", authenticatedURL, branch, "--set-upstream")
	pushCmd.Dir = workDir

	var output strings.Builder
	pushOut, err := pushCmd.CombinedOutput()
	if err != nil {
		output.WriteString(fmt.Sprintf("git push failed: %s\n%s", err, string(pushOut)))
		sendResult(writer, id, map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": output.String(),
				},
			},
		})
		return
	}

	output.WriteString(fmt.Sprintf("Push successful to branch: %s\n", branch))
	if len(pushOut) > 0 {
		output.WriteString(string(pushOut))
	}

	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": output.String(),
			},
		},
	})
}
