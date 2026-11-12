// Package meshimport loads the meshes a GSL program imports with importMesh
// from files, so the distance builtins that read them (bvhSdf, vdbSdf) build:
// it reads each STL a slot names, and hands back the oracle payloads fidget
// registers before meshing or evaluating, and the dense distance grids the WGSL
// preview samples.
//
// The parser only accepts an importMesh path that is a filename beside the GSL
// file, so a Finder resolves a filename in its package's folder. Slots come
// from the program: the numbered `importMesh N "file.stl";` declarations and
// the assembler's table for the expression form (assembler.Result.Slots).
package meshimport

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gradientcontrol.com/omega/core/assembler/meshslots"
	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/meshasset"
)

// Finder returns the bytes of filename as the package at pkgPath imports it.
// pkgPath is "" for the primary package. A file that is not there is an error
// wrapping fs.ErrNotExist.
type Finder func(pkgPath, filename string) ([]byte, error)

// Dirs finds a file in the first directory that holds it: the primary
// package's files in dirs themselves, a dependency's in dir/pkgPath (the
// layout the source manager resolves packages by).
func Dirs(dirs ...string) Finder {
	return func(pkgPath, filename string) ([]byte, error) {
		var tried []string
		for _, dir := range dirs {
			p := filepath.Join(dir, filepath.FromSlash(pkgPath), filename)
			data, err := os.ReadFile(p)
			if err == nil {
				return data, nil
			}
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
			tried = append(tried, filepath.Dir(p))
		}
		return nil, fmt.Errorf("%s is not in %s: %w", filename, strings.Join(tried, ", "), fs.ErrNotExist)
	}
}

// Chain tries each finder in turn and returns the first file found.
func Chain(finders ...Finder) Finder {
	return func(pkgPath, filename string) ([]byte, error) {
		var errs []string
		for _, find := range finders {
			data, err := find(pkgPath, filename)
			if err == nil {
				return data, nil
			}
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
			errs = append(errs, err.Error())
		}
		return nil, fmt.Errorf("%s: %w", strings.Join(errs, "; "), fs.ErrNotExist)
	}
}

// Asset is one imported mesh, parsed.
type Asset struct {
	Filename string
	Hash     meshasset.ContentHash
	Mesh     *meshasset.TriMesh
	// Vertices and Triangles are the triangle soup in the layout fidget's
	// oracles take: three f32 per vertex, three vertices per triangle, and
	// triangle indices 0, 1, 2, ... in order.
	Vertices  []float32
	Triangles []uint32
}

// Use is one asset a distance builtin reads: the slot it names, which oracle,
// and (for vdbSdf, and for every preview grid) the level set's voxel size and
// half band, resolved from the call's arguments as meshasset.ResolveVdbParams
// does.
type Use struct {
	AssetIndex     int
	Kind           program.MeshSdfKind
	Asset          *Asset
	VoxelSize      float32
	HalfBandVoxels float32
}

// Builtin is the GSL name of the call that reads the slot.
func (u Use) Builtin() string {
	if u.Kind == program.MeshSdfKind_MESH_SDF_KIND_VDB {
		return "vdbSdf"
	}
	return "bvhSdf"
}

// References reports whether prog reads any imported mesh through a distance
// builtin; when it does not, nothing needs loading.
func References(prog *program.Program) bool {
	return prog != nil && len(prog.MeshSdfSlots) > 0
}

type declared struct {
	pkgPath, kind, path string
}

// Resolve reads every mesh prog's bvhSdf and vdbSdf calls name. slots is the
// assembler's table for the expression-form imports (assembler.Result.Slots);
// the numbered declarations come from prog itself. Uses come back in the order
// of prog.MeshSdfSlots.
func Resolve(prog *program.Program, slots []meshslots.Import, find Finder) ([]Use, error) {
	if !References(prog) {
		return nil, nil
	}
	table := declarations(prog, slots)
	uses := make([]Use, 0, len(prog.MeshSdfSlots))
	for _, s := range prog.MeshSdfSlots {
		idx := int(s.GetAssetIndex())
		u := Use{AssetIndex: idx, Kind: s.GetKind()}
		decl, ok := table[idx]
		if !ok {
			return nil, fmt.Errorf("%s(%d): no importMesh declares slot %d", u.Builtin(), idx, idx)
		}
		if decl.kind != "mesh" {
			return nil, fmt.Errorf("%s(%d): slot %d is %s %q; only importMesh .stl files load here",
				u.Builtin(), idx, idx, "import"+strings.ToUpper(decl.kind[:1])+decl.kind[1:], decl.path)
		}
		data, err := find(decl.pkgPath, decl.path)
		if err != nil {
			return nil, fmt.Errorf("%s(%d): importMesh %q: %w", u.Builtin(), idx, decl.path, err)
		}
		asset, err := parse(decl.path, data)
		if err != nil {
			return nil, fmt.Errorf("%s(%d): %w", u.Builtin(), idx, err)
		}
		u.Asset = asset
		u.VoxelSize, u.HalfBandVoxels = meshasset.ResolveVdbParams(asset.Mesh, int(s.GetVoxelDivisor()), s.GetHalfBandVoxels())
		uses = append(uses, u)
	}
	return uses, nil
}

// declarations maps each slot to the file it imports: numbered declarations
// in every package, then the expression-form table.
func declarations(prog *program.Program, slots []meshslots.Import) map[int]declared {
	table := map[int]declared{}
	primary := prog.GetPrimary().GetPath()
	pkgs := append([]*program.Package{prog.GetPrimary()}, prog.GetDependencies()...)
	for _, pkg := range pkgs {
		pkgPath := pkg.GetPath()
		if pkgPath == primary {
			pkgPath = ""
		}
		for _, f := range pkg.GetFiles() {
			for _, imp := range f.GetMeshImports() {
				table[int(imp.GetIndex())] = declared{pkgPath, "mesh", imp.GetPath()}
			}
			for _, imp := range f.GetStepImports() {
				table[int(imp.GetIndex())] = declared{pkgPath, "step", imp.GetPath()}
			}
		}
	}
	for _, s := range slots {
		pkgPath := s.PkgPath
		if pkgPath == primary {
			pkgPath = ""
		}
		table[s.Slot] = declared{pkgPath, s.Kind, s.Path}
	}
	return table
}

var (
	parsedMu sync.Mutex
	parsed   = map[meshasset.ContentHash]*Asset{}
)

// maxParsed bounds the parsed-mesh cache; a long-running server sees a new
// entry only when a file's contents change.
const maxParsed = 32

// parse parses an STL, reusing the parse of identical bytes.
func parse(filename string, data []byte) (*Asset, error) {
	hash := meshasset.HashBytes(data)
	parsedMu.Lock()
	a, ok := parsed[hash]
	parsedMu.Unlock()
	if ok {
		if a.Filename == filename {
			return a, nil
		}
		copied := *a
		copied.Filename = filename
		return &copied, nil
	}
	mesh, _, err := meshasset.ParseSTL(data)
	if err != nil {
		return nil, fmt.Errorf("importMesh %q: %w", filename, err)
	}
	if len(mesh.Triangles) == 0 {
		return nil, fmt.Errorf("importMesh %q: the STL has no triangles", filename)
	}
	a = &Asset{Filename: filename, Hash: hash, Mesh: mesh}
	a.Vertices = make([]float32, 0, 9*len(mesh.Triangles))
	for _, t := range mesh.Triangles {
		for _, v := range [3][3]float64{t.V0, t.V1, t.V2} {
			a.Vertices = append(a.Vertices, float32(v[0]), float32(v[1]), float32(v[2]))
		}
	}
	a.Triangles = make([]uint32, 3*len(mesh.Triangles))
	for i := range a.Triangles {
		a.Triangles[i] = uint32(i)
	}
	parsedMu.Lock()
	if len(parsed) >= maxParsed {
		parsed = map[meshasset.ContentHash]*Asset{}
	}
	parsed[hash] = a
	parsedMu.Unlock()
	return a, nil
}

// Oracles is the uses as the payloads fidget registers: a BVH oracle for
// every slot, and a level set for each slot vdbSdf reads. The program lists
// one MeshSdfSlot per asset, of kind VDB when vdbSdf reads it whether or not
// bvhSdf does too (textures.Extract), so a VDB slot gets the BVH oracle as
// well; building it costs little beside the level set's bake. Pass them to
// mesh.MeshOracles / mesh.VdbOracles, tape's options of the same names, or
// Register.
func Oracles(uses []Use) ([]meshasset.MeshSdfOracle, []meshasset.VdbOracle) {
	var bvh []meshasset.MeshSdfOracle
	var vdb []meshasset.VdbOracle
	seenBvh := map[int]bool{}
	seenVdb := map[int]bool{}
	for _, u := range uses {
		if !seenBvh[u.AssetIndex] {
			seenBvh[u.AssetIndex] = true
			bvh = append(bvh, meshasset.MeshSdfOracle{
				AssetID: uint32(u.AssetIndex), Vertices: u.Asset.Vertices, Triangles: u.Asset.Triangles,
			})
		}
		if u.Kind == program.MeshSdfKind_MESH_SDF_KIND_VDB && !seenVdb[u.AssetIndex] {
			seenVdb[u.AssetIndex] = true
			vdb = append(vdb, meshasset.VdbOracle{
				AssetID: uint32(u.AssetIndex), Vertices: u.Asset.Vertices, Triangles: u.Asset.Triangles,
				VoxelSize: u.VoxelSize, HalfBandVoxels: u.HalfBandVoxels,
			})
		}
	}
	return bvh, vdb
}

var (
	registeredMu sync.Mutex
	// registered is what each fidget oracle slot holds, so registering the
	// same mesh again (every query of a session) skips the BVH build or the
	// OpenVDB bake.
	registered = map[string]string{}
)

// Register registers the uses' oracles with fidget for the evaluation paths
// that build a program without taking oracle options (eval's queries, the png
// renderer). A slot already holding the same mesh, with the same level-set
// parameters, is left as it is.
func Register(uses []Use) error {
	bvh, vdb := Oracles(uses)
	registeredMu.Lock()
	defer registeredMu.Unlock()
	for _, o := range bvh {
		slot, key := fmt.Sprintf("bvh:%d", o.AssetID), string(hashOf(uses, int(o.AssetID)))
		if registered[slot] == key {
			continue
		}
		fidget.Lock()
		err := fidget.RegisterMeshOracle(o.AssetID, o.Vertices, o.Triangles)
		fidget.Unlock()
		if err != nil {
			return fmt.Errorf("bvhSdf(%d): %w", o.AssetID, err)
		}
		registered[slot] = key
	}
	for _, o := range vdb {
		slot := fmt.Sprintf("vdb:%d", o.AssetID)
		key := fmt.Sprintf("%s:%v:%v", hashOf(uses, int(o.AssetID)), o.VoxelSize, o.HalfBandVoxels)
		if registered[slot] == key {
			continue
		}
		fidget.Lock()
		err := fidget.RegisterVdbOracle(o.AssetID, o.Vertices, o.Triangles, o.VoxelSize, o.HalfBandVoxels)
		fidget.Unlock()
		if err != nil {
			return fmt.Errorf("vdbSdf(%d): %w", o.AssetID, err)
		}
		registered[slot] = key
	}
	return nil
}

func hashOf(uses []Use, assetIndex int) meshasset.ContentHash {
	for _, u := range uses {
		if u.AssetIndex == assetIndex {
			return u.Asset.Hash
		}
	}
	return ""
}

// Grid dimensions for the WGSL preview, as omega's canvas agent requests
// them: 256 cells along the mesh's longest axis, at least 32 along any.
const (
	gridLongest    = 256
	gridMinPerAxis = 32
)

// GridDims is the dense preview grid's cell count per axis for a mesh's box.
func GridDims(bbox meshasset.Bbox) [3]uint32 {
	size := bbox.Size()
	longest := math.Max(size[0], math.Max(size[1], size[2]))
	dims := [3]uint32{1, 1, 1}
	for i := 0; i < 3; i++ {
		d := int(math.Round(float64(gridLongest) * size[i] / longest))
		if d < gridMinPerAxis {
			d = gridMinPerAxis
		}
		dims[i] = uint32(d)
	}
	return dims
}

type gridKey struct {
	hash                meshasset.ContentHash
	voxelSize, halfBand float32
}

var (
	gridsMu sync.Mutex
	grids   = map[gridKey]*meshasset.SdfGrid{}
)

// maxGrids bounds the preview-grid cache (each is up to 256³ f32).
const maxGrids = 8

// Grids bakes the dense distance grid the WGSL preview samples for each asset
// the uses read, keyed by asset index: the mesh's OpenVDB level set at the
// use's voxel size and half band, sampled at GridDims cells, for bvhSdf and
// vdbSdf alike (the preview of both is the same grid; the oracles differ only
// when meshing). A grid already baked for the same mesh and parameters is
// reused.
//
// Baking registers the level set as fidget's vdbSdf oracle for the slot, so a
// caller sharing fidget across documents serializes Grids with its meshing.
func Grids(uses []Use) (map[int]*meshasset.SdfGrid, error) {
	if len(uses) == 0 {
		return nil, nil
	}
	out := make(map[int]*meshasset.SdfGrid, len(uses))
	for _, u := range uses {
		key := gridKey{u.Asset.Hash, u.VoxelSize, u.HalfBandVoxels}
		gridsMu.Lock()
		g, ok := grids[key]
		gridsMu.Unlock()
		if !ok {
			dims := GridDims(u.Asset.Mesh.Bbox)
			fidget.Lock()
			raw, err := fidget.BakeAndExtractVdbWithProgress(uint32(u.AssetIndex), u.Asset.Vertices, u.Asset.Triangles,
				u.VoxelSize, u.HalfBandVoxels, dims[0], dims[1], dims[2], func(uint32, uint32) {})
			fidget.Unlock()
			if err != nil {
				return nil, fmt.Errorf("%s(%d): baking the preview grid: %w", u.Builtin(), u.AssetIndex, err)
			}
			// The bake left the level set registered under this slot.
			registeredMu.Lock()
			registered[fmt.Sprintf("vdb:%d", u.AssetIndex)] = fmt.Sprintf("%s:%v:%v", u.Asset.Hash, u.VoxelSize, u.HalfBandVoxels)
			registeredMu.Unlock()
			g, err = DecodeDenseGrid(raw)
			if err != nil {
				return nil, fmt.Errorf("%s(%d): %w", u.Builtin(), u.AssetIndex, err)
			}
			gridsMu.Lock()
			if len(grids) >= maxGrids {
				grids = map[gridKey]*meshasset.SdfGrid{}
			}
			grids[key] = g
			gridsMu.Unlock()
		}
		out[u.AssetIndex] = g
	}
	return out, nil
}

// DecodeDenseGrid parses fidget's dense-grid bytes (ExtractVdbDenseBytes:
// dims[3] u32, bbox[6] f32, then the values f32, little-endian) into a grid
// whose box reaches half a cell past the first and last samples, since fidget
// samples on the box faces and SdfGrid values sit at cell centres.
func DecodeDenseGrid(raw []byte) (*meshasset.SdfGrid, error) {
	const headerLen = 3*4 + 6*4
	if len(raw) < headerLen {
		return nil, fmt.Errorf("dense grid: truncated header, %d bytes", len(raw))
	}
	u32 := func(off int) uint32 { return binary.LittleEndian.Uint32(raw[off : off+4]) }
	f32 := func(off int) float64 { return float64(math.Float32frombits(u32(off))) }
	dims := [3]int{int(u32(0)), int(u32(4)), int(u32(8))}
	total := dims[0] * dims[1] * dims[2]
	if len(raw) != headerLen+4*total {
		return nil, fmt.Errorf("dense grid: %d bytes do not match dims %d×%d×%d", len(raw), dims[0], dims[1], dims[2])
	}
	bbox := meshasset.Bbox{
		Min: [3]float64{f32(12), f32(16), f32(20)},
		Max: [3]float64{f32(24), f32(28), f32(32)},
	}
	for k, n := range dims {
		if n > 1 {
			half := 0.5 * (bbox.Max[k] - bbox.Min[k]) / float64(n-1)
			bbox.Min[k] -= half
			bbox.Max[k] += half
		}
	}
	values := make([]float32, total)
	for i := range values {
		values[i] = math.Float32frombits(u32(headerLen + 4*i))
	}
	return &meshasset.SdfGrid{Dims: dims, Bbox: bbox, Values: values}, nil
}
