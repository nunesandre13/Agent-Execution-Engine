package git

import (
	"context"
	"fmt"
)

// Restore restores the workspace to a previous snapshot (commit).
func (g *GitProvider) Restore(ctx context.Context, workspaceDir string, snapshotID string) error {
	// Checkout the specified snapshot.
	if err := g.runGit(ctx, workspaceDir, "checkout", snapshotID, "--", "."); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	// Clean untracked files.
	if err := g.runGit(ctx, workspaceDir, "clean", "-fd"); err != nil {
		return fmt.Errorf("git clean failed: %w", err)
	}

	return nil
}
