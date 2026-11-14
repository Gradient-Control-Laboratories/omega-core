package mesh

// IndexedStats is what an indexed triangle mesh says about itself without a
// field to compare it to: its size, whether its surface is closed, and the
// volume its triangles enclose.
type IndexedStats struct {
	Vertices  int
	Triangles int
	// OpenEdges counts edges only one triangle uses: a hole's rim.
	OpenEdges int
	// NonManifoldEdges counts edges more than two triangles use.
	NonManifoldEdges int
	// SignedVolume is the divergence-theorem volume: positive for a closed
	// surface whose triangles face outward, negative when they face inward.
	// It means nothing for a surface with open edges.
	SignedVolume float64
}

// MeasureIndexed measures an indexed mesh: positions holds xyz per vertex and
// indices three vertex indices per triangle, as threemf.WeldSTL returns them.
// An edge is the unordered pair of its vertex indices, so vertices must be
// welded for two triangles to share one. Indices past the vertices and a
// trailing partial triangle are ignored.
func MeasureIndexed(positions []float32, indices []uint32) IndexedStats {
	s := IndexedStats{Vertices: len(positions) / 3}
	vertex := func(i uint32) [3]float64 {
		return [3]float64{float64(positions[3*i]), float64(positions[3*i+1]), float64(positions[3*i+2])}
	}
	uses := make(map[[2]uint32]int, len(indices)/2)
	var origin [3]float64
	var volume, correction float64
	for t := 0; t+2 < len(indices); t += 3 {
		tri := [3]uint32{indices[t], indices[t+1], indices[t+2]}
		if int(tri[0]) >= s.Vertices || int(tri[1]) >= s.Vertices || int(tri[2]) >= s.Vertices {
			continue
		}
		s.Triangles++
		for k := range tri {
			a, b := tri[k], tri[(k+1)%3]
			if a == b {
				continue
			}
			if b < a {
				a, b = b, a
			}
			uses[[2]uint32{a, b}]++
		}
		// Measured from the first vertex, so a part far from the origin does
		// not lose its volume to cancellation; summed with compensation, as
		// MeasureSTL sums it.
		v := [3][3]float64{vertex(tri[0]), vertex(tri[1]), vertex(tri[2])}
		if s.Triangles == 1 {
			origin = v[0]
		}
		for j := range v {
			for k := range v[j] {
				v[j][k] -= origin[k]
			}
		}
		term := (v[0][0]*(v[1][1]*v[2][2]-v[1][2]*v[2][1]) -
			v[0][1]*(v[1][0]*v[2][2]-v[1][2]*v[2][0]) +
			v[0][2]*(v[1][0]*v[2][1]-v[1][1]*v[2][0])) / 6
		y := term - correction
		next := volume + y
		correction = (next - volume) - y
		volume = next
	}
	for _, n := range uses {
		switch {
		case n == 1:
			s.OpenEdges++
		case n > 2:
			s.NonManifoldEdges++
		}
	}
	s.SignedVolume = volume
	return s
}
