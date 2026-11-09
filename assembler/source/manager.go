package source

import "errors"

// ErrPkgNotFound is returned by Manager implementations when a requested package or asset does not exist.
var ErrPkgNotFound = errors.New("package not found")

// Manager provides access to package sources and assets.
type Manager interface {
	// Return the sources in the order that they should be parsed,
	// along with the file names of the assets in the package.
	Sources(path string) ([]string, []string, error)

	// Asset returns the binary content of the named asset file in the package.
	Asset(pkg, filename string) ([]byte, error)
}
