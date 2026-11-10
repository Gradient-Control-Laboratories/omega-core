package parser

import (
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

func init() {
	for _, libInfo := range gsllibs.All() {
		res, err := ParsePackage(libInfo.Sources)
		if err == nil && res != nil && res.Package != nil {
			program.RegisterStdPackage(libInfo.ImportPath, res.Package)
			stdCachedPkgs = append(stdCachedPkgs, stdCachedPkg{
				key:  sourcesKey(libInfo.Sources),
				pkg:  res.Package,
				size: sourcesSize(libInfo.Sources),
			})
		}
	}
}

// Result holds the parsed package.
//
// Invariants:
//   - Result and Result.Package are never nil.
//   - Result.Package.Files always contains at least one non-nil *program.PackageFile
//     (len(res.Package.Files) >= 1 and res.Package.Files[0] != nil).
type Result struct {
	Package *program.Package
}

// ParsePackage parses all source files in a package into a Result without caching.
//
// Invariants:
//   - The returned *Result and Result.Package are never nil, and Result.Package
//     always contains at least one non-nil *program.PackageFile
//     (len(res.Package.Files) >= 1, res.Package.Files[0] != nil).
//   - When a parse or build error occurs, ParsePackage returns the partially
//     populated Result alongside a non-nil error.
func ParsePackage(sources []string) (*Result, error) {
	res, err := parser2.Parse(sources)
	return &Result{Package: res.Package}, err
}
