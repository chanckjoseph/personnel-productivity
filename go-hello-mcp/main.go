package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type toolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func main() {
	fmt.Fprintln(os.Stderr, "hello-go-mcp: server started")
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

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

		if len(req.ID) == 0 {
			if req.Method == "notifications/initialized" {
				continue
			}
			continue
		}

		id, err := parseID(req.ID)
		if err != nil {
			sendError(writer, nil, -32600, "invalid request id")
			continue
		}

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
					"name":    "hello-go-mcp",
					"version": "1.0.0",
				},
			}
			sendResult(writer, id, result)
		case "tools/list":
			tools := []toolDefinition{
				{
					Name:        "hello_greeting",
					Description: "Returns a greeting message to the agent",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Optional name to greet",
							},
						},
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

			if params.Name != "hello_greeting" {
				sendError(writer, id, -32602, fmt.Sprintf("unknown tool: %s", params.Name))
				continue
			}

			name := "agent"
			if rawName, ok := params.Arguments["name"]; ok {
				if parsedName, ok := rawName.(string); ok && strings.TrimSpace(parsedName) != "" {
					name = strings.TrimSpace(parsedName)
				}
			}

			result := map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": fmt.Sprintf("Hello, %s! Greeting from your Go MCP server.", name),
					},
				},
			}
			sendResult(writer, id, result)
		case "ping":
			sendResult(writer, id, map[string]interface{}{})
		default:
			sendError(writer, id, -32601, "method not found")
		}
	}
}

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

func sendResult(writer *bufio.Writer, id interface{}, result interface{}) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	send(writer, resp)
}

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

func readFrame(reader *bufio.Reader) ([]byte, error) {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(line) > 0 {
				// last line without newline
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					return []byte(trimmed), nil
				}
			}
			return nil, err
		}
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return []byte(trimmed), nil
		}
	}
}
