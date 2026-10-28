package meshmeasure

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestClosedMeasurement(t *testing.T) {
	stl := makeCubeSTL(2, [3]float32{1e6, -1e6, 1e6})
	stats, err := MeasureClosedSTL(stl)
	if err != nil || stats.Volume != 8 || stats.Center != [3]float32{1000001, -999999, 1000001} {
		t.Fatalf("translated cube: %+v, %v", stats, err)
	}
	for _, damage := range []string{"open", "reversed", "degenerate", "nonfinite"} {
		bad := append([]byte(nil), stl...)
		switch damage {
		case "open":
			binary.LittleEndian.PutUint32(bad[80:84], 11)
			bad = bad[:len(bad)-50]
		case "reversed":
			vertex := append([]byte(nil), bad[96:108]...)
			copy(bad[96:108], bad[108:120])
			copy(bad[108:120], vertex)
		case "degenerate":
			copy(bad[96:108], bad[108:120])
		case "nonfinite":
			binary.LittleEndian.PutUint32(bad[96:100], math.Float32bits(float32(math.NaN())))
		}
		if _, err := MeasureClosedSTL(bad); err == nil {
			t.Errorf("accepted %s surface", damage)
		}
	}
}

func TestClosedShellsSharingAnEdge(t *testing.T) {
	stl := makeCubeSTL(2, [3]float32{})
	other := makeCubeSTL(2, [3]float32{2, 2, 0})
	stl = append(stl, other[84:]...)
	binary.LittleEndian.PutUint32(stl[80:84], 24)
	stats, err := MeasureClosedSTL(stl)
	if err != nil || stats.Volume != 16 || stats.Center != [3]float32{2, 2, 1} {
		t.Fatalf("edge contact: %+v, %v", stats, err)
	}
}

func TestClosedMeasurementIgnoresCollapsedFacets(t *testing.T) {
	stl := makeCubeSTL(2, [3]float32{})
	stl = append(stl, make([]byte, 50)...)
	binary.LittleEndian.PutUint32(stl[80:84], 13)
	stats, err := MeasureClosedSTL(stl)
	if err != nil || stats.Volume != 8 {
		t.Fatalf("collapsed facet changed volume: %+v, %v", stats, err)
	}
}

func TestClosedMeasurementRetainsCollinearEdges(t *testing.T) {
	stl := makeCubeSTL(2, [3]float32{})
	var vertices [4][3]float32
	for j := 0; j < 3; j++ {
		for k := range vertices[j] {
			vertices[j][k] = math.Float32frombits(binary.LittleEndian.Uint32(stl[96+(j*3+k)*4:]))
		}
	}
	for k := range vertices[3] {
		vertices[3][k] = (vertices[0][k] + vertices[1][k]) / 2
	}
	stl = append(stl, make([]byte, 100)...)
	binary.LittleEndian.PutUint32(stl[80:84], 14)
	for i, triangle := range [][3]int{{0, 1, 3}, {1, 2, 3}, {2, 0, 3}} {
		facet := []int{0, 12, 13}[i]
		for j, vertex := range triangle {
			for k, coordinate := range vertices[vertex] {
				binary.LittleEndian.PutUint32(stl[96+facet*50+(j*3+k)*4:], math.Float32bits(coordinate))
			}
		}
	}
	stats, err := MeasureClosedSTL(stl)
	if err != nil || stats.Volume != 8 || stats.SurfaceArea != 24 {
		t.Fatalf("collinear facet changed a closed cube: %+v, %v", stats, err)
	}
}
