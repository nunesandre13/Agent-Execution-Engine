package versioning

import (
	"context"
	"fmt"
	"strings"
)

// Init initializes a Git repository in the workspace directory.
// If a Git repo already exists, it does nothing.
func (g *GitProvider) Init(ctx context.Context, workspaceDir string) error {
	// Initialize the repository.
	if err := g.runGit(ctx, workspaceDir, "init"); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// Configure user for commits (required inside containers without global config).
	if err := g.runGit(ctx, workspaceDir, "config", "user.email", "agent@execution-engine.local"); err != nil {
		return fmt.Errorf("git config email failed: %w", err)
	}
	if err := g.runGit(ctx, workspaceDir, "config", "user.name", "Agent Engine"); err != nil {
		return fmt.Errorf("git config name failed: %w", err)
	}

	// Create initial commit if the repo is empty.
	stdout, err := g.runGitOutput(ctx, workspaceDir, "rev-parse", "HEAD")
	if err != nil || strings.TrimSpace(stdout) == "" {
		// Empty repo — create initial commit.
		if err := g.runGit(ctx, workspaceDir, "add", "-A"); err != nil {
			return fmt.Errorf("git add failed: %w", err)
		}
		// Use --allow-empty in case there are no files.
		if err := g.runGit(ctx, workspaceDir, "commit", "--allow-empty", "-m", "initial commit"); err != nil {
			return fmt.Errorf("git initial commit failed: %w", err)
		}
	}

	return nil
}
