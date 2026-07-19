// Package mcp provides the MCP adapter that exposes the Engine's capabilities
// as MCP Tools. This is the ONLY package that imports the mcp-go library.
package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"agents-execution-engine/internal/domain"
)

// Server encapsulates the MCP server and all transport logic.
// No other package needs to import mcp-go directly.
type Server struct {
	inner *server.MCPServer
}

// NewServer creates an MCP Server configured with all registered tools.
// It receives the Engine which provides access to the sandbox and versioning.
func NewServer(engine *domain.Engine) *Server {
	s := server.NewMCPServer(
		"agents-execution-engine",
		"0.1.0",
	)

	h := &Handlers{engine: engine}
	registerTools(s, h)

	return &Server{inner: s}
}

// ServeStdio starts the MCP server via stdin/stdout (standard transport for AI agents).
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.inner)
}

// registerTools registers all MCP tools in the server.
func registerTools(s *server.MCPServer, h *Handlers) {
	// ── Sandbox Tools ──

	s.AddTool(
		mcp.NewTool("create_sandbox",
			mcp.WithDescription("Create a new isolated sandbox environment for code execution"),
			mcp.WithString("image", mcp.Description("Base image for the sandbox (e.g. 'ubuntu:24.04', 'golang:1.22')")),
			mcp.WithString("workspace_dir", mcp.Required(), mcp.Description("Local directory to mount as the workspace")),
			mcp.WithString("memory_limit", mcp.Description("Memory limit (e.g. '512m', '1g')")),
			mcp.WithString("cpu_limit", mcp.Description("CPU limit (e.g. '1.0', '0.5')")),
		),
		h.CreateSandbox,
	)

	s.AddTool(
		mcp.NewTool("execute_command",
			mcp.WithDescription("Execute a command inside the sandbox"),
			mcp.WithString("sandbox_id", mcp.Required(), mcp.Description("ID of the sandbox to execute in")),
			mcp.WithString("command", mcp.Required(), mcp.Description("The command to execute (e.g. 'go build ./...')")),
			mcp.WithString("workdir", mcp.Description("Working directory relative to the workspace")),
			mcp.WithNumber("timeout_seconds", mcp.Description("Maximum execution time in seconds")),
		),
		h.ExecuteCommand,
	)

	s.AddTool(
		mcp.NewTool("write_file",
			mcp.WithDescription("Write content to a file in the sandbox workspace"),
			mcp.WithString("sandbox_id", mcp.Required(), mcp.Description("ID of the sandbox")),
			mcp.WithString("path", mcp.Required(), mcp.Description("File path relative to the workspace")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Content to write to the file")),
		),
		h.WriteFile,
	)

	s.AddTool(
		mcp.NewTool("read_file",
			mcp.WithDescription("Read the content of a file from the sandbox workspace"),
			mcp.WithString("sandbox_id", mcp.Required(), mcp.Description("ID of the sandbox")),
			mcp.WithString("path", mcp.Required(), mcp.Description("File path relative to the workspace")),
		),
		h.ReadFile,
	)

	s.AddTool(
		mcp.NewTool("list_files",
			mcp.WithDescription("List files in a directory of the sandbox workspace"),
			mcp.WithString("sandbox_id", mcp.Required(), mcp.Description("ID of the sandbox")),
			mcp.WithString("path", mcp.Description("Directory path relative to the workspace (empty for root)")),
		),
		h.ListFiles,
	)

	s.AddTool(
		mcp.NewTool("destroy_sandbox",
			mcp.WithDescription("Destroy a sandbox and release all associated resources"),
			mcp.WithString("sandbox_id", mcp.Required(), mcp.Description("ID of the sandbox to destroy")),
		),
		h.DestroySandbox,
	)

	s.AddTool(
		mcp.NewTool("commit_to_host",
			mcp.WithDescription("Copy all modified files from the sandbox back to the host workspace. The host filesystem is NEVER modified until this tool is called. Use this only after the user has reviewed and approved the changes (e.g. via the 'diff' tool)."),
			mcp.WithString("sandbox_id", mcp.Required(), mcp.Description("ID of the sandbox to commit changes from")),
		),
		h.CommitToHost,
	)

	// ── Versioning Tools ──

	s.AddTool(
		mcp.NewTool("snapshot",
			mcp.WithDescription("Create a versioning snapshot (checkpoint) of the current workspace state"),
			mcp.WithString("workspace_dir", mcp.Required(), mcp.Description("Path to the workspace directory")),
			mcp.WithString("message", mcp.Required(), mcp.Description("Descriptive message for this snapshot")),
		),
		h.Snapshot,
	)

	s.AddTool(
		mcp.NewTool("diff",
			mcp.WithDescription("Get the diff of changes since the last snapshot"),
			mcp.WithString("workspace_dir", mcp.Required(), mcp.Description("Path to the workspace directory")),
		),
		h.Diff,
	)

	s.AddTool(
		mcp.NewTool("restore",
			mcp.WithDescription("Restore the workspace to a previous snapshot"),
			mcp.WithString("workspace_dir", mcp.Required(), mcp.Description("Path to the workspace directory")),
			mcp.WithString("snapshot_id", mcp.Required(), mcp.Description("ID of the snapshot to restore to")),
		),
		h.Restore,
	)

	s.AddTool(
		mcp.NewTool("log",
			mcp.WithDescription("List the history of workspace snapshots"),
			mcp.WithString("workspace_dir", mcp.Required(), mcp.Description("Path to the workspace directory")),
			mcp.WithNumber("limit", mcp.Description("Maximum number of snapshots to return (default: 20)")),
		),
		h.Log,
	)

	// ── Prompts ──
	s.AddPrompt(
		mcp.NewPrompt("agent_rules",
			mcp.WithPromptDescription("Mandatory rules to make the agent use the execution engine sandbox instead of local commands"),
		),
		h.AgentRules,
	)
}
