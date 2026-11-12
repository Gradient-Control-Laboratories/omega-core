package source

import (
	"fmt"

	"gradientcontrol.com/omega/core/assembler/source"
	"gradientcontrol.com/omega/core/internal/stdpkg"
	_ "gradientcontrol.com/omega/core/stdlib"
)

// StdLibManager provides access to embedded standard library packages.
type StdLibManager struct{}

var _ source.Manager = &StdLibManager{}

// NewStdLibManager creates a new instance of the embedded standard library source manager.
func NewStdLibManager() *StdLibManager {
	return &StdLibManager{}
}

// Sources searches the embedded standard library for the given package path.
func (m *StdLibManager) Sources(path string) ([]string, []string, error) {
	if srcs, ok := embeddedSources(path); ok {
		return srcs, nil, nil
	}
	return nil, nil, wrapError(path, source.ErrPkgNotFound)
}

// Asset returns ErrPkgNotFound because standard library packages do not contain binary assets.
func (m *StdLibManager) Asset(pkg, filename string) ([]byte, error) {
	return nil, wrapError(pkg, fmt.Errorf("asset %q: %w", filename, source.ErrPkgNotFound))
}

func embeddedSources(path string) ([]string, bool) {
	srcs, ok := stdpkg.StdLibRegistry.Sources(path)
	if !ok {
		return nil, false
	}
	copied := make([]string, len(srcs))
	copy(copied, srcs)
	return copied, true
}
