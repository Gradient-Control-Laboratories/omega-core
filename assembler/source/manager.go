package source

import "gradientcontrol.com/omega/core/internal/source"

// Manager provides access to package sources and assets.
type Manager = source.Manager

// LocalManager searches local filesystem root directories and embedded packages
// for package sources and assets.
type LocalManager = source.LocalManager

// NewLocalManager creates a new instance of the source manager searching rootPaths.
func NewLocalManager(rootPaths ...string) *LocalManager {
	return source.NewLocalManager(rootPaths...)
}
