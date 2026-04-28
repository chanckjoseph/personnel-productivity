package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// readFrame reads a single JSON-RPC frame (line-delimited)
func readFrame(reader *bufio.Reader) ([]byte, error) {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(line) > 0 {
				return []byte(line), nil
			}
			return nil, err
		}

		line = strings.TrimSuffix(line, "\n")
		if line != "" {
			return []byte(line), nil
		}
	}
}

// parseID parses a JSON-RPC request ID (can be string or number)
func parseID(raw json.RawMessage) (interface{}, error) {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString, nil
	}

	var asNumber float64
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return asNumber, nil
	}

	return nil, fmt.Errorf("unsupported id type")
}

// sendResult sends a successful JSON-RPC response
func sendResult(writer *bufio.Writer, id interface{}, result interface{}) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	send(writer, resp)
}

// sendError sends a JSON-RPC error response
func sendError(writer *bufio.Writer, id interface{}, code int, message string) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
		},
	}
	send(writer, resp)
}

// send marshals and sends a JSON-RPC response
func send(writer *bufio.Writer, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcp: marshal error: %v\n", err)
		return
	}
	body = append(body, '\n')
	if _, err := writer.Write(body); err != nil {
		fmt.Fprintf(os.Stderr, "mcp: write error: %v\n", err)
		return
	}
	_ = writer.Flush()
}

// handleSelfBuild rebuilds the devtools-mcp binary using setup scripts asynchronously
func handleSelfBuild(writer *bufio.Writer, id interface{}, args map[string]interface{}) {
	workDir := os.Getenv("GIT_WORK_DIR")
	if workDir == "" {
		workDir = "."
	}

	binaryPath := "bin/devtools-mcp.exe"
	if runtime.GOOS != "windows" {
		binaryPath = "bin/devtools-mcp"
	}

	// Send response FIRST (before server dies)
	sendResult(writer, id, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "Self-build started in background: killing server and rebuilding binary. Restart VS Code when ready to pick up the new version.",
			},
		},
		"success": true,
		"binary":  binaryPath,
		"workdir": workDir,
	})

	// Spawn an EXTERNAL process (not a goroutine) to avoid being killed with the server
	// This way the rebuild can complete even after the server process is terminated
	if runtime.GOOS == "windows" {
		// Create a batch script that will rebuild asynchronously
		script := fmt.Sprintf(`@echo off
timeout /t 2 /nobreak
cd /d "%s"
call setup.bat
if %%errorlevel%% equ 0 (
  echo Build completed successfully. Restart VS Code to pick up the new binary.
) else (
  echo Build failed
)
`, workDir)
		// Write script to temp file and execute it
		tempScript := filepath.Join(os.TempDir(), "devtools-rebuild.bat")
		if err := os.WriteFile(tempScript, []byte(script), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create rebuild script: %v\n", err)
			return
		}
		// Spawn the script in a detached process
		cmd := exec.Command("cmd", "/c", "start", "/B", tempScript)
		cmd.Run() // Detached, so we don't wait
	} else {
		// For Linux/macOS, create a shell script
		script := fmt.Sprintf(`#!/bin/bash
sleep 2
cd "%s"
if ./setup.sh; then
  echo "Build completed successfully. Restart VS Code to pick up the new binary."
else
  echo "Build failed"
fi
`, workDir)
		tempScript := filepath.Join(os.TempDir(), "devtools-rebuild.sh")
		if err := os.WriteFile(tempScript, []byte(script), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create rebuild script: %v\n", err)
			return
		}
		cmd := exec.Command("sh", tempScript)
		cmd.Start() // Start detached
	}
}
