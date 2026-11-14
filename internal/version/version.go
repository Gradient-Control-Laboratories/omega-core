// Package version holds the toolchain's release version, which every command reports with --version.
//
// Keep internal/version here because pyproject.toml, .github/workflows/wheels.yml, and release
// -ldflags reference this path.
package version

// Version is the release a build reports. A release build sets it with
// -ldflags "-X gradientcontrol.com/omega/core/internal/version.Version=X.Y.Z"; a development build reports
// the release it is working toward.
var Version = "0.1.2"
