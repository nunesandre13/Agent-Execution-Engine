package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"

	"agents-execution-engine/internal/domain"
)

// CreateSandbox creates a new sandbox and registers it in the registry.
func (h *Handlers) CreateSandbox(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspaceDir, err := requiredString(request, "workspace_dir")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	config := domain.SandboxConfig{
		Image:        optionalString(request, "image"),
		WorkspaceDir: workspaceDir,
		MemoryLimit:  optionalString(request, "memory_limit"),
		CPULimit:     optionalString(request, "cpu_limit"),
		Env:          make(map[string]string),
	}

	// Convert to absolute path to avoid dangerous relative mounts in Docker.
	absPath, err := filepath.Abs(workspaceDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve absolute path: %v", err)), nil
	}
	workspaceDir = absPath
	config.WorkspaceDir = absPath

	// Validate that the folder exists on the host (required for read-only mount).
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("workspace directory does not exist on host: %s", workspaceDir)), nil
	}

	// Initialize versioning in the workspace.
	if err := h.engine.Versioning().Init(ctx, workspaceDir); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to init versioning: %v", err)), nil
	}

	instance, err := h.engine.Sandbox().Create(ctx, config)
	if err != nil {
		log.Printf("[ERROR] CreateSandbox: failed to create: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to create sandbox: %v", err)), nil
	}

	log.Printf("[INFO] CreateSandbox: Created sandbox %s for workspace %s", instance.ID(), workspaceDir)

	// Register the instance.
	h.mu.Lock()
	if h.sandboxes == nil {
		h.sandboxes = make(map[string]domain.SandboxInstance)
	}
	h.sandboxes[instance.ID()] = instance
	h.mu.Unlock()

	return mcp.NewToolResultText(fmt.Sprintf("Sandbox created successfully.\nID: %s", instance.ID())), nil
}
