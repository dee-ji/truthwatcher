// Package version exposes build metadata injected by release builds.
package version

var (
	// Version is the Truthwatcher release version. Release builds inject the Git tag.
	Version = "dev"
	// Commit is the Git commit SHA for the build.
	Commit = "unknown"
	// BuildDate is the UTC timestamp when the binary was built.
	BuildDate = "unknown"
)
