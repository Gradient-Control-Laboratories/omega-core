package program

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// GetImportFunc resolves an import path to its package. It mirrors
// build.GetImportFunc without taking a dependency on the build package.
type GetImportFunc func(path string) (*Package, error)

var stdPkgs = make(map[string]*Package)

// RegisterStdPackage registers a pre-parsed standard GSL package.
func RegisterStdPackage(path string, pkg *Package) {
	stdPkgs[path] = pkg
}

// GetStdPackage returns a copy of a registered standard package, or nil if not registered.
func GetStdPackage(path string) *Package {
	if pkg, ok := stdPkgs[path]; ok {
		return proto.Clone(pkg).(*Package)
	}
	return nil
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
	if stdPkg, ok := stdPkgs[path]; ok {
		return stdPkg, nil
	}
	return nil, fmt.Errorf("package %q not found in dependencies", path)
}
