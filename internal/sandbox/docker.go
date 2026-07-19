package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"agents-execution-engine/internal/domain"
)

const (
	// defaultImage is the Docker image used when none is specified.
	defaultImage = "ubuntu:24.04"

	// readonlyMountPath is the path inside the container where the host workspace is mounted (READ-ONLY).
	readonlyMountPath = "/workspace-readonly"

	// workspaceMountPath is the path inside the container where the agent works (editable copy).
	workspaceMountPath = "/workspace"
)

// DockerProvider implements domain.SandboxProvider using Docker via CLI.
// It does not import any Docker library — it only uses os/exec to invoke the Docker CLI.
type DockerProvider struct{}

// NewDockerProvider creates a new instance of DockerProvider.
func NewDockerProvider() *DockerProvider {
	return &DockerProvider{}
}

// Create creates a new Docker container with the workspace mounted as read-only
// and copies the files to an editable directory inside the container.
// The host is NEVER modified directly — only via CopyToHost.
func (p *DockerProvider) Create(ctx context.Context, opts domain.SandboxConfig) (domain.SandboxInstance, error) {
	image := opts.Image
	if image == "" {
		image = defaultImage
	}

	args := []string{
		"run", "-d",
		"--label", "managed-by=agents-execution-engine",
	}

	// Mount the host workspace as READ-ONLY to protect the original files.
	if opts.WorkspaceDir != "" {
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", opts.WorkspaceDir, readonlyMountPath))
		args = append(args, "-w", workspaceMountPath)
	}

	// Resource limits.
	if opts.MemoryLimit != "" {
		args = append(args, "--memory", opts.MemoryLimit)
	}
	if opts.CPULimit != "" {
		args = append(args, "--cpus", opts.CPULimit)
	}

	// Environment variables.
	for k, v := range opts.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Image and command to keep the container active.
	args = append(args, image, "tail", "-f", "/dev/null")

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create docker container: %w\nstderr: %s", err, stderr.String())
	}

	containerID := strings.TrimSpace(stdout.String())
	if containerID == "" {
		return nil, fmt.Errorf("docker run returned empty container ID")
	}

	log.Printf("[INFO] DockerProvider: Container %s created, copying workspace files...", containerID[:12])

	// Copy the files from the read-only mount to the editable directory inside the container.
	// This ensures that the agent works on an isolated copy.
	if opts.WorkspaceDir != "" {
		copyArgs := []string{
			"exec", containerID, "sh", "-c",
			fmt.Sprintf("mkdir -p %s && cp -r %s/. %s/", workspaceMountPath, readonlyMountPath, workspaceMountPath),
		}

		var copyStderr bytes.Buffer
		copyCmd := exec.CommandContext(ctx, "docker", copyArgs...)
		copyCmd.Stderr = &copyStderr

		if err := copyCmd.Run(); err != nil {
			// Clean up the container if the copy fails.
			_ = exec.CommandContext(ctx, "docker", "rm", "-f", containerID).Run()
			return nil, fmt.Errorf("failed to copy workspace files into container: %w\nstderr: %s", err, copyStderr.String())
		}

		log.Printf("[INFO] DockerProvider: Workspace files copied successfully into container %s", containerID[:12])
	}

	return &dockerInstance{
		containerID:  containerID,
		workspaceDir: opts.WorkspaceDir,
	}, nil
}
