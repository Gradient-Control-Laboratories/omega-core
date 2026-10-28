package meshmeasure

import (
	"encoding/binary"
	"fmt"
	"math"
)

// MeasureClosedSTL rejects non-finite, open and inconsistently wound surfaces.
func MeasureClosedSTL(stl []byte) (STLStats, error) {
	stats, err := MeasureSTL(stl)
	if err != nil || stats.Empty() {
		return stats, err
	}
	type edge struct{ a, b [3]float32 }
	edges := make(map[edge]int, stats.TriangleCount*3/2)
	less := func(a, b [3]float32) bool {
		for i := range a {
			if a[i] != b[i] {
				return a[i] < b[i]
			}
		}
		return false
	}
	for i := 0; i < stats.TriangleCount; i++ {
		var v [3][3]float32
		for j := range v {
			for k := range v[j] {
				v[j][k] = math.Float32frombits(binary.LittleEndian.Uint32(stl[84+i*50+12+(j*3+k)*4:]))
			}
		}
		for j := range v {
			e, direction := edge{v[j], v[(j+1)%3]}, 1
			if e.a == e.b {
				continue
			}
			if less(e.b, e.a) {
				e.a, e.b, direction = e.b, e.a, -1
			}
			edges[e] += direction
		}
	}
	// Closed shells can share an edge if their directed boundaries still cancel.
	for _, direction := range edges {
		if direction != 0 {
			return STLStats{}, fmt.Errorf("mesh is not a closed consistently oriented surface")
		}
	}
	if stats.Volume <= 0 {
		return STLStats{}, fmt.Errorf("closed mesh has no positive volume")
	}
	return stats, nil
}
