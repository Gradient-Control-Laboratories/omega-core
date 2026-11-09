// Package eval provides ad-hoc GSL expression evaluation, bounding box fitting,
// global extremum search, sound interval certification, surface projection,
// volume estimation, and lit preview rendering against GSL documents.
package eval

import (
	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/source"
	internaleval "gradientcontrol.com/omega/core/internal/eval"
	fidgetapi "gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/fidget/lib"
	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/program"
)

// Evaluator evaluates GSL query expressions against a document via Fidget.
type Evaluator = gsleval.Evaluator

// AssetFetcher returns the mesh-asset context (import slots and oracle payloads)
// for a module.
type AssetFetcher = gsleval.AssetFetcher

// Option configures program assembly inside the evaluator.
type Option = assembler.Option

// BoundsResult reports the axis-aligned box of the region where an expression is negative.
type BoundsResult = gsleval.BoundsResult

// CertifyResult is the outcome of a sound interval-arithmetic proof.
type CertifyResult = gsleval.CertifyResult

// ExtremumResult reports the global minimum and/or maximum of a scalar expression.
type ExtremumResult = gsleval.ExtremumResult

// SurfaceSample is one seed point projected onto an expression's zero level set.
type SurfaceSample = gsleval.SurfaceSample

// VolumeBounds brackets the solid volume inside a containing region in cubic millimetres.
type VolumeBounds = gsleval.VolumeBounds

// GradSample is the field value and spatial gradient at one evaluation point.
type GradSample = fidgetapi.GradSample

// GradGroup is one return value's evaluation (GSL type name and per-component samples).
type GradGroup = fidgetapi.GradGroup

// IntervalSample is a sound [Lower, Upper] bracket of a field over one box.
type IntervalSample = fidgetapi.IntervalSample

// EvalAssetsRequest asks for the mesh-asset context a geometry evaluation needs.
type EvalAssetsRequest = fidgetapi.EvalAssetsRequest

// EvalAssetsResponse carries the import slot table, oracle payloads, and digests for a module.
type EvalAssetsResponse = fidgetapi.EvalAssetsResponse

// ImportSlotRef maps a package asset to a slot index.
type ImportSlotRef = fidgetapi.ImportSlotRef

// MeshOraclePayload bundles triangle data for a bvhSdf(assetID) leaf.
type MeshOraclePayload = fidgetapi.MeshOraclePayload

// VdbOraclePayload bundles triangle and grid settings for a vdbSdf(assetID) leaf.
type VdbOraclePayload = fidgetapi.VdbOraclePayload

// AssetDigest summarizes one referenced mesh asset's placement and orientation.
type AssetDigest = fidgetapi.AssetDigest

// MapRangeDigest summarizes the finite range of a surface-analysis map.
type MapRangeDigest = fidgetapi.MapRangeDigest

// RenderGlslResponse holds generated GLSL source and slider input metadata.
type RenderGlslResponse = fidgetapi.RenderGlslResponse

// RenderInput describes one slider input's default value and bounds.
type RenderInput = fidgetapi.RenderInput

// LitView is the camera and ground of a lit CPU render.
type LitView = lib.LitView

// LitOptions controls a lit CPU render.
type LitOptions = lib.LitOptions

// LitCamera places the camera in world Z-up coordinates.
type LitCamera = lib.LitCamera

// LitSimple holds the Simple render tier's constants.
type LitSimple = lib.LitSimple

// LitLinear is a lit render before display encoding.
type LitLinear = lib.LitLinear

// The Reason values an Indeterminate CertifyResult can carry.
const (
	ReasonBudget    = gsleval.ReasonBudget
	ReasonPool      = gsleval.ReasonPool
	ReasonTolerance = gsleval.ReasonTolerance
)

// New returns an Evaluator backed by Fidget. If srcMgr is nil, embedded
// standard library packages remain available automatically.
func New(srcMgr source.Manager, fetchAssets AssetFetcher, options ...Option) *Evaluator {
	return gsleval.New(srcMgr, fetchAssets, options...)
}

// SliderRanges parses the document and returns every declared slider's [min, max] range.
func SliderRanges(code string) (map[string][2]float64, error) {
	return gsleval.SliderRanges(code)
}

// TopLevelNames returns the set of names declared at the top level of source.
func TopLevelNames(source string) map[string]bool {
	return gsleval.TopLevelNames(source)
}

// TopLevelReturnExprs returns the source text of each expression in the document's
// top-level return statement, in order.
func TopLevelReturnExprs(source string) []string {
	return gsleval.TopLevelReturnExprs(source)
}

// EvalGrad evaluates each scalar return of the program at the given points,
// returning per-return samples (value and spatial gradient).
func EvalGrad(prog *program.Program, xs, ys, zs []float32) ([][]GradSample, error) {
	return lib.EvalGrad(prog, xs, ys, zs)
}

// EvalGradGroups evaluates every return of the program at the given points,
// accepting vector- and matrix-valued returns.
func EvalGradGroups(prog *program.Program, xs, ys, zs []float32) ([]GradGroup, error) {
	return lib.EvalGradGroups(prog, xs, ys, zs)
}

// EvalInterval interval-bounds each scalar return of the program over every
// axis-aligned box, returning a sound [Lower, Upper] bracket per box.
func EvalInterval(prog *program.Program, boxes [][6]float32) ([][]IntervalSample, error) {
	return lib.EvalInterval(prog, boxes)
}

// EvalIntervalParams is EvalInterval with named document parameters ranging over intervals.
func EvalIntervalParams(prog *program.Program, boxes [][6]float32, params map[string][2]float64) ([]IntervalSample, error) {
	return lib.EvalIntervalParams(prog, boxes, params)
}

// EvalIntervalParamBoxes interval-bounds the program's single scalar return over
// each box with per-box parameter ranges.
func EvalIntervalParamBoxes(prog *program.Program, names []string, boxes [][6]float32, bounds [][][2]float64) ([]IntervalSample, error) {
	return lib.EvalIntervalParamBoxes(prog, names, boxes, bounds)
}

// Render generates a PNG preview and GLSL representation of prog.
func Render(prog *program.Program) ([]byte, *RenderGlslResponse, error) {
	return lib.Render(prog)
}

// RenderFramed is Render with the camera framed on the geometry.
func RenderFramed(prog *program.Program) ([]byte, *RenderGlslResponse, error) {
	return lib.RenderFramed(prog)
}

// RenderLit renders prog on the CPU with the viewer's lighting and materials.
func RenderLit(prog *program.Program, opts LitOptions) ([]byte, error) {
	return lib.RenderLit(prog, opts)
}

// RenderLitLinear renders prog to linear-light RGB, per-pixel shape index, and depth.
func RenderLitLinear(prog *program.Program, opts LitOptions) (*LitLinear, error) {
	return lib.RenderLitLinear(prog, opts)
}

// RenderLitSized renders prog on the CPU at the given dimensions and world extent.
func RenderLitSized(prog *program.Program, width, height int, worldExtent, timeAt float64) ([]byte, error) {
	return lib.RenderLitSized(prog, width, height, worldExtent, timeAt)
}

// RenderLitSizedView renders prog on the CPU with a chosen camera and ground.
func RenderLitSizedView(prog *program.Program, width, height int, worldExtent, timeAt float64, view LitView) ([]byte, error) {
	return lib.RenderLitSizedView(prog, width, height, worldExtent, timeAt, view)
}

// RenderLitSizedViewRadiance is RenderLitSizedView returning each pixel's linear radiance beside the PNG.
func RenderLitSizedViewRadiance(prog *program.Program, width, height int, worldExtent, timeAt float64, view LitView) ([]byte, []float32, error) {
	return lib.RenderLitSizedViewRadiance(prog, width, height, worldExtent, timeAt, view)
}

// ComposeSheet tiles PNG panels into one PNG, cols across.
func ComposeSheet(panels [][]byte, cols int) ([]byte, error) {
	return lib.ComposeSheet(panels, cols)
}

// ViewPresetNames lists the camera preset names accepted by ParseView.
func ViewPresetNames() []string {
	return lib.ViewPresetNames()
}

// ParseView reads a camera as a preset name ("iso-below") or an "x,y,z" direction.
func ParseView(s string) (LitView, error) {
	return lib.ParseView(s)
}

// ParseViews reads a comma-separated list of cameras.
func ParseViews(s string) ([]LitView, error) {
	return lib.ParseViews(s)
}

// ParseVec3 reads "x,y,z" into a 3-vector.
func ParseVec3(s string) ([3]float64, error) {
	return lib.ParseVec3(s)
}

// Const evaluates an Operand to a scalar or vector of float64 values using the
// given top-level declarations as the name scope. Scalars return a length-1
// slice; vectors return a length-N slice. Returns an error if the operand
// cannot be folded to a constant.
func Const(op *program.Operand, assignments []*program.Assignment, fnTypes []*program.FnType) ([]float64, error) {
	return internaleval.Const(op, assignments, fnTypes)
}

// ConstPkg is Const with import support: references that do not resolve to one
// of the package's own top-level constants are looked up in dot-imported
// packages, and qualified references (prefix.name) resolve through the
// matching prefixed import. Imported packages are loaded lazily via getImport.
func ConstPkg(op *program.Operand, pkg *program.Package, getImport program.GetImportFunc) ([]float64, error) {
	return internaleval.ConstPkg(op, pkg, getImport)
}

// ConstElementsPkg evaluates an operand as a collection and returns one
// float64 slice per element.
func ConstElementsPkg(op *program.Operand, pkg *program.Package, getImport program.GetImportFunc) ([][]float64, error) {
	return internaleval.ConstElementsPkg(op, pkg, getImport)
}
