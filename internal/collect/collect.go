package collect

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// Program collects all packages starting from the primary package.
// It reports an error if a package cannot be found, cannot be parsed,
// or cyclic package imports are detected.
//
// It returns a program composed of all packages
// directly or indirectly referenced by the primarySources GSL code.
func Program(primarySources []string, srcMgr source.Manager, parsePkg func(sources []string) (*parser.Result, error)) (*program.Program, error) {
	if len(primarySources) == 0 {
		return nil, fmt.Errorf("missing primary sources")
	}

	primaryRes, err := parsePkg(primarySources)
	if err != nil {
		return nil, err
	}
	if primaryRes == nil || primaryRes.Package == nil {
		return nil, fmt.Errorf("no primary package returned")
	}
	primaryPkg := primaryRes.Package
	setPackageSources(primaryPkg, primarySources)
	return Package(primaryPkg, srcMgr, parsePkg)
}

func setPackageSources(pkg *program.Package, sources []string) {
	if pkg == nil {
		return
	}
	for i := 0; i < len(pkg.Files) && i < len(sources); i++ {
		if pkg.Files[i] != nil && pkg.Files[i].Source == "" {
			pkg.Files[i].Source = sources[i]
		}
	}
}

// Package collects all packages starting from a pre-parsed primary package.
func Package(primaryPkg *program.Package, srcMgr source.Manager, parsePkg func(sources []string) (*parser.Result, error)) (*program.Program, error) {
	if primaryPkg == nil {
		return nil, fmt.Errorf("missing primary package")
	}
	program.UpdateImports(primaryPkg, primaryPkg.Path)

	if len(primaryPkg.Files) == 0 {
		return nil, fmt.Errorf("primary package must have at least one file")
	}

	var collected []*program.Package
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var dfs func(p *program.Package) error
	dfs = func(p *program.Package) error {
		path := p.Path
		visiting[path] = true

		for _, file := range p.Files {
			for _, imp := range file.Imports {
				impPath := imp.Path
				if impPath == path {
					return fmt.Errorf("package %s cannot import itself", path)
				}
				if visiting[impPath] {
					return fmt.Errorf("cyclic package import detected: %s imports %s", path, impPath)
				}
				if !visited[impPath] {
					sources, _, err := srcMgr.Sources(impPath)
					if err != nil {
						return fmt.Errorf("failed to load package %s: %w", impPath, err)
					}
					depRes, err := parsePkg(sources)
					if err != nil {
						return fmt.Errorf("failed to parse package %s: %w", impPath, err)
					}
					if depRes == nil || depRes.Package == nil {
						return fmt.Errorf("package not found: %s", impPath)
					}
					depPkg := depRes.Package
					setPackageSources(depPkg, sources)
					program.UpdateImports(depPkg, impPath)
					if depPkg.Path == "" {
						depPkg.Path = impPath
					}
					if err := dfs(depPkg); err != nil {
						return err
					}
				}
			}
		}

		delete(visiting, path)
		visited[path] = true
		if path != primaryPkg.Path {
			collected = append(collected, p)
		}
		return nil
	}

	if err := dfs(primaryPkg); err != nil {
		return nil, err
	}

	// Reverse the collected packages slice to ensure that importing packages
	// appear earlier in the dependencies list than the packages they import.
	for i, j := 0, len(collected)-1; i < j; i, j = i+1, j-1 {
		collected[i], collected[j] = collected[j], collected[i]
	}

	return &program.Program{
		VersionMajor: primaryPkg.VersionMajor,
		VersionMinor: primaryPkg.VersionMinor,
		Primary:      primaryPkg,
		Dependencies: collected,
	}, nil
}
