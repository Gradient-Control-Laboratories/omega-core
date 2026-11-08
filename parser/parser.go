package parser

import (
	"gradientcontrol.com/omega/core/parser/internal/lib"
	"gradientcontrol.com/omega/core/program"
)

// Result holds the parsed package and any structured parse errors.
//
// Invariants:
//   - Result and Result.Package are never nil.
//   - Result.Package.Files always contains at least one non-nil *program.PackageFile
//     (len(res.Package.Files) >= 1 and res.Package.Files[0] != nil).
type Result = program.ParseResult

// ParsePackage parses all source files in a package into a Result without caching.
//
// Invariants:
//   - The returned *Result and Result.Package are never nil, and Result.Package
//     always contains at least one non-nil *program.PackageFile
//     (len(res.Package.Files) >= 1, res.Package.Files[0] != nil).
//   - When a parse or build error occurs, ParsePackage returns the partially
//     populated Result.Package and the structured Result.Errors.
func ParsePackage(sources []string) *Result {
	return lib.ParsePackage(sources)
}

var _ program.ParsePackageFunc = ParsePackage

// CachingParser wraps ParsePackage with an in-memory LRU cache keyed by source
// content hash. It hands out clones of cached packages and never its own copy,
// so callers may rewrite the returned *program.Package in place.
type CachingParser = lib.CachingParser

// NewCachingParser creates a new CachingParser pre-populated with standard library packages.
func NewCachingParser() *CachingParser {
	return lib.NewCachingParser()
}
