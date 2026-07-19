package domain

import "context"

// SandboxProvider abstracts the isolated execution environment.
// Any runtime (Docker, Firecracker, gVisor, etc.) implements this interface.
type SandboxProvider interface {
	// Create creates a new sandbox with the workspace mounted.
	Create(ctx context.Context, opts SandboxConfig) (SandboxInstance, error)
}

// SandboxConfig contains the configuration required to create a sandbox.
type SandboxConfig struct {
	// Image is the base image of the runtime (e.g., "golang:1.22", "ubuntu:24.04").
	Image string

	// WorkspaceDir is the local path of the workspace to mount inside the sandbox.
	WorkspaceDir string

	// Env contains environment variables to inject into the sandbox.
	Env map[string]string

	// MemoryLimit defines the memory limit (e.g., "512m", "1g").
	MemoryLimit string

	// CPULimit defines the CPU limit (e.g., "1.0", "0.5").
	CPULimit string
}

// SandboxInstance represents an active and running sandbox.
type SandboxInstance interface {
	// ID returns the unique identifier of the sandbox.
	ID() string

	// Exec executes a command inside the sandbox.
	Exec(ctx context.Context, cmd Command) (CommandResult, error)

	// WriteFile writes a file to the sandbox workspace.
	WriteFile(ctx context.Context, op FileOperation) error

	// ReadFile reads the content of a file from the sandbox workspace.
	ReadFile(ctx context.Context, path string) (string, error)

	// ListFiles lists the files in a workspace directory.
	ListFiles(ctx context.Context, path string) ([]FileInfo, error)

	// Destroy destroys the sandbox and releases all associated resources.
	Destroy(ctx context.Context) error

	// CopyToHost copies modified files from the sandbox back to the host workspace.
	// This is the ONLY way changes inside the sandbox can reach the host filesystem.
	CopyToHost(ctx context.Context) error
}

// FileInfo contains metadata of a file in the workspace.
type FileInfo struct {
	Path  string
	IsDir bool
	Size  int64
}
