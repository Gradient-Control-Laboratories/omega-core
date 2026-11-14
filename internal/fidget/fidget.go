package fidget

/*
// How the rust wrapper is linked is in link_dynamic.go (development builds)
// and link_static.go (the omega_static tag: release binaries).
#include <stdlib.h>  // include for C.CString
#include <stddef.h>  // include for size_t
#include <stdint.h>  // include for int64_t

typedef int32_t   c_int; //  Define c_int as int32_t for consistency with Go's int32

typedef uint64_t TreeIndex;
typedef uint64_t CanvasIndex;

TreeIndex new_x();
TreeIndex new_y();
TreeIndex new_z();
TreeIndex new_const(double value);
TreeIndex new_input(char* name, double value);

TreeIndex tree_square (TreeIndex arg1);
TreeIndex tree_floor  (TreeIndex arg1);
TreeIndex tree_ceil   (TreeIndex arg1);
TreeIndex tree_round  (TreeIndex arg1);
TreeIndex tree_sqrt   (TreeIndex arg1);
TreeIndex tree_max    (TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_min    (TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_compare(TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_modulo (TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_and    (TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_or     (TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_atan2  (TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_pow_int(TreeIndex arg1, int64_t exp);
TreeIndex tree_neg    (TreeIndex arg1);
TreeIndex tree_sin    (TreeIndex arg1);
TreeIndex tree_cos    (TreeIndex arg1);
TreeIndex tree_tan    (TreeIndex arg1);
TreeIndex tree_asin   (TreeIndex arg1);
TreeIndex tree_acos   (TreeIndex arg1);
TreeIndex tree_atan   (TreeIndex arg1);
TreeIndex tree_exp    (TreeIndex arg1);
TreeIndex tree_ln     (TreeIndex arg1);
TreeIndex tree_not    (TreeIndex arg1);
TreeIndex tree_abs    (TreeIndex arg1);

typedef struct {
	// compilation error code where 0 = success
	c_int code;

	// if code == 0, then this is the PNG image data
	// otherwise this is nil
	unsigned char *data;

	// number of data bytes
    size_t len;
} RenderResult;

TreeIndex tree_add(TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_sub(TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_mul(TreeIndex arg1, TreeIndex arg2);
TreeIndex tree_div(TreeIndex arg1, TreeIndex arg2);

TreeIndex tree_remap(TreeIndex target, TreeIndex x, TreeIndex y, TreeIndex z);

CanvasIndex  build_canvas(TreeIndex target);

// ProgressCallback is the C-ABI function pointer used by long-running
// FFI calls (mesh-out + VDB bake) to report incremental progress to
// the Go host. (completed, total) are unitless; the host decides what
// they map to (slice counts, per-mille, etc.).
typedef void (*ProgressCallback)(unsigned int completed, unsigned int total);

// Mesh-oracle (bvhSdf) registration. Triangles are passed as parallel arrays:
// vertices is vertex_count vec3<f32> packed flat (length = vertex_count * 3),
// triangles is tri_count uvec3 packed flat (length = tri_count * 3) of
// indices into the vertex array. Re-registering the same asset_id replaces
// the prior oracle. Returns 0 on success, negative on validation failure.
c_int     register_mesh_oracle(uint32_t asset_id, const float* vertices,
                               size_t vertex_count, const uint32_t* triangles,
                               size_t tri_count);
c_int     has_mesh_oracle(uint32_t asset_id);
TreeIndex tree_mesh_oracle(uint32_t asset_id);

// Noise-tile oracle (noiseHash3). register_noise_tile installs the baked
// TileSize^3 RGBA8 tile (len = TileSize^3 * 4); noise_hash3 evaluates the
// three RGB channels at a cell (host-side parity testing); tree_noise_channel
// returns a Tree leaf for one channel (0..2) at the implicit (x,y,z), over
// `octaves` tile lookups (1 or 2).
c_int     register_noise_tile(const uint8_t* bytes, size_t len);
void      noise_hash3(float x, float y, float z, float* out);
void      noise_hash3_octaves(float x, float y, float z, uint32_t octaves, float* out);
TreeIndex tree_noise_channel(uint32_t channel, uint32_t octaves);

// Cellular (Worley) noise as single oracles (the worley* builtins).
// tree_cellular evaluates
// it at a lattice coordinate for host-side parity testing; tree_cellular_f1
// returns a Tree leaf evaluating it at the implicit (x,y,z).
TreeIndex tree_cellular(uint32_t comp, uint32_t octaves);

// Gabor spinodoid oracles (gaborSpinodoid / gaborSpinodoidGradient).
// register_spinodoid installs a table baked by common/noise/spinodoid in a
// slot; tree_spinodoid returns a Tree leaf for one component of it (0 the
// distance, 1..3 the gradient) at the implicit (x,y,z).
c_int     register_spinodoid(uint32_t slot, const double* sites, size_t len,
                            const int64_t* origin, const uint32_t* size, const double* bounds, c_int periodic,
                             double spacing, double reach, double a,
                             double sigma, double base_beta, double level,
                             c_int local);
TreeIndex tree_spinodoid(uint32_t slot, uint32_t comp);
uint32_t  spinodoid_oracle_id(uint32_t slot, uint32_t comp);
uint32_t  hodge_spinodoid_oracle_id(uint32_t slot, uint32_t comp);

// hodge's field (the gaborSpinodoid builtins) over its own slots: a table of
// jittered kernels, periodic over hodge's own window or bounded to a driven
// field's domain.
c_int     register_hodge_spinodoid(uint32_t slot, const double* kernels, size_t len,
                             const int64_t* origin, const uint32_t* size, const double* bounds, c_int periodic,
                             size_t impulses, double cell_size, double a,
                             double sigma, double base_beta, double level,
                             c_int local);
TreeIndex tree_hodge_spinodoid(uint32_t slot, uint32_t comp);

// VDB-oracle (vdbSdf) registration. Same triangle layout as the mesh
// oracle, plus a voxel size (world units) and a half-band width (in
// voxels). The Rust side bakes a narrow-band level set via
// openvdb::tools::meshToLevelSet and caches it under asset_id. The
// underlying mesh-oracle registry uses asset_id | 0x8000_0000 to keep
// VDB and BVH oracles for the same asset from colliding.
c_int     register_vdb_oracle(uint32_t asset_id, const float* vertices,
                              size_t vertex_count, const uint32_t* triangles,
                              size_t tri_count, float voxel_size,
                              float half_band_voxels);
c_int     has_vdb_oracle(uint32_t asset_id);
TreeIndex tree_vdb_oracle(uint32_t asset_id);

// Sample the registered VDB level set into a uniform dense grid and
// return packed bytes (dims[3] u32 + bbox_min[3] f32 + bbox_max[3] f32 +
// values f32[dims_x*dims_y*dims_z]). Returns NULL on failure (asset not
// registered or zero dims). Caller must free with free_vdb_dense_bytes.
unsigned char* extract_vdb_dense_bytes(uint32_t asset_id, uint32_t dims_x,
                                       uint32_t dims_y, uint32_t dims_z,
                                       size_t* out_len);

// One-shot bake + extract that takes a progress callback so the host
// can drive a determinate progress bar across both the OpenVDB
// voxelization (0..=500 per-mille) and the dense-grid sample loop
// (501..=1000). Concurrent invocations serialize on a Rust-side mutex.
unsigned char* bake_and_extract_vdb_with_progress(
    uint32_t asset_id,
    const float* vertices, size_t vertex_count,
    const uint32_t* triangles, size_t tri_count,
    float voxel_size, float half_band_voxels,
    uint32_t dims_x, uint32_t dims_y, uint32_t dims_z,
    ProgressCallback progress_cb,
    size_t* out_len);

// The level set registered under asset_id as an encoded tape.VdbLevelSet
// message. Returns NULL when none is registered or it cannot be written out.
// Caller must free with free_vdb_dense_bytes.
unsigned char* extract_vdb_level_set(uint32_t asset_id, size_t* out_len);

void           free_vdb_dense_bytes(unsigned char* ptr, size_t len);

// Bulk-evaluate a canvas at count points. out receives 4*count floats,
// written as (v, dx, dy, dz) per point. Returns 0 on success, negative
// on error.
c_int canvas_eval_grad(CanvasIndex canvas, size_t count, const float* xs,
                       const float* ys, const float* zs, float* out);
// canvas_eval_grad with per-point values for nvars input trees: var_values
// holds nvars*count floats, var-major.
c_int canvas_eval_grad_vars(CanvasIndex canvas, size_t count, const float* xs,
                            const float* ys, const float* zs, size_t nvars,
                            const uint64_t* var_trees, const float* var_values, float* out);

// Bulk interval-bound a canvas over `count` axis-aligned boxes. `boxes` holds
// 6 floats per box (xmin,ymin,zmin,xmax,ymax,zmax); `out` must hold 2*count
// floats, each box written as (lower, upper). Returns 0 on success, negative
// on error.
c_int canvas_eval_interval(CanvasIndex canvas, size_t count, const float* boxes, float* out);
c_int canvas_eval_interval_vars(CanvasIndex canvas, size_t count, const float* boxes, size_t nvars, const uint64_t* var_trees, const float* var_bounds, float* out);

char* canvas_codegen_glsl(CanvasIndex canvas_index);
void  tape_free(char*);

RenderResult canvas_render_png(CanvasIndex canvas, float centerX, float centerY, float centerZ, float scale, float yaw, float pitch);
RenderResult canvas_mesh_stl(CanvasIndex canvas, int depth, float bboxMinX, float bboxMinY, float bboxMinZ, float bboxMaxX, float bboxMaxY, float bboxMaxZ, ProgressCallback progress_cb);
RenderResult canvas_mesh_stl_marching_cubes(CanvasIndex canvas, int depth, float bboxMinX, float bboxMinY, float bboxMinZ, float bboxMaxX, float bboxMaxY, float bboxMaxZ, ProgressCallback progress_cb);
void         render_result_free(RenderResult result);

// Serialize the given output trees into an omega tape (tape::Tape protobuf).
// The returned bytes are the post-lowering fidget-tree DAG; free with
// render_result_free like the mesh/PNG results.
RenderResult serialize_tape(const TreeIndex* outputs, size_t count);

// serialize_tape with each output's colour (3) and material (12) trees, 15
// indices per output, UINT64_MAX for one it lacks, and a displaced flag each.
RenderResult serialize_tape_appearance(const TreeIndex* outputs, size_t count, const TreeIndex* appearance, const uint8_t* displaced);

// Bridge function defined in fidget_progress.go
extern void goProgressCallback(unsigned int completed, unsigned int total);

void clear();

*/
import "C"
import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	"gradientcontrol.com/omega/core/program/noise"
	"gradientcontrol.com/omega/core/program/noise/spinodoid"
)

var mu sync.Mutex

// This current implementation wrapping the fidget library is single-threaded.
// You MUST lock the mutex before calling any of the FFI functions.
func Lock() {
	mu.Lock()
}

func Unlock() {
	mu.Unlock()
}

type TreeIndex C.uint64_t
type CanvasIndex C.uint64_t

func NewX() TreeIndex {
	return TreeIndex(C.new_x())
}

func NewY() TreeIndex {
	return TreeIndex(C.new_y())
}

func NewZ() TreeIndex {
	return TreeIndex(C.new_z())
}

func NewConst(value float64) TreeIndex {
	t := TreeIndex(C.new_const(C.double(value)))
	constValues[t] = value
	return t
}

func NewInput(name string, value float64) TreeIndex {
	return TreeIndex(C.new_input(C.CString(name), C.double(value)))
}

func Square(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, func(a float64) float64 { return a * a }); ok {
		return t
	}
	return TreeIndex(C.tree_square(C.TreeIndex(arg1)))
}

func Floor(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Floor); ok {
		return t
	}
	return TreeIndex(C.tree_floor(C.TreeIndex(arg1)))
}

func Ceil(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Ceil); ok {
		return t
	}
	return TreeIndex(C.tree_ceil(C.TreeIndex(arg1)))
}

func Round(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Round); ok {
		return t
	}
	return TreeIndex(C.tree_round(C.TreeIndex(arg1)))
}

func Sqrt(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Sqrt); ok {
		return t
	}
	return TreeIndex(C.tree_sqrt(C.TreeIndex(arg1)))
}

func Max(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, rustMax); ok {
		return t
	}
	return TreeIndex(C.tree_max(C.TreeIndex(arg1), C.TreeIndex(arg2)))
}

func Min(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, rustMin); ok {
		return t
	}
	return TreeIndex(C.tree_min(C.TreeIndex(arg1), C.TreeIndex(arg2)))
}

func Compare(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, foldCompare); ok {
		return t
	}
	return TreeIndex(C.tree_compare(C.TreeIndex(arg1), C.TreeIndex(arg2)))
}

func Modulo(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, foldMod); ok {
		return t
	}
	return TreeIndex(C.tree_modulo(C.TreeIndex(arg1), C.TreeIndex(arg2)))
}

func And(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, foldAnd); ok {
		return t
	}
	return TreeIndex(C.tree_and(C.TreeIndex(arg1), C.TreeIndex(arg2)))
}

func Or(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, foldOr); ok {
		return t
	}
	return TreeIndex(C.tree_or(C.TreeIndex(arg1), C.TreeIndex(arg2)))
}

func Atan2(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, math.Atan2); ok {
		return t
	}
	return TreeIndex(C.tree_atan2(C.TreeIndex(arg1), C.TreeIndex(arg2)))
}

func PowInt(arg TreeIndex, exp int64) TreeIndex {
	if t, ok := fold1(arg, func(a float64) float64 { return foldPowInt(a, exp) }); ok {
		return t
	}
	return TreeIndex(C.tree_pow_int(C.TreeIndex(arg), C.int64_t(exp)))
}

func Neg(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, func(a float64) float64 { return -a }); ok {
		return t
	}
	return TreeIndex(C.tree_neg(C.TreeIndex(arg1)))
}

func Sin(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Sin); ok {
		return t
	}
	return TreeIndex(C.tree_sin(C.TreeIndex(arg1)))
}

func Cos(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Cos); ok {
		return t
	}
	return TreeIndex(C.tree_cos(C.TreeIndex(arg1)))
}

func Tan(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Tan); ok {
		return t
	}
	return TreeIndex(C.tree_tan(C.TreeIndex(arg1)))
}

func Asin(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Asin); ok {
		return t
	}
	return TreeIndex(C.tree_asin(C.TreeIndex(arg1)))
}

func Acos(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Acos); ok {
		return t
	}
	return TreeIndex(C.tree_acos(C.TreeIndex(arg1)))
}

func Atan(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Atan); ok {
		return t
	}
	return TreeIndex(C.tree_atan(C.TreeIndex(arg1)))
}

func Exp(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Exp); ok {
		return t
	}
	return TreeIndex(C.tree_exp(C.TreeIndex(arg1)))
}

func Ln(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Log); ok {
		return t
	}
	return TreeIndex(C.tree_ln(C.TreeIndex(arg1)))
}

func Not(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, foldNot); ok {
		return t
	}
	return TreeIndex(C.tree_not(C.TreeIndex(arg1)))
}

func Abs(arg1 TreeIndex) TreeIndex {
	if t, ok := fold1(arg1, math.Abs); ok {
		return t
	}
	return TreeIndex(C.tree_abs(C.TreeIndex(arg1)))
}

func Add(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, func(a, b float64) float64 { return a + b }); ok {
		return t
	}
	return TreeIndex(C.tree_add(C.TreeIndex(arg1), C.TreeIndex(C.TreeIndex(arg2))))
}

func Sub(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, func(a, b float64) float64 { return a - b }); ok {
		return t
	}
	return TreeIndex(C.tree_sub(C.TreeIndex(arg1), C.TreeIndex(arg2)))
}

func Mul(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, func(a, b float64) float64 { return a * b }); ok {
		return t
	}
	return TreeIndex(C.tree_mul(C.TreeIndex(arg1), C.TreeIndex(arg2)))
}

func Div(arg1 TreeIndex, arg2 TreeIndex) TreeIndex {
	if t, ok := fold2(arg1, arg2, func(a, b float64) float64 { return a / b }); ok {
		return t
	}
	return TreeIndex(C.tree_div(C.TreeIndex(arg1), C.TreeIndex(C.TreeIndex(arg2))))
}

func Remap(target TreeIndex, x TreeIndex, y TreeIndex, z TreeIndex) TreeIndex {
	// A constant does not depend on the axes it is remapped onto.
	if _, ok := constValues[target]; ok {
		return target
	}
	return TreeIndex(C.tree_remap(C.TreeIndex(target), C.TreeIndex(x), C.TreeIndex(y), C.TreeIndex(z)))
}

func BuildCanvas(target TreeIndex) CanvasIndex {
	return CanvasIndex(C.build_canvas(C.TreeIndex(target)))
}

// Generate a PNG file of the specified object centered on the specified world coordinates
// and with the specified scale (aka zoom factor - larger number to zoom out)
// yaw (rotation around the y-axis in radians),
// and pitch (rotation around the x-axis in radians 0 to pi)
func RenderPng(
	canvasIndex CanvasIndex,
	centerX float32, centerY float32, centerZ float32,
	scale float32, yaw float32, pitch float32,
) ([]byte, error) {
	result := C.canvas_render_png(C.CanvasIndex(canvasIndex), C.float(centerX), C.float(centerY), C.float(centerZ), C.float(scale), C.float(yaw), C.float(pitch))
	defer C.render_result_free(result)

	safeSlice := make([]byte, result.len)
	if result.len > 0 {

		// Create Go slice backed by C data.
		// The C data will be freed when this function exits.

		unsafeSlice := unsafe.Slice(result.data, result.len)

		// Copy the content out of unsafeSlice into an array that can be returned
		// so that the heap space used by unsafeSlice can be released

		for i := range int(result.len) {
			safeSlice[i] = byte(unsafeSlice[i])
		}
	}

	if result.code == 0 {
		return safeSlice, nil
	}

	return nil, fmt.Errorf("render failed: %d", result.code)
}

// Generate STL mesh of the specified object.
// `subdivisionLevel` is the fidget octree depth (3..12, clamped by the
// caller). `bbox` is the world-space [xmin, ymin, zmin, xmax, ymax,
// zmax] meshing region; the Rust side builds a per-axis world_to_model
// matrix from these six floats so anisotropic bboxes mesh tight (no
// cube padding around the bbox center).
func MeshStl(canvasIndex CanvasIndex, subdivisionLevel int, bbox [6]float32, progressFn ProgressFunc) ([]byte, error) {
	setProgressFunc(progressFn)
	defer setProgressFunc(nil)

	result := C.canvas_mesh_stl(C.CanvasIndex(canvasIndex), C.int(subdivisionLevel),
		C.float(bbox[0]), C.float(bbox[1]), C.float(bbox[2]),
		C.float(bbox[3]), C.float(bbox[4]), C.float(bbox[5]),
		C.ProgressCallback(C.goProgressCallback))
	defer C.render_result_free(result)

	safeSlice := make([]byte, result.len)
	if result.len > 0 {
		unsafeSlice := unsafe.Slice(result.data, result.len)
		for i := range int(result.len) {
			safeSlice[i] = byte(unsafeSlice[i])
		}
	}

	if result.code == 0 {
		return safeSlice, nil
	}

	return nil, fmt.Errorf("mesh stl failed: %d", result.code)
}

// ProgressFunc is called with (completed, total) slice counts during marching cubes.
type ProgressFunc func(completed, total uint32)

// Generate STL mesh using streaming marching cubes.
// `subdivisionLevel` controls grid resolution: n = 2^level + 1 grid
// points per axis. `bbox` is the world-space [xmin, ymin, zmin, xmax,
// ymax, zmax] meshing region; the Rust side pads it to a cube around
// the bbox center because the marching cubes grid is currently
// isotropic — for anisotropic regions, prefer MeshStl (MDC) which
// uses a per-axis world_to_model matrix.
func MeshStlMarchingCubes(canvasIndex CanvasIndex, subdivisionLevel int, bbox [6]float32, progressFn ProgressFunc) ([]byte, error) {
	// Store the progress function so the C callback can reach it
	setProgressFunc(progressFn)
	defer setProgressFunc(nil)

	result := C.canvas_mesh_stl_marching_cubes(C.CanvasIndex(canvasIndex), C.int(subdivisionLevel),
		C.float(bbox[0]), C.float(bbox[1]), C.float(bbox[2]),
		C.float(bbox[3]), C.float(bbox[4]), C.float(bbox[5]),
		C.ProgressCallback(C.goProgressCallback))
	defer C.render_result_free(result)

	safeSlice := make([]byte, result.len)
	if result.len > 0 {
		unsafeSlice := unsafe.Slice(result.data, result.len)
		for i := range int(result.len) {
			safeSlice[i] = byte(unsafeSlice[i])
		}
	}

	if result.code == 0 {
		return safeSlice, nil
	}

	return nil, fmt.Errorf("mesh stl marching cubes failed: %d", result.code)
}

// GradResult holds the field value and spatial gradient at one point.
type GradResult struct {
	Value      float32
	Dx, Dy, Dz float32
}

// EvalGrad bulk-evaluates the canvas at the given points, returning the
// value and spatial gradient at each. xs, ys, zs must be the same length.
func EvalGrad(canvasIndex CanvasIndex, xs, ys, zs []float32) ([]GradResult, error) {
	if len(xs) != len(ys) || len(xs) != len(zs) {
		return nil, fmt.Errorf("EvalGrad: mismatched point slice lengths %d/%d/%d", len(xs), len(ys), len(zs))
	}
	count := len(xs)
	if count == 0 {
		return nil, nil
	}
	out := make([]float32, 4*count)
	ret := C.canvas_eval_grad(C.CanvasIndex(canvasIndex), C.size_t(count),
		(*C.float)(unsafe.Pointer(&xs[0])),
		(*C.float)(unsafe.Pointer(&ys[0])),
		(*C.float)(unsafe.Pointer(&zs[0])),
		(*C.float)(unsafe.Pointer(&out[0])))
	if ret != 0 {
		return nil, fmt.Errorf("EvalGrad: rust returned %d", int(ret))
	}
	results := make([]GradResult, count)
	for i := range results {
		results[i] = GradResult{
			Value: out[4*i],
			Dx:    out[4*i+1],
			Dy:    out[4*i+2],
			Dz:    out[4*i+3],
		}
	}
	return results, nil
}

// EvalGradVars is EvalGrad with per-point input values: each listed input
// tree takes varValues[v][i] at point i instead of its current value, so a
// program can read a quantity that changes from point to point without being
// a function of position (a renderer's pixel footprint, which pixelSize()
// stands for). Unlisted inputs keep their current values.
func EvalGradVars(canvasIndex CanvasIndex, xs, ys, zs []float32, varTrees []TreeIndex, varValues [][]float32) ([]GradResult, error) {
	if len(xs) != len(ys) || len(xs) != len(zs) {
		return nil, fmt.Errorf("EvalGradVars: mismatched point slice lengths %d/%d/%d", len(xs), len(ys), len(zs))
	}
	count := len(xs)
	if count == 0 {
		return nil, nil
	}
	nvars := len(varTrees)
	if nvars == 0 {
		return EvalGrad(canvasIndex, xs, ys, zs)
	}
	if len(varValues) != nvars {
		return nil, fmt.Errorf("EvalGradVars: %d vars but %d value rows", nvars, len(varValues))
	}
	trees := make([]uint64, nvars)
	values := make([]float32, nvars*count)
	for v, t := range varTrees {
		trees[v] = uint64(t)
		if len(varValues[v]) != count {
			return nil, fmt.Errorf("EvalGradVars: var %d has %d values, want %d", v, len(varValues[v]), count)
		}
		copy(values[v*count:(v+1)*count], varValues[v])
	}
	out := make([]float32, 4*count)
	ret := C.canvas_eval_grad_vars(C.CanvasIndex(canvasIndex), C.size_t(count),
		(*C.float)(unsafe.Pointer(&xs[0])),
		(*C.float)(unsafe.Pointer(&ys[0])),
		(*C.float)(unsafe.Pointer(&zs[0])),
		C.size_t(nvars),
		(*C.uint64_t)(unsafe.Pointer(&trees[0])),
		(*C.float)(unsafe.Pointer(&values[0])),
		(*C.float)(unsafe.Pointer(&out[0])))
	if ret != 0 {
		return nil, fmt.Errorf("EvalGradVars: rust returned %d", int(ret))
	}
	results := make([]GradResult, count)
	for i := range results {
		results[i] = GradResult{
			Value: out[4*i],
			Dx:    out[4*i+1],
			Dy:    out[4*i+2],
			Dz:    out[4*i+3],
		}
	}
	return results, nil
}

// IntervalResult is a sound [Lower, Upper] bracket of the field over a box:
// the field's true range over the box lies within it.
type IntervalResult struct {
	Lower, Upper float32
}

// EvalInterval bulk interval-bounds the canvas over the given axis-aligned
// boxes, returning a sound (Lower, Upper) bracket per box via oracle-aware
// interval arithmetic. Each box is 6 floats [xMin,yMin,zMin,xMax,yMax,zMax].
func EvalInterval(canvasIndex CanvasIndex, boxes [][6]float32) ([]IntervalResult, error) {
	count := len(boxes)
	if count == 0 {
		return nil, nil
	}
	flat := make([]float32, 6*count)
	for i, b := range boxes {
		copy(flat[6*i:6*i+6], b[:])
	}
	out := make([]float32, 2*count)
	ret := C.canvas_eval_interval(C.CanvasIndex(canvasIndex), C.size_t(count),
		(*C.float)(unsafe.Pointer(&flat[0])),
		(*C.float)(unsafe.Pointer(&out[0])))
	if ret != 0 {
		return nil, fmt.Errorf("EvalInterval: rust returned %d", int(ret))
	}
	results := make([]IntervalResult, count)
	for i := range results {
		results[i] = IntervalResult{Lower: out[2*i], Upper: out[2*i+1]}
	}
	return results, nil
}

// EvalIntervalVars is EvalInterval with per-variable interval overrides: each
// listed input tree evaluates over a [lo, hi] range instead of its current
// point value, PER BOX — varBounds is indexed [box][var] — so a search can
// subdivide parameter ranges alongside space. Unlisted input vars stay pinned
// at their current values.
func EvalIntervalVars(canvasIndex CanvasIndex, boxes [][6]float32, varTrees []TreeIndex, varBounds [][][2]float32) ([]IntervalResult, error) {
	count := len(boxes)
	if count == 0 {
		return nil, nil
	}
	if len(varBounds) != count {
		return nil, fmt.Errorf("EvalIntervalVars: %d boxes but %d bound rows", count, len(varBounds))
	}
	nvars := len(varTrees)
	if nvars == 0 {
		return EvalInterval(canvasIndex, boxes)
	}
	flat := make([]float32, 6*count)
	for i, b := range boxes {
		copy(flat[6*i:6*i+6], b[:])
	}
	trees := make([]uint64, nvars)
	for i, t := range varTrees {
		trees[i] = uint64(t)
	}
	bounds := make([]float32, 2*nvars*count)
	for i, row := range varBounds {
		if len(row) != nvars {
			return nil, fmt.Errorf("EvalIntervalVars: box %d has %d bounds, want %d", i, len(row), nvars)
		}
		for v, b := range row {
			o := 2 * (i*nvars + v)
			bounds[o], bounds[o+1] = b[0], b[1]
		}
	}
	out := make([]float32, 2*count)
	ret := C.canvas_eval_interval_vars(C.CanvasIndex(canvasIndex), C.size_t(count),
		(*C.float)(unsafe.Pointer(&flat[0])),
		C.size_t(len(trees)),
		(*C.uint64_t)(unsafe.Pointer(&trees[0])),
		(*C.float)(unsafe.Pointer(&bounds[0])),
		(*C.float)(unsafe.Pointer(&out[0])))
	if ret != 0 {
		return nil, fmt.Errorf("EvalIntervalVars: rust returned %d", int(ret))
	}
	results := make([]IntervalResult, count)
	for i := range results {
		results[i] = IntervalResult{Lower: out[2*i], Upper: out[2*i+1]}
	}
	return results, nil
}

// SerializeTape walks the given output trees and returns the serialized omega
// tape (a tape::Tape protobuf): the post-lowering fidget-tree DAG that the
// omega_tape Python module (and any other consumer) replays in-process. Must be
// called under Lock() while the trees are still live (i.e. inside the
// convert.ToTrees callback, before the deferred Clear()).
func SerializeTape(treeIndices []TreeIndex) ([]byte, error) {
	if len(treeIndices) == 0 {
		return nil, fmt.Errorf("no implicit shapes to serialize")
	}

	result := C.serialize_tape(
		(*C.TreeIndex)(unsafe.Pointer(&treeIndices[0])),
		C.size_t(len(treeIndices)),
	)
	defer C.render_result_free(result)

	safeSlice := make([]byte, result.len)
	if result.len > 0 {
		unsafeSlice := unsafe.Slice(result.data, result.len)
		for i := range int(result.len) {
			safeSlice[i] = byte(unsafeSlice[i])
		}
	}

	if result.code == 0 {
		return safeSlice, nil
	}

	return nil, fmt.Errorf("serialize tape failed: %d", result.code)
}

// Appearance is one output's colour (3 trees) and material (12 trees, rows
// flattened) for SerializeTapeWithAppearance; nil when the output has none.
type Appearance struct {
	Color     []TreeIndex
	Material  []TreeIndex
	Displaced bool
}

// SerializeTapeWithAppearance is SerializeTape with each output's colour and
// material trees in the tape's `appearance`. Same locking rules.
func SerializeTapeWithAppearance(treeIndices []TreeIndex, appearance []Appearance) ([]byte, error) {
	if len(treeIndices) == 0 {
		return nil, fmt.Errorf("no implicit shapes to serialize")
	}
	if len(appearance) != len(treeIndices) {
		return nil, fmt.Errorf("%d appearances for %d outputs", len(appearance), len(treeIndices))
	}
	const absent = ^TreeIndex(0)
	flat := make([]TreeIndex, 15*len(treeIndices))
	displaced := make([]uint8, len(treeIndices))
	for i, a := range appearance {
		row := flat[15*i : 15*i+15]
		for k := range row {
			row[k] = absent
		}
		if len(a.Color) == 3 {
			copy(row[0:3], a.Color)
		}
		if len(a.Material) == 12 {
			copy(row[3:15], a.Material)
		}
		if a.Displaced {
			displaced[i] = 1
		}
	}

	result := C.serialize_tape_appearance(
		(*C.TreeIndex)(unsafe.Pointer(&treeIndices[0])),
		C.size_t(len(treeIndices)),
		(*C.TreeIndex)(unsafe.Pointer(&flat[0])),
		(*C.uint8_t)(unsafe.Pointer(&displaced[0])),
	)
	defer C.render_result_free(result)

	safeSlice := make([]byte, result.len)
	if result.len > 0 {
		copy(safeSlice, unsafe.Slice((*byte)(unsafe.Pointer(result.data)), result.len))
	}
	if result.code == 0 {
		return safeSlice, nil
	}
	return nil, fmt.Errorf("serialize tape failed: %d", result.code)
}

// Return encode JSON containing the generated GLSL
func CodegenGLSL(canvasIndex CanvasIndex) string {
	strPtr := C.canvas_codegen_glsl(C.CanvasIndex(canvasIndex))
	defer C.tape_free(strPtr)

	if strPtr == nil {
		return "-- null --"
	}

	return C.GoString(strPtr)
}

// Dispose of all existing nodes and scopes
func Clear() {
	C.clear()
	clear(constValues)
}

// RegisterMeshOracle uploads triangle data for a bvhSdf(asset_id) leaf and
// builds the BSP that the VM evaluator consults. vertices is a flat slice of
// xyz triples (3*N values for N vertices); triangles is a flat slice of
// vertex indices (3*M values for M triangles). Re-registering the same id
// replaces the prior oracle. Returns nil on success.
func RegisterMeshOracle(assetID uint32, vertices []float32, triangles []uint32) error {
	if len(vertices)%3 != 0 {
		return fmt.Errorf("RegisterMeshOracle: vertices length %d not divisible by 3", len(vertices))
	}
	if len(triangles)%3 != 0 {
		return fmt.Errorf("RegisterMeshOracle: triangles length %d not divisible by 3", len(triangles))
	}
	vertexCount := len(vertices) / 3
	triCount := len(triangles) / 3
	if vertexCount == 0 || triCount == 0 {
		return fmt.Errorf("RegisterMeshOracle: empty mesh (verts=%d tris=%d)", vertexCount, triCount)
	}
	// CGo guarantees that the slice's backing memory stays alive across the
	// FFI call as long as we hold a Go-side reference, which we do.
	ret := C.register_mesh_oracle(
		C.uint32_t(assetID),
		(*C.float)(unsafe.Pointer(&vertices[0])),
		C.size_t(vertexCount),
		(*C.uint32_t)(unsafe.Pointer(&triangles[0])),
		C.size_t(triCount),
	)
	if ret != 0 {
		return fmt.Errorf("RegisterMeshOracle: rust returned %d", int(ret))
	}
	return nil
}

func init() {
	if err := RegisterNoiseTile(noise.Bake()); err != nil {
		panic(err)
	}
}

// RegisterNoiseTile installs the baked noise tile (from program/noise.Bake())
// for the noiseHash3 oracle. len(bytes) must be TileSize^3 * Channels.
func RegisterNoiseTile(bytes []byte) error {
	if len(bytes) == 0 {
		return fmt.Errorf("RegisterNoiseTile: empty tile")
	}
	ret := C.register_noise_tile((*C.uint8_t)(unsafe.Pointer(&bytes[0])), C.size_t(len(bytes)))
	if ret != 0 {
		return fmt.Errorf("RegisterNoiseTile: rust returned %d", int(ret))
	}
	return nil
}

// NoiseHash3 evaluates the three RGB channels of the noise hash at the integer
// cell (x, y, z). Mirrors common/noise.Hash3; used to verify the Rust port
// matches the Go reference on the same baked bytes.
func NoiseHash3(x, y, z float32) [3]float32 {
	var out [3]C.float
	C.noise_hash3(C.float(x), C.float(y), C.float(z), (*C.float)(unsafe.Pointer(&out[0])))
	return [3]float32{float32(out[0]), float32(out[1]), float32(out[2])}
}

// NoiseHash3Octaves is NoiseHash3 over a chosen octave count (1 or 2) — the same
// hash the one-octave oracle evaluates, reachable from the host so it can be
// checked against common/noise.Hash3Octaves.
func NoiseHash3Octaves(x, y, z float32, octaves uint32) [3]float32 {
	var out [3]C.float
	C.noise_hash3_octaves(C.float(x), C.float(y), C.float(z), C.uint32_t(octaves),
		(*C.float)(unsafe.Pointer(&out[0])))
	return [3]float32{float32(out[0]), float32(out[1]), float32(out[2])}
}

// NoiseChannelTree returns a Tree leaf evaluating channel (0..2) of the noise
// hash at the implicit (x, y, z) over octaves tile lookups (1 or 2, what the
// noiseHash3 builtin's octaves argument selects); the convert layer remaps the
// implicit coordinates to the cell.
func NoiseChannelTree(channel, octaves uint32) TreeIndex {
	return TreeIndex(C.tree_noise_channel(C.uint32_t(channel), C.uint32_t(octaves)))
}

// Cellular oracle components, matching tape_core::cellular::component. The
// numbering is part of the contract between the Go convert layer, the Rust
// oracles and the WGSL emitter.
const (
	WorleyF1           uint32 = 0
	WorleyF2           uint32 = 1
	WorleyFeatureX     uint32 = 2
	WorleyNearCellX    uint32 = 5
	WorleyFarCellX     uint32 = 8
	WorleyFarFeatureX  uint32 = 11
	WorleyWallFeatureX uint32 = 14
	WorleyWall         uint32 = 17
	// WorleyAxes is how many consecutive components a point-valued one spans.
	WorleyAxes uint32 = 3
)

// WorleyTree returns a Tree leaf evaluating one component of a cellular scan
// at the implicit (x, y, z); the convert layer remaps those to the lattice
// coordinate. One oracle call per component, with a tight 1-Lipschitz interval
// for the distances, in place of the 81 hash leaves the unrolled 27-neighbour
// scan used to produce.
func WorleyTree(comp, octaves uint32) TreeIndex {
	return TreeIndex(C.tree_cellular(C.uint32_t(comp), C.uint32_t(octaves)))
}

// Spinodoid oracle components, matching tape_core::spinodoid::component.
const (
	SpinodoidDistance uint32 = 0
	SpinodoidGradX    uint32 = 1
)

// RegisterSpinodoid installs a baked Gabor spinodoid table in slot, leveled
// at level and normalized by the local wavenumber when local is set.
func RegisterSpinodoid(slot uint32, t *spinodoid.Table, level float64, local bool) error {
	flat := t.Flat()
	if len(flat) == 0 {
		return fmt.Errorf("RegisterSpinodoid: empty table")
	}
	localFlag := C.c_int(0)
	if local {
		localFlag = 1
	}
	origin := [3]int64{int64(t.Origin[0]), int64(t.Origin[1]), int64(t.Origin[2])}
	size := [3]uint32{uint32(t.Size[0]), uint32(t.Size[1]), uint32(t.Size[2])}
	periodic := C.c_int(0)
	if t.Periodic {
		periodic = 1
	}
	ret := C.register_spinodoid(C.uint32_t(slot),
		(*C.double)(unsafe.Pointer(&flat[0])), C.size_t(len(flat)),
		(*C.int64_t)(unsafe.Pointer(&origin[0])), (*C.uint32_t)(unsafe.Pointer(&size[0])),
		(*C.double)(unsafe.Pointer(&t.Bounds[0])), periodic,
		C.double(t.Spacing), C.double(t.Reach), C.double(t.A),
		C.double(t.Sigma), C.double(t.BaseBeta), C.double(level), localFlag)
	if ret != 0 {
		return fmt.Errorf("RegisterSpinodoid: rust returned %d", int(ret))
	}
	return nil
}

// SpinodoidTree returns a Tree leaf evaluating one component of the table in
// slot at the implicit (x, y, z); the convert layer remaps those to the query
// point.
func SpinodoidTree(slot, comp uint32) TreeIndex {
	return TreeIndex(C.tree_spinodoid(C.uint32_t(slot), C.uint32_t(comp)))
}

// SpinodoidOracleID returns the mesh-oracle leaf id one component of slot
// carries — the id a tape's oracle payload has to be keyed to, since the leaf
// itself travels as a bare number. Asked of the oracle rather than recomputed
// here so there is one definition of the id layout.
func SpinodoidOracleID(slot, comp uint32) uint32 {
	return uint32(C.spinodoid_oracle_id(C.uint32_t(slot), C.uint32_t(comp)))
}

// HodgeSpinodoidOracleID is SpinodoidOracleID for hodge's slots, which are a
// separate id space.
func HodgeSpinodoidOracleID(slot, comp uint32) uint32 {
	return uint32(C.hodge_spinodoid_oracle_id(C.uint32_t(slot), C.uint32_t(comp)))
}

// RegisterHodgeSpinodoid installs a baked table of hodge's field in slot, on
// the same terms as RegisterSpinodoid. hodge's slots are its own, so the two
// fields never collide, and a document may read both.
func RegisterHodgeSpinodoid(slot uint32, t *spinodoid.HodgeTable, level float64, local bool) error {
	flat := t.Flat()
	if len(flat) == 0 {
		return fmt.Errorf("RegisterHodgeSpinodoid: empty table")
	}
	localFlag := C.c_int(0)
	if local {
		localFlag = 1
	}
	origin := [3]int64{int64(t.Origin[0]), int64(t.Origin[1]), int64(t.Origin[2])}
	size := [3]uint32{uint32(t.Size[0]), uint32(t.Size[1]), uint32(t.Size[2])}
	periodic := C.c_int(0)
	if t.Periodic {
		periodic = 1
	}
	ret := C.register_hodge_spinodoid(C.uint32_t(slot),
		(*C.double)(unsafe.Pointer(&flat[0])), C.size_t(len(flat)),
		(*C.int64_t)(unsafe.Pointer(&origin[0])), (*C.uint32_t)(unsafe.Pointer(&size[0])),
		(*C.double)(unsafe.Pointer(&t.Bounds[0])), periodic,
		C.size_t(t.Impulses), C.double(t.CellSize), C.double(t.A),
		C.double(t.Sigma), C.double(t.BaseBeta), C.double(level), localFlag)
	if ret != 0 {
		return fmt.Errorf("RegisterHodgeSpinodoid: rust returned %d", int(ret))
	}
	return nil
}

// HodgeSpinodoidTree is SpinodoidTree for a hodge table, over hodge's slots.
func HodgeSpinodoidTree(slot, comp uint32) TreeIndex {
	return TreeIndex(C.tree_hodge_spinodoid(C.uint32_t(slot), C.uint32_t(comp)))
}

// ConstValue reports the value of t if it is a constant tree.
func ConstValue(t TreeIndex) (float64, bool) {
	v, ok := constValues[t]
	return v, ok
}

// HasMeshOracle returns true if an oracle has been registered for assetID.
func HasMeshOracle(assetID uint32) bool {
	return C.has_mesh_oracle(C.uint32_t(assetID)) != 0
}

// MeshOracleTree returns a tree leaf that evaluates to the signed distance
// from the implicit (x, y, z) to the mesh registered under assetID.
func MeshOracleTree(assetID uint32) TreeIndex {
	return TreeIndex(C.tree_mesh_oracle(C.uint32_t(assetID)))
}

// RegisterVdbOracle bakes a narrow-band OpenVDB level set for vdbSdf(assetID)
// and caches it on the Rust side. vertices/triangles match the mesh-oracle
// layout (flat xyz / flat indices). voxelSize is the world-space cell size;
// halfBandVoxels is the narrow-band half-width in voxels (background magnitude
// is voxelSize * halfBandVoxels). Re-registering replaces the prior oracle.
func RegisterVdbOracle(assetID uint32, vertices []float32, triangles []uint32, voxelSize, halfBandVoxels float32) error {
	if len(vertices)%3 != 0 {
		return fmt.Errorf("RegisterVdbOracle: vertices length %d not divisible by 3", len(vertices))
	}
	if len(triangles)%3 != 0 {
		return fmt.Errorf("RegisterVdbOracle: triangles length %d not divisible by 3", len(triangles))
	}
	vertexCount := len(vertices) / 3
	triCount := len(triangles) / 3
	if vertexCount == 0 || triCount == 0 {
		return fmt.Errorf("RegisterVdbOracle: empty mesh (verts=%d tris=%d)", vertexCount, triCount)
	}
	if voxelSize <= 0 {
		return fmt.Errorf("RegisterVdbOracle: voxelSize must be > 0, got %f", voxelSize)
	}
	if halfBandVoxels <= 0 {
		return fmt.Errorf("RegisterVdbOracle: halfBandVoxels must be > 0, got %f", halfBandVoxels)
	}
	ret := C.register_vdb_oracle(
		C.uint32_t(assetID),
		(*C.float)(unsafe.Pointer(&vertices[0])),
		C.size_t(vertexCount),
		(*C.uint32_t)(unsafe.Pointer(&triangles[0])),
		C.size_t(triCount),
		C.float(voxelSize),
		C.float(halfBandVoxels),
	)
	if ret != 0 {
		return fmt.Errorf("RegisterVdbOracle: rust returned %d", int(ret))
	}
	return nil
}

// HasVdbOracle returns true if a VDB oracle has been registered for assetID.
func HasVdbOracle(assetID uint32) bool {
	return C.has_vdb_oracle(C.uint32_t(assetID)) != 0
}

// VdbOracleTree returns a tree leaf that evaluates to the signed distance
// from the implicit (x, y, z) to the OpenVDB level set registered under
// assetID.
func VdbOracleTree(assetID uint32) TreeIndex {
	return TreeIndex(C.tree_vdb_oracle(C.uint32_t(assetID)))
}

// ExtractVdbDenseBytes samples the registered VDB level set for assetID
// into a uniform dense grid at the requested dimensions and returns the
// packed binary representation:
//
//	bytes 0..12   : dims[3] u32 little-endian (x, y, z)
//	bytes 12..36  : bbox[6] f32 little-endian (min_x, min_y, min_z, max_x, max_y, max_z)
//	bytes 36..end : values f32 little-endian, row-major (x fastest then y then z)
//
// The returned slice is a copy of the rust-allocated buffer; the caller
// does not need to free anything. Returns an error if no VDB oracle is
// registered for assetID or any dimension is zero.
func ExtractVdbDenseBytes(assetID, dimsX, dimsY, dimsZ uint32) ([]byte, error) {
	if dimsX == 0 || dimsY == 0 || dimsZ == 0 {
		return nil, fmt.Errorf("ExtractVdbDenseBytes: dims must be > 0, got %d×%d×%d", dimsX, dimsY, dimsZ)
	}
	var outLen C.size_t
	ptr := C.extract_vdb_dense_bytes(
		C.uint32_t(assetID),
		C.uint32_t(dimsX), C.uint32_t(dimsY), C.uint32_t(dimsZ),
		&outLen,
	)
	if ptr == nil {
		return nil, fmt.Errorf("ExtractVdbDenseBytes: asset %d not registered or extraction failed", assetID)
	}
	defer C.free_vdb_dense_bytes(ptr, outLen)
	out := C.GoBytes(unsafe.Pointer(ptr), C.int(outLen))
	return out, nil
}

// ExtractVdbLevelSet returns the level set registered under assetID as an
// encoded tape.VdbLevelSet message: its leaves and block values, which a
// reader without OpenVDB samples to the same values.
func ExtractVdbLevelSet(assetID uint32) ([]byte, error) {
	var outLen C.size_t
	ptr := C.extract_vdb_level_set(C.uint32_t(assetID), &outLen)
	if ptr == nil {
		return nil, fmt.Errorf("ExtractVdbLevelSet: asset %d not registered or its tree cannot be written out", assetID)
	}
	defer C.free_vdb_dense_bytes(ptr, outLen)
	return C.GoBytes(unsafe.Pointer(ptr), C.int(outLen)), nil
}

// BakeAndExtractVdbWithProgress is the canvas-side preview entry point
// that bakes a triangle soup into an OpenVDB level set, registers the
// oracle, samples it onto a dense grid, and returns the packed bytes —
// all while ticking `progressFn` with (completed, total) values that
// span both phases (0..=500 from the OpenVDB Interrupter, 501..=1000
// from the dense extract). progressFn may be nil to skip progress.
//
// Concurrent invocations serialize on a Rust-side mutex so the
// callback channel is unique to the active bake.
func BakeAndExtractVdbWithProgress(
	assetID uint32,
	vertices []float32, triangles []uint32,
	voxelSize, halfBandVoxels float32,
	dimsX, dimsY, dimsZ uint32,
	progressFn ProgressFunc,
) ([]byte, error) {
	if len(vertices)%3 != 0 {
		return nil, fmt.Errorf("BakeAndExtractVdbWithProgress: vertices length %d not divisible by 3", len(vertices))
	}
	if len(triangles)%3 != 0 {
		return nil, fmt.Errorf("BakeAndExtractVdbWithProgress: triangles length %d not divisible by 3", len(triangles))
	}
	vertexCount := len(vertices) / 3
	triCount := len(triangles) / 3
	if vertexCount == 0 || triCount == 0 {
		return nil, fmt.Errorf("BakeAndExtractVdbWithProgress: empty mesh (verts=%d tris=%d)", vertexCount, triCount)
	}
	if voxelSize <= 0 || halfBandVoxels <= 0 {
		return nil, fmt.Errorf("BakeAndExtractVdbWithProgress: voxelSize and halfBandVoxels must be > 0 (got %g, %g)", voxelSize, halfBandVoxels)
	}
	if dimsX == 0 || dimsY == 0 || dimsZ == 0 {
		return nil, fmt.Errorf("BakeAndExtractVdbWithProgress: dims must be > 0, got %d×%d×%d", dimsX, dimsY, dimsZ)
	}

	setProgressFunc(progressFn)
	defer setProgressFunc(nil)

	var outLen C.size_t
	ptr := C.bake_and_extract_vdb_with_progress(
		C.uint32_t(assetID),
		(*C.float)(unsafe.Pointer(&vertices[0])),
		C.size_t(vertexCount),
		(*C.uint32_t)(unsafe.Pointer(&triangles[0])),
		C.size_t(triCount),
		C.float(voxelSize),
		C.float(halfBandVoxels),
		C.uint32_t(dimsX), C.uint32_t(dimsY), C.uint32_t(dimsZ),
		C.ProgressCallback(C.goProgressCallback),
		&outLen,
	)
	if ptr == nil {
		return nil, fmt.Errorf("BakeAndExtractVdbWithProgress: rust returned NULL (asset %d, %d×%d×%d)",
			assetID, dimsX, dimsY, dimsZ)
	}
	defer C.free_vdb_dense_bytes(ptr, outLen)
	return C.GoBytes(unsafe.Pointer(ptr), C.int(outLen)), nil
}
