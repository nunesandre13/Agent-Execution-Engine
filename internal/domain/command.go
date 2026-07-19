package domain

import "time"

// Command represents a command to execute inside the sandbox.
type Command struct {
	// ID is the unique identifier of the command (UUID).
	ID string

	// Instruction is the actual command (e.g., "go build ./...").
	Instruction string

	// WorkDir is the working directory relative to the workspace.
	WorkDir string

	// Env contains additional environment variables.
	Env map[string]string

	// Timeout is the maximum execution time.
	Timeout time.Duration
}

// CommandResult is the result of executing a Command.
type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// FileOperation represents a file read/write operation in the workspace.
type FileOperation struct {
	// Path is the path relative to the workspace.
	Path string

	// Content is the content of the file.
	Content string
}
