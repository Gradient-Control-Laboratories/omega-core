package parser

import (
	"gradientcontrol.com/omega/core/internal/stdpkg"
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
)

// Result holds the parsed package.
//
// Invariants:
//   - Result and Result.Package are never nil.
//   - Result.Package.Files always contains at least one non-nil *program.PackageFile
//     (len(res.Package.Files) >= 1 and res.Package.Files[0] != nil).
type Result struct {
	Package *program.Package
}

// ParsePackage parses all source files in a package into a Result without caching
// user packages, while resolving unmodified standard library packages directly
// from the embedded pre-parsed protobuf cache in package program.
//
// Invariants:
//   - The returned *Result and Result.Package are never nil, and Result.Package
//     always contains at least one non-nil *program.PackageFile
//     (len(res.Package.Files) >= 1, res.Package.Files[0] != nil).
//   - When a parse or build error occurs, ParsePackage returns the partially
//     populated Result alongside a non-nil error.
func ParsePackage(sources []string) (*Result, error) {
	if stdpkg.StdLibRegistry != nil {
		if importPath, ok := stdpkg.StdLibRegistry.ImportPathByHash(stdpkg.SourcesHash(sources)); ok {
			return &Result{Package: program.GetStdPackage(importPath)}, nil
		}
	}
	pkg, err := parser2.Parse(sources)
	return &Result{Package: pkg}, err
}
