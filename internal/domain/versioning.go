package domain

import "context"

// VersioningProvider abstracts the workspace versioning system.
// Any backend (Git, OverlayFS, etc.) implements this interface.
type VersioningProvider interface {
	// Init initializes versioning in a workspace directory.
	Init(ctx context.Context, workspaceDir string) error

	// Snapshot creates a restore point (commit, snapshot, etc.).
	// Returns the ID of the created snapshot.
	Snapshot(ctx context.Context, workspaceDir string, message string) (string, error)

	// Diff returns the differences since the last snapshot.
	Diff(ctx context.Context, workspaceDir string) (string, error)

	// Restore restores the workspace to a previous snapshot.
	Restore(ctx context.Context, workspaceDir string, snapshotID string) error

	// Log returns the snapshot history, limited to 'limit' entries.
	Log(ctx context.Context, workspaceDir string, limit int) ([]SnapshotInfo, error)
}

// SnapshotInfo contains metadata of a snapshot.
type SnapshotInfo struct {
	// ID is the unique identifier of the snapshot (e.g., commit hash).
	ID string

	// Message is the descriptive message of the snapshot.
	Message string

	// Timestamp is the creation date/time (ISO 8601 format).
	Timestamp string
}
