package collect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// Program collects all packages starting from the primary package.
// It reports an error if a package cannot be found, cannot be parsed,
// or cyclic package imports are detected.
//
// It returns a program composed of all packages
// directly or indirectly referenced by the primarySources GSL code.
func Program(primarySources []string, srcMgr source.Manager, parse program.ParsePackageFunc) (*program.Program, error) {
	if len(primarySources) == 0 {
		return nil, fmt.Errorf("missing primary sources")
	}

	primaryRes := parse(primarySources)
	if err := primaryRes.Error(); err != nil {
		return nil, err
	}
	if primaryRes == nil || primaryRes.Package == nil {
		return nil, fmt.Errorf("no primary package returned")
	}
	primaryPkg := primaryRes.Package
	setPackageSources(primaryPkg, primarySources)
	return Package(primaryPkg, srcMgr, parse)
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
func Package(primaryPkg *program.Package, srcMgr source.Manager, parse program.ParsePackageFunc) (*program.Program, error) {
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
					depRes := parse(sources)
					if err := depRes.Error(); err != nil {
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

// ProgramSourceHash returns the deterministic SHA-256 hex digest of all GSL
// package sources in prog (primary and transitive dependencies), preserving
// each package's declared source file order.
func ProgramSourceHash(prog *program.Program) string {
	if prog == nil || prog.Primary == nil {
		return ""
	}
	sources := make(map[string][]string, len(prog.Dependencies)+1)
	// Invariant: prog.Primary and prog.Dependencies entries are guaranteed non-nil.
	addPkg := func(pkg *program.Package) {
		// Invariant: pkg.Files elements are guaranteed non-nil by the parser.
		fileSources := make([]string, 0, len(pkg.Files))
		for _, f := range pkg.Files {
			if f.Source != "" {
				fileSources = append(fileSources, f.Source)
			}
		}
		if len(fileSources) == 0 {
			return
		}
		sources[pkg.Path] = fileSources
	}
	addPkg(prog.Primary)
	for _, dep := range prog.Dependencies {
		addPkg(dep)
	}
	// JSON sorts package keys while preserving each package's declared source order.
	data, err := json.Marshal(sources)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256([]byte("gsl-dependencies-v1\n" + string(data)))
	return hex.EncodeToString(sum[:])
}

// SourceHash collects primaryPkgPath and all its transitive GSL imports from
// srcMgr and returns their deterministic SHA-256 hex digest.
func SourceHash(primaryPkgPath string, srcMgr source.Manager, parse program.ParsePackageFunc) (string, error) {
	primary, _, err := srcMgr.Sources(primaryPkgPath)
	if err != nil {
		return "", err
	}
	prog, err := Program(primary, srcMgr, parse)
	if err != nil {
		return "", err
	}
	prog.Primary.Path = primaryPkgPath
	return ProgramSourceHash(prog), nil
}
