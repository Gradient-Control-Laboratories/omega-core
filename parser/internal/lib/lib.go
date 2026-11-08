package lib

import (
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

func init() {
	for _, libInfo := range gsllibs.All() {
		res := ParsePackage(libInfo.Sources)
		if res.Error() == nil && res != nil && res.Package != nil {
			program.RegisterStdPackage(libInfo.ImportPath, res.Package)
			registerStdCachedPkg(libInfo.Sources, res.Package)
		}
	}
}

// Result holds the parsed package and any structured parse errors.
//
// Invariants:
//   - Result and Result.Package are never nil.
//   - Result.Package.Files always contains at least one non-nil *program.PackageFile
//     (len(res.Package.Files) >= 1 and res.Package.Files[0] != nil).
type Result = program.ParseResult

// ParsePackage parses all of the sources in a package without caching.
//
// Invariants:
//   - The returned *Result and Result.Package are never nil, and Result.Package
//     always contains at least one non-nil *program.PackageFile
//     (len(res.Package.Files) >= 1, res.Package.Files[0] != nil).
//   - When a parse or build error occurs, ParsePackage returns the partially
//     populated Result.Package and the structured Result.Errors.
func ParsePackage(sources []string) *Result {
	return parser2.Parse(sources)
}
