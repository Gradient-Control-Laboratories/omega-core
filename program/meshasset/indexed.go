package meshasset

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
)

// FromIndexed builds a TriMesh (flat triangle soup + bbox) from an indexed
// mesh: positions is flat xyz (len 3*N), tris is flat triangle indices
// (len 3*M). This is how the hodge AnnotatedMesh's HiresMesh (OpenCascade
// BRepMesh tessellation) enters the asset registry for manifold booleans,
// without the bloat of an STL round-trip. The ContentHash is a SHA-256 of the
// raw position+index bytes so re-imports of identical geometry dedup like STL.
func FromIndexed(positions []float32, tris []int32) (*TriMesh, ContentHash) {
	mesh := &TriMesh{
		Triangles: make([]Triangle, 0, len(tris)/3),
		Bbox:      Bbox{Min: [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}, Max: [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}},
	}
	n := int32(len(positions) / 3)
	at := func(i int32) [3]float64 {
		if i < 0 || i >= n {
			return [3]float64{}
		}
		return [3]float64{float64(positions[3*i]), float64(positions[3*i+1]), float64(positions[3*i+2])}
	}
	expand := func(p [3]float64) {
		for k := 0; k < 3; k++ {
			if p[k] < mesh.Bbox.Min[k] {
				mesh.Bbox.Min[k] = p[k]
			}
			if p[k] > mesh.Bbox.Max[k] {
				mesh.Bbox.Max[k] = p[k]
			}
		}
	}
	for t := 0; t+2 < len(tris); t += 3 {
		tri := Triangle{V0: at(tris[t]), V1: at(tris[t+1]), V2: at(tris[t+2])}
		mesh.Triangles = append(mesh.Triangles, tri)
		expand(tri.V0)
		expand(tri.V1)
		expand(tri.V2)
	}
	if len(mesh.Triangles) == 0 {
		mesh.Bbox = Bbox{} // leave invalid for an empty mesh
	}

	h := sha256.New()
	var buf [4]byte
	for _, v := range positions {
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		h.Write(buf[:])
	}
	for _, i := range tris {
		binary.LittleEndian.PutUint32(buf[:], uint32(i))
		h.Write(buf[:])
	}
	return mesh, ContentHash(hex.EncodeToString(h.Sum(nil)))
}
