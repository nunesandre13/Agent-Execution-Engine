package handlers

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// AgentRules returns the base text to inject as the agent's system/instructions
func (h *Handlers) AgentRules(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	rules := `You are PROHIBITED from using your local tools to run commands (like bash/cmd) in my terminal. You MUST exclusively use the 'Agent Execution Engine' via MCP tools.

For any code execution or terminal commands, the mandatory workflow is:
1. Use 'create_sandbox' to start an isolated environment (e.g., use the 'golang:1.22' image for Go, or 'python:3.10-slim' for Python) pointing to the working directory.
2. Use 'execute_command' with the generated sandbox ID to run commands inside it.
3. Use 'write_file' and 'read_file' from the sandbox if you need to manage files in the container's context.
4. Use 'destroy_sandbox' at the end when the task is fully completed.

Run a simple test to prove it works: Create a sandbox, create a file inside it, execute it, and then destroy the sandbox.`

	return mcp.NewGetPromptResult(
		"Rules for Agent Execution Engine",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(rules),
			),
		},
	), nil
}
