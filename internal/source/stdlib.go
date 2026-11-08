package source

import (
	"fmt"

	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// StdLibManager provides access to embedded standard library packages.
type StdLibManager struct{}

var _ Manager = &StdLibManager{}

// NewStdLibManager creates a new instance of the embedded standard library source manager.
func NewStdLibManager() *StdLibManager {
	return &StdLibManager{}
}

// Sources searches the embedded standard library for the given package path.
func (m *StdLibManager) Sources(path string) ([]string, []string, error) {
	if srcs, ok := embeddedSources(path); ok {
		return srcs, nil, nil
	}
	return nil, nil, wrapError(path, ErrPkgNotFound)
}

// Asset returns ErrPkgNotFound because standard library packages do not contain binary assets.
func (m *StdLibManager) Asset(pkg, filename string) ([]byte, error) {
	return nil, wrapError(pkg, fmt.Errorf("asset %q: %w", filename, ErrPkgNotFound))
}

func embeddedSources(path string) ([]string, bool) {
	for _, lib := range gsllibs.All() {
		if lib.ImportPath == path {
			return lib.Sources, true
		}
	}
	return nil, false
}

