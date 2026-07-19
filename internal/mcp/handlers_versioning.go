package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ── Versioning Handlers ──

// Snapshot creates a snapshot of the workspace.
func (h *Handlers) Snapshot(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspaceDir, err := requiredString(request, "workspace_dir")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	message, err := requiredString(request, "message")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	snapshotID, err := h.engine.Versioning().Snapshot(ctx, workspaceDir, message)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create snapshot: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Snapshot created.\nID: %s", snapshotID)), nil
}

// Diff returns the differences since the last snapshot.
func (h *Handlers) Diff(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspaceDir, err := requiredString(request, "workspace_dir")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	diff, err := h.engine.Versioning().Diff(ctx, workspaceDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get diff: %v", err)), nil
	}

	if diff == "" {
		return mcp.NewToolResultText("No changes since last snapshot."), nil
	}

	return mcp.NewToolResultText(diff), nil
}

// Restore restores the workspace to a previous snapshot.
func (h *Handlers) Restore(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspaceDir, err := requiredString(request, "workspace_dir")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	snapshotID, err := requiredString(request, "snapshot_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := h.engine.Versioning().Restore(ctx, workspaceDir, snapshotID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to restore: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Workspace restored to snapshot: %s", snapshotID)), nil
}

// Log returns the snapshot history.
func (h *Handlers) Log(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspaceDir, err := requiredString(request, "workspace_dir")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	limit := int(optionalNumber(request, "limit"))
	if limit <= 0 {
		limit = 20
	}

	snapshots, err := h.engine.Versioning().Log(ctx, workspaceDir, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get log: %v", err)), nil
	}

	jsonBytes, _ := json.MarshalIndent(snapshots, "", "  ")
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
