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

// curatedEntries defines the canonical order in which standard library packages
// appear in All() and generated documentation. Adding a library under an
// existing top-level category only requires adding its ImportPath here.
var curatedEntries = []stdpkg.Entry{
	{ImportPath: "builtin"},
	{ImportPath: "map"},
	{ImportPath: "transform"},
	{ImportPath: "lattice"},
	{ImportPath: "lattice/beam"},
	{ImportPath: "lattice/honeycomb"},
	{ImportPath: "lattice/foam"},
	{ImportPath: "lattice/tpms"},
	{ImportPath: "lattice/spinodoid"},
	{ImportPath: "mechanical"},
	{ImportPath: "mechanical/catalog"},
	{ImportPath: "mechanical/gears"},
	{ImportPath: "mechanical/fastener"},
	{ImportPath: "mechanical/thread"},
	{ImportPath: "mechanical/drill"},
	{ImportPath: "mechanical/driver"},
	{ImportPath: "mechanical/joint"},
	{ImportPath: "text"},
	{ImportPath: "ease"},
	{ImportPath: "noise"},
	{ImportPath: "matrix"},
	{ImportPath: "vector"},
	{ImportPath: "color"},
	{ImportPath: "material"},
	{ImportPath: "material/bulk/metal"},
	{ImportPath: "material/bulk/alloy"},
	{ImportPath: "material/bulk/glass"},
	{ImportPath: "material/bulk/physical"},
	{ImportPath: "texture"},
	{ImportPath: "material/bulk/stone"},
	{ImportPath: "material/bulk/wood"},
	{ImportPath: "material/coat/weathering"},
	{ImportPath: "material/collection/playful"},
	{ImportPath: "material/bulk/natural"},
	{ImportPath: "material/collection/animated"},
	{ImportPath: "material/process/machined"},
	{ImportPath: "material/coat/coating"},
	{ImportPath: "material/process/printed"},
	{ImportPath: "material/coat/plated"},
	{ImportPath: "material/process/sheet"},
	{ImportPath: "material/process/cast"},
	{ImportPath: "material/process/molded"},
	{ImportPath: "material/bulk/composite"},
	{ImportPath: "material/bulk/masonry"},
	{ImportPath: "material/bulk/paper"},
	{ImportPath: "material/bulk/foams"},
	{ImportPath: "material/bulk/ceramic"},
	{ImportPath: "material/bulk/glazing"},
	{ImportPath: "material/collection/classic"},
	{ImportPath: "material/bulk/organic"},
}

func init() {
	// Register stdlibFS and curatedEntries with internal/stdpkg without reading
	// or hashing any embedded files during package initialization.
	stdpkg.RegisterFS(stdlibFS, curatedEntries)
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
	loaded := stdpkg.All()
	out := make([]PkgInfo, len(loaded))
	for i, pd := range loaded {
		srcs := make([]string, len(pd.Sources))
		copy(srcs, pd.Sources)
		out[i] = PkgInfo{
			ImportPath: pd.ImportPath,
			Sources:    srcs,
			Domain:     pd.Domain,
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
// curatedEntries — an explicit import is the escape hatch for a name that must
// collide.
//
// builtin is excluded: it is implicitly dot-imported, so it is never named by
// a prefix.
var pathByPrefix = func() map[string]string {
	infos := make([]PkgInfo, len(curatedEntries))
	for i, e := range curatedEntries {
		infos[i] = PkgInfo{ImportPath: e.ImportPath, Domain: e.Domain}
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
