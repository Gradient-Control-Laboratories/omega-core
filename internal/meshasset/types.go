// Package meshasset holds shared types and helpers for server-side imported
// mesh assets — the triangle data, bounding boxes, baked distance-field grids,
// and the registry that caches them by session index + content hash.
//
// This package is intentionally independent of any specific compilation target.
// WGSL codegen consumes the baked SdfGrid; the (future) fidget oracle consumes
// the raw triangle list. Both live here so the wiring agents can share a single
// source of truth per imported mesh.
package meshasset

// Triangle is three vertices in object space. Winding is preserved as-is from
// the source file; callers that care about orientation (e.g. raycast signing)
// handle it explicitly rather than through a convention here.
type Triangle struct {
	V0, V1, V2 [3]float64
}

// TriMesh is a flat triangle soup plus its axis-aligned bounding box.
type TriMesh struct {
	Triangles []Triangle
	Bbox      Bbox
}

// Bbox is an axis-aligned bounding box. Min is the inclusive lower corner;
// Max is the inclusive upper corner. An empty/zero bbox has Min > Max on all
// axes and IsValid() returns false.
type Bbox struct {
	Min [3]float64
	Max [3]float64
}

// IsValid reports whether the bbox describes a non-degenerate volume.
func (b Bbox) IsValid() bool {
	for i := 0; i < 3; i++ {
		if b.Max[i] < b.Min[i] {
			return false
		}
	}
	return true
}

// Size returns Max - Min per axis.
func (b Bbox) Size() [3]float64 {
	return [3]float64{
		b.Max[0] - b.Min[0],
		b.Max[1] - b.Min[1],
		b.Max[2] - b.Min[2],
	}
}

// Padded returns a bbox inflated by `fraction` of its size on every axis
// (both sides). A fraction of 0.2 gives a 20% margin so domain warps that
// shift the query point slightly still land inside the baked volume.
func (b Bbox) Padded(fraction float64) Bbox {
	size := b.Size()
	out := b
	for i := 0; i < 3; i++ {
		pad := size[i] * fraction
		out.Min[i] -= pad
		out.Max[i] += pad
	}
	return out
}

// SdfGrid is a rank-3 scalar field sampled on a regular grid inside a bbox.
// Values are signed distances at each cell *center* (not corner). The grid is
// dense row-major with x fastest, then y, then z: index = x + y*Nx + z*Nx*Ny.
type SdfGrid struct {
	Dims   [3]int // Nx, Ny, Nz
	Bbox   Bbox   // the region covered by the grid (after any padding)
	Values []float32
}

// Len returns the total number of cells.
func (g SdfGrid) Len() int {
	return g.Dims[0] * g.Dims[1] * g.Dims[2]
}

// At returns the signed-distance value at grid coordinates (ix, iy, iz).
// No bounds check; callers must ensure indices are in range.
func (g SdfGrid) At(ix, iy, iz int) float32 {
	return g.Values[ix+iy*g.Dims[0]+iz*g.Dims[0]*g.Dims[1]]
}
