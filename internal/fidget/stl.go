package fidget

import (
	"fmt"
	"log"
	gotime "time"
)

// GenerateSTL meshes the union of treeIndices into binary STL bytes using
// either "marching-cubes" or Manifold Dual Contouring ("mdc").
// It assumes the caller holds fidget.Lock().
func GenerateSTL(
	treeIndices []TreeIndex,
	subdivisionLevel int,
	bbox [6]float32,
	algorithm string,
	progressFn func(completed, total uint32, status string),
) ([]byte, error) {
	if len(treeIndices) == 0 {
		return nil, fmt.Errorf("no implicit shapes to mesh")
	}

	treeIndex := treeIndices[0]
	for i := 1; i < len(treeIndices); i++ {
		treeIndex = Min(treeIndex, treeIndices[i])
	}

	canvasIndex := BuildCanvas(treeIndex)

	var stlBytes []byte
	var err error
	if algorithm == "marching-cubes" {
		marchingFn := func(completed, total uint32) {
			if progressFn != nil {
				progressFn(completed, total, "")
			}
		}
		log.Printf("marching cubes: depth=%d, bbox=%v", subdivisionLevel, bbox)
		stlBytes, err = MeshStlMarchingCubes(canvasIndex, subdivisionLevel, bbox, marchingFn)
		if err != nil {
			return nil, fmt.Errorf("marching cubes failed to generate stl: %v", err)
		}
	} else {
		// MDC sends phase signals: (phase, 0) where total=0 means phase indicator.
		phaseNames := map[uint32]string{
			1: "Building octree",
			2: "Walking dual",
			3: "Writing STL",
		}
		phaseStart := gotime.Now()
		dualFn := func(completed, total uint32) {
			if total != 0 {
				if progressFn != nil {
					progressFn(completed, total, "")
				}
				return
			}
			elapsed := gotime.Since(phaseStart)
			if name, ok := phaseNames[completed]; ok {
				if completed > 1 {
					log.Printf("mdc phase %d done in %s", completed-1, elapsed.Round(gotime.Millisecond))
				}
				if progressFn != nil {
					progressFn(completed, total, fmt.Sprintf("%s...", name))
				}
				phaseStart = gotime.Now()
			} else if completed == 4 {
				log.Printf("mdc phase 3 done in %s", elapsed.Round(gotime.Millisecond))
				if progressFn != nil {
					progressFn(completed, total, "Done")
				}
			}
		}
		stlBytes, err = MeshStl(canvasIndex, subdivisionLevel, bbox, dualFn)
		if err != nil {
			return nil, fmt.Errorf("manifold dual-contouring failed to generate stl: %v", err)
		}
	}
	return stlBytes, nil
}

