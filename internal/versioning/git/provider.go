package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// GitProvider implements domain.VersioningProvider using Git via CLI.
// It does not import any Git library — it only uses os/exec to invoke the Git CLI.
type GitProvider struct{}

// NewGitProvider creates a new instance of GitProvider.
func NewGitProvider() *GitProvider {
	return &GitProvider{}
}

// runGit executes a git command in the given directory and discards the output.
func (g *GitProvider) runGit(ctx context.Context, dir string, args ...string) error {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w\nstderr: %s", err, stderr.String())
	}
	return nil
}

// runGitOutput executes a git command and returns the stdout.
func (g *GitProvider) runGitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w\nstderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}
