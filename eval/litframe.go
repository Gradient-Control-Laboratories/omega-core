//go:build cgo

// Framing a lit render on the geometry when no extent is given.

package eval

import (
	"log"
	"math"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
)

// frameScanRes and frameScanHalfWidth control the camera-framing scan: two
// passes of a coarse grid over the union field, a cell counting as occupied
// where its center's value is negative.
const (
	frameScanRes       = 24
	frameScanHalfWidth = float32(1024)
)

// frameCanvas fits a view to the canvas's negative region: the occupied
// box's center, with the view half-height (scale) at 1.25x the
// box's bounding radius — the radius covers every projected extent under
// any yaw/pitch, and the view is wider than tall. ok is false when nothing
// negative lies within the scan region or the eval fails.
func frameCanvas(canvasIndex fidget.CanvasIndex) (center [3]float32, scale float32, ok bool) {
	box, found := occupiedBox(canvasIndex)
	if !found {
		return center, 0, false
	}
	var radius2 float32
	for k := 0; k < 3; k++ {
		center[k] = 0.5 * (box[0][k] + box[1][k])
		half := 0.5 * (box[1][k] - box[0][k])
		radius2 += half * half
	}
	scale = 1.25 * float32(math.Sqrt(float64(radius2)))
	if !(scale > 0) {
		return center, 0, false
	}
	return center, scale, true
}

// occupiedBox is the box around the canvas's negative region, as the frame
// scan finds it: a coarse grid over the scan region, then up to four finer
// passes over the box found so far, each widened by one of its cells. found
// is false when nothing negative lies within the scan region or the eval
// fails.
func occupiedBox(canvasIndex fidget.CanvasIndex) (box [2][3]float32, found bool) {
	coarse := [2][3]float32{
		{-frameScanHalfWidth, -frameScanHalfWidth, -frameScanHalfWidth},
		{frameScanHalfWidth, frameScanHalfWidth, frameScanHalfWidth},
	}
	box, found = scanNegativeBox(canvasIndex, coarse)
	if !found {
		return box, false
	}
	diag := func(b [2][3]float32) float32 {
		var d2 float32
		for k := 0; k < 3; k++ {
			d := b[1][k] - b[0][k]
			d2 += d * d
		}
		return float32(math.Sqrt(float64(d2)))
	}
	for pass := 0; pass < 4; pass++ {
		inflated := box
		for k := 0; k < 3; k++ {
			cell := (box[1][k] - box[0][k]) / frameScanRes
			inflated[0][k] -= cell
			inflated[1][k] += cell
		}
		refined, foundRefined := scanNegativeBox(canvasIndex, inflated)
		if !foundRefined {
			break
		}
		shrunk := diag(refined) < 0.75*diag(box)
		box = refined
		if !shrunk {
			break
		}
	}
	return box, true
}

func scanNegativeBox(canvasIndex fidget.CanvasIndex, region [2][3]float32) (box [2][3]float32, found bool) {
	var cell [3]float32
	for k := 0; k < 3; k++ {
		cell[k] = (region[1][k] - region[0][k]) / frameScanRes
	}
	halfDiag := 0.5 * float32(math.Sqrt(float64(cell[0]*cell[0]+cell[1]*cell[1]+cell[2]*cell[2])))
	n := frameScanRes * frameScanRes * frameScanRes
	xs := make([]float32, 0, n)
	ys := make([]float32, 0, n)
	zs := make([]float32, 0, n)
	for iz := 0; iz < frameScanRes; iz++ {
		z := region[0][2] + (float32(iz)+0.5)*cell[2]
		for iy := 0; iy < frameScanRes; iy++ {
			y := region[0][1] + (float32(iy)+0.5)*cell[1]
			for ix := 0; ix < frameScanRes; ix++ {
				xs = append(xs, region[0][0]+(float32(ix)+0.5)*cell[0])
				ys = append(ys, y)
				zs = append(zs, z)
			}
		}
	}
	samples, err := fidget.EvalGrad(canvasIndex, xs, ys, zs)
	if err != nil {
		log.Printf("frame scan eval: %v", err)
		return box, false
	}

	inf := float32(math.Inf(1))
	newBox := func() [2][3]float32 {
		return [2][3]float32{{inf, inf, inf}, {-inf, -inf, -inf}}
	}
	grow := func(b *[2][3]float32, c [3]float32) {
		for k := 0; k < 3; k++ {
			if c[k]-0.5*cell[k] < b[0][k] {
				b[0][k] = c[k] - 0.5*cell[k]
			}
			if c[k]+0.5*cell[k] > b[1][k] {
				b[1][k] = c[k] + 0.5*cell[k]
			}
		}
	}
	negative, nearBand := newBox(), newBox()
	negFound, nearCount := false, 0
	for i, s := range samples {
		c := [3]float32{xs[i], ys[i], zs[i]}
		if s.Value < 0 {
			negFound = true
			grow(&negative, c)
		} else if s.Value < halfDiag {
			nearCount++
			grow(&nearBand, c)
		}
	}
	switch {
	case negFound:
		// The near band too: a part too thin for any cell centre to fall inside
		// it is only ever near, and the refinement widens by a cell at most, so
		// one dropped here never comes back.
		for k := 0; k < 3; k++ {
			negative[0][k] = float32(math.Min(float64(negative[0][k]), float64(nearBand[0][k])))
			negative[1][k] = float32(math.Max(float64(negative[1][k]), float64(nearBand[1][k])))
		}
		return negative, true
	case nearCount > 0 && nearCount < n*4/5:
		return nearBand, true
	}
	return box, false
}
