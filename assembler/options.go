package assembler

import (
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/source"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/topofold"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/warpfold"
	internalsource "github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/program/progress"
)

// Option configures program assembly.
type Option func(*options)

type options struct {
	managers              []source.Manager
	parsePkg              ParsePackageFunc
	geometryResolver      topofold.GeometryResolver
	warpResolver          warpfold.WarpResolver
	foldWarpFieldToZero   bool
	enableSlotFold        bool
	enableDeadcodeRemoval bool
	extractUI             bool
	extractMeshPlan       bool
	meshPlanIdPrefix      string
	variants              map[string]string
	progressFn            progress.Callback
	analyticGradient      bool
}

// WithParser supplies a custom package parser (such as parser.CachingParser.ParsePackage)
// to use when parsing primary and imported packages during assembly.
func WithParser(parsePkg ParsePackageFunc) Option {
	return func(o *options) {
		o.parsePkg = parsePkg
	}
}

// Parser is WithParser.
//
// Deprecated: use WithParser; Parser will be removed after a release.
func Parser(parsePkg ParsePackageFunc) Option {
	return WithParser(parsePkg)
}

// WithSourcePaths adds local directory root paths used to resolve imported packages and assets.
func WithSourcePaths(paths ...string) Option {
	return func(o *options) {
		if len(paths) > 0 {
			o.managers = append(o.managers, internalsource.NewLocalManager(paths...))
		}
	}
}

// SourcePaths is WithSourcePaths.
//
// Deprecated: use WithSourcePaths; SourcePaths will be removed after a release.
func SourcePaths(paths ...string) Option {
	return WithSourcePaths(paths...)
}

// WithSourceManager adds a custom source manager used to resolve imported packages and assets.
func WithSourceManager(mgr source.Manager) Option {
	return func(o *options) {
		if mgr != nil {
			o.managers = append(o.managers, mgr)
		}
	}
}

// SourceManager is WithSourceManager.
//
// Deprecated: use WithSourceManager; SourceManager will be removed after a release.
func SourceManager(mgr source.Manager) Option {
	return WithSourceManager(mgr)
}

// WithGeometryResolver supplies a topology geometry resolver for folding STEP topology builtins.
func WithGeometryResolver(r topofold.GeometryResolver) Option {
	return func(o *options) {
		o.geometryResolver = r
	}
}

// GeometryResolver is WithGeometryResolver.
//
// Deprecated: use WithGeometryResolver; GeometryResolver will be removed after a release.
func GeometryResolver(r topofold.GeometryResolver) Option {
	return WithGeometryResolver(r)
}

// WithWarpResolver supplies a warp simulation resolver for baking and folding warped mesh assets.
func WithWarpResolver(r warpfold.WarpResolver) Option {
	return func(o *options) {
		o.warpResolver = r
	}
}

// WarpResolver is WithWarpResolver.
//
// Deprecated: use WithWarpResolver; WarpResolver will be removed after a release.
func WarpResolver(r warpfold.WarpResolver) Option {
	return WithWarpResolver(r)
}

// WithFoldWarpFieldToZero replaces every warpField(...) call across the program with [0, 0, 0]
// during Step 5 domain folding for backends that do not sample live warp textures.
func WithFoldWarpFieldToZero() Option {
	return func(o *options) {
		o.foldWarpFieldToZero = true
	}
}

// FoldWarpFieldToZero is WithFoldWarpFieldToZero.
//
// Deprecated: use WithFoldWarpFieldToZero; FoldWarpFieldToZero will be removed after a release.
func FoldWarpFieldToZero() Option {
	return WithFoldWarpFieldToZero()
}

// WithDisableMeshSlotFolding disables the automatic mesh slot assignment and folding pass.
func WithDisableMeshSlotFolding() Option {
	return func(o *options) {
		o.enableSlotFold = false
	}
}

// DisableMeshSlotFolding is WithDisableMeshSlotFolding.
//
// Deprecated: use WithDisableMeshSlotFolding; DisableMeshSlotFolding will be removed after a release.
func DisableMeshSlotFolding() Option {
	return WithDisableMeshSlotFolding()
}

// WithDisableDeadcodeRemoval disables dead code removal during assembly.
func WithDisableDeadcodeRemoval() Option {
	return func(o *options) {
		o.enableDeadcodeRemoval = false
	}
}

// DisableDeadcodeRemoval is WithDisableDeadcodeRemoval.
//
// Deprecated: use WithDisableDeadcodeRemoval; DisableDeadcodeRemoval will be removed after a release.
func DisableDeadcodeRemoval() Option {
	return WithDisableDeadcodeRemoval()
}

// WithExtractUI enables extraction of sliders, UI controls, and scene metadata into the Result.
func WithExtractUI() Option {
	return func(o *options) {
		o.extractUI = true
	}
}

// ExtractUI is WithExtractUI.
//
// Deprecated: use WithExtractUI; ExtractUI will be removed after a release.
func ExtractUI() Option {
	return WithExtractUI()
}

// WithExtractMeshPlan configures the assembler to partition geometry into a MeshExportPlan
// and analytical SdfSubtrees as Step 8. If idPrefix is empty, it defaults to "leaf-".
func WithExtractMeshPlan(idPrefix string) Option {
	return func(o *options) {
		o.extractMeshPlan = true
		o.meshPlanIdPrefix = idPrefix
	}
}

// ExtractMeshPlan is WithExtractMeshPlan.
//
// Deprecated: use WithExtractMeshPlan; ExtractMeshPlan will be removed after a release.
func ExtractMeshPlan(idPrefix string) Option {
	return WithExtractMeshPlan(idPrefix)
}

// WithVariants configures the assembler to compile and slice the named
// alternative top-level return statements (e.g. map[string]string{"part": "return mainPart();"})
// alongside the primary package in a single pass.
func WithVariants(variants map[string]string) Option {
	return func(o *options) {
		o.variants = variants
	}
}

// GradientNote reports a geometry return whose analytic gradient could not be
// lowered (an op or builtin with no derivative rule). Its gradient is left nil,
// and a renderer falls back to central differences for hits on that return.
type GradientNote struct {
	Return int    // 0-based index into the geometry returns
	Reason string // why the derivative failed
}

// WithAnalyticGradient lowers the spatial gradient d/dp of each geometry return
// alongside its distance (Program.GradientReturns), for a backend to emit as an
// analytic normal. Each return that could not be differentiated is reported in
// Result.GradientNotes. Off by default: it adds a differentiated copy of every
// return to the program.
func WithAnalyticGradient() Option {
	return func(o *options) {
		o.analyticGradient = true
	}
}

// WithProgress supplies a progress callback to report compiler pass progress.
func WithProgress(fn progress.Callback) Option {
	return func(o *options) {
		o.progressFn = fn
	}
}
