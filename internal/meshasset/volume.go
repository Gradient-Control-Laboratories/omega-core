package meshasset

import "math"

// Volume is the volume enclosed by the triangle soup, via the divergence
// theorem: the sum of signed tetrahedron volumes dot(v0, cross(v1, v2)) / 6
// over the triangles. The sum is signed by winding, which the loaders
// preserve from the source (see Triangle); a consistently-wound closed
// surface yields the enclosed volume with a sign that depends only on the
// orientation convention, so the absolute value is returned. Translating the
// mesh does not change the result; an open surface yields the (meaningless)
// flux of its partial sum.
func (m *TriMesh) Volume() float64 {
	var sum float64
	for _, t := range m.Triangles {
		cx := t.V1[1]*t.V2[2] - t.V1[2]*t.V2[1]
		cy := t.V1[2]*t.V2[0] - t.V1[0]*t.V2[2]
		cz := t.V1[0]*t.V2[1] - t.V1[1]*t.V2[0]
		sum += t.V0[0]*cx + t.V0[1]*cy + t.V0[2]*cz
	}
	return math.Abs(sum) / 6
}
