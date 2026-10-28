package source

import (
	"fmt"

	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

type MockManager struct {
	PkgSources map[string][]string
	AssetNames map[string][]string
	Assets     map[string]map[string][]byte
}

var _ Manager = &MockManager{}

func (m *MockManager) Sources(path string) ([]string, []string, error) {
	srcs, ok := m.PkgSources[path]
	if ok {
		return srcs, m.AssetNames[path], nil
	}

	if srcs, ok := embeddedSources(path); ok {
		return srcs, nil, nil
	}

	return nil, nil, wrapError(path, ErrPkgNotFound)
}

func (m *MockManager) Asset(pkg, filename string) ([]byte, error) {
	if m.Assets != nil {
		if byPkg, ok := m.Assets[pkg]; ok {
			if data, ok := byPkg[filename]; ok {
				return data, nil
			}
		}
	}
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
