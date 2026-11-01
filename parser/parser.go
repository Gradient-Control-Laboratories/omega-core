package parser

import (
	"gradientcontrol.com/omega/core/parser/internal/lib"
	"gradientcontrol.com/omega/core/program"
)

// ParsePackage parses all source files in a package into a program.Package.
//
// Invariants:
//   - The returned *program.Package is never nil and always contains at least
//     one non-nil *program.PackageFile (len(pkg.Files) >= 1, pkg.Files[0] != nil).
//   - When a parse or build error occurs, ParsePackage returns the partially
//     populated *program.Package alongside the error without caching it.
func ParsePackage(sources []string) (*program.Package, error) {
	return lib.ParsePackage(sources)
}

var _ program.ParsePackageFunc = ParsePackage
