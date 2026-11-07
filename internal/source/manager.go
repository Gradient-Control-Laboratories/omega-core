package source

import (
	"os"
	"path/filepath"
)

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
type LocalManager struct {
	// Read-only immutable fields set at construction
	rootPaths []string
}

var _ Manager = &LocalManager{}

// NewLocalManager creates a new instance of the source manager searching rootPaths.
func NewLocalManager(rootPaths ...string) *LocalManager {
	return &LocalManager{rootPaths: rootPaths}
}

// Sources searches the root paths for the given package.
func (m *LocalManager) Sources(path string) ([]string, []string, error) {
	for _, rootPath := range m.rootPaths {
		localPath := filepath.Join(rootPath, path)
		if stat, err := os.Stat(localPath); err == nil && stat.IsDir() {
			sources, assetNames, err := loadLocalSources(localPath)
			if err != nil {
				return nil, nil, wrapError(path, err)
			}
			if len(sources) > 0 {
				return sources, assetNames, nil
			}
		}
	}
	if sources, ok := embeddedSources(path); ok {
		return sources, nil, nil
	}
	return nil, nil, wrapError(path, ErrPkgNotFound)
}

// Asset searches root paths for the given asset file in the package.
func (m *LocalManager) Asset(pkg, filename string) ([]byte, error) {
	for _, rootPath := range m.rootPaths {
		candidates := []string{
			filepath.Join(rootPath, pkg, filename),
			filepath.Join(rootPath, pkg, "assets", filename),
		}
		for _, cand := range candidates {
			if stat, err := os.Stat(cand); err == nil && !stat.IsDir() {
				data, err := os.ReadFile(cand)
				if err != nil {
					return nil, wrapError(pkg, err)
				}
				return data, nil
			}
		}
	}
	return nil, nil
}
