package domain

// Engine orchestrates command execution in the sandbox
// with workspace versioning.
// Serves as the central point of access to the configured providers.
type Engine struct {
	sandbox    SandboxProvider
	versioning VersioningProvider
}

// NewEngine creates a new instance of the Engine with the given providers.
func NewEngine(sandbox SandboxProvider, versioning VersioningProvider) *Engine {
	return &Engine{
		sandbox:    sandbox,
		versioning: versioning,
	}
}

// Sandbox returns the configured SandboxProvider.
func (e *Engine) Sandbox() SandboxProvider {
	return e.sandbox
}

// Versioning returns the configured VersioningProvider.
func (e *Engine) Versioning() VersioningProvider {
	return e.versioning
}
