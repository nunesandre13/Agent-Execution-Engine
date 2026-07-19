package git

import (
	"context"
	"fmt"
	"strings"
)

// Snapshot creates a commit with all current changes.
// Returns the short hash of the created commit.
func (g *GitProvider) Snapshot(ctx context.Context, workspaceDir string, message string) (string, error) {
	// Stage all changes.
	if err := g.runGit(ctx, workspaceDir, "add", "-A"); err != nil {
		return "", fmt.Errorf("git add failed: %w", err)
	}

	// Check if there is anything to commit.
	statusOutput, err := g.runGitOutput(ctx, workspaceDir, "status", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}
	if strings.TrimSpace(statusOutput) == "" {
		// Nothing to commit — return the current HEAD.
		hash, err := g.runGitOutput(ctx, workspaceDir, "rev-parse", "--short", "HEAD")
		if err != nil {
			return "", fmt.Errorf("git rev-parse failed: %w", err)
		}
		return strings.TrimSpace(hash), nil
	}

	// Create the commit.
	if err := g.runGit(ctx, workspaceDir, "commit", "-m", message); err != nil {
		return "", fmt.Errorf("git commit failed: %w", err)
	}

	// Return the hash of the created commit.
	hash, err := g.runGitOutput(ctx, workspaceDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}

	return strings.TrimSpace(hash), nil
}
