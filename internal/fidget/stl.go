package fidget

import (
	"fmt"
	"log"
	gotime "time"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/meshrepair"
)

// AlgorithmMDCRepair is Manifold Dual Contouring followed by a repair of the folds it leaves (meshrepair.Folds),
// guided by the field just meshed.
const AlgorithmMDCRepair = "mdc-repair"

// PadBbox grows bbox by cells of an octree of depth subdivisionLevel over it, on every axis: room for a surface lying
// on the box to fall inside the mesher's cells. The room is the mesh's own cells, so it halves with each level. An
// axis of zero size takes its cell from the largest axis.
func PadBbox(bbox [6]float32, subdivisionLevel, cells int) [6]float32 {
	if cells <= 0 {
		return bbox
	}
	n := float32(int(1) << uint(subdivisionLevel))
	largest := float32(0)
	for k := 0; k < 3; k++ {
		largest = max(largest, bbox[k+3]-bbox[k])
	}
	for k := 0; k < 3; k++ {
		size := bbox[k+3] - bbox[k]
		if size <= 0 {
			size = largest
		}
		pad := float32(cells) * size / n
		bbox[k] -= pad
		bbox[k+3] += pad
	}
	return bbox
}

// GenerateSTL meshes the union of treeIndices into binary STL bytes using
// either "marching-cubes" or Manifold Dual Contouring ("mdc"), or MDC with
// its folds repaired against the union's field (AlgorithmMDCRepair).
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
				// With a repair to come, Done waits for it.
				if progressFn != nil && algorithm != AlgorithmMDCRepair {
					progressFn(completed, total, "Done")
				}
			}
		}
		stlBytes, err = MeshStl(canvasIndex, subdivisionLevel, bbox, dualFn)
		if err != nil {
			return nil, fmt.Errorf("manifold dual-contouring failed to generate stl: %v", err)
		}
		if algorithm == AlgorithmMDCRepair {
			if progressFn != nil {
				progressFn(0, 0, "Repairing folds...")
			}
			start := gotime.Now()
			field := func(xs, ys, zs []float32) ([]meshrepair.Sample, error) {
				grads, err := EvalGrad(canvasIndex, xs, ys, zs)
				if err != nil {
					return nil, err
				}
				out := make([]meshrepair.Sample, len(grads))
				for i, g := range grads {
					out[i] = meshrepair.Sample{Value: g.Value, Dx: g.Dx, Dy: g.Dy, Dz: g.Dz}
				}
				return out, nil
			}
			repaired, stats, err := meshrepair.Folds(stlBytes, field)
			if err != nil {
				return nil, err
			}
			log.Printf("mdc fold repair: backward triangles %d -> %d in %d passes (%d vertices moved), %s",
				stats.BackwardBefore, stats.BackwardAfter, stats.Passes, stats.Moved,
				gotime.Since(start).Round(gotime.Millisecond))
			stlBytes = repaired
			if progressFn != nil {
				progressFn(4, 0, "Done")
			}
		}
	}
	return stlBytes, nil
}
