package handlers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"

	"agents-execution-engine/internal/domain"
)

// Handlers contains the handlers for each MCP tool.
// It maintains a registry of active sandbox instances.
type Handlers struct {
	engine    *domain.Engine
	mu        sync.RWMutex
	sandboxes map[string]domain.SandboxInstance
}

// NewHandlers creates a new instance of Handlers.
func NewHandlers(engine *domain.Engine) *Handlers {
	return &Handlers{
		engine:    engine,
		sandboxes: make(map[string]domain.SandboxInstance),
	}
}

// getSandbox returns a sandbox instance by ID, or an error if it does not exist.
func (h *Handlers) getSandbox(id string) (domain.SandboxInstance, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sb, ok := h.sandboxes[id]
	if !ok {
		return nil, fmt.Errorf("sandbox not found: %s", id)
	}
	return sb, nil
}

// ── Helpers to extract arguments from MCP requests ──

func requiredString(req mcp.CallToolRequest, key string) (string, error) {
	args := req.GetArguments()
	val, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required argument: %s", key)
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("argument %s must be a string", key)
	}
	if s == "" {
		return "", fmt.Errorf("argument %s cannot be empty", key)
	}
	s = strings.Trim(s, "\"'")
	return s, nil
}

func optionalString(req mcp.CallToolRequest, key string) string {
	args := req.GetArguments()
	val, ok := args[key]
	if !ok {
		return ""
	}
	s, _ := val.(string)
	s = strings.Trim(s, "\"'")
	return s
}

func optionalNumber(req mcp.CallToolRequest, key string) float64 {
	args := req.GetArguments()
	val, ok := args[key]
	if !ok {
		return 0
	}
	n, _ := val.(float64)
	return n
}
