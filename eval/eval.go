// Package eval provides ad-hoc GSL expression evaluation, bounding box fitting,
// global extremum search, sound interval certification, surface projection,
// volume estimation, and lit preview rendering against GSL documents.
package eval

import (
	"gradientcontrol.com/omega/core/assembler/source"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
	fidgetapi "gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/gsleval"
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
