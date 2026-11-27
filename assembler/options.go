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
	conditionDerivatives  bool
}

// Parser supplies a custom package parser (such as parser.CachingParser.ParsePackage)
// to use when parsing primary and imported packages during assembly.
func Parser(parsePkg ParsePackageFunc) Option {
	return func(o *options) {
		o.parsePkg = parsePkg
	}
}

// SourcePaths adds local directory root paths used to resolve imported packages and assets.
func SourcePaths(paths ...string) Option {
	return func(o *options) {
		if len(paths) > 0 {
			o.managers = append(o.managers, internalsource.NewLocalManager(paths...))
		}
	}
}

// SourceManager adds a custom source manager used to resolve imported packages and assets.
func SourceManager(mgr source.Manager) Option {
	return func(o *options) {
		if mgr != nil {
			o.managers = append(o.managers, mgr)
		}
	}
}

// GeometryResolver supplies a topology geometry resolver for folding STEP topology builtins.
func GeometryResolver(r topofold.GeometryResolver) Option {
	return func(o *options) {
		o.geometryResolver = r
	}
}

// WarpResolver supplies a warp simulation resolver for baking and folding warped mesh assets.
func WarpResolver(r warpfold.WarpResolver) Option {
	return func(o *options) {
		o.warpResolver = r
	}
}

// FoldWarpFieldToZero replaces every warpField(...) call across the program with [0, 0, 0]
// during Step 5 domain folding for backends that do not sample live warp textures.
func FoldWarpFieldToZero() Option {
	return func(o *options) {
		o.foldWarpFieldToZero = true
	}
}

// DisableMeshSlotFolding disables the automatic mesh slot assignment and folding pass.
func DisableMeshSlotFolding() Option {
	return func(o *options) {
		o.enableSlotFold = false
	}
}

// DisableDeadcodeRemoval disables dead code removal during assembly.
func DisableDeadcodeRemoval() Option {
	return func(o *options) {
		o.enableDeadcodeRemoval = false
	}
}

// ExtractUI enables extraction of sliders, UI controls, and scene metadata into the Result.
func ExtractUI() Option {
	return func(o *options) {
		o.extractUI = true
	}
}

// ExtractMeshPlan configures the assembler to partition geometry into a MeshExportPlan
// and analytical SdfSubtrees as Step 8. If idPrefix is empty, it defaults to "leaf-".
func ExtractMeshPlan(idPrefix string) Option {
	return func(o *options) {
		o.extractMeshPlan = true
		o.meshPlanIdPrefix = idPrefix
	}
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

// WithConditionDerivatives has a .d(var) or .dp written in a return taken of the
// field a condition reads (eval.Evaluator.SetCondition): of the return as built,
// and of its displaced distance, the coats on (Program.DisplacedReturns), where
// a shape inside it is layered. For queries, whose expressions are their
// returns; documents keep the default, every derivative taken as built.
func WithConditionDerivatives() Option {
	return func(o *options) {
		o.conditionDerivatives = true
	}
}

// WithProgress supplies a progress callback to report compiler pass progress.
func WithProgress(fn progress.Callback) Option {
	return func(o *options) {
		o.progressFn = fn
	}
}
