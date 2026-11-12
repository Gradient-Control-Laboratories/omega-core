// Package eval provides ad-hoc GSL expression evaluation, bounding box fitting,
// global extremum search, sound interval certification, surface projection,
// volume estimation, and lit preview rendering against GSL documents.
package eval

import (
	"context"

	"gradientcontrol.com/omega/core/assembler"
)

// Option configures program assembly inside the evaluator.
type Option = assembler.Option

// GradSample is the field value and spatial gradient at one evaluation point.
type GradSample struct {
	Value float32 `json:"value"`
	Dx    float32 `json:"dx"`
	Dy    float32 `json:"dy"`
	Dz    float32 `json:"dz"`
}

// GradGroup is one return value's evaluation (GSL type name and per-component samples).
type GradGroup struct {
	TypeName string         `json:"type"`
	Samples  [][]GradSample `json:"samples"`
}

// IntervalSample is a sound [Lower, Upper] bracket of a field over one box.
type IntervalSample struct {
	Lower float32 `json:"lower"`
	Upper float32 `json:"upper"`
}

// EvalAssetsRequest asks for the mesh-asset context a geometry evaluation needs.
type EvalAssetsRequest struct {
	ID            string            `json:"id"`
	AssetManifest map[string]string `json:"assetManifest,omitempty"`
	Module        []byte            `json:"module"`
	DigestOnly    bool              `json:"digestOnly,omitempty"`
}

// EvalAssetsResponse carries the import slot table, oracle payloads, and digests for a module.
type EvalAssetsResponse struct {
	SpinodoidBounds []float64           `json:"spinodoidBounds,omitempty"`
	ID              string              `json:"id"`
	ImportSlots     []ImportSlotRef     `json:"importSlots,omitempty"`
	MeshOracles     []MeshOraclePayload `json:"meshOracles,omitempty"`
	VdbOracles      []VdbOraclePayload  `json:"vdbOracles,omitempty"`
	Digests         []AssetDigest       `json:"digests,omitempty"`
	Error           string              `json:"error,omitempty"`
}

// ImportSlotRef maps a package asset to a slot index.
type ImportSlotRef struct {
	PkgPath string `json:"pkgPath,omitempty"`
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	Slot    int    `json:"slot"`
}

// MeshOraclePayload bundles triangle data for a bvhSdf(assetID) leaf.
type MeshOraclePayload struct {
	AssetID     uint32 `json:"assetID"`
	VerticesB64 string `json:"verticesB64"`
}

// VdbOraclePayload bundles triangle and grid settings for a vdbSdf(assetID) leaf.
type VdbOraclePayload struct {
	AssetID        uint32  `json:"assetID"`
	VerticesB64    string  `json:"verticesB64"`
	VoxelSize      float32 `json:"voxelSize"`
	HalfBandVoxels float32 `json:"halfBandVoxels"`
}

// AssetDigest summarizes one referenced mesh asset's placement and orientation.
type AssetDigest struct {
	Slot          int             `json:"slot"`
	Filename      string          `json:"filename,omitempty"`
	ContentHash   string          `json:"contentHash,omitempty"`
	BboxMin       [3]float64      `json:"bboxMin"`
	BboxMax       [3]float64      `json:"bboxMax"`
	Centroid      [3]float64      `json:"centroid"`
	Area          float64         `json:"area"`
	PrincipalAxes [3][3]float64   `json:"principalAxes"`
	HalfExtents   [3]float64      `json:"halfExtents"`
	Thickness     *MapRangeDigest `json:"thickness,omitempty"`
	Clearance     *MapRangeDigest `json:"clearance,omitempty"`
}

// MapRangeDigest summarizes the finite range of a surface-analysis map.
type MapRangeDigest struct {
	Min      float64    `json:"min"`
	MinAt    [3]float64 `json:"minAt"`
	Max      float64    `json:"max"`
	MaxAt    [3]float64 `json:"maxAt"`
	Coverage float64    `json:"coverage"`
}

// RenderGlslResponse holds generated GLSL source and slider input metadata.
type RenderGlslResponse struct {
	ID     string                  `json:"id"`
	Inputs map[string]*RenderInput `json:"inputs"`
	Glsl   string                  `json:"glsl"`
	Error  string                  `json:"error,omitempty"`
}

// RenderInput describes one slider input's default value and bounds.
type RenderInput struct {
	Value  float64 `json:"value"`
	HasMin bool    `json:"hasMin"`
	HasMax bool    `json:"hasMax"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

// AssetFetcher returns the mesh-asset context (import slots and oracle payloads)
// for a module. digestOnly requests skip the heavy payloads and return just
// slots and digests, for prompt context.
type AssetFetcher func(ctx context.Context, manifest map[string]string, module []byte, digestOnly bool) (*EvalAssetsResponse, error)

// BoundsResult reports the axis-aligned box of the region where an expression
// is negative (inside), fit on a resolution^3 grid, padded by up to one cell.
type BoundsResult struct {
	Found      bool
	BboxMin    [3]float64
	BboxMax    [3]float64
	CellSize   [3]float64
	Resolution int
	MinValue   float32
	MinAt      [3]float64
}

// CertifyResult is the outcome of a sound interval-arithmetic proof that a
// scalar expression is at least a threshold over an entire region.
type CertifyResult struct {
	Expr      string
	Threshold float64
	Tolerance float64
	Evals     int

	RegionLower float64
	RegionUpper float64

	Proven        bool
	Disproven     bool
	Indeterminate bool
	Reason        string

	CounterBox   [6]float64
	CounterUpper float64

	WorstBox   [6]float64
	WorstLower float64

	Params        map[string][2]float64
	CounterParams map[string][2]float64
	WorstParams   map[string][2]float64
}

// The Reason values an Indeterminate CertifyResult can carry.
const (
	ReasonBudget    = "budget"
	ReasonPool      = "pool"
	ReasonTolerance = "tolerance"
)

// ExtremumResult reports the global minimum and/or maximum of a scalar
// expression over a region, found by Lipschitz branch-and-bound.
type ExtremumResult struct {
	Mode      string
	Lipschitz float64
	Tolerance float64
	Evals     int

	Min          float64
	MinAt        [3]float64
	MinLower     float64
	MinConverged bool

	Max          float64
	MaxAt        [3]float64
	MaxUpper     float64
	MaxConverged bool
}

// SurfaceSample is one seed point projected onto an expression's zero level set.
type SurfaceSample struct {
	Point     [3]float64
	Normal    [3]float64
	Value     float64
	Converged bool
}

// VolumeBounds brackets the solid volume inside a containing region in cubic millimetres.
type VolumeBounds struct {
	Lower float64 `json:"lowerMm3"`
	Upper float64 `json:"upperMm3"`
	Evals int     `json:"evals"`
}

// LitView is the camera and ground of a lit CPU render.
type LitView struct {
	Dir             [3]float64
	Up              [3]float64
	Light           bool
	NoAxes          bool
	Environment     string
	EnvironmentTurn float64
	GlassSamples    int
}

// LitOptions controls a lit CPU render. The zero value renders 900x600 at the
// document's default framing.
type LitOptions struct {
	Width, Height   int
	WorldExtent     float64
	Time            float64
	View            [3]float64
	Up              [3]float64
	Light           bool
	NoAxes          bool
	Environment     string
	EnvironmentTurn float64
	GlassSamples    int

	Camera      *LitCamera
	Colors      [][3]float64
	ViewerMarch bool
	BackDim     float64
	NoMaterials bool
	Simple      *LitSimple
}

// LitCamera places the camera in world Z-up coordinates.
type LitCamera struct {
	Pos, Right, Up, Fwd [3]float64
	TanFov              [2]float64
	Ortho               bool
}

// LitSimple holds the Simple render tier's constants.
type LitSimple struct {
	Steps                    int
	Fudge                    float64
	HitFloor, HitPixel       float64
	NormalFloor, NormalPixel float64
	Ambient                  float64
}

// LitLinear is a lit render before display encoding.
type LitLinear struct {
	Width, Height int
	RGB           []float32
	Shape         []int32
	Glass         []bool
	Depth         []float32
	Radiance      []float32
}
