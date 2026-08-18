package git

import (
	"context"
	"fmt"
	"strings"
)

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
		for file := range strings.SplitSeq(untrackedFiles, "\n") {
			file = strings.TrimSpace(file)
			if file == "" {
				continue
			}
			result.WriteString(fmt.Sprintf("\n--- /dev/null\n+++ b/%s\n(new file)\n", file))
		}
	}

	return result.String(), nil
}
