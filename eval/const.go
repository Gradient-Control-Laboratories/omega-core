package eval

import (
	"gradientcontrol.com/omega/core/assembler"
	internaleval "gradientcontrol.com/omega/core/internal/eval"
	fidgetapi "gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

// Option configures program assembly inside the evaluator.
type Option = assembler.Option

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

