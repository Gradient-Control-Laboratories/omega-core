// Package gsllibs exposes the canonical GSL library packages — their
// import paths and source text — as compile-time embedded data. Tools that
// need to introspect, document, or include the libraries (e.g. the AI
// preamble in program/preamble) read them through this package so they don't
// need a source/manager and filesystem layout to do so.
//
// Package ordering is hand-curated rather than discovered by walking the
// filesystem so generated documentation in program/preamble always renders in
// a deterministic, intentional sequence. Individual .gsl files and pre-parsed
// package.pb artifacts inside each package directory are loaded via embed.FS.
package gsllibs

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/internal/stdpkg"
)

// Embed all top-level GSL library directories recursively. Go's embed.FS
// automatically includes nested subpackages (such as lattice/beam or
// material/bulk/metal) while excluding '_'-prefixed helper directories such as
// text/_tools/ so offline font generation assets do not bloat compiled binaries.
//
//go:embed builtin color ease lattice map material matrix mechanical noise text texture transform vector
var stdlibFS embed.FS

// curatedImportPaths defines the canonical order in which standard library
// packages appear in All() and generated documentation. Adding a library under
// an existing top-level category only requires adding its ImportPath here.
var curatedImportPaths = []string{
	"builtin",
	"map",
	"transform",
	"lattice",
	"lattice/beam",
	"lattice/honeycomb",
	"lattice/foam",
	"lattice/tpms",
	"lattice/spinodoid",
	"mechanical",
	"mechanical/catalog",
	"mechanical/gears",
	"mechanical/fastener",
	"mechanical/thread",
	"mechanical/drill",
	"mechanical/driver",
	"mechanical/joint",
	"text",
	"ease",
	"noise",
	"matrix",
	"vector",
	"color",
	"material",
	"material/bulk/metal",
	"material/bulk/alloy",
	"material/bulk/glass",
	"material/bulk/physical",
	"texture",
	"material/bulk/stone",
	"material/bulk/wood",
	"material/coat/weathering",
	"material/collection/playful",
	"material/bulk/natural",
	"material/collection/animated",
	"material/process/machined",
	"material/coat/coating",
	"material/process/printed",
	"material/coat/plated",
	"material/process/sheet",
	"material/process/cast",
	"material/process/molded",
	"material/bulk/composite",
	"material/bulk/masonry",
	"material/bulk/paper",
	"material/bulk/foams",
	"material/bulk/ceramic",
	"material/bulk/glazing",
	"material/collection/classic",
	"material/bulk/organic",
}

func init() {
	// Initialize stdpkg.StdLibRegistry with stdlibFS and curatedImportPaths
	// without reading or hashing any embedded files during package initialization.
	stdpkg.StdLibRegistry = stdpkg.NewRegistry(stdlibFS, curatedImportPaths)
}

// PkgInfo is one canonical library package: the import path users write
// in `import . "<path>";` and the source text of that package's .gsl.
//
// Domain marks a product-specific library: it is empty for the general
// libraries and set for product-specific libraries that should only surface
// in that product's documentation. See LibrariesDescriptiveText.
type PkgInfo struct {
	ImportPath string
	Sources    []string
	Domain     string
}

// All returns every embedded library in the order they should appear in
// generated documentation. Don't depend on map iteration order for output
// stability.
func All() []PkgInfo {
	loaded := stdpkg.StdLibRegistry.All()
	out := make([]PkgInfo, len(loaded))
	for i, pd := range loaded {
		srcs := make([]string, len(pd.Sources))
		copy(srcs, pd.Sources)
		out[i] = PkgInfo{
			ImportPath: pd.ImportPath,
			Sources:    srcs,
		}
	}
	return out
}

// pathByPrefix inverts the import statement: it maps the prefix a reference
// uses to the import path that prefix stands for ("beam" -> "lattice/beam").
// The assembler needs this to auto-import a library the source referenced but
// never imported, where all it has to go on is the prefix.
//
// A library's prefix is the last segment of its path, matching the prefix the
// parser derives for an unaliased `import "lattice/beam";`. That makes the
// last segment the global identity of a library: two libraries may not share
// one, however deeply their paths differ. Registering a duplicate panics here
// rather than binding the prefix to whichever library happens to be first in
// curatedImportPaths — an explicit import is the escape hatch for a name that
// must collide.
//
// builtin is excluded: it is implicitly dot-imported, so it is never named by
// a prefix.
var pathByPrefix = func() map[string]string {
	infos := make([]PkgInfo, len(curatedImportPaths))
	for i, imp := range curatedImportPaths {
		infos[i] = PkgInfo{ImportPath: imp}
	}
	byPrefix, err := indexByPrefix(infos)
	if err != nil {
		panic("gsllibs: " + err.Error())
	}
	return byPrefix
}()

// indexByPrefix builds the prefix index for libs, or reports the collision
// that makes one impossible. Split out from pathByPrefix so the collision is
// reachable from a test — the package-level index can only fail once, at init.
func indexByPrefix(libs []PkgInfo) (map[string]string, error) {
	byPrefix := make(map[string]string, len(libs))
	for _, lib := range libs {
		if lib.ImportPath == "builtin" {
			continue
		}
		prefix := lib.ImportPath
		if i := strings.LastIndex(prefix, "/"); i >= 0 {
			prefix = prefix[i+1:]
		}
		if first, dup := byPrefix[prefix]; dup {
			return nil, fmt.Errorf("%q and %q both claim the reference prefix %q; "+
				"auto-import cannot choose between them", first, lib.ImportPath, prefix)
		}
		byPrefix[prefix] = lib.ImportPath
	}
	return byPrefix, nil
}

// PathForPrefix returns the import path of the standard library that `prefix`
// refers to, and whether there is one. A prefix with no entry is not a
// standard library and needs an explicit import statement to resolve.
func PathForPrefix(prefix string) (string, bool) {
	path, ok := pathByPrefix[prefix]
	return path, ok
}

// Prefixes returns every prefix that auto-import resolves, sorted. Callers
// documenting the auto-imported set read it from here so the documentation
// cannot drift from what the assembler actually resolves.
func Prefixes() []string {
	prefixes := make([]string, 0, len(pathByPrefix))
	for prefix := range pathByPrefix {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)
	return prefixes
}
