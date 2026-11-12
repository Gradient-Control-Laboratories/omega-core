package source

import (
	"os"
	"path/filepath"

	"gradientcontrol.com/omega/core/assembler/source"
)

// LocalManager searches local filesystem root directories
// for package sources and assets.
type LocalManager struct {
	// Read-only immutable fields set at construction
	rootPaths []string
}

var _ source.Manager = &LocalManager{}

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
	return nil, nil, wrapError(path, source.ErrPkgNotFound)
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
