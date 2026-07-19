// Package sandbox provides implementations of domain.SandboxProvider.
//
// The current implementation uses Docker via CLI (os/exec).
// To add a new runtime (e.g. Firecracker), simply create a new file
// in this package that implements domain.SandboxProvider and domain.SandboxInstance.
package sandbox
