package fidget

import "gradientcontrol.com/omega/core/internal/noise/spinodoid"

const (
	// RequestPathMesh is fidget-only.
	RequestPathMesh          = "mesh/request"
	ResponsePathMeshStl      = "mesh/stl"
	ResponsePathMeshProgress = "mesh/progress"

	// RequestPathTape serializes the parsed module into an omega tape
	// (tape::Tape protobuf — the post-lowering fidget-tree DAG) and writes it
	// to the downloads dir. The request reuses MeshRequest (program + oracles;
	// bbox/depth/algorithm are ignored). The response reuses MeshStlResponse
	// (Url points at the written .tape).
	RequestPathTape  = "tape/request"
	ResponsePathTape = "tape/result"

	// RequestPathMeasure asks for numeric geometry measurements of a compiled
	// program (bounding box, volume, surface area) derived from a measuring
	// mesh. It is the headless inspection surface behind the canvas agent's
	// measurement bridge and
	// carries the same program + oracle context as a mesh request, so any
	// program the mesher accepts can be measured.
	RequestPathMeasure  = "fidget/measure"
	ResponsePathMeasure = "fidget/measure/result"

	RequestPathRender = "render"

	// RequestPathRenderLit asks for litrender's frame of a program, drawn the
	// way the viewer's standard frame is (its march, shading and camera):
	// the CPU render tier. The request is RenderLitRequest; the response,
	// RenderLitResponse, points at the raw frame written to the downloads dir.
	RequestPathRenderLit  = "fidget/render-lit"
	ResponsePathRenderLit = "fidget/render-lit/result"

	RequestPathRaycast = "raycast/request"

	// VDB-extract is the canvas -> fidget request that bakes a triangle soup
	// into an OpenVDB narrow-band level set and ships back a uniform
	// dense grid the WGSL preview can sample. The fidget agent already
	// holds the OpenVDB level set in memory once registered, so callers
	// can retry the same request cheaply (the second call hits the cache
	// inside register_vdb_oracle).
	RequestPathVdbExtract = "vdb/extract"
	ResponsePathVdbData   = "vdb/data"

	// VdbProgress carries per-mille progress ticks during a vdb-extract
	// bake — emitted from inside the Rust FFI as OpenVDB's Interrupter
	// fires and as the dense-grid extract walks Z-slices. The canvas
	// agent listens to these and forwards them onto the unified
	// canvas/operation-progress stream the browser renders.
	ResponsePathVdbProgress = "vdb/progress"

	// SpinodoidBake is the canvas -> fidget request for the Gabor spinodoid
	// tables a program's gaborSpinodoid calls read. The kernels are baked
	// here because their scale and orientation drivers are fields of p, and
	// only fidget evaluates fields; the WGSL preview binds the tables the
	// response carries, matched to its calls by spinodoid.Call key. The reply
	// goes to the sender only.
	RequestPathSpinodoidBake  = "fidget/spinodoid-bake"
	ResponsePathSpinodoidBake = "fidget/spinodoid-bake/result"
)

// SpinodoidBakeRequest asks for every spinodoid table prog reads. Program
// carries the serialized pre-folded program.Program.
type SpinodoidBakeRequest struct {
	ID      string `json:"id"`
	Program []byte `json:"program"` // Serialized program.Program
	// Known lists the content ids of tables the requester already holds;
	// those come back without texels.
	Known []string `json:"known,omitempty"`
}

// SpinodoidTable is one baked table in the GPU layout of
// common/noise/spinodoid (GPUTexels): RGBA float32 texels, little-endian,
// base64.
type SpinodoidTable = spinodoid.BakedTable

type SpinodoidBakeResponse struct {
	ID     string           `json:"id"`
	Tables []SpinodoidTable `json:"tables,omitempty"`
	Error  string           `json:"error,omitempty"`
}

type MeshRequest struct {
	ID      string `json:"id"`
	Program []byte `json:"program"` // Serialized program.Program
	// Bbox is the world-space region the fidget octree builds over,
	// packed as [xmin, ymin, zmin, xmax, ymax, zmax].
	Bbox [6]float32 `json:"bbox"`
	// SubdivisionLevel is the fidget octree depth.
	SubdivisionLevel int                 `json:"subdivisionLevel"`
	Time             float64             `json:"time"`
	Algorithm        string              `json:"algorithm"`
	MeshOracles      []MeshOraclePayload `json:"meshOracles,omitempty"`
	VdbOracles       []VdbOraclePayload  `json:"vdbOracles,omitempty"`

	// Inline (tape requests only) returns the serialized tape in the response
	// as TapeB64 instead of writing it to the downloads dir and returning a
	// URL. The browser's Save Tape wants a file to download; a server-side
	// consumer (a simulation, which posts the tape to an out-of-
	// process service) wants the bytes, and canvas has no downloads dir to
	// read them back from — nor would one exist if the two agents were ever
	// split across containers.
	Inline bool `json:"inline,omitempty"`

	// Appearance (tape requests only) also serializes each return's colour
	// and material fields into the tape's `appearance`.
	Appearance bool `json:"appearance,omitempty"`
	// LevelSets (tape requests only) also bakes each vdbSdf's level set and
	// carries it in the tape's oracle payload, so a reader without OpenVDB
	// (the browser's dual contouring) samples the same values. It is large —
	// every voxel of the band — so only a reader that needs it asks.
	LevelSets bool `json:"levelSets,omitempty"`
	// Filename specifies the explicit destination filename under downloads dir
	// (e.g. "part-a1b2c3d4.stl" or "leaf-a1b2c3d4.stl"). If empty, fidget
	// falls back to generating a timestamped name.
	Filename string `json:"filename,omitempty"`
}

// MeasureRequest asks for numeric geometry measurements of a compiled
// program. The program and oracle payloads mean what they mean on MeshRequest;
// a coarse pass over Bbox finds the shape, then a measuring mesh at Depth over
// the shape's own (padded) bounding box produces the numbers.
type MeasureRequest struct {
	ID         string `json:"id"`
	ExportMesh bool   `json:"exportMesh,omitempty"`
	Program    []byte `json:"program"` // Serialized program.Program
	// Depth is the octree depth of the measuring mesh (default 7, max 10).
	Depth int `json:"depth"`
	// Bbox is the region searched for the shape. Zero means the default
	// search region.
	Bbox        [6]float32          `json:"bbox"`
	MeshOracles []MeshOraclePayload `json:"meshOracles,omitempty"`
	VdbOracles  []VdbOraclePayload  `json:"vdbOracles,omitempty"`
}

type MeasureResponse struct {
	ID  string `json:"id"`
	URL string `json:"url,omitempty"`
	// Empty is set when no surface was found in the search region.
	Empty bool `json:"empty"`
	// Bbox is the tight world bounding box [xmin, ymin, zmin, xmax, ymax, zmax].
	Bbox          [6]float32 `json:"bbox"`
	Extents       [3]float32 `json:"extents"`
	Center        [3]float32 `json:"center"` // volume centroid
	Volume        float64    `json:"volume"`
	SurfaceArea   float64    `json:"surfaceArea"`
	TriangleCount int        `json:"triangleCount"`
	Error         string     `json:"error,omitempty"`
}

// ImportSlotRef is one package asset to slot mapping for fidget's import fold.
type ImportSlotRef struct {
	PkgPath string `json:"pkgPath,omitempty"`
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	Slot    int    `json:"slot"`
}

// MeshOraclePayload bundles the triangle data the fidget agent needs to
// register a bvhSdf(assetID) leaf before evaluating the module. The canvas
// agent attaches one entry per asset_id referenced by bvhSdf() in the
// program. VerticesB64 is little-endian f32 packed 3 floats per vertex, 3
// vertices per triangle.
type MeshOraclePayload struct {
	AssetID     uint32 `json:"assetID"`
	VerticesB64 string `json:"verticesB64"`
}

// VdbOraclePayload mirrors MeshOraclePayload for the OpenVDB-backed
// vdbSdf(assetID) leaf. Triangles use the same packed layout. VoxelSize is
// the world-space cell size for openvdb::tools::meshToLevelSet;
// HalfBandVoxels is the narrow-band half-width in voxels (background
// magnitude on the resulting grid is VoxelSize * HalfBandVoxels).
type VdbOraclePayload struct {
	AssetID        uint32  `json:"assetID"`
	VerticesB64    string  `json:"verticesB64"`
	VoxelSize      float32 `json:"voxelSize"`
	HalfBandVoxels float32 `json:"halfBandVoxels"`
}

// RenderCamera is the viewer's camera, in world Z-up: its position, its
// orthonormal basis, and the tangents of the half field of view (for an
// orthographic camera, the half-width and half-height of the view).
type RenderCamera struct {
	Pos    [3]float64 `json:"pos"`
	Right  [3]float64 `json:"right"`
	Up     [3]float64 `json:"up"`
	Fwd    [3]float64 `json:"fwd"`
	TanFov [2]float64 `json:"tanFov"`
	Ortho  bool       `json:"ortho,omitempty"`
}

// RenderLitRequest is a RequestPathRenderLit request. The embedded
// MeshRequest carries the program and the context it needs to build, as for a
// tape (the program, oracles, import slots and folded imports); its bbox,
// depth and algorithm are ignored.
type RenderLitRequest struct {
	MeshRequest
	Width       int          `json:"width"`
	Height      int          `json:"height"`
	WorldExtent float64      `json:"worldExtent"`
	Camera      RenderCamera `json:"camera"`
	// Colors is each return's colour where the program authors none: the
	// viewer's per-shape colours. For "simple", the Simple palette.
	Colors [][3]float64 `json:"colors,omitempty"`
	// Mode is the viewer tier drawn: "standard" (or empty), "color" (the
	// standard frame without the program's materials), or "simple" (the
	// Simple frame, whose constants Simple carries).
	Mode   string      `json:"mode,omitempty"`
	Simple *SimpleSpec `json:"simple,omitempty"`
	// BackDim is how bright a surface seen from behind is drawn: the Render
	// panel's Back faces switch, 0.25 at Dim and 1 at Full. Zero is Dim.
	BackDim float64 `json:"backDim,omitempty"`
}

// SimpleSpec is the Simple frame's constants, as common/web/gsl-simple.js
// defines them.
type SimpleSpec struct {
	Steps       int     `json:"steps"`
	Fudge       float64 `json:"fudge"`
	HitFloor    float64 `json:"hitFloor"`
	HitPixel    float64 `json:"hitPixel"`
	NormalFloor float64 `json:"normalFloor"`
	NormalPixel float64 `json:"normalPixel"`
	Ambient     float64 `json:"ambient"`
}

// RenderLitResponse answers a RenderLitRequest. Url is the frame: Width x
// Height pixels, row-major from the top left, five little-endian float32s
// each: the tone-mapped linear colour, the return hit (numbered from 1, 0 for
// none), and the depth along the view direction.
type RenderLitResponse struct {
	ID     string `json:"id"`
	Url    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Error  string `json:"error,omitempty"`
}

type MeshStlResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Url   string `json:"url,omitempty"`
	Error string `json:"error,omitempty"`

	// TapeB64 carries the serialized tape itself, set instead of Url when the
	// tape request asked for MeshRequest.Inline. Tape responses only.
	TapeB64 string `json:"tapeB64,omitempty"`
}

type MeshProgressResponse struct {
	ID        string `json:"id"`
	Completed uint32 `json:"completed,omitempty"`
	Total     uint32 `json:"total,omitempty"`
	Status    string `json:"status,omitempty"`
}

type RenderPngResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Url    string `json:"url,omitempty"`
	Base64 string `json:"base64,omitempty"`
	Error  string `json:"error,omitempty"`
}

type RenderGlslResponse struct {
	ID     string                  `json:"id"`
	Inputs map[string]*RenderInput `json:"inputs"`
	Glsl   string                  `json:"glsl"`
	Error  string                  `json:"error,omitempty"`
}

type RenderInput struct {
	Value  float64 `json:"value"`
	HasMin bool    `json:"hasMin"`
	HasMax bool    `json:"hasMax"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

// GradSample is the field value and spatial gradient at one eval point.
type GradSample struct {
	Value float32 `json:"value"`
	Dx    float32 `json:"dx"`
	Dy    float32 `json:"dy"`
	Dz    float32 `json:"dz"`
}

// GradGroup is one return value's evaluation: its GSL type name ("num" for
// scalars) and per-component samples indexed [component][point]. A vector
// or matrix return carries one component per scalar element, in order.
type GradGroup struct {
	TypeName string         `json:"type"`
	Samples  [][]GradSample `json:"samples"`
}

// IntervalSample is a sound [Lower, Upper] bracket of a field over one box:
// the field's true range over the box lies within it (oracle-aware interval
// arithmetic). Lower >= 0 proves the field is non-negative over the whole box;
// Upper < 0 proves it negative — no Lipschitz constant required.
type IntervalSample struct {
	Lower float32 `json:"lower"`
	Upper float32 `json:"upper"`
}

// Eval-assets is the ai-agent -> canvas request that gathers the mesh-asset
// context a geometry eval needs: the module's importStep/importMesh slot
// table plus the bvhSdf/vdbSdf oracle payloads from the client's asset
// registry. It mirrors what the canvas packs into a MeshRequest for fidget.
const (
	RequestPathEvalAssets  = "canvas/eval-assets/request"
	ResponsePathEvalAssets = "canvas/eval-assets/response"
)

type EvalAssetsRequest struct {
	ID string `json:"id"`
	// AssetManifest maps asset filenames to their SHA-256 content hashes.
	AssetManifest map[string]string `json:"assetManifest,omitempty"`
	// Module is the proto model.Package whose mesh references to gather.
	Module []byte `json:"module"`
	// DigestOnly skips the (heavy) oracle payload encoding and returns just
	// the slot table and per-asset digests — for prompt context rather than
	// evaluation.
	DigestOnly bool `json:"digestOnly,omitempty"`
}

type EvalAssetsResponse struct {
	// SpinodoidBounds is the shared geometry domain used by preview/export,
	// so chat evaluations bake graded fields over the same imported bodies.
	SpinodoidBounds []float64           `json:"spinodoidBounds,omitempty"`
	ID              string              `json:"id"`
	ImportSlots     []ImportSlotRef     `json:"importSlots,omitempty"`
	MeshOracles     []MeshOraclePayload `json:"meshOracles,omitempty"`
	VdbOracles      []VdbOraclePayload  `json:"vdbOracles,omitempty"`
	Digests         []AssetDigest       `json:"digests,omitempty"`
	Error           string              `json:"error,omitempty"`
}

// AssetDigest summarizes one referenced mesh asset's placement and
// orientation, computed exactly from its triangles (see
// meshasset.ComputeDigest).
type AssetDigest struct {
	Slot        int        `json:"slot"`
	Filename    string     `json:"filename,omitempty"`
	ContentHash string     `json:"contentHash,omitempty"`
	BboxMin     [3]float64 `json:"bboxMin"`
	BboxMax     [3]float64 `json:"bboxMax"`
	Centroid    [3]float64 `json:"centroid"`
	Area        float64    `json:"area"`
	// PrincipalAxes are unit vectors ordered by decreasing spread;
	// HalfExtents[i] is the mesh's half-length along PrincipalAxes[i].
	PrincipalAxes [3][3]float64 `json:"principalAxes"`
	HalfExtents   [3]float64    `json:"halfExtents"`
	// Thickness/Clearance summarize hodge's per-vertex analysis maps when
	// the asset carries them (STEP imports annotated at conversion): wall
	// thickness through material and clearance through air, in mm.
	Thickness *MapRangeDigest `json:"thickness,omitempty"`
	Clearance *MapRangeDigest `json:"clearance,omitempty"`
}

// MapRangeDigest is the whole-mesh summary of one analysis map: the finite
// extremes, where they occur (mesh frame), and the fraction of mapped
// vertices with a finite value.
type MapRangeDigest struct {
	Min      float64    `json:"min"`
	MinAt    [3]float64 `json:"minAt"`
	Max      float64    `json:"max"`
	MaxAt    [3]float64 `json:"maxAt"`
	Coverage float64    `json:"coverage"`
}

// Mesh-maps is the ai-agent -> canvas request that reads hodge's per-vertex
// analysis maps on an imported mesh near given points: the canvas answers
// from the AnnotatedMesh proto in its per-client asset registry, so the
// (large) annotated payload never crosses the bus.
const (
	RequestPathMeshMaps  = "canvas/mesh-maps/request"
	ResponsePathMeshMaps = "canvas/mesh-maps/response"
)

type MeshMapsRequest struct {
	ID string `json:"id"`
	// ContentHash identifies the asset by SHA-256 hash.
	ContentHash string `json:"contentHash"`
	// Points are query positions. Without Placement they are in the mesh's
	// own import frame (the coordinates bvhSdf(slot) and the digest bbox
	// use). With Placement they are WORLD-space, and the returned nearest
	// point and normal come back in world space too.
	Points [][3]float64 `json:"points"`
	// Placement, when set, is the 4x4 world-from-mesh transform (column-
	// major, 16 values) the part is positioned by — the same transform
	// applied in the document. It moves the query into the mesh frame and
	// the results back into world, so the caller never transforms by hand.
	// Empty means the query and results are already in the mesh frame.
	Placement []float64 `json:"placement,omitempty"`
}

type MeshMapsResponse struct {
	ID      string          `json:"id"`
	Samples []MeshMapSample `json:"samples,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// MeshMapSample is one query point's local surface reading. Thickness and
// Clearance are nil where the mesh carries no map or hodge found no
// opposite wall (its +inf sentinel — JSON cannot carry infinities).
type MeshMapSample struct {
	Point     [3]float64 `json:"point"`
	Nearest   [3]float64 `json:"nearest"`
	Distance  float64    `json:"distance"`
	Normal    [3]float64 `json:"normal"`
	Face      string     `json:"face,omitempty"`
	Thickness *float64   `json:"thickness,omitempty"`
	Clearance *float64   `json:"clearance,omitempty"`
}

type RaycastRequest struct {
	ID string  `json:"id"`
	X  float32 `json:"x"`
	Y  float32 `json:"y"`
}

// VdbExtractRequest asks the fidget agent to bake a mesh into an OpenVDB
// narrow-band level set (or reuse a cached one) and return a uniform
// dense sample grid.
type VdbExtractRequest struct {
	ID             string  `json:"id"`
	AssetID        uint32  `json:"assetID"`
	AssetHash      string  `json:"assetHash,omitempty"`
	VerticesB64    string  `json:"verticesB64"`
	VoxelSize      float32 `json:"voxelSize"`
	HalfBandVoxels float32 `json:"halfBandVoxels"`
	DimsX          uint32  `json:"dimsX"`
	DimsY          uint32  `json:"dimsY"`
	DimsZ          uint32  `json:"dimsZ"`
}

// VdbExtractResponse carries the packed dense-grid bytes back to the
// canvas agent. Layout matches fidget.ExtractVdbDenseBytes:
// dims[3] u32 + bbox[6] f32 + values f32[dimsX*dimsY*dimsZ], all
// little-endian. BytesB64 is base64 because the websocket wire format
// is JSON-only and dispatcher events reuse the same encoding.
//
// AssetHash, VoxelSize, and HalfBandVoxels are echoed so the canvas cache
// can distinguish same-slot meshes and explicit vdbSdf parameters.
type VdbExtractResponse struct {
	ID             string  `json:"id"`
	AssetID        uint32  `json:"assetID"`
	AssetHash      string  `json:"assetHash,omitempty"`
	VoxelSize      float32 `json:"voxelSize,omitempty"`
	HalfBandVoxels float32 `json:"halfBandVoxels,omitempty"`
	BytesB64       string  `json:"bytesB64,omitempty"`
	Error          string  `json:"error,omitempty"`
}

// VdbProgressResponse is one progress tick fired during the OpenVDB
// bake. Total is always 1000 (per-mille).
type VdbProgressResponse struct {
	ID        string `json:"id"`
	AssetID   uint32 `json:"assetID"`
	AssetHash string `json:"assetHash,omitempty"`
	Completed uint32 `json:"completed"`
	Total     uint32 `json:"total"`
}
