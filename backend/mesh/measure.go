// Package mesh generates and measures binary STL meshes.
package mesh

import (
	"encoding/binary"
	"fmt"
	"math"
)

// STLStats are geometry measurements derived from a binary STL: the tight
// vertex bounding box, total surface area, enclosed volume (divergence-
// theorem sum of signed tetrahedra, so it is exact for closed meshes), and
// the volume centroid.
type STLStats struct {
	TriangleCount int
	Bbox          [6]float32 // xmin, ymin, zmin, xmax, ymax, zmax
	Volume        float64    // the enclosed volume's magnitude
	// SignedVolume is the enclosed volume with its sign: positive when the
	// triangles face outward, negative when the mesh is inside out.
	SignedVolume float64
	SurfaceArea  float64
	Center       [3]float32
}

// Empty reports whether the mesh had no triangles.
func (s STLStats) Empty() bool {
	return s.TriangleCount == 0
}

// Extents returns the bounding box size along each axis.
func (s STLStats) Extents() [3]float32 {
	return [3]float32{
		s.Bbox[3] - s.Bbox[0],
		s.Bbox[4] - s.Bbox[1],
		s.Bbox[5] - s.Bbox[2],
	}
}

// MeasureSTL computes STLStats from binary STL bytes.
func MeasureSTL(stl []byte) (STLStats, error) {
	var s STLStats
	if len(stl) < 84 {
		return s, fmt.Errorf("stl too short: %d bytes", len(stl))
	}
	n := int(binary.LittleEndian.Uint32(stl[80:84]))
	if n < 0 || len(stl) < 84+n*50 {
		return s, fmt.Errorf("stl truncated: header claims %d triangles for %d bytes", n, len(stl))
	}
	s.TriangleCount = n
	if n == 0 {
		return s, nil
	}

	min := [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	max := [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	var signedVol, correction, area float64
	var cx, cy, cz float64
	var origin [3]float64

	for i := 0; i < n; i++ {
		off := 84 + i*50 + 12 // skip the 12-byte facet normal
		var v [3][3]float64
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				bits := binary.LittleEndian.Uint32(stl[off+(j*3+k)*4:])
				v[j][k] = float64(math.Float32frombits(bits))
				if math.IsNaN(v[j][k]) || math.IsInf(v[j][k], 0) {
					return STLStats{}, fmt.Errorf("triangle %d contains a non-finite vertex", i)
				}
			}
		}

		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				if v[j][k] < min[k] {
					min[k] = v[j][k]
				}
				if v[j][k] > max[k] {
					max[k] = v[j][k]
				}
			}
		}

		e1 := [3]float64{v[1][0] - v[0][0], v[1][1] - v[0][1], v[1][2] - v[0][2]}
		e2 := [3]float64{v[2][0] - v[0][0], v[2][1] - v[0][1], v[2][2] - v[0][2]}
		cross := [3]float64{
			e1[1]*e2[2] - e1[2]*e2[1],
			e1[2]*e2[0] - e1[0]*e2[2],
			e1[0]*e2[1] - e1[1]*e2[0],
		}
		area += 0.5 * math.Sqrt(cross[0]*cross[0]+cross[1]*cross[1]+cross[2]*cross[2])

		if i == 0 {
			origin = v[0]
		}
		for j := range v {
			for k := range v[j] {
				v[j][k] -= origin[k]
			}
		}
		// A nearby origin limits cancellation for translated parts.
		sv := (v[0][0]*(v[1][1]*v[2][2]-v[1][2]*v[2][1]) -
			v[0][1]*(v[1][0]*v[2][2]-v[1][2]*v[2][0]) +
			v[0][2]*(v[1][0]*v[2][1]-v[1][1]*v[2][0])) / 6.0
		y := sv - correction
		next := signedVol + y
		correction = (next - signedVol) - y
		signedVol = next

		// The tetra's centroid is (origin+v0+v1+v2)/4; accumulate the
		// volume-weighted sum for the overall centroid.
		cx += sv * (v[0][0] + v[1][0] + v[2][0]) / 4.0
		cy += sv * (v[0][1] + v[1][1] + v[2][1]) / 4.0
		cz += sv * (v[0][2] + v[1][2] + v[2][2]) / 4.0
	}

	for k := 0; k < 3; k++ {
		s.Bbox[k] = float32(min[k])
		s.Bbox[k+3] = float32(max[k])
	}
	s.SurfaceArea = area
	s.Volume = math.Abs(signedVol)
	s.SignedVolume = signedVol
	if math.Abs(signedVol) > 1e-9 {
		s.Center = [3]float32{
			float32(origin[0] + cx/signedVol),
			float32(origin[1] + cy/signedVol),
			float32(origin[2] + cz/signedVol),
		}
	} else {
		// Degenerate (open or sheet-like) mesh: fall back to the bbox center.
		for k := 0; k < 3; k++ {
			s.Center[k] = float32((min[k] + max[k]) / 2)
		}
	}
	return s, nil
}
