package meshasset

import "math"

// Area is the total surface area of the triangle soup: the sum of the
// triangle areas |cross(v1-v0, v2-v0)| / 2. Winding-independent; degenerate
// triangles contribute zero.
func (m *TriMesh) Area() float64 {
	var sum float64
	for _, t := range m.Triangles {
		ux := t.V1[0] - t.V0[0]
		uy := t.V1[1] - t.V0[1]
		uz := t.V1[2] - t.V0[2]
		vx := t.V2[0] - t.V0[0]
		vy := t.V2[1] - t.V0[1]
		vz := t.V2[2] - t.V0[2]
		cx := uy*vz - uz*vy
		cy := uz*vx - ux*vz
		cz := ux*vy - uy*vx
		sum += math.Sqrt(cx*cx + cy*cy + cz*cz)
	}
	return sum / 2
}
