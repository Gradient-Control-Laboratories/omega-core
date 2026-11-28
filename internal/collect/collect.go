package collect

import (
	"fmt"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler/source"
	"github.com/Gradient-Control-Laboratories/omega-core/parser"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
	gsllibs "github.com/Gradient-Control-Laboratories/omega-core/stdlib"
)

// Program collects all packages starting from the primary package
// along with any non-fatal parser diagnostics, tagging diagnostics from imported
// packages with their import path and source file path.
// It reports an error if a package cannot be found, cannot be parsed,
// or cyclic package imports are detected.
func Program(primarySources []string, srcMgr source.Manager, parsePkg func(sources []string) (*parser.Result, error)) (*program.Program, []program.Diagnostic, error) {
	if len(primarySources) == 0 {
		return nil, nil, fmt.Errorf("missing primary sources")
	}

	primaryRes, err := parsePkg(primarySources)
	if err != nil {
		return nil, nil, err
	}
	if primaryRes == nil || primaryRes.Package == nil {
		return nil, nil, fmt.Errorf("no primary package returned")
	}
	primaryPkg := primaryRes.Package
	setPackageSources(primaryPkg, primarySources)
	var diags []program.Diagnostic
	if len(primaryRes.Diagnostics) > 0 {
		diags = append(diags, primaryRes.Diagnostics...)
	}
	prog, depDiags, err := Package(primaryPkg, srcMgr, parsePkg)
	if err != nil {
		return nil, nil, err
	}
	if len(depDiags) > 0 {
		diags = append(diags, depDiags...)
	}
	return prog, diags, nil
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

// Package collects all packages starting from a pre-parsed primary package
// along with any non-fatal parser diagnostics, tagging diagnostics from imported
// packages with their import path and source file path.
func Package(primaryPkg *program.Package, srcMgr source.Manager, parsePkg func(sources []string) (*parser.Result, error)) (*program.Program, []program.Diagnostic, error) {
	if primaryPkg == nil {
		return nil, nil, fmt.Errorf("missing primary package")
	}
	program.UpdateImports(primaryPkg, primaryPkg.Path)

	if len(primaryPkg.Files) == 0 {
		return nil, nil, fmt.Errorf("primary package must have at least one file")
	}

	var collected []*program.Package
	var diags []program.Diagnostic
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
						if renamed, ok := gsllibs.RenamedPath(impPath); ok {
							return fmt.Errorf("failed to load package %s: the package was renamed %s, and `gsl fix` rewrites the import: %w", impPath, renamed, err)
						}
						return fmt.Errorf("failed to load package %s: %w", impPath, err)
					}
					depRes, err := parsePkg(sources)
					if err != nil {
						return fmt.Errorf("failed to parse package %s: %w", impPath, err)
					}
					if depRes == nil || depRes.Package == nil {
						return fmt.Errorf("package not found: %s", impPath)
					}
					for _, d := range depRes.Diagnostics {
						if d.PackagePath == "" {
							d.PackagePath = impPath
						}
						diags = append(diags, d)
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
		return nil, nil, err
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
	}, diags, nil
}
