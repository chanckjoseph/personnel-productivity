package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
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
