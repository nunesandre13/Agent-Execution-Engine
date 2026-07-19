package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"

	"agents-execution-engine/internal/domain"
)

// WriteFile writes a file to the sandbox workspace.
func (h *Handlers) WriteFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sandboxID, err := requiredString(request, "sandbox_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	path, err := requiredString(request, "path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := requiredString(request, "content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sb, err := h.getSandbox(sandboxID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	op := domain.FileOperation{
		Path:    path,
		Content: content,
	}

	if err := sb.WriteFile(ctx, op); err != nil {
		log.Printf("[ERROR] WriteFile: failed to write %s: %v", path, err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to write file: %v", err)), nil
	}

	log.Printf("[INFO] WriteFile: Wrote file %s in sandbox %s", path, sandboxID)

	return mcp.NewToolResultText(fmt.Sprintf("File written successfully: %s", path)), nil
}

// ReadFile reads the content of a file from the sandbox workspace.
func (h *Handlers) ReadFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sandboxID, err := requiredString(request, "sandbox_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	path, err := requiredString(request, "path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sb, err := h.getSandbox(sandboxID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := sb.ReadFile(ctx, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	return mcp.NewToolResultText(content), nil
}

// ListFiles lists files in a workspace directory.
func (h *Handlers) ListFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sandboxID, err := requiredString(request, "sandbox_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sb, err := h.getSandbox(sandboxID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	path := optionalString(request, "path")

	files, err := sb.ListFiles(ctx, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list files: %v", err)), nil
	}

	jsonBytes, _ := json.MarshalIndent(files, "", "  ")
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
