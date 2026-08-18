package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"agents-execution-engine/internal/domain"
)

// dockerInstance implements domain.SandboxInstance for a Docker container.
type dockerInstance struct {
	containerID  string
	workspaceDir string
}

// ID returns the identifier of the Docker container.
func (d *dockerInstance) ID() string {
	return d.containerID
}

// Exec executes a command inside the Docker container.
func (d *dockerInstance) Exec(ctx context.Context, cmd domain.Command) (domain.CommandResult, error) {
	start := time.Now()

	args := []string{"exec"}

	// Working directory.
	if cmd.WorkDir != "" {
		args = append(args, "-w", fmt.Sprintf("%s/%s", workspaceMountPath, cmd.WorkDir))
	}

	// Environment variables.
	for k, v := range cmd.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, d.containerID, "sh", "-c", cmd.Instruction)

	// Apply timeout if set.
	execCtx := ctx
	if cmd.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, cmd.Timeout)
		defer cancel()
	}

	var stdout, stderr bytes.Buffer
	dockerCmd := exec.CommandContext(execCtx, "docker", args...)
	dockerCmd.Stdout = &stdout
	dockerCmd.Stderr = &stderr

	err := dockerCmd.Run()
	duration := time.Since(start)

	result := domain.CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		}
	}

	return result, nil
}

// WriteFile writes a file to the container workspace.
func (d *dockerInstance) WriteFile(ctx context.Context, op domain.FileOperation) error {
	filePath := fmt.Sprintf("%s/%s", workspaceMountPath, op.Path)

	// Use docker exec with sh -c and heredoc to write the file.
	instruction := fmt.Sprintf("mkdir -p $(dirname %s) && cat > %s << 'AGENTS_EOF'\n%s\nAGENTS_EOF", filePath, filePath, op.Content)

	args := []string{"exec", d.containerID, "sh", "-c", instruction}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write file %s in container %s: %w\nstderr: %s", op.Path, d.containerID, err, stderr.String())
	}

	return nil
}

// ReadFile reads the content of a file from the container workspace.
func (d *dockerInstance) ReadFile(ctx context.Context, path string) (string, error) {
	filePath := fmt.Sprintf("%s/%s", workspaceMountPath, path)

	args := []string{"exec", d.containerID, "cat", filePath}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to read file %s in container %s: %w\nstderr: %s", path, d.containerID, err, stderr.String())
	}

	return stdout.String(), nil
}

// ListFiles lists the files in a container workspace directory.
func (d *dockerInstance) ListFiles(ctx context.Context, path string) ([]domain.FileInfo, error) {
	dirPath := workspaceMountPath
	if path != "" {
		dirPath = fmt.Sprintf("%s/%s", workspaceMountPath, path)
	}

	// Use ls -la with parseable format.
	args := []string{"exec", d.containerID, "find", dirPath, "-maxdepth", "1", "-printf", "%y %s %P\n"}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list files in %s in container %s: %w\nstderr: %s", path, d.containerID, err, stderr.String())
	}

	var files []domain.FileInfo
	for line := range strings.SplitSeq(strings.TrimSpace(stdout.String()), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			continue
		}

		var size int64
		fmt.Sscanf(parts[1], "%d", &size)

		files = append(files, domain.FileInfo{
			Path:  parts[2],
			IsDir: parts[0] == "d",
			Size:  size,
		})
	}

	return files, nil
}

// Destroy removes the Docker container and releases the resources.
func (d *dockerInstance) Destroy(ctx context.Context) error {
	args := []string{"rm", "-f", d.containerID}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to destroy container %s: %w\nstderr: %s", d.containerID, err, stderr.String())
	}

	return nil
}

// CopyToHost copies the modified workspace from the container back to the host.
// This is the ONLY way changes inside the sandbox can reach the host filesystem.
// Uses 'docker cp' to extract /workspace contents to the original host directory.
func (d *dockerInstance) CopyToHost(ctx context.Context) error {
	if d.workspaceDir == "" {
		return fmt.Errorf("no workspace directory configured for this sandbox")
	}

	// docker cp container:/workspace/. host_path/
	// The trailing "/." copies the contents of the directory, not the directory itself.
	src := fmt.Sprintf("%s:%s/.", d.containerID, workspaceMountPath)

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "cp", src, d.workspaceDir)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy workspace to host: %w\nstderr: %s", err, stderr.String())
	}

	return nil
}
