package program

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/stdpkg"
	_ "gradientcontrol.com/omega/core/stdlib"
)

// GetImportFunc resolves an import path to its package. It mirrors
// build.GetImportFunc without taking a dependency on the build package.
type GetImportFunc func(path string) (*Package, error)

// stdLoaders builds a read-only map of per-package sync.OnceValue unmarshalers
// on first lookup.
//
// Why lazy per-package unmarshaling shared with parser:
//   - Package init() performs zero I/O or protobuf decoding.
//   - Only the specific standard library packages actually referenced by a
//     compilation or test are unmarshaled, and each is unmarshaled at most once
//     per process without lock contention on subsequent lookups.
//   - (*Program).GetImport returns the shared read-only *Package directly,
//     whereas GetStdPackage (used by parser.ParsePackage and CachingParser)
//     hands out a proto.Clone copy because callers rewrite returned Package
//     ASTs in place.
var stdLoaders = sync.OnceValue(func() map[string]func() *Package {
	pkgs := stdpkg.StdLibRegistry.All()
	loaders := make(map[string]func() *Package, len(pkgs))
	for _, pd := range pkgs {
		importPath := pd.ImportPath
		loaders[importPath] = sync.OnceValue(func() *Package {
			data, ok := stdpkg.StdLibRegistry.Proto(importPath)
			if !ok {
				return nil
			}
			var pkg Package
			if err := proto.Unmarshal(data, &pkg); err != nil {
				panic(fmt.Sprintf("program: unmarshaling embedded stdlib %q: %v", importPath, err))
			}
			return &pkg
		})
	}
	return loaders
})

func stdPackage(path string) *Package {
	loader, ok := stdLoaders()[path]
	if !ok {
		return nil
	}
	return loader()
}

// GetStdPackage returns a cloned copy of an embedded standard GSL package, or
// nil if path is not a standard library package with a current package.pb.
func GetStdPackage(path string) *Package {
	pkg := stdPackage(path)
	if pkg == nil {
		return nil
	}
	return proto.Clone(pkg).(*Package)
}

// GetImport resolves an import path to its package from the program's dependencies.
func (prog *Program) GetImport(path string) (*Package, error) {
	if prog != nil {
		for _, dep := range prog.Dependencies {
			if dep != nil && dep.Path == path {
				return dep, nil
			}
		}
	}
	// TODO: Remove this when possible. Ideally a program is complete and contains
	// all referenced packages including system and auto-imported packages so no
	// external lookups should be necessary.
	if stdPkg := stdPackage(path); stdPkg != nil {
		return stdPkg, nil
	}
	return nil, fmt.Errorf("package %q not found in dependencies", path)
}
