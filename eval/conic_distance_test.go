//go:build cgo

package eval

import (
	"fmt"
	"math"
	"testing"
)

// The conic and quadric distances, checked against brute force: the nearest point of each curve or surface found
// by dense sampling of its parameters and golden-section refinement, independently of the Lagrange-multiplier
// method the library uses. Near the surface the distance must be right to single precision; far from it, to a
// small part of itself.

// nearestOnCurve is the least distance from (x, y) to the curve c(u), u in [lo, hi].
func nearestOnCurve(c func(float64) (float64, float64), lo, hi, x, y float64) float64 {
	dist2 := func(u float64) float64 { cx, cy := c(u); return (cx-x)*(cx-x) + (cy-y)*(cy-y) }
	const n = 4000
	best, bu := math.Inf(1), lo
	for i := 0; i <= n; i++ {
		u := lo + (hi-lo)*float64(i)/n
		if d := dist2(u); d < best {
			best, bu = d, u
		}
	}
	a, b := bu-(hi-lo)/n, bu+(hi-lo)/n
	for i := 0; i < 80; i++ {
		m1, m2 := a+(b-a)*0.382, a+(b-a)*0.618
		if dist2(m1) < dist2(m2) {
			b = m2
		} else {
			a = m1
		}
	}
	return math.Sqrt(dist2((a + b) / 2))
}

// checkConic compares the field "f" of expr at a grid of points in the plane z = 0 (or the given points) with
// the signed brute-force distance.
func checkConic(t *testing.T, expr string, points [][3]float64, want func(q [3]float64) float64) {
	t.Helper()
	got := evalChamferField(t, "@scene(40)\nf = "+expr+";\nreturn f;\n", "f", points)
	worstNear, worstFar := 0.0, 0.0
	for i, q := range points {
		w := want(q)
		e := math.Abs(got[i] - w)
		if math.Abs(w) < 2 {
			worstNear = math.Max(worstNear, e)
			if !(e <= 2e-5) { // a NaN fails too
				t.Errorf("at %v, near the surface: %.7f, want %.7f", q, got[i], w)
			}
		} else {
			worstFar = math.Max(worstFar, e/math.Abs(w))
			if !(e <= 1e-3*math.Abs(w)) {
				t.Errorf("at %v: %.6f, want %.6f", q, got[i], w)
			}
		}
	}
	t.Logf("worst error near the surface %.2e, relative far from it %.2e", worstNear, worstFar)
}

// planeGrid is a grid over the plane z = 0, plus points exactly on the x and y axes, where a conic's nearest points
// may not be unique (a pair across an ellipse's long axis inside its evolute) and the method needs its special case.
func planeGrid() [][3]float64 {
	var pts [][3]float64
	for i := -10; i <= 10; i++ {
		pts = append(pts, [3]float64{float64(i), 0, 0}, [3]float64{0, float64(i) + 0.5, 0})
	}
	for i := 0; i <= 30; i++ {
		for j := 0; j <= 30; j++ {
			pts = append(pts, [3]float64{-11 + 22*float64(i)/30 + 0.0137, -11 + 22*float64(j)/30 + 0.0291, 0})
		}
	}
	return pts
}

func TestConicDistance_Ellipse(t *testing.T) {
	for _, ab := range [][2]float64{{5, 3}, {10, 1}} {
		a, b := ab[0], ab[1]
		t.Run(fmt.Sprintf("%v by %v", a, b), func(t *testing.T) {
			curve := func(u float64) (float64, float64) { return a * math.Cos(u), b * math.Sin(u) }
			checkConic(t, fmt.Sprintf("ellipseXY([0, 0], [%v, %v])", a, b), planeGrid(), func(q [3]float64) float64 {
				d := nearestOnCurve(curve, 0, 2*math.Pi, q[0], q[1])
				if q[0]*q[0]/(a*a)+q[1]*q[1]/(b*b) < 1 {
					return -d
				}
				return d
			})
		})
	}
}

func TestConicDistance_Parabola(t *testing.T) {
	curve := func(u float64) (float64, float64) { return u, u * u / 8 }
	checkConic(t, "parabolaXY([0, 0], 2)", planeGrid(), func(q [3]float64) float64 {
		d := nearestOnCurve(curve, -60, 60, q[0], q[1])
		if q[1] > q[0]*q[0]/8 {
			return -d
		}
		return d
	})
}

func TestConicDistance_Hyperbola(t *testing.T) {
	right := func(u float64) (float64, float64) { return 2 * math.Cosh(u), 1.5 * math.Sinh(u) }
	left := func(u float64) (float64, float64) { return -2 * math.Cosh(u), 1.5 * math.Sinh(u) }
	checkConic(t, "hyperbolaXY([0, 0], [2, 1.5])", planeGrid(), func(q [3]float64) float64 {
		d := math.Min(nearestOnCurve(right, -4, 4, q[0], q[1]), nearestOnCurve(left, -4, 4, q[0], q[1]))
		if q[0]*q[0]/4-q[1]*q[1]/2.25 > 1 {
			return -d
		}
		return d
	})
}

// The ellipsoid with radii 3, 4 and 5: the brute force minimizes over its two angles, first on a grid, then by
// alternating golden-section refinement.
func TestConicDistance_Ellipsoid(t *testing.T) {
	r := [3]float64{3, 4, 5}
	surf := func(th, ph float64) [3]float64 {
		return [3]float64{r[0] * math.Sin(th) * math.Cos(ph), r[1] * math.Sin(th) * math.Sin(ph), r[2] * math.Cos(th)}
	}
	var pts [][3]float64
	for i := 0; i <= 8; i++ {
		for j := 0; j <= 8; j++ {
			for k := 0; k <= 4; k++ {
				pts = append(pts, [3]float64{-8 + 2*float64(i) + 0.013, -8 + 2*float64(j) + 0.029, -8 + 4*float64(k) + 0.017})
			}
		}
	}
	checkConic(t, "ellipsoid([0, 0, 0], [3, 4, 5])", pts, func(q [3]float64) float64 {
		d2 := func(th, ph float64) float64 { s := surf(th, ph); return sq(s[0]-q[0]) + sq(s[1]-q[1]) + sq(s[2]-q[2]) }
		best, bt, bp := math.Inf(1), 0.0, 0.0
		for i := 0; i <= 200; i++ {
			for j := 0; j <= 400; j++ {
				th, ph := math.Pi*float64(i)/200, 2*math.Pi*float64(j)/400
				if v := d2(th, ph); v < best {
					best, bt, bp = v, th, ph
				}
			}
		}
		for it := 0; it < 30; it++ {
			bt = golden(func(th float64) float64 { return d2(th, bp) }, bt-0.02, bt+0.02)
			bp = golden(func(ph float64) float64 { return d2(bt, ph) }, bp-0.02, bp+0.02)
		}
		d := math.Sqrt(d2(bt, bp))
		if sq(q[0]/r[0])+sq(q[1]/r[1])+sq(q[2]/r[2]) < 1 {
			return -d
		}
		return d
	})
}

// The paraboloid z = (x² + y²)/8, a parabola turned about the z axis: its distance is the parabola's in the plane
// through the axis and the point. The grid holds points on the axis, where above the focus the nearest points are a
// ring.
func TestConicDistance_Paraboloid(t *testing.T) {
	curve := func(u float64) (float64, float64) { return u, u * u / 8 }
	var pts [][3]float64
	for i := 0; i <= 10; i++ {
		for k := 0; k <= 12; k++ {
			a := 0.7 * float64(i)
			pts = append(pts, [3]float64{(-7 + 1.4*float64(i)) * math.Cos(a), (-7 + 1.4*float64(i)) * math.Sin(a), -4 + 1.2*float64(k) + 0.019})
		}
	}
	checkConic(t, "paraboloid([0, 0, 0], 2)", pts, func(q [3]float64) float64 {
		rho := math.Hypot(q[0], q[1])
		d := nearestOnCurve(curve, -60, 60, rho, q[2])
		if q[2] > rho*rho/8 {
			return -d
		}
		return d
	})
}

func sq(x float64) float64 { return x * x }

func golden(f func(float64) float64, a, b float64) float64 {
	for i := 0; i < 60; i++ {
		m1, m2 := a+(b-a)*0.382, a+(b-a)*0.618
		if f(m1) < f(m2) {
			b = m2
		} else {
			a = m1
		}
	}
	return (a + b) / 2
}
