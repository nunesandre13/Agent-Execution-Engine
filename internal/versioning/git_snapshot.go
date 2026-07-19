package versioning

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

// Diff returns the differences since the last commit.
func (g *GitProvider) Diff(ctx context.Context, workspaceDir string) (string, error) {
	// Diff of tracked (modified) files.
	trackedDiff, err := g.runGitOutput(ctx, workspaceDir, "diff", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	// Diff of untracked (new) files.
	untrackedOutput, err := g.runGitOutput(ctx, workspaceDir, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return "", fmt.Errorf("git ls-files failed: %w", err)
	}

	var result strings.Builder
	if trackedDiff != "" {
		result.WriteString(trackedDiff)
	}

	// Add new files to the diff.
	untrackedFiles := strings.TrimSpace(untrackedOutput)
	if untrackedFiles != "" {
		for _, file := range strings.Split(untrackedFiles, "\n") {
			file = strings.TrimSpace(file)
			if file == "" {
				continue
			}
			result.WriteString(fmt.Sprintf("\n--- /dev/null\n+++ b/%s\n(new file)\n", file))
		}
	}

	return result.String(), nil
}
