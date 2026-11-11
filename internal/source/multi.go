package source

import (
	"errors"
	"fmt"

	"gradientcontrol.com/omega/core/assembler/source"
)

// MultiSourceManager delegates package source and asset lookups across an ordered list of Managers.
type MultiSourceManager struct {
	// Read-only immutable fields set at construction
	managers []source.Manager
}

var _ source.Manager = &MultiSourceManager{}

// NewMultiSourceManager creates a Manager that queries each non-nil Manager in order.
func NewMultiSourceManager(managers ...source.Manager) *MultiSourceManager {
	var filtered []source.Manager
	for _, mgr := range managers {
		if mgr != nil {
			filtered = append(filtered, mgr)
		}
	}
	return &MultiSourceManager{managers: filtered}
}

// Sources queries each Manager in order until a package is found or a non-not-found error occurs.
func (m *MultiSourceManager) Sources(path string) ([]string, []string, error) {
	for _, mgr := range m.managers {
		srcs, assets, err := mgr.Sources(path)
		if err != nil {
			// TODO this is inefficient... change Sources to return nil rather than error when package not found
			if errors.Is(err, source.ErrPkgNotFound) {
				continue
			}
			return nil, nil, err
		}
		if len(srcs) > 0 {
			return srcs, assets, nil
		}
	}
	return nil, nil, wrapError(path, source.ErrPkgNotFound)
}

// Asset queries each Manager in order until the asset is found or a non-not-found error occurs.
func (m *MultiSourceManager) Asset(pkg, filename string) ([]byte, error) {
	for _, mgr := range m.managers {
		data, err := mgr.Asset(pkg, filename)
		if err != nil {
			// TODO this is inefficient... change Asset to return nil rather than error when package not found
			if errors.Is(err, source.ErrPkgNotFound) {
				continue
			}
			return nil, err
		}
		if data != nil {
			return data, nil
		}
	}
	return nil, wrapError(pkg, fmt.Errorf("asset %q: %w", filename, source.ErrPkgNotFound))
}
