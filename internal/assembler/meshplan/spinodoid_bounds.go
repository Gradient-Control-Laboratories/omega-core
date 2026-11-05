package meshplan

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/scene"
)

// ComputeSpinodoidBounds computes the 6-element [minX, minY, minZ, maxX, maxY, maxZ]
// driver-baking domain for a program. Imported mesh bounding boxes and their
// transformed 8-corner placements (padded by 5%) take precedence; otherwise an
// explicit positive Scene.Size supplies [-extent, -extent, -extent, extent, extent, extent]
// via scene.ResolveWorldExtent. Headerless analytical programs return nil, nil
// so they preserve their legacy periodic domain.
func ComputeSpinodoidBounds(s *program.Scene, meshBboxes map[int][6]float64, decls []MeshDecl) ([]float64, error) {
	lo := [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	hi := [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	hasGeometry := false
	include := func(p [3]float64) error {
		for axis, v := range p {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return fmt.Errorf("imported geometry has non-finite bounds")
			}
			lo[axis] = math.Min(lo[axis], v)
			hi[axis] = math.Max(hi[axis], v)
		}
		hasGeometry = true
		return nil
	}

	for slot, bb := range meshBboxes {
		if bb[3] < bb[0] || bb[4] < bb[1] || bb[5] < bb[2] {
			return nil, fmt.Errorf("mesh slot %d has invalid bounds", slot)
		}
		if err := include([3]float64{bb[0], bb[1], bb[2]}); err != nil {
			return nil, err
		}
		if err := include([3]float64{bb[3], bb[4], bb[5]}); err != nil {
			return nil, err
		}
	}

	if hasGeometry {
		for _, decl := range decls {
			bb, ok := meshBboxes[decl.Index]
			if !ok || len(decl.Matrix) != 16 {
				continue
			}
			for corner := 0; corner < 8; corner++ {
				p := [3]float64{bb[0], bb[1], bb[2]}
				for axis := range 3 {
					if corner&(1<<axis) != 0 {
						p[axis] = bb[axis+3]
					}
				}
				if err := include(ApplyMat4Col(decl.Matrix, p)); err != nil {
					return nil, err
				}
			}
		}
		for axis := range 3 {
			pad := math.Max(hi[axis]-lo[axis], 1) * 0.05
			lo[axis] -= pad
			hi[axis] += pad
		}
	} else {
		if s == nil || !(s.Size > 0) {
			return nil, nil
		}
		extent := scene.ResolveWorldExtent(s, 0)
		lo = [3]float64{-extent, -extent, -extent}
		hi = [3]float64{extent, extent, extent}
	}

	return []float64{lo[0], lo[1], lo[2], hi[0], hi[1], hi[2]}, nil
}

// SetProgramSpinodoidBounds computes the spinodoid baking domain via
// ComputeSpinodoidBounds and populates prog.Primary.Scene.SpinodoidBounds.
func SetProgramSpinodoidBounds(prog *program.Program, meshBboxes map[int][6]float64, decls []MeshDecl) error {
	bounds, err := ComputeSpinodoidBounds(prog.Scene(), meshBboxes, decls)
	if err != nil {
		return err
	}
	if len(bounds) == 0 {
		if prog.Scene() != nil {
			prog.Primary.Scene.SpinodoidBounds = nil
		}
		return nil
	}
	if prog.Primary.Scene == nil {
		prog.Primary.Scene = &program.Scene{}
	}
	prog.Primary.Scene.SpinodoidBounds = bounds
	return nil
}
