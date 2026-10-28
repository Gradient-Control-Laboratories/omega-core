package source

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ResolveRootPathList resolves the root package paths from the given value.
func ResolveRootPathList(value, baseDir string, required bool) ([]string, error) {
	var roots []string
	for _, root := range filepath.SplitList(value) {
		if root != "" {
			roots = append(roots, root)
		}
	}
	if len(roots) == 0 {
		if required {
			return nil, errors.New("missing")
		}
		return nil, nil
	}
	for i, root := range roots {
		if baseDir != "" && !filepath.IsAbs(root) {
			root = filepath.Join(baseDir, root)
		}
		abs, err := filepath.Abs(root)
		if err != nil {
			return nil, fmt.Errorf("failed to make absolute: %w", err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %q: %w", abs, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("expected directory %q", abs)
		}
		roots[i] = abs
	}
	return roots, nil
}
