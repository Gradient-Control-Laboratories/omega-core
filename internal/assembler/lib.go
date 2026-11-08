package lib

import (
	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/assembler/topofold"
	"gradientcontrol.com/omega/core/internal/assembler/warpfold"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// Option configures program assembly.
type Option = assembler.Option

// ParsePackageFunc parses all source files in a package into a *parser.Result.
type ParsePackageFunc = assembler.ParsePackageFunc

// VariantResult holds the sliced MeshExportPlan, analytical SdfSubtrees, or
// per-variant compilation error for a single caller-supplied variant return.
type VariantResult = assembler.VariantResult

// Result contains the assembled program AST and the assigned import slot table.
type Result = assembler.Result

// ProgressFunc is invoked at discrete compilation milestones during program assembly.
type ProgressFunc = assembler.ProgressFunc

// Parser supplies a custom package parser (such as parser.CachingParser.ParsePackage)
// to use when parsing primary and imported packages during assembly.
func Parser(parsePkg ParsePackageFunc) Option {
	return assembler.Parser(parsePkg)
}

// GeometryResolver supplies a topology geometry resolver for folding STEP topology builtins.
func GeometryResolver(r topofold.GeometryResolver) Option {
	return assembler.GeometryResolver(r)
}

// WarpResolver supplies a warp simulation resolver for baking and folding warped mesh assets.
func WarpResolver(r warpfold.WarpResolver) Option {
	return assembler.WarpResolver(r)
}

// FoldWarpFieldToZero replaces every warpField(...) call across the program with [0, 0, 0]
// during Step 5 domain folding for backends that do not sample live warp textures.
func FoldWarpFieldToZero() Option {
	return assembler.FoldWarpFieldToZero()
}

// DisableMeshSlotFolding disables the automatic mesh slot assignment and folding pass.
func DisableMeshSlotFolding() Option {
	return assembler.DisableMeshSlotFolding()
}

// DisableDeadcodeRemoval disables dead code removal during assembly.
func DisableDeadcodeRemoval() Option {
	return assembler.DisableDeadcodeRemoval()
}

// ExtractUI enables extraction of sliders, UI controls, and scene metadata into the Result.
func ExtractUI() Option {
	return assembler.ExtractUI()
}

// ExtractMeshPlan configures the assembler to partition geometry into a MeshExportPlan
// and analytical SdfSubtrees as Step 8. If idPrefix is empty, it defaults to "leaf-".
func ExtractMeshPlan(idPrefix string) Option {
	return assembler.ExtractMeshPlan(idPrefix)
}

// WithVariants configures the assembler to compile and slice the named
// alternative top-level return statements (e.g. map[string]string{"part": "return mainPart();"})
// alongside the primary package in a single pass.
func WithVariants(variants map[string]string) Option {
	return assembler.WithVariants(variants)
}

// WithProgress supplies a progress callback to report compiler pass progress.
func WithProgress(fn ProgressFunc) Option {
	return assembler.WithProgress(fn)
}

// Assemble collects the packages directly and indirectly referenced by the
// primary package, executes compiler passes and pluggable fold passes, and
// assembles them into a single program along with the assigned import slot table.
func Assemble(primarySources []string, srcMgr source.Manager, opts ...Option) (*Result, error) {
	return assembler.AssembleWithSource(primarySources, srcMgr, opts...)
}

// DependenciesSha256 computes the deterministic SHA-256 hex digest of pkgPath
// and all its transitive GSL package dependencies from srcMgr without reading
// mesh assets.
func DependenciesSha256(pkgPath string, srcMgr source.Manager) (string, error) {
	return assembler.DependenciesSha256(pkgPath, srcMgr)
}

// AnalyzePackage parses source into an un-inlined *program.Package and stamps
// evaluated constant float64 values onto top-level assignments, local function
// expressions/returns, and call arguments while preserving the original AST
// structure for template comparisons and bounds extraction.
func AnalyzePackage(sourceText string, getImport program.GetImportFunc) (*program.Package, error) {
	return assembler.AnalyzePackage(sourceText, getImport)
}
