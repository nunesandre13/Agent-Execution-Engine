package main

import (
	"log"
	"os"

	"agents-execution-engine/internal/domain"
	mcpserver "agents-execution-engine/internal/mcp"
	"agents-execution-engine/internal/sandbox"
	"agents-execution-engine/internal/versioning"
)

func main() {
	log.SetOutput(os.Stderr)
	log.Println("Starting Agents Execution Engine...")

	// 1. Create concrete providers.
	// To replace Docker with micro-VMs, replace this line.
	sandboxProvider := sandbox.NewDockerProvider()

	// To replace Git with OverlayFS, replace this line.
	versioningProvider := versioning.NewGitProvider()

	// 2. Assemble the engine with the providers.
	engine := domain.NewEngine(sandboxProvider, versioningProvider)

	// 3. Create the MCP server (registers all tools).
	mcpSrv := mcpserver.NewServer(engine)

	// 4. Serve via stdio (integration with AI agents).
	log.Println("MCP Server ready. Listening on stdio...")
	if err := mcpSrv.ServeStdio(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
