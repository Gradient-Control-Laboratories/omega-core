// Package gsllibs exposes the canonical GSL library packages — their
// import paths and source text — as compile-time embedded data. Tools that
// need to introspect, document, or include the libraries (e.g. the AI
// preamble in common/program/preamble) read them through this package so
// they don't need a source/manager and filesystem layout to do so.
//
// The set is hand-curated rather than discovered — adding a new library
// means adding a //go:embed line and an entry to All(). That's the
// trade-off for not needing a filesystem at runtime.
//
// The directory holding this file is named "package" (it's the root of
// the GSL package tree), which would shadow Go's `package` keyword as a
// directory-derived name, so the Go package name is gsllibs instead.
package gsllibs

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed builtin/base.gsl
var baseSrc string

//go:embed builtin/booleans.gsl
var booleansSrc string

//go:embed builtin/util.gsl
var utilSrc string

//go:embed builtin/offsets.gsl
var offsetsSrc string

//go:embed builtin/2d.gsl
var _2dSrc string

//go:embed builtin/3d.gsl
var _3dSrc string

//go:embed builtin/quad_bezier_curves.gsl
var quadBezierCurvesSrc string

//go:embed builtin/sweep.gsl
var sweepSrc string

//go:embed map/map.gsl
var mapSrc string

//go:embed transform/transform.gsl
var transformSrc string

//go:embed lattice/lattice.gsl
var latticeSrc string

//go:embed lattice/beam/beam.gsl
var latticeBeamSrc string

//go:embed lattice/honeycomb/honeycomb.gsl
var latticeHoneycombSrc string

//go:embed lattice/foam/foam.gsl
var latticeFoamSrc string

//go:embed lattice/tpms/tpms.gsl
var latticeTpmsSrc string

//go:embed lattice/spinodoid/spinodoid.gsl
var latticeSpinodoidSrc string

//go:embed mechanical/mechanical.gsl
var mechanicalSrc string

//go:embed mechanical/catalog/catalog.gsl
var catalogSrc string

//go:embed mechanical/gears/gears.gsl
var gearsSrc string

//go:embed mechanical/fastener/fastener.gsl
var fastenerSrc string

//go:embed mechanical/drill/drill.gsl
var drillSrc string

//go:embed mechanical/thread/thread.gsl
var threadSrc string

//go:embed mechanical/driver/driver.gsl
var driverSrc string

//go:embed mechanical/joint/joint.gsl
var jointSrc string

//go:embed text/text.gsl
var textSrc string

//go:embed ease/ease.gsl
var easeSrc string

//go:embed noise/noise.gsl
var noiseSrc string

//go:embed matrix/matrix.gsl
var matrixSrc string

//go:embed vector/vector.gsl
var vectorSrc string

//go:embed color/color.gsl
var colorSrc string

//go:embed material/material.gsl
var materialSrc string

//go:embed metal/metal.gsl
var metalSrc string

//go:embed glass/glass.gsl
var glassSrc string

//go:embed texture/texture.gsl
var textureSrc string

//go:embed material/stone/stone.gsl
var textureStoneSrc string

//go:embed material/wood/wood.gsl
var textureWoodSrc string

//go:embed material/weathering/weathering.gsl
var textureWeatheringSrc string

//go:embed material/playful/playful.gsl
var texturePlayfulSrc string

//go:embed material/natural/natural.gsl
var textureNaturalSrc string

//go:embed material/animated/animated.gsl
var textureAnimatedSrc string

//go:embed material/machined/machined.gsl
var materialMachinedSrc string

//go:embed material/coating/coating.gsl
var materialCoatingSrc string

//go:embed material/printed/printed.gsl
var materialPrintedSrc string

//go:embed physical/physical.gsl
var physicalSrc string

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
	return []PkgInfo{
		{ImportPath: "builtin", Sources: []string{
			baseSrc,
			booleansSrc,
			utilSrc,
			offsetsSrc,
			_2dSrc,
			_3dSrc,
			quadBezierCurvesSrc,
			sweepSrc,
		}},
		{ImportPath: "map", Sources: []string{mapSrc}},
		{ImportPath: "transform", Sources: []string{transformSrc}},
		{ImportPath: "lattice", Sources: []string{latticeSrc}},
		{ImportPath: "lattice/beam", Sources: []string{latticeBeamSrc}},
		{ImportPath: "lattice/honeycomb", Sources: []string{latticeHoneycombSrc}},
		{ImportPath: "lattice/foam", Sources: []string{latticeFoamSrc}},
		{ImportPath: "lattice/tpms", Sources: []string{latticeTpmsSrc}},
		{ImportPath: "lattice/spinodoid", Sources: []string{latticeSpinodoidSrc}},
		{ImportPath: "mechanical", Sources: []string{mechanicalSrc}},
		{ImportPath: "mechanical/catalog", Sources: []string{catalogSrc}},
		{ImportPath: "mechanical/gears", Sources: []string{gearsSrc}},
		{ImportPath: "mechanical/fastener", Sources: []string{fastenerSrc}},
		{ImportPath: "mechanical/thread", Sources: []string{threadSrc}},
		{ImportPath: "mechanical/drill", Sources: []string{drillSrc}},
		{ImportPath: "mechanical/driver", Sources: []string{driverSrc}},
		{ImportPath: "mechanical/joint", Sources: []string{jointSrc}},
		{ImportPath: "text", Sources: []string{textSrc}},
		{ImportPath: "ease", Sources: []string{easeSrc}},
		{ImportPath: "noise", Sources: []string{noiseSrc}},
		{ImportPath: "matrix", Sources: []string{matrixSrc}},
		{ImportPath: "vector", Sources: []string{vectorSrc}},
		{ImportPath: "color", Sources: []string{colorSrc}},
		{ImportPath: "material", Sources: []string{materialSrc}},
		{ImportPath: "metal", Sources: []string{metalSrc}},
		{ImportPath: "glass", Sources: []string{glassSrc}},
		{ImportPath: "physical", Sources: []string{physicalSrc}},
		{ImportPath: "texture", Sources: []string{textureSrc}},
		{ImportPath: "material/stone", Sources: []string{textureStoneSrc}},
		{ImportPath: "material/wood", Sources: []string{textureWoodSrc}},
		{ImportPath: "material/weathering", Sources: []string{textureWeatheringSrc}},
		{ImportPath: "material/playful", Sources: []string{texturePlayfulSrc}},
		{ImportPath: "material/natural", Sources: []string{textureNaturalSrc}},
		{ImportPath: "material/animated", Sources: []string{textureAnimatedSrc}},
		{ImportPath: "material/machined", Sources: []string{materialMachinedSrc}},
		{ImportPath: "material/coating", Sources: []string{materialCoatingSrc}},
		{ImportPath: "material/printed", Sources: []string{materialPrintedSrc}},
	}
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
// All() — an explicit import is the escape hatch for a name that must collide.
//
// builtin is excluded: it is implicitly dot-imported, so it is never named by
// a prefix.
var pathByPrefix = func() map[string]string {
	byPrefix, err := indexByPrefix(All())
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
