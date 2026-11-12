//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
)

const (
	defaultSurfaceMaxIters  = 8
	maxSurfaceMaxIters      = 32
	defaultSurfaceTolerance = 1e-4
)

// SurfacePoint projects each seed point onto the zero level set of a single
// scalar expression by a Newton walk along the gradient:
// p <- p - f*grad/|grad|^2. For an exact SDF |grad| = 1, so the step is
// p - value*gradient; the 1/|grad|^2 keeps it correct for scaled or
// min/max-composed fields. maxIters <= 0 and tolerance <= 0 select defaults.
func (e *Evaluator) SurfacePoint(ctx context.Context, manifest map[string]string, code, defs, expr string, points [][3]float64, maxIters int, tolerance float64) ([]SurfaceSample, error) {
	if err := startQuery(ctx); err != nil {
		return nil, err
	}
	defer endQuery()
	if maxIters <= 0 {
		maxIters = defaultSurfaceMaxIters
	}
	if maxIters > maxSurfaceMaxIters {
		maxIters = maxSurfaceMaxIters
	}
	if tolerance <= 0 {
		tolerance = defaultSurfaceTolerance
	}

	prog, err := e.prepareQuery(ctx, manifest, code, defs, []string{expr})
	if err != nil {
		return nil, err
	}

	// The iterate for each point, walked toward the surface in place.
	cur := make([][3]float64, len(points))
	copy(cur, points)
	// done stops a point from stepping further (converged, or no usable
	// gradient); the reported Converged is recomputed from the final residual.
	done := make([]bool, len(points))

	// samples holds the latest value+gradient at each current iterate; the
	// final pass refreshes it for reporting without stepping again.
	var samples []GradSample
	for iter := 0; iter <= maxIters; iter++ {
		xs := make([]float32, len(cur))
		ys := make([]float32, len(cur))
		zs := make([]float32, len(cur))
		for i, pt := range cur {
			xs[i], ys[i], zs[i] = float32(pt[0]), float32(pt[1]), float32(pt[2])
		}
		grads, err := EvalGrad(prog, xs, ys, zs)
		if err != nil {
			return nil, err
		}
		if len(grads) != 1 {
			return nil, fmt.Errorf("surfacePoint expr must be scalar-valued, got %d components", len(grads))
		}
		samples = grads[0]

		if iter == maxIters {
			break
		}
		for i := range cur {
			if done[i] {
				continue
			}
			s := samples[i]
			v := float64(s.Value)
			gx, gy, gz := float64(s.Dx), float64(s.Dy), float64(s.Dz)
			if math.Abs(v) <= tolerance {
				done[i] = true
				continue
			}
			g2 := gx*gx + gy*gy + gz*gz
			if g2 == 0 || math.IsNaN(v) || math.IsInf(v, 0) {
				// Flat region or non-finite sample — no usable step; leave the
				// point where it is (it reports unconverged with its residual).
				done[i] = true
				continue
			}
			step := v / g2
			cur[i][0] -= step * gx
			cur[i][1] -= step * gy
			cur[i][2] -= step * gz
		}
	}

	out := make([]SurfaceSample, len(points))
	for i := range points {
		s := samples[i]
		v := float64(s.Value)
		gx, gy, gz := float64(s.Dx), float64(s.Dy), float64(s.Dz)
		normal := [3]float64{gx, gy, gz}
		if n := math.Sqrt(gx*gx + gy*gy + gz*gz); n > 0 {
			normal = [3]float64{gx / n, gy / n, gz / n}
		}
		out[i] = SurfaceSample{
			Point:     cur[i],
			Normal:    normal,
			Value:     v,
			Converged: math.Abs(v) <= tolerance,
		}
	}
	return out, nil
}
