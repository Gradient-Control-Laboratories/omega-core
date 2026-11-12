package mesh

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/fidget/convert"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/meshasset"
	"gradientcontrol.com/omega/core/program/noise/spinodoid"
	"gradientcontrol.com/omega/core/program/progress"
)

// MeshSdfOracle holds the triangle soup for a bvhSdf(AssetID) leaf.
//
// Deprecated: Use [meshasset.MeshSdfOracle] instead; will be removed in 0.2.0.
type MeshSdfOracle = meshasset.MeshSdfOracle

// VdbOracle holds the triangle soup and narrow-band level-set parameters for a vdbSdf(AssetID) leaf.
//
// Deprecated: Use [meshasset.VdbOracle] instead; will be removed in 0.2.0.
type VdbOracle = meshasset.VdbOracle

// Output holds the generated binary STL mesh bytes and extracted triangle soup.
type Output struct {
	STL            []byte
	Vertices       []float32
	Triangles      []uint32
	SpinodoidCalls []spinodoid.Call
	SpinodoidMs    int64
}

// BakedSpinodoid is a baked Gabor spinodoid table and the ID of its content.
//
// Deprecated: Use [spinodoid.Baked] instead; will be removed in 0.2.0.
type BakedSpinodoid = spinodoid.Baked

// Option configures 3D mesh generation.
type Option func(*options)

type options struct {
	subdivision int
	bbox        [6]float32
	hasBbox     bool
	algo        string
	time        float64
	meshOracles []meshasset.MeshSdfOracle
	vdbOracles  []meshasset.VdbOracle
	progressFn  progress.Callback
}

// Progress supplies a progress callback to report meshing progress.
func Progress(fn progress.Callback) Option {
	return func(o *options) {
		o.progressFn = fn
	}
}

// WithProgress supplies a progress callback to report meshing progress.
func WithProgress(fn progress.Callback) Option {
	return Progress(fn)
}

// Subdivision sets the octree subdivision depth (default: 4).
func Subdivision(level int) Option {
	return func(o *options) {
		o.subdivision = level
	}
}

// WithSubdivision sets the octree subdivision depth (default: 4).
func WithSubdivision(level int) Option {
	return Subdivision(level)
}

// Bbox sets the explicit world-space bounding box for meshing.
func Bbox(bbox [6]float32) Option {
	return func(o *options) {
		o.bbox = bbox
		o.hasBbox = true
	}
}

// WithBbox sets the explicit world-space bounding box for meshing.
// If omitted, the bounding box is automatically derived from the program's @scene or AST bounds.
func WithBbox(bbox [6]float32) Option {
	return Bbox(bbox)
}

// Algorithm selects the polygonization algorithm ("marching-cubes" or "mdc").
func Algorithm(algo string) Option {
	return func(o *options) {
		o.algo = algo
	}
}

// WithAlgorithm selects the polygonization algorithm ("marching-cubes" or "mdc").
func WithAlgorithm(algo string) Option {
	return Algorithm(algo)
}

// Time sets the animation timestamp for evaluating dynamic geometry (default: 0.0).
func Time(timeSec float64) Option {
	return func(o *options) {
		o.time = timeSec
	}
}

// WithTime sets the animation timestamp for evaluating dynamic geometry (default: 0.0).
func WithTime(timeSec float64) Option {
	return Time(timeSec)
}

// MeshOracles registers BVH mesh SDF oracles before meshing.
func MeshOracles(oracles []meshasset.MeshSdfOracle) Option {
	return func(o *options) {
		o.meshOracles = oracles
	}
}

// VdbOracles registers OpenVDB narrow-band level-set oracles before meshing.
func VdbOracles(oracles []meshasset.VdbOracle) Option {
	return func(o *options) {
		o.vdbOracles = oracles
	}
}

// DecodePackedTriangles unpacks base64+f32 triangle data (9 floats per triangle,
// no shared indexing) into flat (vertices, triangles) slices for oracle registration.
func DecodePackedTriangles(verticesB64 string) ([]float32, []uint32, error) {
	raw, err := base64.StdEncoding.DecodeString(verticesB64)
	if err != nil {
		return nil, nil, fmt.Errorf("decode verticesB64: %w", err)
	}
	if len(raw)%4 != 0 {
		return nil, nil, fmt.Errorf("vertex blob length %d not a multiple of 4 (f32)", len(raw))
	}
	floats := make([]float32, len(raw)/4)
	for i := range floats {
		bits := binary.LittleEndian.Uint32(raw[i*4 : i*4+4])
		floats[i] = math.Float32frombits(bits)
	}
	if len(floats)%9 != 0 {
		return nil, nil, fmt.Errorf("vertex count %d not a multiple of 9 floats (3 verts xyz per triangle)", len(floats))
	}
	vertCount := uint32(len(floats) / 3)
	indices := make([]uint32, vertCount)
	for i := range indices {
		indices[i] = uint32(i)
	}
	return floats, indices, nil
}

// RegisterMeshOracle uploads triangle data for a bvhSdf(assetID) leaf.
func RegisterMeshOracle(assetID uint32, vertices []float32, triangles []uint32) error {
	fidget.Lock()
	defer fidget.Unlock()
	return fidget.RegisterMeshOracle(assetID, vertices, triangles)
}

// RegisterVdbOracle bakes a narrow-band OpenVDB level set for vdbSdf(assetID).
func RegisterVdbOracle(assetID uint32, vertices []float32, triangles []uint32, voxelSize, halfBandVoxels float32) error {
	fidget.Lock()
	defer fidget.Unlock()
	return fidget.RegisterVdbOracle(assetID, vertices, triangles, voxelSize, halfBandVoxels)
}

// BakeAndExtractVdbWithProgress bakes a triangle soup into an OpenVDB level set,
// registers the oracle under assetID, samples it onto a uniform dense grid, and
// returns the packed dense-grid bytes.
func BakeAndExtractVdbWithProgress(
	assetID uint32,
	vertices []float32, triangles []uint32,
	voxelSize, halfBandVoxels float32,
	dimsX, dimsY, dimsZ uint32,
	progressFn func(completed, total uint32),
) ([]byte, error) {
	fidget.Lock()
	defer fidget.Unlock()
	return fidget.BakeAndExtractVdbWithProgress(assetID, vertices, triangles, voxelSize, halfBandVoxels, dimsX, dimsY, dimsZ, progressFn)
}

// ExtractVdbDenseBytes samples the registered VDB level set for assetID into a uniform dense grid.
func ExtractVdbDenseBytes(assetID, dimsX, dimsY, dimsZ uint32) ([]byte, error) {
	fidget.Lock()
	defer fidget.Unlock()
	return fidget.ExtractVdbDenseBytes(assetID, dimsX, dimsY, dimsZ)
}

// ExtractVdbLevelSet returns the level set registered under assetID as an encoded tape.VdbLevelSet message.
func ExtractVdbLevelSet(assetID uint32) ([]byte, error) {
	fidget.Lock()
	defer fidget.Unlock()
	return fidget.ExtractVdbLevelSet(assetID)
}

// BakeSpinodoids builds prog and returns the spinodoid table each of its gaborSpinodoid calls reads.
func BakeSpinodoids(prog *program.Program) (map[string]spinodoid.Baked, error) {
	if prog == nil || prog.Primary == nil {
		return nil, fmt.Errorf("program or primary package is nil")
	}
	var gslSource string
	if len(prog.Primary.Files) > 0 {
		gslSource = prog.Primary.Files[0].Source
	}
	d := convert.NewDelegate(gslSource)

	fidget.Lock()
	defer fidget.Unlock()
	defer fidget.Clear()

	if err := build.Build(d, convert.BuildCopy(prog), true); err != nil {
		return nil, build.ErrWithPkgSrc(prog, err)
	}
	return d.BakedSpinodoids(), nil
}

// Body represents a distinct solid body in a multi-body export.
type Body struct {
	Name string
	STL  []byte
}

// Mesh converts an assembled GSL program into a single watertight binary STL mesh.
// It automatically coordinates the underlying geometry engines:
//   - Uses the Fidget dual-contouring mesher for analytical SDF expressions.
//   - Uses the Manifold boolean engine for exact mesh booleans and imported mesh assets.
func Mesh(prog *program.Program, opts ...Option) ([]byte, error) {
	o := options{
		subdivision: 4,
		algo:        "marching-cubes",
		time:        0.0,
		bbox:        [6]float32{-10, -10, -10, 10, 10, 10},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	for _, mo := range o.meshOracles {
		if err := RegisterMeshOracle(mo.AssetID, mo.Vertices, mo.Triangles); err != nil {
			return nil, fmt.Errorf("mesh oracle %d register: %w", mo.AssetID, err)
		}
	}
	for _, vo := range o.vdbOracles {
		if err := RegisterVdbOracle(vo.AssetID, vo.Vertices, vo.Triangles, vo.VoxelSize, vo.HalfBandVoxels); err != nil {
			return nil, fmt.Errorf("vdb oracle %d register: %w", vo.AssetID, err)
		}
	}

	var internalProgress func(completed, total uint32, status string)
	if o.progressFn != nil {
		internalProgress = func(completed, total uint32, status string) {
			o.progressFn(status, completed, total)
		}
	}

	var stlBytes []byte
	var err error
	err = convert.ToTreesAtTime(prog, o.time, func(treeIndices []fidget.TreeIndex, _ []string) error {
		stlBytes, err = fidget.GenerateSTL(treeIndices, o.subdivision, o.bbox, o.algo, internalProgress)
		return err
	})
	return stlBytes, err
}

// MeshWithOutput meshes prog and returns the binary STL along with extracted triangle soup.
func MeshWithOutput(prog *program.Program, opts ...Option) (*Output, error) {
	stl, err := Mesh(prog, opts...)
	if err != nil {
		return nil, err
	}
	verts, tris, err := unpackSTLTriangles(stl)
	if err != nil {
		return nil, err
	}
	return &Output{
		STL:       stl,
		Vertices:  verts,
		Triangles: tris,
	}, nil
}

func unpackSTLTriangles(stl []byte) ([]float32, []uint32, error) {
	if len(stl) < 84 {
		return nil, nil, nil
	}
	n := int(binary.LittleEndian.Uint32(stl[80:84]))
	if n <= 0 || len(stl) < 84+n*50 {
		return nil, nil, nil
	}
	verts := make([]float32, 0, n*9)
	tris := make([]uint32, n*3)
	for i := 0; i < n; i++ {
		off := 84 + i*50 + 12
		for j := 0; j < 9; j++ {
			bits := binary.LittleEndian.Uint32(stl[off+j*4:])
			verts = append(verts, math.Float32frombits(bits))
		}
	}
	for i := range tris {
		tris[i] = uint32(i)
	}
	return verts, tris, nil
}

// MeshBodies converts an assembled program into discrete named bodies.
func MeshBodies(prog *program.Program, opts ...Option) ([]Body, error) {
	stl, err := Mesh(prog, opts...)
	if err != nil {
		return nil, err
	}
	return []Body{
		{
			Name: "default",
			STL:  stl,
		},
	}, nil
}
