package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"agents-execution-engine/internal/domain"
)

// ExecuteCommand executes a command inside an existing sandbox.
func (h *Handlers) ExecuteCommand(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sandboxID, err := requiredString(request, "sandbox_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	command, err := requiredString(request, "command")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sb, err := h.getSandbox(sandboxID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cmd := domain.Command{
		Instruction: command,
		WorkDir:     optionalString(request, "workdir"),
	}

	log.Printf("[INFO] ExecuteCommand: Sandbox %s running: %s", sandboxID, command)

	// Optional timeout.
	if timeoutSec := optionalNumber(request, "timeout_seconds"); timeoutSec > 0 {
		cmd.Timeout = time.Duration(timeoutSec) * time.Second
	}

	result, err := sb.Exec(ctx, cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("execution error: %v", err)), nil
	}

	// Format result as JSON for better parseability.
	output := map[string]interface{}{
		"exit_code": result.ExitCode,
		"stdout":    result.Stdout,
		"stderr":    result.Stderr,
		"duration":  result.Duration.String(),
	}
	jsonBytes, _ := json.MarshalIndent(output, "", "  ")

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
