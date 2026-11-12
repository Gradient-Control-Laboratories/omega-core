package source

import (
	"fmt"

	"gradientcontrol.com/omega/core/assembler/source"
)

type MockManager struct {
	PkgSources map[string][]string
	AssetNames map[string][]string
	Assets     map[string]map[string][]byte
}

var _ source.Manager = &MockManager{}

func (m *MockManager) Sources(path string) ([]string, []string, error) {
	srcs, ok := m.PkgSources[path]
	if ok {
		return srcs, m.AssetNames[path], nil
	}

	if srcs, ok := embeddedSources(path); ok {
		return srcs, nil, nil
	}

	return nil, nil, wrapError(path, source.ErrPkgNotFound)
}

func (m *MockManager) Asset(pkg, filename string) ([]byte, error) {
	if m.Assets != nil {
		if byPkg, ok := m.Assets[pkg]; ok {
			if data, ok := byPkg[filename]; ok {
				return data, nil
			}
		}
	}
	return nil, wrapError(pkg, fmt.Errorf("asset %q: %w", filename, source.ErrPkgNotFound))
}
