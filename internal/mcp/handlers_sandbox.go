package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"agents-execution-engine/internal/domain"
)

// ── Sandbox Handlers ──

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
