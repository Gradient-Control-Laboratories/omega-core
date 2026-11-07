package source

import "gradientcontrol.com/omega/core/internal/source"

// Manager provides access to package sources and assets.
type Manager interface {
	// Return the sources in the order that they should be parsed,
	// along with the file names of the assets in the package.
	Sources(path string) ([]string, []string, error)

	// Asset returns the binary content of the named asset file in the package.
	Asset(pkg, filename string) ([]byte, error)
}

// LocalManager searches local filesystem root directories and embedded packages
// for package sources and assets.
type LocalManager = source.LocalManager

// NewLocalManager creates a new instance of the source manager searching rootPaths.
func NewLocalManager(rootPaths ...string) *LocalManager {
	return source.NewLocalManager(rootPaths...)
}
