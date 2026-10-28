package spinodoid

import (
	"fmt"
	"math"
)

// normEps is ugf.normalize's eps: shorter vectors pass through unchanged.
const normEps = 1e-12

// normalize is hodge's ugf.normalize for one vector: v / max(|v|, eps) when
// |v| > eps, else v. The norm sums in numpy's order, (x² + y²) + z².
func normalize(v [3]float64) [3]float64 {
	n := math.Sqrt(float64(float64(v[0]*v[0])+float64(v[1]*v[1])) + float64(v[2]*v[2]))
	if !(n > normEps) {
		return v
	}
	return [3]float64{v[0] / n, v[1] / n, v[2] / n}
}

// dot3 sums in index order without fusing.
func dot3(a, b [3]float64) float64 {
	return float64(float64(a[0]*b[0])+float64(a[1]*b[1])) + float64(a[2]*b[2])
}

// deg2rad is numpy's radians: x * (NPY_PI / 180.0), the quotient rounded to a
// double first. Dividing a float64 variable (not the untyped constant, which
// Go would evaluate exactly) reproduces that rounding.
func deg2rad(x float64) float64 {
	pi := math.Pi
	return x * (pi / 180.0)
}

// cartesianPoles is spinodoid_waves' default, np.eye(3): the paper's Eq. (2).
var cartesianPoles = [][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}

// Waves is hodge's spinodoid_waves(theta_deg, n_waves, seed, poles) without the
// phases: the unit directions only, which is all the Gabor form reads (its
// kernel phases come from the spatial hash). The draw consumes numpy's stream
// exactly as hodge does: batches of 3-vectors of standard normals, normalized,
// kept when |n . p_k| > cos(theta_k) for any pole k, with hodge's adaptive
// batch size between rounds. Since the phases are drawn after the directions,
// leaving them out does not move the directions.
//
// poles nil means the Cartesian basis; otherwise each is normalized as hodge
// normalizes them.
func Waves(thetaDeg []float64, nWaves int, seed uint64, poles [][3]float64) ([][3]float64, error) {
	if poles == nil {
		poles = cartesianPoles
	} else {
		normed := make([][3]float64, len(poles))
		for i, p := range poles {
			normed[i] = normalize(p)
		}
		poles = normed
	}
	if len(thetaDeg) != len(poles) {
		return nil, fmt.Errorf("theta_deg has %d angles but %d poles were given", len(thetaDeg), len(poles))
	}
	anyOpen := false
	for _, t := range thetaDeg {
		if t < 0 || t > 90 {
			return nil, fmt.Errorf("theta_deg must lie in [0, 90]")
		}
		if t > 0 {
			anyOpen = true
		}
	}
	if !anyOpen {
		return nil, fmt.Errorf("at least one cone angle must be > 0")
	}
	if nWaves < 1 {
		return nil, fmt.Errorf("n_waves must be >= 1, got %d", nWaves)
	}
	cosTheta := make([]float64, len(thetaDeg))
	for i, t := range thetaDeg {
		cosTheta[i] = math.Cos(deg2rad(t))
	}

	g := newGenerator(seed)
	kept := make([][3]float64, 0, nWaves)
	nHave, nDrawn := 0, 0
	batch := max(4*nWaves, 1024)
	for nHave < nWaves {
		accepted := 0
		for range batch {
			var v [3]float64
			for a := range v {
				v[a] = g.standardNormal()
			}
			n := normalizeRow(v)
			for k, p := range poles {
				if math.Abs(dot3(n, p)) > cosTheta[k] {
					kept = append(kept, n)
					accepted++
					break
				}
			}
		}
		nDrawn += batch
		if accepted > 0 {
			nHave += accepted
		} else if nDrawn > 1_000_000 {
			return nil, fmt.Errorf("cone ODF %v deg is too narrow to sample; widen theta_deg", thetaDeg)
		}
		rate := math.Max(float64(nHave)/float64(nDrawn), 1e-6)
		next := float64(float64(float64(nWaves-nHave)/rate)*1.3) + 64
		batch = int(math.Min(math.Max(next, 1024), 1<<22))
	}
	return kept[:nWaves], nil
}

// normalizeRow is the draw's own v / np.linalg.norm(v): no eps guard, unlike
// ugf.normalize.
func normalizeRow(v [3]float64) [3]float64 {
	n := math.Sqrt(float64(float64(v[0]*v[0])+float64(v[1]*v[1])) + float64(v[2]*v[2]))
	return [3]float64{v[0] / n, v[1] / n, v[2] / n}
}
