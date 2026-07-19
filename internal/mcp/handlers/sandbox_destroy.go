package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
)

// DestroySandbox destroys a sandbox and removes it from the registry.
func (h *Handlers) DestroySandbox(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sandboxID, err := requiredString(request, "sandbox_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sb, err := h.getSandbox(sandboxID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := sb.Destroy(ctx); err != nil {
		log.Printf("[ERROR] DestroySandbox: failed to destroy %s: %v", sandboxID, err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to destroy sandbox: %v", err)), nil
	}

	log.Printf("[INFO] DestroySandbox: Destroyed sandbox %s", sandboxID)

	// Remove from the registry.
	h.mu.Lock()
	delete(h.sandboxes, sandboxID)
	h.mu.Unlock()

	return mcp.NewToolResultText(fmt.Sprintf("Sandbox %s destroyed successfully", sandboxID)), nil
}

// CommitToHost copies all modified files from the sandbox back to the host workspace.
// This is the ONLY way changes inside the sandbox can reach the host filesystem.
// Should only be called after the user has reviewed the changes (e.g. via the 'diff' tool).
func (h *Handlers) CommitToHost(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sandboxID, err := requiredString(request, "sandbox_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sb, err := h.getSandbox(sandboxID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	log.Printf("[INFO] CommitToHost: Copying changes from sandbox %s to host...", sandboxID)

	if err := sb.CopyToHost(ctx); err != nil {
		log.Printf("[ERROR] CommitToHost: failed for sandbox %s: %v", sandboxID, err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to commit changes to host: %v", err)), nil
	}

	log.Printf("[INFO] CommitToHost: Changes from sandbox %s committed to host successfully", sandboxID)

	return mcp.NewToolResultText(fmt.Sprintf("Changes from sandbox %s have been committed to the host workspace successfully.", sandboxID)), nil
}
