// Package meshrepair repairs the folds Manifold Dual Contouring leaves in a mesh, guided by the field it meshed.
//
// A fold is a pleat: a few triangles turned over against the surface, the mesh still closed and oriented, most of
// them crossing no other triangle. They come from the dual quads' triangulation and from vertices the octree's QEF
// placed past a neighbour, and they show as area the surface does not have. The repair finds them by the field: a
// triangle whose normal points against the field's gradient at its centroid is backward. Each vertex of a backward
// triangle is placed again, from the tangent planes the field gives at it and at its 1-ring (a QEF, held toward the
// ring's mean so a flat neighbourhood has an answer), and walked back onto the zero set by Newton steps. That repeats
// over the triangles the moves touched until none is backward, or for a fixed number of passes.
//
// Only vertex positions change: the triangles, their order and the welding stay as the mesher made them.
package meshrepair

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"
)

// Sample is the field's value and gradient at a point.
type Sample struct {
	Value, Dx, Dy, Dz float32
}

// Field evaluates the field the mesh was made from at the points (xs[i], ys[i], zs[i]), one Sample per point.
type Field func(xs, ys, zs []float32) ([]Sample, error)

// Stats is what Folds did.
type Stats struct {
	// Passes is the pass the repair ended on, the one that left the fewest backward triangles (0: the mesh as given).
	Passes int
	// BackwardBefore and BackwardAfter count the backward triangles in the mesh given and in the mesh returned.
	BackwardBefore, BackwardAfter int
	// Moved counts the vertices whose place in the mesh returned differs from the mesh given.
	Moved int
}

// The repair's settings: the passes it makes at most, the weight that holds a vertex's QEF toward its ring's mean,
// and the Newton walk onto the zero set (steps, and the |value| that counts as on it). A vertex whose walk does not
// get within the tolerance keeps its QEF place. The field is evaluated in f32, so the tolerance is one a value of a
// few ulps at the part's size can meet.
const (
	maxPasses        = 12
	qefWeight        = 0.1
	projectMaxIters  = 20
	projectTolerance = 1e-5
)

type vec [3]float64

func sub(a, b vec) vec           { return vec{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func add(a, b vec) vec           { return vec{a[0] + b[0], a[1] + b[1], a[2] + b[2]} }
func scale(a vec, s float64) vec { return vec{a[0] * s, a[1] * s, a[2] * s} }
func dot(a, b vec) float64       { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func cross(a, b vec) vec {
	return vec{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}
func norm(a vec) float64 { return math.Sqrt(dot(a, a)) }

// Folds repairs the folds of stl, a binary STL, against field and returns the repaired binary STL: the same
// header and triangles, vertices welded by exact position and moved, each facet normal the cross product of its
// edges as fidget writes it.
func Folds(stl []byte, field Field) ([]byte, Stats, error) {
	var stats Stats
	if len(stl) < 84 {
		return nil, stats, fmt.Errorf("fold repair: STL of %d bytes has no header", len(stl))
	}
	n := int(binary.LittleEndian.Uint32(stl[80:84]))
	if len(stl) < 84+50*n {
		return nil, stats, fmt.Errorf("fold repair: STL says %d triangles but holds %d bytes", n, len(stl))
	}
	m := weld(stl, n)
	r := &repair{mesh: m, field: field}
	if err := r.run(&stats); err != nil {
		return nil, stats, err
	}
	if stats.Passes == 0 {
		return stl, stats, nil
	}
	return m.write(stl[:80]), stats, nil
}

// mesh is a welded triangle list: all of the STL's triangles, and among them the ones with three distinct vertices
// (live), which alone count for backwardness and the 1-rings.
type mesh struct {
	verts []vec
	tris  [][3]int32
	live  []int32
	// ring[ringStart[v]:ringStart[v+1]] are v's neighbours, and vtris[vtriStart[v]:vtriStart[v+1]] the live
	// triangles at v.
	ringStart, ring []int32
	vtriStart, vtri []int32
}

func weld(stl []byte, n int) *mesh {
	m := &mesh{tris: make([][3]int32, n)}
	ids := make(map[[3]float32]int32, n/2)
	for i := 0; i < n; i++ {
		off := 84 + 50*i + 12
		for k := 0; k < 3; k++ {
			var p [3]float32
			for c := 0; c < 3; c++ {
				p[c] = math.Float32frombits(binary.LittleEndian.Uint32(stl[off+12*k+4*c:]))
			}
			id, ok := ids[p]
			if !ok {
				id = int32(len(m.verts))
				ids[p] = id
				m.verts = append(m.verts, vec{float64(p[0]), float64(p[1]), float64(p[2])})
			}
			m.tris[i][k] = id
		}
		t := m.tris[i]
		if t[0] != t[1] && t[1] != t[2] && t[0] != t[2] {
			m.live = append(m.live, int32(i))
		}
	}
	nv := len(m.verts)
	neighbours := make([][]int32, nv)
	m.vtriStart = make([]int32, nv+1)
	for _, ti := range m.live {
		t := m.tris[ti]
		for k := 0; k < 3; k++ {
			v := t[k]
			neighbours[v] = append(neighbours[v], t[(k+1)%3], t[(k+2)%3])
			m.vtriStart[v+1]++
		}
	}
	m.ringStart = make([]int32, nv+1)
	for v := range neighbours {
		slices.Sort(neighbours[v])
		neighbours[v] = slices.Compact(neighbours[v])
		m.ringStart[v+1] = m.ringStart[v] + int32(len(neighbours[v]))
		m.ring = append(m.ring, neighbours[v]...)
		m.vtriStart[v+1] += m.vtriStart[v]
	}
	m.vtri = make([]int32, m.vtriStart[nv])
	fill := slices.Clone(m.vtriStart[:nv])
	for _, ti := range m.live {
		for _, v := range m.tris[ti] {
			m.vtri[fill[v]] = ti
			fill[v]++
		}
	}
	return m
}

func (m *mesh) ringOf(v int32) []int32 { return m.ring[m.ringStart[v]:m.ringStart[v+1]] }
func (m *mesh) trisAt(v int32) []int32 { return m.vtri[m.vtriStart[v]:m.vtriStart[v+1]] }

// write is the mesh as a binary STL under header.
func (m *mesh) write(header []byte) []byte {
	buf := make([]byte, 84+50*len(m.tris))
	copy(buf, header)
	binary.LittleEndian.PutUint32(buf[80:], uint32(len(m.tris)))
	put := func(off int, x float64) { binary.LittleEndian.PutUint32(buf[off:], math.Float32bits(float32(x))) }
	for i, t := range m.tris {
		off := 84 + 50*i
		a, b, c := m.verts[t[0]], m.verts[t[1]], m.verts[t[2]]
		nrm := cross(sub(b, a), sub(c, a))
		for k := 0; k < 3; k++ {
			put(off+4*k, nrm[k])
		}
		for k, p := range [3]vec{a, b, c} {
			for d := 0; d < 3; d++ {
				put(off+12+12*k+4*d, p[d])
			}
		}
	}
	return buf
}

type repair struct {
	*mesh
	field Field
}

func (r *repair) samples(pts []vec) ([]Sample, error) {
	xs, ys, zs := make([]float32, len(pts)), make([]float32, len(pts)), make([]float32, len(pts))
	for i, p := range pts {
		xs[i], ys[i], zs[i] = float32(p[0]), float32(p[1]), float32(p[2])
	}
	s, err := r.field(xs, ys, zs)
	if err != nil {
		return nil, err
	}
	if len(s) != len(pts) {
		return nil, fmt.Errorf("fold repair: the field gave %d samples for %d points", len(s), len(pts))
	}
	return s, nil
}

func gradient(s Sample) vec { return vec{float64(s.Dx), float64(s.Dy), float64(s.Dz)} }

// classify marks each of list backward or not: its normal against the field's gradient at its centroid.
func (r *repair) classify(list []int32, backward []bool) error {
	if len(list) == 0 {
		return nil
	}
	pts := make([]vec, len(list))
	for k, ti := range list {
		t := r.tris[ti]
		pts[k] = scale(add(add(r.verts[t[0]], r.verts[t[1]]), r.verts[t[2]]), 1.0/3)
	}
	s, err := r.samples(pts)
	if err != nil {
		return err
	}
	for k, ti := range list {
		t := r.tris[ti]
		a := r.verts[t[0]]
		nrm := cross(sub(r.verts[t[1]], a), sub(r.verts[t[2]], a))
		backward[ti] = dot(nrm, gradient(s[k])) < 0
	}
	return nil
}

func (r *repair) run(stats *Stats) error {
	backward := make([]bool, len(r.tris))
	if err := r.classify(r.live, backward); err != nil {
		return err
	}
	count := func() (back []int32) {
		for _, ti := range r.live {
			if backward[ti] {
				back = append(back, ti)
			}
		}
		return back
	}
	back := count()
	stats.BackwardBefore = len(back)
	// A pass may turn more triangles over before a later one rights them, so every pass is made and the repair ends
	// on the one that left the fewest backward: the mesh as given, when none did better.
	start := slices.Clone(r.verts)
	best := slices.Clone(r.verts)
	bestCount, bestPass := len(back), 0
	for pass := 1; pass <= maxPasses && len(back) > 0; pass++ {
		var order []int32
		seen := make(map[int32]bool)
		for _, ti := range back {
			for _, v := range r.tris[ti] {
				if !seen[v] {
					seen[v] = true
					order = append(order, v)
				}
			}
		}
		slices.Sort(order)
		// Every new place is found from the positions before this pass, then all are moved at once.
		pts, err := r.qefPlace(order)
		if err != nil {
			return err
		}
		if pts, err = r.project(pts); err != nil {
			return err
		}
		r.keepApart(order, pts)
		touched := make(map[int32]bool)
		for k, v := range order {
			r.verts[v] = pts[k]
			for _, ti := range r.trisAt(v) {
				touched[ti] = true
			}
		}
		list := make([]int32, 0, len(touched))
		for ti := range touched {
			list = append(list, ti)
		}
		if err := r.classify(list, backward); err != nil {
			return err
		}
		back = count()
		if len(back) < bestCount {
			bestCount, bestPass = len(back), pass
			copy(best, r.verts)
		}
	}
	copy(r.verts, best)
	stats.Passes = bestPass
	stats.BackwardAfter = bestCount
	for v := range r.verts {
		if r.verts[v] != start[v] {
			stats.Moved++
		}
	}
	return nil
}

// keepApart leaves a vertex of order where it is when its new place, as the STL writes it (float32), is another
// vertex's: two vertices brought to one point, as both sides of a pinch walked onto the same spot of the zero set,
// would be welded into one and open the mesh.
func (r *repair) keepApart(order []int32, pts []vec) {
	key := func(p vec) [3]float32 { return [3]float32{float32(p[0]), float32(p[1]), float32(p[2])} }
	moving := make(map[int32]bool, len(order))
	for _, v := range order {
		moving[v] = true
	}
	taken := make(map[[3]float32]bool, len(r.verts))
	for v, p := range r.verts {
		if !moving[int32(v)] {
			taken[key(p)] = true
		}
	}
	for k, v := range order {
		if taken[key(pts[k])] {
			pts[k] = r.verts[v]
		}
		taken[key(pts[k])] = true
	}
}

// qefPlace places each vertex of order from the tangent planes the field gives at it and at each vertex of its
// 1-ring, held toward the ring's mean m with weight w: (Σ n nᵀ + w I) x = Σ n nᵀ p + w m, n the unit gradient at p.
// Two faces' planes pin a vertex to their edge, so a sharp edge stays sharp; a plane the field gives no gradient
// for is left out.
func (r *repair) qefPlace(order []int32) ([]vec, error) {
	idx := make(map[int32]int)
	var pts []vec
	need := func(u int32) {
		if _, ok := idx[u]; !ok {
			idx[u] = len(pts)
			pts = append(pts, r.verts[u])
		}
	}
	for _, v := range order {
		need(v)
		for _, u := range r.ringOf(v) {
			need(u)
		}
	}
	s, err := r.samples(pts)
	if err != nil {
		return nil, err
	}
	out := make([]vec, len(order))
	for k, v := range order {
		ring := r.ringOf(v)
		var mean vec
		for _, u := range ring {
			mean = add(mean, r.verts[u])
		}
		if len(ring) > 0 {
			mean = scale(mean, 1/float64(len(ring)))
		} else {
			mean = r.verts[v]
		}
		var A [3][3]float64
		var b vec
		plane := func(u int32) {
			g := gradient(s[idx[u]])
			l := norm(g)
			if !(l > 0) || math.IsInf(l, 0) {
				return
			}
			g = scale(g, 1/l)
			d := dot(g, r.verts[u])
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					A[i][j] += g[i] * g[j]
				}
				b[i] += g[i] * d
			}
		}
		plane(v)
		for _, u := range ring {
			plane(u)
		}
		for i := 0; i < 3; i++ {
			A[i][i] += qefWeight
			b[i] += qefWeight * mean[i]
		}
		out[k] = solve3(A, b)
	}
	return out, nil
}

// solve3 solves A x = b by Cramer's rule; A is symmetric positive definite here (the weight's w I sees to that).
func solve3(A [3][3]float64, b vec) vec {
	det := func(M [3][3]float64) float64 {
		return M[0][0]*(M[1][1]*M[2][2]-M[1][2]*M[2][1]) - M[0][1]*(M[1][0]*M[2][2]-M[1][2]*M[2][0]) +
			M[0][2]*(M[1][0]*M[2][1]-M[1][1]*M[2][0])
	}
	D := det(A)
	var x vec
	for c := 0; c < 3; c++ {
		M := A
		for i := 0; i < 3; i++ {
			M[i][c] = b[i]
		}
		x[c] = det(M) / D
	}
	return x
}

// project walks each point onto the zero set as gsl's surface query does (eval.SurfacePoint):
// p <- p - f ∇f / |∇f|², for at most projectMaxIters steps, stopping a point once |f| <= projectTolerance or the
// gradient is zero or the value not finite. A point that ends off the zero set keeps the place it was given.
func (r *repair) project(start []vec) ([]vec, error) {
	cur := slices.Clone(start)
	done := make([]bool, len(cur))
	var s []Sample
	for iter := 0; iter <= projectMaxIters; iter++ {
		var err error
		if s, err = r.samples(cur); err != nil {
			return nil, err
		}
		if iter == projectMaxIters {
			break
		}
		for i := range cur {
			if done[i] {
				continue
			}
			v := float64(s[i].Value)
			if math.Abs(v) <= projectTolerance {
				done[i] = true
				continue
			}
			g := gradient(s[i])
			g2 := dot(g, g)
			if g2 == 0 || math.IsNaN(v) || math.IsInf(v, 0) {
				done[i] = true
				continue
			}
			cur[i] = sub(cur[i], scale(g, v/g2))
		}
	}
	for i := range cur {
		if !(math.Abs(float64(s[i].Value)) <= projectTolerance) {
			cur[i] = start[i]
		}
	}
	return cur, nil
}
