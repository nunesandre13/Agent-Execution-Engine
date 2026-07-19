// Package versioning provides implementations of domain.VersioningProvider.
//
// The current implementation uses Git via CLI (os/exec).
// To add a new backend (e.g. OverlayFS), simply create a new file
// in this package that implements domain.VersioningProvider.
package versioning
