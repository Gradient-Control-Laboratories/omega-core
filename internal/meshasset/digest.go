package meshasset

import "math"

// Digest summarizes a triangle mesh's placement and orientation: exact
// bounding box, area-weighted centroid, surface area, and the principal
// axes of its area-weighted spread (largest first) with the mesh's
// half-extent along each. Computed directly from the triangles — no
// distance-field probing involved.
type Digest struct {
	BboxMin  [3]float64
	BboxMax  [3]float64
	Centroid [3]float64
	Area     float64
	// PrincipalAxes are unit vectors ordered by decreasing spread;
	// Extents[i] is the mesh's half-length projected onto PrincipalAxes[i].
	PrincipalAxes [3][3]float64
	Extents       [3]float64
}

// ComputeDigest derives a Digest from the mesh. Returns a zero Digest for
// an empty mesh.
func ComputeDigest(mesh *TriMesh) Digest {
	var d Digest
	if mesh == nil || len(mesh.Triangles) == 0 {
		return d
	}

	d.BboxMin = [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	d.BboxMax = [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}

	// Area-weighted centroid over triangle centroids.
	totalArea := 0.0
	var weighted [3]float64
	for _, t := range mesh.Triangles {
		for _, v := range [][3]float64{t.V0, t.V1, t.V2} {
			for k := 0; k < 3; k++ {
				d.BboxMin[k] = math.Min(d.BboxMin[k], v[k])
				d.BboxMax[k] = math.Max(d.BboxMax[k], v[k])
			}
		}
		a := triangleArea(t)
		c := triangleCentroid(t)
		totalArea += a
		for k := 0; k < 3; k++ {
			weighted[k] += a * c[k]
		}
	}
	d.Area = totalArea
	if totalArea <= 0 {
		// Degenerate soup: fall back to the bbox center with axis-aligned axes.
		for k := 0; k < 3; k++ {
			d.Centroid[k] = 0.5 * (d.BboxMin[k] + d.BboxMax[k])
			d.PrincipalAxes[k][k] = 1
			d.Extents[k] = 0.5 * (d.BboxMax[k] - d.BboxMin[k])
		}
		return d
	}
	for k := 0; k < 3; k++ {
		d.Centroid[k] = weighted[k] / totalArea
	}

	// Exact area-weighted covariance of the surface about the centroid:
	// per triangle, ∫ x xᵀ dA = (A/12)·(m mᵀ + Σᵢ pᵢ pᵢᵀ) with m = Σᵢ pᵢ,
	// vertices taken relative to the centroid. Exact, so independent of how
	// the surface happens to be triangulated.
	var cov [3][3]float64
	for _, t := range mesh.Triangles {
		a := triangleArea(t)
		var p0, p1, p2, m [3]float64
		for k := 0; k < 3; k++ {
			p0[k] = t.V0[k] - d.Centroid[k]
			p1[k] = t.V1[k] - d.Centroid[k]
			p2[k] = t.V2[k] - d.Centroid[k]
			m[k] = p0[k] + p1[k] + p2[k]
		}
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				cov[i][j] += a / 12 * (m[i]*m[j] + p0[i]*p0[j] + p1[i]*p1[j] + p2[i]*p2[j])
			}
		}
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			cov[i][j] /= totalArea
		}
	}

	axes := symmetricEigenvectors(cov)
	// Eigenvector signs are arbitrary; make them deterministic (largest-
	// magnitude component positive) and the frame right-handed (flip the
	// third axis if needed) so the axes always form a proper rotation.
	for i := 0; i < 3; i++ {
		maxK := 0
		for k := 1; k < 3; k++ {
			if math.Abs(axes[i][k]) > math.Abs(axes[i][maxK]) {
				maxK = k
			}
		}
		if axes[i][maxK] < 0 {
			for k := 0; k < 3; k++ {
				axes[i][k] = -axes[i][k]
			}
		}
	}
	det := axes[0][0]*(axes[1][1]*axes[2][2]-axes[1][2]*axes[2][1]) -
		axes[0][1]*(axes[1][0]*axes[2][2]-axes[1][2]*axes[2][0]) +
		axes[0][2]*(axes[1][0]*axes[2][1]-axes[1][1]*axes[2][0])
	if det < 0 {
		for k := 0; k < 3; k++ {
			axes[2][k] = -axes[2][k]
		}
	}
	d.PrincipalAxes = axes

	// Half-extents: project every vertex onto each axis.
	minProj := [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxProj := [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, t := range mesh.Triangles {
		for _, v := range [][3]float64{t.V0, t.V1, t.V2} {
			for i := 0; i < 3; i++ {
				p := (v[0]-d.Centroid[0])*axes[i][0] +
					(v[1]-d.Centroid[1])*axes[i][1] +
					(v[2]-d.Centroid[2])*axes[i][2]
				minProj[i] = math.Min(minProj[i], p)
				maxProj[i] = math.Max(maxProj[i], p)
			}
		}
	}
	for i := 0; i < 3; i++ {
		d.Extents[i] = 0.5 * (maxProj[i] - minProj[i])
	}
	return d
}

func triangleArea(t Triangle) float64 {
	var e1, e2, n [3]float64
	for k := 0; k < 3; k++ {
		e1[k] = t.V1[k] - t.V0[k]
		e2[k] = t.V2[k] - t.V0[k]
	}
	n[0] = e1[1]*e2[2] - e1[2]*e2[1]
	n[1] = e1[2]*e2[0] - e1[0]*e2[2]
	n[2] = e1[0]*e2[1] - e1[1]*e2[0]
	return 0.5 * math.Sqrt(n[0]*n[0]+n[1]*n[1]+n[2]*n[2])
}

func triangleCentroid(t Triangle) [3]float64 {
	return [3]float64{
		(t.V0[0] + t.V1[0] + t.V2[0]) / 3,
		(t.V0[1] + t.V1[1] + t.V2[1]) / 3,
		(t.V0[2] + t.V1[2] + t.V2[2]) / 3,
	}
}

// symmetricEigenvectors returns the eigenvectors of a symmetric 3x3 matrix
// as unit rows ordered by decreasing eigenvalue, via cyclic Jacobi
// rotations (plenty for a covariance matrix).
func symmetricEigenvectors(m [3][3]float64) [3][3]float64 {
	a := m
	v := [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	for sweep := 0; sweep < 32; sweep++ {
		off := a[0][1]*a[0][1] + a[0][2]*a[0][2] + a[1][2]*a[1][2]
		if off < 1e-24 {
			break
		}
		for p := 0; p < 2; p++ {
			for q := p + 1; q < 3; q++ {
				if math.Abs(a[p][q]) < 1e-18 {
					continue
				}
				theta := (a[q][q] - a[p][p]) / (2 * a[p][q])
				t := math.Copysign(1, theta) / (math.Abs(theta) + math.Sqrt(theta*theta+1))
				c := 1 / math.Sqrt(t*t+1)
				s := t * c
				for k := 0; k < 3; k++ {
					akp, akq := a[k][p], a[k][q]
					a[k][p] = c*akp - s*akq
					a[k][q] = s*akp + c*akq
				}
				for k := 0; k < 3; k++ {
					apk, aqk := a[p][k], a[q][k]
					a[p][k] = c*apk - s*aqk
					a[q][k] = s*apk + c*aqk
				}
				for k := 0; k < 3; k++ {
					vkp, vkq := v[k][p], v[k][q]
					v[k][p] = c*vkp - s*vkq
					v[k][q] = s*vkp + c*vkq
				}
			}
		}
	}
	// Columns of v are eigenvectors; order rows by decreasing eigenvalue.
	type eig struct {
		val float64
		vec [3]float64
	}
	eigs := []eig{
		{a[0][0], [3]float64{v[0][0], v[1][0], v[2][0]}},
		{a[1][1], [3]float64{v[0][1], v[1][1], v[2][1]}},
		{a[2][2], [3]float64{v[0][2], v[1][2], v[2][2]}},
	}
	for i := 0; i < 2; i++ {
		for j := i + 1; j < 3; j++ {
			if eigs[j].val > eigs[i].val {
				eigs[i], eigs[j] = eigs[j], eigs[i]
			}
		}
	}
	var out [3][3]float64
	for i, e := range eigs {
		n := math.Sqrt(e.vec[0]*e.vec[0] + e.vec[1]*e.vec[1] + e.vec[2]*e.vec[2])
		if n == 0 {
			out[i][i] = 1
			continue
		}
		out[i] = [3]float64{e.vec[0] / n, e.vec[1] / n, e.vec[2] / n}
	}
	return out
}
