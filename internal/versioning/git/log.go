package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"agents-execution-engine/internal/domain"
)

// Log returns the commit history, limited to 'limit' entries.
func (g *GitProvider) Log(ctx context.Context, workspaceDir string, limit int) ([]domain.SnapshotInfo, error) {
	args := []string{
		"log",
		"--format=%H|||%s|||%aI",
		"-n", strconv.Itoa(limit),
	}

	stdout, err := g.runGitOutput(ctx, workspaceDir, args...)
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	var snapshots []domain.SnapshotInfo
	for line := range strings.SplitSeq(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|||", 3)
		if len(parts) < 3 {
			continue
		}
		snapshots = append(snapshots, domain.SnapshotInfo{
			ID:        parts[0],
			Message:   parts[1],
			Timestamp: parts[2],
		})
	}

	return snapshots, nil
}
