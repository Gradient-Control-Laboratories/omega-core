//go:build cgo

package eval

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// Corner blends that are edges of the loop: an arc fillet or a line chamfer
// at a vertex between two straight edges, so the profile stays exact. The
// reference builds the blended loop itself and is held to it as any loop is.

type cornerKind int

const (
	sharpCorner cornerKind = iota
	filletCorner
	chamferCorner
)

type refCorner struct {
	kind cornerKind
	size float64
}

func (c refCorner) gsl() string {
	switch c.kind {
	case filletCorner:
		return fmt.Sprintf("cornerFillet(%v)", c.size)
	case chamferCorner:
		return fmt.Sprintf("cornerChamfer(%v)", c.size)
	}
	return "cornerSharp()"
}

// blendedLoop is the loop through points with each corner blended: a fillet
// of radius r at a corner turning by phi trims both edges by r * tan(|phi| / 2)
// and joins the trim points with an arc of curvature sign(phi) / r; a chamfer
// trims both by its setback and joins them with a line. No trim passes half
// of either edge at the corner: a larger blend is cut down to that.
func blendedLoop(points [][2]float64, corners []refCorner) refLoop {
	n := len(points)
	var out refLoop
	for i, v := range points {
		prev, next := points[(i+n-1)%n], points[(i+1)%n]
		inX, inY := v[0]-prev[0], v[1]-prev[1]
		outX, outY := next[0]-v[0], next[1]-v[1]
		inLen, outLen := math.Hypot(inX, inY), math.Hypot(outX, outY)
		inX, inY, outX, outY = inX/inLen, inY/inLen, outX/outLen, outY/outLen
		turn := math.Atan2(inX*outY-inY*outX, inX*outX+inY*outY)
		limit := math.Min(inLen, outLen) / 2
		c := corners[i]
		trim := 0.0
		switch c.kind {
		case filletCorner:
			trim = math.Min(c.size*math.Tan(math.Abs(turn)/2), limit)
		case chamferCorner:
			trim = math.Min(c.size, limit)
		}
		if trim == 0 {
			out.points = append(out.points, v)
			out.curvatures = append(out.curvatures, 0)
			continue
		}
		a := [2]float64{v[0] - trim*inX, v[1] - trim*inY}
		b := [2]float64{v[0] + trim*outX, v[1] + trim*outY}
		k := 0.0
		if c.kind == filletCorner {
			radius := trim / math.Tan(math.Abs(turn)/2)
			k = math.Copysign(1/radius, turn)
		}
		out.points = append(out.points, a, b)
		out.curvatures = append(out.curvatures, k, 0)
	}
	// Two trims that meet leave a straight edge of no length: drop its start.
	var merged refLoop
	for i, q := range out.points {
		next := out.points[(i+1)%len(out.points)]
		if out.curvatures[i] == 0 && math.Hypot(next[0]-q[0], next[1]-q[1]) < 1e-12 {
			continue
		}
		merged.points = append(merged.points, q)
		merged.curvatures = append(merged.curvatures, out.curvatures[i])
	}
	return merged
}

// cornersExpr is profileXY with straight edges and the given corners.
func cornersExpr(points [][2]float64, corners []refCorner) string {
	pts := make([]string, len(points))
	zeros := make([]string, len(points))
	cs := make([]string, len(points))
	for i, q := range points {
		pts[i] = fmt.Sprintf("[%v, %v]", q[0], q[1])
		zeros[i] = "0"
		cs[i] = corners[i].gsl()
	}
	return fmt.Sprintf("profileXY([%s], [%s], [%s])", strings.Join(pts, ", "), strings.Join(zeros, ", "), strings.Join(cs, ", "))
}

func pointsGSL(points [][2]float64) string {
	pts := make([]string, len(points))
	for i, q := range points {
		pts[i] = fmt.Sprintf("[%v, %v]", q[0], q[1])
	}
	return "[" + strings.Join(pts, ", ") + "]"
}

// uShape has two reflex corners at the bottom of its slot.
var uShape = [][2]float64{{0, 0}, {30, 0}, {30, 25}, {22, 25}, {22, 8}, {8, 8}, {8, 25}, {0, 25}}

func cornerCases(t *testing.T) []profileCase {
	mixed := []refCorner{
		{filletCorner, 3}, {chamferCorner, 2}, {filletCorner, 2}, {sharpCorner, 0},
		{filletCorner, 3}, {chamferCorner, 2}, {sharpCorner, 0}, {filletCorner, 1.5},
	}
	starPoints := star().points
	everyFillet := make([]refCorner, len(starPoints))
	for i := range everyFillet {
		everyFillet[i] = refCorner{filletCorner, 1}
	}
	square := [][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	clamped := []refCorner{{filletCorner, 8}, {filletCorner, 8}, {filletCorner, 8}, {filletCorner, 8}}
	chamfered := []refCorner{{chamferCorner, 7}, {sharpCorner, 0}, {chamferCorner, 1}, {sharpCorner, 0}}
	return []profileCase{
		{name: "U, mixed corners", expr: cornersExpr(uShape, mixed), ref: newRefProfile(t, blendedLoop(uShape, mixed))},
		// The curvatures are written out: a list made by map and handed in as
		// curvatures fails to assemble once the profile is named (its box pass),
		// with or without corners.
		{
			name: "star, everyCorner fillet",
			expr: fmt.Sprintf("profileXY(%s, [0, 0, 0, 0, 0, 0, 0, 0, 0, 0], everyCorner(%s, cornerFillet(1)))",
				pointsGSL(starPoints), pointsGSL(starPoints)),
			ref: newRefProfile(t, blendedLoop(starPoints, everyFillet)),
		},
		// Every fillet clamped to half a side: a circle of radius 5.
		{name: "square, fillets clamped", expr: cornersExpr(square, clamped), ref: newRefProfile(t, blendedLoop(square, clamped))},
		{name: "square, a chamfer clamped", expr: cornersExpr(square, chamfered), ref: newRefProfile(t, blendedLoop(square, chamfered))},
	}
}

// Fillets and chamfers between straight edges keep the profile exact.
func TestProfileXY_CornersMatchReference(t *testing.T) {
	for _, c := range cornerCases(t) {
		t.Run(c.name, func(t *testing.T) {
			points := flat(gridAround(c.ref.box(), 5, 52))
			got := evalProfileField(t, "@scene(120)\nsolid = "+c.expr+";\nreturn solid;\n", "solid", points)
			compareToReference(t, c.name, points, got, func(q [3]float64) float64 {
				return c.ref.distance([2]float64{q[0], q[1]})
			})
		})
	}
}

// The clamped square is a circle.
func TestProfileXY_ClampedFilletsMakeACircle(t *testing.T) {
	square := [][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	clamped := []refCorner{{filletCorner, 8}, {filletCorner, 8}, {filletCorner, 8}, {filletCorner, 8}}
	points := flat(gridAround([4]float64{0, 0, 10, 10}, 4, 30))
	got := evalProfileField(t, "@scene(40)\nsolid = "+cornersExpr(square, clamped)+";\nreturn solid;\n", "solid", points)
	compareToReference(t, "clamped square", points, got, func(q [3]float64) float64 {
		return math.Hypot(q[0]-5, q[1]-5) - 5
	})
}

// checkSurfaceCellsOfHalf is checkSurfaceCells with cells of half-width half:
// sound, and no wider than maxCellSpread of their own diagonals.
func checkSurfaceCellsOfHalf(t *testing.T, label, doc string, centres [][3]float64, half float64, want func([3]float64) float64) {
	t.Helper()
	boxes := make([][6]float32, len(centres))
	for i, q := range centres {
		boxes[i] = [6]float32{
			float32(q[0] - half), float32(q[1] - half), float32(q[2] - half),
			float32(q[0] + half), float32(q[1] + half), float32(q[2] + half),
		}
	}
	bounds := intervalOver(t, doc, "solid", boxes)
	diagonal := 2 * half * math.Sqrt(3)
	widest, failures := 0.0, 0
	for i, q := range centres {
		lo, hi := float64(bounds[i].Lower), float64(bounds[i].Upper)
		spread := (hi - lo) / diagonal
		widest = math.Max(widest, spread)
		if value := want(q); value < lo-profileTolerance || value > hi+profileTolerance || spread > maxCellSpread {
			if failures++; failures <= 3 {
				t.Errorf("%s: cell about %v bounds [%v, %v], reference at its centre %v, spread %.2f diagonals",
					label, q, lo, hi, value, spread)
			}
		}
	}
	t.Logf("%s: %d surface cells of half-width %v, widest %.2f diagonals", label, len(centres), half, widest)
}

// Blended corners stay provable: cells on the boundary are tight. An
// interval over a curve is as loose as the cell is wide against its radius,
// so the cells are no wider than a fifth of the smallest fillet's radius.
func TestProfileXY_CornersSurfaceCellsAreTight(t *testing.T) {
	smallestRadius := []float64{1.5, 1}
	for i, c := range cornerCases(t)[:2] {
		t.Run(c.name, func(t *testing.T) {
			doc := "@scene(120)\nsolid = " + c.expr + ";\nreturn solid;\n"
			half := math.Min(cellHalf, smallestRadius[i]/10)
			checkSurfaceCellsOfHalf(t, c.name, doc, flat(c.ref.boundarySamples(0.75)), half, func(q [3]float64) float64 {
				return c.ref.distance([2]float64{q[0], q[1]})
			})
		})
	}
}

// Until the phase's second part, a corner next to an arc edge stays sharp
// whatever it asks for: the arc L with a fillet asked at every vertex is the
// arc L.
func TestProfileXY_CornerNextToAnArcStaysSharp(t *testing.T) {
	l := arcL()
	asked := make([]string, len(l.points))
	ks := make([]string, len(l.curvatures))
	for i := range asked {
		asked[i] = "cornerSharp()"
		if l.curvatures[i] != 0 || l.curvatures[(i+len(l.curvatures)-1)%len(l.curvatures)] != 0 {
			asked[i] = "cornerFillet(2)"
		}
		if k := l.curvatures[i]; k < 0 {
			ks[i] = fmt.Sprintf("neg(%v)", -k)
		} else {
			ks[i] = fmt.Sprintf("%v", k)
		}
	}
	expr := fmt.Sprintf("profileXY(%s, [%s], [%s])", pointsGSL(l.points), strings.Join(ks, ", "), strings.Join(asked, ", "))
	ref := newRefProfile(t, l)
	points := flat(gridAround(ref.box(), 5, 40))
	got := evalProfileField(t, "@scene(120)\nsolid = "+expr+";\nreturn solid;\n", "solid", points)
	compareToReference(t, "arc L, fillets asked next to its arcs", points, got, func(q [3]float64) float64 {
		return ref.distance([2]float64{q[0], q[1]})
	})
}
